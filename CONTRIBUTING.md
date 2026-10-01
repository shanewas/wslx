# Contributing to wslx

## Build and check

Requires Go 1.22+.

```sh
gofmt -l .        # must print nothing
go build ./...
go vet ./...
go test ./...
```

CI runs the same on Ubuntu (amd64, arm64) and Windows, and the tag
workflow runs them again before GoReleaser publishes anything.

## Before tagging: the live battery

GitHub runners have no WSL, so the hop only gets tested by a person. On a
real WSL2 machine, from both sides where it applies:

```sh
wslx doctor                               # all green on both sides
wslx win ipconfig /all                    # streams, exits 0
wslx pwsh 'Write-Output 日本語; exit 3'    # prints 日本語 intact, exits 3
wslx --json pwsh 'Get-Date'               # stdoutEncoding is utf8
wslx wsl --distro <name> -- uname -a
yes | wslx wsl head -1                    # returns at once
```

Note the WSL version, distro and Windows build in the release notes.

## Rules that protect the design

- **argv arrays, never string-built command lines.** Quoting belongs to
  the OS: Go's `exec` on Windows, the interop layer from Linux,
  `-EncodedCommand` for PowerShell. A new launch path gets an
  argv-building function with a unit test, not a string template.
- **Clean-room only.** Do not copy `wslu` (GPLv3) code into this MIT
  codebase. Reimplement helpers from behavior, not from source.
- **`doctor` reports, never mutates.** No auto-remediation; every failure
  prints cause plus the exact fix command, and a probe only runs on the
  side where it means something.
- **Never hang in scripted paths.** `--json` and `pwsh` scripts use
  non-interactive flags, a TTY stdin reads as empty, and failures exit
  promptly with a clear error. Without `--json` the child owns the
  terminal on purpose; waiting for the user there isn't a hang.
- **Same-side rule.** When already on the target side, run directly with
  no hop.

## Docs

`README.md` is the user-facing contract. Behavior changes update the docs
and `CHANGELOG.md` in the same commit.
