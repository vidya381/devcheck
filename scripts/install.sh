#!/usr/bin/env bash

set -e

REPO="vidya381/devcheck"
BIN="devcheck"

# detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)   OS="linux" ;;
  darwin)  OS="darwin" ;;
  mingw* | msys* | cygwin* | windows*)
    echo "This installer does not support Windows."
    echo "Download devcheck-windows-amd64.exe from https://github.com/$REPO/releases/latest"
    echo "then put it somewhere on your PATH."
    exit 1
    ;;
  *)       echo "Unsupported OS: $OS"; exit 1 ;;
esac

# detect arch
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)       echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

# get latest release tag
echo "Fetching latest release..."
TAG=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$TAG" ]; then
  echo "Could not find a release. Check https://github.com/$REPO/releases"
  exit 1
fi

FILENAME="${BIN}-${OS}-${ARCH}"
BASE="https://github.com/$REPO/releases/download/$TAG"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading $BIN $TAG ($OS/$ARCH)..."
curl -fsSL "$BASE/$FILENAME" -o "$TMPDIR/$FILENAME"

# verify the checksum if we have a tool for it and the release published one
SHA_TOOL=""
if command -v sha256sum >/dev/null 2>&1; then
  SHA_TOOL="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA_TOOL="shasum -a 256"
fi

if [ -n "$SHA_TOOL" ] && curl -fsSL "$BASE/checksums.txt" -o "$TMPDIR/checksums.txt" 2>/dev/null; then
  EXPECTED=$(grep " $FILENAME\$" "$TMPDIR/checksums.txt" | awk '{print $1}')
  if [ -z "$EXPECTED" ]; then
    echo "Warning: $FILENAME is not listed in checksums.txt, skipping verification."
  else
    ACTUAL=$($SHA_TOOL "$TMPDIR/$FILENAME" | awk '{print $1}')
    if [ "$EXPECTED" != "$ACTUAL" ]; then
      echo "Checksum mismatch for $FILENAME."
      echo "  expected: $EXPECTED"
      echo "  got:      $ACTUAL"
      exit 1
    fi
    echo "Checksum verified."
  fi
else
  echo "Skipping checksum verification (no checksum tool or no checksums.txt)."
fi

chmod +x "$TMPDIR/$FILENAME"

# pick an install directory
INSTALL_DIR="/usr/local/bin"
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMPDIR/$FILENAME" "$INSTALL_DIR/$BIN"
elif command -v sudo >/dev/null 2>&1; then
  sudo mv "$TMPDIR/$FILENAME" "$INSTALL_DIR/$BIN"
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  mv "$TMPDIR/$FILENAME" "$INSTALL_DIR/$BIN"
  echo
  echo "$INSTALL_DIR is not on your PATH by default. Add this to your shell profile:"
  echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

echo "Installed $BIN $TAG to $INSTALL_DIR/$BIN"
echo "Run: devcheck --help"
