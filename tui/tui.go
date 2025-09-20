package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atolix/clyst/spec"
	"github.com/atolix/clyst/theme"

	"github.com/alecthomas/chroma/quick"
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
	result   string
	width    int
	height   int
}

func NewStyleDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.NormalTitle = d.Styles.NormalTitle.
		PaddingLeft(2)

	d.Styles.NormalDesc = d.Styles.NormalDesc.
		PaddingLeft(2)

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(theme.Primary).
		MarginLeft(0).
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Border).
		Bold(true)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(theme.Primary).
		MarginLeft(0).
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Border)

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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

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
	listWidth := m.width / 3
	detailWidth := m.width / 3

	height := m.height / 3

	return lipgloss.JoinHorizontal(lipgloss.Top, ListView(m, listWidth, height), DetailBox(m, detailWidth, height))
}

func ListView(m Model, width int, height int) string {
	return lipgloss.NewStyle().Width(width).Height(height).Render(m.list.View())
}

func DetailBox(m Model, width int, height int) string {
	var detail string
	if i, ok := m.list.SelectedItem().(EndpointItem); ok {
		parsed, err := json.MarshalIndent(i.Operation, "", "  ")
		if err == nil {
			var buf bytes.Buffer
			quick.Highlight(&buf, string(parsed), "json", "terminal", "github")
			detail = buf.String()
		} else {
			detail = "error formatting JSON"
		}
	} else {
		detail = "No item selected"
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Render(detail)
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
