package tui

import (
	"fmt"
	"strings"

	"github.com/atolix/clyst/params"
	"github.com/atolix/clyst/request"
	"github.com/atolix/clyst/spec"
	"github.com/atolix/clyst/theme"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PrefilledProvider struct {
	path      map[string]string
	query     map[string]string
	body      string
	recording bool
}

type TUIInput struct {
	Endpoint  request.Endpoint
	collected bool
	provider  PrefilledProvider
	canceled  bool
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
	canceled     bool
	recording    bool
}

func (p PrefilledProvider) GetPathParam(param spec.Parameter) string  { return p.path[param.Name] }
func (p PrefilledProvider) GetQueryParam(param spec.Parameter) string { return p.query[param.Name] }
func (p PrefilledProvider) GetRequestBody() string                    { return p.body }
func (p PrefilledProvider) ShouldRecord() bool                        { return p.recording }

func CollectParams(ep request.Endpoint) (PrefilledProvider, bool, error) {
	var initial PrefilledProvider

	if store, err := params.Load("."); err == nil {
		if presets := store.PresetsFor(ep.Method, ep.Path); len(presets) > 0 {
			selected, canceled, err := SelectPreset(ep, presets)
			if err != nil {
				fmt.Println("failed to select preset:", err)
			} else if canceled {
				return PrefilledProvider{}, true, nil
			} else if selected != nil {
				initial.path = selected.Path
				initial.query = selected.Query
				initial.body = selected.Body
			}
		}
	} else {
		fmt.Println("failed to read saved params:", err)
	}

	m := newParamFormModel(ep, initial)
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return PrefilledProvider{}, false, err
	}
	fm := final.(paramFormModel)
	return fm.toProvider(), fm.canceled, nil
}

func (c *TUIInput) ensureCollected() {
	if c.collected {
		return
	}
	if p, canceled, err := CollectParams(c.Endpoint); err == nil {
		c.provider = p
		c.collected = true
		c.canceled = canceled
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

func (c *TUIInput) ShouldRecord() bool {
	c.ensureCollected()
	return c.provider.ShouldRecord()
}

func (c *TUIInput) Canceled() bool {
	return c.canceled
}

func newParamFormModel(ep request.Endpoint, initial PrefilledProvider) paramFormModel {
	var pathFields []paramField
	var queryFields []paramField
	for _, p := range ep.Operation.Parameters {
		ti := textinput.New()
		ti.Prompt = "> "
		ti.Placeholder = fmt.Sprintf("%s (%s)", p.Name, p.Schema.Type)
		switch p.In {
		case "path":
			if initial.path != nil {
				if v, ok := initial.path[p.Name]; ok {
					ti.SetValue(v)
				}
			}
			pathFields = append(pathFields, paramField{p: p, input: ti})
		case "query":
			if initial.query != nil {
				if v, ok := initial.query[p.Name]; ok {
					ti.SetValue(v)
				}
			}
			queryFields = append(queryFields, paramField{p: p, input: ti})
		}
	}

	ta := textarea.New()
	ta.Placeholder = "{\n  \"example\": \"value\"\n}"
	ta.ShowLineNumbers = false
	ta.MaxWidth = 0
	if strings.TrimSpace(initial.body) != "" {
		ta.SetValue(initial.body)
	}
	hasBody := ep.Operation.RequestBody != nil

	m := paramFormModel{
		ep:           ep,
		pathFields:   pathFields,
		queryFields:  queryFields,
		bodyArea:     ta,
		hasBody:      hasBody,
		focusedIndex: 0,
		recording:    false,
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

func (m paramFormModel) Init() tea.Cmd {
	return nil
}

func (m paramFormModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Render("Parameters")
	section := lipgloss.NewStyle().Bold(true)
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Border).Padding(1, 2)
	outer := lipgloss.NewStyle().MarginBottom(2)
	onStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.DarkText).
		Background(theme.Primary).
		Padding(0, 1)
	offStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Muted).
		Background(lipgloss.Color("#2b2d3a")).
		Padding(0, 1)
	recordStatus := offStyle.Render("Recording OFF")
	if m.recording {
		recordStatus = onStyle.Render("Recording ON")
	}

	var sections []string
	statusLines := []string{recordStatus}

	sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, statusLines...))
	sections = append(sections, "")

	hints := []string{
		"Tab/Shift+Tab: move",
		"Ctrl+r: toggle recording",
		"Enter: submit (newline in Body)",
		"Ctrl+s: submit",
		"Esc: cancel",
	}
	sections = append(sections, lipgloss.NewStyle().Faint(true).Render(strings.Join(hints, "  ")))
	sections = append(sections, "")

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
		case "ctrl+r":
			m.recording = !m.recording
			return m, nil
		case "esc":
			m.canceled = true
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

	return PrefilledProvider{
		path:      pathVals,
		query:     queryVals,
		body:      m.bodyArea.Value(),
		recording: m.recording,
	}
}
