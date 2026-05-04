#!/bin/sh
set -e

CLI_NAME="tickert"
OWNER="petl-cli"
REPO="tickert"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  darwin | linux) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

FILENAME="${CLI_NAME}-${OS}-${ARCH}"
URL="https://github.com/${OWNER}/${REPO}/releases/latest/download/${FILENAME}"

echo "Downloading ${CLI_NAME} for ${OS}/${ARCH}..."
curl -fsSL "$URL" -o "/tmp/${CLI_NAME}"
chmod +x "/tmp/${CLI_NAME}"

INSTALL_DIR="/usr/local/bin"
if [ -w "$INSTALL_DIR" ]; then
  mv "/tmp/${CLI_NAME}" "${INSTALL_DIR}/${CLI_NAME}"
else
  sudo mv "/tmp/${CLI_NAME}" "${INSTALL_DIR}/${CLI_NAME}"
fi

echo "${CLI_NAME} installed to ${INSTALL_DIR}/${CLI_NAME}"
