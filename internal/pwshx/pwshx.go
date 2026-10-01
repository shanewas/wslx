// Package pwshx discovers PowerShell and builds its launch argv.
package pwshx

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os/exec"
	"strings"
	"unicode/utf16"

	"github.com/shanewas/wslx/internal/detect"
)

// utf8Preamble switches the session to UTF-8 output. Without it
// PowerShell writes pipes in the console code page (cp932 on Japanese
// Windows) and nothing on the Linux side can read it. The try guards
// hosts with no console attached.
const utf8Preamble = "try { [Console]::OutputEncoding = New-Object System.Text.UTF8Encoding $false } catch {}; " +
	"$OutputEncoding = New-Object System.Text.UTF8Encoding $false\n"

// maxEncoded keeps the argv under the 32767-char Windows command line.
const maxEncoded = 30000

// Resolve finds pwsh.exe, falling back to powershell.exe. It prefers
// pwsh over powershell in both well-known paths and PATH lookup.
func Resolve() (string, error) {
	var searched []string
	for _, c := range detect.PwshCandidates() {
		searched = append(searched, c)
		if detect.Exists(c) {
			return c, nil
		}
	}
	for _, n := range []string{"pwsh", "pwsh.exe", "powershell", "powershell.exe"} {
		searched = append(searched, n)
		if p, err := exec.LookPath(n); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no PowerShell found (searched: %s)", strings.Join(searched, ", "))
}

// ScriptArgv builds the non-interactive launch. The script travels as
// -EncodedCommand (UTF-16LE base64), so neither quoting nor code pages
// can mangle it and stdin stays free for data. A script too long for
// the command line falls back to -Command - and comes back as stdin.
func ScriptArgv(exe, script string, utf8Out bool) (argv []string, stdin string) {
	if utf8Out {
		script = utf8Preamble + script
	}
	if !strings.HasSuffix(script, "\n") {
		script += "\n"
	}
	argv = []string{exe, "-NoProfile", "-NonInteractive"}
	enc := base64.StdEncoding.EncodeToString(utf16le(script))
	if len(enc) <= maxEncoded {
		return append(argv, "-EncodedCommand", enc), ""
	}
	return append(argv, "-Command", "-"), script
}

// InteractiveArgv starts a real session: profile loaded, prompts allowed.
func InteractiveArgv(exe string) []string {
	return []string{exe, "-NoLogo"}
}

func utf16le(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, 2*len(u))
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[2*i:], c)
	}
	return b
}
