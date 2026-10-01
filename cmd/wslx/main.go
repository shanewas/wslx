// Command wslx is a bidirectional Windows<->WSL interop helper.
package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
	"unicode/utf8"

	"github.com/shanewas/wslx/internal/detect"
	"github.com/shanewas/wslx/internal/doctor"
	"github.com/shanewas/wslx/internal/envx"
	"github.com/shanewas/wslx/internal/pathx"
	"github.com/shanewas/wslx/internal/proc"
	"github.com/shanewas/wslx/internal/pwshx"
	"github.com/shanewas/wslx/internal/winrun"
	"github.com/shanewas/wslx/internal/wslrun"
)

var version = "dev"

var jsonOut bool

func main() {
	args := os.Args[1:]
	for len(args) > 0 && args[0] == "--json" {
		jsonOut = true
		args = args[1:]
	}
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(topHelp)
		return
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "path", "env", "doctor", "version":
		rest = takeJSON(rest)
	}
	if restHasHelp(rest) {
		printCmdHelp(cmd)
		return
	}
	switch cmd {
	case "win":
		cmdWin(rest)
	case "wsl":
		cmdWsl(rest)
	case "pwsh":
		cmdPwsh(rest)
	case "path":
		cmdPath(rest)
	case "env":
		cmdEnv(rest)
	case "doctor":
		cmdDoctor(rest)
	case "version", "--version", "-v":
		cmdVersion()
	default:
		fatalf("unknown command %q (try `wslx --help`)", cmd)
	}
}

// takeJSON accepts --json after the command for commands that never
// forward arguments to a child.
func takeJSON(args []string) []string {
	out := args[:0:0]
	for _, a := range args {
		if a == "--json" {
			jsonOut = true
			continue
		}
		out = append(out, a)
	}
	return out
}

func restHasHelp(rest []string) bool {
	return len(rest) > 0 && (rest[0] == "-h" || rest[0] == "--help")
}

func printCmdHelp(cmd string) {
	if h, ok := cmdHelps[cmd]; ok {
		fmt.Print(h)
		return
	}
	fatalf("unknown command %q (try `wslx --help`)", cmd)
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "wslx: "+format+"\n", a...)
	os.Exit(1)
}

func usagef(cmd, format string, a ...any) {
	fmt.Fprintf(os.Stderr, "wslx: "+format+"\n", a...)
	if h, ok := cmdHelps[cmd]; ok {
		fmt.Fprint(os.Stderr, h)
	}
	os.Exit(2)
}

func stripDashDash(args []string) []string {
	if len(args) > 0 && args[0] == "--" {
		return args[1:]
	}
	return args
}

func stdinIsTTY() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// stdinIfPiped returns os.Stdin when data is piped in, nil on a TTY so
// a scripted (--json) child never blocks on the terminal.
func stdinIfPiped() io.Reader {
	if stdinIsTTY() {
		return nil
	}
	return os.Stdin
}

// exitCode masks to the 0-255 a Unix shell can see; Windows keeps the
// full 32-bit value.
func exitCode(n int) int {
	if runtime.GOOS != "windows" {
		return n & 0xff
	}
	return n
}

// run launches argv: streams inherited normally, so output streams and
// interactive tools work; captured into the JSON envelope with --json.
// A non-nil stdin replaces the terminal or pipe.
func run(argv []string, stdin io.Reader) {
	if !jsonOut {
		exit, err := proc.Inherit(argv, stdin)
		os.Exit(exitCode(childExit(err, exit)))
	}
	if stdin == nil {
		stdin = stdinIfPiped()
	}
	stdout, stderr, exit, err := proc.Capture(argv, stdin)
	emitJSON(stdout, stderr, childExit(err, exit))
}

