package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeProbes() []Probe {
	return []Probe{
		func() Check { return Check{Name: "a green", OK: true} },
		func() Check { return Check{Name: "b red", OK: false, Fix: "run: fix-b"} },
	}
}

func TestRunWithFakes(t *testing.T) {
	checks := RunWith(fakeProbes())
	if len(checks) != 2 || !checks[0].OK || checks[1].OK || checks[1].Fix != "run: fix-b" {
		t.Fatalf("got %+v", checks)
	}
}

func TestLiveRunShape(t *testing.T) {
	checks := Run()
	if len(checks) == 0 {
		t.Fatal("no checks for this side")
	}
	for _, c := range checks {
		if c.Name == "" {
			t.Errorf("check with empty Name: %+v", c)
		}
		if !c.OK && c.Fix == "" {
			t.Errorf("failing check %q has empty Fix", c.Name)
		}
	}
	t.Logf("\n%s", FormatHuman(checks))
}

func TestFormatJSONShape(t *testing.T) {
	var decoded []Check
	if err := json.Unmarshal([]byte(FormatJSON(RunWith(fakeProbes()))), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded) != 2 || decoded[0].Name != "a green" || !decoded[0].OK {
		t.Fatalf("shape: %+v", decoded)
	}
	if decoded[1].OK || decoded[1].Fix != "run: fix-b" {
		t.Fatalf("shape: %+v", decoded)
	}
	if got := FormatJSON(nil); got != "[]\n" {
		t.Fatalf("nil checks: %q", got)
	}
}

func TestFormatHuman(t *testing.T) {
	out := FormatHuman(RunWith(fakeProbes()))
	if !strings.Contains(out, "ok   a green") || !strings.Contains(out, "FAIL b red") {
		t.Fatalf("human:\n%s", out)
	}
	if !strings.Contains(out, "fix: run: fix-b") || !strings.Contains(out, "1/2 checks green") {
		t.Fatalf("human lacks fix or tally:\n%s", out)
	}
}

func TestWslConfFalse(t *testing.T) {
	conf := filepath.Join(t.TempDir(), "wsl.conf")
	body := `
[boot]
systemd = true

[automount]
enabled = false

[Interop]
# enabled = false
appendWindowsPath = FALSE   # trimmed PATH
`
	if err := os.WriteFile(conf, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if wslConfFalse(conf, "interop", "enabled") {
		t.Error("commented-out enabled=false must not count")
	}
	if !wslConfFalse(conf, "interop", "appendWindowsPath") {
		t.Error("appendWindowsPath = FALSE with trailing comment must count")
	}
	if wslConfFalse(conf, "interop", "systemd") {
		t.Error("key from another section leaked")
	}
	if wslConfFalse(filepath.Join(t.TempDir(), "missing"), "interop", "enabled") {
		t.Error("missing file must mean default (true)")
	}
}
