// Command stripyhorse is the command-line client for the Stripy Horse API:
// convert documents to ZPL, render ZPL to PNG, and drive virtual Zebra
// printers — all from the terminal, pipe-friendly, over /v1.
package main

import (
	"fmt"
	"os"
)

// version is overridden at build time via -ldflags "-X main.version=…".
var version = "dev"

type command struct {
	name    string
	summary string
	run     func(cfg *Config, args []string) error
}

var commands = []command{
	{"login", "Save your API key", cmdLogin},
	{"convert", "Convert a PDF or image file to ZPL", cmdConvert},
	{"render", "Render ZPL to a PNG", cmdRender},
	{"view", "Render ZPL and show it inline in the terminal", cmdView},
	{"printers", "Manage virtual printers (list|create|delete)", cmdPrinters},
	{"print", "Send ZPL to a real printer (IP) or a virtual one (prt_…)", cmdPrint},
	{"tail", "Stream a printer's job events live", cmdTail},
	{"version", "Print the CLI version", cmdVersion},
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	name := os.Args[1]
	if name == "-h" || name == "--help" || name == "help" {
		usage()
		return
	}
	for _, c := range commands {
		if c.name == name {
			cfg, err := loadConfig()
			if err != nil {
				fatal(err)
			}
			if err := c.run(cfg, os.Args[2:]); err != nil {
				fatal(err)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "unknown command %q\n\n", name)
	usage()
	os.Exit(2)
}

func usage() {
	fmt.Fprintf(os.Stderr, "stripyhorse — command-line client for the Stripy Horse API\n\n")
	fmt.Fprintf(os.Stderr, "Usage: stripyhorse <command> [flags]\n\n")
	fmt.Fprintln(os.Stderr, "Commands:")
	for _, c := range commands {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", c.name, c.summary)
	}
	fmt.Fprintln(os.Stderr, "\nAuth: `stripyhorse login`, or set STRIPYHORSE_API_KEY.")
	fmt.Fprintln(os.Stderr, "Point at another host with STRIPYHORSE_API_URL (default "+defaultBaseURL+").")
}

func cmdVersion(_ *Config, _ []string) error {
	fmt.Println("stripyhorse " + version)
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error: "+err.Error())
	os.Exit(1)
}
