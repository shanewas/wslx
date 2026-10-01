package pwshx

import (
	"encoding/base64"
	"encoding/binary"
	"strings"
	"testing"
	"unicode/utf16"
)

func decode(t *testing.T, enc string) string {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		t.Fatalf("base64: %v", err)
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(b[2*i:])
	}
	return string(utf16.Decode(u))
}

func TestScriptArgvEncoded(t *testing.T) {
	argv, stdin := ScriptArgv("pwsh.exe", "Get-ChildItem 'C:\\Users\\太郎'", true)
	if stdin != "" {
		t.Fatalf("stdin should be empty for encoded mode, got %q", stdin)
	}
	want := []string{"pwsh.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand"}
	for i, w := range want {
		if argv[i] != w {
			t.Fatalf("argv[%d]=%q want %q", i, argv[i], w)
		}
	}
	script := decode(t, argv[4])
	if !strings.HasPrefix(script, "try { [Console]::OutputEncoding") {
		t.Fatalf("missing utf8 preamble: %q", script)
	}
	if !strings.HasSuffix(script, "Get-ChildItem 'C:\\Users\\太郎'\n") {
		t.Fatalf("script mangled: %q", script)
	}
}

func TestScriptArgvNoPreambleOnWindowsConsole(t *testing.T) {
	argv, _ := ScriptArgv("pwsh.exe", "Get-Date", false)
	if got := decode(t, argv[4]); got != "Get-Date\n" {
		t.Fatalf("got %q", got)
	}
}

func TestScriptArgvLongFallsBackToStdin(t *testing.T) {
	long := strings.Repeat("Write-Output x\n", 2000)
	argv, stdin := ScriptArgv("pwsh.exe", long, false)
	want := []string{"pwsh.exe", "-NoProfile", "-NonInteractive", "-Command", "-"}
	if strings.Join(argv, " ") != strings.Join(want, " ") {
		t.Fatalf("argv=%q", argv)
	}
	if stdin != long {
		t.Fatal("stdin should carry the script verbatim")
	}
}

func TestInteractiveArgv(t *testing.T) {
	argv := InteractiveArgv("pwsh.exe")
	if strings.Join(argv, " ") != "pwsh.exe -NoLogo" {
		t.Fatalf("got %q", argv)
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
