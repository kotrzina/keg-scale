package prometheus

import (
	"testing"
)

func TestNewMonitor(t *testing.T) {
	monitor := New()

	if monitor.Registry == nil {
		t.Fatal("Registry should not be nil")
	}

	tests := []struct {
		name   string
		metric interface{}
	}{
		{"Weight", monitor.Weight},
		{"ActiveKeg", monitor.ActiveKeg},
		{"BeersLeft", monitor.BeersLeft},
		{"BeersTotal", monitor.BeersTotal},
		{"ScaleWifiRssi", monitor.ScaleWifiRssi},
		{"LastPing", monitor.LastPing},
		{"PubIsOpen", monitor.PubIsOpen},
		{"AttendanceUptime", monitor.AttendanceUptime},
		{"AttendanceLastPing", monitor.AttendanceLastPing},
		{"AttendanceScanCount", monitor.AttendanceScanCount},
		{"AttendanceCpuMhz", monitor.AttendanceCpuMhz},
		{"AttendanceHeapSize", monitor.AttendanceHeapSize},
		{"AttendanceFreeHeap", monitor.AttendanceFreeHeap},
		{"AttendanceMinFreeHeap", monitor.AttendanceMinFreeHeap},
		{"AttendanceWifiRssi", monitor.AttendanceWifiRssi},
		{"AttendanceDetectedCount", monitor.AttendanceDetectedCount},
		{"AttendanceIrkCount", monitor.AttendanceIrkCount},
		{"AnthropicInputTokens", monitor.AnthropicInputTokens},
		{"AnthropicOutputTokens", monitor.AnthropicOutputTokens},
		{"OpenAiInputTokens", monitor.OpenAiInputTokens},
		{"OpenAiOutputTokens", monitor.OpenAiOutputTokens},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metric == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
		})
	}
}
