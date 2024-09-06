package main

import "github.com/prometheus/client_golang/prometheus"

// Monitor represents a Prometheus monitor
// It contains Prometheus registry and all available metrics
type Monitor struct {
	Registry *prometheus.Registry

	kegWeight     *prometheus.GaugeVec
	scaleWifiRssi *prometheus.GaugeVec
	lastUpdate    *prometheus.GaugeVec
	pubIsOpen     *prometheus.GaugeVec
}

// NewMonitor creates a new Monitor
func NewMonitor() *Monitor {
	reg := prometheus.NewRegistry()
	monitor := &Monitor{
		Registry: reg,

		kegWeight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_keg_weight",
			Help: "Current weight of the keg in grams",
		}, []string{}),

		scaleWifiRssi: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_wifi_rssi",
			Help: "Current WiFi RSSI",
		}, []string{}),

		lastUpdate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_last_update",
			Help: "Last update time",
		}, []string{}),

		pubIsOpen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_pub_open",
			Help: "Is the pub open/closed",
		}, []string{}),
	}

	reg.MustRegister(monitor.lastUpdate)
	reg.MustRegister(monitor.kegWeight)
	reg.MustRegister(monitor.scaleWifiRssi)
	reg.MustRegister(monitor.pubIsOpen)

	return monitor
}
