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

type CancelAware interface {
	Canceled() bool
}

func AssembleInput(baseURL string, ep Endpoint, provider InputProvider) (InputResult, bool, error) {
	if ca, ok := provider.(CancelAware); ok && ca.Canceled() {
		return InputResult{}, true, nil
	}

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

	if ca, ok := provider.(CancelAware); ok && ca.Canceled() {
		return InputResult{}, true, nil
	}
	u.RawQuery = q.Encode()

	var rawBody string
	if ep.Operation.RequestBody != nil {
		rawBody = provider.GetRequestBody()
		if ca, ok := provider.(CancelAware); ok && ca.Canceled() {
			return InputResult{}, true, nil
		}
	}

	return InputResult{
		URL:     u.String(),
		RawBody: rawBody,
		Body:    strings.NewReader(rawBody),
	}, false, nil
}
