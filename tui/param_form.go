package tui

import (
	"fmt"

	"github.com/atolix/clyst/request"
	"github.com/atolix/clyst/spec"
	"github.com/atolix/clyst/theme"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PrefilledProvider struct {
	path  map[string]string
	query map[string]string
	body  string
}

func (p PrefilledProvider) GetPathParam(param spec.Parameter) string  { return p.path[param.Name] }
func (p PrefilledProvider) GetQueryParam(param spec.Parameter) string { return p.query[param.Name] }
func (p PrefilledProvider) GetRequestBody() string                    { return p.body }

func CollectParams(ep request.Endpoint) (PrefilledProvider, error) {
	m := newParamFormModel(ep)
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return PrefilledProvider{}, err
	}
	fm := final.(paramFormModel)
	return fm.toProvider(), nil
}

type TUIInput struct {
	Endpoint  request.Endpoint
	collected bool
	provider  PrefilledProvider
}

func (c *TUIInput) ensureCollected() {
	if c.collected {
		return
	}
	if p, err := CollectParams(c.Endpoint); err == nil {
		c.provider = p
		c.collected = true
	}
}

func (c *TUIInput) GetPathParam(p spec.Parameter) string {
	c.ensureCollected()
	return c.provider.GetPathParam(p)
}

func (c *TUIInput) GetQueryParam(p spec.Parameter) string {
	c.ensureCollected()
	return c.provider.GetQueryParam(p)
}

func (c *TUIInput) GetRequestBody() string {
	c.ensureCollected()
	return c.provider.GetRequestBody()
}

type paramField struct {
	p     spec.Parameter
	input textinput.Model
}

type paramFormModel struct {
	ep           request.Endpoint
	pathFields   []paramField
	queryFields  []paramField
	bodyArea     textarea.Model
	hasBody      bool
	focusedIndex int
	width        int
	height       int
}

func newParamFormModel(ep request.Endpoint) paramFormModel {
	var pathFields []paramField
	var queryFields []paramField
	for _, p := range ep.Operation.Parameters {
		ti := textinput.New()
		ti.Prompt = "> "
		ti.Placeholder = fmt.Sprintf("%s (%s)", p.Name, p.Schema.Type)
		switch p.In {
		case "path":
			pathFields = append(pathFields, paramField{p: p, input: ti})
		case "query":
			queryFields = append(queryFields, paramField{p: p, input: ti})
		}
	}

	ta := textarea.New()
	ta.Placeholder = "{\n  \"example\": \"value\"\n}"
	ta.ShowLineNumbers = false
	ta.MaxWidth = 0
	hasBody := ep.Operation.RequestBody != nil

	m := paramFormModel{
		ep:           ep,
		pathFields:   pathFields,
		queryFields:  queryFields,
		bodyArea:     ta,
		hasBody:      hasBody,
		focusedIndex: 0,
	}

	if len(m.pathFields) > 0 {
		m.pathFields[0].input.Focus()
	} else if len(m.queryFields) > 0 {
		m.queryFields[0].input.Focus()
	} else if m.hasBody {
		m.bodyArea.Focus()
	}

	return m
}

func (m paramFormModel) Init() tea.Cmd { return nil }

