// Package login is the user's side of POST /v1/login: it asks the
// gateway which identity provider to use, sends the user to that
// provider in the browser (authorization code with PKCE, loopback
// redirect), exchanges the code for an id_token, and hands the id_token to
// the gateway, which issues the keys.
package login

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// Flow is one sign-in.
type Flow struct {
	// Gateway is the base address of the gateway.
	Gateway string
	// Listen is the loopback address for the redirect; 127.0.0.1:0 by default.
	Listen string
	// Open opens the browser on the sign-in address; nil prints it only.
	Open func(url string) error
	// Out receives the sign-in address and the progress.
	Out io.Writer
	// HTTP talks to the gateway and the provider; http.DefaultClient when nil.
	HTTP *http.Client
}

// Answer is what the gateway returns.
type Answer struct {
	Login  string  `json:"login"`
	Key    string  `json:"key"`
	Brains []Brain `json:"brains"`
}

// Brain is one brain the user may reach, with its key.
type Brain struct {
	Brain  string `json:"brain"`
	Access string `json:"access"`
	Key    string `json:"key"`
}

// Run signs the user in and returns the keys.
func (f *Flow) Run(ctx context.Context) (*Answer, error) {
	gateway := strings.TrimRight(f.Gateway, "/")
	var id struct{ Issuer, ClientID string }
	var idRaw struct {
		Issuer   string `json:"issuer"`
		ClientID string `json:"client_id"`
	}
	if err := f.getJSON(ctx, gateway+"/v1/identity", &idRaw); err != nil {
		return nil, fmt.Errorf("gateway: %w", err)
	}
	id.Issuer, id.ClientID = idRaw.Issuer, idRaw.ClientID
	if id.ClientID == "" {
		return nil, errors.New("gateway: the identity block has no client_id")
	}
	var disc struct {
		Authorize string `json:"authorization_endpoint"`
		Token     string `json:"token_endpoint"`
	}
	if err := f.getJSON(ctx, strings.TrimRight(id.Issuer, "/")+"/.well-known/openid-configuration", &disc); err != nil {
		return nil, fmt.Errorf("issuer discovery: %w", err)
	}
	if disc.Authorize == "" || disc.Token == "" {
		return nil, errors.New("issuer discovery: no authorization_endpoint or token_endpoint")
	}
	verifier, state, nonce := random(), random(), random()
	sum := sha256.Sum256([]byte(verifier))
	listen := f.Listen
	if listen == "" {
		listen = "127.0.0.1:0"
	}
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, fmt.Errorf("listen for the redirect: %w", err)
	}
	defer func() { _ = lis.Close() }()
	redirect := "http://" + lis.Addr().String() + "/callback"
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {id.ClientID},
		"redirect_uri":          {redirect},
		"scope":                 {"openid email"},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
		"code_challenge_method": {"S256"},
	}
	// The callback is served before the browser opens: a browser can come
	// back at once.
	done, stop := serveCallback(lis, state)
	defer stop()
	fmt.Fprintf(f.Out, "Open this address to sign in:\n%s?%s\n", disc.Authorize, q.Encode())
	if f.Open != nil {
		if err := f.Open(disc.Authorize + "?" + q.Encode()); err != nil {
			fmt.Fprintf(f.Out, "cannot open the browser: %v\nOpen the address yourself.\n", err)
		}
	}
	var code string
	select {
	case res := <-done:
		if res.err != nil {
			return nil, res.err
		}
		code = res.code
	case <-ctx.Done():
		return nil, errors.New("sign-in cancelled before the browser came back")
	}
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirect},
		"client_id":     {id.ClientID},
		"code_verifier": {verifier},
	}
	var tok struct {
		IDToken string `json:"id_token"`
	}
	if err := f.postJSON(ctx, disc.Token, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()), "", &tok); err != nil {
		return nil, fmt.Errorf("token endpoint: %w", err)
	}
	if tok.IDToken == "" {
		return nil, errors.New("token endpoint: no id_token in the answer")
	}
	if nonceOf(tok.IDToken) != nonce {
		return nil, errors.New("token endpoint: the id_token does not carry the nonce of this sign-in")
	}
	var a Answer
	if err := f.postJSON(ctx, gateway+"/v1/login", "", nil, tok.IDToken, &a); err != nil {
		return nil, fmt.Errorf("gateway login: %w", err)
	}
	fmt.Fprintf(f.Out, "Signed in as %s.\n", a.Login)
	return &a, nil
}

type callback struct {
	code string
	err  error
}

// serveCallback serves the loopback until the provider redirects the
// browser back with the code; the first callback wins.
func serveCallback(lis net.Listener, state string) (<-chan callback, func()) {
	done := make(chan callback, 1)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var res callback
		switch {
		case q.Get("state") != state:
			res.err = errors.New("callback: the state does not match this sign-in")
		case q.Get("error") != "":
			res.err = fmt.Errorf("the provider refused: %s", q.Get("error"))
		default:
			res.code = q.Get("code")
		}
		if res.err != nil {
			http.Error(w, res.err.Error(), http.StatusBadRequest)
		} else {
			_, _ = io.WriteString(w, "<!doctype html><title>Thalamus</title><p>Signed in. You can close this window.")
		}
		select {
		case done <- res:
		default:
		}
	})}
	go func() { _ = srv.Serve(lis) }()
	return done, func() { _ = srv.Close() }
}

// nonceOf reads the nonce claim of a JWT without checking it: the gateway
// checks the signature, the client checks that the token is the one it
// asked for.
func nonceOf(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct{ Nonce string }
	_ = json.Unmarshal(data, &claims)
	return claims.Nonce
}

func random() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (f *Flow) client() *http.Client {
	if f.HTTP != nil {
		return f.HTTP
	}
	return http.DefaultClient
}

func (f *Flow) getJSON(ctx context.Context, u string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	return f.do(req, out)
}

func (f *Flow) postJSON(ctx context.Context, u, contentType string, body io.Reader, bearer string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	return f.do(req, out)
}

// do sends the request and decodes the JSON answer; a status that is not
// 200 is an error with the body's message.
func (f *Flow) do(req *http.Request, out any) error {
	res, err := f.client().Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	data, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != http.StatusOK {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(data, &e)
		if e.Error == "" {
			e.Error = strings.TrimSpace(string(data))
		}
		return fmt.Errorf("%d %s", res.StatusCode, e.Error)
	}
	return json.Unmarshal(data, out)
}
