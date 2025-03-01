package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/jsonschema"
	"github.com/sirupsen/logrus"
)

type Anthropic struct {
	client       *anthropic.Client
	toolsFactory *ToolFactory

	config  *config.Config
	monitor *prometheus.Monitor
	scale   *scale.Scale
	ctx     context.Context
	logger  *logrus.Logger
}

func NewAnthropic(ctx context.Context, conf *config.Config, s *scale.Scale, m *prometheus.Monitor, l *logrus.Logger) *Anthropic {
	return &Anthropic{
		client: anthropic.NewClient(conf.AnthropicAPIKey),
		toolsFactory: &ToolFactory{
			scale:  s,
			config: conf,
			logger: l,
		},

		config:  conf,
		monitor: m,
		scale:   s,
		ctx:     ctx,
		logger:  l,
	}
}

func (ai *Anthropic) GetQuality(_ ModelQuality) string {
	// we use sonnet only
	return string(anthropic.ModelClaude3Dot5SonnetLatest)
}

func (ai *Anthropic) GetResponse(history []ChatMessage, quality ModelQuality) (Response, error) {
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

	model := ai.GetQuality(quality)

	messages := make([]anthropic.Message, len(history))
	for i, message := range history {
		switch {
		case message.From == Me:
			// all other messages from user
			messages[i] = anthropic.NewUserTextMessage(message.Text)
		default:
			// replies from assistant
			messages[i] = anthropic.NewAssistantTextMessage(message.Text)
		}
	}

	tools := ai.toolsFactory.GetTools()
	toolDefinitions := ai.getToolsDefinitions(tools)

	running := true
	sem := 0
	for running && sem < safetyLoopLimit {
		sem++

		resp, err := ai.client.CreateMessages(ai.ctx, anthropic.MessagesRequest{
			Model:     anthropic.Model(model),
			System:    renderPrompt(),
			Messages:  messages,
			MaxTokens: 1000,
			Tools:     toolDefinitions,
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
						if aiTool.Name == requestedTool.Name {
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

func (ai *Anthropic) getToolsDefinitions(tools []Tool) []anthropic.ToolDefinition {
	at := make([]anthropic.ToolDefinition, len(tools))
	for i, t := range tools {
		at[i] = anthropic.ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: ai.convertDefinition(t.Schema),
		}
	}

	return at
}

func (ai *Anthropic) convertDefinition(d Property) jsonschema.Definition {
	t := jsonschema.Object
	switch d.Type {
	case SchemaTypeObject:
		t = jsonschema.Object
	case SchemaTypeArray:
		t = jsonschema.Array
	case SchemaTypeBoolean:
		t = jsonschema.Boolean
	case SchemaTypeInteger:
		t = jsonschema.Integer
	case SchemaTypeString:
		t = jsonschema.String
	}

	ret := jsonschema.Definition{
		Type:        t,
		Description: d.Description,
	}

	if len(d.Enum) > 0 {
		ret.Enum = d.GetEnumAsStrings()
	}

	if len(d.Required) > 0 {
		ret.Required = d.Required
	}

	if d.Properties != nil {
		ret.Properties = make(map[string]jsonschema.Definition, len(d.Properties))
		for k, v := range d.Properties {
			ret.Properties[k] = ai.convertDefinition(v)
		}
	}

	return ret
}
