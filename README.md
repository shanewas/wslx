# wslx

One static binary, installed on both Windows and WSL. It runs the other
side for you: Windows tools and PowerShell from bash, Linux commands from
PowerShell, paths translated, and a `doctor` that names the exact command
to repair a broken bridge. No quoting tricks, no hangs.

> Status: v0.1.x. CI builds and tests on Ubuntu (amd64, arm64) and Windows.
> GitHub runners ship without WSL, so the hop itself isn't under CI; run
> the live battery in [CONTRIBUTING.md](CONTRIBUTING.md) on a real WSL2
> machine before tagging. WSL1 is gated on the binfmt entry alone and hasn't
> been exercised live yet.

## Usage

```sh
# inside WSL
wslx win ipconfig /all                        # any Windows executable, output streams
wslx win cmd /c dir 'C:\Users'
wslx pwsh 'Get-Service | Where-Object Status -eq Running'
wslx pwsh                                     # interactive PowerShell in this terminal
wslx path --win ~/projects/app                # \\wsl.localhost\Ubuntu\home\me\projects\app

# inside PowerShell
wslx wsl uname -a
wslx wsl -- bash -lc 'ls ~/projects | wc -l'  # shell features need a shell
wslx path --unix 'C:\Users\me\src'            # /mnt/c/Users/me/src

# anywhere
wslx doctor
wslx --json wsl cat /etc/os-release           # machine-readable envelope
```

Two rules cover most surprises:

- `win` runs executables (`ipconfig`, `cmd`, `code`), not PowerShell
  cmdlets. `Get-ChildItem` is a cmdlet, so hand it to `pwsh`.
- `win` and `wsl` pass arguments straight through with no shell in
  between, which is what keeps quoting safe. It also means `~`, `*`, `|`
  and `$HOME` only expand when you wrap the command in `bash -lc '...'`
  (or `cmd /c` going the other way).

### Commands

- `win [--] <exe> [args]`: run a Windows executable. On Windows, PATH
  decides. From WSL the name is tried as typed, then with `.exe`, then in
  `C:\Windows\System32`, so it keeps working when `/etc/wsl.conf` sets
  `appendWindowsPath=false`.
- `wsl [--distro D] [--] <command>`: run a Linux command. Inside WSL with
  no `--distro` it runs right here, no hop. Otherwise it goes through
  `wsl.exe -e`.
- `pwsh [--] <statements>`: arguments join with spaces into one script
  (`;` separates statements). Pipe a script in with no arguments, or get an
  interactive session on a terminal. Scripts run with `-NoProfile
  -NonInteractive`; the interactive session loads your profile.
- `path --win|--unix|--mixed <p>`: `wslpath` inside WSL. On Windows, drive
  and `\\wsl.localhost` paths translate in-process and paths inside the
  distro go through `wsl.exe`.
- `env show|suggest`: read or build `WSLENV` values.
- `doctor`: see below. `version` (also `--version`).

Without `--json`, the child owns your terminal: output streams as it
happens, `tail -f` and `htop` work, Ctrl-C reaches the child, and piped
stdin passes through untouched.

## doctor

Probes depend on which side you're on. Inside WSL it checks the
`/etc/wsl.conf` interop settings, the kernel's `WSLInterop` binfmt entry
(either spelling, `WSLInterop-late` included), a live `WSL_INTEROP` socket
on WSL2, `wslpath`, `wsl.exe` and PowerShell. On Windows it checks
`wsl.exe`, that the default distro boots, and PowerShell. Each failure
prints its cause and the command that fixes it; nothing gets changed for
you. Exit 0 when all green, 1 otherwise. `wslx doctor --json` gives the
same list as JSON.

## --json

`wslx --json <command>` (or `wslx doctor --json`) prints one JSON object
and still exits with the child's code:

```json
{"side":"wsl","distro":"Ubuntu","exit":0,"exitFull":0,
 "stdout":"...","stderr":"","stdoutEncoding":"utf8","stderrEncoding":"utf8"}
```

