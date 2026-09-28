package wslrun

import (
	"reflect"
	"testing"
)

func TestBuildWslArgv(t *testing.T) {
	got := BuildWslArgv("Ubuntu-22.04", []string{"ls", "-la"})
	want := []string{"wsl.exe", "-d", "Ubuntu-22.04", "-e", "ls", "-la"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
	got = BuildWslArgv("", []string{"ls"})
	want = []string{"wsl.exe", "-e", "ls"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRunDirectEcho(t *testing.T) {
	stdout, stderr, exit, err := RunDirect([]string{"echo", "hello-direct"}, "")
	if err != nil {
		t.Fatalf("run: %v stderr=%q", err, stderr)
	}
	if exit != 0 || stdout != "hello-direct\n" {
		t.Fatalf("exit=%d stdout=%q", exit, stdout)
	}
}

func TestRunDirectExitCode(t *testing.T) {
	_, _, exit, err := RunDirect([]string{"sh", "-c", "exit 7"}, "")
	if err == nil {
		t.Fatal("expected error for exit 7")
	}
	if exit != 7 {
		t.Fatalf("exit=%d", exit)
	}
}

func TestRunEmpty(t *testing.T) {
	if _, _, _, err := RunDirect(nil, ""); err == nil {
		t.Fatal("expected error for empty argv")
	}
}
