package main

import "time"

type Storage interface {
	SetWeight(weight float64) error // set weight
	GetWeight() (float64, error)    // get weight

	SetWeightAt(weightAt time.Time) error // set weight at
	GetWeightAt() (time.Time, error)      // get weight at

	SetActiveKeg(weight int) error // set active keg
	GetActiveKeg() (int, error)    // get active keg

	SetBeersLeft(beersLeft int) error // set beers left
	GetBeersLeft() (int, error)       // get beers left

	SetIsLow(isLow bool) error // set is low flag
	GetIsLow() (bool, error)   // get is low flag

	SetWarehouse(warehouse [5]int) error // set warehouse
	GetWarehouse() ([5]int, error)       // get warehouse
}
