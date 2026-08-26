# Releasing

Releases are cut by pushing a tag; GitHub Actions runs GoReleaser
(`.github/workflows/release.yml`).

```sh
git tag v0.2.0
git push origin v0.2.0
```

That publishes cross-platform archives (`linux`/`darwin`/`windows` ×
`amd64`/`arm64`) plus `checksums.txt` to a GitHub release. The `install.sh`
one-liner in the README pulls the latest of these assets.

## Required secret

Set in the repo (Settings → Secrets and variables → Actions):

| Secret | Needed for | Notes |
|--------|-----------|-------|
| `GH_MODULE_TOKEN` | Building | Fine-grained PAT with **Contents: read** on `Stripy-Horse/stripyhorse-go` (the private Go binding). Remove it once both repos are public. |

`GITHUB_TOKEN` is provided automatically and is what creates the release.

## Local dry run

```sh
goreleaser check                                        # validate the config
goreleaser release --snapshot --clean --skip=publish    # build everything, publish nothing
```
