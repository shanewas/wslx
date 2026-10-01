# Changelog

## [Unreleased]

## [0.1.2] - 2026-10-01

### Added

- `doctor` picks its probes by side. Windows checks `wsl.exe`, that the
  default distro boots, and PowerShell. WSL adds the `/etc/wsl.conf`
  interop settings (`enabled`, `appendWindowsPath`) and skips the socket
  probe on WSL1. Plain Linux gets one clear failure instead of six.
- `--version` and `-v`; `--json` accepted after the command for `path`,
  `env`, `doctor` and `version`.
- README: `--json` envelope, encoding notes, a stability statement, and
  install steps that resolve the latest release, verify `checksums.txt`
  and put `wslx` on PATH.

### Changed

- `win`, `wsl` and `pwsh` without `--json` hand the child your terminal:
  output streams, `tail -f` and interactive tools work, Ctrl-C reaches the
  child, and piped stdin passes through instead of being read into memory
  first (`yes | wslx wsl head -1` used to hang).
- `pwsh` ships the script as `-EncodedCommand` with UTF-8 output, so
  quoting and console code pages (cp932) can't mangle it and stdin stays
  free for `$input`. Arguments join with spaces into one script; `;`
  separates statements. The interactive session loads your profile.
- `win` resolves names from WSL via PATH, then `.exe`, then
  `C:\Windows\System32`, so `wslx win ipconfig` works with
  `appendWindowsPath=false`; the `wsl --distro` hop finds `wsl.exe` the
  same way.
- The Windows-side gate needs only the binfmt entry (`WSLInterop`, or the
  `WSLInterop-late` that newer WSL registers), plus a live `WSL_INTEROP`
  socket on WSL2. WSL1 passes the gate.
- `path` on Windows hops through `wsl.exe -e wslpath` for paths inside the
  distro; `wslpath` failures show its stderr.
- `version` without the release ldflag reports what the Go toolchain
  recorded (module version for `go install`, pseudo-version for local
  builds) instead of `dev`.
- Exit codes keep their 32-bit value on Windows; the 0-255 mask applies on
  Unix only.
- Release workflow runs build, vet and tests before GoReleaser; CI checks
  `gofmt`.

### Removed

- `internal/quote.WindowsCommandLine` and its corpus test: nothing called
  it. Quoting is the OS's job (Go `exec` on Windows, WSL interop from
  Linux).

### Fixed

- README examples that couldn't work: `win Get-ChildItem` (a cmdlet, not
  an executable) and `wsl ls ~/projects` from PowerShell (no shell expands
  `~` there).
- Bare invocation (no arguments) exits 0 with usage help instead of 2.

## [0.1.1] - 2026-09-30

### Added

- arm64 binaries for Linux and Windows; CI covers ubuntu-24.04-arm.

### Fixed

- Dev builds report `dev` instead of a stale release number.

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
