package main

import (
	"flag"
	"os"
)

// cmdView renders ZPL to a PNG and shows it inline in the terminal (Kitty /
// Ghostty / iTerm2 / WezTerm graphics protocols, with a truecolor half-block
// fallback). Piped (non-TTY) output gets the raw PNG so it stays composable.
func cmdView(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("view", flag.ExitOnError)
	preset := fs.String("preset", "", "label size preset, e.g. 4x6")
	dpmm := fs.Int("dpmm", 0, "print density in dots/mm (6,8,12,24)")
	width := fs.Int("width", 80, "max width in terminal columns (half-block fallback only)")
	out := fs.String("o", "", "save the PNG to a file instead of displaying it")
	fs.Parse(args)

	zpl, err := readInput(fs.Arg(0)) // file arg, or stdin when omitted
	if err != nil {
		return err
	}
	png, err := cfg.renderPNG(string(zpl), *preset, *dpmm)
	if err != nil {
		return err
	}

	if *out != "" {
		return writeOut(*out, png)
	}
	if !isTTY(os.Stdout) {
		_, err := os.Stdout.Write(png) // piped: emit raw PNG, stay composable
		return err
	}
	return displayImage(os.Stdout, png, *width)
}
