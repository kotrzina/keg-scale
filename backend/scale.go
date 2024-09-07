package main

import (
	"encoding/json"
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

	Measurements []Measurement `json:"measurements"`
	index        int
	size         int
	valid        int // number of valid measurements

	Pub Pub `json:"pub"`

	LastOk time.Time `json:"last_ok"`
	Rssi   float64   `json:"rssi"`
}

func NewScale(bufferSize int, monitor *Monitor) *Scale {
	s := &Scale{
		mux:     sync.Mutex{},
		monitor: monitor,

		Measurements: make([]Measurement, bufferSize),
		index:        -1,
		size:         bufferSize,
		valid:        0,

		Pub: Pub{
			IsOpen:   false,
			OpenedAt: time.Now().Add(-9999 * time.Hour),
			ClosedAt: time.Now().Add(-9999 * time.Hour),
		},

		LastOk: time.Now().Add(-9999 * time.Hour),
	}

	// periodically call recheck
	go func(s *Scale) {
		for {
			time.Sleep(15 * time.Second)
			s.Recheck()
		}
		// @todo - I don't really care about cancellation right now
	}(s)

	return s
}

func (s *Scale) AddMeasurement(weight float64) {
	s.monitor.kegWeight.WithLabelValues().Set(weight)

	s.mux.Lock()
	defer s.mux.Unlock()

	s.index++
	if s.index >= len(s.Measurements) {
		s.index = 0
	}

	s.Measurements[s.index] = Measurement{
		Index:  s.index,
		Weight: weight,
		At:     time.Now(),
	}

	if s.valid < s.size {
		s.valid++
	}
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

// Recheck sets the scale to not open
// it should be called everytime we want to get some calculations
// to recalculate the state of the scale
func (s *Scale) Recheck() {
	ok := s.IsOk() // mutex

	if ok {
		return
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if s.Pub.IsOpen { // we haven't received any data for [OkLimit] minutes and pub is open
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
