package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/kotrzina/keg-scale/pkg/ai"
	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/promector"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type HandlerRepository struct {
	scale     *scale.Scale
	promector *promector.Promector
	ai        *ai.Ai
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
				"message_id": message.MessageID,
			}).Infof("Scale new value: %0.2f", message.Value)
		}

		_, err = w.Write([]byte(hr.scale.GetPushResponse()))
		if err != nil {
			hr.logger.Errorf("Could not write response: %v", err)
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
		_, err = w.Write(utils.GetOkJSON())
		if err != nil {
			hr.logger.Errorf("Could not write response: %v", err)
		}
	}
}

func (hr *HandlerRepository) scaleDashboardHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
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
		_, err = w.Write(res)
		if err != nil {
			hr.logger.Errorf("Could not write response: %v", err)
		}
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

		if strings.EqualFold(data.Way, "up") {
			if err := hr.scale.IncreaseWarehouse(data.Keg); err != nil {
				http.Error(w, "Could not increase warehouse", http.StatusInternalServerError)
				return
			}
		}

		if strings.EqualFold(data.Way, "down") {
			if err := hr.scale.DecreaseWarehouse(data.Keg); err != nil {
				http.Error(w, "Could not increase warehouse", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(utils.GetOkJSON())
		if err != nil {
			hr.logger.Errorf("Could not write response: %v", err)
		}
	}
}

func (hr *HandlerRepository) aiTestHandler() func(http.ResponseWriter, *http.Request) {
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

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Could not read post body", http.StatusInternalServerError)
			return
		}

		var data []ai.ChatMessage
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Could not unmarshal data", http.StatusBadRequest)
			return
		}

		resp, err := hr.ai.GetResponse(data)
		if err != nil {
			hr.logger.Warnf("could not get response because %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.Text = strings.ReplaceAll(resp.Text, "\n", "<br/>")

		output, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Could not marshal data", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(output)
		if err != nil {
			hr.logger.Errorf("Could not write response: %v", err)
		}
	}
}

func (hr *HandlerRepository) checkPassword() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != hr.config.Password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
