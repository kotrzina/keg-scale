package scale

import (
	"fmt"
	"time"
)

type EventType string

const (
	EventOpen         EventType = "pub_open"
	EventClose        EventType = "pub_close"
	EventNewKegTapped EventType = "new_keg_tapped"
)

func (s *Scale) makeEvent(event EventType) {
	go func() {
		eventString := fmt.Sprintf("%s AT %s", event, time.Now().Format(time.RFC3339))
		err := s.store.AddEvent(eventString)
		if err != nil {
			s.logger.Error("failed to add event", "event", event, "error", err)
		}
	}()
}
