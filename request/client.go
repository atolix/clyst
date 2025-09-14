package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Endpoint struct {
	Method  string
	Path    string
	Summary string
}

func (e Endpoint) Title() string {
	return fmt.Sprintf("%s %s", e.Method, e.Path)
}

func (e Endpoint) Description() string {
	return e.Summary
}

func (e Endpoint) FilterValue() string {
	return e.Path
}

func Send(baseURL string, ep Endpoint) (map[string]interface{}, error) {
	path := ep.Path
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

	output := map[string]interface{}{
		"request": map[string]string{
			"method":  method,
			"path":    path,
			"summary": ep.Summary,
		},
		"response": map[string]interface{}{
			"status": res.StatusCode,
		},
	}

	var result interface{}

	if err := json.Unmarshal(body, &result); err == nil {
		output["response"].(map[string]interface{})["body"] = result
	} else {
		output["response"].(map[string]interface{})["body"] = string(body)
	}

	return output, nil
}
