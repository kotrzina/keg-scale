package scale

import (
	"fmt"
	"time"
)

type Irk struct {
	IdentityAddress string `json:"identity_address"`
	Irk             string `json:"irk"`
	DeviceName      string `json:"device_name"`
}

type Device struct {
	IdentityAddress string    `json:"identity_address"`
	RSSI            int       `json:"rssi"`
	Bounded         bool      `json:"bounded"`
	LastSeen        time.Time `json:"last_seen"`
}

type attendance struct {
	irks   []Irk
	active map[string]Device // list of active devices
	known  map[string]string // list of known devices -> translated names
	lastOk time.Time
}

func (s *Scale) AddIrk(irk Irk) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	// If IRK already exists, remove it
	for i, existingIrk := range s.attendance.irks {
		if existingIrk.IdentityAddress == irk.IdentityAddress {
			s.attendance.irks = append(s.attendance.irks[:i], s.attendance.irks[i+1:]...)
			break
		}
	}

	// Store device name (only if it's not already known)
	knownName, f := s.attendance.known[irk.IdentityAddress]
	if !f || knownName == "" {
		s.attendance.known[irk.IdentityAddress] = irk.DeviceName
	}

	// store IRK
	s.attendance.irks = append(s.attendance.irks, irk)

	if err := s.store.SetAttendanceKnownDevices(s.attendance.known); err != nil {
		s.logger.Errorf("failed to store known devices: %s => %s because %s", irk.IdentityAddress, irk.DeviceName, err)
		return fmt.Errorf("failed to store known devices: %w", err)
	}

	irks := make(map[string]string, len(s.attendance.irks))
	for _, item := range s.attendance.irks {
		irks[item.IdentityAddress] = item.Irk
	}
	if err := s.store.SetAttendanceIrks(irks); err != nil {
		s.logger.Errorf("failed to store irk: %s => %s because %s", irk.IdentityAddress, irk.Irk, err)
		return fmt.Errorf("failed to store irk: %w", err)
	}

	return nil
}

func (s *Scale) GetIrks() []Irk {
	s.mux.Lock()
	defer s.mux.Unlock()

	irks := make([]Irk, len(s.attendance.irks))
	copy(irks, s.attendance.irks)
	return irks
}

func (s *Scale) SetDevices(devices map[string]Device) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for address, device := range devices {
		s.attendance.active[address] = device
	}

	// delete inactive devices
	for address, device := range s.attendance.active {
		if device.LastSeen.Before(time.Now().Add(-15 * time.Minute)) {
			delete(s.attendance.active, address)
		}
	}

	s.attendance.lastOk = time.Now()
}

func (s *Scale) GetKnownDevices() map[string]string {
	s.mux.Lock()
	defer s.mux.Unlock()

	// create a copy of known devices
	r := make(map[string]string, len(s.attendance.known))
	for k, v := range s.attendance.known {
		r[k] = v
	}

	return r
}

func (s *Scale) RenameKnownDevice(address string, newName string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.attendance.known[address] = newName
	return s.store.SetAttendanceKnownDevices(s.attendance.known)
}
