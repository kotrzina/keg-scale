package ai

import (
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
)

type SdhEvent struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func ProvideSdhEvents() (string, error) {
	url := "https://sdhveselice.cz/novinky/1/"

	document, err := provideParsePage(url)
	if err != nil {
		return "", fmt.Errorf("could not parse page: %w", err)
	}

	linksXpath := "//div[contains(@class, 'col-lg-8')]/a"
	links, err := htmlquery.QueryAll(document, linksXpath)
	if err != nil {
		return "", fmt.Errorf("could not parse links: %w", err)
	}

	output := strings.Builder{}

	for i, link := range links {
		for _, attr := range link.Attr {
			if attr.Key == "href" {
				event, err := ProvideSdhEvent(attr.Val)
				if err != nil {
					return "", fmt.Errorf("could not provide event: %w", err)
				}

				output.WriteString(fmt.Sprintf("%d. %s\n%s\n\n", i+1, event.Title, event.Content))
			}
		}
	}

	return output.String(), nil
}

func ProvideSdhEvent(path string) (SdhEvent, error) {
	emptyEvent := SdhEvent{}
	url := fmt.Sprintf("https://sdhveselice.cz%s", path)

	document, err := provideParsePage(url)
	if err != nil {
		return emptyEvent, fmt.Errorf("could not parse page: %w", err)
	}

	titleXpath := "//h1"
	titleElement, err := htmlquery.QueryAll(document, titleXpath)
	if err != nil {
		return emptyEvent, fmt.Errorf("could not parse title element: %w", err)
	}
	if len(titleElement) != 1 {
		return emptyEvent, fmt.Errorf("could not find title element")
	}
	title := htmlquery.InnerText(titleElement[0])

	contentXpath := "//div[contains(@class, 'main')]/div[contains(@class, 'container')]"
	contentElement, err := htmlquery.QueryAll(document, contentXpath)
	if err != nil {
		return emptyEvent, fmt.Errorf("could not parse content element: %w", err)
	}
	if len(contentElement) != 1 {
		return emptyEvent, fmt.Errorf("could not find content element")
	}

	content := htmlquery.OutputHTML(contentElement[0], false)

	return SdhEvent{
		Title:   title,
		Content: removeUnwantedHTML(content),
	}, nil
}
