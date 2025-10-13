package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atolix/clyst/config"
	"github.com/atolix/clyst/output"
	"github.com/atolix/clyst/request"
	"github.com/atolix/clyst/spec"
	"github.com/atolix/clyst/tui"
	"github.com/atolix/clyst/tui/selector"

	"github.com/charmbracelet/bubbles/list"
)

func main() {
	names := specNamesOrExit()

Outer:
	for {
		specPath, ok := discoverSpecPath(names)
		if !ok {
			return
		}

		specDoc := mustLoadSpec(specPath)
		if specDoc == nil {
			return
		}

		switchSpec, exit := runEndpointSession(specDoc)
		if exit {
			return
		}
		if switchSpec {
			continue Outer
		}
	}
}

func specNamesOrExit() []string {
	cfg, _ := config.Load()
	names, err := config.DefineSpecNames(cfg)
	if err != nil {
		fmt.Println("Config error:", err)
		os.Exit(1)
	}
	return names
}

func discoverSpecPath(names []string) (string, bool) {
	found, err := spec.DiscoverSpecFiles(".", names)
	if err != nil {
		panic(err)
	}

	if len(found) == 0 {
		fmt.Printf("No spec file found. Looked for: %s\n", strings.Join(names, ", "))
		os.Exit(1)
	}

	specPath := found[0]
	if len(found) == 1 {
		return specPath, true
	}

	opts := make([]selector.SpecItem, 0, len(found))
	for _, p := range found {
		opts = append(opts, selector.SpecItem{
			TitleText: p,
			DescText:  filepath.Dir(p),
			Value:     p,
		})
	}

	selected, err := selector.SelectSpec("Select an OpenAPI spec", opts)
	if err != nil {
		fmt.Println("TUI running error:", err)
		os.Exit(1)
	}

	if strings.TrimSpace(selected) == "" {
		return "", false
	}

	return selected, true
}

func mustLoadSpec(path string) *spec.OpenApiSpec {
	doc, err := spec.Load(path)
	if doc == nil {
		return nil
	}
	if err != nil {
		panic(err)
	}
	return doc
}

func runEndpointSession(doc *spec.OpenApiSpec) (bool, bool) {
	items := buildEndpointItems(doc)

EndpointLoop:
	for {
		runRes, err := selector.RunEndpoints(items)
		if err != nil {
			panic(err)
		}

		if runRes.SwitchSpecSelect {
			return true, false
		}

		if runRes.Selected == nil {
			return false, false
		}

		ep := request.Endpoint{
			Method:    runRes.Selected.Method,
			Path:      runRes.Selected.Path,
			Operation: runRes.Selected.Operation,
		}

		baseURL := doc.BaseURL
		if strings.TrimSpace(baseURL) == "" {
			fmt.Println("Not found BaseURL")
			os.Exit(1)
		}

		tuiInput := &tui.TUIInput{Endpoint: ep}
		input, canceled, err := request.AssembleInput(baseURL, ep, tuiInput)
		if err != nil {
			panic(err)
		}

		if canceled {
			if tuiInput.ShouldReselectEndpoint() {
				continue EndpointLoop
			}
			return false, true
		}

		result, err := request.Send(ep, input)
		if err != nil {
			panic(err)
		}

		handlePresetRecording(ep, tuiInput)

		fmt.Println(output.Render(result))
		return false, true
	}
}

func buildEndpointItems(doc *spec.OpenApiSpec) []list.Item {
	var endpoints []selector.EndpointItem
	for path, methods := range doc.Paths {
		for method, op := range methods {
			endpoints = append(endpoints, selector.EndpointItem{
				Method:    method,
				Path:      path,
				Operation: op,
			})
		}
	}

	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Method == endpoints[j].Method {
			return endpoints[i].Path < endpoints[j].Path
		}
		return endpoints[i].Method < endpoints[j].Method
	})

	items := make([]list.Item, 0, len(endpoints))
	for _, ep := range endpoints {
		items = append(items, ep)
	}
	return items
}

func handlePresetRecording(ep request.Endpoint, tuiInput *tui.TUIInput) {
	if !tuiInput.ShouldRecord() {
		return
	}
	if err := request.SavePreset(".", ep, tuiInput); err != nil {
		fmt.Println("failed to save params:", err)
	}
}
