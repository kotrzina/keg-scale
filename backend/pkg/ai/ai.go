package ai

import (
	"context"
	"fmt"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/sirupsen/logrus"
)

// Prompt is the most important part of the AI. It is the soul of the bot.
// Mr. Botka lives here
const Prompt = `
You are a bot in a pub. Your name is <name>Pan Botka</name> (Mr. Botka in English).
The pub has a keg scale connected to the internet via wifi. 
The scale is used to measure the amount of beer left in the active keg. 
There is always only one active keg.

Facts:
	- current date and time: <datetime>${datetime}</datetime>
	- there is various goods in the pub - beer, non alcoholic drinks, snacks, wine, coffee, tea
	- usual beer - Policka Otakar 11 
	- prices in the pub are fixed: everything is 25 CZK except for a bottle of wine which is 130 CZK
	- we do not sell kegs, only 0.5 liter beers
	- kegs are used only as a supply for the pub
	- existing kegs: 10, 15, 20, 30, 50 liters
	- suppliers: baracek, maneo - always prefer baracek supplier
	- pub is located in the small village call Veselice in Moravsky Kras in the Czech Republic
	- the nearest town is Blansko
	- Veselice is a feminine word in Czech language
	- contact person for intimate questions (hookes, sex, private) is Radim Boháč
	- there is a table tennis team representing Veselice in the district league (okresni soutez)
	- we organize running competition called Lesempolem every year

Generate a response to the following message:
<message>
${msg}
</message>

The answer will be brief and clear. Always in Czech language. No XML tags. Do not moralize guests.
Czech synonyms for beer keg: bečka = sud = keg
Preferred wording: hospoda, bečka. 
Provide only verified and up-to-date information. If unsure, use the appropriate tool instead of making assumptions.
`

type Provider interface {
	GetResponse(history []ChatMessage) (Response, error)
}

type Ai struct {
	providers map[string]Provider
}

func NewAi(ctx context.Context, conf *config.Config, s *scale.Scale, m *prometheus.Monitor, l *logrus.Logger) *Ai {
	return &Ai{
		providers: map[string]Provider{
			"openai":    NewOpenAi(ctx, conf, s, m, l),
			"anthropic": NewAnthropic(ctx, conf, s, m, l),
		},
	}
}

func (ai *Ai) GetResponse(history []ChatMessage) (Response, error) {
	const providerName = "anthropic"
	p, ok := ai.providers[providerName]
	if !ok {
		return Response{}, fmt.Errorf("unknown provider: %s", providerName)
	}

	return p.GetResponse(history)
}

type ChatMessage struct {
	Text string `json:"text"`
	From string `json:"from"` // me means the user. Anything else is the assistant
}

type Response struct {
	Text string `json:"text"`
	Cost Cost   `json:"cost"`
}

type Cost struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type SchemaType uint8

const (
	SchemaTypeObject SchemaType = iota
	SchemaTypeArray
	SchemaTypeBoolean
	SchemaTypeInteger
	SchemaTypeString
)

type Tool struct {
	Name        string
	Description string
	HasSchema   bool
	Schema      Property

	Fn func(string) (string, error)
}

type Property struct {
	Type        SchemaType
	Description string
	Properties  map[string]Property
	Enum        []interface{} // depends on the type
	Required    []string
}

// GetEnumAsStrings returns the Enum field as a slice of strings.
// it is useful for services which does support strings only (like Anthropic)
// basically we convert all values to strings
func (d *Property) GetEnumAsStrings() []string {
	if d.Enum == nil {
		return nil
	}

	ret := make([]string, len(d.Enum))
	for i, v := range d.Enum {
		ret[i] = fmt.Sprint(v)
	}

	return ret
}

const Me = "me" // user
