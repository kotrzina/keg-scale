package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ToolFactory struct {
	scale  *scale.Scale
	config *config.Config
	logger *logrus.Logger
}

func (tf *ToolFactory) GetTools() []Tool {
	tools := []Tool{
		tf.currentTimeTool(),
		tf.isOpenTool(),
		tf.pubOpenedAtTool(),
		tf.pubClosedAtTool(),
		tf.currentKegTools(),
		tf.beersLeftTool(),
		tf.kegTappedAtTool(),
		tf.warehouseTotalTool(),
		tf.warehouseKegTool(),
		tf.scaleWifiStrengthTool(),
		tf.suppliersTool(),
		tf.localNewsTool(),
		tf.tennisTool(),
		tf.lunchMenuTool(),
		tf.eventBlanskoTool(),
		tf.siestaMenuTool(),
		tf.weatherTool(),
		tf.sdhEventsTool(),
		tf.tableTennisResultsTool(),
		tf.tableTennisTableTool(),
		tf.lesempolemRegisteredTool(),
		tf.musicConcertsTool(),
		tf.pubCalendarTool(),
	}

	// concat with static tools
	staticTools, err := tf.StaticTools()
	if err != nil {
		// ignore the error
		// request will be handled without static tools
		tf.logger.Errorf("could not get static tools: %v", err)
	} else {
		return slices.Concat(tools, staticTools)
	}

	return tools
}

func (tf *ToolFactory) currentTimeTool() Tool {
	return Tool{
		Name:        "current_time",
		Description: "Provides current day, date and time in Europe/Prague timezone",
		Fn: func(_ string) (string, error) {
			now := time.Now()
			return fmt.Sprintf("%s, %s", utils.FormatWeekday(now), utils.FormatDate(now)), nil
		},
	}
}

func (tf *ToolFactory) isOpenTool() Tool {
	return Tool{
		Name:        "is_pub_open",
		Description: "Returns whether the pub is open or closed.",
		Fn: func(_ string) (string, error) {
			data := tf.scale.GetScale()
			if data.Pub.IsOpen {
				return "The pub is open.", nil
			}

			return "The pub is closed.", nil
		},
	}
}

func (tf *ToolFactory) pubOpenedAtTool() Tool {
	return Tool{
		Name:        "pub_open_at",
		Description: "Returns the date and time the pub was opened in Europe/Prague timezone.",
		Fn: func(_ string) (string, error) {
			data := tf.scale.GetScale()
			if !data.Pub.IsOpen {
				return "The pub is closed.", nil
			}

			return data.Pub.OpenedAt, nil
		},
	}
}

func (tf *ToolFactory) pubClosedAtTool() Tool {
	return Tool{
		Name:        "pub_close_at",
		Description: "Returns the date and time when the pub was closed last.",
		Fn: func(_ string) (string, error) {
			data := tf.scale.GetScale()
			if data.Pub.IsOpen {
				return "The pub is open.", nil
			}

			return data.Pub.ClosedAt, nil
		},
	}
}

func (tf *ToolFactory) currentKegTools() Tool {
	return Tool{
		Name:        "current_keg",
		Description: "If there is an active keg, it provides its size in liters",
		Fn: func(_ string) (string, error) {
			data := tf.scale.GetScale()
			if data.ActiveKeg == 0 {
				return "There is no active keg.", nil
			}

			return fmt.Sprintf("<size>%d</size> liter keg is tapped.",
				data.ActiveKeg,
			), nil
		},
	}
}

func (tf *ToolFactory) beersLeftTool() Tool {
	return Tool{
		Name:        "beers_left",
		Description: "Returns the number of beers left in the active keg.",
		Fn: func(_ string) (string, error) {
			data := tf.scale.GetScale()
			if data.ActiveKeg == 0 {
				return "There is no active keg.", nil
			}

			return fmt.Sprintf("%d beers", data.BeersLeft), nil
		},
	}
}

func (tf *ToolFactory) kegTappedAtTool() Tool {
	return Tool{
		Name:        "keg_tapped_at",
		Description: "Returns the date and time the active keg was tapped in Europe/Prague timezone.",
		Fn: func(_ string) (string, error) {
			data := tf.scale.GetScale()
			if data.ActiveKeg == 0 {
				return "There is no active keg.", nil
			}

			return utils.FormatDate(data.ActiveKegAt), nil
		},
	}
}

func (tf *ToolFactory) warehouseTotalTool() Tool {
	return Tool{
		Name:        "warehouse_total",
		Description: "Returns the total number of beers in the warehouse",
		Fn: func(_ string) (string, error) {
			data := tf.scale.GetScale()

			return fmt.Sprintf("%d beers", data.WarehouseBeerLeft), nil
		},
	}
}

