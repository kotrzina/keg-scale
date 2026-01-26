package web

import (
	"encoding/json"
	"net/http"

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
			Bounded []struct {
				Address string `json:"address"`
				Rssi    int    `json:"rssi"`
			} `json:"bounded"`
			Telemetry struct {
				UptimeS      int `json:"uptime_s"`
				ScanCount    int `json:"scan_count"`
				CpuMhz       int `json:"cpu_mhz"`
				HeapSize     int `json:"heap_size"`
				FreeHeap     int `json:"free_heap"`
				MinFreeHeap  int `json:"min_free_heap"`
				WifiRssi     int `json:"wifi_rssi"`
				IrkCount     int `json:"irk_count"`
				DevicesFound int `json:"devices_found"`
			} `json:"telemetry"`
		}

		var req AttendanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		devices := make([]scale.Device, len(req.Bounded)+len(req.Ble))
		i := 0

		for _, dev := range req.Bounded {
			devices[i] = scale.Device{
				IdentityAddress: dev.Address,
				RSSI:            dev.Rssi,
				Bounded:         true,
			}
			i++
		}

		for _, dev := range req.Ble {
			devices[i] = scale.Device{
				IdentityAddress: dev.Address,
				RSSI:            dev.Rssi,
				Bounded:         false,
			}
			i++
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
		hr.monitor.AttendanceIrkCount.WithLabelValues().Set(float64(req.Telemetry.IrkCount))

		hr.scale.SetDevices(devices)

		w.WriteHeader(http.StatusNoContent)
	}
}

func (hr *HandlerRepository) attendanceIrksHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if r.Method == http.MethodPost {
			var req IRKUploadRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			hr.logger.Infof("IRK upload: %s %s %s %d %d", req.IdentityAddress, req.IRK, req.DeviceName, *req.RSSI, *req.Appearance)

			hr.scale.AddIrk(scale.Irk{
				IdentityAddress: req.IdentityAddress,
				Irk:             req.IRK,
				DeviceName:      req.DeviceName,
			})

			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == http.MethodGet {
			irks := hr.scale.GetIrks()
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(irks); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}

			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
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
