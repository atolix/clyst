package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/atolix/clyst/spec"
)

type Endpoint struct {
	Method    string
	Path      string
	Operation spec.Operation
}

func Send(ep Endpoint, input InputResult) (map[string]any, error) {
	req, err := http.NewRequest(strings.ToUpper(ep.Method), input.URL, input.Body)
	if input.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	elapsed := time.Since(start)

	var resBody any
	json.NewDecoder(res.Body).Decode(&resBody)

	return map[string]any{
		"request": map[string]any{
			"method": ep.Method,
			"url":    input.URL,
			"body":   input.RawBody,
		},
		"response": map[string]any{
			"status":  res.StatusCode,
			"elapsed": elapsed.String(),
			"body":    resBody,
		},
	}, nil
}
