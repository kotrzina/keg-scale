package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func ProvideEventsBlansko() (string, error) {
	type BkEvent struct {
		Title    string `json:"title"`
		Category string `json:"category"`
		Place    string `json:"place"`
		Date     string `json:"date"`
	}

	url := "https://www.akceblansko.cz/"
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not get response from akceblansko: %w", err)
	}

	defer resp.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read body from akceblansko: %w", err)
	}

	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("could not parse body from akceblansko: %w", err)
	}

	xpath := "//div[contains(@class, 'eventCard')]"

	els, err := htmlquery.QueryAll(root, xpath)
	if err != nil {
		return "", fmt.Errorf("could not parse pages from akceblansko: %w", err)
	}

	if len(els) == 0 {
		return "", fmt.Errorf("could not find any event")
	}

	xpathTitle := "//p[contains(@class, 'card-title')]"
	xpathCategory := "//span[contains(@class, 'card-category')]"
	xpathPlace := "//span[contains(@class, 'place-name')]"
	xpathDate := "//div[contains(@class, 'card-eventDate')]"
	events := make([]BkEvent, len(els))
	for i, el := range els {
		title, err := htmlquery.QueryAll(el, xpathTitle)
		if err != nil {
			return "", fmt.Errorf("could not parse title from Baracek: %w", err)
		}

		category, err := htmlquery.QueryAll(el, xpathCategory)
		if err != nil {
			return "", fmt.Errorf("could not parse category from akceblansko: %w", err)
		}

		place, err := htmlquery.QueryAll(el, xpathPlace)
		if err != nil {
			return "", fmt.Errorf("could not parse place from akceblansko: %w", err)
		}

		date, err := htmlquery.QueryAll(el, xpathDate)
		if err != nil {
			return "", fmt.Errorf("could not parse date from akceblansko: %w", err)
		}

		events[i] = BkEvent{
			Title:    strings.TrimSuffix(removeUnwantedHTML(htmlquery.InnerText(title[0])), " kino"),
			Category: removeUnwantedHTML(htmlquery.InnerText(category[0])),
			Place:    removeUnwantedHTML(htmlquery.InnerText(place[0])),
			Date:     removeUnwantedHTML(htmlquery.InnerText(date[0])),
		}
	}

	output, err := json.Marshal(events)
	if err != nil {
		return "", fmt.Errorf("could not marshal events from akceblansko: %w", err)
	}

	return string(output), nil
}
