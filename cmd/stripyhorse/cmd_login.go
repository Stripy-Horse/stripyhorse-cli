package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func cmdLogin(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	key := fs.String("key", "", "API key (sh_live_…); skips the browser flow")
	apiURL := fs.String("url", "", "override the API base URL")
	webURL := fs.String("web-url", "", "override the website URL used for browser login")
	noBrowser := fs.Bool("no-browser", false, "print the sign-in URL instead of opening a browser")
	fs.Parse(args)

	if *apiURL != "" {
		cfg.BaseURL = *apiURL
	}
	if *webURL != "" {
		cfg.WebURL = *webURL
	}

	if *key != "" {
		cfg.APIKey = strings.TrimSpace(*key)
		if err := cfg.save(); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Saved. Try: stripyhorse printers list")
		return nil
	}
	return browserLogin(cfg, *noBrowser)
}

// browserLogin runs the loopback OAuth flow: stand up a local callback server,
// open the browser to the site's consent page, and wait for it to redirect
// back with the freshly minted key.
func browserLogin(cfg *Config, noBrowser bool) error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting local callback server: %w", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	state, err := randomState()
	if err != nil {
		return err
	}
	host, _ := os.Hostname()

	keyCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- errors.New("state mismatch on callback (possible CSRF)")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if k := q.Get("key"); k != "" {
			fmt.Fprint(w, callbackHTML("You're signed in", "You can close this tab and return to your terminal."))
			keyCh <- k
		} else {
			fmt.Fprint(w, callbackHTML("Sign-in failed", "No key was returned. Please try again."))
			errCh <- errors.New("no key in callback")
		}
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	authURL := fmt.Sprintf("%s/cli/authorize?port=%d&state=%s&host=%s",
		strings.TrimRight(cfg.webURL(), "/"), port, url.QueryEscape(state), url.QueryEscape(host))

	fmt.Fprintln(os.Stderr, "Opening your browser to sign in…")
	fmt.Fprintln(os.Stderr, "  "+authURL)
	if !noBrowser {
		if err := openBrowser(authURL); err != nil {
			fmt.Fprintln(os.Stderr, "(couldn't open a browser automatically — open the URL above)")
		}
	}

	select {
	case k := <-keyCh:
		cfg.APIKey = k
		if err := cfg.save(); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "✓ Logged in. Try: stripyhorse printers list")
		return nil
	case err := <-errCh:
		return err
	case <-time.After(3 * time.Minute):
		return errors.New("timed out waiting for browser sign-in")
	}
}

func randomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func callbackHTML(title, msg string) string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>Stripy Horse CLI</title>
<style>body{margin:0;font:16px/1.6 system-ui,sans-serif;background:#faf9f7;color:#1a1a1a;display:flex;align-items:center;justify-content:center;min-height:100vh}
.card{background:#fff;border:1px solid #e2e2e2;border-radius:10px;padding:30px 34px;text-align:center;max-width:380px}</style>
</head><body><div class="card"><h1 style="font-size:20px;margin:0 0 6px">` + title + `</h1>
<p style="color:#6b6b6b;font-size:14px;margin:0">` + msg + `</p></div></body></html>`
}
