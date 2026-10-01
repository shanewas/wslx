// Package doctor checks cross-boundary health and prints exact fixes.
// The probe set depends on the side: WSL checks the bridge into
// Windows, Windows checks the bridge into WSL, plain Linux gets one
// clear failure instead of six misleading ones.
package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shanewas/wslx/internal/detect"
	"github.com/shanewas/wslx/internal/pwshx"
	"github.com/shanewas/wslx/internal/winrun"
)

// Check is one health gate: OK plus the Fix command when failing.
type Check struct {
	Name string `json:"name"`
	OK   bool   `json:"ok"`
	Fix  string `json:"fix"`
}

// Probe yields one Check.
type Probe func() Check

// WslConfPath is where WSL reads per-distro settings.
var WslConfPath = "/etc/wsl.conf"

// RunWith runs the given probes in order.
func RunWith(probes []Probe) []Check {
	out := make([]Check, 0, len(probes))
	for _, p := range probes {
		out = append(out, p())
	}
	return out
}

// Probes returns the live-machine probe set for the current side.
func Probes() []Probe {
	switch detect.DetectSide() {
	case detect.SideWindows:
		return []Probe{checkWslExe, checkDefaultDistro, checkPwsh}
	case detect.SideWSL:
		p := []Probe{checkInteropConf, checkBinfmt}
		if detect.IsWSL2() {
			p = append(p, checkSocket)
		}
		return append(p, checkWindowsPathConf, checkWslpath, checkWslExe, checkPwsh, checkDistro)
	default:
		return []Probe{checkSide}
	}
}

// Run runs the live-machine probe set.
func Run() []Check {
	return RunWith(Probes())
}

func checkSide() Check {
	return Check{
		Name: "running inside WSL or on Windows",
		OK:   false,
		Fix:  "wslx bridges Windows and WSL; this is plain " + runtime.GOOS + ". Install it inside a WSL distro or on Windows",
	}
}

func checkInteropConf() Check {
	return Check{
		Name: "wsl.conf keeps interop on",
		OK:   !wslConfFalse(WslConfPath, "interop", "enabled"),
		Fix:  "In /etc/wsl.conf set [interop] enabled=true (or drop the line), then from Windows run `wsl --shutdown` and reopen WSL",
	}
}

func checkBinfmt() Check {
	return Check{
		Name: "binfmt WSLInterop present",
		OK:   detect.BinfmtPresent(),
		Fix:  "Run `wsl --update` from Windows (systemd-binfmt clears the entry on older WSL), then `wsl --shutdown`, reopen WSL and re-run `wslx doctor`",
	}
}

func checkSocket() Check {
	sock := os.Getenv("WSL_INTEROP")
	return Check{
		Name: "interop socket live",
		OK:   sock != "" && detect.Exists(sock),
		Fix:  "Reopen your WSL shell so WSL_INTEROP points at a live socket. Inside tmux/screen: export WSL_INTEROP=$(ls -t /run/WSL/*_interop | head -1). Still failing: `wsl --shutdown` from Windows",
	}
}

func checkWindowsPathConf() Check {
	return Check{
		Name: "wsl.conf appends the Windows PATH",
		OK:   !wslConfFalse(WslConfPath, "interop", "appendWindowsPath"),
		Fix:  "Only tools under C:\\Windows\\System32 resolve by name. To run any Windows tool, set [interop] appendWindowsPath=true in /etc/wsl.conf, then `wsl --shutdown` from Windows",
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
	_, err := winrun.Resolve("wsl.exe")
	return Check{
		Name: "wsl.exe present",
		OK:   err == nil,
		Fix:  "Install WSL from an elevated Windows PowerShell: wsl --install",
	}
}

func checkDefaultDistro() Check {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, "wsl.exe", "-e", "true").Run()
	return Check{
		Name: "default distro boots",
		OK:   err == nil,
		Fix:  "Run `wsl -e true` to see the error. No distro: wsl --install -d Ubuntu. Wrong default: wsl --set-default <name> (list: wsl -l -v)",
	}
}

func checkDistro() Check {
	return Check{
		Name: "distro known",
		OK:   detect.Distro() != "",
		Fix:  "Set WSL_DISTRO_NAME or pass --distro <name> (list distros with: wsl -l -q)",
	}
}

// wslConfFalse reports whether the INI file at path sets key=false
// under [section]. A missing file or key means the default, true.
func wslConfFalse(path, section, key string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	in := false
	for _, line := range strings.Split(string(b), "\n") {
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			in = strings.EqualFold(line, "["+section+"]")
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if in && ok && strings.EqualFold(strings.TrimSpace(k), key) {
			return strings.EqualFold(strings.TrimSpace(v), "false")
		}
	}
	return false
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
