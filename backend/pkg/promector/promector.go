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
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Promector represents a Prometheus collector
// We download data periodically and store it in cache
type Promector struct {
	config *config.Config

	logger *logrus.Logger
	ctx    context.Context
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

type RangeRecord struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

func NewPromector(ctx context.Context, c *config.Config, logger *logrus.Logger) *Promector {
	return &Promector{
		ctx: ctx,

		config: c,
		logger: logger,
	}
}

func (p *Promector) GetRangeData(query string, start, end time.Time, step time.Duration) ([]RangeRecord, error) {
	url := fmt.Sprintf("%s/api/v1/query_range?", p.config.PrometheusURL)
	req, err := http.NewRequestWithContext(p.ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("step", formatStep(step))
	q.Add("start", fmt.Sprintf("%d", start.Unix()))
	q.Add("end", fmt.Sprintf("%d", end.Unix()))
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", getBaseAuth(p.config.PrometheusUser, p.config.PrometheusPassword))
	req.Header.Add("X-Scope-OrgID", p.config.PrometheusOrg)
	client := &http.Client{
		Timeout: 15 * time.Second, // Set the timeout duration here
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get value from prometheus: %w", err)
	}

	defer response.Body.Close() //nolint: errcheck

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

		tf, ok := record[0].(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected time format: %T", record[0])
		}
		t := time.Unix(int64(tf), 0)

		vs, ok := record[1].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected value format: %T", record[1])
		}
		v, e := strconv.Atoi(vs)
		if e != nil {
			return nil, fmt.Errorf("could not convert value to int: %w", e)
		}

		type labelFormatter func(time.Time) string
		var labelFunc labelFormatter
		labelFunc = utils.FormatTime
		if step >= 1*time.Hour {
			labelFunc = func(t time.Time) string {
				return t.In(utils.GetTz()).Format("2.1.")
			}
		}

		records[i] = RangeRecord{
			Label: labelFunc(t),
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
