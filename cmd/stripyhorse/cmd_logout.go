package main

import (
	"flag"
	"fmt"
	"os"
)

// cmdLogout forgets the saved API key and cached printer ingest URLs. Endpoint
// overrides (API/web URLs) are preserved — logout is about credentials, not
// which server you point at.
func cmdLogout(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("logout", flag.ExitOnError)
	fs.Parse(args)

	had := cfg.APIKey != "" || len(cfg.Ingest) > 0
	cfg.APIKey = ""
	cfg.Ingest = map[string]string{}
	if err := cfg.save(); err != nil {
		return err
	}

	if os.Getenv("STRIPYHORSE_API_KEY") != "" {
		fmt.Fprintln(os.Stderr, "Cleared the saved key — but STRIPYHORSE_API_KEY is still set in your environment; unset it to fully log out.")
		return nil
	}
	if had {
		fmt.Fprintln(os.Stderr, "Logged out.")
	} else {
		fmt.Fprintln(os.Stderr, "Already logged out.")
	}
	return nil
}
