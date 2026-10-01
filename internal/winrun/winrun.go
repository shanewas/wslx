// Package winrun resolves Windows executables from either side.
package winrun

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shanewas/wslx/internal/detect"
)

// System32 is where the Windows tools live as seen from WSL.
const System32 = "/mnt/c/Windows/System32"

// Resolve finds the executable for name. On Windows PATH plus PATHEXT
// decide. From WSL it tries PATH as given, then with .exe, then
// System32, so `wslx win ipconfig` works even when wsl.conf sets
// appendWindowsPath=false. Names with a path separator pass through.
func Resolve(name string) (string, error) {
	if strings.ContainsAny(name, `/\`) {
		return name, nil
	}
	if detect.DetectSide() == detect.SideWindows {
		return exec.LookPath(name)
	}
	var tried []string
	for _, c := range candidates(name) {
		tried = append(tried, c)
		if strings.HasPrefix(c, "/") {
			if detect.Exists(c) {
				return c, nil
			}
			continue
		}
		if p, err := exec.LookPath(c); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("%q not found on the Windows side (tried %s); wsl.conf may set appendWindowsPath=false", name, strings.Join(tried, ", "))
}

func candidates(name string) []string {
	names := []string{name}
	if !strings.HasSuffix(strings.ToLower(name), ".exe") {
		names = append(names, name+".exe")
	}
	out := append([]string{}, names...)
	for _, n := range names {
		out = append(out, System32+"/"+n)
	}
	return out
}
