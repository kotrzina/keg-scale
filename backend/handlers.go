package main

import (
	"encoding/json"
	"fmt"
	"github.com/hako/durafmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
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
		_, _ = w.Write(getOkJson())
	}
}

const localizationUnits = "r:r,t:t,d:d,h:h,m:m,s:s,ms:ms,microsecond"

func (hr *HandlerRepository) scaleDashboardHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		hr.scale.Recheck()

		type warehouseItem struct {
			Keg    int `json:"keg"`
			Amount int `json:"amount"`
		}

		type pubOutput struct {
			IsOpen   bool   `json:"is_open"`
			OpenedAt string `json:"opened_at"`
			ClosedAt string `json:"closed_at"`
		}

		type output struct {
			IsOk               bool            `json:"is_ok"`
			BeersLeft          int             `json:"beers_left"`
			LastWeight         float64         `json:"last_weight"`
			LastWeightFormated string          `json:"last_weight_formated"`
			LastAt             string          `json:"last_at"`
			LastAtDuration     string          `json:"last_at_duration"`
			Rssi               float64         `json:"rssi"`
			LastUpdate         string          `json:"last_update"`
			LastUpdateDuration string          `json:"last_update_duration"`
			Pub                pubOutput       `json:"pub"`
			ActiveKeg          int             `json:"active_keg"`
			IsLow              bool            `json:"is_low"`
			Warehouse          []warehouseItem `json:"warehouse"`
			WarehouseBeerLeft  int             `json:"warehouse_beer_left"`
		}

		units, err := durafmt.DefaultUnitsCoder.Decode(localizationUnits)
		if err != nil {
			http.Error(w, "Could not decode units", http.StatusInternalServerError)
			return
		}

		warehouse := []warehouseItem{
			{Keg: 10, Amount: hr.scale.Warehouse[0]},
			{Keg: 15, Amount: hr.scale.Warehouse[1]},
			{Keg: 20, Amount: hr.scale.Warehouse[2]},
			{Keg: 30, Amount: hr.scale.Warehouse[3]},
			{Keg: 50, Amount: hr.scale.Warehouse[4]},
		}

		data := output{
			IsOk:               hr.scale.IsOk(),
			BeersLeft:          hr.scale.BeersLeft,
			LastWeight:         hr.scale.Weight,
			LastWeightFormated: fmt.Sprintf("%.2f", hr.scale.Weight/1000),
			LastAt:             formatDate(hr.scale.WeightAt),
			LastAtDuration:     durafmt.Parse(time.Since(hr.scale.WeightAt).Round(time.Second)).LimitFirstN(2).Format(units),
			Rssi:               hr.scale.Rssi,
			LastUpdate:         formatDate(hr.scale.LastOk),
			LastUpdateDuration: durafmt.Parse(time.Since(hr.scale.LastOk).Round(time.Second)).LimitFirstN(2).Format(units),
			Pub: pubOutput{
				IsOpen:   hr.scale.Pub.IsOpen,
				OpenedAt: formatTime(hr.scale.Pub.OpenedAt),
				ClosedAt: formatTime(hr.scale.Pub.ClosedAt),
			},
			ActiveKeg:         hr.scale.ActiveKeg,
			IsLow:             hr.scale.IsLow,
			Warehouse:         warehouse,
			WarehouseBeerLeft: GetWarehouseBeersLeft(hr.scale.Warehouse),
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
		_, _ = w.Write(getOkJson())
	}
}