A stream that isn't valid UTF-8 comes back base64 with its `*Encoding`
field set to `base64`. `exit` is masked to 0-255 the way a Unix shell sees
it; `exitFull` keeps the 32-bit Windows value. `path` adds `winPath` or
`unixPath`, `doctor` prints the check array, `version` prints
`{"version":...}`. In JSON mode the child never gets your terminal: piped
stdin passes through, a TTY reads as empty, so scripts and agents never
block on a prompt.

## Non-UTF-8 consoles (Japanese Windows and friends)

PowerShell writes to a pipe in the console code page, cp932 on Japanese
Windows, and bash can't read that. `wslx pwsh` sends the script as
`-EncodedCommand` (UTF-16LE, nothing left to misdecode) and switches the
session's output to UTF-8 before your first statement, so 日本語 paths and
output survive the trip. On the Windows side without `--json` it leaves
the console alone; that one already renders Unicode.

## Install

You install it twice: once inside your WSL distro, once on Windows. Each
side runs its own binary and reaches the other through interop.

The snippets fetch the latest release, verify it against the release's
`checksums.txt`, and use `amd64`; swap in `arm64` on ARM machines.

WSL:

```sh
VERSION=$(curl -fsSL https://api.github.com/repos/shanewas/wslx/releases/latest | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p')
TAR="wslx_${VERSION}_linux_amd64.tar.gz"
curl -fsSLO "https://github.com/shanewas/wslx/releases/download/v${VERSION}/${TAR}"
curl -fsSLO "https://github.com/shanewas/wslx/releases/download/v${VERSION}/checksums.txt"
sha256sum --ignore-missing -c checksums.txt
tar xzf "$TAR" wslx
install -Dm0755 wslx ~/.local/bin/wslx     # ~/.local/bin must be on PATH
```

Debian/Ubuntu or Fedora/RHEL packages, same release page:

```sh
curl -fsSLO "https://github.com/shanewas/wslx/releases/download/v${VERSION}/wslx_${VERSION}_linux_amd64.deb" && sudo dpkg -i "wslx_${VERSION}_linux_amd64.deb"
curl -fsSLO "https://github.com/shanewas/wslx/releases/download/v${VERSION}/wslx_${VERSION}_linux_amd64.rpm" && sudo rpm -i "wslx_${VERSION}_linux_amd64.rpm"
```

Windows (PowerShell):

```powershell
$Version = (Invoke-RestMethod https://api.github.com/repos/shanewas/wslx/releases/latest).tag_name.TrimStart('v')
$Zip = "wslx_${Version}_windows_amd64.zip"
Invoke-WebRequest -OutFile $Zip "https://github.com/shanewas/wslx/releases/download/v$Version/$Zip"
Invoke-WebRequest -OutFile checksums.txt "https://github.com/shanewas/wslx/releases/download/v$Version/checksums.txt"
if ((Get-FileHash $Zip).Hash -ne (Select-String $Zip checksums.txt).Line.Split(' ')[0]) { throw 'checksum mismatch' }
Expand-Archive $Zip "$env:LOCALAPPDATA\wslx" -Force
[Environment]::SetEnvironmentVariable('Path', "$env:LOCALAPPDATA\wslx;" + [Environment]::GetEnvironmentVariable('Path', 'User'), 'User')
```

Open a new terminal afterwards; `wslx version` should print the version on both sides.

Package managers:

```powershell
winget install Shanewas.wslx
scoop bucket add wslx https://github.com/shanewas/scoop-bucket
scoop install wslx
```

From source (Go 1.22+), on either side:

```sh
go install github.com/shanewas/wslx/cmd/wslx@latest
```

## Stability

Pre-1.0. What 1.0 will freeze: the command names and flags listed above,
exit code behaviour, and the `--json` field names (fields may be added,
existing ones won't change meaning). Until then a breaking change gets
a CHANGELOG entry and a minor version bump.

## Layout

- `cmd/wslx/`: CLI entry
- `internal/`: side detection, process launch, tool resolution, PowerShell argv, paths, env, doctor
- `.github/workflows/`: CI (Ubuntu amd64/arm64 + Windows) and the tag release

## Build

Requires Go 1.22+.

```sh
go build ./...
go vet ./...
go test ./...
```

## License

MIT, see [LICENSE](LICENSE).
