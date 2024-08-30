package main

import (
	"encoding/json"
	"sync"
	"time"
)

type Measurement struct {
	Index  int       `json:"index"`
	Weight float64   `json:"weight"`
	At     time.Time `json:"at"`
}

type Scale struct {
	mux sync.Mutex

	Measurements [10]Measurement `json:"measurements"`
	Index        int             `json:"index"`

	IsOk   bool      `json:"is_ok"`
	LastOk time.Time `json:"last_ok"`
}

func NewScale() *Scale {
	return &Scale{
		mux: sync.Mutex{},

		Measurements: [10]Measurement{},
		Index:        -1,

		IsOk:   false,
		LastOk: time.Now().Add(-9999 * time.Hour),
	}
}

func (s *Scale) AddMeasurement(weight float64) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.Index++
	if s.Index >= len(s.Measurements) {
		s.Index = 0
	}

	s.Measurements[s.Index] = Measurement{
		Index:  s.Index,
		Weight: weight,
		At:     time.Now(),
	}
}

func (s *Scale) JsonState() ([]byte, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return json.Marshal(s)
}

func (s *Scale) Ping() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.IsOk = true
	s.LastOk = time.Now()
}
