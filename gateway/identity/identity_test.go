package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// A fake provider: a discovery document, a JWKS with one RSA key, and a
// way to sign tokens with it or with a stranger's key.
type provider struct {
	key   *rsa.PrivateKey
	other *rsa.PrivateKey
	srv   *httptest.Server
	hits  int
}

func newProvider(t *testing.T) *provider {
	t.Helper()
	p := &provider{}
	var err error
	if p.key, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		t.Fatal(err)
	}
	if p.other, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": p.srv.URL, "jwks_uri": p.srv.URL + "/jwks"})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		p.hits++
		b64 := func(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{
			{"kty": "RSA", "kid": "k1", "alg": "RS256", "use": "sig", "n": b64(p.key.N.Bytes()), "e": b64(big.NewInt(int64(p.key.E)).Bytes())},
			{"kty": "EC", "kid": "ec1"},
		}})
	})
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	return p
}

// token signs claims with the key, under the kid, RS256 unless alg says
// otherwise.
func (p *provider) token(t *testing.T, key *rsa.PrivateKey, kid, alg string, claims map[string]any) string {
	t.Helper()
	if _, ok := claims["iss"]; !ok {
		claims["iss"] = p.srv.URL
	}
	if _, ok := claims["aud"]; !ok {
		claims["aud"] = "thalamus"
	}
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
	}
	return sign(t, key, kid, alg, claims)
}

