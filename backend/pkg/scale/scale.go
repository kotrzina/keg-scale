package scale

import (
	"context"
	"fmt"
	"github.com/hako/durafmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"

	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/store"
)

const OkLimit = 5 * time.Minute
const localizationUnits = "r:r,t:t,d:d,h:h,m:m,s:s,ms:ms,microsecond"

type Scale struct {
	mux     sync.RWMutex
	monitor *prometheus.Monitor

	weight       float64 // current scale value
	weightAt     time.Time
	candidateKeg int    // candidate keg size
	activeKeg    int    // int value of the active keg in liters
	beersLeft    int    // how many beers are left in the keg
	isLow        bool   // is the keg low and needs to be replaced soon
	warehouse    [5]int // warehouse of kegs [10l, 15l, 20l, 30l, 50l]

	pub pub

	lastOk time.Time
	rssi   float64

	store    store.Storage
	logger   *logrus.Logger
	ctx      context.Context
	fmtUnits durafmt.Units
}

type pub struct {
	isOpen   bool
	openedAt time.Time
	closedAt time.Time
}

func NewScale(monitor *prometheus.Monitor, storage store.Storage, logger *logrus.Logger, ctx context.Context) *Scale {
	fmtUnits, err := durafmt.DefaultUnitsCoder.Decode(localizationUnits)
	if err != nil {
		logger.Fatalf("could not decode units: %v", err)
	}

	s := &Scale{
		mux:     sync.RWMutex{},
		monitor: monitor,

		weight:       0,
		weightAt:     time.Unix(0, 0), // time of last weight measurement
		candidateKeg: 0,
		activeKeg:    0,
		beersLeft:    0,
		isLow:        false,
		warehouse:    [5]int{0, 0, 0, 0, 0},

		pub: pub{
			isOpen:   false,
			openedAt: time.Now().Add(-9999 * time.Hour),
			closedAt: time.Now().Add(-9999 * time.Hour),
		},

		lastOk: time.Now().Add(-9999 * time.Hour),

		store:    storage,
		logger:   logger,
		ctx:      ctx,
		fmtUnits: fmtUnits,
	}

	s.loadDataFromStore()

	// periodically call recheck
	go func(s *Scale) {
		tick := time.NewTicker(15 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-s.ctx.Done():
				s.logger.Debug("Scale recheck stopped")
				return
			case <-tick.C:
				s.Recheck()
			}
		}
	}(s)

	return s
}

func (s *Scale) loadDataFromStore() {
	weight, err := s.store.GetWeight()
	if err == nil {
		s.weight = weight
		s.monitor.Weight.WithLabelValues().Set(weight)
	}

	weightAt, err := s.store.GetWeightAt()
	if err == nil {
		s.weightAt = weightAt
	}

	activeKeg, err := s.store.GetActiveKeg()
	if err == nil {
		s.activeKeg = activeKeg
		s.monitor.ActiveKeg.WithLabelValues().Set(float64(activeKeg))
	}

	beersLeft, err := s.store.GetBeersLeft()
	if err == nil {
		s.beersLeft = beersLeft
		s.monitor.BeersLeft.WithLabelValues().Set(float64(beersLeft))
	}

	isLow, err := s.store.GetIsLow()
	if err == nil {
		s.isLow = isLow
	}

	warehouse, err := s.store.GetWarehouse()
	if err == nil {
		s.warehouse = warehouse
	}

	lastOk, err := s.store.GetLastOk()
	if err == nil {
		s.lastOk = lastOk
	}

	isOpen, err := s.store.GetIsOpen()
	if err == nil {
		s.pub.isOpen = isOpen
	}

	openAt, err := s.store.GetOpenAt()
	if err == nil {
		s.pub.openedAt = openAt
	}

	closeAt, err := s.store.GetCloseAt()
	if err == nil {
		s.pub.closedAt = closeAt
	}
}

