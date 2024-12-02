package ai

import (
	"fmt"

	"github.com/antchfx/htmlquery"
)

func ProvideHrabenkaMenu() (string, error) {
	url := "https://hrabenka.cz/restaurace/tydenni-obedove-menu/"
	root, err := provideParsePage(url)
	if err != nil {
		return "", fmt.Errorf("could not get response from Hrabenka: %w", err)
	}

	xpath := "//div[contains(@class, 'entry-content')]"
	els, err := htmlquery.QueryAll(root, xpath)
	if err != nil {
		return "", fmt.Errorf("could not parse menu from Hrabenka: %w", err)
	}

	if len(els) != 1 {
		return "", fmt.Errorf("could not find Hrabenka menu data")
	}

	doc := htmlquery.OutputHTML(els[0], false)
	doc = removeUnwantedHTML(doc)

	return doc, nil
}
