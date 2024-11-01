package scale

import "time"

type OpeningOutput struct {
	IsOpen   bool      `json:"is_open"`
	OpenedAt time.Time `json:"open_at"`
	ClosedAt time.Time `json:"closed_at"`
}

func (s *Scale) GetOpeningOutput() OpeningOutput {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return OpeningOutput{
		IsOpen:   s.pub.isOpen,
		OpenedAt: s.pub.openedAt,
		ClosedAt: s.pub.closedAt,
	}
}
