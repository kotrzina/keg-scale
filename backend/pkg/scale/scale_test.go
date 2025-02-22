package scale

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
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
	s := New(
		context.Background(),
		prometheus.New(),
		&store.FakeStore{},
		config.NewConfig(),
		logger,
	)
	for _, weight := range weights {
		assert.Nil(t, s.AddMeasurement(weight*1000))
	}
	return s
}

func TestScale_shouldSendOpen(t *testing.T) {
	cases := []struct {
		name        string
		openBefore  time.Duration
		closeBefore time.Duration
		shouldSend  bool
	}{
		{"should send open", 24 * time.Hour, 19 * time.Hour, true},
		{"once per 12 hours", 9 * time.Hour, 19 * time.Hour, false},
		{"at least 3 hours closed", 24 * time.Hour, 2 * time.Hour, false},
	}

	for _, tt := range cases {
		s := &Scale{
			pub: pub{
				openedAt: time.Now().Add(-tt.openBefore),
				closedAt: time.Now().Add(-tt.closeBefore),
			},
		}

		assert.Equal(t, tt.shouldSend, s.shouldSendOpen(), tt.name)
	}
}