func (tf *ToolFactory) warehouseKegTool() Tool {
	return Tool{
		Name:        "warehouse_kegs",
		Description: "Returns amount of kegs in the warehouse for a specific keg size",
		HasSchema:   true,
		Schema: Property{
			Type: SchemaTypeObject,
			Properties: map[string]Property{
				"keg_size": {
					Type:        SchemaTypeString,
					Enum:        []interface{}{"10", "15", "20", "30", "50"},
					Description: "The size of the keg in liters",
				},
			},
			Required: []string{"keg_size"},
		},
		Fn: func(input string) (string, error) {
			s := tf.scale.GetScale()

			var data struct {
				KegSize string `json:"keg_size"`
			}

			if err := json.Unmarshal([]byte(input), &data); err != nil {
				return "", fmt.Errorf("error unmarshalling input: %w", err)
			}

			kegSize, err := strconv.Atoi(data.KegSize)
			if err != nil {
				return "Invalid keg size", fmt.Errorf("invalid keg size: %w", err)
			}

			for _, keg := range s.Warehouse {
				if keg.Keg == kegSize {
					return fmt.Sprintf("%d", keg.Amount), nil
				}
			}

			return "This keg size does not exist", fmt.Errorf("keg size %s does not exist", data.KegSize)
		},
	}
}

func (tf *ToolFactory) scaleWifiStrengthTool() Tool {
	return Tool{
		Name:        "scale_wifi_strength",
		Description: "Returns the wifi strength of the scale in dBm.",
		Fn: func(_ string) (string, error) {
			s := tf.scale.GetScale()

			if !s.Pub.IsOpen {
				return "The pub is closed and scale is not working", nil
			}

			return fmt.Sprintf("%.2f dBm", s.Rssi), nil
		},
	}
}

