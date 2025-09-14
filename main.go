package main

import (
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
	title, desc string
}

func (i endpointItem) Title() string       { return i.title }
func (i endpointItem) Description() string { return i.desc }
func (i endpointItem) FilterValue() string { return i.title }

type model struct {
	list list.Model
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
				title: fmt.Sprintf("%s %s", method, path),
				desc:  op.Summary,
			})
		}
	}

	m := newModel(items)
	_, err = tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("bye")
}
