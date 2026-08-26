package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

func cmdLogin(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	key := fs.String("key", "", "API key (sh_live_…); read from stdin if omitted")
	apiURL := fs.String("url", "", "override the API base URL")
	fs.Parse(args)

	k := *key
	if k == "" {
		fmt.Fprint(os.Stderr, "API key: ")
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		k = strings.TrimSpace(line)
	}
	if k == "" {
		return errors.New("no API key provided")
	}
	cfg.APIKey = k
	if *apiURL != "" {
		cfg.BaseURL = *apiURL
	}
	if err := cfg.save(); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Saved. Try: stripyhorse printers list")
	return nil
}
