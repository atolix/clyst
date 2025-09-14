package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/atolix/catalyst/request"
	"github.com/atolix/catalyst/spec"
	"github.com/atolix/catalyst/tui"
)

func main() {
	items, err := spec.Load("api_spec.yml")
	if err != nil {
		panic(err)
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

	baseURL := "https://jsonplaceholder.typicode.com"

	result, err := request.Send(baseURL, request.Endpoint{
		Method:  selected.Method,
		Path:    selected.Path,
		Summary: selected.Summary,
	})

	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
