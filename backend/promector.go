package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Promector represents a Prometheus collector
type Promector struct {
	baseUrl  string
	user     string
	password string
}

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func NewPromector(baseUrl, user, password string) *Promector {
	return &Promector{
		baseUrl:  baseUrl,
		user:     user,
		password: password,
	}
}

func (p *Promector) GetLastValue(query string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", p.baseUrl, query)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", getBaseAuth(p.user, p.password))
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("could not get value from prometheus: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("could not read response body: %w", err)
	}

	var prometheusResponse PrometheusResponse
	err = json.Unmarshal(data, &prometheusResponse)
	if err != nil {
		return 0, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	if prometheusResponse.Status != "success" {
		return 0, fmt.Errorf("prometheus query failed: %s", prometheusResponse.Status)
	}

	if len(prometheusResponse.Data.Result) != 1 {
		return 0, fmt.Errorf("unexpected number of results: %d", len(prometheusResponse.Data.Result))
	}

	if len(prometheusResponse.Data.Result[0].Value) != 2 {
		return 0, fmt.Errorf("unexpected number of values: %d", len(prometheusResponse.Data.Result[0].Value))
	}

	val, err := strconv.ParseFloat(prometheusResponse.Data.Result[0].Value[1].(string), 64)
	if err != nil {
		return 0, fmt.Errorf("could not convert value to float: %w", err)

	}

	return val, nil
}

func getBaseAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
