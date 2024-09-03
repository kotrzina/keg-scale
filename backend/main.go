package main

import "github.com/prometheus/client_golang/prometheus"

func main() {
	monitor := NewMonitor()
	monitor.kegWeight.With(prometheus.Labels{}).Set(1000) // @todo: get value from prometheus

	config := NewConfig()
	scale := NewScale(config.BufferSize)

	StartServer(NewRouter(&HandlerRepository{
		scale:   scale,
		config:  config,
		monitor: monitor,
	}), 8080)
}
