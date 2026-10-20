package keys

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"

	"github.com/lfoletti/granular-cortex/gateway/rights"
)

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

// Issue writes a key confined to the brain's collection into the server's
// key file, with the clearance of the access, and keeps the keys already
// there. The file it leaves is one the server's own loader accepts.
func TestAnIssuedKeyIsConfinedToTheBrainsCollection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	if err := os.WriteFile(path, []byte(`{"keys":[{"id":"operator","secret":"op-secret","clearance":"read-write"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	k, err := Issue(path, "lfoletti", "team/acme", "team-acme", rights.Write)
	if err != nil {
		t.Fatal(err)
	}
	if k.ID != "gc/lfoletti/team/acme" || len(k.Secret) < 32 || k.Clearance != authz.ReadWrite {
		t.Fatalf("issued: %+v", k)
	}
	r, err := Issue(path, "mdupont", "team/acme", "team-acme", rights.Read)
	if err != nil || r.Clearance != authz.ReadOnly || r.Secret == k.Secret {
		t.Fatalf("reader: %+v, %v", r, err)
	}
	ks := keysOf(t, path)
	if len(ks) != 3 || ks[0].ID != "operator" || ks[1].ID != k.ID || ks[1].Secret != k.Secret || strings.Join(ks[1].Scope.Collections, ",") != "team-acme" || ks[2].Clearance != authz.ReadOnly {
		t.Fatalf("file: %+v", ks)
	}
	if _, err := authz.Load(path); err != nil {
		t.Fatalf("the server's loader refuses the file: %v", err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode().Perm())
	}
}

// No access, no key. A key already issued to the user for the brain is not
// issued twice: revoke it first.
func TestNoAccessAndNoDuplicateKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	if _, err := Issue(path, "stranger", "team/acme", "team-acme", rights.None); err == nil || !strings.Contains(err.Error(), "no access") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("a refused issue wrote the file")
	}
	if _, err := Issue(path, "lfoletti", "team/acme", "team-acme", rights.Write); err != nil {
		t.Fatal(err)
	}
	if _, err := Issue(path, "lfoletti", "team/acme", "team-acme", rights.Write); err == nil || !strings.Contains(err.Error(), "already issued") {
		t.Fatalf("got %v", err)
	}
	if len(keysOf(t, path)) != 1 {
		t.Fatal("the refused issue changed the file")
	}
}

// Revoke removes the user's key for the brain and nothing else; revoking a
// key that is not there is an error.
func TestRevokeRemovesOneKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	for _, u := range []string{"lfoletti", "mdupont"} {
		if _, err := Issue(path, u, "team/acme", "team-acme", rights.Write); err != nil {
			t.Fatal(err)
		}
	}
	if err := Revoke(path, "lfoletti", "team/acme"); err != nil {
		t.Fatal(err)
	}
	ks := keysOf(t, path)
	if len(ks) != 1 || ks[0].ID != "gc/mdupont/team/acme" {
		t.Fatalf("file: %+v", ks)
	}
	if err := Revoke(path, "lfoletti", "team/acme"); err == nil || !strings.Contains(err.Error(), "no key") {
		t.Fatalf("got %v", err)
	}
	if err := Revoke(path, "mdupont", "team/acme"); err != nil {
		t.Fatal(err)
	}
	if _, err := authz.Load(path); err == nil || !strings.Contains(err.Error(), "no keys") {
		t.Fatalf("an empty file should be one the server refuses to load: %v", err)
	}
}

// Holder maps a secret back to the user and brain the gateway issued it
// for. A secret that is not in the file, or a key the gateway did not issue,
// is refused.
func TestHolderNamesTheUserBehindASecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	if err := os.WriteFile(path, []byte(`{"keys":[{"id":"operator","secret":"op-secret","clearance":"read-write"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	k, err := Issue(path, "lfoletti", "team/acme", "team-acme", rights.Write)
	if err != nil {
		t.Fatal(err)
	}
	user, brain, err := Holder(path, k.Secret)
	if err != nil || user != "lfoletti" || brain != "team/acme" {
		t.Fatalf("got %q %q %v", user, brain, err)
	}
	if _, _, err := Holder(path, "op-secret"); err == nil || !strings.Contains(err.Error(), "not issued by the gateway") {
		t.Fatalf("operator: %v", err)
	}
	if _, _, err := Holder(path, "nope"); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("unknown: %v", err)
	}
}

// The file is replaced in one step: when the replacement cannot land, the
// file that was there is unchanged.
func TestAFailedWriteLeavesTheFileAsItWas(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.json")
	if _, err := Issue(path, "lfoletti", "team/acme", "team-acme", rights.Write); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	if _, err := Issue(path, "mdupont", "team/acme", "team-acme", rights.Read); err == nil {
		t.Fatal("the write succeeded in a read-only directory")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("the file changed")
	}
	if err := os.WriteFile(filepath.Join(t.TempDir(), "bad.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// A key file the gateway cannot read, or cannot parse, is an error that
// names the file; the gateway never starts a new file over a broken one.
func TestABrokenKeyFileIsNotOverwritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Issue(path, "lfoletti", "team/acme", "team-acme", rights.Write); err == nil || !strings.Contains(err.Error(), "keys.json") {
		t.Fatalf("got %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "{" {
		t.Fatal("the broken file was replaced")
	}
	if _, _, err := Holder(path, "x"); err == nil {
		t.Fatal("Holder read a broken file")
	}
	if err := Revoke(path, "lfoletti", "team/acme"); err == nil {
		t.Fatal("Revoke rewrote a broken file")
	}
	dir := t.TempDir()
	if _, err := Issue(dir, "lfoletti", "team/acme", "team-acme", rights.Write); err == nil || !strings.Contains(err.Error(), "read "+dir) {
		t.Fatalf("a directory: %v", err)
	}
}

// A key file that parses but that the server would refuse (a key without
// a clearance) is not written back with one more key: the gateway never
// leaves a file the server cannot load.
func TestAFileTheServerWouldRefuseIsNotWritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	bad := `{"keys":[{"id":"operator","secret":"op-secret"}]}`
	if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Issue(path, "lfoletti", "team/acme", "team-acme", rights.Write); err == nil || !strings.Contains(err.Error(), "the server refuses") {
		t.Fatalf("got %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != bad {
		t.Fatal("the file changed")
	}
}

// SecretOf finds the key the gateway issued to a user for a brain, so the
// gateway can call that brain on the user's behalf; no key, no secret.
func TestSecretOfFindsTheUsersKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	k, err := Issue(path, "lfoletti", "team/acme", "team-acme", rights.Write)
	if err != nil {
		t.Fatal(err)
	}
	got, err := SecretOf(path, "lfoletti", "team/acme")
	if err != nil || got != k.Secret {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := SecretOf(path, "mdupont", "team/acme"); err == nil || !strings.Contains(err.Error(), "no key") {
		t.Fatalf("got %v", err)
	}
	if _, err := SecretOf(filepath.Join(t.TempDir(), "absent", "keys.json"), "lfoletti", "team/acme"); err == nil || !strings.Contains(err.Error(), "no key") {
		t.Fatalf("absent file: got %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := SecretOf(path, "lfoletti", "team/acme"); err == nil || !strings.Contains(err.Error(), "parse") {
		t.Fatalf("broken file: got %v", err)
	}
}

// A key for a brain that is the whole server (no collection) has no scope:
// the engine then opens every tool its clearance allows, the graph included.
func TestAKeyForAWholeServerIsNotConfined(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	k, err := Issue(path, "lfoletti", "team/acme", "", rights.Read)
	if err != nil {
		t.Fatal(err)
	}
	ks := keysOf(t, path)
	if len(ks) != 1 || !ks[0].Scope.IsZero() || ks[0].Clearance != authz.ReadOnly || k.Collection != "" {
		t.Fatalf("file: %+v", ks)
	}
	if _, err := authz.Load(path); err != nil {
		t.Fatal(err)
	}
}

// List reads the keys the gateway issued in a file, by brain, and leaves
// the other keys (an operator's) out but counted.
func TestListReadsTheIssuedKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	if err := os.WriteFile(path, []byte(`{"keys":[{"id":"operator","secret":"op","clearance":"read-write"},{"id":"gc/odd","secret":"x","clearance":"read-only"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		user, brain, coll string
		a                 rights.Access
	}{
		{"lfoletti", "team/acme", "team-acme", rights.Write},
		{"mdupont", "team/acme", "team-acme", rights.Read},
		{"lfoletti", "org/ref", "", rights.Read},
	} {
		if _, err := Issue(path, c.user, c.brain, c.coll, c.a); err != nil {
			t.Fatal(err)
		}
	}
	held, others, err := List(path)
	if err != nil {
		t.Fatal(err)
	}
	if others != 2 || len(held) != 3 {
		t.Fatalf("held %+v, others %d", held, others)
	}
	// Sorted by brain then user.
	if held[0].Brain != "org/ref" || held[0].User != "lfoletti" || held[0].Clearance != authz.ReadOnly || held[0].Collection != "" {
		t.Fatalf("held[0] %+v", held[0])
	}
	if held[1].User != "lfoletti" || held[1].Brain != "team/acme" || held[1].Clearance != authz.ReadWrite || held[1].Collection != "team-acme" || held[2].User != "mdupont" || held[2].Clearance != authz.ReadOnly {
		t.Fatalf("held %+v", held)
	}
	if _, _, err := List(filepath.Join(t.TempDir(), "absent.json")); err != nil {
		t.Fatalf("a missing file is an empty list: %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := List(path); err == nil {
		t.Fatal("a broken file was listed")
	}
}
