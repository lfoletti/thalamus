package relay

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// rest sends one request to the bench with a bearer and returns the
// status and the body.
func (b *bench) rest(t *testing.T, method, path, secret, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, b.url+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	data, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(data)
}

// GET /v1/brains lists the brains of the caller, the same list as
// th_brains, under the same key, and leaves no audit line either.
func TestRestListsTheBrains(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	status, body := b.rest(t, http.MethodGet, "/v1/brains", k.Secret, "")
	var got []brainEntry
	if err := json.Unmarshal([]byte(body), &got); status != http.StatusOK || err != nil || len(got) != 2 || got[0].Brain != "team/acme" || got[0].Access != "write" || !got[0].Key || got[1].Brain != "team/other" || got[1].Key {
		t.Fatalf("%d %s", status, body)
	}
	if lines := b.auditLines(t); len(lines) != 0 {
		t.Fatalf("audit: %+v", lines)
	}
	for _, c := range []struct {
		method, path, secret string
		status               int
		want                 string
	}{
		{http.MethodGet, "/v1/brains", "", http.StatusUnauthorized, "required"},
		{http.MethodGet, "/v1/brains", "nope", http.StatusUnauthorized, "unknown key"},
		{http.MethodPost, "/v1/brains", k.Secret, http.StatusMethodNotAllowed, "GET"},
		{http.MethodGet, "/v1/nothing", k.Secret, http.StatusNotFound, "no such"},
		{http.MethodGet, "/v1/", k.Secret, http.StatusNotFound, "no such"},
	} {
		if status, body := b.rest(t, c.method, c.path, c.secret, ""); status != c.status || !strings.Contains(body, c.want) {
			t.Fatalf("%s %s: %d %s", c.method, c.path, status, body)
		}
	}
	if err := os.WriteFile(b.keysFile, []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if status, body := b.rest(t, http.MethodGet, "/v1/brains", k.Secret, ""); status != http.StatusInternalServerError || !strings.Contains(body, "key file") {
		t.Fatalf("broken key file: %d %s", status, body)
	}
}

// POST /v1/add proposes a document the way th_add does: same brain, same
// key, same author and review status, same audit line. Each refusal has
// its status: 400 bad body or no content, 403 no write right, 404 unknown
// brain, 502 the brain refused.
func TestRestAddsADocument(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	status, body := b.rest(t, http.MethodPost, "/v1/add", k.Secret, `{"brain":"team/acme","document_id":"spec-9","title":"Spec","content":"The ledger moves on Sunday.","source_url":"smb://x/spec.docx"}`)
	if status != http.StatusOK || !strings.Contains(body, `"echo"`) {
		t.Fatalf("%d %s", status, body)
	}
	var args map[string]any
	_ = json.Unmarshal([]byte(b.brain.calls[0].ArgsJson), &args)
	if b.brain.calls[0].Name != "ingest_document" || args["collection"] != "team-acme" || args["document_id"] != "spec-9" || args["author"] != "lfoletti" || args["review_status"] != "proposed" {
		t.Fatalf("args: %s", b.brain.calls[0].ArgsJson)
	}
	if l := b.auditLines(t)[0]; l["tool"] != "th_add" || l["ids"].(map[string]any)["document_id"] != "spec-9" || l["ok"] != true {
		t.Fatalf("audit: %+v", l)
	}
	km, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Read)
	for _, c := range []struct {
		secret, body string
		status       int
		want         string
	}{
		{k.Secret, `{"brain":`, http.StatusBadRequest, "body"},
		{k.Secret, `{"brain":"team/acme"}`, http.StatusBadRequest, "content is required"},
		{km.Secret, `{"brain":"team/acme","content":"x"}`, http.StatusForbidden, "may not write"},
		{k.Secret, `{"brain":"org/none","content":"x"}`, http.StatusNotFound, "unknown brain"},
		{"", `{"brain":"team/acme","content":"x"}`, http.StatusUnauthorized, "required"},
	} {
		if status, body := b.rest(t, http.MethodPost, "/v1/add", c.secret, c.body); status != c.status || !strings.Contains(body, c.want) {
			t.Fatalf("%q: %d %s", c.body, status, body)
		}
	}
	if status, body := b.rest(t, http.MethodGet, "/v1/add", k.Secret, ""); status != http.StatusMethodNotAllowed || !strings.Contains(body, "POST") {
		t.Fatalf("GET: %d %s", status, body)
	}
	b.brain.fail = true
	if status, body := b.rest(t, http.MethodPost, "/v1/add", k.Secret, `{"brain":"team/acme","content":"x"}`); status != http.StatusBadGateway || !strings.Contains(body, "brain says no") {
		t.Fatalf("brain failure: %d %s", status, body)
	}
	if len(b.brain.calls) != 2 {
		t.Fatalf("calls: %d", len(b.brain.calls))
	}
}
