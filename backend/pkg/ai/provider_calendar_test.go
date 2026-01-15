package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatCalendarDate(t *testing.T) {
	tests := []struct {
		name     string
		date     *time.Time
		expected string
	}{
		{
			name:     "nil date",
			date:     nil,
			expected: "unknown",
		},
		{
			name:     "valid date",
			date:     timePtr(time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)),
			expected: "2025-06-15",
		},
		{
			name:     "date with time ignored",
			date:     timePtr(time.Date(2024, 12, 25, 23, 59, 59, 0, time.UTC)),
			expected: "2024-12-25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCalendarDate(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
