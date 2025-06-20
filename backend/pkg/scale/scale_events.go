package scale

import (
	"fmt"
	"time"
)

type EventType string

type Event func(et EventType) error

const (
	EventOpen         EventType = "pub_open"
	EventClose        EventType = "pub_close"
	EventNewKegTapped EventType = "new_keg_tapped"
)

// RegisterEvent registers a callback for a specific event
// The function checks if the event type is already registered and appends the callback to the list of callbacks
func (s *Scale) RegisterEvent(eventType EventType, callback Event) {
	if s.events[eventType] == nil {
		s.events[eventType] = []Event{callback}
	} else {
		s.events[eventType] = append(s.events[eventType], callback)
	}
}

func (s *Scale) dispatchEvent(event EventType) {
	go func() {
		// log event to storage
		eventString := fmt.Sprintf("%s AT %s", event, time.Now().Format(time.RFC3339))
		err := s.store.AddEvent(eventString)
		if err != nil {
			s.logger.Error("failed to add event", "event", event, "error", err)
		}

		// actually dispatch events
		if s.events[event] != nil {
			for _, hook := range s.events[event] {
				if err = hook(event); err != nil {
					s.logger.Errorf("failed to run hook for event %s: %s", event, err)
				}
			}
		}
	}()
}
