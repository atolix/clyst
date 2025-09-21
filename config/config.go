package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SpecFiles []string `yaml:"spec_files"`
}

var DefaultSpecNames = []string{
	"api_spec.yml",
	"spec.yml",
	"openapi.yml",
	"openapi.yaml",
}

var DefaultCandicates = []string{
	".clyst.yml",
	".clyst.yaml",
	"clyst.yml",
	"clyst.yaml",
}

func Load() (*Config, error) {
	var path string
	for _, c := range DefaultCandicates {
		if _, err := os.Stat(c); err == nil {
			path = c
			break
		}
	}

	if path == "" {
		return &Config{}, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	for i, p := range cfg.SpecFiles {
		cfg.SpecFiles[i] = filepath.Clean(p)
	}

	return &cfg, nil
}

func DefineSpecNames(cfg *Config) ([]string, error) {
	if cfg == nil {
		return DefaultSpecNames, nil
	}

	if len(cfg.SpecFiles) == 0 {
		return DefaultSpecNames, nil
	}

	out := make([]string, 0, len(cfg.SpecFiles))
	for _, s := range cfg.SpecFiles {
		if s == "" {
			return nil, errors.New("config.spec_files contains an empty entry")
		}
		out = append(out, s)
	}

	return out, nil
}
