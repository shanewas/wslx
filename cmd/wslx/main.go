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
	"strings"
	"unicode/utf8"

	"github.com/shanewas/wslx/internal/detect"
	"github.com/shanewas/wslx/internal/doctor"
	"github.com/shanewas/wslx/internal/envx"
	"github.com/shanewas/wslx/internal/pathx"
	"github.com/shanewas/wslx/internal/pwshx"
	"github.com/shanewas/wslx/internal/winrun"
	"github.com/shanewas/wslx/internal/wslrun"
)

var version = "v0.1.0"

var jsonOut bool

func main() {
	args := os.Args[1:]
	for len(args) > 0 && args[0] == "--json" {
		jsonOut = true
		args = args[1:]
	}
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(topHelp)
		if len(args) == 0 {
			os.Exit(2)
		}
		return
	}
	cmd, rest := args[0], args[1:]
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
	case "version":
		cmdVersion()
	default:
		fatalf("unknown command %q (try `wslx --help`)", cmd)
	}
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

// pipedStdin returns piped stdin content, or "" when stdin is a TTY.
func pipedStdin() string {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice != 0 {
		return ""
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		fatalf("read stdin: %v", err)
	}
	return string(b)
}

func stdinIsTTY() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
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

func emit(stdout, stderr string, exit int) {
	if jsonOut {
		out, outEnc := encodeStream(stdout)
		errS, errEnc := encodeStream(stderr)
		b, _ := json.Marshal(envelope{
			Side: jsonSide(), Distro: detect.Distro(),
			Exit: exit & 0xff, ExitFull: exit,
			Stdout: out, Stderr: errS,
			StdoutEncoding: outEnc, StderrEncoding: errEnc,
		})
		fmt.Println(string(b))
	} else {
		io.WriteString(os.Stdout, stdout)
		io.WriteString(os.Stderr, stderr)
	}
	os.Exit(exit & 0xff)
}

// childExit maps a Run error to its exit code, fatal on start failure.
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
	stdout, stderr, exit, err := winrun.Run(args[0], args, pipedStdin())
	emit(stdout, stderr, childExit(err, exit))
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
	stdout, stderr, exit, err := wslrun.Run(args, distro, pipedStdin())
	emit(stdout, stderr, childExit(err, exit))
}

func cmdPwsh(args []string) {
	args = stripDashDash(args)
	needWindowsSide()
	exe, err := pwshx.Resolve()
	if err != nil {
		fatalf("%v; run `wslx doctor` for fix guidance", err)
	}
	if len(args) == 0 {
		if s := pipedStdin(); s != "" {
			runPwshScript(exe, s)
		}
		if stdinIsTTY() {
			interactivePwsh(exe)
		}
		fatalf("no script given and stdin is empty; pass statements, pipe a script, or run from a TTY")
	}
	runPwshScript(exe, strings.Join(args, "\n")+"\n")
}

func runPwshScript(exe, script string) {
	argv := winrun.BuildPowershellArgv(exe)
	stdout, stderr, exit, err := winrun.Run(exe, argv, script)
	emit(stdout, stderr, childExit(err, exit))
}

func interactivePwsh(exe string) {
	c := exec.Command(exe, pwshx.BaseFlags()...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			os.Exit(ee.ExitCode() & 0xff)
		}
		fatalf("%v", err)
	}
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

func cmdVersion() {
	if jsonOut {
		b, _ := json.Marshal(map[string]string{"version": version})
		fmt.Println(string(b))
		return
	}
	fmt.Println("wslx " + version)
}

const topHelp = `wslx - run the other side natively

usage: wslx [--json] <command> [args...]

commands:
  win [--] <command...>               run Windows command from anywhere
  wsl [--distro D] [--] <command...>  run Linux command from anywhere
  pwsh [--] <script...>               PowerShell via stdin, never hangs
  path (--win|--unix|--mixed) <p>      translate a path across the boundary
  env show|suggest ...                parse or build WSLENV values
  doctor                              health + fix guidance
  version                             print version

global flags: --json  machine-readable envelope
per-command help: wslx <command> --help
`

var cmdHelps = map[string]string{
	"win": `wslx win - run a Windows command from anywhere

usage: wslx win [--] <command...>

Runs directly when already on Windows, else via interop .exe launch.
Fails clean (pointing at wslx doctor) when interop is broken.

examples:
  wslx win Get-ChildItem C:\Users\me
`,
	"wsl": `wslx wsl - run a Linux command from anywhere

usage: wslx wsl [--distro D] [--] <command...>

Same-side rule: inside WSL without --distro runs with no hop.
From WSL, --distro hops via interop and fails clean (pointing at
wslx doctor) when interop is broken.

examples:
  wslx wsl ls -la ~/projects
  wslx wsl grep -R "TODO" .
  wslx wsl --distro Ubuntu-22.04 -- uname -a
`,
	"pwsh": `wslx pwsh - PowerShell specifically, stdin-piped, never hangs

usage: wslx pwsh [--] <statement...>

Each argument is one PowerShell statement, joined and piped to
pwsh -NoProfile -NonInteractive -File - (exit code preserved).
With no args it reads the script from piped stdin; on an
interactive TTY it starts an interactive session.

examples:
  wslx pwsh "Get-Service | Where-Object Status -eq Running"
`,
	"path": `wslx path - translate a path across the boundary

usage: wslx path (--win|--unix|--mixed) <p>

examples:
  wslx path --win ~/projects/app
  # -> \\wsl.localhost\Ubuntu-22.04\home\...\app
`,
	"env": `wslx env - parse or build WSLENV values

usage: wslx env show
       wslx env suggest NAME[/flags]...

examples:
  wslx env show
  wslx env suggest BASH_ENV/u MY_VAR/p
`,
	"doctor": `wslx doctor - health + fix guidance

usage: wslx doctor

Green means everything works. Every failure prints cause plus the
exact command that fixes it. Exit 0 when all green, 1 otherwise.

examples:
  wslx doctor
`,
	"version": `wslx version - print version

usage: wslx version
`,
}
