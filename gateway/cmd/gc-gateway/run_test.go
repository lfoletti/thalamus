package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"

	"github.com/lfoletti/granular-cortex/gateway/keys"
	"github.com/lfoletti/granular-cortex/gateway/rights"
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

func writeRightsTwo(t *testing.T, dir string) (rightsPath, keysPath string) {
	t.Helper()
	keysPath = filepath.Join(dir, "keys.json")
	rightsPath = filepath.Join(dir, "rights.yaml")
	text := "brains:\n  team/acme:\n    addr: 127.0.0.1:1\n    keys_file: " + keysPath + "\n    collection: team-acme\n    readers: [group:acme]\n    writers: [user:lfoletti]\n  org/ref:\n    addr: 127.0.0.1:1\n    keys_file: " + keysPath + "\n    writers: [user:kmlead]\ngroups:\n  acme: [mdupont, lfoletti]\n"
	if err := os.WriteFile(rightsPath, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return rightsPath, keysPath
}

// key list shows, per brain, who holds a key and whether it still matches
// the right: ok, no right any more, clearance or collection out of date.
func TestKeyListShowsTheKeysAgainstTheRights(t *testing.T) {
	rightsPath, keysPath := writeRightsTwo(t, t.TempDir())
	if err := os.WriteFile(keysPath, []byte(`{"keys":[{"id":"operator","secret":"op","clearance":"read-write"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ user, brain, coll, access string }{
		{"lfoletti", "team/acme", "team-acme", "write"},
		{"mdupont", "team/acme", "old-coll", "read"},
		{"ghost", "team/acme", "team-acme", "read"},
		{"kmlead", "org/ref", "", "read"},
	} {
		a := map[string]rights.Access{"read": rights.Read, "write": rights.Write}[c.access]
		if _, err := keys.Issue(keysPath, c.user, c.brain, c.coll, a); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := keys.Issue(keysPath, "other", "org/ref", "some-coll", rights.Write); err != nil {
		t.Fatal(err)
	}
	code, out, _ := call(t, "key", "list", rightsPath)
	want := "org/ref  kmlead  read-only  stale: right is write\n" +
		"org/ref  other  read-write  stale: no right\n" +
		"team/acme  ghost  read-only  stale: no right\n" +
		"team/acme  lfoletti  read-write  ok\n" +
		"team/acme  mdupont  read-only  stale: collection is team-acme\n" +
		"1 other key(s) in " + keysPath + "\n"
	if code != 0 || out != want {
		t.Fatalf("code %d, out:\n%s\nwant:\n%s", code, out, want)
	}
	if err := os.WriteFile(keysPath, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := call(t, "key", "list", rightsPath); code != 1 || !strings.Contains(stderr, "keys.json") {
		t.Fatalf("broken file: %d %q", code, stderr)
	}
}

// key sync issues the missing keys, replaces the stale ones (new secret),
// revokes the ones without a right, prints each action with the secret
// when there is one, and does nothing the second time. --dry-run prints
// the plan and changes nothing.
func TestKeySyncBringsTheKeysToTheRights(t *testing.T) {
	rightsPath, keysPath := writeRightsTwo(t, t.TempDir())
	if _, err := keys.Issue(keysPath, "ghost", "team/acme", "team-acme", rights.Read); err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Issue(keysPath, "mdupont", "team/acme", "old-coll", rights.Read); err != nil {
		t.Fatal(err)
	}
	code, out, _ := call(t, "key", "sync", rightsPath, "--dry-run")
	if code != 0 || out != "would issue gc/kmlead/org/ref\nwould revoke gc/ghost/team/acme\nwould issue gc/lfoletti/team/acme\nwould reissue gc/mdupont/team/acme\n" {
		t.Fatalf("dry run: %d %q", code, out)
	}
	ks, _ := authz.Load(keysPath)
	if _, ok := ks.Lookup("x"); ok || len(keysOf(t, keysPath)) != 2 {
		t.Fatal("the dry run changed the file")
	}
	code, out, _ = call(t, "key", "sync", rightsPath)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if code != 0 || len(lines) != 4 || !strings.HasPrefix(lines[0], "issued gc/kmlead/org/ref ") || lines[1] != "revoked gc/ghost/team/acme" || !strings.HasPrefix(lines[2], "issued gc/lfoletti/team/acme ") || !strings.HasPrefix(lines[3], "reissued gc/mdupont/team/acme ") {
		t.Fatalf("sync: %d %q", code, out)
	}
	secret := strings.Fields(lines[3])[2]
	held := keysOf(t, keysPath)
	if len(held) != 3 {
		t.Fatalf("file: %+v", held)
	}
	for _, k := range held {
		switch k.ID {
		case "gc/mdupont/team/acme":
			if k.Secret != secret || strings.Join(k.Scope.Collections, ",") != "team-acme" {
				t.Fatalf("mdupont: %+v", k)
			}
		case "gc/kmlead/org/ref":
			if !k.Scope.IsZero() || k.Clearance != authz.ReadWrite {
				t.Fatalf("kmlead: %+v", k)
			}
		case "gc/lfoletti/team/acme":
		default:
			t.Fatalf("unexpected %+v", k)
		}
	}
	if code, out, _ := call(t, "key", "sync", rightsPath); code != 0 || out != "nothing to do\n" {
		t.Fatalf("second sync: %d %q", code, out)
	}
	if code, out, _ := call(t, "key", "list", rightsPath); code != 0 || strings.Contains(out, "stale") || strings.Count(out, "  ok\n") != 3 {
		t.Fatalf("list after sync: %d %q", code, out)
	}
	if code, _, stderr := call(t, "key", "sync", rightsPath, "--nope"); code != 2 || !strings.Contains(stderr, "usage") {
		t.Fatalf("bad flag: %d %q", code, stderr)
	}
	if err := os.WriteFile(keysPath, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := call(t, "key", "sync", rightsPath); code != 1 || !strings.Contains(stderr, "keys.json") {
		t.Fatalf("broken file: %d %q", code, stderr)
	}
}

func keysOf(t *testing.T, path string) []authz.Key {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Keys []authz.Key `json:"keys"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	return f.Keys
}

// A key with a collection on a brain that is the whole server is stale.
func TestAKeyWithACollectionOnAWholeServerIsStale(t *testing.T) {
	rightsPath, keysPath := writeRightsTwo(t, t.TempDir())
	if _, err := keys.Issue(keysPath, "kmlead", "org/ref", "some-coll", rights.Write); err != nil {
		t.Fatal(err)
	}
	if code, out, _ := call(t, "key", "list", rightsPath); code != 0 || out != "org/ref  kmlead  read-write  stale: brain is the whole server\n" {
		t.Fatalf("%d %q", code, out)
	}
}

// key sync stops at the first action that fails and names the file; a
// missing rights file or none given is an error.
func TestKeySyncStopsOnAFailedAction(t *testing.T) {
	for _, verb := range []string{"issue", "revoke"} {
		dir := t.TempDir()
		rightsPath, keysPath := writeRightsTwo(t, dir)
		if verb == "revoke" {
			if _, err := keys.Issue(keysPath, "ghost", "team/acme", "team-acme", rights.Read); err != nil {
				t.Fatal(err)
			}
			// A first sync settles everything, ghost included; ghost's key is
			// put back so that the revoke is the only action pending.
			if _, _, err := call(t, "key", "sync", rightsPath); err != "" {
				t.Fatal(err)
			}
			if _, err := keys.Issue(keysPath, "ghost", "team/acme", "team-acme", rights.Read); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		code, _, stderr := call(t, "key", "sync", rightsPath)
		_ = os.Chmod(dir, 0o700)
		if code != 1 || !strings.Contains(stderr, "keys.json") {
			t.Fatalf("%s: code %d, stderr %q", verb, code, stderr)
		}
	}
	if code, _, stderr := call(t, "key", "sync"); code != 2 || !strings.Contains(stderr, "usage") {
		t.Fatalf("no file: %d %q", code, stderr)
	}
	if code, _, stderr := call(t, "key", "sync", filepath.Join(t.TempDir(), "absent.yaml")); code != 1 || !strings.Contains(stderr, "absent.yaml") {
		t.Fatalf("absent file: %d %q", code, stderr)
	}
}
