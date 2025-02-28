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
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/sirupsen/logrus"
)

const providerName = "openai"

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

func renderPrompt() string {
	renderedPrompt := strings.ReplaceAll(prompt, "${datetime}", utils.FormatDate(time.Now()))

	return renderedPrompt
}

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
	p, ok := ai.providers[providerName]
	if !ok {
		return Response{}, fmt.Errorf("unknown provider: %s", providerName)
	}

	return p.GetResponse(history)
}

// GenerateGeneralOpenMessage generates a message for group WhatsApp chat
func (ai *Ai) GenerateGeneralOpenMessage() (string, error) {
	messages := []ChatMessage{
		{
			From: Me,
			Text: customGeneralMessage,
		},
	}

	resp, err := ai.GetResponse(messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate open message %w", err)
	}

	return resp.Text, nil
}

// GenerateCustomOpenMessage generates a custom open message
// for the user with the given name
// Is creates a
func (ai *Ai) GenerateCustomOpenMessage(name string) (string, error) {
	firstMsg := strings.ReplaceAll(customOpenPrompt, "${name}", name)
	messages := []ChatMessage{
		{
			From: Me,
			Text: firstMsg,
		},
	}

	resp, err := ai.GetResponse(messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate open message for %s: %w", name, err)
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
