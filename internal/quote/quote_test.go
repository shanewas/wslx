package quote

import (
	"reflect"
	"testing"
)

func TestWindowsCommandLineCorpus(t *testing.T) {
	cases := []struct {
		name string
		argv []string
		want string
	}{
		{"plain", []string{"echo", "hello"}, "echo hello"},
		{"dollar", []string{"echo", "$var"}, "echo $var"},
		{"braces", []string{"echo", "${HOME}"}, "echo ${HOME}"},
		{"subshell", []string{"bash", "-c", "$(ls)"}, "bash -c $(ls)"},
		{"awk", []string{"awk", "{print $1}"}, `awk "{print $1}"`},
		{"spaces", []string{"a b", "c"}, `"a b" c`},
		{"embedded quote", []string{`a"b`}, `"a\"b"`},
		{"backslash plain", []string{`C:\path\to`}, `C:\path\to`},
		{"backslash trailing quoted", []string{`a b\`}, `"a b\\"`},
		{"backslash before quote", []string{`a\"b c`}, `"a\\\"b c"`},
		{"trailing newline", []string{"a\n"}, "\"a\n\""},
		{"empty arg", []string{""}, `""`},
		{"empty among args", []string{"a", "", "b"}, `a "" b`},
		{"tab", []string{"a\tb"}, "\"a\tb\""},
	}
	for _, c := range cases {
		if got := WindowsCommandLine(c.argv); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestWslExecArgv(t *testing.T) {
	got := WslExecArgv("Ubuntu-22.04", []string{"ls", "-la"})
	want := []string{"-d", "Ubuntu-22.04", "-e", "ls", "-la"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
	got = WslExecArgv("", []string{"ls"})
	want = []string{"-e", "ls"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}
