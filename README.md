# wslx

One static binary, installed on both Windows and WSL. Makes the other side
feel native: Linux commands from PowerShell, PowerShell from WSL. No manual
copy-paste, no quoting voodoo.

> Status: v0.1.0 released. The CLI slice (`win`, `wsl`, `pwsh`, `path`,
> `env`, `doctor`) is covered by CI on Ubuntu + Windows and verified
> live across the boundary (`wslx doctor` reports all green on a
> healthy instance).

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

> Artifacts live on the [Releases](https://github.com/shanewas/wslx/releases)
> page. Replace `0.1.0` below with the release you want.

Linux tarball:

```sh
VERSION=0.1.0
curl -sSL -o wslx.tar.gz "https://github.com/shanewas/wslx/releases/download/v${VERSION}/wslx_${VERSION}_linux_amd64.tar.gz"
tar xzf wslx.tar.gz wslx
install -m 0755 wslx ~/.local/bin/wslx
```

Windows zip (PowerShell):

```powershell
$Version = "0.1.0"
Invoke-WebRequest -OutFile wslx.zip "https://github.com/shanewas/wslx/releases/download/v$Version/wslx_${Version}_windows_amd64.zip"
Expand-Archive wslx.zip "$env:LOCALAPPDATA\wslx"
```

Debian/Ubuntu and Fedora/RHEL packages:

```sh
VERSION=0.1.0
curl -sSL -o wslx.deb "https://github.com/shanewas/wslx/releases/download/v${VERSION}/wslx_${VERSION}_linux_amd64.deb"
sudo dpkg -i wslx.deb
curl -sSL -o wslx.rpm "https://github.com/shanewas/wslx/releases/download/v${VERSION}/wslx_${VERSION}_linux_amd64.rpm"
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