// envelope is the --json output shape.
type envelope struct {
	Side           string `json:"side"`
	Distro         string `json:"distro,omitempty"`
	Exit           int    `json:"exit"`
	ExitFull       int    `json:"exitFull"`
	Stdout         string `json:"stdout"`
	Stderr         string `json:"stderr"`
	StdoutEncoding string `json:"stdoutEncoding"`
	StderrEncoding string `json:"stderrEncoding"`
	WinPath        string `json:"winPath,omitempty"`
	UnixPath       string `json:"unixPath,omitempty"`
}

func encodeStream(s string) (string, string) {
	if utf8.ValidString(s) {
		return s, "utf8"
	}
	return base64.StdEncoding.EncodeToString([]byte(s)), "base64"
}

func jsonSide() string {
	switch detect.DetectSide() {
	case detect.SideWindows:
		return "win"
	case detect.SideWSL:
		return "wsl"
	default:
		return detect.DetectSide().String()
	}
}

func emitJSON(stdout, stderr string, exit int) {
	out, outEnc := encodeStream(stdout)
	errS, errEnc := encodeStream(stderr)
	b, _ := json.Marshal(envelope{
		Side: jsonSide(), Distro: detect.Distro(),
		Exit: exit & 0xff, ExitFull: exit,
		Stdout: out, Stderr: errS,
		StdoutEncoding: outEnc, StderrEncoding: errEnc,
	})
	fmt.Println(string(b))
	os.Exit(exitCode(exit))
}

// childExit maps a run error to its exit code, fatal on start failure.
func childExit(err error, exit int) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return exit
	}
	fatalf("%v", err)
	return 1
}

// needWindowsSide fails clean when .exe launch cannot work.
func needWindowsSide() {
	if detect.DetectSide() == detect.SideWindows {
		return
	}
	if h := detect.CheckInterop(); !h.Healthy() {
		fatalf("cannot reach Windows side (%s); run `wslx doctor` for fix guidance", h.Detail)
	}
}

func cmdWin(args []string) {
	args = stripDashDash(args)
	if len(args) == 0 {
		usagef("win", "need a command")
	}
	needWindowsSide()
	exe, err := winrun.Resolve(args[0])
	if err != nil {
		fatalf("%v; run `wslx doctor` for fix guidance", err)
	}
	run(append([]string{exe}, args[1:]...), nil)
}

func cmdWsl(args []string) {
	var distro string
	for len(args) > 0 && strings.HasPrefix(args[0], "--distro") {
		a := args[0]
		args = args[1:]
		if v, ok := strings.CutPrefix(a, "--distro="); ok {
			distro = v
		} else if a == "--distro" && len(args) > 0 {
			distro = args[0]
			args = args[1:]
		} else {
			usagef("wsl", "--distro needs a value")
		}
	}
	args = stripDashDash(args)
	if len(args) == 0 {
		usagef("wsl", "need a command")
	}
	if distro != "" && detect.DetectSide() == detect.SideWSL {
		needWindowsSide()
	}
	argv, err := wslrun.Argv(distro, args)
	if err != nil {
		fatalf("%v; run `wslx doctor` for fix guidance", err)
	}
	run(argv, nil)
}

func cmdPwsh(args []string) {
	args = stripDashDash(args)
	needWindowsSide()
	exe, err := pwshx.Resolve()
	if err != nil {
		fatalf("%v; run `wslx doctor` for fix guidance", err)
	}
	script := strings.Join(args, " ")
	if script == "" {
		if stdinIsTTY() {
			if jsonOut {
				fatalf("no script given; pass statements or pipe a script")
			}
			run(pwshx.InteractiveArgv(exe), nil)
		}
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fatalf("read stdin: %v", err)
		}
		script = string(b)
		if strings.TrimSpace(script) == "" {
			fatalf("no script given and stdin is empty; pass statements, pipe a script, or run from a TTY")
		}
	}
	// A Windows console already shows Unicode; only pipes and the WSL
	// side need PowerShell forced to UTF-8.
	utf8Out := jsonOut || detect.DetectSide() != detect.SideWindows
	argv, stdin := pwshx.ScriptArgv(exe, script, utf8Out)
	if stdin != "" {
		run(argv, strings.NewReader(stdin))
	}
	run(argv, nil)
}

