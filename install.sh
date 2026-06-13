#!/bin/sh
# hookfire installer: downloads the latest release binary for this OS/arch from
# GitHub Releases and installs it to $HOME/.local/bin (override with HOOKFIRE_BIN_DIR).
#
#   curl -fsSL https://raw.githubusercontent.com/knakul853/hookfire/main/install.sh | sh
set -eu

REPO="knakul853/hookfire"
BIN_DIR="${HOOKFIRE_BIN_DIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) echo "hookfire: unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux | darwin) ;;
  *) echo "hookfire: unsupported OS: $os (use the Windows zip from Releases)" >&2; exit 1 ;;
esac

tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
  grep '"tag_name":' | head -n1 | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "${tag:-}" ]; then
  echo "hookfire: could not determine the latest release tag" >&2
  exit 1
fi
version="${tag#v}"

archive="hookfire_${version}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$tag/$archive"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
echo "hookfire: downloading $archive ($tag)" >&2
curl -fsSL "$url" -o "$tmp/$archive"
tar -xzf "$tmp/$archive" -C "$tmp"

mkdir -p "$BIN_DIR"
install -m 0755 "$tmp/hookfire" "$BIN_DIR/hookfire"
echo "hookfire: installed to $BIN_DIR/hookfire" >&2

case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) echo "hookfire: add $BIN_DIR to your PATH to run 'hookfire'." >&2 ;;
esac
