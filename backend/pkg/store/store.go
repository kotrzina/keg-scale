package store

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

	SetBeersTotal(beersTotal int) error // set beers total
	GetBeersTotal() (int, error)        // get beers total

	SetIsLow(isLow bool) error // set is low flag
	GetIsLow() (bool, error)   // get is low flag

	SetWarehouse(warehouse [5]int) error // set warehouse
	GetWarehouse() ([5]int, error)       // get warehouse

	SetLastOk(lastOk time.Time) error // set last ok
	GetLastOk() (time.Time, error)    // get last ok

	SetOpenAt(openAt time.Time) error // set open at
	GetOpenAt() (time.Time, error)    // get open at

	SetCloseAt(closeAt time.Time) error // set close at
	GetCloseAt() (time.Time, error)     // get close at

	SetIsOpen(isOpen bool) error // set is open flag
	GetIsOpen() (bool, error)    // get is open flag

	SetTotalBeers(totalBeers int) error // set total beers
	GetTotalBeers() (int, error)        // get total beers
}
