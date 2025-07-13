package prometheus

import "github.com/prometheus/client_golang/prometheus"

// Monitor represents a Prometheus monitor
// It contains Prometheus registry and all available metrics
type Monitor struct {
	Registry *prometheus.Registry

	Weight        *prometheus.GaugeVec
	ActiveKeg     *prometheus.GaugeVec
	BeersLeft     *prometheus.GaugeVec
	BeersTotal    *prometheus.GaugeVec
	ScaleWifiRssi *prometheus.GaugeVec
	LastPing      *prometheus.GaugeVec
	PubIsOpen     *prometheus.GaugeVec

	AnthropicInputTokens  *prometheus.CounterVec
	AnthropicOutputTokens *prometheus.CounterVec
	OpenAiInputTokens     *prometheus.CounterVec
	OpenAiOutputTokens    *prometheus.CounterVec
}

// New creates a new Monitor
func New() *Monitor {
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

		BeersTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "scale_beers_total",
			Help: "Total number of beers consumed",
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

		AnthropicInputTokens: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "anthropic_input_tokens_total",
			Help: "Number of input tokens processed by the AI",
		}, []string{}),

		AnthropicOutputTokens: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "anthropic_output_tokens_total",
			Help: "Number of output tokens processed by the AI",
		}, []string{}),

		OpenAiInputTokens: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "openai_input_tokens_total",
			Help: "Number of input tokens processed by the Open AI",
		}, []string{}),

		OpenAiOutputTokens: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "openai_output_tokens_total",
			Help: "Number of output tokens processed by the Open AI",
		}, []string{}),
	}

	reg.MustRegister(
		monitor.Weight,
		monitor.ActiveKeg,
		monitor.BeersLeft,
		monitor.BeersTotal,
		monitor.ScaleWifiRssi,
		monitor.LastPing,
		monitor.PubIsOpen,
		monitor.AnthropicInputTokens,
		monitor.AnthropicOutputTokens,
		monitor.OpenAiInputTokens,
		monitor.OpenAiOutputTokens,
	)

	return monitor
}
