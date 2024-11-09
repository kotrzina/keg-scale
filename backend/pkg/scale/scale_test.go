package scale

import (
	"bytes"
	"context"
	"testing"

	"github.com/kotrzina/keg-scale/pkg/hook"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/store"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestScale_AddMeasurement(t *testing.T) {
	s := createScaleWithMeasurements(t, []float64{10, 3, 20, 30, 40, 81, 50, 60}...)
	assert.Equal(t, 60000.0, s.weight)
}

func createScaleWithMeasurements(t *testing.T, weights ...float64) *Scale {
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	s := New(context.Background(), prometheus.New(), &store.FakeStore{}, &hook.Discord{}, logger)
	for _, weight := range weights {
		assert.Nil(t, s.AddMeasurement(weight*1000))
	}
	return s
}
