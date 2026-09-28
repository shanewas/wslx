package doctor

import (
	"encoding/json"
	"strings"
	"testing"
)

func fakeProbes() []Prober {
	return []Prober{
		ProbeFunc(func() Check { return Check{Name: "a green", OK: true} }),
		ProbeFunc(func() Check { return Check{Name: "b red", OK: false, Fix: "run: fix-b"} }),
	}
}

func TestRunWithFakes(t *testing.T) {
	checks := RunWith(fakeProbes())
	if len(checks) != 2 || !checks[0].OK || checks[1].OK {
		t.Fatalf("got %+v", checks)
	}
}

func TestFixNonEmptyOnFailure(t *testing.T) {
	for _, c := range Run() {
		if !c.OK && c.Fix == "" {
			t.Errorf("failing check %q has empty Fix", c.Name)
		}
	}
	// Fake red check keeps its fix through RunWith.
	red := RunWith(fakeProbes())[1]
	if red.Fix == "" {
		t.Error("fake failing check lost Fix")
	}
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
}

func TestFormatHuman(t *testing.T) {
	out := FormatHuman(RunWith(fakeProbes()))
	if !strings.Contains(out, "ok   a green") || !strings.Contains(out, "FAIL b red") {
		t.Fatalf("human:\n%s", out)
	}
	if !strings.Contains(out, "fix: run: fix-b") {
		t.Fatalf("human lacks fix:\n%s", out)
	}
}

func TestLiveRunHasSixChecks(t *testing.T) {
	if got := len(Run()); got != 6 {
		t.Fatalf("got %d checks", got)
	}
}
