package spec

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func DiscoverSpecFiles(root string, patterns []string) ([]string, error) {
	var matches []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(root, path)
		base := filepath.Base(path)

		for _, pat := range patterns {
			if !hasGlob(pat) && filepath.Base(pat) == pat {
				if base == pat {
					matches = append(matches, path)
					break
				}
				continue
			}

			if strings.Contains(pat, string(filepath.Separator)) {
				if ok, _ := filepath.Match(pat, rel); ok {
					matches = append(matches, path)
					break
				}
			} else {
				if ok, _ := filepath.Match(pat, base); ok {
					matches = append(matches, path)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return distinct(matches), nil
}

func hasGlob(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func distinct(in []string) []string {
	m := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
