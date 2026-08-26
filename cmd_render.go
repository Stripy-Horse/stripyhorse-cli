package main

import (
	"flag"
	"strings"

	stripyhorse "github.com/Stripy-Horse/stripyhorse-go"
)

func cmdRender(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	out := fs.String("o", "", "output PNG file (default: stdout)")
	preset := fs.String("preset", "", "label size preset, e.g. 4x6")
	dpmm := fs.Int("dpmm", 0, "print density in dots/mm (6,8,12,24)")
	fs.Parse(args)

	zpl, err := readInput(fs.Arg(0)) // file arg, or stdin when omitted
	if err != nil {
		return err
	}

	client, ctx, err := cfg.apiClient()
	if err != nil {
		return err
	}
	body := stripyhorse.NewRenderInputBody(strings.TrimSpace(string(zpl)))
	if *preset != "" {
		body.SetPreset(*preset)
	}
	if *dpmm > 0 {
		body.SetDpmm(int64(*dpmm))
	}
	png, _, err := client.RenderAPI.RenderZplPng(ctx).RenderInputBody(*body).Execute()
	if err != nil {
		return apiError(err)
	}
	return writeOut(*out, []byte(png))
}
