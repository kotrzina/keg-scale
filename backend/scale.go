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

type Measurement struct {
	Index  int       `json:"index"`
	Weight float64   `json:"weight"`
	At     time.Time `json:"at"`
}

type Pub struct {
	IsOpen   bool      `json:"is_open"`
	OpenedAt time.Time `json:"open_at"`
	ClosedAt time.Time `json:"closed_at"`
}

type Scale struct {
	mux     sync.Mutex
	monitor *Monitor

	Measurements []Measurement `json:"measurements"` // @todo - this might be useless (we need just the last value)
	index        int
	size         int
	valid        int // number of valid measurements

	ActiveKeg int    `json:"active_keg"` // int value of the active keg in liters
	BeersLeft int    `json:"beers_left"` // how many beers are left in the keg
	IsLow     bool   `json:"is_low"`     // is the keg low and needs to be replaced soon
	Warehouse [5]int `json:"warehouse"`  // warehouse of kegs [10l, 15l, 20l, 30l, 50l]

	Pub Pub `json:"pub"`

	LastOk time.Time `json:"last_ok"`
	Rssi   float64   `json:"rssi"`

	store  Storage
	logger *logrus.Logger
	ctx    context.Context
}

func NewScale(bufferSize int, monitor *Monitor, store Storage, logger *logrus.Logger, ctx context.Context) *Scale {
	s := &Scale{
		mux:     sync.Mutex{},
		monitor: monitor,

		Measurements: make([]Measurement, bufferSize),
		index:        -1,
		size:         bufferSize,
		valid:        0,

		ActiveKeg: 0,
		BeersLeft: 0,
		IsLow:     false,
		Warehouse: [5]int{0, 0, 0, 0, 0},

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
	measurements, err := s.store.GetMeasurements()
	if err == nil {
		s.index = len(measurements) - 1
		i := 0
		for _, m := range measurements {
			s.Measurements[i] = m
		}
	}

	activeKeg, err := s.store.GetActiveKeg()
	if err == nil {
		s.ActiveKeg = activeKeg
	}

	beersLeft, err := s.store.GetBeersLeft()
	if err == nil {
		s.BeersLeft = beersLeft
	}

	isLow, err := s.store.GetIsLow()
	if err == nil {
		s.IsLow = isLow
	}

	warehouse, err := s.store.GetWarehouse()
	if err == nil {
		s.Warehouse = warehouse
	}
}

func (s *Scale) AddMeasurement(weight float64) error {
	if weight < 6000 || weight > 65000 {
		s.logger.Infof("Invalid weight: %f", weight)
		return nil
	}

	s.monitor.kegWeight.WithLabelValues().Set(weight)

	s.mux.Lock()
	defer s.mux.Unlock()

	s.index++
	if s.index >= len(s.Measurements) {
		s.index = 0
	}

	m := Measurement{
		Index:  s.index,
		Weight: weight,
		At:     time.Now(),
	}

	s.Measurements[s.index] = m
	err := s.store.AddMeasurement(m)
	if err != nil {
		return fmt.Errorf("could not store measurement: %w", err)
	}

	if s.valid < s.size {
		s.valid++
	}

	// check if keg is low
	if !s.IsLow {
		s.IsLow = IsKegLow(s.ActiveKeg, weight)
		if serr := s.store.SetIsLow(s.IsLow); serr != nil {
			return fmt.Errorf("could not store is_low: %w", serr)
		}
	}

	// we expect a new keg
	if s.ActiveKeg == 0 || s.IsLow {
		keg, err := GuessNewKegSize(weight)
		if err == nil {
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
		}
	}

	s.BeersLeft = CalcBeersLeft(s.ActiveKeg, weight)
	if serr := s.store.SetBeersLeft(s.BeersLeft); serr != nil {
		return fmt.Errorf("could not store beers_left: %w", serr)
	}
	s.monitor.beersLeft.WithLabelValues().Set(float64(s.BeersLeft))
	s.monitor.activeKeg.WithLabelValues().Set(float64(s.ActiveKeg))

	return nil
}

func (s *Scale) GetLastMeasurement() Measurement {
	return s.GetMeasurement(0)
}

// GetMeasurement GetValidCount return number of valid measurements
func (s *Scale) GetMeasurement(index int) Measurement {
	if index > s.GetValidCount() || index > s.size {
		return Measurement{
			Weight: 0,
			Index:  -1,
		}
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	idx := (s.index - index + s.size) % s.size
	return s.Measurements[idx]
}

func (s *Scale) JsonState() ([]byte, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return json.Marshal(s)
}

func (s *Scale) Ping() {
	s.monitor.lastUpdate.WithLabelValues().SetToCurrentTime()

	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.Pub.IsOpen {
		s.monitor.pubIsOpen.WithLabelValues().Set(1)
		s.Pub.IsOpen = true
		s.Pub.OpenedAt = time.Now()
	}

	s.LastOk = time.Now()
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
		s.monitor.pubIsOpen.WithLabelValues().Set(0)
		s.Pub.IsOpen = false
		s.Pub.ClosedAt = time.Now().Add(-1 * OkLimit)
	}
}

func (s *Scale) IsOk() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	return time.Since(s.LastOk) < OkLimit
}

func (s *Scale) SetRssi(rssi float64) {
	s.monitor.scaleWifiRssi.WithLabelValues().Set(rssi)

	s.mux.Lock()
	defer s.mux.Unlock()

	s.Rssi = rssi
}

// GetValidCount returns the number of valid measurements
func (s *Scale) GetValidCount() int {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.valid
}

// HasLastN returns true if the last n measurements are not empty
func (s *Scale) HasLastN(n int) bool {
	if n > s.size {
		n = s.size
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	return s.valid >= n
}

// SumLastN returns the sum of the last n measurements
func (s *Scale) SumLastN(n int) float64 {
	if n > s.size {
		n = s.size
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	sum := 0.0
	for i := 0; i < n; i++ {
		idx := (s.index - i + s.size) % s.size
		sum += s.Measurements[idx].Weight
	}

	return sum
}

// AvgLastN returns the average of the last n measurements
// It ignores empty measurements - you should call HasLastN before calling this
func (s *Scale) AvgLastN(n int) float64 {
	if n > s.size {
		n = s.size
	}

	if n == 0 {
		return 0
	}

	return s.SumLastN(n) / float64(n)
}

func (s *Scale) SetActiveKeg(keg int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

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
