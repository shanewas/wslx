package pwshx

import (
	"reflect"
	"strings"
	"testing"
)

func TestBaseFlags(t *testing.T) {
	want := []string{"-NoProfile", "-NonInteractive"}
	if got := BaseFlags(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolve(t *testing.T) {
	p, err := Resolve()
	if err == nil {
		if p == "" {
			t.Fatal("nil error with empty path")
		}
		t.Logf("resolved %s", p)
		return
	}
	if !strings.Contains(err.Error(), "searched:") {
		t.Fatalf("error names nothing searched: %v", err)
	}
	if !strings.Contains(err.Error(), "pwsh") {
		t.Fatalf("error lacks pwsh: %v", err)
	}
	t.Logf("unresolved (ok on bare linux): %v", err)
}
