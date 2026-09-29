# Changelog

## [Unreleased]

### Added

- CLI slice 1: `win`, `wsl`, `pwsh`, `path`, `env`, `doctor`,
  `version`, plus the `--json` envelope for machine-readable output.
- Quoting corpus tests, path/env/doctor unit tests; CI on Ubuntu +
  Windows.
- Release pipeline: GoReleaser (Linux + Windows archives, deb/rpm),
  tag workflow, install docs.

### Fixed

- `wsl --distro` from inside WSL now fails clean with `doctor`
  guidance when interop is broken, instead of leaking an exec error.
