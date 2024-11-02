package prometheus

import (
	"testing"
)

func TestNewMonitor(t *testing.T) {
	monitor := NewMonitor()

	if monitor.Registry == nil {
		t.Error("Registry should not be nil")
	}

	if monitor.Weight == nil {
		t.Error("Weight should not be nil")
	}

	if monitor.ActiveKeg == nil {
		t.Error("ActiveKeg should not be nil")
	}

	if monitor.BeersLeft == nil {
		t.Error("BeersLeft should not be nil")
	}

	if monitor.ScaleWifiRssi == nil {
		t.Error("ScaleWifiRssi should not be nil")
	}

	if monitor.LastPing == nil {
		t.Error("LastPing should not be nil")
	}

	if monitor.PubIsOpen == nil {
		t.Error("PubIsOpen should not be nil")
	}
}
