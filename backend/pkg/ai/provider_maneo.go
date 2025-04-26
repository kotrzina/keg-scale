package ai

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"
)

type ManeoProvider struct{}

func (provider *ManeoProvider) GetItems() ([]BeerItem, error) {
	const url = "https://www.eshop.maneo.cz/pivo-sudove-vse-katskup5.1.A.php?KATALOG_ZBOZI_VYPISOVAT_OD=0&ajax=1&KATALOG_POCET_ZBOZI_VYPISOVAT=1000"

	var (
		err     error
		resp    *http.Response
		retries = 3
	)

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	for retries > 0 {
		resp, err = client.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			retries--
		} else {
			break
		}
	}

	defer resp.Body.Close() //nolint: errcheck

	if err != nil {
		return nil, fmt.Errorf("could not get response from Maneo: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("response body is empty")
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read data from Maneo: %w", err)
	}

	body, err := provider.decodeWindows1250(content)
	if err != nil {
		return nil, fmt.Errorf("could not decode windows-1250: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("could not parse html from Maneo: %w", err)
	}

	xpath := "//div[contains(@class, 'produkt')]/div[contains(@class, 'produkt-in')]"
	els, err := htmlquery.QueryAll(doc, xpath)
	if err != nil {
		return nil, fmt.Errorf("could not parse items from Maneo: %w", err)
	}

	items := make([]BeerItem, len(els))
	for i, el := range els {
		title, path, err := provider.parseTitleAndURL(el)
		if err != nil {
			return nil, fmt.Errorf("could not parse title from Maneo: %w", err)
		}

		price, err := provider.parsePrice(el)
		if err != nil {
			return nil, fmt.Errorf("could not parse price from Maneo: %w", err)
		}

		stock, err := provider.parseStock(el)
		if err != nil {
			return nil, fmt.Errorf("could not parse stock from Maneo: %w", err)
		}

		items[i] = BeerItem{
			Title: title,
			link:  fmt.Sprintf("%s/%s", "https://www.eshop.maneo.cz/", path),
			Price: price,
			stock: stock,
		}
	}

	return items, nil
}

func (provider *ManeoProvider) parseTitleAndURL(node *html.Node) (title, url string, err error) {
	xpath := "//div[contains(@class, 'pr-top')]/h2/a"
	els, err := htmlquery.QueryAll(node, xpath)
	if err != nil {
		return "", "", fmt.Errorf("could not parse title: %w", err)
	}
	if len(els) != 1 {
		return "", "", fmt.Errorf("could not parse title: %w", err)
	}

	title = sanitizeBeerTitle(els[0].FirstChild.Data)

	url = ""
	for _, attr := range els[0].Attr {
		if attr.Key == "href" {
			url = attr.Val
		}
	}

	return title, url, nil
}

func (provider *ManeoProvider) parsePrice(node *html.Node) (string, error) {
	xpath := "//div[contains(@class, 'castka')]/strong"
	els, err := htmlquery.QueryAll(node, xpath)
	if err != nil {
		return "", fmt.Errorf("could not parse price: %w", err)
	}
	if len(els) != 1 {
		return "", fmt.Errorf("could not parse price: %w", err)
	}

	return els[0].FirstChild.Data, nil
}

func (provider *ManeoProvider) parseStock(node *html.Node) (StockType, error) {
	xpath := "//p[contains(@class, 'p_stav_skladu_vypis')]/span"
	els, err := htmlquery.QueryAll(node, xpath)
	if err != nil {
		return StockTypeUnknown, fmt.Errorf("could not parse stock: %w", err)
	}
	if len(els) != 1 {
		return StockTypeUnknown, fmt.Errorf("could not parse stock: %w", err)
	}

	if strings.EqualFold(els[0].FirstChild.Data, "skladem") {
		return StockTypeAvailable, nil
	}

	return StockTypeUnknown, nil
}

// decodeWindows1250 decodes windows-1250 encoded string to utf-8
// because Maneo is using windows-1250 encoding
func (provider *ManeoProvider) decodeWindows1250(enc []byte) (string, error) {
	dec := charmap.Windows1250.NewDecoder()
	out, err := dec.Bytes(enc)
	if err != nil {
		return "", fmt.Errorf("could not decode windows-1250: %w", err)
	}
	return string(out), nil
}
