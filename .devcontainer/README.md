# Dev Container

This repository includes a VS Code / Cursor dev container configuration for Go development.

## What it provides

- Go 1.23 toolchain in a reproducible container
- Automatic dependency download (`go mod download`) on first create
- Recommended VS Code extensions and Go formatting on save

## Usage

1. Open this repository in VS Code or Cursor.
2. Run the command: **Dev Containers: Reopen in Container**.
3. After the container starts, run `go test ./...`.
