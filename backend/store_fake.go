package main

// FakeStore is primarily used for testing purposes
type FakeStore struct {
	beersLeft int
	isLow     bool
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

func (s *FakeStore) SetBeersLeft(beersLeft int) error {
	s.beersLeft = beersLeft
	return nil
}

func (s *FakeStore) GetBeersLeft() (int, error) {
	return s.beersLeft, nil
}

func (s *FakeStore) SetIsLow(isLow bool) error {
	s.isLow = isLow
	return nil
}

func (s *FakeStore) GetIsLow() (bool, error) {
	return s.isLow, nil
}

func (s *FakeStore) SetWarehouse(warehouse [5]int) error {
	return nil
}

func (s *FakeStore) GetWarehouse() ([5]int, error) {
	var warehouse = [5]int{1, 2, 3, 4, 5}
	return warehouse, nil
}
