package scale

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hako/durafmt"
	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/store"
	"github.com/kotrzina/keg-scale/pkg/wa"
	"github.com/sirupsen/logrus"
)

type Scale struct {
	mux     sync.RWMutex
	monitor *prometheus.Monitor

	weight       float64 // current scale value
	weightAt     time.Time
	candidateKeg int       // candidate keg size
	activeKeg    int       // int value of the active keg in liters
	activeKegAt  time.Time // time when the active keg was set
	beersLeft    int       // how many beers are left in the keg
	beersTotal   int       // how many beers were consumed ever
	isLow        bool      // is the keg low and needs to be replaced soon
	warehouse    [5]int    // warehouse of kegs [10l, 15l, 20l, 30l, 50l]

	pub pub

	lastOk time.Time
	rssi   float64

	events map[EventType][]Event

	store    store.Storage
	config   *config.Config
	whatsapp *wa.WhatsAppClient
	logger   *logrus.Logger
	ctx      context.Context
	fmtUnits durafmt.Units
}

type pub struct {
	isOpen   bool
	openedAt time.Time
	closedAt time.Time
}

const okLimit = 10 * time.Minute

const localizationUnits = "r:r,t:t,d:d,h:h,m:m,s:s,ms:ms,microsecond"

func New(
	ctx context.Context,
	monitor *prometheus.Monitor,
	storage store.Storage,
	conf *config.Config,
	whatsapp *wa.WhatsAppClient,
	logger *logrus.Logger,
) *Scale {
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
		activeKegAt:  time.Unix(0, 0),
		beersLeft:    0,
		beersTotal:   0,
		isLow:        false,
		warehouse:    [5]int{0, 0, 0, 0, 0},

		pub: pub{
			isOpen:   false,
			openedAt: time.Now().Add(-9999 * time.Hour),
			closedAt: time.Now().Add(-9999 * time.Hour),
		},

		lastOk: time.Now().Add(-9999 * time.Hour),

		events: map[EventType][]Event{},

		store:    storage,
		config:   conf,
		whatsapp: whatsapp,
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

	activeKegAt, err := s.store.GetActiveKegAt()
	if err == nil {
		s.activeKegAt = activeKegAt
	}

	beersLeft, err := s.store.GetBeersLeft()
	if err == nil {
		s.beersLeft = beersLeft
		s.monitor.BeersLeft.WithLabelValues().Set(float64(beersLeft))
	}

	beersTotal, err := s.store.GetBeersTotal()
	if err == nil {
		s.beersTotal = beersTotal
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

	s.monitor.BeersTotal.WithLabelValues().Add(float64(s.getBeersTotal()))
}

// AddMeasurement handles a new measurement from the scale
// the most important function in the scale
func (s *Scale) AddMeasurement(weight float64) error {
	if weight < 6000 || weight > 65000 {
		s.logger.Infof("Invalid weight: %.0f", weight)
		return nil
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// set new values to the structure
	s.weight = weight
	s.weightAt = time.Now()
	if serr := s.store.SetWeight(weight); serr != nil {
		return fmt.Errorf("could not store weight: %w", serr)
	}
	if serr := s.store.SetWeightAt(s.weightAt); serr != nil {
		return fmt.Errorf("could not store weight_at: %w", serr)
	}

	// recalculate beers left
	s.beersLeft = CalcBeersLeft(s.activeKeg, weight)
	if serr := s.store.SetBeersLeft(s.beersLeft); serr != nil {
		return fmt.Errorf("could not store beers_left: %w", serr)
	}

	// check empty keg
	if s.beersLeft == 0 {
		if serr := s.addCurrentKegToTotal(); serr != nil {
			return fmt.Errorf("could not add current keg to total: %w", serr)
		}
		s.activeKeg = 0
		if serr := s.store.SetActiveKeg(s.activeKeg); serr != nil {
			return fmt.Errorf("could not store active_keg: %w", serr)
		}
	}

	// check if keg is low
	if !s.isLow {
		s.isLow = IsKegLow(s.activeKeg, weight)
		if s.isLow {
			if serr := s.store.SetIsLow(s.isLow); serr != nil {
				return fmt.Errorf("could not store is_low: %w", serr)
			}
		}
	}

	// check if we expect a new keg
	if s.activeKeg == 0 || s.isLow {
		if serr := s.tryNewKeg(); serr != nil {
			return fmt.Errorf("could not try new keg: %w", serr)
		}
	}

	s.updateMetrics()

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
// - sets the scale to not open after [okLimit] minutes
// it should be called everytime we want to get some calculations
// to recalculate the state of the scale
func (s *Scale) Recheck() {
	s.mux.Lock()
	defer s.mux.Unlock()

	// we haven't received any data for [okLimit] minutes and pub is open
	if !s.isOk() && s.pub.isOpen {
		s.updatePub(false) // close the pub
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

	// manually empty the keg
	if keg == 0 {
		s.isLow = true // enable rekeg
		s.beersLeft = 0
		if serr := s.store.SetBeersLeft(s.beersLeft); serr != nil {
			return fmt.Errorf("could not store beers_left: %w", serr)
		}
		if serr := s.addCurrentKegToTotal(); serr != nil {
			return fmt.Errorf("could not add current keg to total: %w", serr)
		}
	}

	s.activeKeg = keg
	if err := s.store.SetActiveKeg(s.activeKeg); err != nil {
		return err
	}

	if err := s.store.SetIsLow(s.isLow); err != nil {
		return err
	}

	s.updateMetrics()

	return nil
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
	return time.Since(s.lastOk) < okLimit
}

// updatePub updates the pub state
// opening or closing the pub
func (s *Scale) updatePub(isOpen bool) {
	s.pub.isOpen = isOpen
	if err := s.store.SetIsOpen(isOpen); err != nil {
		s.logger.Errorf("Could not set is_open flag: %v", err)
	}

	if isOpen {
		if s.shouldSendOpen() {
			s.dispatchEvent(EventOpen)
		} else {
			s.logger.Warningf("Pub is open, but the opening message has been skipped. Diff: %s", time.Since(s.pub.openedAt).String())
		}

		s.pub.openedAt = time.Now()
		if err := s.store.SetOpenAt(s.pub.openedAt); err != nil {
			s.logger.Errorf("Could not set open_at time: %v", err)
		}
	} else {
		s.pub.closedAt = time.Now().Add(-1 * okLimit)
		if err := s.store.SetCloseAt(s.pub.closedAt); err != nil {
			s.logger.Errorf("Could not set close_at time: %v", err)
		}
		s.dispatchEvent(EventClose)
	}

	fIsOpen := 0.
	if isOpen {
		fIsOpen = 1.
	}

	s.monitor.PubIsOpen.WithLabelValues().Set(fIsOpen)
}

// tryNewKeg tries to find a new keg based on the current weight
// we need at least two measurements to be sure
// first measurement sets the candidate keg
// second measurement sets the active keg
func (s *Scale) tryNewKeg() error {
	keg, err := GuessNewKegSize(s.weight)
	if err == nil {
		// we found a good candidate
		if s.candidateKeg > 0 && s.candidateKeg == keg {
			// we have two measurements with the same keg - rekeg successful !!!

			if s.activeKeg == 50 && keg == 10 {
				// known bug - there is a conflict between 50l and 10l kegs
				// when the 50l keg is empty, the weight is the same as full 10l keg
				// we don't want to rekeg in this case because in many cases it's not true
				s.logger.Warnf("Conflict between 50l and 10l kegs detected. Ignoring rekeg.")
				return nil
			}

			if serr := s.addCurrentKegToTotal(); serr != nil {
				return fmt.Errorf("could not add current keg to total: %w", serr)
			}

			s.candidateKeg = 0
			s.activeKeg = keg
			if serr := s.store.SetActiveKeg(keg); serr != nil {
				return fmt.Errorf("could not store active_keg: %w", serr)
			}
			s.activeKegAt = time.Now()
			if serr := s.store.SetActiveKegAt(s.activeKegAt); serr != nil {
				return fmt.Errorf("could not store active_keg_at: %w", serr)
			}
			s.beersLeft = CalcBeersLeft(s.activeKeg, s.weight)
			if serr := s.store.SetBeersLeft(s.beersLeft); serr != nil {
				return fmt.Errorf("could not store beers_left: %w", serr)
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

			s.dispatchEvent(EventNewKegTapped)
			s.logger.Infof("New keg (%d l) CONFIRMED with current value %.0f", keg, s.weight)
		} else {
			// new candidate keg
			// we already know that the new keg is there, but we need to confirm it
			s.logger.Infof("New keg candidate (%d l) REGISTERED with current value %.0f", keg, s.weight)
			s.candidateKeg = keg
		}
	}

	return nil
}

// getBeersTotal calculates the total amount of beers consumed
// adds two values together - total from the store and the current active keg
func (s *Scale) getBeersTotal() int {
	total := s.beersTotal

	if s.activeKeg > 0 {
		total += CalcBeersConsumed(s.activeKeg, s.weight)
	}

	return total
}

func (s *Scale) addCurrentKegToTotal() error {
	if s.activeKeg == 0 {
		return nil // there is no active keg
	}

	s.beersTotal += s.activeKeg * 2 // liters to beers
	s.monitor.BeersTotal.WithLabelValues().Add(float64(s.getBeersTotal()))
	if err := s.store.SetBeersTotal(s.beersTotal); err != nil {
		return fmt.Errorf("could not store beers_total: %w", err)
	}

	return nil
}

// shouldSendOpen applies the rules for sending a message when the pub is open
// we don't want to spam the group with messages
// it could happen for example when the scale is restarted or lost Wi-Fi connection for a while
func (s *Scale) shouldSendOpen() bool {
	// send message only once in 12 hours
	if time.Since(s.pub.openedAt) < 12*time.Hour {
		return false
	}

	// send message only if the pub was closed for at least 3 hours
	if time.Since(s.pub.closedAt) < 3*time.Hour {
		return false
	}

	return true
}

// updateMetrics updates beer/keg related metrics for prometheus
func (s *Scale) updateMetrics() {
	s.monitor.Weight.WithLabelValues().Set(s.weight)
	s.monitor.BeersLeft.WithLabelValues().Set(float64(s.beersLeft))
	s.monitor.ActiveKeg.WithLabelValues().Set(float64(s.activeKeg))
	s.monitor.BeersTotal.WithLabelValues().Add(float64(s.getBeersTotal()))
}
