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

## Install

> Pre-release: this repo is still private. Artifacts appear on the
> [Releases](https://github.com/shanewas/wslx/releases) page after the first
> tag. Replace `0.1.0` below with the release you want.

Linux tarball:

```sh
VERSION=0.1.0
curl -sSL -o wslx.tar.gz "https://github.com/shanewas/wslx/releases/download/v${VERSION}/wslx_${VERSION}_Linux_x86_64.tar.gz"
tar xzf wslx.tar.gz wslx
install -m 0755 wslx ~/.local/bin/wslx
```

Windows zip (PowerShell):

```powershell
$Version = "0.1.0"
Invoke-WebRequest -OutFile wslx.zip "https://github.com/shanewas/wslx/releases/download/v$Version/wslx_${Version}_Windows_x86_64.zip"
Expand-Archive wslx.zip "$env:LOCALAPPDATA\wslx"
```

Debian/Ubuntu and Fedora/RHEL packages:

```sh
VERSION=0.1.0
curl -sSL -o wslx.deb "https://github.com/shanewas/wslx/releases/download/v${VERSION}/wslx_${VERSION}_amd64.deb"
sudo dpkg -i wslx.deb
curl -sSL -o wslx.rpm "https://github.com/shanewas/wslx/releases/download/v${VERSION}/wslx_${VERSION}_x86_64.rpm"
sudo rpm -i wslx.rpm
```

From source:

```sh
go install github.com/shanewas/wslx/cmd/wslx@latest
```

Package managers (winget, scoop, chocolatey) are planned at public launch.

## Layout

- `cmd/wslx/` — CLI entry
- `internal/` — side detection, runners, quoting, paths, env, doctor
- `SPEC.md` — full v0.1 specification
- `.github/workflows/` — CI (Ubuntu + Windows)

## Build

Requires Go 1.22+.

```sh
go build ./...
go vet ./...
go test ./...
```

## License

MIT — see [LICENSE](LICENSE).
