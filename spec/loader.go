package spec

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

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

type componentsRaw struct {
	Parameters    map[string]Parameter   `yaml:"parameters"`
	RequestBodies map[string]RequestBody `yaml:"requestBodies"`
}

type openAPISpecRaw struct {
	BaseURL    string                             `yaml:"base_url"`
	Paths      map[string]map[string]operationRaw `yaml:"paths"`
	Components componentsRaw                      `yaml:"components"`
}

type parameterOrRef struct {
	Ref      string `yaml:"$ref"`
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Required bool   `yaml:"required"`
	Schema   struct {
		Type string `yaml:"type"`
	} `yaml:"schema"`
}

type requestBodyOrRef struct {
	Ref     string `yaml:"$ref"`
	Content map[string]struct {
		Schema map[string]any `yaml:"schema"`
	} `yaml:"content"`
}

type operationRaw struct {
	Summary     string              `yaml:"summary"`
	Parameters  []parameterOrRef    `yaml:"parameters"`
	RequestBody *requestBodyOrRef   `yaml:"requestBody"`
	Responses   map[string]Response `yaml:"responses"`
}

func Load(filename string) (*OpenApiSpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var raw openAPISpecRaw
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	resolved := &OpenApiSpec{
		BaseURL: raw.BaseURL,
		Paths:   make(map[string]map[string]Operation, len(raw.Paths)),
	}

	for p, methods := range raw.Paths {
		outMethods := make(map[string]Operation, len(methods))
		for method, op := range methods {
			rop, err := resolveOperation(op, raw.Components)
			if err != nil {
				return nil, fmt.Errorf("resolve %s %s: %w", strings.ToUpper(method), p, err)
			}
			outMethods[method] = rop
		}
		resolved.Paths[p] = outMethods
	}

	return resolved, nil
}

func resolveOperation(in operationRaw, comps componentsRaw) (Operation, error) {
	var out Operation
	out.Summary = in.Summary
	out.Responses = in.Responses

	for _, pr := range in.Parameters {
		if strings.TrimSpace(pr.Ref) != "" {
			name, kind, err := parseLocalRef(pr.Ref)
			if err != nil {
				return Operation{}, err
			}

			if kind != "parameters" {
				return Operation{}, fmt.Errorf("unsupported $ref kind for parameter: %s", kind)
			}

			param, ok := comps.Parameters[name]
			if !ok {
				return Operation{}, fmt.Errorf("unresolved parameter ref: %s", pr.Ref)
			}

			out.Parameters = append(out.Parameters, param)
			continue
		}
		out.Parameters = append(out.Parameters, Parameter{
			Name:     pr.Name,
			In:       pr.In,
			Required: pr.Required,
			Schema:   pr.Schema,
		})
	}

	if in.RequestBody != nil {
		rb := in.RequestBody
		if strings.TrimSpace(rb.Ref) != "" {
			name, kind, err := parseLocalRef(rb.Ref)
			if err != nil {
				return Operation{}, err
			}

			if kind != "requestBodies" {
				return Operation{}, fmt.Errorf("unsupported $ref kind for requestBody: %s", kind)

			}
			body, ok := comps.RequestBodies[name]
			if !ok {
				return Operation{}, fmt.Errorf("unresolved requestBody ref: %s", rb.Ref)
			}

			b := body
			out.RequestBody = &b
		} else {
			out.RequestBody = &RequestBody{Content: rb.Content}
		}
	}

	return out, nil
}

func parseLocalRef(ref string) (name string, kind string, err error) {
	if !strings.HasPrefix(ref, "#/") {
		return "", "", errors.New("only local $ref supported (must start with #/)")
	}

	parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	if len(parts) < 3 || parts[0] != "components" {
		return "", "", fmt.Errorf("invalid $ref format: %s", ref)
	}

	kind = parts[1]
	name = path.Clean(strings.Join(parts[2:], "/"))

	return name, kind, nil
}
