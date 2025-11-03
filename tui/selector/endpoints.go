package selector

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

type endpointModel struct {
	list             list.Model
	selected         *EndpointItem
	width            int
	height           int
	switchSpecSelect bool
}

type EndpointResult struct {
	Selected         *EndpointItem
	SwitchSpecSelect bool
}

func newEndpointsModel(items []list.Item) endpointModel {
	const defaultWidth = 50
	l := list.New(items, NewStyleDelegate(), defaultWidth, 40)
	l.Title = "Api Endpoints  (press 'Ctrl+b' to back to spec selection)"
	l.SetShowStatusBar(false)

	return endpointModel{list: l}
}

func (m endpointModel) Init() tea.Cmd { return nil }

func (m endpointModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(EndpointItem); ok {
				m.selected = &i
				return m, tea.Quit
			}
		case "ctrl+b":
			m.switchSpecSelect = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m endpointModel) View() string {
	listWidth := m.width / 2
	detailWidth := m.width / 2
	height := m.height / 3

	return lipgloss.JoinHorizontal(lipgloss.Top, listView(m, listWidth, height), detailBox(m, detailWidth, height))
}

func listView(m endpointModel, width int, height int) string {
	return lipgloss.NewStyle().Width(width).Height(height).Render(m.list.View())
}

func detailBox(m endpointModel, width int, height int) string {
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

func RunEndpoints(items []list.Item) (EndpointResult, error) {
	m := newEndpointsModel(items)
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return EndpointResult{}, err
	}

	fm := final.(endpointModel)

	return EndpointResult{Selected: fm.selected, SwitchSpecSelect: fm.switchSpecSelect}, nil
}
