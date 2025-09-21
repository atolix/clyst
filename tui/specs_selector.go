package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type StringItem struct {
	TitleText string
	DescText  string
	Value     string
}

func (i StringItem) Title() string       { return i.TitleText }
func (i StringItem) Description() string { return i.DescText }
func (i StringItem) FilterValue() string { return i.TitleText }

type selectModel struct {
	list     list.Model
	Selected *StringItem
}

func newSelectModel(title string, items []list.Item) selectModel {
	const defaultWidth = 60
	l := list.New(items, NewStyleDelegate(), defaultWidth, 20)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	return selectModel{list: l}
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(StringItem); ok {
				m.Selected = &i
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectModel) View() string { return m.list.View() }

// SelectOne renders a simple selection list with a title and returns the chosen value.
func SelectOne(title string, options []StringItem) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options provided")
	}
	items := make([]list.Item, 0, len(options))
	for _, o := range options {
		items = append(items, o)
	}
	m := newSelectModel(title, items)
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	fm := final.(selectModel)
	if fm.Selected == nil {
		return "", nil
	}
	return fm.Selected.Value, nil
}
