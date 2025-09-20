package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "sort"
    "strings"

    "github.com/atolix/clyst/request"
    "github.com/atolix/clyst/spec"
    "github.com/atolix/clyst/tui"

    "github.com/alecthomas/chroma/quick"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/lipgloss"
)

func main() {
	spec, err := spec.Load("api_spec.yml")
	if err != nil {
		panic(err)
	}

	var endpoints []tui.EndpointItem
	for path, methods := range spec.Paths {
		for method, op := range methods {
			endpoints = append(endpoints, tui.EndpointItem{
				Method:    method,
				Path:      path,
				Operation: op,
			})
		}
	}

	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Method == endpoints[j].Method {
			return endpoints[i].Path < endpoints[j].Path
		}
		return endpoints[i].Method < endpoints[j].Method
	})

	var items []list.Item
	for _, ep := range endpoints {
		items = append(items, ep)
	}

	selected, err := tui.Run(items)
	if err != nil {
		fmt.Println("TUI running error:", err)
		os.Exit(1)
	}

	if selected == nil {
		fmt.Println("No endpoint selected")
		return
	}

	ep := request.Endpoint{
		Method:    selected.Method,
		Path:      selected.Path,
		Operation: selected.Operation,
	}

	baseURL := "https://jsonplaceholder.typicode.com"
	input, err := request.AssembleInput(baseURL, ep, request.CLIInput{})
	if err != nil {
		panic(err)
	}

    result, err := request.Send(ep, input)
    if err != nil {
        panic(err)
    }

    // Styles for lipgloss output
    title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87cefa"))
    label := lipgloss.NewStyle().Foreground(lipgloss.Color("#8a8f98")).Bold(true)
    value := lipgloss.NewStyle().Foreground(lipgloss.Color("#e6e6e6"))
    box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6495ed")).Padding(1, 2)
    codeBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#5f87af")).Padding(0, 1).MarginTop(0)

    // Request box content
    reqLines := []string{
        label.Render("Method:") + " " + value.Render(strings.ToUpper(result.Request.Method)),
        label.Render("URL:") + "    " + value.Render(result.Request.URL),
    }
    if strings.TrimSpace(result.Request.Body) != "" {
        // Pretty print request JSON body if possible
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

    // Prepare response body with syntax highlighting
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
        switch {
        case strings.Contains(strings.ToLower(result.Response.ContentType), "json"):
            lexer = "json"
        case strings.Contains(strings.ToLower(result.Response.ContentType), "xml"):
            lexer = "xml"
        case strings.Contains(strings.ToLower(result.Response.ContentType), "html"):
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

    // Headers section
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

    // Print joined layout
    out := lipgloss.JoinVertical(lipgloss.Left, reqBox, "\n", respBox)
    fmt.Println(out)
}

// httpStatusText extracts the textual status after the code, if available.
func httpStatusText(status string) string {
    // status is usually like "200 OK"; return the trailing part.
    parts := strings.SplitN(status, " ", 2)
    if len(parts) == 2 {
        return parts[1]
    }
    return status
}
