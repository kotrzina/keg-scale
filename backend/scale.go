package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const OkLimit = 5 * time.Minute

type Pub struct {
	IsOpen   bool      `json:"is_open"`
	OpenedAt time.Time `json:"open_at"`
	ClosedAt time.Time `json:"closed_at"`
}

type Scale struct {
	mux     sync.Mutex
	monitor *Monitor

	Weight       float64   `json:"weight"` // current scale value
	WeightAt     time.Time `json:"last_weight_at"`
	CandidateKeg int       `json:"candidate_keg"` // candidate keg size
	ActiveKeg    int       `json:"active_keg"`    // int value of the active keg in liters
	BeersLeft    int       `json:"beers_left"`    // how many beers are left in the keg
	IsLow        bool      `json:"is_low"`        // is the keg low and needs to be replaced soon
	Warehouse    [5]int    `json:"warehouse"`     // warehouse of kegs [10l, 15l, 20l, 30l, 50l]

	Pub Pub `json:"pub"`

	LastOk time.Time `json:"last_ok"`
	Rssi   float64   `json:"rssi"`

	store  Storage
	logger *logrus.Logger
	ctx    context.Context
}

func NewScale(monitor *Monitor, store Storage, logger *logrus.Logger, ctx context.Context) *Scale {
	s := &Scale{
		mux:     sync.Mutex{},
		monitor: monitor,

		Weight:       0,
		WeightAt:     time.Unix(0, 0), // time of last weight measurement
		CandidateKeg: 0,
		ActiveKeg:    0,
		BeersLeft:    0,
		IsLow:        false,
		Warehouse:    [5]int{0, 0, 0, 0, 0},

		Pub: Pub{
			IsOpen:   false,
			OpenedAt: time.Now().Add(-9999 * time.Hour),
			ClosedAt: time.Now().Add(-9999 * time.Hour),
		},

		LastOk: time.Now().Add(-9999 * time.Hour),

		store:  store,
		logger: logger,
		ctx:    ctx,
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
		s.Weight = weight
		s.monitor.weight.WithLabelValues().Set(weight)
	}

	weightAt, err := s.store.GetWeightAt()
	if err == nil {
		s.WeightAt = weightAt
	}

	activeKeg, err := s.store.GetActiveKeg()
	if err == nil {
		s.ActiveKeg = activeKeg
		s.monitor.activeKeg.WithLabelValues().Set(float64(activeKeg))
	}

	beersLeft, err := s.store.GetBeersLeft()
	if err == nil {
		s.BeersLeft = beersLeft
		s.monitor.beersLeft.WithLabelValues().Set(float64(beersLeft))
	}

	isLow, err := s.store.GetIsLow()
	if err == nil {
		s.IsLow = isLow
	}

	warehouse, err := s.store.GetWarehouse()
	if err == nil {
		s.Warehouse = warehouse
	}

	lastOk, err := s.store.GetLastOk()
	if err == nil {
		s.LastOk = lastOk
	}

	isOpen, err := s.store.GetIsOpen()
	if err == nil {
		s.Pub.IsOpen = isOpen
	}

	openAt, err := s.store.GetOpenAt()
	if err == nil {
		s.Pub.OpenedAt = openAt
	}

	closeAt, err := s.store.GetCloseAt()
	if err == nil {
		s.Pub.ClosedAt = closeAt
	}
}

