package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/apognu/gocal"
)

func ProvideCalendar(url string, start, end time.Time) (string, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not get calendar: %w", err)
	}
	if resp.Body == nil {
		defer resp.Body.Close() //nolint: errcheck
	}

	calendar := gocal.NewParser(resp.Body)
	calendar.Start, calendar.End = &start, &end

	err = calendar.Parse()
	if err != nil {
		return "", fmt.Errorf("could not parse calendar: %w", err)
	}

	type Event struct {
		Summary     string `json:"summary"`
		Start       string `json:"start"`
		End         string `json:"end"`
		Description string `json:"description"`
		Location    string `json:"location"`
	}

	events := make([]Event, len(calendar.Events))
	for i, e := range calendar.Events {
		events[i] = Event{
			Summary:     e.Summary,
			Start:       formatCalendarDate(e.Start),
			End:         formatCalendarDate(e.End),
			Description: e.Description,
			Location:    e.Location,
		}
	}

	jsonData, err := json.Marshal(events)
	if err != nil {
		return "", fmt.Errorf("could not marshal calendar events: %w", err)
	}

	return fmt.Sprintf("Calendar contains following events:\n<json>\n%s\n</json>", string(jsonData)), nil
}

func formatCalendarDate(date *time.Time) string {
	if date == nil {
		return "unknown"
	}
	return date.Format("2006-01-02")
}
