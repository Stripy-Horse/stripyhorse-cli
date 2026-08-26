# stripyhorse

The command-line client for the **[Stripy Horse](https://stripyhorse.io)** API
— *ZPL, tamed.* Convert documents to ZPL, render ZPL to pixel-true PNGs, and
drive hosted virtual Zebra printers, all from your terminal.

Every command reads **stdin** / writes **stdout** and exits non-zero on
failure, so it drops straight into shell pipelines and CI. It's a thin,
statically-linked client over the `/v1` API, built on the generated Go SDK
([`stripyhorse-go`](https://github.com/Stripy-Horse/stripyhorse-go)).

```sh
stripyhorse convert invoice.pdf | stripyhorse print 192.168.1.50
```

---

## Install

**With Go** (installs the `stripyhorse` binary into `$(go env GOPATH)/bin`):

```sh
go install github.com/Stripy-Horse/stripyhorse-cli/cmd/stripyhorse@latest
```

**From source:**

```sh
git clone https://github.com/Stripy-Horse/stripyhorse-cli
cd stripyhorse-cli
go install ./cmd/stripyhorse
```

Make sure Go's bin dir is on your `PATH`:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

Verify:

```sh
stripyhorse version
```

> Prebuilt binaries and a Homebrew tap (`brew install stripy-horse/tap/stripyhorse`)
> land with the first tagged release.

---

## Authenticate

```sh
stripyhorse login          # opens your browser to sign in (recommended)
```

`login` starts a tiny local listener, opens the browser to
[stripyhorse.io](https://stripyhorse.io), and — once you approve — captures a
freshly-minted API key named after your machine (revocable any time on your
account page). `--no-browser` prints the URL to open manually.

Prefer a key you already have, or running headless/CI:

```sh
stripyhorse login --key sh_live_xxxxxxxxxxxx    # paste a key
export STRIPYHORSE_API_KEY=sh_live_xxxxxxxxxxxx  # or set it in the environment
```

---

## Commands

| Command | Description |
|---------|-------------|
| `stripyhorse login` | Save your API key |
| `stripyhorse convert <file> [-o out.zpl]` | Convert a PDF or image to ZPL |
| `stripyhorse render [file] [-o out.png]` | Render ZPL (file or stdin) to a PNG |
| `stripyhorse view [file]` | Render ZPL and show it inline in the terminal |
| `stripyhorse printers list` | List your virtual printers |
| `stripyhorse printers create [flags]` | Create a virtual printer |
| `stripyhorse printers delete <id>` | Delete a printer |
| `stripyhorse print <ip\|prt_id> [file]` | Send ZPL to a real printer (raw TCP) or a virtual one |
| `stripyhorse tail <id>` | Stream a printer's job events live |
| `stripyhorse version` | Print the CLI version |

### Convert & render

```sh
# PDF or image → ZPL (each page concatenated)
stripyhorse convert invoice.pdf -o label.zpl --preset 4x6

# ZPL → PNG preview (reads a file, or stdin)
stripyhorse render label.zpl -o preview.png
cat label.zpl | stripyhorse render -o preview.png
```

### View a label right in the terminal

```sh
stripyhorse view label.zpl
cat label.zpl | stripyhorse view
```

Renders the ZPL and shows the label **inline** using your terminal's graphics
protocol — Kitty, Ghostty, iTerm2, and WezTerm display a real image; any other
truecolor terminal gets a unicode half-block rendering. Piped into another
command it emits the raw PNG instead, so `stripyhorse view label.zpl > out.png`
still works.

### Virtual printers

Hosted Zebra printers you can send ZPL to and watch render — no hardware.

```sh
stripyhorse printers create --size 4x6 --mode persistent --anonymize
stripyhorse printers list
stripyhorse printers delete prt_x9k2
```

`--anonymize` masks PII and strips graphics from every captured job before it's
stored.

### Print

`stripyhorse print` sends ZPL to a target — the target decides where it goes.
The ZPL comes from a trailing file argument, or from stdin.

```sh
# A REAL Zebra printer on your network — raw TCP to port 9100, no cloud round-trip
stripyhorse print 192.168.1.50 label.zpl
cat label.zpl | stripyhorse print 192.168.1.50
stripyhorse print 192.168.1.50:9100 label.zpl      # explicit port

# One of YOUR virtual printers (via its cached ingest URL)
stripyhorse print prt_x9k2 label.zpl

# Or an ingest URL directly
stripyhorse print --ingest-url https://api.stripyhorse.io/ingest/pit_… label.zpl
```

Printing to a real printer is pure client-side raw TCP — it needs no API key
and never leaves your network. A `prt_…` target routes through the API to your
hosted virtual printer instead.

### Tail

```sh
# Watch a virtual printer's jobs arrive in real time (Server-Sent Events)
stripyhorse tail prt_x9k2
```

### In CI

Fail the build when a label won't render, or print to a virtual printer as an
integration check:

```sh
stripyhorse convert label.pdf | stripyhorse print --printer "$CI_PRINTER"
```

---

## Configuration

| Env var | Purpose | Default |
|---------|---------|---------|
| `STRIPYHORSE_API_KEY` | API key (overrides the saved one) | — |
| `STRIPYHORSE_API_URL` | API base URL | `https://api.stripyhorse.io` |
| `STRIPYHORSE_WEB_URL` | Website URL used for browser login | `https://stripyhorse.io` |

Config is stored at `${XDG_CONFIG_HOME:-~/.config}/stripyhorse/config.json`
(`%AppData%\stripyhorse\` on Windows). It holds your key and the ingest URLs of
printers this CLI created — the ingest token is only returned once, at creation,
so `print` relies on that cache. Point at a self-hosted instance with
`STRIPYHORSE_API_URL`.

---

## How it works

`convert`, `render`, and `printers` call the typed `/v1` operations via the
generated Go SDK. `print` to a real printer is direct raw TCP to port 9100 (no
API involved); `print` to a virtual printer and `tail` use plain HTTP — ingest
is a capability URL (the token is the auth) and the event stream is Server-Sent
Events, neither of which is part of the OpenAPI surface.

## Development

```sh
go build ./cmd/stripyhorse    # build
go vet ./...                  # lint
```

The SDK is regenerated from the API's OpenAPI spec; see
[`stripyhorse-go`](https://github.com/Stripy-Horse/stripyhorse-go).
