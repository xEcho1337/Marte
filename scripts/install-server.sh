#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="xEcho1337"
REPO_NAME="Marte"
IMAGE="ghcr.io/${REPO_OWNER}/${REPO_NAME}:latest"
DATA_DIR="${HOME}/.marte"
CONFIG_FILE="${DATA_DIR}/config.yml"
COMPOSE_FILE="${DATA_DIR}/docker-compose.yml"

if ! command -v docker &>/dev/null; then
    echo "Error: Docker is required. Install it from https://docs.docker.com/engine/install/"
    exit 1
fi

mkdir -p "$DATA_DIR"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating default config in ${CONFIG_FILE} ..."
    docker run --rm "$IMAGE" cat config.yml > "$CONFIG_FILE" 2>/dev/null || {
        curl -sSfL \
            "https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/main/backend/config/default.yml" \
            -o "$CONFIG_FILE"
    }
    echo "Config created. Edit ${CONFIG_FILE} with your settings."
fi

if [ ! -f "$COMPOSE_FILE" ]; then
    echo "Creating docker-compose.yml in ${DATA_DIR} ..."
    cat > "$COMPOSE_FILE" <<EOF
services:
  marte:
    image: ${IMAGE}
    container_name: marte
    ports:
      - "14100:14100"
      - "14101:14101"
    volumes:
      - "${DATA_DIR}:/app"
    restart: unless-stopped
EOF
fi

echo "Pulling latest image ..."
docker pull "$IMAGE"

echo ""
echo "Setup complete!"
echo "  Config:  ${CONFIG_FILE}"
echo "  Data:    ${DATA_DIR}"
echo ""
echo "Start the server:"
echo "  cd ${DATA_DIR} && docker compose up -d"
echo ""
echo "View logs:"
echo "  cd ${DATA_DIR} && docker compose logs -f"
echo ""
echo "Stop:"
echo "  cd ${DATA_DIR} && docker compose down"
