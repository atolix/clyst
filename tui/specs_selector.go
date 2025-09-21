package tui

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

type selectSpecModel struct {
	list     list.Model
	Selected *SpecItem
}

func newSpecSelectorModel(title string, items []list.Item) selectSpecModel {
	const defaultWidth = 60
	l := list.New(items, NewStyleDelegate(), defaultWidth, 20)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return selectSpecModel{list: l}
}

func (m selectSpecModel) Init() tea.Cmd { return nil }

func (m selectSpecModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(SpecItem); ok {
				m.Selected = &i
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m selectSpecModel) View() string { return m.list.View() }

func SelectSpec(title string, options []SpecItem) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options provided")
	}

	items := make([]list.Item, 0, len(options))
	for _, o := range options {
		items = append(items, o)
	}

	m := newSpecSelectorModel(title, items)
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}

	fm := final.(selectSpecModel)
	if fm.Selected == nil {
		return "", nil
	}

	return fm.Selected.Value, nil
}
