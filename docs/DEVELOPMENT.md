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

Stable installs download the prebuilt `imap-scrub-<os>-<arch>.tar.gz` binaries that [`build-release.yml`](../.github/workflows/build-release.yml) already attaches to each GitHub release, so `brew install imap-scrub` needs no Go toolchain. `--HEAD` builds `develop` from source instead (`depends_on "go"` only applies there).

When cutting a new release, bump the formula's `version` and the `url`/`sha256` pair for each of the four supported platforms (darwin-arm64, darwin-amd64, linux-amd64, linux-arm64) once the release binaries are uploaded:

```sh
for asset in darwin-arm64 darwin-amd64 linux-amd64 linux-arm64; do
  curl -sL -o "/tmp/imap-scrub-$asset.tar.gz" \
    "https://github.com/mixeme/imap-scrub/releases/download/v<new-version>/imap-scrub-$asset.tar.gz"
  echo "$asset: $(sha256sum "/tmp/imap-scrub-$asset.tar.gz" | awk '{print $1}')"
done
```

Test locally before pushing:

```sh
brew install ./Formula/imap-scrub.rb          # prebuilt binary path
brew install --HEAD ./Formula/imap-scrub.rb   # source build path
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
