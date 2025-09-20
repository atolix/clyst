package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atolix/clyst/request"

	"github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/lipgloss"
)

type styles struct {
	title   lipgloss.Style
	label   lipgloss.Style
	value   lipgloss.Style
	box     lipgloss.Style
	codeBox lipgloss.Style
}

func defaultStyles() styles {
	return styles{
		title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87cefa")),
		label:   lipgloss.NewStyle().Foreground(lipgloss.Color("#8a8f98")).Bold(true),
		value:   lipgloss.NewStyle().Foreground(lipgloss.Color("#e6e6e6")),
		box:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6495ed")).Padding(1, 2),
		codeBox: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#5f87af")).Padding(0, 1).MarginTop(0),
	}
}

// Render returns a pretty, styled string for a SendResult.
func Render(result request.ResultInfo) string {
	s := defaultStyles()
	reqBox := renderRequestBox(result, s)
	bodyStr, lexer := prepareResponseBody(result)
	headers := renderHeaders(result, s)
	respBox := renderResponseBox(result, headers, bodyStr, lexer, s)
	return layout(reqBox, respBox)
}

func renderRequestBox(result request.ResultInfo, s styles) string {
	lines := []string{
		s.label.Render("Method:") + " " + s.value.Render(strings.ToUpper(result.Request.Method)),
		s.label.Render("URL:") + "    " + s.value.Render(result.Request.URL),
	}
	if strings.TrimSpace(result.Request.Body) != "" {
		var pretty bytes.Buffer
		var rendered string
		if json.Indent(&pretty, []byte(result.Request.Body), "", "  ") == nil {
			var buf bytes.Buffer
			_ = quick.Highlight(&buf, pretty.String(), "json", "terminal", "github")
			rendered = buf.String()
		} else {
			rendered = result.Request.Body
		}
		lines = append(lines, s.label.Render("Body:")+"\n"+s.codeBox.Render(rendered))
	}
	return s.title.Render("Request") + "\n" + s.box.Render(strings.Join(lines, "\n"))
}

func prepareResponseBody(result request.ResultInfo) (string, string) {
	if result.Response.JSONBody != nil {
		if enc, err := json.MarshalIndent(result.Response.JSONBody, "", "  "); err == nil {
			return string(enc), "json"
		}
	}
	bodyStr := string(result.Response.RawBody)
	ct := strings.ToLower(result.Response.ContentType)
	switch {
	case strings.Contains(ct, "json"):
		return bodyStr, "json"
	case strings.Contains(ct, "xml"):
		return bodyStr, "xml"
	case strings.Contains(ct, "html"):
		return bodyStr, "html"
	default:
		return bodyStr, "plaintext"
	}
}

func renderHeaders(result request.ResultInfo, s styles) string {
	var lines []string
	for k, v := range result.Response.Headers {
		lines = append(lines, s.label.Render(k+":")+" "+s.value.Render(strings.Join(v, ", ")))
	}
	return strings.Join(lines, "\n")
}

func renderResponseBox(result request.ResultInfo, headersSection, bodyStr, lexer string, s styles) string {
	var bodyBuf bytes.Buffer
	if err := quick.Highlight(&bodyBuf, bodyStr, lexer, "terminal", "github"); err != nil {
		bodyBuf.Reset()
		bodyBuf.WriteString(bodyStr)
	}

	meta := []string{
		s.label.Render("Status:") + " " + s.value.Render(fmt.Sprintf("%d %s", result.Response.StatusCode, httpStatusText(result.Response.Status))),
		s.label.Render("Time:") + "   " + s.value.Render(result.Response.Elapsed.String()),
	}
	if ct := strings.TrimSpace(result.Response.ContentType); ct != "" {
		meta = append(meta, s.label.Render("Type:")+"   "+s.value.Render(ct))
	}
	content := strings.Join(meta, "\n")
	if headersSection != "" {
		content += "\n" + s.label.Render("Headers:") + "\n" + headersSection
	}
	content += "\n" + s.label.Render("Body:") + "\n" + s.codeBox.Render(bodyBuf.String())
	return s.title.Render("Response") + "\n" + s.box.Render(content)
}

func layout(reqBox, respBox string) string {
	return lipgloss.JoinVertical(lipgloss.Left, reqBox, "\n", respBox)
}

func httpStatusText(status string) string {
	parts := strings.SplitN(status, " ", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return status
}
