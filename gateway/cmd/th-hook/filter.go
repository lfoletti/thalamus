package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// filter is a person's choice of brains for the automatic recall: only
// names the brains to keep (empty is every brain), skip the brains to leave
// out. Local brains have the prefix local:. The rights stay the gateway's;
// a filter only chooses among the brains the person may already read.
type filter struct {
	Only []string
	Skip []string
}

const localPrefix = "local:"

func (f filter) on(name string) bool {
	return (len(f.Only) == 0 || slices.Contains(f.Only, name)) && !slices.Contains(f.Skip, name)
}

// parseFilter reads the recall file: one key=value per line, the values
// lists separated by commas. Comments and unknown keys are ignored.
func parseFilter(data string) filter {
	var f filter
	for _, line := range strings.Split(data, "\n") {
		k, v, _ := strings.Cut(line, "=")
		switch strings.TrimSpace(k) {
		case "only":
			f.Only = names(v)
		case "skip":
			f.Skip = names(v)
		}
	}
	return f
}

func names(v string) []string {
	var out []string
	for _, n := range strings.Split(v, ",") {
		if n = strings.TrimSpace(n); n != "" {
			out = append(out, n)
		}
	}
	return out
}

func (f filter) String() string {
	var b strings.Builder
	if len(f.Only) > 0 {
		b.WriteString("only=" + strings.Join(f.Only, ",") + "\n")
	}
	if len(f.Skip) > 0 {
		b.WriteString("skip=" + strings.Join(f.Skip, ",") + "\n")
	}
	return b.String()
}

// with applies one command. on takes a brain out of skip, and adds it to
// only when only is not empty; off adds it to skip; only sets the list,
// and only --clear empties it.
func (f filter) with(cmd string, brains []string) filter {
	switch cmd {
	case "on":
		for _, b := range brains {
			f.Skip = without(f.Skip, b)
			if len(f.Only) > 0 && !slices.Contains(f.Only, b) {
				f.Only = append(f.Only, b)
			}
		}
	case "off":
		for _, b := range brains {
			if !slices.Contains(f.Skip, b) {
				f.Skip = append(f.Skip, b)
			}
		}
	case "only":
		f.Only = brains
		if slices.Equal(brains, []string{"--clear"}) {
			f.Only = nil
		}
	}
	return f
}

func without(list []string, name string) []string {
	var out []string
	for _, n := range list {
		if n != name {
			out = append(out, n)
		}
	}
	return out
}

// shared splits the filter for th_recall: the server brains of only, and
// those of skip. ask is false when only names local brains alone.
func (f filter) shared() (only, skip []string, ask bool) {
	for _, n := range f.Only {
		if !strings.HasPrefix(n, localPrefix) {
			only = append(only, n)
		}
	}
	for _, n := range f.Skip {
		if !strings.HasPrefix(n, localPrefix) {
			skip = append(skip, n)
		}
	}
	return only, skip, len(f.Only) == 0 || len(only) > 0
}

// recallFile is TH_RECALL_FILE, or ~/.thalamus/recall.
func recallFile(get func(string) string) string {
	if p := get("TH_RECALL_FILE"); p != "" {
		return p
	}
	return filepath.Join(get("HOME"), ".thalamus", "recall")
}

func readFilter(path string) filter {
	data, _ := os.ReadFile(path)
	return parseFilter(string(data))
}

func writeFilter(path string, f filter) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("recall file: %w", err)
	}
	if err := os.WriteFile(path, []byte(f.String()), 0o600); err != nil {
		return fmt.Errorf("recall file: %w", err)
	}
	return nil
}
