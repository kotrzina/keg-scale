package main

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestScale_GetMeasurement(t *testing.T) {
	type testcase struct {
		size    int
		weights []float64
		index   int
		weight  float64
	}

	testcases := []testcase{
		{10, []float64{10, 20, 30, 40, 50, 60}, 0, 60},
		{10, []float64{10, 20, 30, 40, 50, 60}, 1, 50},
		{10, []float64{10, 20, 30, 40, 50, 60}, 5, 10},
		{4, []float64{10, 20, 30, 40, 50, 60}, 2, 40}, // overflown buffer
	}

	for _, tc := range testcases {
		s := CreateScaleWithMeasurements(tc.size, tc.weights...)
		measurement := s.GetMeasurement(tc.index)

		if measurement.Weight != tc.weight*1000 {
			t.Errorf("Expected weight to be %f, got %f", tc.weight*1000, measurement.Weight)
		}
	}
}

func TestScale_GetValidCount(t *testing.T) {
	type testcase struct {
		size    int
		weights []float64
		count   int
	}

	testcases := []testcase{
		{10, []float64{}, 0},
		{10, []float64{10, 20, 30, 40, 50, 60}, 6},
		{1000, []float64{10, 20, 30, 40, 50, 60}, 6},
		{4, []float64{10, 20, 30, 40, 50, 60}, 4}, // buffer is full and overflown
	}

	for _, tc := range testcases {
		s := CreateScaleWithMeasurements(tc.size, tc.weights...)
		count := s.GetValidCount()

		if count != tc.count {
			t.Errorf("Expected count to be %d, got %d", tc.count, count)
		}
	}
}

func TestScale_SumLastN(t *testing.T) {
	type testcase struct {
		size    int
		weights []float64
		n       int
		sum     float64
	}

	testcases := []testcase{
		{10, []float64{10, 20, 30, 40, 50, 60}, 0, 0},
		{10, []float64{10, 20, 30, 40, 50, 60}, 1, 60},
		{10, []float64{10, 20, 30, 40, 50, 60}, 3, 150},
		{10, []float64{10, 20, 30, 40, 50, 60}, 4, 180},
		{10, []float64{11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, 7, 154},
		{10, []float64{11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, 10, 205}, // all values in the buffer
		{10, []float64{}, 2, 0},                                                              // empty buffer
		{10, []float64{10, 20}, 4, 30},                                                       // not enough values in the buffer
	}

	for _, tc := range testcases {
		s := CreateScaleWithMeasurements(tc.size, tc.weights...)
		sum := s.SumLastN(tc.n)

		if sum != tc.sum*1000 {
			t.Errorf("Expected sum to be %f, got %f", tc.sum*1000, sum)
		}
	}

}

func TestScale_AvgLastN(t *testing.T) {
	type testcase struct {
		size    int
		weights []float64
		n       int
		avg     float64
	}

	testcases := []testcase{
		{10, []float64{10, 20, 30, 40, 50, 60}, 0, 0},
		{10, []float64{10, 20, 30, 40, 50, 60}, 1, 60},
		{10, []float64{10, 20, 30, 40, 50, 60}, 3, 50},
		{10, []float64{10, 20}, 4, 7.5}, // not enough values in the buffer
	}

	for _, tc := range testcases {
		s := CreateScaleWithMeasurements(tc.size, tc.weights...)
		avg := s.AvgLastN(tc.n)

		if avg != tc.avg*1000 {
			t.Errorf("Expected avg to be %f, got %f", tc.avg*1000, avg)
		}
	}
}

func TestScale_HasLastN(t *testing.T) {
	type testcase struct {
		size    int
		weights []float64
		n       int
		has     bool
	}

	testcases := []testcase{
		{10, []float64{10, 20, 30, 40, 50, 60}, 3, true},
		{10, []float64{10, 20, 30, 40, 50, 60}, 6, true},
		{10, []float64{10, 20, 30, 40, 50, 60}, 7, false},
		{10, []float64{10, 20}, 4, false},
	}

	for _, tc := range testcases {
		s := CreateScaleWithMeasurements(tc.size, tc.weights...)
		has := s.HasLastN(tc.n)

		if has != tc.has {
			t.Errorf("Expected has to be %t, got %t", tc.has, has)
		}
	}
}

func TestScale_AddMeasurement(t *testing.T) {
	type testcase struct {
		size       int
		weights    []float64
		validCount int
	}

	testcases := []testcase{
		{10, []float64{10, 3, 20, 30, 40, 81, 50, 60}, 6}, // 3 and 81 are invalid
	}

	for _, tc := range testcases {
		s := CreateScaleWithMeasurements(tc.size, tc.weights...)
		count := s.GetValidCount()

		if count != tc.validCount {
			t.Errorf("Expected count to be %d, got %d", tc.validCount, count)
		}
	}
}

func CreateScaleWithMeasurements(size int, weights ...float64) *Scale {
	logger := logrus.New()
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	s := NewScale(size, NewMonitor(), &FakeStore{}, logger)
	for _, weight := range weights {
		_ = s.AddMeasurement(weight * 1000)
	}
	return s
}
