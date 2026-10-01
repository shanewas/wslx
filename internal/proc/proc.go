// Package proc launches child processes two ways: streams inherited, so
// output streams and interactive tools work, or captured for --json.
package proc

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
)

// Inherit runs argv on this process's stdin/stdout/stderr; a non-nil
// stdin replaces os.Stdin. Ctrl-C is left to the child: wslx catches
// the interrupt and keeps waiting, then exits with the child's code.
func Inherit(argv []string, stdin io.Reader) (int, error) {
	if len(argv) == 0 {
		return -1, errors.New("empty argv")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if stdin != nil {
		cmd.Stdin = stdin
	}
	// signal.Notify, not Ignore: SIG_IGN would be inherited by the child.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)
	return exitOf(cmd.Run())
}

// Capture runs argv with both streams buffered; nil stdin means none.
func Capture(argv []string, stdin io.Reader) (stdout, stderr string, exit int, err error) {
	if len(argv) == 0 {
		return "", "", -1, errors.New("empty argv")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, &outBuf, &errBuf
	exit, err = exitOf(cmd.Run())
	return outBuf.String(), errBuf.String(), exit, err
}

// exitOf maps a Run error to the child's exit code; -1 means it never
// started (err says why).
func exitOf(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), err
	}
	return -1, err
}