func cmdPath(args []string) {
	var target string
	var rest []string
	for _, a := range args {
		switch a {
		case "--win":
			target = "win"
		case "--unix":
			target = "unix"
		case "--mixed":
			target = "mixed"
		default:
			rest = append(rest, a)
		}
	}
	if target == "" || len(rest) != 1 {
		usagef("path", "need exactly --win|--unix|--mixed plus one path")
	}
	got, err := pathx.Translate(rest[0], target)
	if err != nil {
		fatalf("%v", err)
	}
	if jsonOut {
		out, outEnc := encodeStream(got + "\n")
		e := envelope{
			Side: jsonSide(), Distro: detect.Distro(),
			Stdout: out, StdoutEncoding: outEnc, StderrEncoding: "utf8",
		}
		if target == "unix" {
			e.UnixPath = got
		} else {
			e.WinPath = got
		}
		b, _ := json.Marshal(e)
		fmt.Println(string(b))
		return
	}
	fmt.Println(got)
}

var flagGloss = map[rune]string{
	'p': "translate path separators",
	'l': "value is a path list",
	'u': "share WSL->Windows (unix form)",
	'w': "share Windows->WSL (win form)",
}

func cmdEnv(args []string) {
	if len(args) == 0 {
		usagef("env", "need show|suggest")
	}
	switch args[0] {
	case "show":
		entries := envx.ParseWSLENV(os.Getenv("WSLENV"))
		if jsonOut {
			b, _ := json.Marshal(entries)
			fmt.Println(string(b))
			return
		}
		if len(entries) == 0 {
			fmt.Println("WSLENV is empty or unset")
			return
		}
		for _, e := range entries {
			if e.Flags == "" {
				fmt.Printf("%s (shared as-is)\n", e.Name)
				continue
			}
			var glosses []string
			for _, f := range e.Flags {
				if g, ok := flagGloss[f]; ok {
					glosses = append(glosses, g)
				} else {
					glosses = append(glosses, "unknown flag /"+string(f))
				}
			}
			fmt.Printf("%s /%s (%s)\n", e.Name, e.Flags, strings.Join(glosses, "; "))
		}
	case "suggest":
		if len(args) < 2 {
			usagef("env", "suggest needs NAME[/flags]...")
		}
		var entries []envx.Entry
		for _, a := range args[1:] {
			entries = append(entries, envx.ParseWSLENV(a)...)
		}
		fmt.Println(envx.Suggest(entries))
	default:
		usagef("env", "unknown env verb %q", args[0])
	}
}

func cmdDoctor(args []string) {
	if len(args) != 0 {
		usagef("doctor", "doctor takes no arguments")
	}
	checks := doctor.Run()
	if jsonOut {
		fmt.Print(doctor.FormatJSON(checks))
	} else {
		fmt.Print(doctor.FormatHuman(checks))
	}
	for _, c := range checks {
		if !c.OK {
			os.Exit(1)
		}
	}
}

// versionString prefers the release ldflag, then the module version a
// `go install ...@vX.Y.Z` build records, then "dev".
func versionString() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}

func cmdVersion() {
	v := versionString()
	if jsonOut {
		b, _ := json.Marshal(map[string]string{"version": v})
		fmt.Println(string(b))
		return
	}
	fmt.Println("wslx " + v)
}

