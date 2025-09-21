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

	"github.com/charmbracelet/bubbles/list"
)

func main() {
	cfg, _ := config.Load()
	names, err := config.DefineSpecNames(cfg)
	if err != nil {
		fmt.Println("Config error:", err)
		os.Exit(1)
	}

	found, err := spec.DiscoverSpecFiles(".", names)
	if err != nil {
		fmt.Println("Discovery error:", err)
		os.Exit(1)
	}

	if len(found) == 0 {
		fmt.Printf("No spec file found. Looked for: %s\n", strings.Join(names, ", "))
		os.Exit(1)
	}

	specPath := found[0]
	if len(found) > 1 {
		var opts []tui.SpecItem
		for _, p := range found {
			opts = append(opts, tui.SpecItem{
				TitleText: p,
				DescText:  filepath.Dir(p),
				Value:     p,
			})
		}
		chosen, err := tui.SelectSpec("Select an OpenAPI spec", opts)
		if err != nil {
			fmt.Println("TUI running error:", err)
			os.Exit(1)
		}
		if chosen == "" {
			fmt.Println("No spec selected")
			return
		}
		specPath = chosen
	}

	spec, err := spec.Load(specPath)
	if err != nil {
		panic(err)
	}

	var endpoints []tui.EndpointItem
	for path, methods := range spec.Paths {
		for method, op := range methods {
			endpoints = append(endpoints, tui.EndpointItem{
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

	var items []list.Item
	for _, ep := range endpoints {
		items = append(items, ep)
	}

	selected, err := tui.Run(items)
	if err != nil {
		fmt.Println("TUI running error:", err)
		os.Exit(1)
	}

	if selected == nil {
		fmt.Println("No endpoint selected")
		return
	}

	ep := request.Endpoint{
		Method:    selected.Method,
		Path:      selected.Path,
		Operation: selected.Operation,
	}

	baseURL := spec.BaseURL
	if strings.TrimSpace(baseURL) == "" {
		fmt.Println("No BaseURL")
		os.Exit(1)
	}

	input, err := request.AssembleInput(baseURL, ep, &tui.TUIInput{Endpoint: ep})
	if err != nil {
		panic(err)
	}

	result, err := request.Send(ep, input)
	if err != nil {
		panic(err)
	}

	fmt.Println(output.Render(result))
}
