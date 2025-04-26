package ai

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type StockType string

const (
	StockTypeAvailable StockType = "available"
	StockTypeUnknown   StockType = "unknown"
)

var (
	// compile regexes compile time
	reSvgs          = regexp.MustCompile(`<svg.*?</svg>`)
	reImages        = regexp.MustCompile(`<img.*?>`)
	reStyles        = regexp.MustCompile(`style=".*?"`)
	reClasses       = regexp.MustCompile(`class=".*?"`)
	reSpaces        = regexp.MustCompile(`\s+`)
	endTagSpace     = regexp.MustCompile(` >`)
	tagBetweenSpace = regexp.MustCompile(`> <`)
)

func provideParsePage(url string) (*html.Node, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get response from page: %w", err)
	}

	defer resp.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body from page: %w", err)
	}

	els, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("could not parse body from page: %w", err)
	}

	return els, nil
}

type beerProvider interface {
	GetItems() ([]BeerItem, error)
}

type BeerItem struct {
	Title string    `json:"title"`
	link  string    // do not export for language model (savings)
	Price string    `json:"price"`
	stock StockType // currently not exported for language model
}

// title sanitization - remove useless words and prefixes
func sanitizeBeerTitle(s string) string {
	titlePrefixes := []string{"sud", "pivo"}
	uselessWords := []string{"AKCE"}
	for _, prefix := range titlePrefixes {
		s = strings.TrimPrefix(s, prefix)
	}
	for _, word := range uselessWords {
		s = strings.ReplaceAll(s, word, "")
	}
	s = reSpaces.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}

// removeUnwantedHTML removes unwanted HTML elements from the document
// images, svgs, styles, classes, multiple spaces, spaces between tags
// none of these are needed for the AI to understand the content
func removeUnwantedHTML(doc string) string {
	doc = reSvgs.ReplaceAllString(doc, "")
	doc = reImages.ReplaceAllString(doc, "")
	doc = reStyles.ReplaceAllString(doc, "")
	doc = reClasses.ReplaceAllString(doc, "")
	doc = reSpaces.ReplaceAllString(doc, " ")
	doc = endTagSpace.ReplaceAllString(doc, ">")
	doc = tagBetweenSpace.ReplaceAllString(doc, "><")
	doc = strings.TrimSpace(doc)
	return doc
}

var reHref = regexp.MustCompile(`href=".*?"`)

func removeHrefs(doc string) string {
	return reHref.ReplaceAllString(doc, "")
}
