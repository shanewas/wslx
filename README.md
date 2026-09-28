# wslx

One static binary, installed on both Windows and WSL. Makes the other side
feel native: Linux commands from PowerShell, PowerShell from WSL. No manual
copy-paste, no quoting voodoo.

> Status: pre-release scaffold. The plan is settled in [SPEC.md](SPEC.md);
> the CLI is not implemented yet.

## Vision

```sh
# inside WSL: PowerShell without leaving the shell
wslx win Get-ChildItem C:\Users\me
wslx pwsh "Get-Service | Where-Object Status -eq Running"

# inside PowerShell: Linux tools, quoting intact
wslx wsl ls -la ~/projects

# paths just work, health is one command
wslx path --win ~/projects/app
wslx doctor
```

## Layout

- `cmd/wslx/` — CLI entry
- `internal/` — side detection, runners, quoting, paths, env, doctor
- `SPEC.md` — full v0.1 specification
- `.github/workflows/` — CI (Ubuntu + Windows)

## Build

Requires Go 1.24+.

```sh
go build ./...
go vet ./...
go test ./...
```

## License

MIT — see [LICENSE](LICENSE).
