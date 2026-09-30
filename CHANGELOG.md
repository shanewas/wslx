# Changelog

## [Unreleased]

### Added

- arm64 binaries for Linux and Windows; CI covers ubuntu-24.04-arm.

### Fixed

- Dev builds report `dev` instead of a stale release number.

- Bare invocation (no arguments) exits 0 with usage help instead of 2.

## [0.1.0] - 2026-09-30

### Added

- CLI slice 1: `win`, `wsl`, `pwsh`, `path`, `env`, `doctor`,
  `version`, plus the `--json` envelope for machine-readable output.
- Quoting corpus tests, path/env/doctor unit tests; CI on Ubuntu +
  Windows.
- Release pipeline: GoReleaser (Linux + Windows archives, deb/rpm),
  tag workflow, install docs.

### Fixed

- PowerShell stdin mode uses `-Command -` instead of `-File -`:
  same exit fidelity, clean stdout (no `PS <cwd>>` echo lines).
- `wsl --distro` from inside WSL now fails clean with `doctor`
  guidance when interop is broken, instead of leaking an exec error.
