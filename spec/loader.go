package spec

import (
	"os"

	"gopkg.in/yaml.v3"
)

type OpenApiSpec struct {
	Paths map[string]map[string]Operation `yaml:"paths"`
}

type Parameter struct {
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Required bool   `yaml:"required"`
	Schema   struct {
		Type string `yaml:"type"`
	} `yaml:"schema"`
}

type Response struct {
	Description string `yaml:"description"`
}

type Operation struct {
	Summary    string              `yaml:"summary"`
	Parameters []Parameter         `yaml:parameters`
	Response   map[string]Response `yaml:"responses"`
}

func Load(filename string) (*OpenApiSpec, error) {
	data, err := os.ReadFile("api_spec.yml")
	if err != nil {
		panic(err)
	}

	var spec OpenApiSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		panic(err)
	}

	return &spec, nil
}
