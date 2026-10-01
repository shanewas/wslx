package winrun

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolvePassesPaths(t *testing.T) {
	for _, p := range []string{"/mnt/c/Windows/System32/cmd.exe", `C:\Windows\notepad.exe`, "./tool"} {
		if got, err := Resolve(p); err != nil || got != p {
			t.Errorf("Resolve(%q)=%q,%v", p, got, err)
		}
	}
}

func TestResolveFromWSLSide(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exercises the Linux-side lookup")
	}
	dir := t.TempDir()
	exe := filepath.Join(dir, "foo.exe")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	for _, name := range []string{"foo", "foo.exe"} {
		if got, err := Resolve(name); err != nil || got != exe {
			t.Errorf("Resolve(%q)=%q,%v want %q", name, got, err, exe)
		}
	}
	_, err := Resolve("nope")
	if err == nil {
		t.Fatal("expected error for missing tool")
	}
	if !strings.Contains(err.Error(), System32+"/nope.exe") || !strings.Contains(err.Error(), "appendWindowsPath") {
		t.Fatalf("error lacks tried list or hint: %v", err)
	}
}

func TestResolveOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows PATHEXT lookup")
	}
	got, err := Resolve("cmd")
	if err != nil || !strings.HasSuffix(strings.ToLower(got), "cmd.exe") {
		t.Fatalf("Resolve(cmd)=%q,%v", got, err)
	}
}
