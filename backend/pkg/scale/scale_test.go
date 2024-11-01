package scale

import (
	"bytes"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/store"
)

func TestScale_AddMeasurement(t *testing.T) {
	s := CreateScaleWithMeasurements([]float64{10, 3, 20, 30, 40, 81, 50, 60}...)
	assert.Equal(t, 60000.0, s.weight)
}

func CreateScaleWithMeasurements(weights ...float64) *Scale {
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	s := NewScale(prometheus.NewMonitor(), &store.FakeStore{}, logger, context.Background())
	for _, weight := range weights {
		_ = s.AddMeasurement(weight * 1000)
	}
	return s
}
