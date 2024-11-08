package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrip(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello, World!", ""},
		{"Hello, 123!", "123"},
		{"Hello, 123.74! hello", "123.74"},
		{"Hello, 0.12", "0.12"},
	}

	for _, test := range tests {
		if got := Strip(test.input); got != test.expected {
			t.Errorf("Strip(%q) = %q; want %q", test.input, got, test.expected)
		}
	}
}

func TestGetOkJson(t *testing.T) {
	got := GetOkJSON()
	assert.Contains(t, string(got), "ok")
}
