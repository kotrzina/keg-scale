package main

import (
	"bytes"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScale_AddMeasurement(t *testing.T) {
	s := CreateScaleWithMeasurements([]float64{10, 3, 20, 30, 40, 81, 50, 60}...)
	assert.Equal(t, 60000.0, s.Weight)
}

func CreateScaleWithMeasurements(weights ...float64) *Scale {
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	s := NewScale(NewMonitor(), &FakeStore{}, logger, context.Background())
	for _, weight := range weights {
		_ = s.AddMeasurement(weight * 1000)
	}
	return s
}
