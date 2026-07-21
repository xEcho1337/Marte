#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="xEcho1337"
REPO_NAME="Marte"
REPO="https://github.com/${REPO_OWNER}/${REPO_NAME}.git"
INSTALL_DIR="Marte"

if ! command -v docker &>/dev/null; then
    echo "Error: Docker is required. Install it from https://docs.docker.com/engine/install/"
    exit 1
fi

if ! command -v git &>/dev/null; then
    echo "Error: git is required. Install it first."
    exit 1
fi

if [ -d "$INSTALL_DIR" ]; then
    if [ -d "$INSTALL_DIR/.git" ]; then
        echo "Marte directory exists, pulling latest changes..."
        git -C "$INSTALL_DIR" pull
    else
        echo "Error: '$INSTALL_DIR' exists but is not a git repository."
        echo "Remove it or rename it and re-run this script."
        exit 1
    fi
else
    echo "Cloning Marte..."
    git clone "$REPO" "$INSTALL_DIR"
fi

if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    echo "Error: docker-compose.yml not found after clone/pull."
    exit 1
fi

if [ ! -f "$INSTALL_DIR/data/config.yml" ]; then
    echo "Creating default config..."
    mkdir -p "$INSTALL_DIR/data"
    curl -sSfL "https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/main/backend/config/default.yml" \
        -o "$INSTALL_DIR/data/config.yml"
    echo "Edit $INSTALL_DIR/data/config.yml with your settings before starting."
fi

COMPOSE_CMD="docker compose"
if ! docker compose version &>/dev/null 2>&1; then
    if command -v docker-compose &>/dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        echo "Error: docker compose or docker-compose not found."
        exit 1
    fi
fi

echo ""
echo "Setup complete!"
echo "  cd $INSTALL_DIR"
echo "  ${COMPOSE_CMD} up -d"
