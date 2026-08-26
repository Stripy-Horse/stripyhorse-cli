package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// cmdPrint sends ZPL to a printer's HTTPS ingest endpoint. Ingest is a raw
// capability URL (not part of the typed /v1 spec), so this uses plain HTTP —
// the token in the URL is the auth. The URL is the one cached when this CLI
// created the printer.
func cmdPrint(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	printer := fs.String("printer", "", "printer id (one you created with this CLI)")
	ingestURL := fs.String("ingest-url", "", "ingest URL to POST to (overrides --printer)")
	fs.Parse(args)

	zpl, err := readInput(fs.Arg(0)) // file arg, or stdin when omitted
	if err != nil {
		return err
	}
	u := *ingestURL
	if u == "" && *printer != "" {
		u = cfg.Ingest[*printer]
	}
	if u == "" {
		return errors.New("need --printer <id created with this CLI> or --ingest-url <url>")
	}

	resp, err := http.Post(u, "text/plain", bytes.NewReader(zpl))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ingest failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	fmt.Fprintln(os.Stderr, strings.TrimSpace(string(b)))
	return nil
}
