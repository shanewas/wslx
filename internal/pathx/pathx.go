// Package pathx translates paths across the Windows/WSL boundary,
// using wslpath when present and a pure fallback otherwise.
package pathx

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/shanewas/wslx/internal/detect"
)

var (
	mntRe   = regexp.MustCompile(`^/mnt/([a-zA-Z])(?:/(.*))?$`)
	driveRe = regexp.MustCompile(`^([a-zA-Z]):[\\/](.*)$`)
	bareRe  = regexp.MustCompile(`^([a-zA-Z]):$`)
	wslRe   = regexp.MustCompile(`^(?:\\\\|//)wsl\.localhost[/\\]([^/\\]+)[/\\](.*)$`)
)

// Translate converts p to target form ("win", "unix" or "mixed"):
// wslpath when on PATH, else TranslatePure, else (on Windows) wslpath
// inside the default distro, which knows the distro's own paths.
func Translate(p, target string) (string, error) {
	flag, err := wslFlag(target)
	if err != nil {
		return "", err
	}
	if wslpath, err := exec.LookPath("wslpath"); err == nil {
		return viaWslpath(wslpath, flag, p)
	}
	out, perr := TranslatePure(p, target)
	if perr == nil {
		return out, nil
	}
	if detect.DetectSide() == detect.SideWindows {
		return viaWslpath("wsl.exe", "-e", "wslpath", flag, p)
	}
	return "", perr
}

func viaWslpath(argv ...string) (string, error) {
	out, err := exec.Command(argv[0], argv[1:]...).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("%s: %s", strings.Join(argv, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("%s: %w", strings.Join(argv, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func wslFlag(target string) (string, error) {
	switch target {
	case "win":
		return "-w", nil
	case "unix":
		return "-u", nil
	case "mixed":
		return "-m", nil
	default:
		return "", fmt.Errorf("target must be win|unix|mixed, got %q", target)
	}
}

// TranslatePure converts /mnt/<l>/... <-> <L>:\... forms plus the
// \\wsl.localhost\<distro>\... form without external tools.
func TranslatePure(p, target string) (string, error) {
	switch target {
	case "win":
		return pureToWin(p, `\`)
	case "mixed":
		return pureToWin(p, "/")
	case "unix":
		return pureToUnix(p)
	default:
		return "", fmt.Errorf("target must be win|unix|mixed, got %q", target)
	}
}

func pureToWin(p, sep string) (string, error) {
	if m := mntRe.FindStringSubmatch(p); m != nil {
		rest := strings.ReplaceAll(m[2], "/", sep)
		if rest == "" {
			return strings.ToUpper(m[1]) + ":" + sep, nil
		}
		return strings.ToUpper(m[1]) + ":" + sep + rest, nil
	}
	if driveRe.MatchString(p) || bareRe.MatchString(p) || wslRe.MatchString(p) {
		return strings.ReplaceAll(p, "/", sep), nil
	}
	return "", fmt.Errorf("cannot translate %q to windows form without wslpath", p)
}

func pureToUnix(p string) (string, error) {
	if m := driveRe.FindStringSubmatch(p); m != nil {
		return "/mnt/" + strings.ToLower(m[1]) + "/" + strings.ReplaceAll(m[2], `\`, "/"), nil
	}
	if m := bareRe.FindStringSubmatch(p); m != nil {
		return "/mnt/" + strings.ToLower(m[1]), nil
	}
	if m := wslRe.FindStringSubmatch(p); m != nil {
		return "/" + strings.ReplaceAll(m[2], `\`, "/"), nil
	}
	if strings.HasPrefix(p, "/") {
		return p, nil
	}
	return "", fmt.Errorf("cannot translate %q to unix form without wslpath", p)
}

// ComposeWSLHostPath builds the Windows-side path into a distro,
// e.g. \\wsl.localhost\Ubuntu-22.04\tmp for (Ubuntu-22.04, /tmp).
func ComposeWSLHostPath(distro, unixPath string) string {
	rest := strings.ReplaceAll(strings.TrimPrefix(unixPath, "/"), "/", `\`)
	return `\\wsl.localhost\` + distro + `\` + rest
}
