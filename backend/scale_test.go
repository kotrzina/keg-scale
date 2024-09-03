package main

import "testing"

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
		{10, []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150}, 7, 840},
		{10, []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150}, 10, 1050}, // all values in the buffer
		{10, []float64{}, 2, 0},        // empty buffer
		{10, []float64{10, 20}, 4, 30}, // not enough values in the buffer
	}

	for _, tc := range testcases {
		s := CreateScaleWithMeasurements(tc.size, tc.weights...)
		sum := s.SumLastN(tc.n)

		if sum != tc.sum {
			t.Errorf("Expected sum to be %f, got %f", tc.sum, sum)
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

		if avg != tc.avg {
			t.Errorf("Expected avg to be %f, got %f", tc.avg, avg)
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

func CreateScaleWithMeasurements(size int, weights ...float64) *Scale {
	s := NewScale(size)
	for _, weight := range weights {
		s.AddMeasurement(weight)
	}

	return s
}
