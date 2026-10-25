package main

import (
	"strings"
	"testing"

	"github.com/lfoletti/thalamus/gateway/version"
)

// version, or --version, prints the build of the binary; with an argument
// it is a wrong command line.
func TestVersionIsPrinted(t *testing.T) {
	t.Cleanup(version.SetForTest("v9.9.9", nil))
	// TH_ENV names a file that does not exist: version reads no key file.
	env := map[string]string{"TH_ENV": "/nonexistent/thalamus/env"}
	for _, arg := range []string{"version", "--version"} {
		if code, out, errOut := hook(t, env, "", arg); code != 0 || out != "th-hook v9.9.9\n" || errOut != "" {
			t.Fatalf("%s: %d %q %q", arg, code, out, errOut)
		}
	}
	if code, _, errOut := hook(t, nil, "", "version", "now"); code != 2 || !strings.Contains(errOut, "usage:") {
		t.Fatalf("version now: %d %q", code, errOut)
	}
}
