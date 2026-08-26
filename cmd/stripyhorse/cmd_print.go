package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// cmdPrint sends ZPL to a printer. The target decides how:
//
//	stripyhorse print 192.168.1.50            real printer, raw TCP :9100
//	stripyhorse print 192.168.1.50:9100 f.zpl explicit port, from a file
//	stripyhorse print prt_x9k2 label.zpl      your virtual printer (via ingest)
//	stripyhorse print --ingest-url <url>      an ingest URL directly
//
// ZPL is read from the trailing file argument, or from stdin.
func cmdPrint(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	ingestURL := fs.String("ingest-url", "", "ingest URL to POST to (instead of a target)")
	port := fs.Int("port", 9100, "TCP port for a real printer")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: stripyhorse print <ip|host[:port]|printer-id> [file]")
		fmt.Fprintln(os.Stderr, "  ip/host      a real Zebra printer on your network (raw TCP :9100)")
		fmt.Fprintln(os.Stderr, "  prt_…        one of your virtual printers (via its ingest URL)")
		fmt.Fprintln(os.Stderr, "  file / stdin the ZPL to send")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	rest := fs.Args()
	var target, file string
	if *ingestURL != "" {
		// target-less: the URL is the destination, first arg (if any) is the file.
		if len(rest) >= 1 {
			file = rest[0]
		}
	} else {
		if len(rest) < 1 {
			fs.Usage()
			return errors.New("need a target: a printer IP/host, a prt_… id, or --ingest-url")
		}
		target = rest[0]
		if len(rest) >= 2 {
			file = rest[1]
		}
	}

	zpl, err := readInput(file)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(zpl))) == 0 {
		return errors.New("no ZPL to send (empty file/stdin)")
	}

	switch {
	case *ingestURL != "":
		return sendIngest(*ingestURL, zpl)
	case strings.HasPrefix(target, "prt_"):
		url := cfg.Ingest[target]
		if url == "" {
			return fmt.Errorf("no cached ingest URL for %s — create it with this CLI, or pass --ingest-url", target)
		}
		return sendIngest(url, zpl)
	default:
		return sendRawTCP(target, *port, zpl)
	}
}

// sendRawTCP streams ZPL straight to a real printer's port-9100 socket. No
// read deadline: Zebra printers pause the TCP connection while the head is
// printing, and a timeout would abort a long job mid-run.
func sendRawTCP(host string, port int, zpl []byte) error {
	addr := host
	if !strings.Contains(host, ":") {
		addr = fmt.Sprintf("%s:%d", host, port)
	}
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connecting to printer %s: %w", addr, err)
	}
	defer conn.Close()
	if _, err := conn.Write(zpl); err != nil {
		return fmt.Errorf("sending to printer %s: %w", addr, err)
	}
	fmt.Fprintf(os.Stderr, "sent %d bytes to %s\n", len(zpl), addr)
	return nil
}

func sendIngest(url string, zpl []byte) error {
	resp, err := http.Post(url, "text/plain", bytes.NewReader(zpl))
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