func (tf *ToolFactory) suppliersTool() Tool {
	return Tool{
		Name:        "supplier_beer_price_list",
		Description: "Provides beer prices list for the specific supplier. Response contains a json document with beer title and price. There might be multiple results for a single brand with various sizes and beer types! The structure is flat - it means there is no structure for brands and keg sizes. Title usually contains the size of the keg and the type of beer. Always try to find all sizes and prices.",
		HasSchema:   true,
		Schema: Property{
			Type: SchemaTypeObject,
			Properties: map[string]Property{
				"supplier_name": {
					Type:        SchemaTypeString,
					Enum:        []interface{}{"baracek", "maneo"},
					Description: "Supplier name",
				},
			},
			Required: []string{"supplier_name"},
		},
		Fn: func(input string) (string, error) {
			var data struct {
				Supplier string `json:"supplier_name"`
			}

			if err := json.Unmarshal([]byte(input), &data); err != nil {
				return "", fmt.Errorf("error unmarshalling input: %w", err)
			}

			var provider beerProvider
			if data.Supplier == "maneo" {
				provider = &ManeoProvider{}
			} else {
				provider = &BaracekProvider{}
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

func (tf *ToolFactory) localNewsTool() Tool {
	return Tool{
		Name:        "local_news_events",
		Description: "Returns local news, events, announcements as a json. Source is the government website.",
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

func (tf *ToolFactory) tennisTool() Tool {
	return Tool{
		Name:        "tennis_results",
		Description: "Results of the tennis tournament called Veselice Open. You can get results for every tournament separately. The result is a webpage with results. Page also contains links to the photo library.",
		HasSchema:   true,
		Schema: Property{
			Type: SchemaTypeObject,
			Properties: map[string]Property{
				"tournament_name": {
					Type:        SchemaTypeString,
					Enum:        []interface{}{"2023-debl", "2024-singl"},
					Description: "The name of the tournament. Usually the year and type of the tournament.",
				},
			},
			Required: []string{"tournament_name"},
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

func (tf *ToolFactory) lunchMenuTool() Tool {
	return Tool{
		Name:        "lunch_menu",
		Description: "Provides lunch menu for the restaurants nearby (Restaurace Obůrka, Hotel Broušek, Penzion U Hrabenky). The result is a webpage with the menu. The page might be outdated for current week. Restaurants usually provide unique meals for each day. They also might have some meals for the whole week.",
		HasSchema:   true,
		Schema: Property{
			Type: SchemaTypeObject,
			Properties: map[string]Property{
				"restaurant_name": {
					Type:        SchemaTypeString,
					Enum:        []interface{}{"oburka", "brousek", "hrabenka"},
					Description: "The name of the restaurant",
				},
			},
			Required: []string{"restaurant_name"},
		},
		Fn: func(input string) (string, error) {
			var i struct {
				RestaurantName string `json:"restaurant_name"`
			}

			err := json.Unmarshal([]byte(input), &i)
			if err != nil {
				return "", fmt.Errorf("could not unmarshal input: %w", err)
			}
			restaurant := i.RestaurantName

			var data string
			switch restaurant {
			case "oburka":
				data, err = ProvideOburkaMenu()
			case "brousek":
				data, err = ProvideBrousekMenu()
			case "hrabenka":
				data, err = ProvideHrabenkaMenu()
			default:
				return "", fmt.Errorf("unknown restaurant: %s", restaurant)
			}

			if err != nil {
				return "", fmt.Errorf("could not get lunch menu: %w", err)
			}

			return data, nil
		},
	}
}

func (tf *ToolFactory) eventBlanskoTool() Tool {
	return Tool{
		Name:        "events_blansko",
		Description: "Provides culture events in Blansko. Tool returns json structure. The source is the website akceblansko.cz. Includes timetable of the cinema, concerts, etc. Make sure to check exact dates for events and provide exact event names/movies",
		Fn: func(_ string) (string, error) {
			data, err := ProvideEventsBlansko()
			if err != nil {
				return "", fmt.Errorf("could not get events: %w", err)
			}

			return data, nil
		},
	}
}

func (tf *ToolFactory) siestaMenuTool() Tool {
	return Tool{
		Name:        "siesta_pizza",
		Description: "Provides pizza menu from Siesta Pizza. The Siesta is usually used for the pizza delivery in our pub. Pizza is always 32 cm.",
		Fn: func(_ string) (string, error) {
			output, err := ProvideSiestaMenu()
			if err != nil {
				return "", fmt.Errorf("could not get menu: %w", err)
			}

			return output, nil
		},
	}
}

func (tf *ToolFactory) weatherTool() Tool {
	return Tool{
		Name:        "weather",
		Description: "Provides current weather in Veselice. Json contains hourly forecast and contains temperature, participation, wind speed, cloudiness and wind direction. It also containers forecast for the next 3 days.",
		Fn: func(_ string) (string, error) {
			output, err := ProvideWeather()
			if err != nil {
				return "", fmt.Errorf("could not get weather: %w", err)
			}

			return output, nil
		},
	}
}

func (tf *ToolFactory) sdhEventsTool() Tool {
	return Tool{
		Name:        "sdh_events",
		Description: "Provides news and events from the local fire department (SDH Veselice). The source is the website sdhveselice.cz.",
		Fn: func(_ string) (string, error) {
			output, err := ProvideSdhEvents()
			if err != nil {
				return "", fmt.Errorf("could not get sdh events: %w", err)
			}

			return output, nil
		},
	}
}

func (tf *ToolFactory) tableTennisResultsTool() Tool {
	return Tool{
		Name:        "table_tennis_results",
		Description: "Provides matches schedule and results of the table tennis league.",
		Fn: func(_ string) (string, error) {
			output, err := ProvideTableTennisResults()
			if err != nil {
				return "", fmt.Errorf("could not get table tennis data: %w", err)
			}

			return output, nil
		},
	}
}

func (tf *ToolFactory) tableTennisTableTool() Tool {
	return Tool{
		Name:        "table_tennis_league_table",
		Description: "Provides leaderboard table of the tennis league.",
		Fn: func(_ string) (string, error) {
			output, err := ProvideTableTennisLeagueTable()
			if err != nil {
				return "", fmt.Errorf("could not get table tennis data: %w", err)
			}

			return output, nil
		},
	}
}

func (tf *ToolFactory) lesempolemRegisteredTool() Tool {
	return Tool{
		Name:        "running_lesempolem_registered",
		Description: "Provides list of registered runners for the Lesempolem run with going to happen on 2025-05-10",
		Fn: func(_ string) (string, error) {
			output, err := ProviderLesempolemRegistered()
			if err != nil {
				return "", fmt.Errorf("could not get lesempolem data: %w", err)
			}

			return output, nil
		},
	}
}

func (tf *ToolFactory) musicConcertsTool() Tool {
	return Tool{
		Name:        "music_concerts",
		Description: "Provides music concerts and festivals which might be interesting for the pub visitors",
		Fn: func(_ string) (string, error) {
			output, err := ProvideCalendar(tf.config.CalendarConcertsURL, time.Now().Add(-30*24*time.Hour), time.Now().Add(365*24*time.Hour))
			if err != nil {
				return "", fmt.Errorf("could not get concerts calendar data: %w", err)
			}

			return output, nil
		},
	}
}

func (tf *ToolFactory) pubCalendarTool() Tool {
	return Tool{
		Name:        "pub_calendar",
		Description: "Calendar of events related to the pub or Veselice village (e.g. birthdays, parties, events, tap sanitizations, special beer days, etc.)",
		Fn: func(_ string) (string, error) {
			output, err := ProvideCalendar(tf.config.CalendarPubURL, time.Now().Add(-60*24*time.Hour), time.Now().Add(365*24*time.Hour))
			if err != nil {
				return "", fmt.Errorf("could not get pub calendar data: %w", err)
			}

			return output, nil
		},
	}
}

type StaticConfig struct {
	Tools []struct {
		Name        string `yaml:"name"`
		Type        string `yaml:"type"`
		Description string `yaml:"description"`
		Result      string `yaml:"result"`
	} `yaml:"tools"`
}

func (tf *ToolFactory) StaticTools() ([]Tool, error) {
	url := "https://static.kozak.in/pub-ai/static.yml"
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get static tools: %w", err)
	}

	defer resp.Body.Close()

	staticContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	var sConfig StaticConfig
	if err := yaml.Unmarshal(staticContent, &sConfig); err != nil {
		return nil, fmt.Errorf("could not unmarshal StaticConfig content: %w", err)
	}

	tools := make([]Tool, len(sConfig.Tools))
	for i, t := range sConfig.Tools {
		tools[i] = Tool{
			Name:        t.Name,
			Description: t.Description,
			Fn: func(_ string) (string, error) {
				return t.Result, nil
			},
		}
	}

	return tools, nil
}
