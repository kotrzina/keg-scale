package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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
			Scale scale.FullOutput `json:"scale"`
		}

		data := output{
			Scale: hr.scale.GetScale(),
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
			hr.logger.Errorf("could not get response from AI: %v", err)
			resp = ai.Response{
				Text: "Teď bohužel nedokážu odpovědět. Zkus to prosím později.",
				Cost: ai.Cost{
					Input:  0,
					Output: 0,
				},
			}
		}

		resp.Text = utils.UnwrapHTML(resp.Text)

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

func (hr *HandlerRepository) scaleChartHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		metric := req.URL.Query().Get("metric")
		interval := req.URL.Query().Get("interval")

		allowedMetrics := []string{
			"scale_beers_left",
			"scale_active_keg",
		}
		if !listContains(allowedMetrics, metric) {
			http.Error(w, "Now allowed metric", http.StatusBadRequest)
			return
		}

		var err error
		var start time.Time
		end := time.Now()

		if strings.EqualFold(interval, "ted") {
			// if open return current session
			// if closed return last session
			pub := hr.scale.GetOpeningOutput()
			start = pub.OpenedAt.Add(-10 * time.Minute)
			if !pub.IsOpen {
				end = pub.ClosedAt.Add(10 * time.Minute)
			}
		} else {
			d, err := parseCustomDuration(interval)
			if err != nil {
				http.Error(w, "could not parse interval", http.StatusBadRequest)
				return
			}

			// older than 4 years
			if d > 4*365*24*time.Hour {
				http.Error(w, "Interval too long", http.StatusBadRequest)
				return
			}

			start = end.Add(-d)
		}

		delta := end.Sub(start) // duration between start and end

		// we want to set reasonable step for the delta
		step := 5 * time.Minute
		if delta > 7*24*time.Hour {
			step = 1 * time.Hour
		}
		if delta > 30*24*time.Hour {
			step = 24 * time.Hour
		}

		data, err := hr.promector.GetRangeData(metric, start, end, step)
		if err != nil {
			hr.logger.Errorf("could not get range data for %s: %v", metric, err)
			http.Error(w, "could not get range data", http.StatusInternalServerError)
			return
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

var reCustomDuration = regexp.MustCompile(`^(\d{1,2})([hdwmy])$`)

// parseCustomDuration parses custom duration string
// e.g. 1h, 2d, 3w, 4m, 5y
// it supports units which are not supported by time.ParseDuration
func parseCustomDuration(input string) (time.Duration, error) {
	matches := reCustomDuration.FindStringSubmatch(input)
	if len(matches) != 3 {
		return 0, fmt.Errorf("could not parse custom duration: %s", input)
	}

	var duration time.Duration
	switch matches[2] {
	case "h":
		duration = time.Hour
	case "d":
		duration = 24 * time.Hour
	case "w":
		duration = 7 * 24 * time.Hour
	case "m":
		duration = 30 * 24 * time.Hour
	case "y":
		duration = 365 * 24 * time.Hour
	}

	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("could not parse custom duration: %s", input)
	}

	return time.Duration(n) * duration, nil
}

func listContains(arr []string, item string) bool {
	for _, v := range arr {
		if strings.EqualFold(v, item) {
			return true
		}
	}
	return false
}
