package main

// FakeStore is primarily used for testing purposes
type FakeStore struct {
}

func (s *FakeStore) SetActiveKeg(weight int) error {
	return nil
}

func (s *FakeStore) GetActiveKeg() (int, error) {
	return 0, nil
}

func (s *FakeStore) AddMeasurement(m Measurement) error {
	return nil
}

func (s *FakeStore) GetMeasurements() ([]Measurement, error) {
	var measurements []Measurement
	return measurements, nil
}
