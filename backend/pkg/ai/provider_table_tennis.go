package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func ProvideTableTennisResults() (string, error) {
	url := "https://stis.ping-pong.cz/los-vse/svaz-420701/rocnik-2024/soutez-5685/druzstvo-58442"
	body, err := providerTableTennisPage(url)
	if err != nil {
		return "", fmt.Errorf("could not get response from Stis: %w", err)
	}

	root, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("could not parse body from page: %w", err)
	}

	contentXpath := "//div[@id='print_content']"
	contentEls, err := htmlquery.QueryAll(root, contentXpath)
	if err != nil {
		return "", fmt.Errorf("could not parse primary content from Stis: %w", err)
	}
	if len(contentEls) != 1 {
		return "", fmt.Errorf("could not find Stis content div")
	}

	return removeHrefs(removeUnwantedHTML(htmlquery.OutputHTML(contentEls[0], false))), nil
}

func ProvideTableTennisLeagueTable() (string, error) {
	url := "https://stis.ping-pong.cz/tabulka/svaz-420701/rocnik-2024/soutez-5685"
	body, err := providerTableTennisPage(url)
	if err != nil {
		return "", fmt.Errorf("could not get response from Stis: %w", err)
	}

	jsonLine := ""
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "var initTabData") {
			jsonLine = line
		}
	}

	if jsonLine == "" {
		return "", fmt.Errorf("could not find league table data")
	}

	jsonLine = strings.TrimPrefix(jsonLine, "var initTabData = \"")
	jsonLine = strings.TrimSuffix(jsonLine, "\";")
	jsonLine = strings.ReplaceAll(jsonLine, `\"`, `"`) // unescape quotes

	input := &struct {
		Tables []struct {
			Data []struct {
				Nazev       string `json:"nazev"`
				Vyhry       int    `json:"vyhry"`
				Remizy      int    `json:"remizy"`
				Prohry      int    `json:"prohry"`
				Kontprohry  int    `json:"kontprohry"`
				Vyhrbody    int    `json:"vyhrbody"`
				Prohrbody   int    `json:"prohrbody"`
				Lepsi       int    `json:"lepsi"`
				OdpocetBodu string `json:"odpocet_bodu"`
				Vyhryhorsi  int    `json:"vyhryhorsi"`
				Prohrylepsi int    `json:"prohrylepsi"`
				Body        int    `json:"body"`
				Zapasy      struct {
					Body float64 `json:"body"`
				} `json:"zapasy"`
				Pocet    int    `json:"pocet,omitempty"`
				Bodu     int    `json:"bodu,omitempty"`
				OdhlText string `json:"odhl_text,omitempty"`
			} `json:"data"`
		} `json:"tables"`
	}{}

	err = json.Unmarshal([]byte(jsonLine), &input)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal league table data: %w", err)
	}

	decoded, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("could not marshal league table data: %w", err)
	}

	return string(decoded), nil
}

func providerTableTennisPage(url string) ([]byte, error) {
	initReq, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("could not create init request: %w", err)
	}
	initReq.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0")
	initReq.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	initResp, err := http.DefaultClient.Do(initReq)
	if err != nil {
		return nil, fmt.Errorf("could not get response for init request: %w", err)
	}
	defer initResp.Body.Close()

	realReq, err := http.NewRequest(http.MethodGet, url+"?nuser=1", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("could not create real request: %w", err)
	}
	realReq.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0")
	realReq.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	realReq.Header.Add("Cookie", strings.Split(initResp.Header.Get("Set-Cookie"), ";")[0])
	realResp, err := http.DefaultClient.Do(realReq)
	if err != nil {
		return nil, fmt.Errorf("could not get response from page: %w", err)
	}
	defer realResp.Body.Close()

	body, err := io.ReadAll(realResp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body from page: %w", err)
	}

	return body, nil
}
