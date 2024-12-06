package scale

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
		20: 9250,
		30: 10000,
		50: 13500,
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
	if keg == 0 {
		return 0
	}
	kegWeight, found := GetEmptyWeights()[keg]
	if !found {
		kegWeight = 0
	}

	if kegWeight/1000 > weight/1000 {
		return 0
	}

	return int(math.Floor((weight/1000 - kegWeight/1000) * 2))
}

// CalcBeersConsumed calculates the number of beers consumed from a keg based on its size and current weight
func CalcBeersConsumed(keg int, weight float64) int {
	kegWeight, found := GetEmptyWeights()[keg]
	if !found {
		return 0
	}
	fullKeg := float64(keg) * 2 // how many beers do we have in full keg

	w := weight - kegWeight

	if w <= 0 {
		return keg * 2
	}

	if w >= float64(keg)*1000 {
		return int(fullKeg)
	}

	return int(math.Floor(fullKeg - (w / 500)))
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
	delta := 2500.0
	for keg, fullWeight := range kegs {
		if math.Abs(weight-fullWeight) < delta {
			return keg, nil
		}
	}

	return 0, fmt.Errorf("could not guess keg size based on weight: %f", weight)
}
