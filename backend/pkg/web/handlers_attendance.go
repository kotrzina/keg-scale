package web

import (
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kotrzina/keg-scale/pkg/scale"
)

func (hr *HandlerRepository) attendanceHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != hr.config.AuthToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		type AttendanceRequest struct {
			Ble []struct {
				Address string `json:"address"`
				Rssi    int    `json:"rssi"`
			} `json:"ble"`
			Telemetry struct {
				UptimeS      int `json:"uptime_s"`
				ScanCount    int `json:"scan_count"`
				CpuMhz       int `json:"cpu_mhz"`
				HeapSize     int `json:"heap_size"`
				FreeHeap     int `json:"free_heap"`
				MinFreeHeap  int `json:"min_free_heap"`
				WifiRssi     int `json:"wifi_rssi"`
				DevicesFound int `json:"devices_found"`
			} `json:"telemetry"`
		}

		var req AttendanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var devices map[string]scale.Device

		irks := hr.scale.GetIrks()

		for _, dev := range req.Ble {
			addr := dev.Address
			bounded := false
			irk, found := resolveRPA(dev.Address, irks)
			if found {
				addr = irk.IdentityAddress // rewrite current RPA by bond address
				bounded = true
			}
			devices[addr] = scale.Device{
				IdentityAddress: addr,
				RSSI:            dev.Rssi,
				Bounded:         bounded,
				LastSeen:        time.Now(),
			}
		}

		hr.monitor.AttendanceUptime.WithLabelValues().Set(float64(req.Telemetry.UptimeS))
		hr.monitor.AttendanceLastPing.WithLabelValues().SetToCurrentTime()
		hr.monitor.AttendanceScanCount.WithLabelValues().Set(float64(req.Telemetry.ScanCount))
		hr.monitor.AttendanceCpuMhz.WithLabelValues().Set(float64(req.Telemetry.CpuMhz))
		hr.monitor.AttendanceHeapSize.WithLabelValues().Set(float64(req.Telemetry.HeapSize))
		hr.monitor.AttendanceFreeHeap.WithLabelValues().Set(float64(req.Telemetry.FreeHeap))
		hr.monitor.AttendanceMinFreeHeap.WithLabelValues().Set(float64(req.Telemetry.MinFreeHeap))
		hr.monitor.AttendanceWifiRssi.WithLabelValues().Set(float64(req.Telemetry.WifiRssi))
		hr.monitor.AttendanceDetectedCount.WithLabelValues().Set(float64(len(devices)))
		hr.monitor.AttendanceIrkCount.WithLabelValues().Set(float64(len(irks)))

		hr.scale.SetDevices(devices)

		w.WriteHeader(http.StatusNoContent)
	}
}

func (hr *HandlerRepository) attendanceIrksHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != hr.config.AuthToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		type IRKUploadRequest struct {
			IdentityAddress string `json:"identity_address"`
			IRK             string `json:"irk"`
			DeviceName      string `json:"device_name,omitempty"`
			RSSI            *int   `json:"rssi,omitempty"`
			Appearance      *int   `json:"appearance,omitempty"`
		}

		var req IRKUploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := hr.scale.AddIrk(scale.Irk{
			IdentityAddress: req.IdentityAddress,
			Irk:             req.IRK,
			DeviceName:      req.DeviceName,
		}); err != nil {
			http.Error(w, "Could not add IRK", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func (hr *HandlerRepository) attendanceDeviceRenameHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type RenameRequest struct {
			IdentityAddress string `json:"identity_address"`
			DeviceName      string `json:"device_name"`
		}

		if r.Method != http.MethodPut {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}

		auth := r.Header.Get("Authorization")
		if auth != hr.config.Password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req RenameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		hr.logger.Infof("Renaming device %s to %s", req.IdentityAddress, req.DeviceName)

		if err := hr.scale.RenameKnownDevice(req.IdentityAddress, req.DeviceName); err != nil {
			http.Error(w, "Could not rename device", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func resolveRPA(rpa string, irks []scale.Irk) (scale.Irk, bool) {
	for _, irk := range irks {
		if ok, _ := matchRPA(irk.Irk, rpa); ok {
			return irk, true
		}
	}

	return scale.Irk{}, false
}

func matchRPA(irkHex string, rpaHex string) (bool, error) {
	irk, err := parseHexBytes(irkHex, 16)
	if err != nil {
		return false, err
	}
	rpa, err := parseHexBytes(rpaHex, 6)
	if err != nil {
		return false, err
	}

	// Reverse IRK (BLE uses little-endian)
	reverseBytes(irk)

	// After reversing: rpa[0:3] is prand, rpa[3:6] is hash
	prand := rpa[0:3]
	hashPart := rpa[3:6]

	// AES input = 13*0x00 + prand (3 bytes)
	plaintext := make([]byte, 16)
	copy(plaintext[13:], prand)

	block, err := aes.NewCipher(irk)
	if err != nil {
		return false, err
	}

	out := make([]byte, 16)
	block.Encrypt(out, plaintext)

	// Compare lowest 3 bytes of AES output with hash
	return out[13] == hashPart[0] &&
		out[14] == hashPart[1] &&
		out[15] == hashPart[2], nil
}

func reverseBytes(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

// parseHexBytes parses a hex string like "0011aabb" or "00:11:aa:bb"
// and enforces an exact byte length.
func parseHexBytes(s string, wantLen int) ([]byte, error) {
	clean := strings.ReplaceAll(s, ":", "")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ToLower(clean)

	b, err := hex.DecodeString(clean)
	if err != nil {
		return nil, err
	}
	if len(b) != wantLen {
		return nil, fmt.Errorf("invalid length, want %d, got %d", wantLen, len(b))
	}

	return b, nil
}
