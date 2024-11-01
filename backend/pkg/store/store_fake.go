package store

import "time"

// FakeStore is primarily used for testing purposes
type FakeStore struct {
	beersLeft int
	isLow     bool
}

func (s *FakeStore) SetWeight(weight float64) error {
	return nil
}

func (s *FakeStore) GetWeight() (float64, error) {
	return 12, nil
}

func (s *FakeStore) SetWeightAt(weightAt time.Time) error {
	return nil
}

func (s *FakeStore) GetWeightAt() (time.Time, error) {
	return time.Now(), nil
}

func (s *FakeStore) SetActiveKeg(weight int) error {
	return nil
}

func (s *FakeStore) GetActiveKeg() (int, error) {
	return 0, nil
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
func (s *FakeStore) SetLastOk(lastOk time.Time) error {
	return nil
}
func (s *FakeStore) GetLastOk() (time.Time, error) {
	return time.Now(), nil
}
func (s *FakeStore) SetOpenAt(openAt time.Time) error {
	return nil
}
func (s *FakeStore) GetOpenAt() (time.Time, error) {
	return time.Now(), nil
}
func (s *FakeStore) SetCloseAt(closeAt time.Time) error {
	return nil
}
func (s *FakeStore) GetCloseAt() (time.Time, error) {
	return time.Now(), nil
}
func (s *FakeStore) SetIsOpen(isOpen bool) error {
	return nil
}
func (s *FakeStore) GetIsOpen() (bool, error) {
	return false, nil
}
