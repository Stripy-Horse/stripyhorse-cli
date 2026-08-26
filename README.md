# stripyhorse

The command-line client for the [Stripy Horse](https://stripyhorse.io) API —
convert documents to ZPL, render ZPL to PNG, and drive virtual Zebra printers,
all from the terminal. Pipe-friendly and CI-ready.

It's a thin client over the `/v1` API, built on the generated Go SDK
([stripyhorse-go](https://github.com/Stripy-Horse/stripyhorse-go)).

## Install

```sh
go install github.com/Stripy-Horse/stripyhorse-cli@latest
# installs the `stripyhorse-cli` binary; rename to `stripyhorse` if you like,
# or use the release archives / Homebrew tap once published.
```

## Auth

```sh
stripyhorse login                 # prompts for your key, saves to ~/.config/stripyhorse
# or, non-interactive:
export STRIPYHORSE_API_KEY=sh_live_…
```

Point at a different host with `STRIPYHORSE_API_URL` (default
`https://api.stripyhorse.io`).

## Examples

```sh
# Convert a PDF (or image) to ZPL
stripyhorse convert invoice.pdf -o label.zpl --preset 4x6

# Render ZPL to a PNG (reads a file or stdin)
stripyhorse render label.zpl -o preview.png
cat label.zpl | stripyhorse render -o preview.png

# Virtual printers
stripyhorse printers create --size 4x6 --mode persistent --anonymize
stripyhorse printers list
stripyhorse printers delete prt_xxxxxxxx

# Send ZPL to a printer you created (uses its cached ingest URL)
stripyhorse print --printer prt_xxxxxxxx label.zpl
cat label.zpl | stripyhorse print --printer prt_xxxxxxxx

# Watch a printer's jobs arrive live
stripyhorse tail prt_xxxxxxxx
```

### In CI

```sh
stripyhorse convert label.pdf | stripyhorse print --printer "$CI_PRINTER"
```

Every command reads stdin / writes stdout and exits non-zero on failure, so it
drops into shell pipelines and CI gates.

## Commands

| Command | What it does |
|---------|--------------|
| `login` | Save your API key |
| `convert <file>` | Convert a PDF or image to ZPL |
| `render [file]` | Render ZPL (file or stdin) to a PNG |
| `printers list\|create\|delete` | Manage virtual printers |
| `print [file]` | Send ZPL to a printer's ingest endpoint |
| `tail <id>` | Stream a printer's job events (SSE) |
| `version` | Print the CLI version |
