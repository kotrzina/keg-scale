package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kotrzina/keg-scale/pkg/shops"
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/jsonschema"
	"github.com/mmcdole/gofeed"
)

func (ai *Ai) currentTimeTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "current_time",
			Description: "Provides current time in Europe/Prague timezone. The place where the pub is located.",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			return utils.FormatDate(time.Now()), nil
		},
	}
}

func (ai *Ai) isOpenTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "is_pub_open",
			Description: "Returns whether the pub is open or closed.",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			data := ai.scale.GetScale()
			if data.Pub.IsOpen {
				return "The pub is open.", nil
			}

			return "The pub is closed.", nil
		},
	}
}

func (ai *Ai) pubOpenedAtTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "pub_open_at",
			Description: "Returns the date and time the pub was opened in Europe/Prague timezone.",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			data := ai.scale.GetScale()
			if !data.Pub.IsOpen {
				return "The pub is closed.", nil
			}

			return data.Pub.OpenedAt, nil
		},
	}
}

func (ai *Ai) currentKegTools() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "current_keg",
			Description: "If there is an active keg, it provides its size in liters",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			data := ai.scale.GetScale()
			if data.ActiveKeg == 0 {
				return "There is no active keg.", nil
			}

			return fmt.Sprintf("<size>%d</size> liter keg is tapped.",
				data.ActiveKeg,
			), nil
		},
	}
}

func (ai *Ai) beersLeftTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "beers_left",
			Description: "Returns the number of beers left in the active keg.",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			data := ai.scale.GetScale()
			if data.ActiveKeg == 0 {
				return "There is no active keg.", nil
			}

			return fmt.Sprintf("%d beers", data.BeersLeft), nil
		},
	}
}

func (ai *Ai) kegTappedAtTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "keg_tapped_at",
			Description: "Returns the date and time the active keg was tapped in Europe/Prague timezone.",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			data := ai.scale.GetScale()
			if data.ActiveKeg == 0 {
				return "There is no active keg.", nil
			}

			return utils.FormatDate(data.ActiveKegAt), nil
		},
	}
}

func (ai *Ai) warehouseTotalTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "warehouse_total",
			Description: "Returns the total number of beers in the warehouse",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			data := ai.scale.GetScale()

			return fmt.Sprintf("%d beers", data.WarehouseBeerLeft), nil
		},
	}
}

func (ai *Ai) warehouseKegTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "warehouse_kegs",
			Description: "Returns amount of kegs in the warehouse for a specific keg size",
			InputSchema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"keg_size": {
						Type:        jsonschema.Integer,
						Enum:        []string{"10", "15", "20", "30", "50"},
						Description: "The size of the keg in liters",
					},
				},
				Required: []string{"keg_size"},
			},
		},
		Fn: func(input string) (string, error) {
			scale := ai.scale.GetScale()

			var data struct {
				KegSize int `json:"keg_size"`
			}

			if err := json.Unmarshal([]byte(input), &data); err != nil {
				return "", fmt.Errorf("error unmarshalling input: %w", err)
			}

			for _, keg := range scale.Warehouse {
				if keg.Keg == data.KegSize {
					return fmt.Sprintf("%d", keg.Amount), nil
				}
			}

			return "This keg size does not exist", fmt.Errorf("keg size %d does not exist", data.KegSize)
		},
	}
}

func (ai *Ai) scaleWifiStrengthTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "scale_wifi_strength",
			Description: "Returns the wifi strength of the scale in dBm.",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			scale := ai.scale.GetScale()

			if !scale.Pub.IsOpen {
				return "The pub is closed and scale is not working", nil
			}

			return fmt.Sprintf("%.2f dBm", scale.Rssi), nil
		},
	}
}

func (ai *Ai) suppliersTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "supplier_beer_price_list",
			Description: "Returns a beer prices for the specific supplier. Response contains a json document with beer title and price. There might be multiple results for a single brand with various sizes and beer types! The structure is flat - it means there is no structure for brands and keg sizes. Title usually contains the size of the keg and the type of beer.",
			InputSchema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"supplier_name": {
						Type:        jsonschema.String,
						Enum:        []string{"baracek", "maneo"},
						Description: "Supplier name",
					},
				},
				Required: []string{"supplier_name"},
			},
		},
		Fn: func(input string) (string, error) {
			var data struct {
				Supplier string `json:"supplier_name"`
			}

			if err := json.Unmarshal([]byte(input), &data); err != nil {
				return "", fmt.Errorf("error unmarshalling input: %w", err)
			}

			var provider shops.Provider
			if data.Supplier == "maneo" {
				provider = &shops.ManeoProvider{}
			} else {
				provider = &shops.BaracekProvider{}
			}

			items, err := provider.GetItems()
			if err != nil {
				return "", fmt.Errorf("could not get items: %w", err)
			}
			j, err := json.Marshal(items)
			if err != nil {
				return "", fmt.Errorf("could not marshal items: %w", err)
			}
			return string(j), nil
		},
	}
}

