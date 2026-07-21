#!/bin/sh
set -e
cd "$(dirname "$0")/.."
mkdir -p backend/out
cd backend
CGO_ENABLED=0 go build -ldflags="-s -w" -o out/marte .
echo "Built backend/out/marte"