const topHelp = `wslx - run the other side natively

usage: wslx [--json] <command> [args...]

commands:
  win [--] <exe> [args...]            run a Windows executable from anywhere
  wsl [--distro D] [--] <command...>  run a Linux command from anywhere
  pwsh [--] <statements...>           PowerShell script, never hangs
  path (--win|--unix|--mixed) <p>      translate a path across the boundary
  env show|suggest ...                parse or build WSLENV values
  doctor                              health + fix guidance for this side
  version                             print version (also --version, -v)

global flags: --json  machine-readable envelope (before the command, or
              after it for path/env/doctor/version)
per-command help: wslx <command> --help

Two rules: 'win' runs executables, not cmdlets (use pwsh for those), and
'win'/'wsl' pass arguments through with no shell, so ~, globs, pipes and
$VARS only expand inside 'bash -lc' or 'cmd /c'.
`

var cmdHelps = map[string]string{
	"win": `wslx win - run a Windows executable from anywhere

usage: wslx win [--] <exe> [args...]

Runs an executable (.exe, .cmd, .bat), not a PowerShell cmdlet; give
cmdlets to wslx pwsh. On Windows PATH decides. From WSL the name is
tried as given, with .exe, then in C:\Windows\System32, so it works
when wsl.conf sets appendWindowsPath=false. Output streams and the
child gets your terminal; with --json both streams are captured.
Fails clean (pointing at wslx doctor) when interop is broken.

examples:
  wslx win ipconfig /all
  wslx win cmd /c dir 'C:\Users'
  wslx win code .
`,
	"wsl": `wslx wsl - run a Linux command from anywhere

usage: wslx wsl [--distro D] [--] <command...>

Same-side rule: inside WSL without --distro runs with no hop. From
Windows, or with --distro, it hops through wsl.exe -e. Exec form: no
shell runs on the far side, so ~, globs, pipes and $VARS need one.
Fails clean (pointing at wslx doctor) when interop is broken.

examples:
  wslx wsl uname -a
  wslx wsl grep -R "TODO" .
  wslx wsl -- bash -lc 'ls ~/projects | wc -l'
  wslx wsl --distro Ubuntu-22.04 -- cat /etc/os-release
`,
	"pwsh": `wslx pwsh - run PowerShell, never hangs

usage: wslx pwsh [--] <statements...>
       <script> | wslx pwsh
       wslx pwsh            (on a terminal: interactive session)

Arguments join with spaces into one script; separate statements with ;
or newlines. The script travels as -EncodedCommand with -NoProfile
-NonInteractive, so quoting and console code pages cannot mangle it,
and PowerShell is switched to UTF-8 output on the way into WSL. Piped
stdin is the script when no statements are given, otherwise it is data
for $input. Exit codes come back as PowerShell set them.

examples:
  wslx pwsh "Get-Service | Where-Object Status -eq Running"
  wslx pwsh 'Get-ChildItem C:\Users\me; exit 3'
  wslx pwsh < setup.ps1
`,
	"path": `wslx path - translate a path across the boundary

usage: wslx path (--win|--unix|--mixed) <p>

Uses wslpath inside WSL. On Windows, drive and \\wsl.localhost paths
translate in-process; paths inside the distro go through wsl.exe.

examples:
  wslx path --win ~/projects/app
  # -> \\wsl.localhost\Ubuntu-22.04\home\...\app
  wslx path --unix 'C:\Users\me\src'
  # -> /mnt/c/Users/me/src
`,
	"env": `wslx env - parse or build WSLENV values

usage: wslx env show
       wslx env suggest NAME[/flags]...

examples:
  wslx env show
  wslx env suggest BASH_ENV/u MY_VAR/p
`,
	"doctor": `wslx doctor - health + fix guidance

usage: wslx doctor [--json]

Probes depend on the side. Inside WSL: wsl.conf interop settings, the
binfmt entry, the interop socket (WSL2), wslpath, wsl.exe, PowerShell,
distro name. On Windows: wsl.exe, that the default distro boots,
PowerShell. Every failure prints cause plus the exact command that
fixes it. Exit 0 when all green, 1 otherwise.

examples:
  wslx doctor
  wslx doctor --json
`,
	"version": `wslx version - print version

usage: wslx version [--json]
       wslx --version
`,
}
