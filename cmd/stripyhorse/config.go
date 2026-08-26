package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	defaultBaseURL = "https://api.stripyhorse.io"
	defaultWebURL  = "https://stripyhorse.io"
)

// Config is the on-disk CLI state: the API key, the API/web base URLs, and the
// ingest URLs of printers this CLI created (the ingest token is only ever
// returned once, at creation, so `print` relies on this cache).
type Config struct {
	APIKey  string            `json:"apiKey,omitempty"`
	BaseURL string            `json:"baseUrl,omitempty"`
	WebURL  string            `json:"webUrl,omitempty"`
	Ingest  map[string]string `json:"ingest,omitempty"` // printerID -> ingest URL
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "stripyhorse", "config.json"), nil
}

func loadConfig() (*Config, error) {
	cfg := &Config{Ingest: map[string]string{}}
	path, err := configPath()
	if err != nil {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	_ = json.Unmarshal(data, cfg)
	if cfg.Ingest == nil {
		cfg.Ingest = map[string]string{}
	}
	return cfg, nil
}

func (c *Config) save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// apiKey resolves the key: env wins over the config file.
func (c *Config) apiKey() string {
	if k := os.Getenv("STRIPYHORSE_API_KEY"); k != "" {
		return k
	}
	return c.APIKey
}

// baseURL resolves the API base URL: env, then config, then the default.
func (c *Config) baseURL() string {
	if u := os.Getenv("STRIPYHORSE_API_URL"); u != "" {
		return u
	}
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return defaultBaseURL
}

// webURL resolves the website base URL, where browser login and the OAuth
// session cookie live (distinct from the API host).
func (c *Config) webURL() string {
	if u := os.Getenv("STRIPYHORSE_WEB_URL"); u != "" {
		return u
	}
	if c.WebURL != "" {
		return c.WebURL
	}
	return defaultWebURL
}
