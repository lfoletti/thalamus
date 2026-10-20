package login

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// A fake identity provider for the authorization-code flow: discovery,
// authorize (redirects with a code), token (checks the PKCE verifier and
// returns an id_token whose payload carries the nonce). The gateway of
// these tests does not verify signatures, so the token is unsigned.
type idp struct {
	srv      *httptest.Server
	mu       sync.Mutex
	codes    map[string]string // code -> challenge
	nonces   map[string]string // code -> nonce
	clientID string
	tokenErr string // when set, the token endpoint answers 400 with it
	noToken  bool   // when set, the token answer has no id_token
	wrongN   bool   // when set, the id_token carries another nonce
	authErr  string // when set, authorize redirects with error=
}

func newIDP(t *testing.T) *idp {
	t.Helper()
	p := &idp{codes: map[string]string{}, nonces: map[string]string{}, clientID: "thalamus-cli"}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"authorization_endpoint": p.srv.URL + "/authorize", "token_endpoint": p.srv.URL + "/token"})
	})
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("client_id") != p.clientID || q.Get("response_type") != "code" || q.Get("code_challenge_method") != "S256" || q.Get("scope") != "openid email" || q.Get("code_challenge") == "" || q.Get("nonce") == "" || q.Get("state") == "" {
			http.Error(w, "bad authorize request: "+r.URL.RawQuery, http.StatusBadRequest)
			return
		}
		if p.authErr != "" {
			http.Redirect(w, r, q.Get("redirect_uri")+"?error="+p.authErr+"&state="+q.Get("state"), http.StatusFound)
			return
		}
		code := fmt.Sprintf("code-%d", len(p.codes)+1)
		p.mu.Lock()
		p.codes[code] = q.Get("code_challenge")
		p.nonces[code] = q.Get("nonce")
		p.mu.Unlock()
		http.Redirect(w, r, q.Get("redirect_uri")+"?code="+code+"&state="+url.QueryEscape(q.Get("state")), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if p.tokenErr != "" {
			http.Error(w, p.tokenErr, http.StatusBadRequest)
			return
		}
		sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		p.mu.Lock()
		challenge, nonce := p.codes[r.Form.Get("code")], p.nonces[r.Form.Get("code")]
		p.mu.Unlock()
		if r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("client_id") != p.clientID || r.Form.Get("redirect_uri") == "" || challenge != base64.RawURLEncoding.EncodeToString(sum[:]) {
			http.Error(w, "bad token request", http.StatusBadRequest)
			return
		}
		if p.noToken {
			_, _ = w.Write([]byte(`{"token_type":"Bearer"}`))
			return
		}
		if p.wrongN {
			nonce = "other"
		}
		payload, _ := json.Marshal(map[string]string{"nonce": nonce, "email": "lfoletti@acme.example"})
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"})
	})
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	return p
}

