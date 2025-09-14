package spec

import (
	"os"

	"github.com/atolix/catalyst/tui"
	"github.com/charmbracelet/bubbles/list"
	"gopkg.in/yaml.v3"
)

type OpenApiSpec struct {
	Paths map[string]map[string]Operation `yaml:"paths"`
}

type Operation struct {
	Summary string `yaml:"summary"`
}

func Load(filename string) ([]list.Item, error) {
	data, err := os.ReadFile("api_spec.yml")
	if err != nil {
		panic(err)
	}

	var spec OpenApiSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		panic(err)
	}

	var items []list.Item
	for path, methods := range spec.Paths {
		for method, op := range methods {
			items = append(items, tui.EndpointItem{
				Method:  method,
				Path:    path,
				Summary: op.Summary,
			})
		}
	}

	return items, nil
}
