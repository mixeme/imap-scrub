#!/usr/bin/env bash

set -euo pipefail

# Cross-compile macOS binaries (works from Linux, macOS, or a Linux container).
# Produces dist/imap-scrub-darwin-amd64 and dist/imap-scrub-darwin-arm64.

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
version="${VERSION:-$(git -C "$project_dir" describe --tags --always --dirty 2>/dev/null || echo dev)}"

cd "$project_dir"
mkdir -p dist

for arch in amd64 arm64; do
	echo "Building darwin/${arch}..."
	CGO_ENABLED=0 GOOS=darwin GOARCH="${arch}" \
		go build -trimpath -ldflags "-s -w -X main.appVersion=${version}" \
		-o "dist/imap-scrub-darwin-${arch}" .
done

echo "Done. Artifacts:"
ls -lh dist/imap-scrub-darwin-*
