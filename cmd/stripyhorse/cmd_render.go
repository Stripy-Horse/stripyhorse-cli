package main

import (
	"flag"
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
	png, err := cfg.renderPNG(string(zpl), *preset, *dpmm)
	if err != nil {
		return err
	}
	return writeOut(*out, png)
}
