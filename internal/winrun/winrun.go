// Package winrun executes argv on the Windows side.
package winrun

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

// BuildPowershellArgv returns full argv for stdin-mode PowerShell:
// -Command - reads script from stdin with clean stdout. -File - echoes
// PS <cwd>> prompt lines to stdout on both engines (live battery
// 2026-09-30). Explicit exit N preserved in both modes; native failure
// collapses to 1 in both; -NonInteractive never hangs.
func BuildPowershellArgv(exe string) []string {
	return []string{exe, "-NoProfile", "-NonInteractive", "-Command", "-"}
}

// Run executes argv (argv[0] is the program, normally exe) with stdin
// fed when non-empty. Exit comes from ExitError.ExitCode; a start
// failure returns exit -1 with err set.
func Run(exe string, argv []string, stdin string) (stdout, stderr string, exit int, err error) {
	prog := exe
	var args []string
	if len(argv) > 0 {
		prog = argv[0]
		args = argv[1:]
	}
	cmd := exec.Command(prog, args...)
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
