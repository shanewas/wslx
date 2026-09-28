// Package wslrun executes argv on the WSL side.
package wslrun

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/shanewas/wslx/internal/detect"
	"github.com/shanewas/wslx/internal/quote"
)

// WslExe is the default launcher used to hop distros.
const WslExe = "wsl.exe"

// BuildWslArgv returns full argv for the hop: wsl.exe -d <distro>
// (when non-empty) -e plus argv verbatim.
func BuildWslArgv(distro string, argv []string) []string {
	return append([]string{WslExe}, quote.WslExecArgv(distro, argv)...)
}

func runArgs(full []string, stdin string) (stdout, stderr string, exit int, err error) {
	if len(full) == 0 {
		return "", "", -1, fmt.Errorf("empty argv")
	}
	cmd := exec.Command(full[0], full[1:]...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	if err = cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return outBuf.String(), errBuf.String(), ee.ExitCode(), err
		}
		return outBuf.String(), errBuf.String(), -1, err
	}
	return outBuf.String(), errBuf.String(), 0, nil
}

// RunDirect executes argv in the current environment with no hop.
func RunDirect(argv []string, stdin string) (string, string, int, error) {
	return runArgs(argv, stdin)
}

// RunViaWsl hops via wslExe into distro (default distro when empty).
func RunViaWsl(wslExe, distro string, argv []string, stdin string) (string, string, int, error) {
	return runArgs(append([]string{wslExe}, quote.WslExecArgv(distro, argv)...), stdin)
}

// Run applies the same-side rule: inside WSL with no --distro it runs
// directly, otherwise it hops via wsl.exe.
func Run(argv []string, distro, stdin string) (string, string, int, error) {
	if distro == "" && detect.DetectSide() == detect.SideWSL {
		return RunDirect(argv, stdin)
	}
	return RunViaWsl(WslExe, distro, argv, stdin)
}
