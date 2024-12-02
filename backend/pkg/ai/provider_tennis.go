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

func ProvideTennisData(name string) (string, error) {
	url := fmt.Sprintf("https://hriste.kotrzina.cz/turnaj/%s", name)
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not get response from Baracek: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read body from Baracek: %w", err)
	}

	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("could not parse body from Baracek: %w", err)
	}

	xpath := "//main/div[2]/div"

	els, err := htmlquery.QueryAll(root, xpath)
	if err != nil {
		return "", fmt.Errorf("could not parse pages from Baracek: %w", err)
	}

	if len(els) != 1 {
		return "", fmt.Errorf("could not find tennis data")
	}

	doc := htmlquery.OutputHTML(els[0], false)
	doc = removeUnwantedHTML(doc)
	doc = strings.ReplaceAll(doc, "<!-- -->", "")

	return doc, nil
}
