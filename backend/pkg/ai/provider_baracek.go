package ai

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type BaracekProvider struct{}

func (provider *BaracekProvider) GetItems() ([]BeerItem, error) {
	var (
		err      error
		document *html.Node
	)

	pages, err := provider.GetPages("https://www.baracek.cz/sortiment/pivo-sudy")
	if err != nil {
		return nil, fmt.Errorf("could not get pages from Baracek: %w", err)
	}

	items := []BeerItem{}
	for _, page := range pages {

		document, err = provider.getParserForPage(page)
		if err != nil {
			return nil, fmt.Errorf("could not get response from Baracek: %w", err)
		}

		xpath := "//div[contains(@class, 'view-content')]/div[contains(@class, 'views-row')]"
		els, err := htmlquery.QueryAll(document, xpath)
		if err != nil {
			return nil, fmt.Errorf("could not parse items from Baracek: %w", err)
		}

		for _, el := range els {
			title := ""
			href := ""
			price := ""

			titleXpath := "//span/h2/a"
			titles, err := htmlquery.QueryAll(el, titleXpath)
			if err != nil {
				return nil, fmt.Errorf("could not parse title from Baracek: %w", err)
			}
			if len(titles) == 1 {
				for _, attr := range titles[0].Attr {
					if attr.Key == "href" {
						href = attr.Val
					}

					if attr.Key == "title" {
						title = sanitizeBeerTitle(attr.Val)
					}
				}
			}

			priceXpath := "//span[contains(@class, 'uc-price')]"
			prices, err := htmlquery.QueryAll(el, priceXpath)
			if err != nil {
				return nil, fmt.Errorf("could not parse price from Baracek: %w", err)
			}
			if len(prices) == 1 {
				price = prices[0].FirstChild.Data
			}

			stock := StockTypeUnknown
			if strings.Contains(htmlquery.InnerText(el), "Skladem: ano") {
				stock = StockTypeAvailable
			}

			if title != "" {
				items = append(items, BeerItem{
					Title: title,
					link:  "https://www.baracek.cz" + href,
					Price: price,
					stock: stock,
				})
			}
		}
	}

	return items, nil
}

func (provider *BaracekProvider) GetPages(baseURL string) ([]string, error) {
	document, err := provider.getParserForPage(baseURL)
	if err != nil {
		return nil, fmt.Errorf("could not get response from Baracek: %w", err)
	}

	xpath := "//ul[contains(@class, 'pager')]/li[contains(@class, 'pager-item')]/a"
	els, err := htmlquery.QueryAll(document, xpath)
	if err != nil {
		return nil, fmt.Errorf("could not parse pages from Baracek: %w", err)
	}

	pages := []string{baseURL}
	for _, el := range els {
		for _, attr := range el.Attr {
			if attr.Key == "href" {
				pages = append(pages, "https://www.baracek.cz"+attr.Val)
			}
		}
	}

	return pages, nil
}

func (provider *BaracekProvider) getParserForPage(url string) (*html.Node, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get response from Baracek: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body from Baracek: %w", err)
	}

	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("could not parse body from Baracek: %w", err)
	}

	return root, nil
}