// A fake gateway: /v1/identity names the provider, /v1/login turns the
// bearer into an answer when the token carries the expected email.
func newGateway(t *testing.T, p *idp, identityStatus int, clientID string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/identity", func(w http.ResponseWriter, r *http.Request) {
		if identityStatus != http.StatusOK {
			w.WriteHeader(identityStatus)
			_, _ = w.Write([]byte(`{"error":"no identity provider is configured"}`))
			return
		}
		issuer := p.srv.URL
		if clientID == "bad-issuer" {
			issuer = "http://127.0.0.1:1"
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": issuer, "audience": "thalamus", "client_id": clientID})
	})
	mux.HandleFunc("/v1/login", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		parts := strings.Split(token, ".")
		payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
		if !strings.Contains(string(payload), "lfoletti@acme.example") {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"token: the signature does not verify"}`))
			return
		}
		_, _ = w.Write([]byte(`{"login":"lfoletti","key":"k-acme","brains":[{"brain":"team/acme","access":"write","key":"k-acme"},{"brain":"team/other","access":"write","key":"k-other"}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// testCtx bounds every flow: a flow that waits for a browser that never
// comes back ends in 10 s.
func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// browser acts as the employee's browser: it opens the address and
// follows the redirects to the loopback callback.
func browser(t *testing.T, wantBody string) func(string) error {
	return func(u string) error {
		// A bounded browser: a callback that nobody serves ends in 5 s.
		res, err := (&http.Client{Timeout: 5 * time.Second}).Get(u)
		if err != nil {
			t.Errorf("the browser got no answer: %v", err)
			return err
		}
		defer func() { _ = res.Body.Close() }()
		var body strings.Builder
		_, _ = fmt.Fprint(&body, res.StatusCode, " ")
		buf := make([]byte, 4096)
		n, _ := res.Body.Read(buf)
		body.Write(buf[:n])
		if !strings.Contains(body.String(), wantBody) {
			t.Errorf("browser saw %q, want %q", body.String(), wantBody)
		}
		return nil
	}
}

// The whole flow: the gateway names the provider, the browser signs in,
// the code is exchanged with the PKCE verifier, the nonce is checked, the
// id_token goes to the gateway, and the keys come back. The address is
// printed before the browser opens, so an employee without a browser can
// paste it.
func TestTheFlowSignsInAndBringsTheKeys(t *testing.T) {
	p := newIDP(t)
	gw := newGateway(t, p, http.StatusOK, "thalamus-cli")
	var out strings.Builder
	// The browser comes back twice (a reload): the first callback wins,
	// the second is answered and dropped. The HTTP client is the caller's.
	f := &Flow{Gateway: gw.URL + "/", Out: &out, HTTP: &http.Client{}, Open: func(u string) error {
		_ = browser(t, "Signed in")(u)
		return browser(t, "Signed in")(u)
	}}
	a, err := f.Run(testCtx(t))
	if err != nil {
		t.Fatal(err)
	}
	if a.Login != "lfoletti" || a.Key != "k-acme" || len(a.Brains) != 2 || a.Brains[1].Key != "k-other" {
		t.Fatalf("answer: %+v", a)
	}
	if !strings.Contains(out.String(), "Open this address to sign in:\n"+p.srv.URL+"/authorize?") || !strings.Contains(out.String(), "redirect_uri=http%3A%2F%2F127.0.0.1%3A") {
		t.Fatalf("out: %s", out.String())
	}
	// A browser that fails to open is told, and the flow still waits for
	// the callback: the employee opens the printed address by hand.
	out.Reset()
	f = &Flow{Gateway: gw.URL, Out: &out, Open: func(u string) error {
		go func() { _ = browser(t, "Signed in")(u) }()
		return errors.New("no display")
	}}
	if a, err = f.Run(testCtx(t)); err != nil || a.Login != "lfoletti" {
		t.Fatalf("no display: %+v, %v", a, err)
	}
	if !strings.Contains(out.String(), "cannot open the browser: no display\nOpen the address yourself.") {
		t.Fatalf("out: %s", out.String())
	}
}

// Every refusal along the way is an error that names the step.
func TestTheFlowRefusals(t *testing.T) {
	p := newIDP(t)
	okGW := newGateway(t, p, http.StatusOK, "thalamus-cli")
	cancelled, cancel := context.WithCancel(testCtx(t))
	for _, c := range []struct {
		name string
		f    *Flow
		ctx  context.Context
		want string
	}{
		{"gateway down", &Flow{Gateway: "http://127.0.0.1:1"}, testCtx(t), "gateway: "},
		{"bad gateway address", &Flow{Gateway: "%zz"}, testCtx(t), "gateway: "},
		{"no identity block", &Flow{Gateway: newGateway(t, p, http.StatusNotFound, "").URL}, testCtx(t), "no identity provider"},
		{"no client_id", &Flow{Gateway: newGateway(t, p, http.StatusOK, "").URL}, testCtx(t), "client_id"},
		{"provider down", &Flow{Gateway: newGateway(t, p, http.StatusOK, "bad-issuer").URL}, testCtx(t), "discovery"},
		{"cannot listen", &Flow{Gateway: okGW.URL, Listen: "127.0.0.1:1"}, testCtx(t), "listen"},
		{"wrong client_id", &Flow{Gateway: newGateway(t, p, http.StatusOK, "other-client").URL, Open: func(u string) error { _ = browser(t, "bad authorize request")(u); cancel(); return nil }}, cancelled, "cancelled"},
	} {
		c.f.Out = &strings.Builder{}
		a, err := c.f.Run(c.ctx)
		if err == nil || a != nil || !strings.Contains(err.Error(), c.want) {
			t.Fatalf("%s: %v, %v", c.name, a, err)
		}
	}
	// The provider refuses the sign-in; the state does not match; the
	// token endpoint fails; no id_token; another nonce; the gateway
	// refuses the token.
	for _, c := range []struct {
		name  string
		setup func()
		open  func(string) error
		want  string
	}{
		{"provider refuses", func() { p.authErr = "access_denied" }, browser(t, "access_denied"), "access_denied"},
		{"state mismatch", func() {}, func(u string) error {
			cb := strings.SplitN(u, "redirect_uri=", 2)[1]
			cb, _ = url.QueryUnescape(strings.SplitN(cb, "&", 2)[0])
			return browser(t, "state does not match")(cb + "?code=x&state=forged")
		}, "state does not match"},
		{"token endpoint fails", func() { p.tokenErr = "invalid_grant" }, browser(t, "Signed in"), "token endpoint: 400 invalid_grant"},
		{"no id_token", func() { p.noToken = true }, browser(t, "Signed in"), "no id_token"},
		{"other nonce", func() { p.wrongN = true }, browser(t, "Signed in"), "nonce"},
		{"gateway refuses", func() { p.clientID = "thalamus-cli" }, browser(t, "Signed in"), "signature"},
	} {
		p.authErr, p.tokenErr, p.noToken, p.wrongN = "", "", false, false
		c.setup()
		gw := okGW
		if c.name == "gateway refuses" {
			gw = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1/login" {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"error":"token: the signature does not verify"}`))
					return
				}
				okGW.Config.Handler.ServeHTTP(w, r)
			}))
			t.Cleanup(gw.Close)
		}
		f := &Flow{Gateway: gw.URL, Out: &strings.Builder{}, Open: c.open}
		a, err := f.Run(testCtx(t))
		if err == nil || a != nil || !strings.Contains(err.Error(), c.want) {
			t.Fatalf("%s: %v, %v", c.name, a, err)
		}
	}
}