func sign(t *testing.T, key *rsa.PrivateKey, kid, alg string, claims map[string]any) string {
	t.Helper()
	b64 := func(v any) string {
		data, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(data)
	}
	head := b64(map[string]string{"alg": alg, "kid": kid, "typ": "JWT"})
	body := b64(claims)
	sig, err := signRS256(key, head+"."+body)
	if err != nil {
		t.Fatal(err)
	}
	return head + "." + body + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// A valid token gives the login from the claim, the domain stripped when
// asked; the JWKS is fetched once and kept.
func TestAValidTokenGivesTheLogin(t *testing.T) {
	p := newProvider(t)
	v := &Verifier{Issuer: p.srv.URL, Audience: "thalamus", Claim: "email", StripDomain: true}
	login, err := v.Login(context.Background(), p.token(t, p.key, "k1", "RS256", map[string]any{"email": "LFoletti@acme.example"}))
	if err != nil || login != "lfoletti" {
		t.Fatalf("got %q, %v", login, err)
	}
	if login, err = v.Login(context.Background(), p.token(t, p.key, "k1", "RS256", map[string]any{"email": "mdupont@acme.example"})); err != nil || login != "mdupont" {
		t.Fatalf("second: %q, %v", login, err)
	}
	if p.hits != 1 {
		t.Fatalf("jwks fetched %d times", p.hits)
	}
	// aud as a list, the claim kept whole, the default claim.
	v2 := &Verifier{Issuer: p.srv.URL, Audience: "thalamus"}
	if login, err = v2.Login(context.Background(), p.token(t, p.key, "k1", "RS256", map[string]any{"email": "lfoletti@acme.example", "aud": []string{"other", "thalamus"}})); err != nil || login != "lfoletti@acme.example" {
		t.Fatalf("list aud: %q, %v", login, err)
	}
	v3 := &Verifier{Issuer: p.srv.URL, Audience: "thalamus", Claim: "preferred_username"}
	if login, err = v3.Login(context.Background(), p.token(t, p.key, "k1", "RS256", map[string]any{"preferred_username": "lfoletti"})); err != nil || login != "lfoletti" {
		t.Fatalf("other claim: %q, %v", login, err)
	}
}

// Every way a token can be wrong is refused, with a reason, and never
// gives a login.
func TestAWrongTokenIsRefused(t *testing.T) {
	p := newProvider(t)
	v := &Verifier{Issuer: p.srv.URL, Audience: "thalamus", Claim: "email"}
	good := map[string]any{"email": "lfoletti@acme.example"}
	for _, c := range []struct {
		name, token, want string
	}{
		{"not a jwt", "abc", "three parts"},
		{"bad header", "!!!.e30.e30", "header"},
		{"bad claims", base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"k1"}`)) + ".!!!.e30", "claims"},
		{"other algorithm", p.token(t, p.key, "k1", "HS256", good), "algorithm"},
		{"unknown kid", p.token(t, p.key, "k9", "RS256", good), "key \"k9\""},
		{"a key that is not RSA", p.token(t, p.key, "ec1", "RS256", good), "key \"ec1\""},
		{"stranger's key", p.token(t, p.other, "k1", "RS256", good), "signature"},
		{"bad signature bytes", p.token(t, p.key, "k1", "RS256", good) + "!", "signature"},
		{"wrong issuer", p.token(t, p.key, "k1", "RS256", map[string]any{"email": "x", "iss": "https://elsewhere"}), "issuer"},
		{"wrong audience", p.token(t, p.key, "k1", "RS256", map[string]any{"email": "x", "aud": "other"}), "audience"},
		{"wrong audience list", p.token(t, p.key, "k1", "RS256", map[string]any{"email": "x", "aud": []string{"a", "b"}}), "audience"},
		{"expired", p.token(t, p.key, "k1", "RS256", map[string]any{"email": "x", "exp": time.Now().Add(-time.Minute).Unix()}), "expired"},
		{"no expiry", p.token(t, p.key, "k1", "RS256", map[string]any{"email": "x", "exp": nil}), "expiry"},
		{"no claim", p.token(t, p.key, "k1", "RS256", map[string]any{"name": "x"}), "claim \"email\""},
		{"empty claim", p.token(t, p.key, "k1", "RS256", map[string]any{"email": "  "}), "claim \"email\""},
	} {
		login, err := v.Login(context.Background(), c.token)
		if err == nil || login != "" || !strings.Contains(err.Error(), c.want) {
			t.Fatalf("%s: got %q, %v", c.name, login, err)
		}
	}
}

// A provider that cannot be reached, or that publishes nonsense, is an
// error, and nothing is kept from it.
func TestAProviderThatFailsIsAnError(t *testing.T) {
	p := newProvider(t)
	good := p.token(t, p.key, "k1", "RS256", map[string]any{"email": "x@y"})
	for _, c := range []struct {
		issuer, want string
	}{
		{"http://127.0.0.1:1", "discovery"},
		{"%zz", "discovery"},
	} {
		v := &Verifier{Issuer: c.issuer, Audience: "thalamus"}
		if _, err := v.Login(context.Background(), good); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Fatalf("%s: %v", c.issuer, err)
		}
	}
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			if r.URL.Query().Get("x") == "" && strings.Contains(r.Host, "127.0.0.1") {
				_ = json.NewEncoder(w).Encode(map[string]string{"jwks_uri": "http://" + r.Host + "/jwks"})
			}
		case "/jwks":
			_, _ = w.Write([]byte("not json"))
		}
	}))
	defer bad.Close()
	v := &Verifier{Issuer: bad.URL, Audience: "thalamus"}
	tok := sign(t, p.key, "k1", "RS256", map[string]any{"iss": bad.URL, "aud": "thalamus", "exp": time.Now().Add(time.Hour).Unix(), "email": "x"})
	if _, err := v.Login(context.Background(), tok); err == nil || !strings.Contains(err.Error(), "jwks") {
		t.Fatalf("bad jwks: %v", err)
	}
	if _, err := v.Login(context.Background(), tok); err == nil || !strings.Contains(err.Error(), "jwks") {
		t.Fatalf("bad jwks again (nothing kept): %v", err)
	}
	// A discovery document without jwks_uri, and a JWKS whose key has bad numbers.
	odd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{"jwks_uri": "http://" + r.Host + "/jwks"})
		case "/jwks":
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kty": "RSA", "kid": "k1", "n": "!!!", "e": "AQAB"}, {"kty": "RSA", "kid": "k2", "n": "AQAB", "e": "!!!"}}})
		}
	}))
	defer odd.Close()
	v = &Verifier{Issuer: odd.URL, Audience: "thalamus"}
	for _, kid := range []string{"k1", "k2"} {
		tok = sign(t, p.key, kid, "RS256", map[string]any{"iss": odd.URL, "aud": "thalamus", "exp": time.Now().Add(time.Hour).Unix(), "email": "x"})
		if _, err := v.Login(context.Background(), tok); err == nil || !strings.Contains(err.Error(), "key \""+kid+"\"") {
			t.Fatalf("bad key %s: %v", kid, err)
		}
	}
	none := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("{}")) }))
	defer none.Close()
	v = &Verifier{Issuer: none.URL, Audience: "thalamus"}
	tok = sign(t, p.key, "k1", "RS256", map[string]any{"iss": none.URL, "aud": "thalamus", "exp": time.Now().Add(time.Hour).Unix(), "email": "x"})
	if _, err := v.Login(context.Background(), tok); err == nil || !strings.Contains(err.Error(), "jwks_uri") {
		t.Fatalf("no jwks_uri: %v", err)
	}
}