func (s *Scale) AddMeasurement(weight float64) error {
	if weight < 6000 || weight > 65000 {
		s.logger.Infof("Invalid weight: %.0f", weight)
		return nil
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	s.weight = weight
	s.weightAt = time.Now()
	if serr := s.store.SetWeight(weight); serr != nil {
		return fmt.Errorf("could not store weight: %w", serr)
	}
	if serr := s.store.SetWeightAt(s.weightAt); serr != nil {
		return fmt.Errorf("could not store weight_at: %w", serr)
	}

	// check if keg is low
	if !s.isLow {
		s.isLow = IsKegLow(s.activeKeg, weight)
		if serr := s.store.SetIsLow(s.isLow); serr != nil {
			return fmt.Errorf("could not store is_low: %w", serr)
		}
	}

	// we expect a new keg
	// we need at least two measurements to be sure
	// first measurement sets the candidate keg
	// second measurement sets the active keg
	if s.activeKeg == 0 || s.isLow {
		keg, err := GuessNewKegSize(weight)
		if err == nil {
			// we found a good candidate

			if s.candidateKeg > 0 && s.candidateKeg == keg {
				// we have two measurements with the same keg
				s.candidateKeg = 0
				s.activeKeg = keg
				if serr := s.store.SetActiveKeg(keg); serr != nil {
					return fmt.Errorf("could not store active_keg: %w", serr)
				}

				s.isLow = false
				if serr := s.store.SetIsLow(false); serr != nil {
					return fmt.Errorf("could not store is_low: %w", serr)
				}

				// remove keg from warehouse
				index, err := GetWarehouseIndex(keg)
				if err != nil {
					return err
				}
				if s.warehouse[index] > 0 {
					s.warehouse[index]--
					if serr := s.store.SetWarehouse(s.warehouse); serr != nil {
						return fmt.Errorf("could not update store warehouse: %w", serr)
					}
				} else {
					s.logger.Warnf("Keg %d is not available in the warehouse", keg)
				}

				s.logger.Infof("New keg (%d l) CONFIRMED with current value %.0f", keg, weight)
			} else {
				// new candidate keg
				// we already know that the new keg is there, but we need to confirm it
				s.logger.Infof("New keg candidate (%d l) REGISTERED with current value %.0f", keg, weight)
				s.candidateKeg = keg
				s.activeKeg = 0
			}
		}
	}

	// calculate values only if we know active keg
	if s.activeKeg > 0 {
		s.beersLeft = CalcBeersLeft(s.activeKeg, weight)
		if serr := s.store.SetBeersLeft(s.beersLeft); serr != nil {
			return fmt.Errorf("could not store beers_left: %w", serr)
		}

		s.monitor.Weight.WithLabelValues().Set(s.weight)
		s.monitor.BeersLeft.WithLabelValues().Set(float64(s.beersLeft))
		s.monitor.ActiveKeg.WithLabelValues().Set(float64(s.activeKeg))
	}

	return nil
}

func (s *Scale) Ping() {
	s.monitor.LastPing.WithLabelValues().SetToCurrentTime()
	now := time.Now()
	err := s.store.SetLastOk(s.lastOk)
	if err != nil {
		s.logger.Errorf("Could not set last_ok time: %v", err)
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.pub.isOpen {
		s.updatePub(true)
	}

	s.lastOk = now
}

// Recheck checks various conditions and states
// - sets the scale to not open after [OkLimit] minutes
// it should be called everytime we want to get some calculations
// to recalculate the state of the scale
func (s *Scale) Recheck() {
	s.mux.Lock()
	defer s.mux.Unlock()

	// we haven't received any data for [OkLimit] minutes and pub is open
	if !s.isOk() && s.pub.isOpen {
		s.updatePub(false)
	}
}

// SetRssi sets the RSSI value of the WiFi signal
func (s *Scale) SetRssi(rssi float64) {
	s.monitor.ScaleWifiRssi.WithLabelValues().Set(rssi)

	s.mux.Lock()
	defer s.mux.Unlock()

	s.rssi = rssi
}

// SetActiveKeg sets the current active keg
func (s *Scale) SetActiveKeg(keg int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isLow = false
	if err := s.store.SetIsLow(false); err != nil {
		return err
	}

	s.activeKeg = keg
	return s.store.SetActiveKeg(keg)
}

func (s *Scale) IncreaseWarehouse(keg int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	index, err := GetWarehouseIndex(keg)
	if err != nil {
		return err
	}

	s.warehouse[index]++
	return s.store.SetWarehouse(s.warehouse)
}

func (s *Scale) DecreaseWarehouse(keg int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	index, err := GetWarehouseIndex(keg)
	if err != nil {
		return err
	}

	if s.warehouse[index] > 0 {
		s.warehouse[index]--
		return s.store.SetWarehouse(s.warehouse)
	}

	return nil
}

// isOk returns true if the scale is ok based on the last update time
func (s *Scale) isOk() bool {
	return time.Since(s.lastOk) < OkLimit
}

// updatePub updates the pub state
// opening or closing the pub
// function is not thread safe
func (s *Scale) updatePub(isOpen bool) {
	s.pub.isOpen = isOpen
	if err := s.store.SetIsOpen(true); err != nil {
		s.logger.Errorf("Could not set is_open flag: %v", err)
	}

	if isOpen {
		s.pub.openedAt = time.Now()
		if err := s.store.SetOpenAt(s.pub.openedAt); err != nil {
			s.logger.Errorf("Could not set open_at time: %v", err)
		}
	} else {
		s.pub.closedAt = time.Now().Add(-1 * OkLimit)
		if err := s.store.SetCloseAt(s.pub.closedAt); err != nil {
			s.logger.Errorf("Could not set close_at time: %v", err)
		}
	}

	fIsOpen := 0.
	if isOpen {
		fIsOpen = 1.
	}

	s.monitor.PubIsOpen.WithLabelValues().Set(fIsOpen)
}
