package request

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	u, _ := url.Parse(baseURL + path)
	q := u.Query()

	for _, p := range ep.Operation.Parameters {
		if p.In == "path" {
			fmt.Printf("Enter %s (%s): ", p.Name, p.Schema.Type)
			var v string
			fmt.Scan(&v)
			path = strings.Replace(path, "{"+p.Name+"}", v, 1)
		}

		if p.In == "query" {
			fmt.Printf("Enter %s (%s) [optional]: ", p.Name, p.Schema.Type)
			var v string
			fmt.Scanln(&v)
			if v != "" {
				q.Set(p.Name, v)
			}
		}
	}

	var rb io.Reader
	var rawBody string

	if ep.Operation.RequestBody != nil {
		fmt.Println("Enter JSON body:")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			rawBody = scanner.Text()
			rb = strings.NewReader(rawBody)
		}
	}

	method := strings.ToUpper(ep.Method)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(method, u.String(), rb)
	if rb != nil {
		req.Header.Set("Content-Type", "application/json")
	}
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
			"method":      method,
			"url":         u.String(),
			"requestBody": rawBody,
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
