package ai

import (
	"testing"

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
