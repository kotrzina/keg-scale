package scale

import (
	"fmt"
	"strings"
)

// GetPushResponse is a response for scale push event
// Scale has display and it is able to display four digits
func (s *Scale) GetPushResponse() string {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return leftPad(fmt.Sprintf("%d", s.beersLeft), " ", 4)
}

func leftPad(input, padChar string, length int) string {
	if len(input) >= length {
		return input
	}
	padding := strings.Repeat(padChar, length-len(input))
	return padding + input
}
