package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/sirupsen/logrus"
)

type Ai struct {
	client *anthropic.Client

	config  *config.Config
	monitor *prometheus.Monitor
	scale   *scale.Scale
	ctx     context.Context
	logger  *logrus.Logger

	tools []tool
}

type tool struct {
	Definition anthropic.ToolDefinition
	Fn         func(input string) (string, error)
}

func NewAi(ctx context.Context, conf *config.Config, s *scale.Scale, m *prometheus.Monitor, l *logrus.Logger) *Ai {
	ai := &Ai{
		client: anthropic.NewClient(conf.AnthropicAPIKey),

		config:  conf,
		monitor: m,
		scale:   s,
		ctx:     ctx,
		logger:  l,
	}

	tools := []tool{
		ai.currentTimeTool(),

		ai.isOpenTool(),
		ai.pubOpenedAtTool(),

		ai.currentKegTools(),
		ai.beersLeftTool(),
		ai.kegTappedAtTool(),

		ai.warehouseTotalTool(),
		ai.warehouseKegTool(),

		ai.scaleWifiStrengthTool(),

		ai.suppliersTool(),

		ai.localNewsTool(),
	}
	ai.tools = tools

	return ai
}

const Prompt = `
All communication will be in Czech language.
czech synonyms for beer keg: bečka = sud = keg
Preferred wording: hospoda, bečka.
Functions: Pub provides various functions and pubic data such as:
	- how many beers are left in the active (tapped) keg
	- is pub open
	- warehouse state (how many beers are in the warehouse) - separated by keg size
	- price list for various suppliers
	- wifi signal strength
	- information from village (news, events, local government announcements)
	- calculate bill for multiple guests at the same time
Facts:
	- there is various sentiment in the pub - we sell beer, non alcoholic drinks, snacks, wine, coffee, tea
	- prices in the pub are fixed: everything is 25 Kč except for a bottle of wine which is 130 Kč
	- you can get total price for a specific amount of beers by multiplying the price by the amount of beers
	- we do not sell kegs, only 0.5 liter beers
	- keg are used only as a supply for the pub
	- existing kegs: 10, 15, 20, 30, 50 liters
	- at the moment, the pub has only one active keg
	- scale is connected to the internet via wifi
	- suppliers: baracek, maneo
	- always prefer baracek supplier
	- pub is located in the small village of Veselice

Generate a response to the following message:
<message>
${msg}
</message>

The answer will be brief and clear. Always in Czech. No XML tags.

For supplier price list try to find all keg sizes. If you can't find the price for a specific keg size, return a message that the price is not available.
`

type ChatMessage struct {
	Text string `json:"text"`
	From string `json:"from"` // me means the user. Anything else is the assistant
}

const Me = "me" // user

func (ai *Ai) GetResponse(history []ChatMessage) (string, error) {
	if len(history) == 0 {
		return "", errors.New("no messages")
	}

	messages := make([]anthropic.Message, len(history))
	for i, message := range history {
		switch {
		case message.From == Me && i == 0:
			// first message from user is special
			// we want to use full Prompt
			messages[i] = anthropic.NewUserTextMessage(strings.ReplaceAll(Prompt, "${msg}", message.Text))
		case message.From == Me:
			// all other messages from user
			messages[i] = anthropic.NewUserTextMessage(message.Text)
		default:
			// replies from assistant
			messages[i] = anthropic.NewAssistantTextMessage(message.Text)
		}
	}

	for _, msg := range messages {
		ai.logger.WithField("message", msg.Content).Info("Anthropic message")
	}

	running := true
	sem := 0
	lastMessage := ""

	for running && sem < 10 {
		sem++

		requestTools := make([]anthropic.ToolDefinition, len(ai.tools))
		for i, tool := range ai.tools {
			requestTools[i] = tool.Definition
		}
		resp, err := ai.client.CreateMessages(ai.ctx, anthropic.MessagesRequest{
			Model:     anthropic.ModelClaude3Dot5SonnetLatest,
			Messages:  messages,
			MaxTokens: 1000,
			Tools:     requestTools,
		})
		if err != nil {
			var e *anthropic.APIError
			if errors.As(err, &e) {
				return "", fmt.Errorf("messages error, type: %s, message: %s", e.Type, e.Message)
			}

			return "", fmt.Errorf("messages error: %w", err)
		}

		messages = append(messages, anthropic.Message{
			Role:    anthropic.RoleAssistant,
			Content: resp.Content,
		})

		// solve all requested tools from the response and push results back to the messages
		if resp.StopReason == anthropic.MessagesStopReasonToolUse {
			// combined response for all tools
			toolsResponse := anthropic.Message{
				Role:    anthropic.RoleUser,
				Content: []anthropic.MessageContent{},
			}

			// find all requested tool to solve
			for _, content := range resp.Content {
				requestedTool := content.MessageContentToolUse
				if requestedTool != nil {
					for _, aiTool := range ai.tools {
						if aiTool.Definition.Name == requestedTool.Name {
							toolResponse, err := aiTool.Fn(string(requestedTool.Input))
							if err != nil {
								return "", fmt.Errorf("error running tool %s: %w", requestedTool.Name, err)
							}
							toolsResponse.Content = append(
								toolsResponse.Content,
								anthropic.NewToolResultMessageContent(requestedTool.ID, toolResponse, err != nil),
							)
						}
					}
				}
			}

			messages = append(messages, toolsResponse)
		}

		if resp.StopReason != anthropic.MessagesStopReasonToolUse {
			running = false
		}

		if len(resp.Content) > 0 {
			lastMessage = resp.Content[len(resp.Content)-1].GetText()
		}

		ai.monitor.AnthropicInputTokens.WithLabelValues().Add(float64(resp.Usage.InputTokens))
		ai.monitor.AnthropicOutputTokens.WithLabelValues().Add(float64(resp.Usage.OutputTokens))

		ai.logger.WithField("billing", "input").Infof("Anthropic input tokens: %d", resp.Usage.InputTokens)
		ai.logger.WithField("billing", "output").Infof("Anthropic output tokens: %d", resp.Usage.OutputTokens)
	}

	if lastMessage == "" {
		lastMessage = "Nemohu odpovědět na tuto otázku."
	}
	return lastMessage, nil
}