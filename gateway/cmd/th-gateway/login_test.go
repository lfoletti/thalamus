package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A fake provider and a fake gateway, enough for the command: the
// provider redirects with a code and returns an unsigned id_token with
// the nonce; the gateway answers /v1/identity and /v1/login.
func fakes(t *testing.T) (gateway string) {
	t.Helper()
	var idp *httptest.Server
	nonces := map[string]string{}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"authorization_endpoint": idp.URL + "/authorize", "token_endpoint": idp.URL + "/token"})
	})
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		nonces["c1"] = q.Get("nonce")
		http.Redirect(w, r, q.Get("redirect_uri")+"?code=c1&state="+url.QueryEscape(q.Get("state")), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		payload, _ := json.Marshal(map[string]string{"nonce": nonces["c1"]})
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": "h." + base64.RawURLEncoding.EncodeToString(payload) + ".s"})
	})
	idp = httptest.NewServer(mux)
	t.Cleanup(idp.Close)
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/identity":
			_ = json.NewEncoder(w).Encode(map[string]string{"issuer": idp.URL, "client_id": "thalamus-cli"})
		case "/v1/login":
			_, _ = w.Write([]byte(`{"login":"lfoletti","key":"k-acme","brains":[{"brain":"team/acme","access":"write","key":"k-acme"},{"brain":"team/other","access":"read","key":"k-other"}]}`))
		}
	}))
	t.Cleanup(gw.Close)
	return gw.URL
}

// visit acts as the browser: it opens the address and follows the
// redirect to the loopback.
func visit(u string) error {
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	return res.Body.Close()
}

// th-gateway login signs in through the browser and writes TH_GATEWAY
// and TH_KEY into the env file, created with its directory, readable by
// the owner only; it prints the brains and how to load the file.
func TestLoginWritesTheEnvFile(t *testing.T) {
	gw := fakes(t)
	dir := t.TempDir()
	envPath := filepath.Join(dir, "deep", ".thalamus", "env")
	saved := openBrowser
	openBrowser = visit
	defer func() { openBrowser = saved }()
	code, out, errOut := call(t, "login", gw+"/", "--env", envPath)
	if code != 0 || errOut != "" {
		t.Fatalf("login: %d %s %s", code, out, errOut)
	}
	for _, want := range []string{"Signed in as lfoletti.", "team/acme (write), team/other (read)", "TH_GATEWAY and TH_KEY written to " + envPath, "set -a; . " + envPath + "; set +a"} {
		if !strings.Contains(out, want) {
			t.Fatalf("out lacks %q:\n%s", want, out)
		}
	}
	data, err := os.ReadFile(envPath)
	if err != nil || string(data) != "TH_GATEWAY="+gw+"\nTH_KEY=k-acme\n" {
		t.Fatalf("env file: %q, %v", data, err)
	}
	if info, _ := os.Stat(envPath); info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode().Perm())
	}
	// Without --env, the file is under the home directory.
	t.Setenv("HOME", dir)
	if code, out, errOut := call(t, "login", gw); code != 0 || !strings.Contains(out, filepath.Join(dir, ".thalamus", "env")) {
		t.Fatalf("default path: %d %s %s", code, out, errOut)
	}
	// --no-browser prints the address and waits: the employee opens it.
	openBrowser = func(string) error { t.Error("--no-browser opened the browser"); return nil }
	var lout lockedBuffer
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			if s := lout.String(); strings.Contains(s, "sign in:\n") {
				u := strings.SplitN(strings.SplitN(s, "sign in:\n", 2)[1], "\n", 2)[0]
				_ = visit(u)
				return
			}
		}
	}()
	if code := run(context.Background(), []string{"login", gw, "--env", envPath, "--no-browser"}, &lout, os.Stderr); code != 0 || !strings.Contains(lout.String(), "Signed in as lfoletti.") {
		t.Fatalf("no browser: %d %s", code, lout.String())
	}
}

// Refusals: no gateway argument, a bad flag, a gateway that is down, an
// env file that cannot be written, a home directory that is unknown.
func TestLoginRefusals(t *testing.T) {
	gw := fakes(t)
	saved := openBrowser
	openBrowser = visit
	defer func() { openBrowser = saved }()
	for _, c := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"login"}, 2, "login needs the gateway"},
		{[]string{"login", gw, "--bogus"}, 2, "bogus"},
		{[]string{"login", "http://127.0.0.1:1", "--env", filepath.Join(t.TempDir(), "env")}, 1, "gateway: "},
		{[]string{"login", gw, "--env", t.TempDir()}, 1, "env file"},
	} {
		code, _, errOut := call(t, c.args...)
		if code != c.code || !strings.Contains(errOut, c.want) {
			t.Fatalf("%v: %d %s", c.args, code, errOut)
		}
	}
	// The env file's parent is a file: it cannot be created.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "parent"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut := call(t, "login", gw, "--env", filepath.Join(dir, "parent", "env")); code != 1 || !strings.Contains(errOut, "env file") {
		t.Fatalf("unwritable: %d %s", code, errOut)
	}
	t.Setenv("HOME", "")
	if code, _, errOut := call(t, "login", gw); code != 1 || !strings.Contains(errOut, "home directory") {
		t.Fatalf("no home: %d %s", code, errOut)
	}
}

// openBrowser runs the TH_BROWSER command, or the system opener.
func TestOpenBrowserRunsTheOpener(t *testing.T) {
	t.Setenv("TH_BROWSER", "true")
	if err := openBrowser("http://127.0.0.1:1/x"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TH_BROWSER", "/nonexistent/opener")
	if err := openBrowser("http://127.0.0.1:1/x"); err == nil {
		t.Fatal("a missing opener is not an error")
	}
	t.Setenv("TH_BROWSER", "")
	if a, b := openerName("darwin"), openerName("linux"); a != "open" || b != "xdg-open" {
		t.Fatalf("openers %q %q", a, b)
	}
}
