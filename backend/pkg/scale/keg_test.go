package scale

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalcBeersLeft(t *testing.T) {
	type testcase struct {
		keg    int
		weight float64
		beers  int
	}

	testcases := []testcase{
		{10, 6100, 0},
		{10, 7500, 3},
		{15, 7000, 0},
		{15, 8500, 3},
		{20, 7250, 0},
		{20, 8750, 2},
		{30, 7500, 0},
		{30, 11200, 2},
		{50, 10100, 0},
		{50, 12100, 2},
		{90, 10000, 20}, // unknown keg - ignore weight
	}

	for _, tc := range testcases {
		beers := CalcBeersLeft(tc.keg, tc.weight)
		assert.Equal(t, tc.beers, beers, "Keg %d with weight %f - Expected beers to be %d, got %d", tc.keg, tc.weight, tc.beers, beers)
	}
}

func TestIsKegLow(t *testing.T) {
	type testcase struct {
		keg    int
		weight float64
		isLow  bool
	}

	testcases := []testcase{
		{0, 0, true},     // always low
		{0, 10000, true}, // always low
		{10, 7100, true},
		{10, 10500, false},
		{30, 8100, true},
		{30, 17500, false},
		{90, 10000, true}, // unknown keg - always low
	}

	for _, tc := range testcases {
		ready := IsKegLow(tc.keg, tc.weight)
		assert.Equal(t, tc.isLow, ready, "Expected is_low to be %t, got %t", tc.isLow, ready)
	}
}

func TestGuessNewKegSize(t *testing.T) {
	type testcase struct {
		weight float64
		keg    int
	}

	testcases := []testcase{
		{16500, 10},
		{22200, 15},
	}

	for _, tc := range testcases {
		keg, err := GuessNewKegSize(tc.weight)
		assert.Nil(t, err)
		assert.Equal(t, tc.keg, keg, "Expected keg to be %d, got %d", tc.keg, keg)
	}
}
