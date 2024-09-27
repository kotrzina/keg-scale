package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
		if got := strip(test.input); got != test.expected {
			t.Errorf("strip(%q) = %q; want %q", test.input, got, test.expected)
		}
	}
}

func TestGetOkJson(t *testing.T) {
	got := getOkJson()
	assert.Contains(t, string(got), "ok")
}
