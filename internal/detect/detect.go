// Package detect identifies which side of the Windows/WSL boundary
// the binary is running on and probes interop health.
package detect

import (
	"os"
	"path/filepath"
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

// BinfmtGlob matches the kernel interop gate. Newer WSL registers it
// again as WSLInterop-late after systemd-binfmt flushes the table, so
// either name means .exe launch works.
const BinfmtGlob = "/proc/sys/fs/binfmt_misc/WSLInterop*"

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
	v := procVersion()
	if strings.Contains(v, "microsoft") || strings.Contains(v, "wsl") {
		return SideWSL
	}
	return SideLinux
}

func procVersion() string {
	b, err := os.ReadFile("/proc/version")
	if err != nil {
		return ""
	}
	return strings.ToLower(string(b))
}

// IsWSL2 reports a WSL2 kernel. WSL1 does interop in-kernel and sets
// no WSL_INTEROP socket, so that probe only applies here.
func IsWSL2() bool {
	return os.Getenv("WSL_INTEROP") != "" || strings.Contains(procVersion(), "wsl2")
}

// Distro returns the current distro name, or "" when unknown.
func Distro() string {
	return os.Getenv("WSL_DISTRO_NAME")
}

// BinfmtPresent reports whether the kernel can launch Windows .exe files.
func BinfmtPresent() bool {
	m, _ := filepath.Glob(BinfmtGlob)
	return len(m) > 0
}

// InteropHealth is the WSL->Windows launch gate state.
type InteropHealth struct {
	Binfmt bool
	Socket bool
	WSL2   bool
	Detail string
}

// Healthy reports whether .exe launch should work: the binfmt entry,
// plus a live interop socket on WSL2.
func (h InteropHealth) Healthy() bool {
	return h.Binfmt && (h.Socket || !h.WSL2)
}

// CheckInterop probes the binfmt entry and the $WSL_INTEROP socket.
func CheckInterop() InteropHealth {
	h := InteropHealth{Binfmt: BinfmtPresent(), WSL2: IsWSL2()}
	sock := os.Getenv("WSL_INTEROP")
	h.Socket = sock != "" && Exists(sock)
	var missing []string
	if !h.Binfmt {
		missing = append(missing, "binfmt WSLInterop entry")
	}
	if h.WSL2 && !h.Socket {
		missing = append(missing, "WSL_INTEROP socket")
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
