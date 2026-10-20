package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"
)

func writeRights(t *testing.T, dir string) (rightsPath, keysPath string) {
	t.Helper()
	keysPath = filepath.Join(dir, "keys.json")
	rightsPath = filepath.Join(dir, "rights.yaml")
	text := "brains:\n  team/acme:\n    addr: 127.0.0.1:1\n    keys_file: " + keysPath + "\n    collection: team-acme\n    readers: [group:acme]\n    writers: [user:lfoletti]\ngroups:\n  acme: [mdupont]\n"
	if err := os.WriteFile(rightsPath, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return rightsPath, keysPath
}

func call(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = run(context.Background(), args, &out, &errb)
	return code, out.String(), errb.String()
}

// A wrong command line is exit 2 with the usage; a command that fails is
// exit 1 with the error.
func TestTheCommandLineIsChecked(t *testing.T) {
	dir := t.TempDir()
	rightsPath, _ := writeRights(t, dir)
	for _, args := range [][]string{{}, {"rights"}, {"rights", "list", rightsPath}, {"rights", "check"}, {"key", "issue", rightsPath, "lfoletti"}, {"serve"}, {"serve", rightsPath, "--port", "1"}} {
		if code, _, stderr := call(t, args...); code != 2 || !strings.Contains(stderr, "usage:") {
			t.Fatalf("%v: code %d, stderr %q", args, code, stderr)
		}
	}
	if code, _, stderr := call(t, "rights", "check", filepath.Join(dir, "absent.yaml")); code != 1 || !strings.Contains(stderr, "absent.yaml") {
		t.Fatalf("absent file: code %d, stderr %q", code, stderr)
	}
}

// rights check lists the brains; rights of lists a user's grants.
func TestRightsCheckAndOf(t *testing.T) {
	rightsPath, _ := writeRights(t, t.TempDir())
	code, out, _ := call(t, "rights", "check", rightsPath)
	if code != 0 || !strings.Contains(out, "1 brain(s), 1 group(s)") || !strings.Contains(out, "team/acme  127.0.0.1:1  collection=team-acme  readers=1  writers=1") {
		t.Fatalf("check: %d %q", code, out)
	}
	code, out, _ = call(t, "rights", "of", rightsPath, "mdupont")
	if code != 0 || strings.TrimSpace(out) != "team/acme  read" {
		t.Fatalf("of: %d %q", code, out)
	}
	code, out, _ = call(t, "rights", "of", rightsPath, "stranger")
	if code != 0 || !strings.Contains(out, "stranger has access to no brain") {
		t.Fatalf("of stranger: %d %q", code, out)
	}
}

// key issue prints the secret alone, writes the confined key, and refuses a
// user without access or an unknown brain; key revoke removes it.
func TestKeyIssueAndRevoke(t *testing.T) {
	rightsPath, keysPath := writeRights(t, t.TempDir())
	code, out, _ := call(t, "key", "issue", rightsPath, "mdupont", "team/acme")
	secret := strings.TrimSpace(out)
	if code != 0 || len(secret) < 32 || strings.Contains(secret, "\n") {
		t.Fatalf("issue: %d %q", code, out)
	}
	ks, err := authz.Load(keysPath)
	if err != nil {
		t.Fatal(err)
	}
	k, ok := ks.Lookup(secret)
	if !ok || k.ID != "gc/mdupont/team/acme" || k.Clearance != authz.ReadOnly || strings.Join(k.Scope.Collections, ",") != "team-acme" {
		t.Fatalf("key: %+v %v", k, ok)
	}
	if code, _, stderr := call(t, "key", "issue", rightsPath, "stranger", "team/acme"); code != 1 || !strings.Contains(stderr, "no access") {
		t.Fatalf("stranger: %d %q", code, stderr)
	}
	if code, _, stderr := call(t, "key", "issue", rightsPath, "lfoletti", "team/acm"); code != 1 || !strings.Contains(stderr, "unknown brain") {
		t.Fatalf("unknown brain: %d %q", code, stderr)
	}
	if code, out, _ := call(t, "key", "revoke", rightsPath, "mdupont", "team/acme"); code != 0 || !strings.Contains(out, "revoked gc/mdupont/team/acme") {
		t.Fatalf("revoke: %d %q", code, out)
	}
	if code, _, stderr := call(t, "key", "revoke", rightsPath, "mdupont", "team/acme"); code != 1 || !strings.Contains(stderr, "no key") {
		t.Fatalf("revoke again: %d %q", code, stderr)
	}
	if code, _, stderr := call(t, "key", "revoke", rightsPath, "mdupont", "team/acm"); code != 1 || !strings.Contains(stderr, "unknown brain") {
		t.Fatalf("revoke unknown brain: %d %q", code, stderr)
	}
}

// serve listens, prints its address, answers on /mcp/<brain>, and stops
// when the context ends. A port in use, a bad rights file and an audit file
// that cannot be opened are exit 1.
func TestServeListensAndStops(t *testing.T) {
	dir := t.TempDir()
	rightsPath, _ := writeRights(t, dir)
	audit := filepath.Join(dir, "audit.jsonl")
	ctx, cancel := context.WithCancel(context.Background())
	out := &lockedBuffer{}
	done := make(chan int, 1)
	go func() {
		done <- run(ctx, []string{"serve", rightsPath, "--listen", "127.0.0.1:0", "--audit", audit}, out, os.Stderr)
	}()
	var addr string
	for i := 0; i < 100 && addr == ""; i++ {
		time.Sleep(20 * time.Millisecond)
		if line := out.String(); strings.Contains(line, "listening on http://") {
			addr = strings.Fields(strings.SplitN(line, "http://", 2)[1])[0]
		}
	}
	if addr == "" {
		t.Fatalf("no address printed: %q", out.String())
	}
	resp, err := http.Post("http://"+addr+"/mcp/team/acme", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if code, _, stderr := call(t, "serve", rightsPath, "--listen", addr, "--audit", audit); code != 1 || !strings.Contains(stderr, "listen") {
		t.Fatalf("port in use: %d %q", code, stderr)
	}
	cancel()
	if code := <-done; code != 0 {
		t.Fatalf("serve exited %d", code)
	}
	if _, err := os.Stat(audit); err != nil {
		t.Fatalf("audit file: %v", err)
	}
	if code, _, stderr := call(t, "serve", filepath.Join(dir, "absent.yaml")); code != 1 || !strings.Contains(stderr, "absent.yaml") {
		t.Fatalf("bad rights: %d %q", code, stderr)
	}
	if code, _, stderr := call(t, "serve", rightsPath, "--listen", "127.0.0.1:0", "--audit", filepath.Join(dir, "missing", "a.jsonl")); code != 1 || !strings.Contains(stderr, "audit") {
		t.Fatalf("bad audit: %d %q", code, stderr)
	}
}

// lockedBuffer is a bytes.Buffer the test may read while serve writes it.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}
