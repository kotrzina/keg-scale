package ai

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/utils"
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

	ai.tools = []tool{
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
		ai.tennisTool(),
		ai.lunchMenuTool(),
		ai.eventBlanskoTool(),
		ai.siestaMenuTool(),
		ai.weatherTool(),
		ai.sdhEventsTool(),
	}

	return ai
}

const Prompt = `
You are a bot in a pub. Your name is <name>Pan Botka</name> (Mr. Botka in English).
The pub has a keg scale connected to the internet via wifi. The scale is used to measure the amount of beer left in the active keg. There is always only one active keg.

Facts:
	- current date and time: <datetime>${datetime}</datetime>
	- there is various sentiment in the pub - beer, non alcoholic drinks, snacks, wine, coffee, tea
	- usually we have Policka - Otakar 11 beer
	- prices in the pub are fixed: everything is 25 Kč except for a bottle of wine which is 130 Kč
	- we do not sell kegs, only 0.5 liter beers
	- kegs are used only as a supply for the pub
	- existing kegs: 10, 15, 20, 30, 50 liters
	- suppliers: baracek, maneo - always prefer baracek supplier
	- pub is located in the small village call Veselice
	- Veselice is a village in Moravsky Kras in the Czech Republic
	- the nearest town is Blansko
	- Veselice is a feminine word in Czech language

Generate a response to the following message:
<message>
${msg}
</message>

The answer will be brief and clear. Always in Czech language. No XML tags.
Czech synonyms for beer keg: bečka = sud = keg
Preferred wording: hospoda, bečka.
For supplier price list try to find all keg sizes. If you can't find the price for a specific keg size, return a message that the price is not available.
`

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

const Me = "me" // user

func (ai *Ai) GetResponse(history []ChatMessage) (Response, error) {
	output := Response{
		Text: "",
		Cost: Cost{
			Input:  0,
			Output: 0,
		},
	}

	if len(history) == 0 {
		return output, errors.New("no messages")
	}

	messages := make([]anthropic.Message, len(history))
	for i, message := range history {
		switch {
		case message.From == Me && i == 0:
			// first message from user is special
			// we want to use full Prompt
			m := strings.ReplaceAll(Prompt, "${msg}", message.Text)
			m = strings.ReplaceAll(m, "${datetime}", utils.FormatDate(time.Now()))
			messages[i] = anthropic.NewUserTextMessage(m)
		case message.From == Me:
			// all other messages from user
			messages[i] = anthropic.NewUserTextMessage(message.Text)
		default:
			// replies from assistant
			messages[i] = anthropic.NewAssistantTextMessage(message.Text)
		}
	}

	staticTools, err := ai.staticTools()
	if err != nil {
		ai.logger.Errorf("could not load StaticConfig tools: %v", err)
		return output, fmt.Errorf("could not load StaticConfig tools: %w", err)
	}

	tools := slices.Concat(ai.tools, staticTools) // merge default and static tools

	running := true
	sem := 0
	for running && sem < 10 {
		sem++

		requestTools := make([]anthropic.ToolDefinition, len(tools))
		for i, tool := range tools {
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
				return output, fmt.Errorf("messages error, type: %s, message: %s", e.Type, e.Message)
			}

			return output, fmt.Errorf("messages error: %w", err)
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
					for _, aiTool := range tools {
						if aiTool.Definition.Name == requestedTool.Name {
							ai.logger.Infof("running tool %s", requestedTool.Name)
							toolResponse, err := aiTool.Fn(string(requestedTool.Input))
							if err != nil {
								return output, fmt.Errorf("error running tool %s: %w", requestedTool.Name, err)
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
			output.Text = resp.Content[len(resp.Content)-1].GetText()
		}

		output.Cost.Output += resp.Usage.OutputTokens
		output.Cost.Input += resp.Usage.InputTokens

		ai.monitor.AnthropicInputTokens.WithLabelValues().Add(float64(resp.Usage.InputTokens))
		ai.monitor.AnthropicOutputTokens.WithLabelValues().Add(float64(resp.Usage.OutputTokens))

		ai.logger.WithField("billing", "input").Infof("Anthropic input tokens: %d", resp.Usage.InputTokens)
		ai.logger.WithField("billing", "output").Infof("Anthropic output tokens: %d", resp.Usage.OutputTokens)
	}

	if output.Text == "" {
		output.Text = "Nemohu odpovědět na tuto otázku."
	}
	return output, nil
}
