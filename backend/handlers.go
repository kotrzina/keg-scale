package main

import (
	"encoding/json"
	"fmt"
	"github.com/hako/durafmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type HandlerRepository struct {
	scale   *Scale
	config  *Config
	monitor *Monitor
	logger  *logrus.Logger
}

func (hr *HandlerRepository) scaleStatusHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		hr.logger.Info("Scale status requested")

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
			hr.scale.AddMeasurement(message.Value)

			hr.logger.WithFields(logrus.Fields{
				"message_id": message.MessageId,
			}).Infof("Scale new value: %0.2f", message.Value)
		}

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

const localizationUnits = "r:r,t:t,d:d,h:h,m:m,s:s,ms:ms,microsecond"

func (hr *HandlerRepository) scaleDashboardHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		hr.scale.Recheck()

		type pubOutput struct {
			IsOpen   bool   `json:"is_open"`
			OpenedAt string `json:"opened_at"`
			ClosedAt string `json:"closed_at"`
		}

		type output struct {
			IsOk               bool      `json:"is_ok"`
			LastWeight         float64   `json:"last_weight"`
			LastWeightFormated string    `json:"last_weight_formated"`
			LastAt             string    `json:"last_at"`
			LastAtDuration     string    `json:"last_at_duration"`
			Rssi               float64   `json:"rssi"`
			LastUpdate         string    `json:"last_update"`
			LastUpdateDuration string    `json:"last_update_duration"`
			Pub                pubOutput `json:"pub"`
		}

		if !hr.scale.HasLastN(1) {
			// we don't have any measurements yet
			http.Error(w, "No measurements yet", http.StatusTooEarly)
			return
		}

		last := hr.scale.GetLastMeasurement()

		units, err := durafmt.DefaultUnitsCoder.Decode(localizationUnits)
		if err != nil {
			http.Error(w, "Could not decode units", http.StatusInternalServerError)
			return
		}

		data := output{
			IsOk:               hr.scale.IsOk(),
			LastWeight:         last.Weight,
			LastWeightFormated: fmt.Sprintf("%.2f", last.Weight/1000),
			LastAt:             last.At.Format("2006-01-02 15:04:05"),
			LastAtDuration:     durafmt.Parse(time.Since(last.At).Round(time.Second)).LimitFirstN(2).Format(units),
			Rssi:               hr.scale.Rssi,
			LastUpdate:         hr.scale.LastOk.Format("2006-01-02 15:04:05"),
			LastUpdateDuration: durafmt.Parse(time.Since(hr.scale.LastOk).Round(time.Second)).LimitFirstN(2).Format(units),
			Pub: pubOutput{
				IsOpen:   hr.scale.Pub.IsOpen,
				OpenedAt: hr.scale.Pub.OpenedAt.Format("15:04"),
				ClosedAt: hr.scale.Pub.ClosedAt.Format("15:04"),
			},
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
