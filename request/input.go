package request

import (
	"io"
	"net/url"
	"strings"

	"github.com/atolix/clyst/spec"
)

type InputResult struct {
	URL     string
	RawBody string
	Body    io.Reader
}

type InputProvider interface {
	GetPathParam(p spec.Parameter) string
	GetQueryParam(p spec.Parameter) string
	GetRequestBody() string
}

func AssembleInput(baseURL string, ep Endpoint, provider InputProvider) (InputResult, error) {
	path := ep.Path

	for _, p := range ep.Operation.Parameters {
		if p.In == "path" {
			v := provider.GetPathParam(p)
			path = strings.Replace(path, "{"+p.Name+"}", v, 1)
		}
	}

	u, _ := url.Parse(baseURL + path)
	q := u.Query()

	for _, p := range ep.Operation.Parameters {
		if p.In == "query" {
			v := provider.GetQueryParam(p)
			if v != "" {
				q.Set(p.Name, v)
			}
		}
	}
	u.RawQuery = q.Encode()

	var rawBody string
	if ep.Operation.RequestBody != nil {
		rawBody = provider.GetRequestBody()
	}

	return InputResult{
		URL:     u.String(),
		RawBody: rawBody,
		Body:    strings.NewReader(rawBody),
	}, nil
}
