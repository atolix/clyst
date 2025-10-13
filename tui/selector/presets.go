package selector

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/atolix/clyst/params"
	"github.com/atolix/clyst/request"
	"github.com/atolix/clyst/theme"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type presetItem struct {
	index int
	title string
	desc  string
}

func (i presetItem) Title() string       { return i.title }
func (i presetItem) Description() string { return i.desc }
func (i presetItem) FilterValue() string { return i.title }

type presetModel struct {
	list     list.Model
	selected int
	canceled bool
	reselect bool
}

func newPresetModel(ep request.Endpoint, presets []params.StoredParams) presetModel {
	items := make([]list.Item, 0, len(presets)+1)
	items = append(items, presetItem{
		index: 0,
		title: "New values",
		desc:  "Open form with empty fields",
	})

	for idx, preset := range presets {
		items = append(items, presetItem{
			index: idx + 1,
			title: presetTitle(preset),
			desc:  presetSummary(preset),
		})
	}

	const defaultWidth = 60
	l := list.New(items, NewStyleDelegate(), defaultWidth, 20)
	l.Title = fmt.Sprintf("Saved presets: %s %s (Esc: cancel, Ctrl+b: back)", strings.ToUpper(ep.Method), ep.Path)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return presetModel{list: l, selected: 0}
}

func (m presetModel) Init() tea.Cmd { return nil }

func (m presetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b":
			m.reselect = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(presetItem); ok {
				m.selected = item.index
				return m, tea.Quit
			}
		case "esc":
			m.canceled = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m presetModel) View() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(1, 2)

	return box.Render(m.list.View())
}

func presetTitle(p params.StoredParams) string {
	if p.RecordedAt.IsZero() {
		return "Saved preset"
	}
	return p.RecordedAt.Local().Format(time.DateTime)
}

func presetSummary(p params.StoredParams) string {
	var parts []string
	if len(p.Path) > 0 {
		parts = append(parts, "Path "+joinPairs(p.Path))
	}
	if len(p.Query) > 0 {
		parts = append(parts, "Query "+joinPairs(p.Query))
	}
	if strings.TrimSpace(p.Body) != "" {
		parts = append(parts, fmt.Sprintf("Body %d chars", len(p.Body)))
	} else {
		parts = append(parts, "Body empty")
	}
	return strings.Join(parts, "  ")
}

func joinPairs(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return strings.Join(out, ", ")
}

func SelectPreset(ep request.Endpoint, presets []params.StoredParams) (*params.StoredParams, bool, bool, error) {
	if len(presets) == 0 {
		return nil, false, false, nil
	}

	model := newPresetModel(ep, presets)
	final, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, false, false, err
	}
	res := final.(presetModel)
	if res.canceled {
		return nil, false, true, nil
	}
	if res.reselect {
		return nil, true, false, nil
	}
	if res.selected == 0 {
		return nil, false, false, nil
	}
	return &presets[res.selected-1], false, false, nil
}
