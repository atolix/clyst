package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/atolix/clyst/spec"
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

	d.Styles.NormalTitle = d.Styles.NormalTitle.
		PaddingLeft(2)

	d.Styles.NormalDesc = d.Styles.NormalDesc.
		PaddingLeft(2)

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87cefa")).
		MarginLeft(0).
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#6495ed")).
		Bold(true)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87cefa")).
		MarginLeft(0).
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#6495ed"))

	return d
}

func NewModel(items []list.Item) Model {
	const defaultWidth = 50
	l := list.New(items, NewStyleDelegate(), defaultWidth, 50)
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
	left := m.list.View()

	var right string
	if i, ok := m.list.SelectedItem().(EndpointItem); ok {
		parsed, err := json.MarshalIndent(i.Operation, "", "  ")
		if err == nil {
			var buf bytes.Buffer
			quick.Highlight(&buf, string(parsed), "json", "terminal", "github")
			right = buf.String()
		} else {
			right = "error formatting JSON"
		}
	} else {
		right = "No item selected"
	}

	rightBox := lipgloss.NewStyle().Width(100).Padding(1, 2).Border(lipgloss.RoundedBorder()).Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, rightBox)
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