func (m paramFormModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Render("Parameters")
	section := lipgloss.NewStyle().Bold(true)
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Border).Padding(1, 2)
	outer := lipgloss.NewStyle().MarginBottom(2)

	var sections []string

	if len(m.pathFields) > 0 {
		var pathViews []string
		for _, f := range m.pathFields {
			label := lipgloss.NewStyle().Foreground(theme.Muted).Render(fmt.Sprintf("%s (%s)", f.p.Name, f.p.Schema.Type))
			pathViews = append(pathViews, label+"\n"+f.input.View())
		}
		sections = append(sections, section.Render("Path Params"))
		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, pathViews...))
	}

	if len(m.queryFields) > 0 {
		if len(sections) > 0 {
			sections = append(sections, "")
		}
		var queryViews []string
		for _, f := range m.queryFields {
			label := lipgloss.NewStyle().Foreground(theme.Muted).Render(fmt.Sprintf("%s (%s)", f.p.Name, f.p.Schema.Type))
			queryViews = append(queryViews, label+"\n"+f.input.View())
		}
		sections = append(sections, section.Render("Query Params"))
		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, queryViews...))
	}

	if m.hasBody {
		if len(sections) > 0 {
			sections = append(sections, "")
		}
		sections = append(sections, section.Render("Body (JSON)"))
		sections = append(sections, m.bodyArea.View())
	}

	if len(sections) > 0 {
		sections = append(sections, "")
	}
	sections = append(sections, lipgloss.NewStyle().Faint(true).Render("Tab/Shift+Tab: move  Enter: submit (Bodyでは改行)  Esc: cancel"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return outer.Render(lipgloss.JoinVertical(lipgloss.Left, title, box.Render(content)))
}

func (m paramFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.hasBody {
			m.bodyArea.SetWidth(m.width - 8)
			m.bodyArea.SetHeight(m.height / 3)
		}
		for i := range m.pathFields {
			m.pathFields[i].input.Width = m.width - 8
		}
		for i := range m.queryFields {
			m.queryFields[i].input.Width = m.width - 8
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		case "tab":
			m.focusNext()
			m.applyFocus()
			return m, nil
		case "shift+tab":
			m.focusPrev()
			m.applyFocus()
			return m, nil
		case "enter":
			if _, kind := m.currentIndex(); kind == "body" {
				var cmd tea.Cmd
				m.bodyArea, cmd = m.bodyArea.Update(msg)
				return m, cmd
			}
			return m, tea.Quit
		case "up":
			if _, kind := m.currentIndex(); kind == "body" {
				var cmd tea.Cmd
				m.bodyArea, cmd = m.bodyArea.Update(msg)
				return m, cmd
			}
			m.focusPrev()
			m.applyFocus()
			return m, nil
		case "down":
			if _, kind := m.currentIndex(); kind == "body" {
				var cmd tea.Cmd
				m.bodyArea, cmd = m.bodyArea.Update(msg)
				return m, cmd
			}
			m.focusNext()
			m.applyFocus()
			return m, nil
		}
	}

	if idx, kind := m.currentIndex(); kind == "path" {
		var cmd tea.Cmd
		m.pathFields[idx].input, cmd = m.pathFields[idx].input.Update(msg)
		return m, cmd
	} else if kind == "query" {
		var cmd tea.Cmd
		m.queryFields[idx].input, cmd = m.queryFields[idx].input.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.bodyArea, cmd = m.bodyArea.Update(msg)
	return m, cmd
}

func (m *paramFormModel) currentIndex() (int, string) {
	if m.focusedIndex < len(m.pathFields) {
		return m.focusedIndex, "path"
	}
	j := m.focusedIndex - len(m.pathFields)
	if j < len(m.queryFields) {
		return j, "query"
	}
	if m.hasBody {
		return -1, "body"
	}
	if len(m.queryFields) > 0 {
		return len(m.queryFields) - 1, "query"
	}
	if len(m.pathFields) > 0 {
		return len(m.pathFields) - 1, "path"
	}
	return -1, "none"
}

func (m *paramFormModel) focusNext() {
	total := len(m.pathFields) + len(m.queryFields)
	if m.hasBody {
		total++
	}
	m.focusedIndex = (m.focusedIndex + 1) % total
}

func (m *paramFormModel) focusPrev() {
	total := len(m.pathFields) + len(m.queryFields)
	if m.hasBody {
		total++
	}
	m.focusedIndex = (m.focusedIndex - 1 + total) % total
}

func (m *paramFormModel) blurAll() {
	for i := range m.pathFields {
		m.pathFields[i].input.Blur()
	}
	for i := range m.queryFields {
		m.queryFields[i].input.Blur()
	}
	if m.hasBody {
		m.bodyArea.Blur()
	}
}

func (m *paramFormModel) applyFocus() {
	m.blurAll()
	if idx, kind := m.currentIndex(); kind == "path" {
		m.pathFields[idx].input.Focus()
	} else if kind == "query" {
		m.queryFields[idx].input.Focus()
	} else if kind == "body" && m.hasBody {
		m.bodyArea.Focus()
	}
}

func (m paramFormModel) toProvider() PrefilledProvider {
	pathVals := map[string]string{}
	queryVals := map[string]string{}
	for _, f := range m.pathFields {
		pathVals[f.p.Name] = f.input.Value()
	}
	for _, f := range m.queryFields {
		queryVals[f.p.Name] = f.input.Value()
	}
	return PrefilledProvider{path: pathVals, query: queryVals, body: m.bodyArea.Value()}
}
