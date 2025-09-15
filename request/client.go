package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/atolix/clyst/spec"
)

type Endpoint struct {
	Method    string
	Path      string
	Summary   string
	Operation spec.Operation
}

func Send(baseURL string, ep Endpoint) (map[string]any, error) {
	path := ep.Path

	for _, p := range ep.Operation.Parameters {
		if p.In == "path" {
			fmt.Printf("Enter %s (%s): ", p.Name, p.Schema.Type)
			var v string
			fmt.Scan(&v)
			path = strings.Replace(path, "{"+p.Name+"}", v, 1)
		}
	}

	url := baseURL + path
	method := strings.ToUpper(ep.Method)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println("Error Creating Request:", err)
		os.Exit(1)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	output := map[string]any{
		"request": map[string]string{
			"method":  method,
			"path":    path,
			"summary": ep.Summary,
		},
		"response": map[string]any{
			"status": res.StatusCode,
		},
	}

	var result any

	if err := json.Unmarshal(body, &result); err == nil {
		output["response"].(map[string]any)["body"] = result
	} else {
		output["response"].(map[string]any)["body"] = string(body)
	}

	return output, nil
}