func (s *Scale) AddMeasurement(weight float64) error {
	if weight < 6000 || weight > 65000 {
		s.logger.Infof("Invalid weight: %.0f", weight)
		return nil
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	s.Weight = weight
	s.WeightAt = time.Now()
	if serr := s.store.SetWeight(weight); serr != nil {
		return fmt.Errorf("could not store weight: %w", serr)
	}
	if serr := s.store.SetWeightAt(s.WeightAt); serr != nil {
		return fmt.Errorf("could not store weight_at: %w", serr)
	}

	// check if keg is low
	if !s.IsLow {
		s.IsLow = IsKegLow(s.ActiveKeg, weight)
		if serr := s.store.SetIsLow(s.IsLow); serr != nil {
			return fmt.Errorf("could not store is_low: %w", serr)
		}
	}

	// we expect a new keg
	// we need at least two measurements to be sure
	// first measurement sets the candidate keg
	// second measurement sets the active keg
	if s.ActiveKeg == 0 || s.IsLow {
		keg, err := GuessNewKegSize(weight)
		if err == nil {
			// we found a good candidate

			if s.CandidateKeg > 0 && s.CandidateKeg == keg {
				// we have two measurements with the same keg
				s.CandidateKeg = 0
				s.ActiveKeg = keg
				if serr := s.store.SetActiveKeg(keg); serr != nil {
					return fmt.Errorf("could not store active_keg: %w", serr)
				}

				s.IsLow = false
				if serr := s.store.SetIsLow(false); serr != nil {
					return fmt.Errorf("could not store is_low: %w", serr)
				}

				// remove keg from warehouse
				index, err := GetWarehouseIndex(keg)
				if err != nil {
					return err
				}
				if s.Warehouse[index] > 0 {
					s.Warehouse[index]--
					if serr := s.store.SetWarehouse(s.Warehouse); serr != nil {
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
				s.CandidateKeg = keg
				s.ActiveKeg = 0
			}
		}
	}

	// calculate values only if we know active keg
	if s.ActiveKeg > 0 {
		s.BeersLeft = CalcBeersLeft(s.ActiveKeg, weight)
		if serr := s.store.SetBeersLeft(s.BeersLeft); serr != nil {
			return fmt.Errorf("could not store beers_left: %w", serr)
		}

		s.monitor.weight.WithLabelValues().Set(s.Weight)
		s.monitor.beersLeft.WithLabelValues().Set(float64(s.BeersLeft))
		s.monitor.activeKeg.WithLabelValues().Set(float64(s.ActiveKeg))
	}

	return nil
}

func (s *Scale) JsonState() ([]byte, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return json.Marshal(s)
}

func (s *Scale) Ping() {
	s.monitor.lastPing.WithLabelValues().SetToCurrentTime()
	now := time.Now()
	err := s.store.SetLastOk(s.LastOk)
	if err != nil {
		s.logger.Errorf("Could not set last_ok time: %v", err)
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.Pub.IsOpen {
		s.updatePub(true)
	}

	s.LastOk = now
}

// Recheck checks various conditions and states
// - sets the scale to not open after [OkLimit] minutes
// it should be called everytime we want to get some calculations
// to recalculate the state of the scale
func (s *Scale) Recheck() {
	ok := s.IsOk() // mutex

	s.mux.Lock()
	defer s.mux.Unlock()

	// we haven't received any data for [OkLimit] minutes and pub is open
	if !ok && s.Pub.IsOpen {
		s.updatePub(false)
	}
}

// IsOk returns true if the scale is ok based on the last update time
func (s *Scale) IsOk() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	return time.Since(s.LastOk) < OkLimit
}

// SetRssi sets the RSSI value of the WiFi signal
func (s *Scale) SetRssi(rssi float64) {
	s.monitor.scaleWifiRssi.WithLabelValues().Set(rssi)

	s.mux.Lock()
	defer s.mux.Unlock()

	s.Rssi = rssi
}

// SetActiveKeg sets the current active keg
func (s *Scale) SetActiveKeg(keg int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.IsLow = false
	if err := s.store.SetIsLow(false); err != nil {
		return err
	}

	s.ActiveKeg = keg
	return s.store.SetActiveKeg(keg)
}

func (s *Scale) IncreaseWarehouse(keg int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	index, err := GetWarehouseIndex(keg)
	if err != nil {
		return err
	}

	s.Warehouse[index]++
	return s.store.SetWarehouse(s.Warehouse)
}

func (s *Scale) DecreaseWarehouse(keg int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	index, err := GetWarehouseIndex(keg)
	if err != nil {
		return err
	}

	if s.Warehouse[index] > 0 {
		s.Warehouse[index]--
		return s.store.SetWarehouse(s.Warehouse)
	}

	return nil
}

// updatePub updates the pub state
// opening or closing the pub
// function is not thread safe
func (s *Scale) updatePub(isOpen bool) {
	s.Pub.IsOpen = isOpen
	if err := s.store.SetIsOpen(true); err != nil {
		s.logger.Errorf("Could not set is_open flag: %v", err)
	}

	if isOpen {
		s.Pub.OpenedAt = time.Now()
		if err := s.store.SetOpenAt(s.Pub.OpenedAt); err != nil {
			s.logger.Errorf("Could not set open_at time: %v", err)
		}
	} else {
		s.Pub.ClosedAt = time.Now().Add(-1 * OkLimit)
		if err := s.store.SetCloseAt(s.Pub.ClosedAt); err != nil {
			s.logger.Errorf("Could not set close_at time: %v", err)
		}
	}

	fIsOpen := 0.
	if isOpen {
		fIsOpen = 1.
	}

	s.monitor.pubIsOpen.WithLabelValues().Set(fIsOpen)
}
