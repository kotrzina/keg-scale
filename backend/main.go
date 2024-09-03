package main

import (
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"log"
)

func main() {
	// for development purposes
	// we don't care about errors here
	_ = godotenv.Load(".env")
	config := NewConfig()

	promector := NewPromector(config.PrometheusUrl, config.PrometheusUser, config.PrometheusPassword)
	scaleCurrentValue, err := promector.GetLastValue("scale_keg_weight")
	if err != nil {
		log.Fatalf("Error getting last value: %v", err)
	}
	scaleLastUpdate, err := promector.GetLastValue("scale_last_update")
	if err != nil {
		log.Fatalf("Error getting last value: %v", err)
	}

	monitor := NewMonitor()
	monitor.kegWeight.With(prometheus.Labels{}).Set(scaleCurrentValue)
	monitor.lastUpdate.With(prometheus.Labels{}).Set(scaleLastUpdate)

	scale := NewScale(config.BufferSize)
	StartServer(NewRouter(&HandlerRepository{
		scale:   scale,
		config:  config,
		monitor: monitor,
	}), 8080)
}
