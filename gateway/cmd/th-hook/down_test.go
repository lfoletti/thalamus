package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// slowGateway accepts the connection and answers nothing until the hook
// goes away.
func slowGateway(t *testing.T) string {
	t.Helper()
	stop := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-stop:
		}
	}))
	t.Cleanup(ts.Close)
	t.Cleanup(func() { close(stop) }) // runs first: the server closes at once
	return ts.URL
}

// The recall waits 1.5 s by default, for the gateway and the local brains
// together. TH_RECALL_TIMEOUT sets it, as a Go duration. A value that is
// not a positive duration is said on stderr and the default applies: the
// prompt still gets its extracts.
func TestTheRecallWaitsAShortTimeThatCanBeSet(t *testing.T) {
	setting := func(v string) func(string) string {
		return func(k string) string {
			if k == "TH_RECALL_TIMEOUT" {
				return v
			}
			return ""
		}
	}
	if d, err := recallTimeout(setting("")); d != 1500*time.Millisecond || err != nil {
		t.Fatalf("default: %v %v", d, err)
	}
	if d, err := recallTimeout(setting("250ms")); d != 250*time.Millisecond || err != nil {
		t.Fatalf("set: %v %v", d, err)
	}
	for _, bad := range []string{"soon", "0s", "-1s"} {
		if d, err := recallTimeout(setting(bad)); d != 1500*time.Millisecond || err == nil || !strings.Contains(err.Error(), "TH_RECALL_TIMEOUT") {
			t.Fatalf("%q: %v %v", bad, d, err)
		}
	}

	slow := slowGateway(t)
	home := t.TempDir() // no .thalamus folder: the failure is not remembered
	began := time.Now()
	code, out, errOut := hook(t, map[string]string{"HOME": home, "TH_GATEWAY": slow, "TH_KEY": "s3"}, prompt)
	if took := time.Since(began); took < 1300*time.Millisecond || took > 3*time.Second {
		t.Fatalf("default: the hook waited %v", took)
	}
	if code != 0 || out != "" || errOut == "" {
		t.Fatalf("default: code %d, stdout %q, stderr %q", code, out, errOut)
	}
	began = time.Now()
	hook(t, map[string]string{"HOME": home, "TH_GATEWAY": slow, "TH_KEY": "s3", "TH_RECALL_TIMEOUT": "200ms"}, prompt)
	if took := time.Since(began); took > 1200*time.Millisecond {
		t.Fatalf("200ms: the hook waited %v", took)
	}

	g := &fakeGateway{secret: "s3", results: []map[string]any{{"brain": "team/acme", "knowledge_id": "spec-1", "snippet": "The ledger moves on Sunday.", "provenance": "[team/acme] Spec, spec-1"}}}
	code, out, errOut = hook(t, map[string]string{"HOME": home, "TH_GATEWAY": g.start(t), "TH_KEY": "s3", "TH_RECALL_TIMEOUT": "soon"}, prompt)
	if code != 0 || !strings.Contains(contextOf(t, out), "The ledger moves on Sunday.") || !strings.Contains(errOut, `TH_RECALL_TIMEOUT "soon"`) {
		t.Fatalf("bad value: code %d, stdout %q, stderr %q", code, out, errOut)
	}
}

// A gateway that does not answer (refused connection, time out) is
// remembered in gateway-down, next to the env file (the folder of TH_ENV,
// or ~/.thalamus). For 5 minutes the hook does not ask it again: the local
// brains are still served, and the block ends with a line that says why
// the shared brains are missing. Without a local brain, nothing is
// injected. From 5 minutes on, the hook asks again. A gateway that answers
// with a refusal did answer: it is not remembered.
func TestTheHookRemembersAGatewayThatDidNotAnswer(t *testing.T) {
	defer func(f func() time.Time) { now = f }(now)
	at := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	now = func() time.Time { return at }

	dir := t.TempDir()
	envFile := filepath.Join(dir, "env")
	down := filepath.Join(dir, "gateway-down")
	g := &fakeGateway{secret: "s3", results: []map[string]any{{"brain": "team/acme", "knowledge_id": "spec-1", "snippet": "The ledger moves on Sunday.", "provenance": "[team/acme] Spec, spec-1"}}}
	url := g.start(t)
	fake := fakeRecall(t) + " --recall"
	local := brainFile(t, dir, "ok.db", "- Kept memory")
	env := func(gateway, key string, locals bool) map[string]string {
		e := map[string]string{"TH_ENV": envFile, "TH_GATEWAY": gateway, "TH_KEY": key}
		if locals {
			e["TH_LOCAL_RECALL"], e["TH_LOCAL_BRAINS"] = fake, "ok="+local
		}
		return e
	}

	_, out, errOut := hook(t, env("http://127.0.0.1:1", "s3", false), prompt)
	if out != "" || !strings.Contains(errOut, "connect") {
		t.Fatalf("down: stdout %q, stderr %q", out, errOut)
	}
	if data, err := os.ReadFile(down); err != nil || strings.TrimSpace(string(data)) != "2026-10-23T10:00:00Z" {
		t.Fatalf("remembered: %q %v", data, err)
	}

	note := "Shared brains skipped: the gateway did not answer at 10:00:00 UTC; the hook asks it again from 10:05:00 UTC."
	at = at.Add(4*time.Minute + 59*time.Second)
	_, out, errOut = hook(t, env(url, "s3", true), prompt)
	ctx := contextOf(t, out)
	if !strings.Contains(ctx, "[local:ok]\n- Kept memory") || !strings.HasSuffix(ctx, "\n\n"+note) || strings.Contains(ctx, "The ledger") {
		t.Fatalf("within 5 minutes:\n%s", ctx)
	}
	if !strings.Contains(errOut, note) || len(g.calls) != 0 {
		t.Fatalf("within 5 minutes: stderr %q, calls %d", errOut, len(g.calls))
	}
	if _, out, errOut = hook(t, env(url, "s3", false), prompt); out != "" || !strings.Contains(errOut, note) || len(g.calls) != 0 {
		t.Fatalf("within 5 minutes, no local brain: stdout %q, stderr %q, calls %d", out, errOut, len(g.calls))
	}

	at = time.Date(2026, 10, 23, 10, 5, 0, 0, time.UTC)
	_, out, _ = hook(t, env(url, "s3", true), prompt)
	if ctx := contextOf(t, out); !strings.Contains(ctx, "The ledger moves on Sunday.") || !strings.Contains(ctx, "[local:ok]") || strings.Contains(ctx, "Shared brains skipped") || len(g.calls) != 1 {
		t.Fatalf("after 5 minutes (%d calls):\n%s", len(g.calls), ctx)
	}

	if err := os.Remove(down); err != nil {
		t.Fatal(err)
	}
	if _, _, errOut = hook(t, env(url, "wrong", false), prompt); !strings.Contains(errOut, "Unauthorized") {
		t.Fatalf("refused key: %s", errOut)
	}
	if _, err := os.Stat(down); !os.IsNotExist(err) {
		t.Fatalf("a refused key is remembered as a gateway down: %v", err)
	}

	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".thalamus"), 0o700); err != nil {
		t.Fatal(err)
	}
	hook(t, map[string]string{"HOME": home, "TH_GATEWAY": "http://127.0.0.1:1", "TH_KEY": "s3"}, prompt)
	if _, err := os.Stat(filepath.Join(home, ".thalamus", "gateway-down")); err != nil {
		t.Fatalf("without TH_ENV: %v", err)
	}
}
