package request

import (
	"github.com/atolix/clyst/params"
)

func SavePreset(dir string, ep Endpoint, provider InputProvider) error {
	store, err := params.Load(dir)
	if err != nil {
		return err
	}

	pathVals := map[string]string{}
	queryVals := map[string]string{}
	for _, p := range ep.Operation.Parameters {
		switch p.In {
		case "path":
			pathVals[p.Name] = provider.GetPathParam(p)
		case "query":
			if v := provider.GetQueryParam(p); v != "" {
				queryVals[p.Name] = v
			}
		}
	}

	body := ""
	if ep.Operation.RequestBody != nil {
		body = provider.GetRequestBody()
	}

	return store.AppendPreset(ep.Method, ep.Path, params.StoredParams{
		Path:  pathVals,
		Query: queryVals,
		Body:  body,
	})
}
