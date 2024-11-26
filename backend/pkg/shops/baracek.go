package shops

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Baracek struct {
	client http.Client
}

func NewBaracek() *Baracek {
	return &Baracek{
		client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (shop *Baracek) GetBeer(url string) (*Beer, error) {
	var (
		err     error
		resp    *http.Response
		retries = 3
	)

	for retries > 0 {
		resp, err = shop.client.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			retries--
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not get response from Baracek: %w", err)
	}

	defer resp.Body.Close()
	document, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("count not read data from Baracek: %w", err)
	}

	root, err := shop.parseRoot(document)
	if err != nil {
		return nil, fmt.Errorf("could not parse html from Baracek: %w", err)
	}

	title, err := shop.parseTitle(root)
	if err != nil {
		return nil, fmt.Errorf("could not parse title from Baracek: %w", err)
	}
	price, err := shop.parsePrice(root)
	if err != nil {
		return nil, fmt.Errorf("could not parse price from Baracek: %w", err)
	}
	stock, err := shop.parseStock(root)
	if err != nil {
		return nil, fmt.Errorf("could not parse stock from Baracek: %w", err)
	}

	return &Beer{
		Title: title,
		Price: price,
		Stock: stock,
	}, nil
}

func (shop *Baracek) parseRoot(body []byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(body))
}

var reSpaces = regexp.MustCompile(`\s+`)

func (shop *Baracek) parseTitle(node *html.Node) (string, error) {
	xpath := "//h1[contains(@class, 'page-title')]"
	els, err := htmlquery.QueryAll(node, xpath)
	if err != nil {
		return "", fmt.Errorf("could not parse title: %w", err)
	}
	if len(els) != 1 {
		return "", fmt.Errorf("could not parse title: %w", err)
	}

	title := els[0].FirstChild.Data
	title = reSpaces.ReplaceAllString(title, " ")

	return title, nil
}

func (shop *Baracek) parsePrice(node *html.Node) (int, error) {
	xpath := "//span[contains(@class, 'uc-price')]"
	els, err := htmlquery.QueryAll(node, xpath)
	if err != nil {
		return 0, fmt.Errorf("could not parse price: %w", err)
	}
	if len(els) != 1 {
		return 0, fmt.Errorf("could not parse price: %w", err)
	}

	price, err := parsePriceString(els[0].FirstChild.Data)
	if err != nil {
		return 0, fmt.Errorf("could not parse price string: %w", err)
	}

	return price, nil
}

func (shop *Baracek) parseStock(node *html.Node) (StockType, error) {
	xpath := "//div[contains(@class, 'field-name-field-skladem')]/div[contains(@class, 'field-items')]/div"
	els, err := htmlquery.QueryAll(node, xpath)
	if err != nil {
		return 0, fmt.Errorf("could not parse stock: %w", err)
	}
	if len(els) != 1 {
		return 0, fmt.Errorf("could not parse stock: %w", err)
	}

	if shop.sanitizeString(els[0].FirstChild.Data) == "ano" {
		return StockTypeAvailable, nil
	}

	return StockTypeUnknown, nil
}

func (shop *Baracek) sanitizeString(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	return s
}
