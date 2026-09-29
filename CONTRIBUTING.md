# Contributing to wslx

## Build and check

Requires Go 1.22+.

```sh
go build ./...
go vet ./...
go test ./...
```

`go vet` must be clean and all tests green before pushing. CI runs the
same three steps on Ubuntu + Windows.

## Rules that protect the design

- **argv arrays plus stdin piping, never string-concatenated commands.**
  Every new quoting path needs a case in the corpus test
  (`internal/quote/quote_test.go`). The corpus is the regression gate.
- **Clean-room only.** Do not copy `wslu` (GPLv3) code into this MIT
  codebase. Reimplement helpers from behavior, not from source.
- **`doctor` reports, never mutates.** No auto-remediation; every failure
  prints cause plus the exact fix command.
- **Never hang.** Cross-boundary calls always use non-interactive flags
  with clean errors on failure. Prove new failure paths exit promptly.
- **Same-side rule.** When already on the target side, run directly with
  no hop.

## Spec

[SPEC.md](SPEC.md) is the design authority. Behavior changes update the
spec in the same commit.
