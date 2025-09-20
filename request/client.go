package request

import (
	"encoding/json"
	"fmt"
	"io"
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

type RequestInfo struct {
	Method string
	URL    string
	Body   string
}

type ResponseInfo struct {
	StatusCode  int
	Status      string
	Elapsed     time.Duration
	Headers     http.Header
	ContentType string
	RawBody     []byte
	JSONBody    any // if JSON parse succeeded; otherwise nil
}

type ResultInfo struct {
	Request  RequestInfo
	Response ResponseInfo
}

func Send(ep Endpoint, input InputResult) (ResultInfo, error) {
	req, err := http.NewRequest(strings.ToUpper(ep.Method), input.URL, input.Body)
	if input.Body != nil && strings.TrimSpace(input.RawBody) != "" {
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

	bodyBytes, _ := io.ReadAll(res.Body)
	contentType := res.Header.Get("Content-Type")

	var jsonBody any
	if strings.Contains(contentType, "application/json") ||
		(len(bodyBytes) > 0 && (bodyBytes[0] == '{' || bodyBytes[0] == '[')) {
		var v any
		if err := json.Unmarshal(bodyBytes, &v); err == nil {
			jsonBody = v
		}
	}

	return ResultInfo{
		Request: RequestInfo{
			Method: ep.Method,
			URL:    input.URL,
			Body:   input.RawBody,
		},
		Response: ResponseInfo{
			StatusCode:  res.StatusCode,
			Status:      res.Status,
			Elapsed:     elapsed,
			Headers:     res.Header.Clone(),
			ContentType: contentType,
			RawBody:     bodyBytes,
			JSONBody:    jsonBody,
		},
	}, nil
}
