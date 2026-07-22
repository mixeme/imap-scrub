# Development

Notes for contributors and anyone building or hacking on IMAP-Scrub.

[![Go Report Card](https://goreportcard.com/badge/github.com/mixeme/imap-scrub)](https://goreportcard.com/report/github.com/mixeme/imap-scrub)

Related docs: [Roadmap](ROADMAP.md) · [Changelog](CHANGELOG.md) · [User README](../README.md)

Requirements: **Go >= 1.23**.


## Install from source

```sh
go install github.com/mixeme/imap-scrub@latest
```


## Building local binaries

The build scripts write artifacts to the `dist` directory and can be run from any
working directory.

GitHub release builds inject the version from the release tag via CI. Local
builds report `dev` unless you pass `-ldflags "-X main.appVersion=<version>"`
(or set `VERSION` when running `scripts/build-macos.sh`).

### Current OS / architecture

```sh
bash scripts/build.sh
```

### Windows (amd64)

```bat
scripts\build-windows.bat
```

### Linux (amd64)

```sh
bash scripts/build-linux.sh
```

### Linux (Docker)

```sh
bash scripts/build-linux-docker.sh
```

### macOS (cross-compile)

Cross-compiles amd64 + arm64; works from Linux or a container:

```sh
bash scripts/build-macos.sh
```

This writes:

- `dist/imap-scrub-darwin-amd64` (Intel Macs)
- `dist/imap-scrub-darwin-arm64` (Apple Silicon)


## Homebrew formula

[`Formula/imap-scrub.rb`](../Formula/imap-scrub.rb) lets users `brew tap mixeme/imap-scrub https://github.com/mixeme/imap-scrub && brew install imap-scrub` directly from this repo — no separate `homebrew-*` tap repo needed.

When cutting a new release, bump the formula's `url` (tag) and `sha256` to match:

```sh
curl -sL -o /tmp/imap-scrub.tar.gz https://github.com/mixeme/imap-scrub/archive/refs/tags/v<new-version>.tar.gz
sha256sum /tmp/imap-scrub.tar.gz
```

Test locally before pushing:

```sh
brew install --build-from-source ./Formula/imap-scrub.rb
brew test imap-scrub
brew audit --strict --online ./Formula/imap-scrub.rb
```


## Nix flake

With [Nix](https://nixos.org/) and flakes enabled:

```sh
nix build
```

For a development shell (Go toolchain):

```sh
nix develop
```

With [direnv](https://direnv.net/), `.envrc` loads the flake automatically.


## Dev container

For a ready-made Go 1.23 toolchain in VS Code / Cursor, see [`.devcontainer/README.md`](../.devcontainer/README.md).


## Tests

```sh
go test ./...
```
