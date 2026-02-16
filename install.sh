#!/bin/sh
set -eu

REPO="openhoo/vibecontainer"
INSTALL_DIR="/usr/local/bin"
BINARY="vibecontainer"

main() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "$os" in
    linux)  os="linux" ;;
    darwin) os="darwin" ;;
    *)      err "Unsupported OS: $os" ;;
  esac

  case "$arch" in
    x86_64|amd64)  arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)             err "Unsupported architecture: $arch" ;;
  esac

  tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4)"
  [ -n "$tag" ] || err "Could not determine latest release"

  name="${BINARY}-${os}-${arch}"
  url="https://github.com/${REPO}/releases/download/${tag}/${name}"
  checksums_url="https://github.com/${REPO}/releases/download/${tag}/checksums.txt"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  log "Downloading ${name} ${tag}..."
  curl -fSL -o "${tmpdir}/${name}" "$url"

  log "Verifying checksum..."
  curl -fsSL -o "${tmpdir}/checksums.txt" "$checksums_url"
  expected="$(grep "${name}" "${tmpdir}/checksums.txt" | awk '{print $1}')"
  [ -n "$expected" ] || err "Checksum for ${name} not found"
  actual="$(sha256sum "${tmpdir}/${name}" 2>/dev/null || shasum -a 256 "${tmpdir}/${name}" | awk '{print $1}')"
  actual="$(echo "$actual" | awk '{print $1}')"
  [ "$actual" = "$expected" ] || err "Checksum mismatch: expected ${expected}, got ${actual}"

  chmod +x "${tmpdir}/${name}"

  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmpdir}/${name}" "${INSTALL_DIR}/${BINARY}"
  else
    log "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "${tmpdir}/${name}" "${INSTALL_DIR}/${BINARY}"
  fi

  log "Installed ${BINARY} ${tag} to ${INSTALL_DIR}/${BINARY}"
}

log() { printf '%s\n' "$*"; }
err() { log "Error: $*" >&2; exit 1; }

main
