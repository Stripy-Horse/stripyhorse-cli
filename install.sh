#!/bin/sh
# Install the stripyhorse CLI. Usage:
#   curl -fsSL https://raw.githubusercontent.com/Stripy-Horse/stripyhorse-cli/main/install.sh | sh
#
# Env:
#   STRIPYHORSE_VERSION      version to install (default: latest), e.g. v0.1.1
#   STRIPYHORSE_INSTALL_DIR  where to put the binary (default: /usr/local/bin,
#                            falling back to ~/.local/bin if that's not writable)
set -eu

REPO="Stripy-Horse/stripyhorse-cli"
BIN="stripyhorse"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) echo "stripyhorse: unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
	linux | darwin) ;;
	*) echo "stripyhorse: unsupported OS: $os — grab the Windows zip from the Releases page" >&2; exit 1 ;;
esac

version="${STRIPYHORSE_VERSION:-latest}"
if [ "$version" = "latest" ]; then
	version=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
		grep -m1 '"tag_name"' | sed -E 's/.*"tag_name" *: *"([^"]+)".*/\1/')
fi
if [ -z "$version" ]; then
	echo "stripyhorse: could not determine the latest version (set STRIPYHORSE_VERSION)" >&2
	exit 1
fi

ver="${version#v}"
asset="${BIN}_${ver}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$version/$asset"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
echo "stripyhorse: downloading $version ($os/$arch)…" >&2
curl -fsSL "$url" -o "$tmp/pkg.tar.gz"
tar -xzf "$tmp/pkg.tar.gz" -C "$tmp"

dir="${STRIPYHORSE_INSTALL_DIR:-}"
if [ -z "$dir" ]; then
	if [ -w /usr/local/bin ]; then dir=/usr/local/bin; else dir="$HOME/.local/bin"; fi
fi
mkdir -p "$dir"
if ! install -m 0755 "$tmp/$BIN" "$dir/$BIN" 2>/dev/null; then
	cp "$tmp/$BIN" "$dir/$BIN"
	chmod 0755 "$dir/$BIN"
fi

echo "stripyhorse: installed to $dir/$BIN" >&2
case ":$PATH:" in
	*":$dir:"*) ;;
	*) echo "stripyhorse: add it to your PATH →  export PATH=\"$dir:\$PATH\"" >&2 ;;
esac
"$dir/$BIN" version 2>/dev/null || true
