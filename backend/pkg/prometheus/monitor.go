package prometheus

import "github.com/prometheus/client_golang/prometheus"

// Monitor represents a Prometheus monitor
// It contains Prometheus registry and all available metrics
type Monitor struct {
	Registry *prometheus.Registry

	Weight        *prometheus.GaugeVec
	ActiveKeg     *prometheus.GaugeVec
	BeersLeft     *prometheus.GaugeVec
	ScaleWifiRssi *prometheus.GaugeVec
	LastPing      *prometheus.GaugeVec
	PubIsOpen     *prometheus.GaugeVec
}

// NewMonitor creates a new Monitor
func NewMonitor() *Monitor {
	reg := prometheus.NewRegistry()
	monitor := &Monitor{
		Registry: reg,

		Weight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_weight",
			Help: "Current weight of the keg in grams",
		}, []string{}),

		ActiveKeg: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_active_keg",
			Help: "Size of current keg in liters",
		}, []string{}),

		BeersLeft: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_beers_left",
			Help: "How to beers are left in the current keg",
		}, []string{}),

		ScaleWifiRssi: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_wifi_rssi",
			Help: "Current WiFi RSSI",
		}, []string{}),

		LastPing: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_last_ping",
			Help: "Last update time",
		}, []string{}),

		PubIsOpen: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_pub_open",
			Help: "Is the pub open/closed",
		}, []string{}),
	}

	reg.MustRegister(monitor.Weight)
	reg.MustRegister(monitor.ActiveKeg)
	reg.MustRegister(monitor.BeersLeft)
	reg.MustRegister(monitor.ScaleWifiRssi)
	reg.MustRegister(monitor.LastPing)
	reg.MustRegister(monitor.PubIsOpen)

	return monitor
}
