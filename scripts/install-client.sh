#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="xEcho1337"
REPO_NAME="Marte"
BINARY_NAME="marte"
INSTALL_DIR="${HOME}/.local/bin"

OS=$(uname -s)
case "$OS" in
    Linux)  OS_TARGET="linux" ;;
    Darwin) OS_TARGET="darwin" ;;
    *)
        echo "Unsupported OS: $OS. This script only supports Linux and macOS."
        exit 1
        ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH_TARGET="amd64" ;;
    aarch64|arm64) ARCH_TARGET="arm64" ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Detected: $OS_TARGET/$ARCH_TARGET"

TARGET_ASSET="marte-client-${OS_TARGET}-${ARCH_TARGET}"

echo "Fetching latest release ..."
LATEST=$(curl -sSf "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest")

DOWNLOAD_URL=$(echo "$LATEST" | grep -oE '"browser_download_url": *"[^"]*'"${TARGET_ASSET}"'[^"]*"' | grep -oE 'https://[^"]+' || true)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: asset '${TARGET_ASSET}' not found in latest release."
    echo "Available assets:"
    echo "$LATEST" | grep -oE '"name": *"[^"]*"' | sed 's/"name": *"\(.*\)"/  \1/' || true
    exit 1
fi

echo "Downloading $TARGET_ASSET ..."

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -sSfL "$DOWNLOAD_URL" -o "$TMP_DIR/$BINARY_NAME"

mkdir -p "$INSTALL_DIR"
mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "Installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"

case ":$PATH:" in
    *:"$INSTALL_DIR":*) ;;
    *)
        echo ""
        echo "Warning: $INSTALL_DIR is not in your PATH."
        echo "Add it by running:"
        echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ${HOME}/.bashrc"
        echo "  source ${HOME}/.bashrc"
        if [ "$OS" = "Darwin" ]; then
            echo "Or if using zsh:"
            echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ${HOME}/.zshrc"
            echo "  source ${HOME}/.zshrc"
        fi
        ;;
esac

echo "Done."
