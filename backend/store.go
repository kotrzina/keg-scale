package main

type Storage interface {
	SetActiveKeg(weight int) error           // set active keg
	GetActiveKeg() (int, error)              // get active keg
	AddMeasurement(m Measurement) error      // add measurement to the store
	GetMeasurements() ([]Measurement, error) // get all measurements
	SetIsLow(isLow bool) error               // set is low flag
	GetIsLow() (bool, error)                 // get is low flag
}
