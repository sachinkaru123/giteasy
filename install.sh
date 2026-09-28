#!/usr/bin/env bash
# giteasy installer for Linux and macOS.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/<owner>/giteasy/main/install.sh | bash
#
# Downloads the correct prebuilt binary for your OS/architecture from the
# latest GitHub Release and installs it onto your PATH. No Go required.

set -euo pipefail

# ---- Configure this before publishing your repo -----------------------
REPO="sachinkaru123/giteasy"   # <-- change to "your-github-username/giteasy"
# -------------------------------------------------------------------------

BINARY_NAME="giteasy"

info()  { printf "\033[1;34m→\033[0m %s\n" "$1"; }
ok()    { printf "\033[1;32m✔\033[0m %s\n" "$1"; }
fail()  { printf "\033[1;31m✘\033[0m %s\n" "$1" >&2; exit 1; }

# ---- Detect OS ----------------------------------------------------------
OS_RAW="$(uname -s)"
case "$OS_RAW" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *) fail "Unsupported OS: $OS_RAW. Please build from source instead: go install github.com/${REPO}@latest" ;;
esac

# ---- Detect architecture -------------------------------------------------
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) fail "Unsupported architecture: $ARCH_RAW. Please build from source instead: go install github.com/${REPO}@latest" ;;
esac

info "Detected platform: ${OS}/${ARCH}"

# ---- Find latest release tag via GitHub API ------------------------------
info "Looking up latest release ..."
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
TAG=$(curl -sSL "$API_URL" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

if [ -z "${TAG:-}" ]; then
  fail "Could not determine the latest release. Check https://github.com/${REPO}/releases"
fi
ok "Latest version: ${TAG}"

# ---- Build the expected asset name (matches .goreleaser.yaml naming) -----
ASSET="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

info "Downloading ${ASSET} ..."
if ! curl -sSL -f -o "${TMP_DIR}/${ASSET}" "$URL"; then
  fail "Download failed: ${URL}
Check that a release for ${OS}/${ARCH} exists at https://github.com/${REPO}/releases/tag/${TAG}"
fi

info "Extracting ..."
tar -xzf "${TMP_DIR}/${ASSET}" -C "$TMP_DIR"

if [ ! -f "${TMP_DIR}/${BINARY_NAME}" ]; then
  fail "Extracted archive did not contain '${BINARY_NAME}'."
fi
chmod +x "${TMP_DIR}/${BINARY_NAME}"

# ---- Choose an install directory -----------------------------------------
if [ -w "/usr/local/bin" ] 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
  USE_SUDO=1
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

info "Installing to ${INSTALL_DIR} ..."
if [ "${USE_SUDO:-0}" = "1" ]; then
  sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
  mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null || {
    sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  }
fi

ok "giteasy ${TAG} installed to ${INSTALL_DIR}/${BINARY_NAME}"

if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
  echo
  echo "Note: ${INSTALL_DIR} doesn't appear to be on your PATH yet."
  echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
  echo
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  echo
else
  echo
  ok "Run 'giteasy' to get started."
fi
