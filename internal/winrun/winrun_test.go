package winrun

import (
	"reflect"
	"strings"
	"testing"

	"github.com/shanewas/wslx/internal/detect"
	"github.com/shanewas/wslx/internal/pwshx"
)

func TestBuildPowershellArgv(t *testing.T) {
	got := BuildPowershellArgv("pwsh.exe")
	want := []string{"pwsh.exe", "-NoProfile", "-NonInteractive", "-Command", "-"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRunStartFailure(t *testing.T) {
	_, _, exit, err := Run("/no/such/bin/38f6b2", []string{"/no/such/bin/38f6b2"}, "")
	if err == nil || exit != -1 {
		t.Fatalf("exit=%d err=%v", exit, err)
	}
}

func TestRunLivePwsh(t *testing.T) {
	if !detect.CheckInterop().Healthy() {
		t.Skip("interop broken, no live .exe")
	}
	exe, err := pwshx.Resolve()
	if err != nil {
		t.Skipf("no pwsh: %v", err)
	}
	stdout, _, exit, err := Run(exe, BuildPowershellArgv(exe), "Write-Output hello-live\n")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if exit != 0 || !strings.Contains(stdout, "hello-live") {
		t.Fatalf("exit=%d stdout=%q", exit, stdout)
	}
}
