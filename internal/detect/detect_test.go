package detect

import (
	"runtime"
	"strings"
	"testing"
)

func TestDetectSidePlausible(t *testing.T) {
	s := DetectSide()
	switch runtime.GOOS {
	case "windows":
		if s != SideWindows {
			t.Fatalf("GOOS=windows but side=%v", s)
		}
	case "linux":
		if s != SideWSL && s != SideLinux {
			t.Fatalf("GOOS=linux but side=%v", s)
		}
	default:
		if s != SideUnknown {
			t.Fatalf("unexpected side=%v", s)
		}
	}
}

func TestDistroMirrorsEnv(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "Ubuntu-22.04")
	if got := Distro(); got != "Ubuntu-22.04" {
		t.Fatalf("Distro()=%q", got)
	}
}

func TestPwshCandidatesOrder(t *testing.T) {
	c := PwshCandidates()
	if len(c) == 0 {
		t.Fatal("empty candidates")
	}
	joined := strings.Join(c, "\n")
	if !strings.Contains(joined, "pwsh.exe") || !strings.Contains(joined, "powershell.exe") {
		t.Fatalf("candidates lack pwsh/powershell:\n%s", joined)
	}
	if !strings.Contains(joined, "/mnt/c/") {
		t.Fatalf("candidates lack /mnt/c mirrors:\n%s", joined)
	}
	pwshIdx := strings.Index(joined, "pwsh.exe")
	psIdx := strings.Index(joined, "powershell.exe")
	if pwshIdx > psIdx {
		t.Fatal("pwsh must precede powershell")
	}
}

func TestExists(t *testing.T) {
	if !Exists(t.TempDir()) {
		t.Fatal("Exists(tempdir)=false")
	}
	if Exists("/no/such/path/38f6b2") {
		t.Fatal("Exists(bogus)=true")
	}
}

func TestHealthyGate(t *testing.T) {
	cases := []struct {
		h    InteropHealth
		want bool
	}{
		{InteropHealth{Binfmt: true, WSL2: false, Socket: false}, true}, // WSL1: no socket exists
		{InteropHealth{Binfmt: true, WSL2: true, Socket: true}, true},
		{InteropHealth{Binfmt: true, WSL2: true, Socket: false}, false},
		{InteropHealth{Binfmt: false, WSL2: false, Socket: false}, false},
	}
	for _, c := range cases {
		if got := c.h.Healthy(); got != c.want {
			t.Errorf("%+v Healthy()=%v want %v", c.h, got, c.want)
		}
	}
}

func TestCheckInteropDetailNonEmpty(t *testing.T) {
	h := CheckInterop()
	if h.Detail == "" {
		t.Fatal("empty Detail")
	}
	if h.Healthy() && strings.HasPrefix(h.Detail, "missing:") {
		t.Fatalf("healthy but detail=%q", h.Detail)
	}
	t.Logf("health=%+v", h)
}
