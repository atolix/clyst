package tui

import (
	"fmt"
	"strings"

	"github.com/atolix/catalyst/spec"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EndpointItem struct {
	Method    string
	Path      string
	Operation spec.Operation
}

func (i EndpointItem) Title() string       { return fmt.Sprintf("%s %s", strings.ToUpper(i.Method), i.Path) }
func (i EndpointItem) Description() string { return i.Operation.Summary }
func (i EndpointItem) FilterValue() string { return i.Path }

type Model struct {
	list     list.Model
	Selected *EndpointItem
}

func NewStyleDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87cefa")).
		Bold(true)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87cefa"))

	return d
}

func NewModel(items []list.Item) Model {
	const defaultWidth = 50
	l := list.New(items, NewStyleDelegate(), defaultWidth, 40)
	l.Title = "Api Endpoints"
	l.SetShowStatusBar(false)
	return Model{list: l}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(EndpointItem); ok {
				m.Selected = &i
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
}

func Run(items []list.Item) (*EndpointItem, error) {
	m := NewModel(items)
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}
	fm := final.(Model)

	return fm.Selected, nil
}
