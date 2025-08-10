package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/prometheus"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/sirupsen/logrus"
)

type OpenAi struct {
	client       openai.Client
	toolsFactory *ToolFactory

	config  *config.Config
	monitor *prometheus.Monitor
	scale   *scale.Scale
	ctx     context.Context
	logger  *logrus.Logger
}

func NewOpenAi(ctx context.Context, conf *config.Config, s *scale.Scale, m *prometheus.Monitor, l *logrus.Logger) *OpenAi {
	return &OpenAi{
		client: openai.NewClient(
			option.WithAPIKey(conf.OpenAiAPIKey),
		),
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

func (ai *OpenAi) GetQuality(quality ModelQuality) string {
	if quality == ModelQualityHigh {
		return openai.ChatModelGPT5
	}

	return openai.ChatModelGPT5Mini
}

func (ai *OpenAi) GetResponse(history []ChatMessage, quality ModelQuality) (Response, error) {
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

	// always build a new param (list of messages for API)
	messages := make([]openai.ChatCompletionMessageParamUnion, len(history)+1)
	messages[0] = openai.SystemMessage(renderPrompt())
	for i, message := range history {
		switch message.From {
		case Me:
			// all other messages from user
			messages[i+1] = openai.UserMessage(message.Text)
		default:
			// replies from assistant
			messages[i+1] = openai.AssistantMessage(message.Text)
		}
	}

	tools := ai.toolsFactory.GetTools(model)

	param := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
		Tools:    ai.convertTools(tools),
	}

	running := true
	sem := 0
	for running && sem < safetyLoopLimit {
		sem++

		running = false
		resp, err := ai.client.Chat.Completions.New(ai.ctx, param)
		if err != nil {
			return output, fmt.Errorf("openai client error: %w", err)
		}

		// add response to the array of messages
		param.Messages = append(param.Messages, resp.Choices[0].Message.ToParam())

		// check for tools and solve them
		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			running = true // tools used, run again
			for _, t := range tools {
				if toolCall.Function.Name == t.Name {
					ai.logger.Infof("running tool %s", t.Name)
					toolResp, err := t.Fn(toolCall.Function.Arguments)
					if err != nil {
						return output, fmt.Errorf("error running tool %s: %w", toolResp, err)
					}
					param.Messages = append(param.Messages, openai.ToolMessage(toolResp, toolCall.ID))
				}
			}
		}

		output.Text = resp.Choices[0].Message.Content

		output.Cost.Input += int(resp.Usage.PromptTokens)
		output.Cost.Output += int(resp.Usage.CompletionTokens)

		ai.monitor.OpenAiInputTokens.WithLabelValues().Add(float64(resp.Usage.PromptTokens))
		ai.monitor.OpenAiOutputTokens.WithLabelValues().Add(float64(resp.Usage.CompletionTokens))

		ai.logger.WithField("billing", "input").Infof("OpenAI input tokens: %d", resp.Usage.PromptTokens)
		ai.logger.WithField("billing", "output").Infof("OpenAI output tokens: %d", resp.Usage.CompletionTokens)
	}

	if output.Text == "" {
		output.Text = "Nemohu odpovědět na tuto otázku."
	}
	return output, nil
}

func (ai *OpenAi) convertTools(tools []Tool) []openai.ChatCompletionToolUnionParam {
	ret := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, t := range tools {
		if t.HasSchema {
			ret[i] = openai.ChatCompletionFunctionTool(
				openai.FunctionDefinitionParam{
					Name:        t.Name,
					Description: openai.String(t.Description),
					Parameters:  ai.convertField(t.Schema),
				},
			)
		} else {
			ret[i] = openai.ChatCompletionFunctionTool(
				openai.FunctionDefinitionParam{
					Name:        t.Name,
					Description: openai.String(t.Description),
				},
			)
		}
	}

	return ret
}

func (ai *OpenAi) convertField(v Property) map[string]interface{} {
	var t string
	switch v.Type {
	case SchemaTypeObject:
		t = "object"
	case SchemaTypeArray:
		t = "array"
	case SchemaTypeBoolean:
		t = "boolean"
	case SchemaTypeInteger:
		t = "integer"
	case SchemaTypeString:
		t = "string"
	default:
		panic("unknown property type")
	}

	ret := make(map[string]interface{})
	ret["type"] = t
	if v.Type == SchemaTypeObject {

		if len(v.Properties) > 0 {
			ret["properties"] = ai.convertFieldMap(v.Properties)
		}

		if v.Description != "" {
			ret["description"] = v.Description
		}
		if len(v.Required) > 0 {
			ret["required"] = v.Required
		}
	} else {
		if v.Description != "" {
			ret["description"] = v.Description
		}
		if len(v.Enum) > 0 {
			ret["enum"] = v.Enum
		}
	}

	return ret
}

func (ai *OpenAi) convertFieldMap(props map[string]Property) map[string]map[string]interface{} {
	ret := make(map[string]map[string]interface{})
	for k, v := range props {
		ret[k] = ai.convertField(v)
	}

	return ret
}
