package promector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Promector represents a Prometheus collector
// We download data periodically and store it in cache
type Promector struct {
	baseURL  string
	user     string
	password string

	scale  *scale.Scale
	logger *logrus.Logger
	ctx    context.Context

	mtx   sync.RWMutex
	cache Charts
}

type Response struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Values []interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func NewPromector(ctx context.Context, baseURL, user, password string, s *scale.Scale, logger *logrus.Logger) *Promector {
	prom := &Promector{
		baseURL:  baseURL,
		user:     user,
		password: password,

		scale:  s,
		logger: logger,
		ctx:    ctx,

		mtx: sync.RWMutex{},
		cache: Charts{
			BeersLeft: []ChartInterval{},
			ActiveKeg: []ChartInterval{},
		},
	}

	prom.Refresh() // first download on start
	// periodically download data
	go func(prom *Promector) {
		tick := time.NewTicker(90 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-prom.ctx.Done():
				prom.logger.Debug("Promector downloading stopped")
				return
			case <-tick.C:
				prom.Refresh()
				prom.logger.Debug("Promector cache refreshed")
			}
		}
	}(prom)

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
	ActiveKeg []ChartInterval `json:"active_keg"`
}

func (p *Promector) Refresh() {
	type request struct {
		key   string
		query string
		start time.Time
		end   time.Time
		step  time.Duration
	}

	now := time.Now()
	step := 5 * time.Minute
	requests := []request{
		{"scale_active_keg_7d", "scale_active_keg", now.Add(-7 * 24 * time.Hour), now, step},
		{"scale_active_keg_14d", "scale_active_keg", now.Add(-14 * 24 * time.Hour), now, step},
		{"scale_beers_left_1h", "scale_beers_left", now.Add(-1 * time.Hour), now, step},
		{"scale_beers_left_3h", "scale_beers_left", now.Add(-3 * time.Hour), now, step},
		{"scale_beers_left_6h", "scale_beers_left", now.Add(-6 * time.Hour), now, step},
		{"scale_beers_left_24h", "scale_beers_left", now.Add(-24 * time.Hour), now, step},
	}

	opening := p.scale.GetOpeningOutput()

	if opening.IsOpen {
		// chart for current session
		// let's take 10 minutes before opening to now
		// step is calculated based on open duration - we want to have approx 30 points in the chart
		dataStart := opening.OpenedAt.Add(-10 * time.Minute)
		requests = append(requests, request{"scale_beers_left_now", "scale_beers_left", dataStart, now, step})
	} else {
		// chart for last session
		// let's take 10 minutes before opening to closing time plus 10 minutes
		// step is calculated based on open duration - we want to have approx 14 points in the chart
		timeStart := opening.OpenedAt.Add(-10 * time.Minute)
		timeEnd := opening.ClosedAt.Add(10 * time.Minute)
		requests = append(requests, request{"scale_beers_left_last", "scale_beers_left", timeStart, timeEnd, step})
	}

	wg := sync.WaitGroup{}
	wg.Add(len(requests))
	results := make(map[string][]RangeRecord, len(requests))
	mapMux := sync.Mutex{}

	for _, req := range requests {
		go func(r request) {
			defer wg.Done()
			data, err := p.GetRangeData(r.query, r.start, r.end, r.step)
			if err != nil {
				p.logger.Errorf("could not get range data for %s: %v", r.query, err)
				return
			}

			// safely write result to the map
			mapMux.Lock()
			defer mapMux.Unlock()
			results[r.key] = data
		}(req)
	}

	wg.Wait()

	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.cache = Charts{
		BeersLeft: []ChartInterval{
			{"1h", results["scale_beers_left_1h"]},
			{"3h", results["scale_beers_left_3h"]},
			{"6h", results["scale_beers_left_6h"]},
			{"24h", results["scale_beers_left_24h"]},
			{"now", results["scale_beers_left_now"]},
			{"last", results["scale_beers_left_last"]},
		},
		ActiveKeg: []ChartInterval{
			{"7d", results["scale_active_keg_7d"]},
			{"14d", results["scale_active_keg_14d"]},
		},
	}
}

func (p *Promector) GetChartData() Charts {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return p.cache
}

func (p *Promector) GetRangeData(query string, start, end time.Time, step time.Duration) ([]RangeRecord, error) {
	url := fmt.Sprintf("%s/api/v1/query_range?", p.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("step", formatStep(step))
	q.Add("start", fmt.Sprintf("%d", start.Unix()))
	q.Add("end", fmt.Sprintf("%d", end.Unix()))
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

	var prometheusResponse Response
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
			Label: utils.FormatTime(t),
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

func formatStep(step time.Duration) string {
	f := step.String()

	if strings.Contains(f, "m0s") {
		f = strings.TrimSuffix(f, "0s")
	}

	if strings.Contains(f, "h0m") {
		f = strings.TrimSuffix(f, "0m")
	}

	return f
}
