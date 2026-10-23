package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// DEMO (v2.138.0): REST keeps a revoked key until restart.
func TestDemoARevokedKeyStillWorksOverREST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	write := func(body string) {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"keys":[{"id":"a","secret":"a-secret","clearance":"read-write"},{"id":"b","secret":"b-secret","clearance":"read-only"}]}`)
	srv := newPolicyServer(t, Options{KeyFile: path})
	wantStatus(t, srv, http.MethodGet, "/v1/health", "a-secret", "", http.StatusOK)
	write(`{"keys":[{"id":"b","secret":"b-secret","clearance":"read-only"},{"id":"c","secret":"c-secret","clearance":"read-only"}]}`)
	sa, _ := do(t, srv, http.MethodGet, "/v1/health", "a-secret", "")
	sc, _ := do(t, srv, http.MethodGet, "/v1/health", "c-secret", "")
	t.Logf("revoked a: %d (want 401); new c: %d (want 200)", sa, sc)
	if sa != http.StatusUnauthorized || sc != http.StatusOK {
		t.Errorf("DEFECT: the key file is read once at start")
	}
}
