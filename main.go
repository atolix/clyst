package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type OpenApiSpec struct {
	Paths map[string]map[string]Operation `yaml:"paths"`
}

type Operation struct {
	Summary string `yaml:"summary"`
}

func main() {
	data, err := os.ReadFile("api_spec.yml")
	if err != nil {
		panic(err)
	}

	var spec OpenApiSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		panic(err)
	}

	for path, methods := range spec.Paths {
		for method, op := range methods {
			fmt.Printf("%s %s - %s\n", method, path, op.Summary)
		}
	}
}
