package pathx

import (
	"strings"
	"testing"
)

func TestTranslatePure(t *testing.T) {
	cases := []struct {
		p, target, want string
	}{
		{"/mnt/c/Windows", "win", `C:\Windows`},
		{"/mnt/c", "win", `C:\`},
		{"/mnt/c/Windows", "mixed", "C:/Windows"},
		{`C:\Windows`, "unix", "/mnt/c/Windows"},
		{"C:/Windows", "unix", "/mnt/c/Windows"},
		{"C:", "unix", "/mnt/c"},
		{`\\wsl.localhost\Ubuntu-22.04\tmp`, "unix", "/tmp"},
		{"//wsl.localhost/Ubuntu-22.04/tmp", "unix", "/tmp"},
		{"/tmp", "unix", "/tmp"},
		{`C:\Windows`, "win", `C:\Windows`},
	}
	for _, c := range cases {
		got, err := TranslatePure(c.p, c.target)
		if err != nil {
			t.Errorf("TranslatePure(%q,%q) err: %v", c.p, c.target, err)
			continue
		}
		if got != c.want {
			t.Errorf("TranslatePure(%q,%q)=%q want %q", c.p, c.target, got, c.want)
		}
	}
}

func TestTranslatePureErrors(t *testing.T) {
	if _, err := TranslatePure("/tmp", "win"); err == nil {
		t.Error("expected error for /tmp -> win without wslpath")
	}
	if _, err := TranslatePure("relative", "unix"); err == nil {
		t.Error("expected error for relative -> unix")
	}
	if _, err := TranslatePure("/tmp", "bogus"); err == nil {
		t.Error("expected error for bogus target")
	}
	if _, err := Translate("/tmp", "bogus"); err == nil {
		t.Error("expected error for Translate bogus target")
	}
}

func TestComposeWSLHostPath(t *testing.T) {
	got := ComposeWSLHostPath("Ubuntu-22.04", "/tmp")
	want := `\\wsl.localhost\Ubuntu-22.04\tmp`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = ComposeWSLHostPath("Ubuntu-22.04", "/home/me/app")
	if want := `\\wsl.localhost\Ubuntu-22.04\home\me\app`; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if !strings.HasPrefix(got, `\\wsl.localhost\`) {
		t.Fatalf("missing host prefix: %q", got)
	}
}
