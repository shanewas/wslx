package proc

import (
	"runtime"
	"strings"
	"testing"
)

func shell(script string) []string {
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c", script}
	}
	return []string{"sh", "-c", script}
}

func TestCaptureEcho(t *testing.T) {
	stdout, stderr, exit, err := Capture(shell("echo hello"), nil)
	if err != nil || exit != 0 {
		t.Fatalf("exit=%d err=%v stderr=%q", exit, err, stderr)
	}
	if strings.TrimSpace(stdout) != "hello" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestCaptureExitCode(t *testing.T) {
	_, _, exit, err := Capture(shell("exit 7"), nil)
	if err == nil || exit != 7 {
		t.Fatalf("exit=%d err=%v", exit, err)
	}
}

func TestCaptureStdin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no cat on windows")
	}
	stdout, _, exit, err := Capture([]string{"cat"}, strings.NewReader("piped"))
	if err != nil || exit != 0 || stdout != "piped" {
		t.Fatalf("stdout=%q exit=%d err=%v", stdout, exit, err)
	}
}

func TestStartFailure(t *testing.T) {
	_, _, exit, err := Capture([]string{"/no/such/bin/38f6b2"}, nil)
	if err == nil || exit != -1 {
		t.Fatalf("exit=%d err=%v", exit, err)
	}
	if _, _, _, err := Capture(nil, nil); err == nil {
		t.Fatal("expected error for empty argv")
	}
	if _, err := Inherit(nil, nil); err == nil {
		t.Fatal("expected error for empty argv")
	}
}

func TestInheritExitCode(t *testing.T) {
	exit, err := Inherit(shell("exit 3"), nil)
	if err == nil || exit != 3 {
		t.Fatalf("exit=%d err=%v", exit, err)
	}
}
