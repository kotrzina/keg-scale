package ai

import (
	"fmt"

	"github.com/antchfx/htmlquery"
)

func ProvideWeather() (string, error) {
	const url = "https://www.yr.no/en/forecast/daily-table/2-3063044/Czech Republic/South Moravian/Blansko District/Veselice"

	root, err := provideParsePage(url)
	if err != nil {
		return "", fmt.Errorf("could not parse page: %w", err)
	}

	// find the weather
	forecastXpath := "//div[@id='page-content']"
	els, err := htmlquery.QueryAll(root, forecastXpath)
	if err != nil {
		return "", fmt.Errorf("could not parse weather data: %w", err)
	}

	if len(els) != 1 {
		return "", fmt.Errorf("could not find weather data")
	}

	return removeUnwantedHTML(htmlquery.OutputHTML(els[0], false)), nil
}
