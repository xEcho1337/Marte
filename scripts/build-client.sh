#!/bin/sh
set -e
cd "$(dirname "$0")/.."
mkdir -p client/out
cd client
CGO_ENABLED=0 go build -ldflags="-s -w" -o out/marte .
echo "Built client/out/marte"
