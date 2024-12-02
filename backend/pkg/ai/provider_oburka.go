package ai

import (
	"fmt"

	"github.com/antchfx/htmlquery"
)

func ProvideOburkaMenu() (string, error) {
	url := "https://www.restaurace-oburka.cz/tydenni-menu-od-11-00-do-14-00hod/"
	root, err := provideParsePage(url)
	if err != nil {
		return "", fmt.Errorf("could not get response from Oburka: %w", err)
	}

	xpath := "//div[@id='left']/div[contains(@class, 'text')]"
	els, err := htmlquery.QueryAll(root, xpath)
	if err != nil {
		return "", fmt.Errorf("could not parse menu from Oburka: %w", err)
	}

	if len(els) != 1 {
		return "", fmt.Errorf("could not find Oburka menu data")
	}

	doc := htmlquery.OutputHTML(els[0], false)
	doc = removeUnwantedHTML(doc)

	return doc, nil
}
