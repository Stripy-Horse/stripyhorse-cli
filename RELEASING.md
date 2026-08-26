# Releasing

Releases are cut by pushing a tag; GitHub Actions runs GoReleaser
(`.github/workflows/release.yml`).

```sh
git tag v0.2.0
git push origin v0.2.0
```

That publishes cross-platform archives (`linux`/`darwin`/`windows` ×
`amd64`/`arm64`) plus `checksums.txt` to a GitHub release, and — when the tap
token is configured — updates the Homebrew cask.

## Required secrets

Set these in the `stripyhorse-cli` repo settings (Settings → Secrets and
variables → Actions):

| Secret | Needed for | Notes |
|--------|-----------|-------|
| `GH_MODULE_TOKEN` | Building | Fine-grained PAT with **Contents: read** on `Stripy-Horse/stripyhorse-go` (the private Go binding). Drop it once both repos are public. |
| `HOMEBREW_TAP_TOKEN` | Homebrew (optional) | PAT with **Contents: write** on `Stripy-Horse/homebrew-tap`. Without it the release still succeeds — the Homebrew step is skipped. |

## Enabling Homebrew

1. Create an empty repo `Stripy-Horse/homebrew-tap`.
2. Add the `HOMEBREW_TAP_TOKEN` secret.
3. Cut a release. Users then:

   ```sh
   brew install stripy-horse/tap/stripyhorse
   ```

## Local dry run

```sh
goreleaser check                                  # validate the config
goreleaser release --snapshot --clean --skip=publish   # build everything, publish nothing
```
