package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	stripyhorse "github.com/Stripy-Horse/stripyhorse-go"
)

// apiClient builds a configured SDK client and an authenticated context
// carrying the bearer key.
func (c *Config) apiClient() (*stripyhorse.APIClient, context.Context, error) {
	key := c.apiKey()
	if key == "" {
		return nil, nil, fmt.Errorf("no API key set — run `stripyhorse login` or set STRIPYHORSE_API_KEY")
	}
	base := strings.TrimRight(c.baseURL(), "/")
	u, err := url.Parse(base)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid API URL %q: %w", base, err)
	}

	cfg := stripyhorse.NewConfiguration()
	cfg.Scheme = u.Scheme
	cfg.Host = u.Host
	cfg.Servers = stripyhorse.ServerConfigurations{{URL: base}}
	cfg.UserAgent = "stripyhorse-cli/" + version

	ctx := context.WithValue(context.Background(), stripyhorse.ContextAccessToken, key)
	return stripyhorse.NewAPIClient(cfg), ctx, nil
}
