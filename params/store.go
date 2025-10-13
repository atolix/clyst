package params

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

const defaultFilename = ".clyst_params"

type StoredParams struct {
	Path  map[string]string `json:"path,omitempty"`
	Query map[string]string `json:"query,omitempty"`
	Body  string            `json:"body,omitempty"`
}

type Store struct {
	path string
	data map[string][]StoredParams
}

func Load(dir string) (*Store, error) {
	if strings.TrimSpace(dir) == "" {
		dir = "."
	}
	fp := filepath.Join(dir, defaultFilename)

	data := map[string][]StoredParams{}
	if b, err := os.ReadFile(fp); err == nil {
		if len(b) > 0 {
			if err := json.Unmarshal(b, &data); err != nil {
				return nil, err
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &Store{
		path: fp,
		data: data,
	}, nil
}

func (s *Store) PresetsFor(method, path string) []StoredParams {
	if s == nil {
		return nil
	}
	key := keyOf(method, path)
	items := s.data[key]
	out := make([]StoredParams, 0, len(items))
	for _, item := range items {
		out = append(out, StoredParams{
			Path:  cloneMap(item.Path),
			Query: cloneMap(item.Query),
			Body:  item.Body,
		})
	}
	return out
}

func (s *Store) AppendPreset(method, path string, preset StoredParams) error {
	if s == nil {
		return nil
	}
	key := keyOf(method, path)
	preset.Path = cloneMap(preset.Path)
	preset.Query = cloneMap(preset.Query)
	s.data[key] = append(s.data[key], preset)
	return s.persist()
}

func (s *Store) persist() error {
	payload, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, payload, 0o644)
}

func keyOf(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func cloneMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	maps.Copy(dst, src)
	return dst
}
