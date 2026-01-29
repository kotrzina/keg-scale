package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveRPA(t *testing.T) {
	tests := []struct {
		name   string
		rpa    string
		irk    string
		result bool
	}{
		{
			name:   "android_1",
			rpa:    "68:70:55:67:5F:E5",                // real bonding RPA from device
			irk:    "c98f3d364c091847a1cdf7c058686524", // real irk from my testing android phone
			result: true,
		},
		{
			name:   "test_1",
			rpa:    "49:C3:3F:F4:F3:6D",
			irk:    "1b:c8:b0:4a:ce:8f:88:b2:03:3f:e2:90:5b:3c:22:d5",
			result: true,
		},
		{
			name:   "test_2",
			rpa:    "73:9E:92:B7:01:81",
			irk:    "1b:c8:b0:4a:ce:8f:88:b2:03:3f:e2:90:5b:3c:22:d5",
			result: true,
		},
		{
			name:   "test_3",
			rpa:    "76:39:F4:FA:58:59",
			irk:    "1b:c8:b0:4a:ce:8f:88:b2:03:3f:e2:90:5b:3c:22:d5",
			result: true,
		},
		{
			name:   "invalid_rpa",
			rpa:    "AA:BB:CC:DD:EE:FF",
			irk:    "1c:c8:b0:4a:ce:8f:88:b2:03:3f:e2:90:5b:3c:22:d5",
			result: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matchRPA(tt.irk, tt.rpa)
			assert.NoError(t, err)
			assert.Equal(t, tt.result, result)
		})
	}
}
