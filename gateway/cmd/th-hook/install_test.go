package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// executable writes a file at path with mode.
func executable(t *testing.T, path string, mode os.FileMode) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), mode); err != nil {
		t.Fatal(err)
	}
	return path
}

var asset = "cortexdb-mcp-" + runtime.GOOS + "-" + runtime.GOARCH

// install finds the newest native binary among the data folder of the
// CortexDB plugin and the CortexDB cache (a file that cannot run is not a
// candidate), writes TH_LOCAL_RECALL into the env file and keeps its other
// lines, and turns the native auto-recall off.
func TestInstallSetsTheLocalRecall(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	home := t.TempDir()
	newest := executable(t, filepath.Join(home, ".claude/plugins/data/cortexdb-cortexdb/bin", asset+"-2.132.0"), 0o755)
	executable(t, filepath.Join(home, ".cache/cortexdb/bin", asset+"-2.99.0"), 0o755)
	executable(t, filepath.Join(home, ".cache/cortexdb/bin", asset), 0o755)
	executable(t, filepath.Join(home, ".cache/cortexdb/bin", asset+"-3.0.0"), 0o644)
	executable(t, filepath.Join(home, ".thalamus/env"), 0o600)
	if err := os.WriteFile(filepath.Join(home, ".thalamus/env"), []byte("TH_GATEWAY=http://gw\nTH_LOCAL_RECALL=/old --recall\nTH_KEY=k\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := hook(t, map[string]string{"HOME": home}, "", "install")
	if code != 0 || errOut != "" {
		t.Fatalf("code %d, stderr %s", code, errOut)
	}
	env, _ := os.ReadFile(filepath.Join(home, ".thalamus/env"))
	if string(env) != "TH_GATEWAY=http://gw\nTH_LOCAL_RECALL="+newest+" --recall\nTH_KEY=k\n" {
		t.Fatalf("env file %q", env)
	}
	if off, _ := os.ReadFile(filepath.Join(home, ".cache/cortexdb/autorecall")); string(off) != "off\n" {
		t.Fatalf("autorecall %q", off)
	}
	for _, want := range []string{"native recall: " + newest, "TH_LOCAL_RECALL written to " + filepath.Join(home, ".thalamus/env"), "cortexdb-recall --enable"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout lacks %q:\n%s", want, out)
		}
	}
}

// CORTEXDB_MCP_BIN wins; then the folders; then cortexdb-mcp on the PATH.
// XDG_CACHE_HOME moves the cache. A new env file is made 0600, with its
// folder; TH_ENV names another one.
func TestInstallLooksInOrder(t *testing.T) {
	home, cache, path := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("PATH", path)
	onPath := executable(t, filepath.Join(path, "cortexdb-mcp"), 0o755)
	inCache := executable(t, filepath.Join(cache, "cortexdb/bin", asset+"-2.1.0"), 0o755)
	named := executable(t, filepath.Join(t.TempDir(), "mine"), 0o755)
	envFile := filepath.Join(t.TempDir(), "sub", "env")
	for _, c := range []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"HOME": home, "TH_ENV": envFile, "XDG_CACHE_HOME": cache, "CORTEXDB_MCP_BIN": named}, named},
		{map[string]string{"HOME": home, "TH_ENV": envFile, "XDG_CACHE_HOME": cache}, inCache},
		{map[string]string{"HOME": home, "TH_ENV": envFile}, onPath},
	} {
		if code, _, errOut := hook(t, c.env, "", "install"); code != 0 {
			t.Fatalf("%v: %d %s", c.env, code, errOut)
		}
		if env, _ := os.ReadFile(envFile); string(env) != "TH_LOCAL_RECALL="+c.want+" --recall\n" {
			t.Fatalf("%v: %q", c.env, env)
		}
	}
	if info, _ := os.Stat(envFile); info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode())
	}
	if off, _ := os.ReadFile(filepath.Join(cache, "cortexdb/autorecall")); string(off) != "off\n" {
		t.Fatalf("autorecall in XDG_CACHE_HOME: %q", off)
	}
}

// No binary is exit 1 with what to do; a file that cannot be written is
// exit 1; arguments are a usage error.
func TestInstallReportsItsProblems(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	home := t.TempDir()
	if code, _, errOut := hook(t, map[string]string{"HOME": home}, "", "install"); code != 1 || !strings.Contains(errOut, "no CortexDB binary") || !strings.Contains(errOut, "CORTEXDB_MCP_BIN") {
		t.Fatalf("none: %d %s", code, errOut)
	}
	bin := executable(t, filepath.Join(t.TempDir(), "mine"), 0o755)
	if code, _, errOut := hook(t, map[string]string{"HOME": home, "CORTEXDB_MCP_BIN": bin, "TH_ENV": t.TempDir()}, "", "install"); code != 1 || !strings.Contains(errOut, "env file") {
		t.Fatalf("env is a folder: %d %s", code, errOut)
	}
	blocked := t.TempDir()
	executable(t, filepath.Join(blocked, "cortexdb"), 0o600)
	if code, _, errOut := hook(t, map[string]string{"HOME": home, "CORTEXDB_MCP_BIN": bin, "XDG_CACHE_HOME": blocked}, "", "install"); code != 1 || !strings.Contains(errOut, "autorecall") {
		t.Fatalf("cache blocked: %d %s", code, errOut)
	}
	if code, _, errOut := hook(t, map[string]string{"HOME": home}, "", "install", "now"); code != 2 || !strings.Contains(errOut, "usage") {
		t.Fatalf("usage: %d %s", code, errOut)
	}
}

// newer compares dotted versions as numbers; no version is the oldest.
func TestNewerComparesVersions(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{{"2.132.0", "2.99.0", true}, {"2.99.0", "2.132.0", false}, {"2.1", "2.1.0", false}, {"2.1.1", "2.1", true}, {"1.0.0", "", true}, {"", "1.0.0", false}, {"x", "", false}} {
		if newer(c.a, c.b) != c.want {
			t.Errorf("newer(%q, %q) = %v", c.a, c.b, !c.want)
		}
	}
}
