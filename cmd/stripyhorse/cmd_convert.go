package main

import (
	"errors"
	"flag"
	"os"
	"strings"
)

func cmdConvert(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	out := fs.String("o", "", "output file (default: stdout)")
	preset := fs.String("preset", "", "label size preset, e.g. 4x6")
	dpmm := fs.Int("dpmm", 0, "print density in dots/mm (6,8,12,24)")
	fs.Parse(args)
	if fs.NArg() < 1 {
		return errors.New("usage: stripyhorse convert <file.pdf|image> [-o out.zpl] [--preset 4x6]")
	}

	f, err := os.Open(fs.Arg(0))
	if err != nil {
		return err
	}
	defer f.Close()

	client, ctx, err := cfg.apiClient()
	if err != nil {
		return err
	}
	req := client.ConvertAPI.ConvertDocument(ctx).File(f)
	if *preset != "" {
		req = req.Preset(*preset)
	}
	if *dpmm > 0 {
		req = req.Dpmm(int64(*dpmm))
	}
	res, _, err := req.Execute()
	if err != nil {
		return apiError(err)
	}

	var sb strings.Builder
	for _, p := range res.GetPages() {
		sb.WriteString(p.GetZpl())
	}
	return writeOut(*out, []byte(sb.String()))
}
