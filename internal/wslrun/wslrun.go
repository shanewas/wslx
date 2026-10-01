// Package wslrun builds the argv that runs a command on the WSL side.
package wslrun

import (
	"github.com/shanewas/wslx/internal/detect"
	"github.com/shanewas/wslx/internal/winrun"
)

// WslExe is the launcher used to hop distros.
const WslExe = "wsl.exe"

// HopArgv returns the exec-form hop: wslExe -d <distro> (when
// non-empty) -e plus argv verbatim. No shell runs on the far side.
func HopArgv(wslExe, distro string, argv []string) []string {
	out := []string{wslExe}
	if distro != "" {
		out = append(out, "-d", distro)
	}
	out = append(out, "-e")
	return append(out, argv...)
}

// Argv applies the same-side rule: inside WSL with no --distro the
// command runs as is, otherwise it hops through wsl.exe, resolved the
// same way `win` resolves tools so a trimmed PATH still works.
func Argv(distro string, argv []string) ([]string, error) {
	if distro == "" && detect.DetectSide() == detect.SideWSL {
		return argv, nil
	}
	wsl, err := winrun.Resolve(WslExe)
	if err != nil {
		return nil, err
	}
	return HopArgv(wsl, distro, argv), nil
}
