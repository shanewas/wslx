// Package envx parses and builds WSLENV values.
package envx

import "strings"

// Entry is one WSLENV element: NAME plus optional /flags.
type Entry struct {
	Name  string
	Flags string
}

// ParseWSLENV splits s on ':' into entries; flags follow the first '/'.
func ParseWSLENV(s string) []Entry {
	var out []Entry
	for _, part := range strings.Split(s, ":") {
		if part == "" {
			continue
		}
		name, flags, _ := strings.Cut(part, "/")
		if name == "" {
			continue
		}
		out = append(out, Entry{Name: name, Flags: flags})
	}
	return out
}

// Suggest joins entries into a WSLENV value with ':' separators.
func Suggest(entries []Entry) string {
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Flags == "" {
			parts = append(parts, e.Name)
		} else {
			parts = append(parts, e.Name+"/"+e.Flags)
		}
	}
	return strings.Join(parts, ":")
}
