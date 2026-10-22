package relay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// idp is a fake OpenID Connect provider with one RSA key.
type idp struct {
	key *rsa.PrivateKey
	srv *httptest.Server
}

func newIDP(t *testing.T) *idp {
	t.Helper()
	p := &idp{}
	var err error
	if p.key, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"jwks_uri": p.srv.URL + "/jwks"})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		b64 := func(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kty": "RSA", "kid": "k1", "n": b64(p.key.N.Bytes()), "e": b64(big.NewInt(int64(p.key.E)).Bytes())}}})
	})
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	return p
}

func (p *idp) token(t *testing.T, email string) string {
	t.Helper()
	b64 := func(v any) string {
		data, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(data)
	}
	head := b64(map[string]string{"alg": "RS256", "kid": "k1"})
	body := b64(map[string]any{"iss": p.srv.URL, "aud": "thalamus", "exp": time.Now().Add(time.Hour).Unix(), "email": email})
	sum := sha256.Sum256([]byte(head + "." + body))
	sig, err := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	return head + "." + body + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// POST /v1/login with a valid id_token: the login comes from the claim,
// the keys the rights allow are issued into the brains' key files, and
// the answer carries one key for the federated endpoint and one per
// brain. A second login returns the same keys. The login is audited.
func TestLoginIssuesTheKeysTheRightsAllow(t *testing.T) {
	b := start(t)
	p := newIDP(t)
	b.rights.Identity = &rights.Identity{Issuer: p.srv.URL, Audience: "thalamus", Claim: "email", StripDomain: true}
	acme := b.rights.Brains["team/acme"]
	acme.Pyramid = &rights.Pyramid{Collection: "acme-pyramid"}
	b.rights.Brains["team/acme"] = acme
	status, body := b.rest(t, http.MethodPost, "/v1/login", "", "")
	if status != http.StatusUnauthorized || !strings.Contains(body, "id_token") {
		t.Fatalf("no token: %d %s", status, body)
	}
	req, _ := http.NewRequest(http.MethodPost, b.url+"/v1/login", nil)
	req.Header.Set("Authorization", "Bearer "+p.token(t, "LFoletti@acme.example"))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var got loginAnswer
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("login: %d %v", res.StatusCode, err)
	}
	_ = res.Body.Close()
	if got.Login != "lfoletti" || len(got.Brains) != 2 || got.Brains[0].Brain != "team/acme" || got.Brains[0].Access != "write" || got.Brains[1].Brain != "team/other" || got.Key == "" || got.Key != got.Brains[0].Key {
		t.Fatalf("answer: %+v", got)
	}
	for _, g := range got.Brains {
		secret, err := keys.SecretOf(b.keysFile, "lfoletti", g.Brain)
		if err != nil || secret != g.Key {
			t.Fatalf("%s: key file has %q, %v; answer has %q", g.Brain, secret, err, g.Key)
		}
	}
	// The key of a brain with a pyramid holds both collections.
	if held, _, _ := keys.List(b.keysFile); len(held) != 2 || strings.Join(held[0].Collections, ",") != "team-acme,acme-pyramid" || strings.Join(held[1].Collections, ",") != "other" {
		t.Fatalf("collections: %+v", held)
	}
	// The key works on the federated endpoint at once.
	if st, body := b.rest(t, http.MethodGet, "/v1/brains", got.Key, ""); st != http.StatusOK || !strings.Contains(body, "team/acme") {
		t.Fatalf("key in use: %d %s", st, body)
	}
	if l := b.auditLines(t)[0]; l["tool"] != "login" || l["user"] != "lfoletti" || l["ok"] != true || fmt.Sprint(l["ids"].(map[string]any)["brains"]) != "[team/acme team/other]" {
		t.Fatalf("audit: %+v", l)
	}
	// Again: the same keys, nothing issued twice.
	req, _ = http.NewRequest(http.MethodPost, b.url+"/v1/login", nil)
	req.Header.Set("Authorization", "Bearer "+p.token(t, "lfoletti@acme.example"))
	res, _ = http.DefaultClient.Do(req)
	var again loginAnswer
	_ = json.NewDecoder(res.Body).Decode(&again)
	_ = res.Body.Close()
	if again.Key != got.Key || again.Brains[1].Key != got.Brains[1].Key {
		t.Fatalf("second login changed the keys: %+v", again)
	}
}

