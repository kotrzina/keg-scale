package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Příliš žluťoučký kůň", "prilis zlutoucky kun"},
		{"ÁÉÍÓÚ", "aeiou"},
		{"Hello World", "hello world"},
		{"Čeština", "cestina"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterSignificantWords(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "filter stop words",
			input:    []string{"hasicska", "a", "zbrojnice"},
			expected: []string{"hasicska", "zbrojnice"},
		},
		{
			name:     "filter short words",
			input:    []string{"on", "je", "hasic"},
			expected: []string{"hasic"},
		},
		{
			name:     "keep significant words",
			input:    []string{"oprava", "budovy", "skoly"},
			expected: []string{"oprava", "budovy", "skoly"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "all filtered",
			input:    []string{"a", "je", "na", "to"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterSignificantWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateRecencyBonus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		date     string
		expected int
	}{
		{
			name:     "last week",
			date:     now.AddDate(0, 0, -7).Format("2006-01-02"),
			expected: 50,
		},
		{
			name:     "last month",
			date:     now.AddDate(0, 0, -25).Format("2006-01-02"),
			expected: 50,
		},
		{
			name:     "two months ago",
			date:     now.AddDate(0, -2, 0).Format("2006-01-02"),
			expected: 30,
		},
		{
			name:     "five months ago",
			date:     now.AddDate(0, -5, 0).Format("2006-01-02"),
			expected: 15,
		},
		{
			name:     "ten months ago",
			date:     now.AddDate(0, -10, 0).Format("2006-01-02"),
			expected: 5,
		},
		{
			name:     "two years ago",
			date:     now.AddDate(-2, 0, 0).Format("2006-01-02"),
			expected: 0,
		},
		{
			name:     "invalid date",
			date:     "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateRecencyBonus(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}
