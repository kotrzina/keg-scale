package scale

type Irk struct {
	IdentityAddress string `json:"identity_address"`
	Irk             string `json:"irk"`
	DeviceName      string `json:"device_name"`
}

type Device struct {
	IdentityAddress string `json:"identity_address"`
	RSSI            int    `json:"rssi"`
	Bounded         bool   `json:"bounded"`
}

type attendance struct {
	irks   []Irk
	active []Device          // list of active devices
	known  map[string]string // list of known devices -> translated names
}

func (s *Scale) AddIrk(irk Irk) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for i, existingIrk := range s.attendance.irks {
		if existingIrk.IdentityAddress == irk.IdentityAddress {
			s.attendance.irks = append(s.attendance.irks[:i], s.attendance.irks[i+1:]...)
			break
		}
	}

	s.attendance.known[irk.IdentityAddress] = irk.DeviceName
	s.attendance.irks = append(s.attendance.irks, irk)

	// it could fail, we don't really care
	_ = s.store.SetAttendanceKnownDevices(s.attendance.known)

	irks := make(map[string]string, len(s.attendance.irks))
	for _, item := range s.attendance.irks {
		irks[item.IdentityAddress] = item.Irk
	}
	if err := s.store.SetAttendanceIrks(irks); err != nil {
		s.logger.Errorf("failed to store irk: %s => %s because %s", irk.IdentityAddress, irk.Irk, err)
	}
}

func (s *Scale) GetIrks() []Irk {
	s.mux.Lock()
	defer s.mux.Unlock()

	irks := make([]Irk, len(s.attendance.irks))
	copy(irks, s.attendance.irks)
	return irks
}

func (s *Scale) SetDevices(devices []Device) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.attendance.active = devices
}

func (s *Scale) RenameKnownDevice(address string, newName string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.attendance.known[address] = newName
	return s.store.SetAttendanceKnownDevices(s.attendance.known)
}
