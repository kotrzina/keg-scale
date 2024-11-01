package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/promector"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/utils"
)

type HandlerRepository struct {
	scale     *scale.Scale
	promector *promector.Promector
	config    *config.Config
	monitor   *prometheus.Monitor
	logger    *logrus.Logger
}

func (hr *HandlerRepository) scaleMessageHandler() func(http.ResponseWriter, *http.Request) {
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
		if err != nil {
			http.Error(w, "Could not read post body", http.StatusInternalServerError)
			return
		}

		message, err := ParseScaleMessage(string(body))
		if err != nil {
			hr.logger.Warnf("Could not parse scale message: %s because %v", string(body), err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hr.scale.Ping()
		hr.scale.SetRssi(message.Rssi)

		if message.MessageType == PushMessageType {
			err = hr.scale.AddMeasurement(message.Value)
			if err != nil {
				hr.logger.Warnf("Could not create measurement: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			hr.logger.WithFields(logrus.Fields{
				"message_id": message.MessageId,
			}).Infof("Scale new value: %0.2f", message.Value)
		}

		_, _ = w.Write([]byte("OK"))
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

func (hr *HandlerRepository) activeKegHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != hr.config.Password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		type input struct {
			Keg int `json:"keg"`
		}

		var data input
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Could not read post body", http.StatusBadRequest)
			return
		}

		switch data.Keg {
		case 10, 15, 20, 30, 50:
			// all is well
		default:
			http.Error(w, "Invalid keg size", http.StatusBadRequest)
			return
		}

		if err = hr.scale.SetActiveKeg(data.Keg); err != nil {
			http.Error(w, "Could not set active keg", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(utils.GetOkJson())
	}
}

func (hr *HandlerRepository) scaleDashboardHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		hr.scale.Recheck()

		type output struct {
			Scale  scale.FullOutput `json:"scale"`
			Charts promector.Charts `json:"charts"`
		}

		data := output{
			Scale:  hr.scale.GetScale(),
			Charts: hr.promector.GetChartData(),
		}

		res, err := json.Marshal(data)

		if err != nil {
			http.Error(w, "Could not marshal data to JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(res)
	}
}

func (hr *HandlerRepository) scaleWarehouseHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != hr.config.Password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		type input struct {
			Keg int    `json:"keg"`
			Way string `json:"way"` // up or down
		}

		var data input
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Could not read post body", http.StatusBadRequest)
			return
		}

		if strings.ToLower(data.Way) == "up" {
			if err := hr.scale.IncreaseWarehouse(data.Keg); err != nil {
				http.Error(w, "Could not increase warehouse", http.StatusInternalServerError)
				return
			}
		}

		if strings.ToLower(data.Way) == "down" {
			if err := hr.scale.DecreaseWarehouse(data.Keg); err != nil {
				http.Error(w, "Could not increase warehouse", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(utils.GetOkJson())
	}
}