// The helpers that read the provider's answers refuse what is not JSON.
func TestBadAnswersAreErrors(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/identity":
			_, _ = w.Write([]byte("not json"))
		case "/.well-known/openid-configuration":
			_, _ = w.Write([]byte(`{"token_endpoint":"x"}`))
		}
	}))
	defer bad.Close()
	f := &Flow{Gateway: bad.URL, Out: &strings.Builder{}}
	if _, err := f.Run(testCtx(t)); err == nil || !strings.Contains(err.Error(), "gateway: ") {
		t.Fatalf("not json: %v", err)
	}
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": bad.URL, "client_id": "c"})
	}))
	defer gw.Close()
	f = &Flow{Gateway: gw.URL, Out: &strings.Builder{}}
	if _, err := f.Run(testCtx(t)); err == nil || !strings.Contains(err.Error(), "authorization_endpoint") {
		t.Fatalf("no endpoints: %v", err)
	}
	// A token endpoint whose address cannot be a request.
	p := newIDP(t)
	odd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"authorization_endpoint": p.srv.URL + "/authorize", "token_endpoint": "%zz"})
	}))
	defer odd.Close()
	gw2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": odd.URL, "client_id": "thalamus-cli"})
	}))
	defer gw2.Close()
	f = &Flow{Gateway: gw2.URL, Out: &strings.Builder{}, Open: browser(t, "Signed in")}
	if _, err := f.Run(testCtx(t)); err == nil || !strings.Contains(err.Error(), "token endpoint: parse") {
		t.Fatalf("bad token address: %v", err)
	}
	// A token whose payload is not base64, or not JSON, has no nonce.
	for _, tok := range []string{"a", "a.!!!.c", "a." + base64.RawURLEncoding.EncodeToString([]byte("x")) + ".c"} {
		if n := nonceOf(tok); n != "" {
			t.Fatalf("%q: nonce %q", tok, n)
		}
	}
}
