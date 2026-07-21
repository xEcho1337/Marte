#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="xEcho1337"
REPO_NAME="Marte"
IMAGE="ghcr.io/${REPO_OWNER}/${REPO_NAME}:latest"
DATA_DIR="${HOME}/.marte"
CONFIG_FILE="${DATA_DIR}/config.yml"

if ! command -v docker &>/dev/null; then
    echo "Error: Docker is required. Install it from https://docs.docker.com/engine/install/"
    exit 1
fi

mkdir -p "$DATA_DIR"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating default config in ${CONFIG_FILE} ..."
    docker run --rm "$IMAGE" cat config.yml > "$CONFIG_FILE" 2>/dev/null || {
        echo "Downloading default config from GitHub ..."
        curl -sSfL \
            "https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/main/backend/config/default.yml" \
            -o "$CONFIG_FILE"
    }
    echo "Edit ${CONFIG_FILE} with your settings, then re-run this script."
    exit 0
fi

echo "Pulling latest image ..."
docker pull "$IMAGE"

echo "Starting Marte server ..."

docker rm -f marte 2>/dev/null || true

docker run -d \
    --name marte \
    --restart unless-stopped \
    -p 14100:14100 \
    -p 14101:14101 \
    -v "$DATA_DIR:/app" \
    "$IMAGE"

echo ""
echo "Marte server started!"
echo "  Dashboard: http://localhost:14100"
echo "  TCP port:  14101"
echo "  Data dir:  $DATA_DIR"
echo ""
echo "To view logs:   docker logs -f marte"
echo "To stop:        docker stop marte"
echo "To restart:     docker restart marte"
