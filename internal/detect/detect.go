// Package detect identifies which side of the Windows/WSL boundary
// the binary is running on and probes interop health.
package detect

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Side is the host side the binary runs on.
type Side int

const (
	SideUnknown Side = iota
	SideWindows
	SideWSL
	SideLinux
)

// String returns the short side name.
func (s Side) String() string {
	switch s {
	case SideWindows:
		return "windows"
	case SideWSL:
		return "wsl"
	case SideLinux:
		return "linux"
	default:
		return "unknown"
	}
}

// BinfmtPath is the kernel interop gate. Absent means broken interop.
const BinfmtPath = "/proc/sys/fs/binfmt_misc/WSLInterop"

// DetectSide reports the current side via GOOS plus /proc/version
// (microsoft/WSL marker, case-insensitive) or WSL_DISTRO_NAME.
func DetectSide() Side {
	if runtime.GOOS == "windows" {
		return SideWindows
	}
	if runtime.GOOS != "linux" {
		return SideUnknown
	}
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return SideWSL
	}
	if b, err := os.ReadFile("/proc/version"); err == nil {
		v := strings.ToLower(string(b))
		if strings.Contains(v, "microsoft") || strings.Contains(v, "wsl") {
			return SideWSL
		}
	}
	return SideLinux
}

// Distro returns the current distro name, or "" when unknown.
func Distro() string {
	return os.Getenv("WSL_DISTRO_NAME")
}

// InteropHealth is the WSL->Windows launch gate state.
type InteropHealth struct {
	Binfmt  bool
	Socket  bool
	Wslpath bool
	Detail  string
}

// Healthy reports whether .exe launch should work.
func (h InteropHealth) Healthy() bool {
	return h.Binfmt && h.Socket && h.Wslpath
}

// CheckInterop probes the binfmt entry, the $WSL_INTEROP socket,
// and wslpath availability.
func CheckInterop() InteropHealth {
	h := InteropHealth{}
	h.Binfmt = Exists(BinfmtPath)
	sock := os.Getenv("WSL_INTEROP")
	h.Socket = sock != "" && Exists(sock)
	_, err := exec.LookPath("wslpath")
	h.Wslpath = err == nil
	var missing []string
	if !h.Binfmt {
		missing = append(missing, "binfmt WSLInterop entry")
	}
	if !h.Socket {
		missing = append(missing, "WSL_INTEROP socket")
	}
	if !h.Wslpath {
		missing = append(missing, "wslpath")
	}
	if len(missing) == 0 {
		h.Detail = "interop healthy"
	} else {
		h.Detail = "missing: " + strings.Join(missing, ", ")
	}
	return h
}

// PwshCandidates lists well-known PowerShell locations, pwsh before
// powershell, native Windows paths before /mnt/c mirrors.
func PwshCandidates() []string {
	return []string{
		`C:\Program Files\PowerShell\7\pwsh.exe`,
		`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
		"/mnt/c/Program Files/PowerShell/7/pwsh.exe",
		"/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe",
	}
}

// Exists reports whether path stats without error.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
