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

	for {
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

		specDoc, err := spec.Load(specPath)
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

		runRes, err := tui.Run(items)
		if err != nil {
			fmt.Println("TUI running error:", err)
			os.Exit(1)
		}

		if runRes.SwitchSpecSelect {
			continue
		}

		if runRes.Selected == nil {
			fmt.Println("No endpoint selected")
			return
		}

		ep := request.Endpoint{
			Method:    runRes.Selected.Method,
			Path:      runRes.Selected.Path,
			Operation: runRes.Selected.Operation,
		}

		baseURL := specDoc.BaseURL
		if strings.TrimSpace(baseURL) == "" {
			baseURL = "https://jsonplaceholder.typicode.com"
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
		return
	}
}
