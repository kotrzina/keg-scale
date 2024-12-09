package ai

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func ProvideWeather() (string, error) {
	now := time.Now()
	url := fmt.Sprintf("https://data.pocasi-data.cz/data/pocasi/v1/aladin/%d/%02d/%02d/06/100/337/dnes.json", now.Year(), int(now.Month()), now.Day())
	fmt.Println(url)
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not get weather data: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body: %w", err)
	}

	fmt.Println(string(body))
	return string(body), nil
}
