// Package quote renders argv arrays into safe command lines per side.
package quote

import "strings"

// WindowsCommandLine renders argv as one command line using
// CommandLineToArgvW-inverse escaping: args with whitespace or quotes
// are wrapped in quotes, inner backslashes and quotes escaped.
func WindowsCommandLine(argv []string) string {
	out := make([]string, len(argv))
	for i, a := range argv {
		out[i] = quoteArg(a)
	}
	return strings.Join(out, " ")
}

func quoteArg(s string) string {
	if s == "" {
		return `""`
	}
	need := false
	for i := 0; i < len(s); i++ {
		if c := s[i]; c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '"' {
			need = true
			break
		}
	}
	if !need {
		return s
	}
	var b strings.Builder
	b.WriteByte('"')
	slashes := 0
	flush := func() {
		if slashes > 0 {
			b.WriteString(strings.Repeat(`\`, slashes))
			slashes = 0
		}
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			slashes++
		case '"':
			b.WriteString(strings.Repeat(`\`, slashes*2+1))
			slashes = 0
			b.WriteByte('"')
		default:
			flush()
			b.WriteByte(s[i])
		}
	}
	b.WriteString(strings.Repeat(`\`, slashes*2))
	b.WriteByte('"')
	return b.String()
}

// WslExecArgv builds the wsl.exe exec-form args: -d <distro> when
// distro is non-empty, then -e plus argv verbatim.
func WslExecArgv(distro string, argv []string) []string {
	out := make([]string, 0, len(argv)+3)
	if distro != "" {
		out = append(out, "-d", distro)
	}
	out = append(out, "-e")
	out = append(out, argv...)
	return out
}
