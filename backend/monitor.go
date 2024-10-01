package main

import "github.com/prometheus/client_golang/prometheus"

// Monitor represents a Prometheus monitor
// It contains Prometheus registry and all available metrics
type Monitor struct {
	Registry *prometheus.Registry

	weight        *prometheus.GaugeVec
	activeKeg     *prometheus.GaugeVec
	beersLeft     *prometheus.GaugeVec
	scaleWifiRssi *prometheus.GaugeVec
	lastPing      *prometheus.GaugeVec
	pubIsOpen     *prometheus.GaugeVec
}

// NewMonitor creates a new Monitor
func NewMonitor() *Monitor {
	reg := prometheus.NewRegistry()
	monitor := &Monitor{
		Registry: reg,

		weight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_weight",
			Help: "Current weight of the keg in grams",
		}, []string{}),

		activeKeg: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_active_keg",
			Help: "Size of current keg in liters",
		}, []string{}),

		beersLeft: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_beers_left",
			Help: "How to beers are left in the current keg",
		}, []string{}),

		scaleWifiRssi: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_wifi_rssi",
			Help: "Current WiFi RSSI",
		}, []string{}),

		lastPing: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_last_ping",
			Help: "Last update time",
		}, []string{}),

		pubIsOpen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_pub_open",
			Help: "Is the pub open/closed",
		}, []string{}),
	}

	reg.MustRegister(monitor.weight)
	reg.MustRegister(monitor.activeKeg)
	reg.MustRegister(monitor.beersLeft)
	reg.MustRegister(monitor.scaleWifiRssi)
	reg.MustRegister(monitor.lastPing)
	reg.MustRegister(monitor.pubIsOpen)

	return monitor
}
