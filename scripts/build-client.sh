#!/bin/sh
set -e
cd "$(dirname "$0")/.."
mkdir -p client/out
cd client
go build -o out/marte .
echo "Built client/out/marte"
