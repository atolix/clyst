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

// Render returns a pretty, styled string for a SendResult.
func Render(result request.ResultInfo) string {
	// Styles
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87cefa"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#8a8f98")).Bold(true)
	value := lipgloss.NewStyle().Foreground(lipgloss.Color("#e6e6e6"))
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6495ed")).Padding(1, 2)
	codeBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#5f87af")).Padding(0, 1).MarginTop(0)

	// Request box
	reqLines := []string{
		label.Render("Method:") + " " + value.Render(strings.ToUpper(result.Request.Method)),
		label.Render("URL:") + "    " + value.Render(result.Request.URL),
	}
	if strings.TrimSpace(result.Request.Body) != "" {
		var prettyReq bytes.Buffer
		var reqBodyOut string
		if json.Indent(&prettyReq, []byte(result.Request.Body), "", "  ") == nil {
			var reqBuf bytes.Buffer
			_ = quick.Highlight(&reqBuf, prettyReq.String(), "json", "terminal", "github")
			reqBodyOut = reqBuf.String()
		} else {
			reqBodyOut = result.Request.Body
		}
		reqLines = append(reqLines, label.Render("Body:")+"\n"+codeBox.Render(reqBodyOut))
	}
	reqBox := title.Render("Request") + "\n" + box.Render(strings.Join(reqLines, "\n"))

	// Response body prep
	var bodyStr string
	var lexer string
	if result.Response.JSONBody != nil {
		if enc, err := json.MarshalIndent(result.Response.JSONBody, "", "  "); err == nil {
			bodyStr = string(enc)
			lexer = "json"
		}
	}
	if bodyStr == "" {
		bodyStr = string(result.Response.RawBody)
		ct := strings.ToLower(result.Response.ContentType)
		switch {
		case strings.Contains(ct, "json"):
			lexer = "json"
		case strings.Contains(ct, "xml"):
			lexer = "xml"
		case strings.Contains(ct, "html"):
			lexer = "html"
		default:
			lexer = "plaintext"
		}
	}
	var bodyBuf bytes.Buffer
	if err := quick.Highlight(&bodyBuf, bodyStr, lexer, "terminal", "github"); err != nil {
		bodyBuf.Reset()
		bodyBuf.WriteString(bodyStr)
	}

	// Headers
	var headerLines []string
	for k, v := range result.Response.Headers {
		headerLines = append(headerLines, label.Render(k+":")+" "+value.Render(strings.Join(v, ", ")))
	}
	headersSection := strings.Join(headerLines, "\n")

	// Response box
	respMeta := []string{
		label.Render("Status:") + " " + value.Render(fmt.Sprintf("%d %s", result.Response.StatusCode, httpStatusText(result.Response.Status))),
		label.Render("Time:") + "   " + value.Render(result.Response.Elapsed.String()),
	}
	if ct := strings.TrimSpace(result.Response.ContentType); ct != "" {
		respMeta = append(respMeta, label.Render("Type:")+"   "+value.Render(ct))
	}
	respContent := strings.Join(respMeta, "\n")
	if headersSection != "" {
		respContent += "\n" + label.Render("Headers:") + "\n" + headersSection
	}
	respContent += "\n" + label.Render("Body:") + "\n" + codeBox.Render(bodyBuf.String())
	respBox := title.Render("Response") + "\n" + box.Render(respContent)

	// Layout
	return lipgloss.JoinVertical(lipgloss.Left, reqBox, "\n", respBox)
}

func httpStatusText(status string) string {
	parts := strings.SplitN(status, " ", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return status
}
