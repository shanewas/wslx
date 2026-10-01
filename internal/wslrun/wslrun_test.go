package wslrun

import (
	"reflect"
	"testing"
)

func TestHopArgv(t *testing.T) {
	got := HopArgv("wsl.exe", "Ubuntu-22.04", []string{"ls", "-la"})
	want := []string{"wsl.exe", "-d", "Ubuntu-22.04", "-e", "ls", "-la"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
	got = HopArgv("/mnt/c/Windows/System32/wsl.exe", "", []string{"ls"})
	want = []string{"/mnt/c/Windows/System32/wsl.exe", "-e", "ls"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestHopArgvDoesNotAliasInput(t *testing.T) {
	argv := make([]string, 1, 8)
	argv[0] = "ls"
	HopArgv("wsl.exe", "", argv)
	if argv[0] != "ls" || len(argv) != 1 {
		t.Fatalf("input mutated: %q", argv)
	}
}
