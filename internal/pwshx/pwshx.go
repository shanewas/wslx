// Package pwshx discovers PowerShell and owns its safe flags.
package pwshx

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shanewas/wslx/internal/detect"
)

// BaseFlags returns the always-on safe flags: no profile, no prompts.
func BaseFlags() []string {
	return []string{"-NoProfile", "-NonInteractive"}
}

// Resolve finds pwsh.exe, falling back to powershell.exe. It prefers
// pwsh over powershell in both well-known paths and PATH lookup.
func Resolve() (string, error) {
	var searched []string
	for _, c := range detect.PwshCandidates() {
		searched = append(searched, c)
		if detect.Exists(c) {
			return c, nil
		}
	}
	for _, n := range []string{"pwsh", "pwsh.exe", "powershell", "powershell.exe"} {
		searched = append(searched, n)
		if p, err := exec.LookPath(n); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no PowerShell found (searched: %s)", strings.Join(searched, ", "))
}
