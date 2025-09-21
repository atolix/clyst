package spec

import (
	"os"

	"gopkg.in/yaml.v3"
)

type OpenApiSpec struct {
	BaseURL string                          `yaml:"base_url"`
	Paths   map[string]map[string]Operation `yaml:"paths"`
}

type Parameter struct {
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Required bool   `yaml:"required"`
	Schema   struct {
		Type string `yaml:"type"`
	} `yaml:"schema"`
}

type RequestBody struct {
	Content map[string]struct {
		Schema map[string]any `yaml:"schema"`
	} `yaml:"content"`
}

type Response struct {
	Description string `yaml:"description"`
}

type Operation struct {
	Summary     string              `yaml:"summary"`
	Parameters  []Parameter         `yaml:"parameters"`
	RequestBody *RequestBody        `yaml:"requestBody"`
	Responses   map[string]Response `yaml:"responses"`
}

func Load(filename string) (*OpenApiSpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec OpenApiSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}
