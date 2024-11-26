package shops

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

type Maneo struct {
	client http.Client
}

func NewManeo() *Maneo {
	return &Maneo{
		client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (shop *Maneo) GetBeer(url string) (*Beer, error) {
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
		return nil, fmt.Errorf("could not get response from Maneo: %w", err)
	}

	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("count not read data from Maneo: %w", err)
	}

	body, err := shop.decodeWindows1250(content)
	if err != nil {
		return nil, fmt.Errorf("could not decode windows-1250: %w", err)
	}
	doc, err := shop.parseRoot(body)
	if err != nil {
		return nil, fmt.Errorf("could not parse html from Maneo: %w", err)
	}

	price, err := shop.parsePrice(doc)
	if err != nil {
		return nil, fmt.Errorf("could not parse price from Maneo: %w", err)
	}
	stock, err := shop.parseStock(doc)
	if err != nil {
		return nil, fmt.Errorf("could not parse stock from Maneo: %w", err)
	}
	title, err := shop.parseTitle(doc)
	if err != nil {
		return nil, fmt.Errorf("could not parse title from Maneo: %w", err)
	}

	return &Beer{
		Title: title,
		Price: price,
		Stock: stock,
	}, nil
}

func (shop *Maneo) parseRoot(body string) (*html.Node, error) {
	return html.Parse(strings.NewReader(body))
}

func (shop *Maneo) parseTitle(node *html.Node) (string, error) {
	xpath := "//h1[contains(@class, 'dp-name')]"
	els, err := htmlquery.QueryAll(node, xpath)
	if err != nil {
		return "", fmt.Errorf("could not parse title: %w", err)
	}
	if len(els) != 1 {
		return "", fmt.Errorf("could not parse title: %w", err)
	}

	title := els[0].FirstChild.Data
	return title, nil
}

func (shop *Maneo) parsePrice(node *html.Node) (int, error) {
	xpath := "//div[contains(@class, 'dp-cena') and contains(@class, 'dp-right')]"
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

func (shop *Maneo) parseStock(node *html.Node) (StockType, error) {
	xpath := "//span[contains(@class, 'tucne')]"
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
func (shop *Maneo) decodeWindows1250(enc []byte) (string, error) {
	dec := charmap.Windows1250.NewDecoder()
	out, err := dec.Bytes(enc)
	if err != nil {
		return "", fmt.Errorf("could not decode windows-1250: %w", err)
	}
	return string(out), nil
}
