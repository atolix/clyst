package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atolix/catalyst/request"
	"github.com/atolix/catalyst/spec"
)

type model struct {
	list     list.Model
	selected *request.Endpoint
}

func newModel(items []list.Item) model {
	const defaultWidth = 50
	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, 40)
	l.Title = "Api Endpoints"
	return model{list: l}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(request.Endpoint); ok {
				m.selected = &i
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
}

func main() {
	spec, err := spec.Load("api_spec.yml")
	if err != nil {
		panic(err)
	}

	var items []list.Item
	for path, methods := range spec.Paths {
		for method, op := range methods {
			items = append(items, request.Endpoint{
				Method:  method,
				Path:    path,
				Summary: op.Summary,
			})
		}
	}

	m := newModel(items)
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(model); ok && fm.selected != nil {
		baseURL := "https://jsonplaceholder.typicode.com"

		result, err := request.Send(baseURL, *fm.selected)
		if err != nil {
			panic(err)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result)
	}
}
