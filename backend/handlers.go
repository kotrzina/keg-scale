package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"strconv"
)

type HandlerRepository struct {
	scale   *Scale
	config  *Config
	monitor *Monitor
}

func (hr *HandlerRepository) scaleStatusHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		data, err := hr.scale.JsonState()

		if err != nil {
			http.Error(w, "Could not marshal state to JSON", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}
}

func (hr *HandlerRepository) scalePingHandler() func(http.ResponseWriter, *http.Request) {
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

		log.Printf("Ping received")

		hr.scale.Ping()
		hr.monitor.lastUpdate.WithLabelValues().SetToCurrentTime()

		_, _ = w.Write([]byte("OK"))
	}
}

func (hr *HandlerRepository) scaleValueHandler() func(http.ResponseWriter, *http.Request) {
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

		body, err := io.ReadAll(r.Body)
		fmt.Printf("body: %q\n", body)
		if err != nil {
			http.Error(w, "Could not read post body", http.StatusInternalServerError)
			return
		}

		current, err := strconv.ParseFloat(string(body), 64)
		if err != nil {
			http.Error(w, "Could not convert value to number", http.StatusBadRequest)
			return
		}

		log.Printf("Keg weight: %f", current)

		hr.monitor.kegWeight.WithLabelValues().Set(current)
		hr.monitor.lastUpdate.WithLabelValues().SetToCurrentTime()

		hr.scale.AddMeasurement(current)
		hr.scale.Ping() // we can also ping with the new value

		_, _ = w.Write([]byte("OK"))
	}
}

func (hr *HandlerRepository) scalePrintHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		c := hr.scale.GetValidCount()
		for i := 0; i < c; i++ {
			m := hr.scale.GetMeasurement(i)
			_, _ = w.Write([]byte(fmt.Sprintf("%.2f;", m.Weight)))
		}
	}
}

// metricsHandler returns HTTP handler for metrics endpoint
func (hr *HandlerRepository) metricsHandler() http.Handler {
	return promhttp.HandlerFor(
		hr.monitor.Registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          hr.monitor.Registry,
		},
	)
}

// homepageHandler returns HTTP handler for homepage
func (hr *HandlerRepository) homepageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>Keg scale exporter</title></head>
			<body>
			<h1>Keg scale exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})
}
