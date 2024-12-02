package ai

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"golang.org/x/net/html"
)

var (
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

	defer resp.Body.Close()

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

func removeUnwantedHTML(doc string) string {
	doc = reSvgs.ReplaceAllString(doc, "")
	doc = reImages.ReplaceAllString(doc, "")
	doc = reStyles.ReplaceAllString(doc, "")
	doc = reClasses.ReplaceAllString(doc, "")
	doc = reSpaces.ReplaceAllString(doc, " ")
	doc = endTagSpace.ReplaceAllString(doc, ">")
	doc = tagBetweenSpace.ReplaceAllString(doc, "><")

	return doc
}
