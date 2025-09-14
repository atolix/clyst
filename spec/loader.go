package spec

import (
	"os"
	"sort"

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

	var endpoints []tui.EndpointItem
	for path, methods := range spec.Paths {
		for method, op := range methods {
			endpoints = append(endpoints, tui.EndpointItem{
				Method:  method,
				Path:    path,
				Summary: op.Summary,
			})
		}
	}

	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Method == endpoints[j].Method {
			return endpoints[i].Path < endpoints[j].Path
		}
		return endpoints[i].Method < endpoints[j].Method
	})

	var items []list.Item
	for _, ep := range endpoints {
		items = append(items, ep)
	}

	return items, nil
}
