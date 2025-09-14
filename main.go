package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gopkg.in/yaml.v3"
)

type OpenApiSpec struct {
	Paths map[string]map[string]Operation `yaml:"paths"`
}

type Operation struct {
	Summary string `yaml:"summary"`
}

type endpointItem struct {
	method  string
	path    string
	summary string
}

func (i endpointItem) Title() string {
	return fmt.Sprintf("%s %s", i.method, i.path)
}

func (i endpointItem) Description() string {
	return i.summary
}

func (i endpointItem) FilterValue() string {
	return i.path
}

type model struct {
	list     list.Model
	selected *endpointItem
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
			if i, ok := m.list.SelectedItem().(endpointItem); ok {
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
	data, err := os.ReadFile("api_spec.yml")
	if err != nil {
		panic(err)
	}

	var spec OpenApiSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		panic(err)
	}

	var items []list.Item
	for path, methods := range spec.Paths {
		for method, op := range methods {
			items = append(items, endpointItem{
				method:  method,
				path:    path,
				summary: op.Summary,
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
		result := map[string]string{
			"method":  fm.selected.method,
			"path":    fm.selected.path,
			"summary": fm.selected.summary,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		if err := enc.Encode(result); err != nil {
			fmt.Println("Error Encoding Result:", err)
		}
	}
}
