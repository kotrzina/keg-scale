package main

import "github.com/prometheus/client_golang/prometheus"

func main() {
	monitor := NewMonitor()
	monitor.kegWeight.With(prometheus.Labels{}).Set(1000)

	config := NewConfig()
	scale := NewScale(MEASUREMENTS)

	StartServer(NewRouter(&HandlerRepository{
		scale:   scale,
		config:  config,
		monitor: monitor,
	}), 8080)
}
