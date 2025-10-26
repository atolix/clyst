package selector

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type SpecItem struct {
	TitleText string
	DescText  string
	Value     string
}

func (i SpecItem) Title() string       { return i.TitleText }
func (i SpecItem) Description() string { return i.DescText }
func (i SpecItem) FilterValue() string { return i.TitleText }

type specModel struct {
	list     list.Model
	selected *SpecItem
}

func newSpecModel(title string, items []list.Item) specModel {
	const defaultWidth = 60
	l := list.New(items, NewStyleDelegate(), defaultWidth, 20)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return specModel{list: l}
}

func (m specModel) Init() tea.Cmd { return nil }

func (m specModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(SpecItem); ok {
				m.selected = &i
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m specModel) View() string { return m.list.View() }

func SelectSpec(title string, options []SpecItem) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options provided")
	}

	items := make([]list.Item, 0, len(options))
	for _, o := range options {
		items = append(items, o)
	}

	m := newSpecModel(title, items)
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return "", err
	}

	fm := final.(specModel)
	if fm.selected == nil {
		return "", nil
	}

	return fm.selected.Value, nil
}
