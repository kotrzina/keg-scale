package ai

import (
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
)

func ProvideBrousekMenu() (string, error) {
	url := "https://hotelbrousek.cz/denni-menu/"
	root, err := provideParsePage(url)
	if err != nil {
		return "", fmt.Errorf("could not get response from Brousek: %w", err)
	}

	sb := &strings.Builder{}

	xpathPrimary := "//div[@id='primary']"
	elsPrimary, err := htmlquery.QueryAll(root, xpathPrimary)
	if err != nil {
		return "", fmt.Errorf("could not parse primary content from Brousek: %w", err)
	}
	if len(elsPrimary) != 1 {
		return "", fmt.Errorf("could not find Brousek primary data")
	}
	sb.WriteString(htmlquery.OutputHTML(elsPrimary[0], false))

	xpathMenu := "//div[contains(@class, 'main')]"
	elsMenu, err := htmlquery.QueryAll(root, xpathMenu)
	if err != nil {
		return "", fmt.Errorf("could not parse menu from Brousek: %w", err)
	}
	if len(elsMenu) != 1 {
		return "", fmt.Errorf("could not find Brousek menu data")
	}
	sb.WriteString(htmlquery.OutputHTML(elsMenu[0], false))

	doc := removeUnwantedHTML(sb.String())

	return doc, nil
}
