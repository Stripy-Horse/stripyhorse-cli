# stripyhorse

The command-line client for the **[Stripy Horse](https://stripyhorse.io)** API
— *ZPL, tamed.* Convert documents to ZPL, render ZPL to pixel-true PNGs, and
drive hosted virtual Zebra printers, all from your terminal.

Every command reads **stdin** / writes **stdout** and exits non-zero on
failure, so it drops straight into shell pipelines and CI. It's a thin,
statically-linked client over the `/v1` API, built on the generated Go SDK
([`stripyhorse-go`](https://github.com/Stripy-Horse/stripyhorse-go)).

```sh
stripyhorse convert invoice.pdf | stripyhorse print --printer prt_x9k2
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

Get an API key from your [account page](https://stripyhorse.io/account), then:

```sh
stripyhorse login          # prompts for the key, saves it to your config
```

Or non-interactively / in CI:

```sh
export STRIPYHORSE_API_KEY=sh_live_xxxxxxxxxxxx
```

---

## Commands

| Command | Description |
|---------|-------------|
| `stripyhorse login` | Save your API key |
| `stripyhorse convert <file> [-o out.zpl]` | Convert a PDF or image to ZPL |
| `stripyhorse render [file] [-o out.png]` | Render ZPL (file or stdin) to a PNG |
| `stripyhorse printers list` | List your virtual printers |
| `stripyhorse printers create [flags]` | Create a virtual printer |
| `stripyhorse printers delete <id>` | Delete a printer |
| `stripyhorse print [file] --printer <id>` | Send ZPL to a printer |
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

### Virtual printers

Hosted Zebra printers you can send ZPL to and watch render — no hardware.

```sh
stripyhorse printers create --size 4x6 --mode persistent --anonymize
stripyhorse printers list
stripyhorse printers delete prt_x9k2
```

`--anonymize` masks PII and strips graphics from every captured job before it's
stored.

### Print & tail

```sh
# Send ZPL to a printer you created with this CLI (uses its cached ingest URL)
stripyhorse print --printer prt_x9k2 label.zpl
cat label.zpl | stripyhorse print --printer prt_x9k2

# Or point at any ingest URL directly
stripyhorse print --ingest-url https://api.stripyhorse.io/ingest/pit_… label.zpl

# Watch jobs arrive in real time (Server-Sent Events)
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

Config is stored at `${XDG_CONFIG_HOME:-~/.config}/stripyhorse/config.json`
(`%AppData%\stripyhorse\` on Windows). It holds your key and the ingest URLs of
printers this CLI created — the ingest token is only returned once, at creation,
so `print` relies on that cache. Point at a self-hosted instance with
`STRIPYHORSE_API_URL`.

---

## How it works

`convert`, `render`, and `printers` call the typed `/v1` operations via the
generated Go SDK. `print` and `tail` use plain HTTP: ingest is a capability URL
(the token is the auth) and the event stream is Server-Sent Events — neither is
part of the OpenAPI surface.

## Development

```sh
go build ./cmd/stripyhorse    # build
go vet ./...                  # lint
```

The SDK is regenerated from the API's OpenAPI spec; see
[`stripyhorse-go`](https://github.com/Stripy-Horse/stripyhorse-go).
