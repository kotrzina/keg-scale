package store

import (
	"time"
)

// FakeStore is primarily used for testing purposes
type FakeStore struct {
	beersLeft int
	isLow     bool
}

func (s *FakeStore) AddEvent(_ string) error {
	return nil
}

func (s *FakeStore) GetEvents() ([]string, error) {
	return []string{}, nil
}

func (s *FakeStore) SetWeight(_ float64) error {
	return nil
}

func (s *FakeStore) GetWeight() (float64, error) {
	return 12, nil
}

func (s *FakeStore) SetWeightAt(_ time.Time) error {
	return nil
}

func (s *FakeStore) GetWeightAt() (time.Time, error) {
	return time.Now(), nil
}

func (s *FakeStore) SetActiveKeg(_ int) error {
	return nil
}

func (s *FakeStore) GetActiveKeg() (int, error) {
	return 0, nil
}

func (s *FakeStore) SetActiveKegAt(_ time.Time) error {
	return nil
}

func (s *FakeStore) GetActiveKegAt() (time.Time, error) {
	return time.Now(), nil
}

func (s *FakeStore) SetBeersLeft(beersLeft int) error {
	s.beersLeft = beersLeft
	return nil
}

func (s *FakeStore) GetBeersLeft() (int, error) {
	return s.beersLeft, nil
}

func (s *FakeStore) SetBeersTotal(_ int) error {
	return nil
}

func (s *FakeStore) GetBeersTotal() (int, error) {
	return 0, nil
}

func (s *FakeStore) SetIsLow(isLow bool) error {
	s.isLow = isLow
	return nil
}

func (s *FakeStore) GetIsLow() (bool, error) {
	return s.isLow, nil
}

func (s *FakeStore) SetWarehouse(_ [5]int) error {
	return nil
}

func (s *FakeStore) GetWarehouse() ([5]int, error) {
	warehouse := [5]int{1, 2, 3, 4, 5}
	return warehouse, nil
}

func (s *FakeStore) SetLastOk(_ time.Time) error {
	return nil
}

func (s *FakeStore) GetLastOk() (time.Time, error) {
	return time.Now(), nil
}

func (s *FakeStore) SetOpenAt(_ time.Time) error {
	return nil
}

func (s *FakeStore) GetOpenAt() (time.Time, error) {
	return time.Now(), nil
}

func (s *FakeStore) SetCloseAt(_ time.Time) error {
	return nil
}

func (s *FakeStore) GetCloseAt() (time.Time, error) {
	return time.Now(), nil
}

func (s *FakeStore) SetIsOpen(_ bool) error {
	return nil
}

func (s *FakeStore) GetIsOpen() (bool, error) {
	return false, nil
}

func (s *FakeStore) SetTodayBeer(_ string) error {
	return nil
}

func (s *FakeStore) ResetTodayBeer() error {
	return nil
}

func (s *FakeStore) GetTodayBeer() (string, error) {
	return "", nil
}

func (s *FakeStore) AddConversationMessage(_ string, _ ConservationMessage) error {
	return nil
}

func (s *FakeStore) GetConversation(_ string) ([]ConservationMessage, error) {
	return nil, nil
}

func (s *FakeStore) ResetConversation(_ string) error {
	return nil
}

func (s *FakeStore) SetAttendanceKnownDevices(_ map[string]string) error {
	return nil
}

func (s *FakeStore) GetAttendanceKnownDevices() (map[string]string, error) {
	return map[string]string{}, nil
}

func (s *FakeStore) SetAttendanceIrks(_ map[string]string) error {
	return nil
}

func (s *FakeStore) GetAttendanceIrks() (map[string]string, error) {
	return map[string]string{}, nil
}
