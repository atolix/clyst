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

Outer:
	for {
		found, err := spec.DiscoverSpecFiles(".", names)
		if err != nil {
			panic(err)
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
			selected, err := tui.SelectSpec("Select an OpenAPI spec", opts)
			if err != nil {
				fmt.Println("TUI running error:", err)
				os.Exit(1)
			}

			specPath = selected
		}

		specDoc, err := spec.Load(specPath)
		if specDoc == nil {
			return
		}

		if err != nil {
			panic(err)
		}

		var endpoints []tui.EndpointItem
		for path, methods := range specDoc.Paths {
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

	EndpointLoop:
		for {
			runRes, err := tui.Run(items)
			if err != nil {
				panic(err)
			}

			if runRes.SwitchSpecSelect {
				continue Outer
			}

			ep := request.Endpoint{
				Method:    runRes.Selected.Method,
				Path:      runRes.Selected.Path,
				Operation: runRes.Selected.Operation,
			}

			baseURL := specDoc.BaseURL
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
				if tuiInput.BackRequested() {
					continue EndpointLoop
				}
				return
			}

			result, err := request.Send(ep, input)
			if err != nil {
				panic(err)
			}

			if tuiInput.ShouldRecord() {
				if err := request.SavePreset(".", ep, tuiInput); err != nil {
					fmt.Println("failed to save params:", err)
				}
			}

			fmt.Println(output.Render(result))
			return
		}
	}
}
