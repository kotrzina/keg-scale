package main

import (
	"fmt"
	"math"
)

type KegWeights map[int]float64

// GetEmptyWeights returns a map of keg sizes and their empty weights in grams
func GetEmptyWeights() KegWeights {
	return KegWeights{
		10: 6000,
		15: 7000,
		20: 7500,
		30: 10000,
		50: 11000,
	}
}

// GetFullWeights returns a map of keg sizes and their full weights in grams
func GetFullWeights() KegWeights {
	w := make(map[int]float64)
	empty := GetEmptyWeights()
	for keg, weight := range empty {
		w[keg] = float64(keg)*1000 + weight
	}

	return w
}

// CalcBeersLeft calculates the number of beers left in a keg based on its size and current weight
func CalcBeersLeft(keg int, weight float64) int {
	kegWeight, found := GetEmptyWeights()[keg]
	if !found {
		kegWeight = 0
	}

	if kegWeight/1000 > weight/1000 {
		return 0
	}

	return int(math.Floor((weight/1000 - kegWeight/1000) * 2))
}

func IsKegLow(keg int, weight float64) bool {
	if keg == 0 {
		return true // no keg is set - is low for a new one
	}

	kegWeight, found := GetEmptyWeights()[keg]
	if !found {
		return true // unknown keg - islow for a new one
	}

	return math.Abs(weight-kegWeight) < 2500 // we are 2500 grams close to the empty keg
}

func GuessNewKegSize(weight float64) (int, error) {
	kegs := GetFullWeights()
	delta := 2000.0
	for keg, fullWeight := range kegs {
		if math.Abs(weight-fullWeight) < delta {
			return keg, nil
		}
	}

	return 0, fmt.Errorf("could not guess keg size based on weight: %f", weight)
}
