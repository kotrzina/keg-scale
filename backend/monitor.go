package main

import "github.com/prometheus/client_golang/prometheus"

// Monitor represents a Prometheus monitor
// It contains Prometheus registry and all available metrics
type Monitor struct {
	Registry  *prometheus.Registry
	kegWeight *prometheus.GaugeVec
}

// NewMonitor creates a new Monitor
func NewMonitor() *Monitor {
	reg := prometheus.NewRegistry()
	monitor := &Monitor{
		Registry: reg,

		kegWeight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "keg_weight",
			Help: "Current weight of the keg in grams",
		}, []string{}),
	}

	reg.MustRegister(monitor.kegWeight)

	return monitor
}
