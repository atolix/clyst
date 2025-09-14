package spec

import (
	"os"

	"gopkg.in/yaml.v3"
)

type OpenApiSpec struct {
	Paths map[string]map[string]Operation `yaml:"paths"`
}

type Operation struct {
	Summary string `yaml:"summary"`
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
