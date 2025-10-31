package ai

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/store"
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/sirupsen/logrus"
)

const providerName = "openai"

const safetyLoopLimit = 10

type ModelQuality uint8

const (
	ModelQualityLow ModelQuality = iota
	ModelQualityMedium
	ModelQualityHigh
)

// prompt is the most important part of the AI. It is the soul of the bot.
// Mr. Botka lives here
//
//go:embed prompts/ai.prompt
var prompt string

// customOpenPrompt custom opening message prompt
//
//go:embed prompts/custom_open.prompt
var customOpenPrompt string

// customGeneralMessage general opening message prompt
//
//go:embed prompts/general_open.prompt
var customGeneralMessage string

// customGeneralMessage general opening message prompt
//
//go:embed prompts/regulars_request.prompt
var regularsRequestPrompt string

//go:embed prompts/volleyball.prompt
var volleyballMessage string

func renderPrompt() string {
	renderedPrompt := strings.ReplaceAll(prompt, "${datetime}", utils.FormatDate(time.Now()))

	return renderedPrompt
}

type Provider interface {
	GetResponse(history []ChatMessage, quality ModelQuality) (Response, error)
	GetQuality(quality ModelQuality) string
}

type Ai struct {
	providers map[string]Provider
	storage   store.Storage
	logger    *logrus.Logger
}

func NewAi(ctx context.Context, conf *config.Config, s *scale.Scale, m *prometheus.Monitor, storage store.Storage, l *logrus.Logger) *Ai {
	return &Ai{
		providers: map[string]Provider{
			"openai":    NewOpenAi(ctx, conf, s, m, l),
			"anthropic": NewAnthropic(ctx, conf, s, m, l),
		},
		storage: storage,
		logger:  l,
	}
}

func (ai *Ai) GetResponse(history []ChatMessage, quality ModelQuality) (Response, error) {
	p, ok := ai.providers[providerName]
	if !ok {
		return Response{}, fmt.Errorf("unknown provider: %s", providerName)
	}

	return p.GetResponse(history, quality)
}

// GenerateGeneralOpenMessage generates a message for group WhatsApp chat
func (ai *Ai) GenerateGeneralOpenMessage() (string, error) {
	req := strings.Builder{}

	// check if today we have a special beer
	beer, err := ai.storage.GetTodayBeer()
	if err == nil && beer != "" {
		req.WriteString(fmt.Sprintf(" - must contain an information that today we have %s beer on tap", beer))
		if err := ai.storage.ResetTodayBeer(); err != nil {
			return "", fmt.Errorf("failed to reset today beer: %w", err)
		}
	}

	templatedPrompt := strings.ReplaceAll(customGeneralMessage, "${requirements}", req.String())
	fmt.Println(templatedPrompt)
	messages := []ChatMessage{
		{
			From: Me,
			Text: templatedPrompt,
		},
	}

	resp, err := ai.GetResponse(messages, ModelQualityHigh)
	if err != nil {
		return "", fmt.Errorf("failed to generate open message %w", err)
	}

	return resp.Text, nil
}

// GenerateRegularsMessage generates a message for regulars that somebody wants to go to the pub today
func (ai *Ai) GenerateRegularsMessage(msg string) (string, error) {
	templated := strings.ReplaceAll(regularsRequestPrompt, "${message}", msg)
	messages := []ChatMessage{
		{
			From: Me,
			Text: templated,
		},
	}

	resp, err := ai.GetResponse(messages, ModelQualityHigh)
	if err != nil {
		return "", fmt.Errorf("failed to generate regulars request message %w", err)
	}

	return resp.Text, nil
}

// GenerateCustomOpenMessage generates a custom open message
// for the user with the given name
func (ai *Ai) GenerateCustomOpenMessage(name string) (string, error) {
	firstMsg := strings.ReplaceAll(customOpenPrompt, "${name}", name)
	messages := []ChatMessage{
		{
			From: Me,
			Text: firstMsg,
		},
	}

	resp, err := ai.GetResponse(messages, ModelQualityMedium)
	if err != nil {
		return "", fmt.Errorf("failed to generate open message for %s: %w", name, err)
	}

	return resp.Text, nil
}

// GenerateVolleyballMessage generates a volleyball message
func (ai *Ai) GenerateVolleyballMessage() (string, error) {
	messages := []ChatMessage{
		{
			From: Me,
			Text: volleyballMessage,
		},
	}

	resp, err := ai.GetResponse(messages, ModelQualityHigh)
	if err != nil {
		return "", fmt.Errorf("failed to generate volleyball message: %w", err)
	}

	return resp.Text, nil
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
