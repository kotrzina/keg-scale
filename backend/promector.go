package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Promector represents a Prometheus collector
type Promector struct {
	baseUrl  string
	user     string
	password string

	logger *logrus.Logger
	ctx    context.Context
	mtx    sync.RWMutex

	data map[string][]RangeRecord
}

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Values []interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func NewPromector(baseUrl, user, password string, logger *logrus.Logger, ctx context.Context) *Promector {
	prom := &Promector{
		baseUrl:  baseUrl,
		user:     user,
		password: password,

		logger: logger,
		ctx:    ctx,
		mtx:    sync.RWMutex{},

		data: make(map[string][]RangeRecord, 4),
	}

	// periodically call recheck
	go func(prom *Promector) {
		tick := time.NewTicker(2 * time.Minute)
		defer tick.Stop()
		for {
			select {
			case <-prom.ctx.Done():
				prom.logger.Debug("Promector downloading stopped")
				return
			case <-tick.C:
				prom.Refresh()
				prom.logger.Debug("Promector data refreshed")
			}
		}
	}(prom)
	prom.Refresh()

	return prom
}

type RangeRecord struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type ChartInterval struct {
	Interval string        `json:"interval"`
	Values   []RangeRecord `json:"values"`
}

type Charts struct {
	BeersLeft []ChartInterval `json:"beers_left"`
}

func (p *Promector) Refresh() {
	type request struct {
		key   string
		query string
		hours int
		step  string
	}

	requests := []request{
		{"scale_beers_left_1h", "scale_beers_left", 1, "5m"},
		{"scale_beers_left_3h", "scale_beers_left", 3, "10m"},
		{"scale_beers_left_6h", "scale_beers_left", 6, "20m"},
		{"scale_beers_left_24h", "scale_beers_left", 24, "1h"},
	}

	tmp := make(map[string][]RangeRecord, 4)

	for _, r := range requests {
		data, err := p.GetRangeData(r.query, r.hours, r.step)
		if err != nil {
			p.logger.Errorf("could not get range data for %s: %v", r.key, err)
			continue
		}

		tmp[r.key] = data
	}

	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.data = tmp
}

func (p *Promector) GetChartData() Charts {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return Charts{
		BeersLeft: []ChartInterval{
			{"1h", p.data["scale_beers_left_1h"]},
			{"3h", p.data["scale_beers_left_3h"]},
			{"6h", p.data["scale_beers_left_6h"]},
			{"24h", p.data["scale_beers_left_24h"]},
		},
	}

}

func (p *Promector) GetRangeData(query string, hours int, step string) ([]RangeRecord, error) {
	url := fmt.Sprintf("%s/api/v1/query_range?", p.baseUrl)
	req, _ := http.NewRequest("GET", url, nil)

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("step", step)
	q.Add("start", fmt.Sprintf("%d", time.Now().Unix()-(60*60*int64(hours))))
	q.Add("end", fmt.Sprintf("%d", time.Now().Unix()))
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", getBaseAuth(p.user, p.password))
	client := &http.Client{
		Timeout: 10 * time.Second, // Set the timeout duration here
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get value from prometheus: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	var prometheusResponse PrometheusResponse
	err = json.Unmarshal(data, &prometheusResponse)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	if prometheusResponse.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s", prometheusResponse.Status)
	}

	if len(prometheusResponse.Data.Result) != 1 {
		return nil, fmt.Errorf("unexpected number of results: %d", len(prometheusResponse.Data.Result))
	}

	records := make([]RangeRecord, len(prometheusResponse.Data.Result[0].Values))

	i := 0
	for _, value := range prometheusResponse.Data.Result[0].Values {
		record, ok := value.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected value type: %T", value)
		}

		if len(record) != 2 {
			return nil, fmt.Errorf("unexpected number of values: %d", len(record))
		}

		t := time.Unix(int64(record[0].(float64)), 0)
		v, e := strconv.Atoi(record[1].(string))
		if e != nil {
			return nil, fmt.Errorf("could not convert value to int: %w", e)
		}

		records[i] = RangeRecord{
			Label: formatTime(t),
			Value: v,
		}

		i++

	}

	return records, nil
}

func getBaseAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
