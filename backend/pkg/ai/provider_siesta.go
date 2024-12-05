package ai

import (
	"encoding/json"
	"fmt"

	"github.com/antchfx/htmlquery"
)

type SiestaItem struct {
	Title       string `json:"title"`
	Price       string `json:"price"`
	Description string `json:"description"`
}

func ProvideSiestaMenu() (string, error) {
	pages := []string{
		"pizza",
		"grill",
		"testoviny",
		"ostatni",
		"dezerty",
	}

	var items []SiestaItem

	for _, page := range pages {
		itemsPage, err := provideSiestaPage(page)
		if err != nil {
			return "", fmt.Errorf("could not get menu from Siesta: %w", err)
		}
		items = append(items, itemsPage...)
	}

	output, err := json.Marshal(items)
	if err != nil {
		return "", fmt.Errorf("could not marshal menu items: %w", err)
	}

	return string(output), nil
}

func provideSiestaPage(page string) ([]SiestaItem, error) {
	url := fmt.Sprintf("https://www.siestapizza.cz/index.php?url=%s", page)
	root, err := provideParsePage(url)
	if err != nil {
		return nil, fmt.Errorf("could not get response from Siesta: %w", err)
	}

	xpath := "//div[contains(@class, 'item')]"

	els, err := htmlquery.QueryAll(root, xpath)
	if err != nil {
		return nil, fmt.Errorf("could not parse menu from Siesta: %w", err)
	}

	if len(els) == 1 {
		return nil, fmt.Errorf("could not find the menu data")
	}

	titleXpath := "//h3/strong"
	priceXpath := "//span[contains(@class, 'price')]"
	description := "//h4"
	items := make([]SiestaItem, len(els))
	for i, el := range els {
		title, err := htmlquery.Query(el, titleXpath)
		if err != nil {
			return nil, fmt.Errorf("could not find title for menu item: %w", err)
		}

		price, err := htmlquery.Query(el, priceXpath)
		if err != nil {
			return nil, fmt.Errorf("could not find price for menu item: %w", err)
		}

		ingredients, err := htmlquery.Query(el, description)
		if err != nil {
			return nil, fmt.Errorf("could not find ingredients for menu item: %w", err)
		}

		items[i] = SiestaItem{
			Title:       htmlquery.InnerText(title),
			Price:       htmlquery.InnerText(price),
			Description: removeUnwantedHTML(htmlquery.InnerText(ingredients)),
		}
	}

	return items, nil
}
