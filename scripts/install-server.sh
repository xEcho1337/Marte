#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="xEcho1337"
REPO_NAME="Marte"
REPO="https://github.com/${REPO_OWNER}/${REPO_NAME}.git"

if ! command -v docker &>/dev/null; then
    echo "Error: Docker is required. Install it from https://docs.docker.com/engine/install/"
    exit 1
fi

if command -v git &>/dev/null; then
    git clone "$REPO" Marte 2>/dev/null || {
        cd Marte && git pull
    }
else
    echo "Error: git is required. Install it first."
    exit 1
fi

cd Marte

if [ ! -f data/config.yml ]; then
    echo "Creating default config..."
    mkdir -p data
    curl -sSfL "https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/main/backend/config/default.yml" \
        -o data/config.yml
    echo "Edit data/config.yml with your settings."
fi

echo ""
echo "Setup complete!"
echo "  cd $(pwd)"
COMPOSE_CMD="docker compose"
if ! docker compose version &>/dev/null; then
    COMPOSE_CMD="docker-compose"
fi
echo "  ${COMPOSE_CMD} up -d"