// Refusals: a bad token (401), a login with no brain (403), no identity
// block in the rights (404), a wrong method (405), a key file that cannot
// be written (500). Nothing is issued on a refusal.
func TestLoginRefusals(t *testing.T) {
	b := start(t)
	p := newIDP(t)
	if status, body := b.rest(t, http.MethodPost, "/v1/login", p.token(t, "lfoletti@acme.example"), ""); status != http.StatusNotFound || !strings.Contains(body, "no identity provider") {
		t.Fatalf("no block: %d %s", status, body)
	}
	b.rights.Identity = &rights.Identity{Issuer: p.srv.URL, Audience: "thalamus", StripDomain: true}
	for _, c := range []struct {
		method, token string
		status        int
		want          string
	}{
		{http.MethodGet, p.token(t, "lfoletti@acme.example"), http.StatusMethodNotAllowed, "POST"},
		{http.MethodPost, "not.a.token", http.StatusUnauthorized, "token"},
		{http.MethodPost, p.token(t, "nobody@acme.example"), http.StatusForbidden, "no brain"},
	} {
		if status, body := b.rest(t, c.method, "/v1/login", c.token, ""); status != c.status || !strings.Contains(body, c.want) {
			t.Fatalf("%s %q: %d %s", c.method, c.token[:8], status, body)
		}
	}
	if data, _ := os.ReadFile(b.keysFile); len(data) != 0 {
		t.Fatalf("something was issued: %s", data)
	}
	if err := os.WriteFile(b.keysFile, []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if status, body := b.rest(t, http.MethodPost, "/v1/login", p.token(t, "lfoletti@acme.example"), ""); status != http.StatusInternalServerError || !strings.Contains(body, "key file") {
		t.Fatalf("broken key file: %d %s", status, body)
	}
	if l := b.auditLines(t)[0]; l["tool"] != "login" || l["ok"] != false {
		t.Fatalf("audit: %+v", l)
	}
	// A changed identity block makes a new verifier (another issuer).
	b.rights.Identity = &rights.Identity{Issuer: "http://127.0.0.1:1", Audience: "thalamus"}
	if status, body := b.rest(t, http.MethodPost, "/v1/login", p.token(t, "lfoletti@acme.example"), ""); status != http.StatusUnauthorized || !strings.Contains(body, "discovery") {
		t.Fatalf("other issuer: %d %s", status, body)
	}
}

// GET /v1/identity is public: it names the provider and the client id, so
// that a tool needs only the gateway's address. 404 without the block,
// 405 on another method.
func TestIdentityIsPublic(t *testing.T) {
	b := start(t)
	if status, body := b.rest(t, http.MethodGet, "/v1/identity", "", ""); status != http.StatusNotFound || !strings.Contains(body, "no identity provider") {
		t.Fatalf("no block: %d %s", status, body)
	}
	b.rights.Identity = &rights.Identity{Issuer: "https://login.example", Audience: "thalamus", ClientID: "thalamus-cli"}
	status, body := b.rest(t, http.MethodGet, "/v1/identity", "", "")
	if status != http.StatusOK || body != `{"issuer":"https://login.example","audience":"thalamus","client_id":"thalamus-cli"}`+"\n" {
		t.Fatalf("identity: %d %s", status, body)
	}
	if status, body := b.rest(t, http.MethodPost, "/v1/identity", "", ""); status != http.StatusMethodNotAllowed || !strings.Contains(body, "GET") {
		t.Fatalf("post: %d %s", status, body)
	}
	if lines := b.auditLines(t); len(lines) != 0 {
		t.Fatalf("a public read is not audited: %+v", lines)
	}
}
