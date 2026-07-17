#!/usr/bin/env bash

set -euo pipefail

# Build a local binary for the current OS/architecture.

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)

cd "$project_dir"
mkdir -p dist

echo "Building imap-scrub for $(go env GOOS)/$(go env GOARCH)..."
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "dist/imap-scrub" .

echo "Done. Artifact: dist/imap-scrub"
