package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScale_ParseScaleMessage(t *testing.T) {
	type testcases struct {
		raw    string
		parsed ScaleMessage
	}

	tests := []testcases{
		{"push|2887417|-74.7|1923.23", ScaleMessage{"push", 2887417, -74.7, 1923.23}},
		{"push|2887417|-74.7|1923.23|", ScaleMessage{"push", 2887417, -74.7, 1923.23}}, // extra pipe
		{"ping|2887417|-74.7|", ScaleMessage{"ping", 2887417, -74.7, 0}},
		{"ping|2887417|-74.7||", ScaleMessage{"ping", 2887417, -74.7, 0}},   // extra pipe
		{"push|471|-74.7|-47.25", ScaleMessage{"push", 471, -74.7, -47.25}}, // negative value
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			parsed, err := ParseScaleMessage(test.raw)
			require.NoError(t, err)

			if test.parsed.MessageID != parsed.MessageID {
				t.Errorf("Expected MessageID to be %d, got %d", test.parsed.MessageID, parsed.MessageID)
			}

			if test.parsed.MessageType != parsed.MessageType {
				t.Errorf("Expected MessageType to be %s, got %s", test.parsed.MessageType, parsed.MessageType)
			}

			if test.parsed.Rssi != parsed.Rssi {
				t.Errorf("Expected rssi to be %f, got %f", test.parsed.Rssi, parsed.Rssi)
			}

			if test.parsed.Value != parsed.Value {
				t.Errorf("Expected Value to be %f, got %f", test.parsed.Value, parsed.Value)
			}
		})
	}
}
