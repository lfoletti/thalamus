package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// install sets up the local recall on this machine: it finds the native
// recall of CortexDB, writes TH_LOCAL_RECALL into the env file, and turns
// the native auto-recall off, since the Thalamus hook runs it now.
func install(get func(string) string, stdout io.Writer) error {
	bin := nativeBinary(get)
	if bin == "" {
		return errors.New("no CortexDB binary found: install the CortexDB plugin and start one session with it, or set CORTEXDB_MCP_BIN")
	}
	envFile := get("TH_ENV")
	if envFile == "" {
		envFile = filepath.Join(get("HOME"), ".thalamus", "env")
	}
	data, _ := os.ReadFile(envFile)
	if err := writeFile(envFile, setLine(string(data), "TH_LOCAL_RECALL", bin+" --recall")); err != nil {
		return fmt.Errorf("env file: %w", err)
	}
	sentinel := filepath.Join(cacheRoot(get), "cortexdb", "autorecall")
	if err := writeFile(sentinel, "off\n"); err != nil {
		return fmt.Errorf("CortexDB autorecall: %w", err)
	}
	fmt.Fprintf(stdout, "native recall: %s\nTH_LOCAL_RECALL written to %s\nCortexDB auto-recall turned off: the Thalamus hook runs it now (undo: cortexdb-recall --enable)\n", bin, envFile)
	return nil
}

func cacheRoot(get func(string) string) string {
	if c := get("XDG_CACHE_HOME"); c != "" {
		return c
	}
	return filepath.Join(get("HOME"), ".cache")
}

// nativeBinary is CORTEXDB_MCP_BIN; else the newest cortexdb-mcp of this
// platform in the data folder of the CortexDB plugin or in the CortexDB
// cache; else cortexdb-mcp on the PATH. Only a file that can run counts.
func nativeBinary(get func(string) string) string {
	if b := get("CORTEXDB_MCP_BIN"); runnable(b) {
		return b
	}
	asset := "cortexdb-mcp-" + runtime.GOOS + "-" + runtime.GOARCH
	var found []string
	for _, dir := range []string{filepath.Join(get("HOME"), ".claude", "plugins", "data", "cortexdb*", "bin"), filepath.Join(cacheRoot(get), "cortexdb", "bin")} {
		m, _ := filepath.Glob(filepath.Join(dir, asset+"*"))
		found = append(found, m...)
	}
	best, bestVersion := "", ""
	for _, f := range found {
		v := strings.TrimPrefix(strings.TrimPrefix(filepath.Base(f), asset), "-")
		if runnable(f) && (best == "" || newer(v, bestVersion)) {
			best, bestVersion = f, v
		}
	}
	if best != "" {
		return best
	}
	p, _ := exec.LookPath("cortexdb-mcp")
	return p
}

func runnable(path string) bool {
	info, err := os.Stat(path)
	return path != "" && err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}

// newer reports whether version a is after version b, comparing the dotted
// parts as numbers; a missing or unreadable part is 0.
func newer(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < max(len(pa), len(pb)); i++ {
		x, y := part(pa, i), part(pb, i)
		if x != y {
			return x > y
		}
	}
	return false
}

func part(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, _ := strconv.Atoi(parts[i])
	return n
}

// setLine sets key=value in an env file text: in place when the key is
// there, at the end otherwise.
func setLine(text, key, value string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if text == "" {
		lines = nil
	}
	for i, l := range lines {
		if k, _, _ := strings.Cut(l, "="); strings.TrimSpace(k) == key {
			lines[i] = key + "=" + value
			return strings.Join(lines, "\n") + "\n"
		}
	}
	return strings.Join(append(lines, key+"="+value), "\n") + "\n"
}

func writeFile(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o600)
}