func (ai *Ai) localNewsTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "local_news_events",
			Description: "Returns local news, events, announcements as a json. Source is the government website.",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			type Entry struct {
				Title    string `json:"title"`
				Summary  string `json:"summary"`
				Created  string `json:"created"`
				Category string `json:"category"`
			}
			type Output struct {
				Title    string  `json:"title"`
				Link     string  `json:"link"`
				Author   string  `json:"author"`
				Language string  `json:"language"`
				Entries  []Entry `json:"entries"`
			}

			resp, err := http.DefaultClient.Get("https://www.vavrinec.cz/api/rss/")
			if err != nil {
				return "", fmt.Errorf("could not get rss feed: %w", err)
			}
			defer resp.Body.Close()

			fp := gofeed.NewParser()
			feed, err := fp.Parse(resp.Body)
			if err != nil {
				return "", fmt.Errorf("could not parse rss feed: %w", err)
			}

			var entries []Entry
			for _, item := range feed.Items {
				if time.Since(*item.PublishedParsed) < 90*24*time.Hour {
					entries = append(entries, Entry{
						Title:    item.Title,
						Summary:  item.Description,
						Created:  item.PublishedParsed.Format(time.RFC3339),
						Category: item.Categories[0],
					})
				}
			}

			output := Output{
				Title:    feed.Title,
				Link:     feed.Link,
				Author:   feed.Description,
				Language: feed.Language,
				Entries:  entries,
			}

			j, err := json.Marshal(output)
			if err != nil {
				return "", fmt.Errorf("could not marshal output: %w", err)
			}

			return string(j), nil
		},
	}
}

// @todo - this does not work and its huge - probably just scrape it from the website
// @todo optimalizations
func (ai *Ai) lesempolemTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "lesempolem_running_results",
			Description: "All results of the Lesempolem running competition from Veselice. The result is JSON combining all years and categories. 'cp' means category place. 'p' means position. 'lt' means lap time. 'c' means category.",
			InputSchema: jsonschema.Definition{
				Type:       jsonschema.Object,
				Properties: map[string]jsonschema.Definition{},
				Required:   []string{""},
			},
		},
		Fn: func(_ string) (string, error) {
			resp, err := http.DefaultClient.Get("https://lesempolem.cz/results_merged.json")
			if err != nil {
				return "", fmt.Errorf("could not get results: %w", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("could not read response body: %w", err)
			}

			result := strings.ReplaceAll(string(body), "\"category_place\"", "\"cp\"")
			result = strings.ReplaceAll(result, "\"place\"", "\"p\"")
			result = strings.ReplaceAll(result, "\"position\"", "\"p\"")
			result = strings.ReplaceAll(result, "\"lapTime\"", "\"lt\"")
			result = strings.ReplaceAll(result, "\"category\"", "\"c\"")

			return result, nil
		},
	}
}

func (ai *Ai) tennisTool() tool {
	return tool{
		Definition: anthropic.ToolDefinition{
			Name:        "tennis_results",
			Description: "Results of the tennis tournament called Veselice Open. You can get results for every tournament separately. The result is a webpage with results.",
			InputSchema: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"tournament_name": {
						Type:        jsonschema.Integer,
						Enum:        []string{"2023-debl", "2024-singl"},
						Description: "The name of the tournament. Usually the year and type of the tournament.",
					},
				},
				Required: []string{"tournament_name"},
			},
		},
		Fn: func(input string) (string, error) {
			var i struct {
				TournamentName string `json:"tournament_name"`
			}

			err := json.Unmarshal([]byte(input), &i)
			if err != nil {
				return "", fmt.Errorf("could not unmarshal input: %w", err)
			}
			tournamentName := i.TournamentName

			data, err := ProvideTennisData(tournamentName)
			if err != nil {
				return "", fmt.Errorf("could not get tennis data: %w", err)
			}

			return data, nil
		},
	}
}
