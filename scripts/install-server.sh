#!/usr/bin/env bash
set -euo pipefail

REPO="https://github.com/xEcho1337/Marte.git"

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
    docker run --rm ghcr.io/xecho1337/marte:latest cat config.yml > data/config.yml 2>/dev/null || {
        curl -sSfL "https://raw.githubusercontent.com/xEcho1337/Marte/main/backend/config/default.yml" \
            -o data/config.yml
    }
    echo "Edit data/config.yml with your settings."
fi

echo ""
echo "Setup complete!"
echo "  cd $(pwd)"
echo "  docker compose up -d"
