package ai

import (
	"fmt"
	"io"
	"net/http"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/liushuangls/go-anthropic/v2/jsonschema"
	"gopkg.in/yaml.v3"
)

type StaticConfig struct {
	Tools []struct {
		Name        string `yaml:"name"`
		Type        string `yaml:"type"`
		Description string `yaml:"description"`
		Result      string `yaml:"result"`
	} `yaml:"tools"`
}

func (ai *Ai) staticTools() ([]tool, error) {
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

	var config StaticConfig
	if err := yaml.Unmarshal(staticContent, &config); err != nil {
		return nil, fmt.Errorf("could not unmarshal StaticConfig content: %w", err)
	}

	tools := make([]tool, len(config.Tools))
	for i, t := range config.Tools {
		tools[i] = tool{
			Definition: anthropic.ToolDefinition{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: jsonschema.Definition{
					Type:       jsonschema.Object,
					Properties: map[string]jsonschema.Definition{},
					Required:   []string{""},
				},
			},
			Fn: func(_ string) (string, error) {
				return t.Result, nil
			},
		}
	}

	return tools, nil
}
