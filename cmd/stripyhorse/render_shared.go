package main

import (
	"strings"

	stripyhorse "github.com/Stripy-Horse/stripyhorse-go"
)

// renderPNG renders a ZPL string to PNG bytes via the API (the first label).
// Shared by `render` and `view`.
func (c *Config) renderPNG(zpl, preset string, dpmm int) ([]byte, error) {
	client, ctx, err := c.apiClient()
	if err != nil {
		return nil, err
	}
	body := stripyhorse.NewRenderInputBody(strings.TrimSpace(zpl))
	if preset != "" {
		body.SetPreset(preset)
	}
	if dpmm > 0 {
		body.SetDpmm(int64(dpmm))
	}
	png, _, err := client.RenderAPI.RenderZplPng(ctx).RenderInputBody(*body).Execute()
	if err != nil {
		return nil, apiError(err)
	}
	return []byte(png), nil
}
