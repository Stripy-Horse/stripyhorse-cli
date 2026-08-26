package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// cmdTail streams a printer's live job events (Server-Sent Events). The SSE
// endpoint is chi-native, not a typed /v1 op, so this uses plain HTTP; the
// key rides in the query string since EventSource-style streams can't set
// headers.
func cmdTail(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("tail", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() < 1 {
		return errors.New("usage: stripyhorse tail <printer-id>")
	}
	id := fs.Arg(0)
	key := cfg.apiKey()
	if key == "" {
		return errors.New("no API key — run `stripyhorse login`")
	}
	base := strings.TrimRight(cfg.baseURL(), "/")
	endpoint := base + "/v1/printers/" + url.PathEscape(id) + "/events?api_key=" + url.QueryEscape(key)

	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tail failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	fmt.Fprintf(os.Stderr, "tailing %s (ctrl-c to stop)…\n", id)

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for sc.Scan() {
		if line := sc.Text(); strings.HasPrefix(line, "data: ") {
			fmt.Println(strings.TrimPrefix(line, "data: "))
		}
	}
	return sc.Err()
}
