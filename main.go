package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/atolix/clyst/output"
	"github.com/atolix/clyst/request"
	"github.com/atolix/clyst/spec"
	"github.com/atolix/clyst/tui"

	"github.com/charmbracelet/bubbles/list"
)

func main() {
	spec, err := spec.Load("api_spec.yml")
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

	baseURL := "https://jsonplaceholder.typicode.com"
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
