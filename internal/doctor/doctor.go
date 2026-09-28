// Package doctor checks cross-boundary health and prints exact fixes.
package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shanewas/wslx/internal/detect"
	"github.com/shanewas/wslx/internal/pwshx"
)

// Check is one health gate: OK plus the Fix command when failing.
type Check struct {
	Name string `json:"name"`
	OK   bool   `json:"ok"`
	Fix  string `json:"fix"`
}

// Prober yields one Check. RunWith runs a probe set so tests can
// inject fakes instead of depending on the live machine.
type Prober interface {
	Probe() Check
}

// ProbeFunc adapts a func to a Prober.
type ProbeFunc func() Check

func (f ProbeFunc) Probe() Check { return f() }

// RunWith runs the given probers in order.
func RunWith(probers []Prober) []Check {
	out := make([]Check, 0, len(probers))
	for _, p := range probers {
		out = append(out, p.Probe())
	}
	return out
}

// DefaultProbers returns the live-machine probe set.
func DefaultProbers() []Prober {
	return []Prober{
		ProbeFunc(checkBinfmt),
		ProbeFunc(checkSocket),
		ProbeFunc(checkWslpath),
		ProbeFunc(checkPwsh),
		ProbeFunc(checkWslExe),
		ProbeFunc(checkDistro),
	}
}

// Run runs the live-machine probe set.
func Run() []Check {
	return RunWith(DefaultProbers())
}

func checkBinfmt() Check {
	return Check{
		Name: "binfmt WSLInterop present",
		OK:   detect.Exists(detect.BinfmtPath),
		Fix:  "From Windows PowerShell run: wsl --shutdown ; then reopen WSL and re-run `wslx doctor`",
	}
}

func checkSocket() Check {
	sock := os.Getenv("WSL_INTEROP")
	return Check{
		Name: "interop socket present",
		OK:   sock != "" && detect.Exists(sock),
		Fix:  "Reopen your WSL shell so WSL_INTEROP is set; if still failing run `wsl --shutdown` from Windows PowerShell",
	}
}

func checkWslpath() Check {
	_, err := exec.LookPath("wslpath")
	return Check{
		Name: "wslpath present",
		OK:   err == nil,
		Fix:  "Ensure /usr/bin/wslpath exists (repair or reinstall your WSL distro)",
	}
}

func checkPwsh() Check {
	_, err := pwshx.Resolve()
	return Check{
		Name: "pwsh present",
		OK:   err == nil,
		Fix:  "Install PowerShell 7 on Windows: winget install Microsoft.PowerShell",
	}
}

func checkWslExe() Check {
	ok := detect.Exists(`/mnt/c/Windows/System32/wsl.exe`)
	if _, err := exec.LookPath("wsl.exe"); err == nil {
		ok = true
	}
	if _, err := exec.LookPath("wsl"); err == nil {
		ok = true
	}
	return Check{
		Name: "wsl.exe present",
		OK:   ok,
		Fix:  "Install WSL from an elevated Windows PowerShell: wsl --install",
	}
}

func checkDistro() Check {
	return Check{
		Name: "distro known",
		OK:   detect.Distro() != "",
		Fix:  "Set WSL_DISTRO_NAME or pass --distro <name> (list distros with: wsl -l -q)",
	}
}

// FormatHuman renders checks as ok/FAIL lines with fix guidance.
func FormatHuman(checks []Check) string {
	var b strings.Builder
	ok := 0
	for _, c := range checks {
		if c.OK {
			ok++
			fmt.Fprintf(&b, "ok   %s\n", c.Name)
		} else {
			fmt.Fprintf(&b, "FAIL %s\n     fix: %s\n", c.Name, c.Fix)
		}
	}
	fmt.Fprintf(&b, "%d/%d checks green\n", ok, len(checks))
	return b.String()
}

// FormatJSON renders checks as indented JSON.
func FormatJSON(checks []Check) string {
	if checks == nil {
		checks = []Check{}
	}
	b, _ := json.MarshalIndent(checks, "", "  ")
	return string(b) + "\n"
}
