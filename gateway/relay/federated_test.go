package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// recallResult is what th_recall returns.
type recallOut struct {
	Results []struct {
		Brain       string   `json:"brain"`
		KnowledgeID string   `json:"knowledge_id"`
		Title       string   `json:"title"`
		Snippet     string   `json:"snippet"`
		Score       float64  `json:"score"`
		Provenance  string   `json:"provenance"`
		AlsoIn      []string `json:"also_in,omitempty"`
		Level       string   `json:"level,omitempty"`
	} `json:"results"`
	Skipped []struct {
		Brain  string `json:"brain"`
		Reason string `json:"reason"`
	} `json:"skipped,omitempty"`
}

func (b *bench) federated(t *testing.T, token string) *mcp.ClientSession {
	t.Helper()
	tr := &mcp.StreamableClientTransport{Endpoint: b.url + "/mcp", HTTPClient: &http.Client{Transport: bearer{token}}, MaxRetries: -1}
	s, err := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(context.Background(), tr, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func decode[T any](t *testing.T, res *mcp.CallToolResult) T {
	t.Helper()
	var out T
	if res.IsError {
		t.Fatalf("tool error: %s", res.Content[0].(*mcp.TextContent).Text)
	}
	if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// th_brains lists the caller's brains with the access and whether a key
// was issued; the caller is known from any one of their gateway keys. It
// reads no brain, so it leaves no audit line.
func TestGcBrainsListsTheCallersBrains(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	tools, err := s.ListTools(context.Background(), nil)
	if err != nil || len(tools.Tools) != 7 || tools.Tools[0].Name != "th_add" || tools.Tools[1].Name != "th_brains" || tools.Tools[2].Name != "th_descend" || tools.Tools[3].Name != "th_get" || tools.Tools[4].Name != "th_pending" || tools.Tools[5].Name != "th_recall" || tools.Tools[6].Name != "th_review" {
		t.Fatalf("tools: %+v %v", tools, err)
	}
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_brains"})
	if err != nil {
		t.Fatal(err)
	}
	got := decode[[]struct {
		Brain  string `json:"brain"`
		Access string `json:"access"`
		Key    bool   `json:"key"`
	}](t, res)
	if len(got) != 2 || got[0].Brain != "team/acme" || got[0].Access != "write" || !got[0].Key || got[1].Brain != "team/other" || got[1].Access != "write" || got[1].Key {
		t.Fatalf("got %+v", got)
	}
	// th_brains reads no brain: the audit keeps nothing of it.
	if lines := b.auditLines(t); len(lines) != 0 {
		t.Fatalf("audit: %+v", lines)
	}
}

// th_recall asks every readable brain that has a key, with that key, and
// fuses the lists by reciprocal rank, weighted by the brain; a brain without
// a key or that fails is listed as skipped; the same text in two brains is
// one result that names both; the audit line lists brains and ids read.
func TestGcRecallFusesTheBrains(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	other, _ := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	b.brain.hits = map[string][]fakeHit{
		"team-acme": {{"a-1", "Ledger plan", "The ledger moves on Sunday."}, {"a-2", "Rollback", "Two hours of rollback."}},
		"other":     {{"o-1", "Copy", "The ledger moves on Sunday."}, {"o-2", "Other", "Something else."}},
	}
	s := b.federated(t, k.Secret)
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: map[string]any{"query": "ledger", "top_k": 3}})
	if err != nil {
		t.Fatal(err)
	}
	got := decode[recallOut](t, res)
	if len(got.Results) != 3 || got.Results[0].KnowledgeID != "a-1" || got.Results[0].Brain != "team/acme" || strings.Join(got.Results[0].AlsoIn, ",") != "team/other" {
		t.Fatalf("results: %+v", got.Results)
	}
	// team/other has weight 0.5: its second hit ranks under team/acme's second.
	if got.Results[1].KnowledgeID != "a-2" || got.Results[2].KnowledgeID != "o-2" || got.Results[0].Score <= got.Results[1].Score || got.Results[1].Score <= got.Results[2].Score {
		t.Fatalf("order: %+v", got.Results)
	}
	if p := got.Results[0].Provenance; p != "[team/acme] Ledger plan, a-1" {
		t.Fatalf("provenance %q", p)
	}
	if len(got.Skipped) != 0 {
		t.Fatalf("skipped: %+v", got.Skipped)
	}
	// Each brain saw its own key and its own collection.
	seen := map[string]string{}
	for _, c := range b.brain.calls {
		var args map[string]any
		_ = json.Unmarshal([]byte(c.ArgsJson), &args)
		seen[args["collection"].(string)] = c.Name
	}
	if len(seen) != 2 || seen["team-acme"] != "knowledge_search" || seen["other"] != "knowledge_search" {
		t.Fatalf("calls: %+v", b.brain.calls)
	}
	if b.brain.bearers["team-acme"] != "Bearer "+k.Secret || b.brain.bearers["other"] != "Bearer "+other.Secret {
		t.Fatalf("bearers: %+v", b.brain.bearers)
	}
	l := b.auditLines(t)[0]
	ids := l["ids"].(map[string]any)
	if l["tool"] != "th_recall" || l["ok"] != true || len(ids["brains"].([]any)) != 2 || len(ids["knowledge_ids"].([]any)) != 4 {
		t.Fatalf("audit: %+v", l)
	}

	// Without a key on team/other, it is skipped and said so.
	if err := keys.Revoke(b.keysFile, "lfoletti", "team/other"); err != nil {
		t.Fatal(err)
	}
	res, err = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: map[string]any{"query": "ledger"}})
	if err != nil {
		t.Fatal(err)
	}
	got = decode[recallOut](t, res)
	if len(got.Results) != 2 || len(got.Skipped) != 1 || got.Skipped[0].Brain != "team/other" || !strings.Contains(got.Skipped[0].Reason, "no key") {
		t.Fatalf("got %+v", got)
	}

	// A brain that fails is skipped with the error, the others answer.
	k2, _ := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	_ = k2
	b.brain.failOn = "other"
	res, err = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: map[string]any{"query": "ledger"}})
	if err != nil {
		t.Fatal(err)
	}
	got = decode[recallOut](t, res)
	if len(got.Results) != 2 || len(got.Skipped) != 1 || !strings.Contains(got.Skipped[0].Reason, "brain says no") {
		t.Fatalf("got %+v", got)
	}
	// A query is required; a brain the caller names but may not read is refused.
	res, _ = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: map[string]any{"query": ""}})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "query") {
		t.Fatalf("empty query: %+v", res)
	}
	res, _ = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: map[string]any{"query": "x", "brains": []string{"org/none"}}})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "org/none") {
		t.Fatalf("unknown brain: %+v", res)
	}
	res, _ = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: json.RawMessage(`{"query": 5}`)})
	if !res.IsError {
		t.Fatalf("bad arguments: %+v", res)
	}
}

// th_recall restricted to named brains asks only those; a brain that is
// the whole server is asked without a collection.
func TestGcRecallCanBeRestrictedToSomeBrains(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	_, _ = keys.Issue(b.keysFile, "lfoletti", "team/other", "", rights.Write)
	whole := b.rights.Brains["team/other"]
	whole.Collection = ""
	b.rights.Brains["team/other"] = whole
	b.brain.hits = map[string][]fakeHit{"team-acme": {{"a-1", "A", "a"}}, "": {{"o-1", "O", "o"}}}
	s := b.federated(t, k.Secret)
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: map[string]any{"query": "x", "brains": []string{"team/other"}}})
	if err != nil {
		t.Fatal(err)
	}
	got := decode[recallOut](t, res)
	if len(got.Results) != 1 || got.Results[0].Brain != "team/other" || len(b.brain.calls) != 1 || strings.Contains(b.brain.calls[0].ArgsJson, "collection") {
		t.Fatalf("got %+v, calls %+v", got, b.brain.calls)
	}
}

// th_recall with skip_brains leaves those brains out, after brains: a
// brain left out is not asked. A name the caller may not read, or that does
// not exist, is ignored: a filter left over from an old right is no error.
func TestGcRecallLeavesOutTheSkippedBrains(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	_, _ = keys.Issue(b.keysFile, "lfoletti", "team/other", "team-other", rights.Write)
	b.brain.hits = map[string][]fakeHit{"team-acme": {{"a-1", "A", "a"}}, "team-other": {{"o-1", "O", "o"}}}
	s := b.federated(t, k.Secret)
	for _, args := range []map[string]any{
		{"query": "x", "skip_brains": []string{"team/other", "org/gone"}},
		{"query": "x", "brains": []string{"team/acme", "team/other"}, "skip_brains": []string{"team/other"}},
	} {
		b.brain.calls = nil
		res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		got := decode[recallOut](t, res)
		if len(got.Results) != 1 || got.Results[0].Brain != "team/acme" || len(b.brain.calls) != 1 {
			t.Fatalf("%v: got %+v, calls %+v", args, got, b.brain.calls)
		}
	}
}

// Every brain can be skipped: no brain reachable, a brain whose address is
// wrong, a brain whose answer is not JSON. The skips are sorted by brain.
func TestGcRecallReportsEverySkippedBrain(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	_, _ = keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	s := b.federated(t, k.Secret)
	b.brain.fail = true
	got := decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "x"}))
	if len(got.Results) != 0 || len(got.Skipped) != 2 || got.Skipped[0].Brain != "team/acme" || got.Skipped[1].Brain != "team/other" {
		t.Fatalf("got %+v", got)
	}
	b.brain.fail = false
	other := b.rights.Brains["team/other"]
	other.Addr = "%zz"
	b.rights.Brains["team/other"] = other
	got = decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "x"}))
	if len(got.Skipped) != 1 || got.Skipped[0].Brain != "team/other" || !strings.Contains(got.Skipped[0].Reason, "connect to %zz") {
		t.Fatalf("got %+v", got)
	}
	b.brain.badJSON = true
	got = decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "x", "brains": []string{"team/acme"}}))
	if len(got.Skipped) != 1 || !strings.Contains(got.Skipped[0].Reason, "decode knowledge_search") {
		t.Fatalf("got %+v", got)
	}
}

func mustCall(t *testing.T, s *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// The provenance line names the brain, the title or the id, the source and
// the author, each only when known.
func TestProvenanceNamesWhatIsKnown(t *testing.T) {
	for _, c := range []struct {
		h    hit
		want string
	}{
		{hit{KnowledgeID: "d-1"}, "[b] d-1"},
		{hit{KnowledgeID: "d-1", Title: "Spec"}, "[b] Spec, d-1"},
		{hit{KnowledgeID: "d-1", Title: "Spec", SourceURL: "smb://x/spec.docx", Author: "mdupont"}, "[b] Spec, d-1, smb://x/spec.docx, by mdupont"},
	} {
		if got := provenance("b", c.h, ""); got != c.want {
			t.Fatalf("got %q, want %q", got, c.want)
		}
	}
}

// The federated endpoint admits any key the gateway issued, and refuses the
// rest before anything is called: an unknown secret, an operator key, no
// token, a holder who lost every right.
func TestTheFederatedEndpointAdmitsOnlyGatewayKeys(t *testing.T) {
	b := start(t)
	ghost, _ := keys.Issue(b.keysFile, "ghost", "team/acme", "team-acme", rights.Write)
	opPath := filepath.Join(filepath.Dir(b.keysFile), "keys.json")
	data, _ := os.ReadFile(opPath)
	_ = os.WriteFile(opPath, []byte(strings.Replace(string(data), `"keys": [`, `"keys": [{"id":"operator","secret":"op-secret","clearance":"read-write"},`, 1)), 0o600)
	for _, c := range []struct {
		name, token string
		want        int
	}{
		{"unknown", "nope", http.StatusUnauthorized},
		{"operator", "op-secret", http.StatusUnauthorized},
		{"none", "", http.StatusUnauthorized},
		{"no right left", ghost.Secret, http.StatusForbidden},
	} {
		req, _ := http.NewRequest(http.MethodPost, b.url+"/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != c.want {
			t.Fatalf("%s: got %d, want %d", c.name, resp.StatusCode, c.want)
		}
	}
	if len(b.brain.calls) != 0 {
		t.Fatalf("the brain was reached: %+v", b.brain.calls)
	}
	// A key file that cannot be read is a server error, not a refusal.
	_ = os.WriteFile(b.keysFile, []byte("{"), 0o600)
	req, _ := http.NewRequest(http.MethodPost, b.url+"/mcp", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer x")
	resp, _ := http.DefaultClient.Do(req)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("broken key file: %d", resp.StatusCode)
	}
}

// The RRF fusion on its own: rank 1 in a brain of weight 1 scores 1/61,
// the same text in two brains adds up, and ties keep the input order.
func TestFuseByReciprocalRank(t *testing.T) {
	lists := []brainHits{
		{Brain: "a", Weight: 1, Hits: []hit{{KnowledgeID: "a-1", Title: "T", Snippet: "same"}, {KnowledgeID: "a-2", Title: "T", Snippet: "x"}}},
		{Brain: "b", Weight: 0.5, Hits: []hit{{KnowledgeID: "b-1", Title: "T", Snippet: "same"}, {KnowledgeID: "b-2", Title: "T", Snippet: "y"}}},
	}
	out := fuse(lists, 10, "")
	if len(out) != 3 || out[0].KnowledgeID != "a-1" || strings.Join(out[0].AlsoIn, ",") != "b" {
		t.Fatalf("%+v", out)
	}
	if want := 1.0/61 + 0.5/61; out[0].Score < want-1e-9 || out[0].Score > want+1e-9 {
		t.Fatalf("score %v", out[0].Score)
	}
	if out[1].KnowledgeID != "a-2" || out[2].KnowledgeID != "b-2" {
		t.Fatalf("%+v", out)
	}
	if got := fuse(lists, 1, ""); len(got) != 1 {
		t.Fatalf("top_k: %+v", got)
	}
	// Rank 1 of the second brain comes before rank 2 of the first: the
	// lists are interleaved by score, not concatenated.
	equal := []brainHits{
		{Brain: "a", Weight: 1, Hits: []hit{{KnowledgeID: "a-1", Snippet: "1"}, {KnowledgeID: "a-2", Snippet: "2"}}},
		{Brain: "b", Weight: 1, Hits: []hit{{KnowledgeID: "b-1", Snippet: "3"}, {KnowledgeID: "b-2", Snippet: "4"}}},
	}
	var ids []string
	for _, f := range fuse(equal, 10, "") {
		ids = append(ids, f.KnowledgeID)
	}
	if strings.Join(ids, ",") != "a-1,b-1,a-2,b-2" {
		t.Fatalf("order %v", ids)
	}
}

func withCurator(t *testing.T, b *bench) {
	t.Helper()
	acme := b.rights.Brains["team/acme"]
	acme.Curators = []string{"user:kmlead"}
	acme.Writers = append(acme.Writers, "user:mdupont")
	b.rights.Brains["team/acme"] = acme
}

// th_add proposes a document into a brain the caller may write: after
// reading the id (free here), the brain receives ingest_document with the caller's key, in its collection,
// with the caller as author and the explicit field review_status proposed,
// never a metadata key. A curator's own document is accepted at once. The
// audit line names the document.
func TestGcAddProposesADocumentForReview(t *testing.T) {
	b := start(t)
	withCurator(t, b)
	b.brain.missing = true
	k, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "spec-2", "title": "Spec", "content": "The ledger moves on Sunday.", "source_url": "smb://x/spec.docx", "metadata": map[string]any{"classification": "internal"}})
	if res.IsError {
		t.Fatalf("th_add: %s", res.Content[0].(*mcp.TextContent).Text)
	}
	var args map[string]any
	_ = json.Unmarshal([]byte(b.brain.calls[1].ArgsJson), &args)
	md, _ := args["metadata"].(map[string]any)
	if b.brain.calls[1].Name != "ingest_document" || args["collection"] != "team-acme" || args["document_id"] != "spec-2" || args["title"] != "Spec" || args["author"] != "mdupont" || args["source_url"] != "smb://x/spec.docx" || args["review_status"] != "proposed" || args["reviewed_by"] != nil || md["classification"] != "internal" || len(md) != 1 {
		t.Fatalf("args: %s", b.brain.calls[1].ArgsJson)
	}
	if b.brain.bearers["team-acme"] != "Bearer "+k.Secret {
		t.Fatalf("bearer %q", b.brain.bearers["team-acme"])
	}
	l := b.auditLines(t)[0]
	ids := l["ids"].(map[string]any)
	if l["tool"] != "th_add" || l["brain"] != "team/acme" || l["user"] != "mdupont" || ids["document_id"] != "spec-2" || l["ok"] != true {
		t.Fatalf("audit: %+v", l)
	}
	// A curator's document is accepted at once, by the curator.
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	sc := b.federated(t, kc.Secret)
	if res := mustCall(t, sc, "th_add", map[string]any{"brain": "team/acme", "content": "A curated note."}); res.IsError {
		t.Fatalf("curator add: %s", res.Content[0].(*mcp.TextContent).Text)
	}
	_ = json.Unmarshal([]byte(b.brain.calls[2].ArgsJson), &args)
	md, _ = args["metadata"].(map[string]any)
	if args["review_status"] != "accepted" || args["reviewed_by"] != nil || len(md) != 0 || args["document_id"] == nil || args["document_id"] == "" {
		t.Fatalf("curator args: %s", b.brain.calls[2].ArgsJson)
	}
}

// th_add is refused without a write right, without content, on an
// unknown brain, and when the brain refuses; nothing reaches the brain in
// the first three cases, and the audit says the fourth failed.
func TestGcAddIsRefusedWithoutAWriteRight(t *testing.T) {
	b := start(t)
	withCurator(t, b)
	acme := b.rights.Brains["team/acme"]
	acme.Readers = []string{"user:guest"}
	b.rights.Brains["team/acme"] = acme
	k, _ := keys.Issue(b.keysFile, "guest", "team/acme", "team-acme", rights.Read)
	s := b.federated(t, k.Secret)
	for _, c := range []struct {
		args map[string]any
		want string
	}{
		{map[string]any{"brain": "team/acme", "content": "x"}, "may not write"},
		{map[string]any{"brain": "org/none", "content": "x"}, "unknown brain"},
		{map[string]any{"brain": "team/acme"}, "content is required"},
	} {
		res := mustCall(t, s, "th_add", c.args)
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, c.want) {
			t.Fatalf("%v: %+v", c.args, res)
		}
	}
	if len(b.brain.calls) != 0 {
		t.Fatalf("the brain was reached: %+v", b.brain.calls)
	}
	kw, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Write)
	sw := b.federated(t, kw.Secret)
	b.brain.fail = true
	res := mustCall(t, sw, "th_add", map[string]any{"brain": "team/acme", "content": "x"})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "brain says no") {
		t.Fatalf("brain failure: %+v", res)
	}
	if l := b.auditLines(t)[0]; l["tool"] != "th_add" || l["ok"] != false {
		t.Fatalf("audit: %+v", l)
	}
	res, _ = sw.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_add", Arguments: json.RawMessage(`{"brain": 5}`)})
	if !res.IsError {
		t.Fatal("bad arguments passed")
	}
}

// th_pending lists what waits for review in a brain the caller curates:
// aggregate_metadata groups the chunks by document under a filter on the
// engine's key _review_status (which lifts the review gate for that status),
// then knowledge_get reads each, with the curator's key. th_review decides:
// knowledge_update with the explicit decision fields and the curator as
// reviewer, and no metadata, so the stored metadata is kept.
// Both are refused to a writer who is not a curator.
func TestGcPendingAndGcReviewAreForCurators(t *testing.T) {
	b := start(t)
	withCurator(t, b)
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	long := strings.Repeat("The ledger moves on Sunday. ", 12)
	b.brain.pending = map[string][]fakeHit{"team-acme": {{"spec-2", "Spec", long}, {"spec-3", "", "Short."}}}
	sc := b.federated(t, kc.Secret)
	type listing struct {
		Brain   string       `json:"brain"`
		Results []pendingHit `json:"results"`
	}
	got := decode[listing](t, mustCall(t, sc, "th_pending", map[string]any{"brain": "team/acme", "top_k": 1}))
	if got.Brain != "team/acme" || len(got.Results) != 1 || got.Results[0].KnowledgeID != "spec-2" || got.Results[0].Title != "Spec" || got.Results[0].Author != "mdupont" || !strings.HasSuffix(got.Results[0].Snippet, " The…") || len(got.Results[0].Snippet) > 203 {
		t.Fatalf("pending: %+v", got)
	}
	var args map[string]any
	_ = json.Unmarshal([]byte(b.brain.calls[0].ArgsJson), &args)
	filters, _ := args["filters"].(map[string]any)
	if b.brain.calls[0].Name != "aggregate_metadata" || args["collection"] != "team-acme" || args["type"] != "group_by" || fmt.Sprint(args["group_by"]) != "[document_id]" || filters["_review_status"] != "proposed" || args["limit"] != float64(1) {
		t.Fatalf("pending args: %s", b.brain.calls[0].ArgsJson)
	}
	args = nil // Unmarshal into a map keeps the old keys.
	_ = json.Unmarshal([]byte(b.brain.calls[1].ArgsJson), &args)
	if _, has := args["collection"]; b.brain.calls[1].Name != "knowledge_get" || args["knowledge_id"] != "spec-2" || has {
		t.Fatalf("get args: %s", b.brain.calls[1].ArgsJson)
	}
	// The default is 20: both documents come, the short one whole.
	if got = decode[listing](t, mustCall(t, sc, "th_pending", map[string]any{"brain": "team/acme"})); len(got.Results) != 2 || got.Results[1].Snippet != "Short." {
		t.Fatalf("pending default: %+v", got)
	}
	if l := b.auditLines(t)[1]; l["tool"] != "th_pending" || l["ok"] != true || fmt.Sprint(l["ids"].(map[string]any)["knowledge_ids"]) != "[spec-2 spec-3]" {
		t.Fatalf("audit: %+v", l)
	}
	b.brain.calls = nil
	res := mustCall(t, sc, "th_review", map[string]any{"brain": "team/acme", "knowledge_id": "spec-2", "decision": "accepted"})
	if res.IsError {
		t.Fatalf("review: %s", res.Content[0].(*mcp.TextContent).Text)
	}
	_ = json.Unmarshal([]byte(b.brain.calls[0].ArgsJson), &args)
	if _, md := args["metadata"]; b.brain.calls[0].Name != "knowledge_update" || args["knowledge_id"] != "spec-2" || args["collection"] != "team-acme" || args["review_status"] != "accepted" || args["reviewed_by"] != nil || md {
		t.Fatalf("review args: %s", b.brain.calls[0].ArgsJson)
	}
	if l := b.auditLines(t)[2]; l["tool"] != "th_review" || l["ids"].(map[string]any)["knowledge_id"] != "spec-2" || l["ids"].(map[string]any)["decision"] != "accepted" {
		t.Fatalf("audit: %+v", l)
	}
	for _, c := range []struct {
		args map[string]any
		want string
	}{
		{map[string]any{"brain": "team/acme", "knowledge_id": "spec-2", "decision": "maybe"}, "decision must be accepted or rejected"},
		{map[string]any{"brain": "team/acme", "decision": "accepted"}, "knowledge_id is required"},
		{map[string]any{"brain": "org/none", "knowledge_id": "x", "decision": "accepted"}, "unknown brain"},
	} {
		res := mustCall(t, sc, "th_review", c.args)
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, c.want) {
			t.Fatalf("%v: %+v", c.args, res)
		}
	}
	if res := mustCall(t, sc, "th_pending", map[string]any{"brain": "org/none"}); !res.IsError {
		t.Fatal("pending on an unknown brain passed")
	}
	kw, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Write)
	sw := b.federated(t, kw.Secret)
	for _, name := range []string{"th_pending", "th_review"} {
		res := mustCall(t, sw, name, map[string]any{"brain": "team/acme", "knowledge_id": "spec-2", "decision": "accepted"})
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "not a curator") {
			t.Fatalf("%s by a writer: %+v", name, res)
		}
	}
	b.brain.fail = true
	for _, name := range []string{"th_pending", "th_review"} {
		res := mustCall(t, sc, name, map[string]any{"brain": "team/acme", "knowledge_id": "spec-2", "decision": "rejected"})
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "brain says no") {
			t.Fatalf("%s failure: %+v", name, res)
		}
	}
	if l := b.auditLines(t)[3]; l["tool"] != "th_pending" || l["ok"] != false {
		t.Fatalf("failed pending not audited: %+v", l)
	}
	b.brain.fail = false
	// A failure or bad JSON on either step of the listing is reported.
	for _, c := range []struct{ failTool, badTool, want string }{
		{"knowledge_get", "", "brain says no to knowledge_get"},
		{"", "aggregate_metadata", "decode aggregate_metadata"},
		{"", "knowledge_get", "decode knowledge_get"},
	} {
		b.brain.failTool, b.brain.badTool = c.failTool, c.badTool
		res := mustCall(t, sc, "th_pending", map[string]any{"brain": "team/acme"})
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, c.want) {
			t.Fatalf("%+v: %+v", c, res)
		}
	}
	b.brain.failTool, b.brain.badTool = "", ""
	for _, name := range []string{"th_pending", "th_review"} {
		res, _ := sc.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: json.RawMessage(`{"brain": 5}`)})
		if !res.IsError {
			t.Fatalf("%s bad arguments passed", name)
		}
	}
	if len(b.brain.calls) != 8 {
		t.Fatalf("calls: %d", len(b.brain.calls))
	}
}

func TestSnippetCutsOnAWord(t *testing.T) {
	if got := snippet(strings.Repeat("a", 250)); got != strings.Repeat("a", 200)+"…" {
		t.Fatalf("no space: %q", got)
	}
}

// th_get reads one document whole, in a brain the caller may read, with the
// caller's key: knowledge_get without a collection (the brain's row guard
// checks it). Refused on an unknown brain, without a read right, and when
// the brain refuses; audited with the knowledge_id.
func TestThGetReadsADocumentWhole(t *testing.T) {
	b := start(t)
	b.brain.pending = map[string][]fakeHit{"team-acme": {{"cutover-plan", "Cutover plan", "The code word is MARMOT."}}}
	k, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Read)
	s := b.federated(t, k.Secret)
	res := mustCall(t, s, "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "cutover-plan"})
	if res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "MARMOT") {
		t.Fatalf("th_get: %+v", res)
	}
	var args map[string]any
	_ = json.Unmarshal([]byte(b.brain.calls[0].ArgsJson), &args)
	if _, has := args["collection"]; b.brain.calls[0].Name != "knowledge_get" || args["knowledge_id"] != "cutover-plan" || has || b.brain.bearers[""] != "Bearer "+k.Secret {
		t.Fatalf("call: %+v", b.brain.calls[0])
	}
	if l := b.auditLines(t)[0]; l["tool"] != "th_get" || l["brain"] != "team/acme" || l["ids"].(map[string]any)["knowledge_id"] != "cutover-plan" || l["ok"] != true {
		t.Fatalf("audit: %+v", l)
	}
	for _, c := range []struct {
		args map[string]any
		want string
	}{
		{map[string]any{"brain": "org/none", "knowledge_id": "x"}, "unknown brain"},
		{map[string]any{"brain": "team/other", "knowledge_id": "x"}, "may not read"},
		{map[string]any{"brain": "team/acme"}, "knowledge_id is required"},
	} {
		res := mustCall(t, s, "th_get", c.args)
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, c.want) {
			t.Fatalf("%v: %+v", c.args, res)
		}
	}
	b.brain.fail = true
	if res := mustCall(t, s, "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "cutover-plan"}); !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "brain says no") {
		t.Fatalf("failure: %+v", res)
	}
	if l := b.auditLines(t)[1]; l["tool"] != "th_get" || l["ok"] != false {
		t.Fatalf("failed get not audited: %+v", l)
	}
	if res, _ := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_get", Arguments: json.RawMessage(`{"brain": 5}`)}); !res.IsError {
		t.Fatal("bad arguments passed")
	}
	if len(b.brain.calls) != 2 {
		t.Fatalf("calls: %d", len(b.brain.calls))
	}
}

// pyramidBench is the bench with a pyramid on team/acme, its syntheses
// listing their sources under "preuves", and a key for lfoletti on both
// brains.
func pyramidBench(t *testing.T) (*bench, string) {
	t.Helper()
	b := startWith(t, func(y string) string {
		return strings.Replace(y, "    collection: team-acme\n", "    collection: team-acme\n    pyramid: {collection: acme-pyramid, sources: preuves}\n", 1)
	})
	k, err := keys.IssueScoped(b.keysFile, "lfoletti", "team/acme", []string{"team-acme", "acme-pyramid"}, rights.Write)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write); err != nil {
		t.Fatal(err)
	}
	return b, k.Secret
}

// th_recall serves the syntheses of a brain's pyramid first, marked as
// such, then the passages: the pyramid is searched in its own collection,
// with the same key, for at most `syntheses` results (4 by default, 0 for
// none). A brain without a pyramid is not searched twice. A pyramid that
// fails is said in skipped, and the passages stay.
func TestRecallServesTheSynthesesFirst(t *testing.T) {
	b, secret := pyramidBench(t)
	b.brain.hits = map[string][]fakeHit{
		"team-acme":    {{"a-1", "Ledger plan", "The ledger moves on Sunday."}, {"a-2", "Rollback", "Two hours of rollback."}},
		"acme-pyramid": {{"pyramide/ledger", "Ledger", "The ledger moves on Sunday, with two hours of rollback."}, {"pyramide/cutover", "Cutover", "The cutover."}, {"pyramide/third", "Third", "A third one."}},
		"other":        {{"o-1", "Other", "Something else."}},
	}
	s := b.federated(t, secret)
	got := decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "ledger", "top_k": 2}))
	if len(got.Results) != 5 || got.Results[0].KnowledgeID != "pyramide/ledger" || got.Results[1].KnowledgeID != "pyramide/cutover" || got.Results[3].KnowledgeID != "a-1" {
		t.Fatalf("order: %+v", got.Results)
	}
	if got.Results[0].Level != "synthesis" || got.Results[3].Level != "" || got.Results[0].Provenance != "[team/acme] synthesis: Ledger, pyramide/ledger" {
		t.Fatalf("level: %+v", got.Results[0])
	}
	asked := map[string]float64{}
	for _, c := range b.brain.calls {
		var args map[string]any
		_ = json.Unmarshal([]byte(c.ArgsJson), &args)
		asked[args["collection"].(string)] = args["top_k"].(float64)
	}
	if len(asked) != 3 || asked["acme-pyramid"] != 4 || asked["team-acme"] != 2 || b.brain.bearers["acme-pyramid"] != "Bearer "+secret {
		t.Fatalf("calls: %+v", b.brain.calls)
	}
	ids := b.auditLines(t)[0]["ids"].(map[string]any)["knowledge_ids"].([]any)
	if fmt.Sprint(ids) != "[team/acme:a-1 team/acme:a-2 team/acme:pyramide/cutover team/acme:pyramide/ledger team/acme:pyramide/third team/other:o-1]" {
		t.Fatalf("audit ids: %v", ids)
	}

	got = decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "ledger", "top_k": 1, "syntheses": 3}))
	if len(got.Results) != 4 || got.Results[2].KnowledgeID != "pyramide/third" || got.Results[3].Level != "" {
		t.Fatalf("three syntheses: %+v", got.Results)
	}
	calls := len(b.brain.calls)
	got = decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "ledger", "top_k": 1, "syntheses": 0}))
	if len(got.Results) != 1 || got.Results[0].Level != "" || len(b.brain.calls) != calls+2 {
		t.Fatalf("no synthesis: %+v, %d calls", got.Results, len(b.brain.calls)-calls)
	}

	b.brain.failOn = "acme-pyramid"
	got = decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "ledger", "top_k": 2}))
	if len(got.Results) != 2 || got.Results[0].KnowledgeID != "a-1" || len(got.Skipped) != 1 || got.Skipped[0].Brain != "team/acme" || !strings.HasPrefix(got.Skipped[0].Reason, "pyramid: ") {
		t.Fatalf("failed pyramid: %+v", got)
	}
}

// th_recall asks each brain for the chunks of its documents, in graph mode,
// top_k chunks in at most 1 000 characters each (10 chunks and 4 syntheses by
// default): a chunk is a result, with the id, title and source of its
// document from the same answer, or its document id alone. A brain that
// answers no chunk is served its documents' snippets. A document read
// through two chunks is audited once.
func TestRecallServesTheChunksOfTheDocuments(t *testing.T) {
	b, secret := pyramidBench(t)
	b.brain.hits = map[string][]fakeHit{
		"team-acme":    {{"a-1", "Ledger plan", "The ledger..."}},
		"acme-pyramid": {{"pyramide/ledger", "Ledger", "The ledger moves..."}},
		"other":        {{"o-1", "Other", "Something else."}},
	}
	b.brain.chunks = map[string][]fakeChunk{
		"team-acme":    {{"a-1", "The ledger moves on Sunday night."}, {"a-1", "The rollback window is two hours."}, {"zz", "A chunk of a document the answer does not list."}, {"a-1", "A fourth chunk."}},
		"acme-pyramid": {{"pyramide/ledger", "The ledger moves on Sunday, with two hours of rollback."}},
	}
	s := b.federated(t, secret)
	got := decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "ledger", "top_k": 3}))
	var lines []string
	for _, r := range got.Results {
		lines = append(lines, r.Level+"|"+r.Provenance+"|"+r.Snippet)
	}
	want := []string{
		"synthesis|[team/acme] synthesis: Ledger, pyramide/ledger, https://wiki/pyramide/ledger, by mdupont|The ledger moves on Sunday, with two hours of rollback.",
		"|[team/acme] Ledger plan, a-1, https://wiki/a-1, by mdupont|The ledger moves on Sunday night.",
		"|[team/acme] Ledger plan, a-1, https://wiki/a-1, by mdupont|The rollback window is two hours.",
		"|[team/acme] zz|A chunk of a document the answer does not list.",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("results:\n%s\nskipped: %+v", strings.Join(lines, "\n"), got.Skipped)
	}
	for _, c := range b.brain.calls {
		var args map[string]any
		_ = json.Unmarshal([]byte(c.ArgsJson), &args)
		chunks := map[string]float64{"team-acme": 3, "other": 3, "acme-pyramid": 4}[args["collection"].(string)]
		if args["retrieval_mode"] != "graph" || args["top_k"] != chunks || args["max_context_chunks"] != chunks || args["max_context_chars"] != 1000*chunks {
			t.Fatalf("args: %v", args)
		}
	}
	ids := b.auditLines(t)[0]["ids"].(map[string]any)["knowledge_ids"].([]any)
	if fmt.Sprint(ids) != "[team/acme:a-1 team/acme:pyramide/ledger team/acme:zz team/other:o-1]" {
		t.Fatalf("audit ids: %v", ids)
	}

	got = decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "ledger", "brains": []string{"team/other"}}))
	if len(got.Results) != 1 || got.Results[0].Snippet != "Something else." || got.Results[0].Provenance != "[team/other] Other, o-1" {
		t.Fatalf("a brain without chunks: %+v", got.Results)
	}
	b.brain.calls = nil
	mustCall(t, s, "th_recall", map[string]any{"query": "ledger"})
	for _, c := range b.brain.calls {
		var args map[string]any
		_ = json.Unmarshal([]byte(c.ArgsJson), &args)
		want := map[string]float64{"team-acme": 10, "other": 10, "acme-pyramid": 4}[args["collection"].(string)]
		if args["max_context_chunks"] != want {
			t.Fatalf("defaults: %v", args)
		}
	}
}

// th_recall searches the query and each of its alternate_queries apart,
// in the passages and in the pyramid, and fuses every list by reciprocal
// rank: the engine stops at the query when it fills top_k, so a variant in
// another language reaches the brain only as its own search. At most 4
// variants. A variant that fails is said in skipped; the query's results
// stay.
func TestRecallSearchesEachAlternateQuery(t *testing.T) {
	b, secret := pyramidBench(t)
	b.brain.hits = map[string][]fakeHit{"team-acme": {{"a-1", "Ledger plan", "..."}, {"b-1", "Grand livre", "..."}}, "acme-pyramid": {}}
	b.brain.chunks = map[string][]fakeChunk{
		"team-acme|grand livre":    {{"b-1", "Le grand livre migre dimanche."}, {"a-1", "The ledger moves on Sunday."}},
		"team-acme|ledger":         {{"a-1", "The ledger moves on Sunday."}, {"c-1", "The cutover is on Thursday."}},
		"acme-pyramid|grand livre": {{"pyramide/fr", "Synthèse."}},
		"acme-pyramid|ledger":      {{"pyramide/en", "Synthesis."}},
	}
	s := b.federated(t, secret)
	order := func(args map[string]any) string {
		got := decode[recallOut](t, mustCall(t, s, "th_recall", args))
		var ids []string
		for _, r := range got.Results {
			ids = append(ids, r.KnowledgeID)
		}
		return strings.Join(ids, " ")
	}
	if got := order(map[string]any{"query": "grand livre", "brains": []string{"team/acme"}}); got != "pyramide/fr b-1 a-1" {
		t.Fatalf("without a variant: %s", got)
	}
	b.brain.calls = nil
	if got := order(map[string]any{"query": "grand livre", "alternate_queries": []string{"ledger"}, "brains": []string{"team/acme"}}); got != "pyramide/fr pyramide/en a-1 b-1 c-1" {
		t.Fatalf("with a variant: %s", got)
	}
	var asked []string
	for _, c := range b.brain.calls {
		var args map[string]any
		_ = json.Unmarshal([]byte(c.ArgsJson), &args)
		if _, sent := args["alternate_queries"]; sent {
			t.Fatalf("the variants went to the engine: %v", args)
		}
		asked = append(asked, args["collection"].(string)+"|"+args["query"].(string))
	}
	sort.Strings(asked)
	if strings.Join(asked, " ") != "acme-pyramid|grand livre acme-pyramid|ledger team-acme|grand livre team-acme|ledger" {
		t.Fatalf("searches: %v", asked)
	}
	ids := b.auditLines(t)[1]["ids"].(map[string]any)["knowledge_ids"].([]any)
	if fmt.Sprint(ids) != "[team/acme:a-1 team/acme:b-1 team/acme:c-1 team/acme:pyramide/en team/acme:pyramide/fr]" {
		t.Fatalf("audit ids: %v", ids)
	}

	b.brain.failQuery = "ledger"
	got := decode[recallOut](t, mustCall(t, s, "th_recall", map[string]any{"query": "grand livre", "alternate_queries": []string{"ledger"}, "brains": []string{"team/acme"}, "syntheses": 0}))
	if len(got.Results) != 2 || got.Results[0].KnowledgeID != "b-1" || len(got.Skipped) != 1 || !strings.HasPrefix(got.Skipped[0].Reason, "alternate query 1: ") || !strings.Contains(got.Skipped[0].Reason, "brain says no") {
		t.Fatalf("a failed variant: %+v", got)
	}
	r, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: map[string]any{"query": "q", "alternate_queries": []string{"a", "b", "c", "d", "e"}}})
	if err != nil || !r.IsError || !strings.Contains(r.Content[0].(*mcp.TextContent).Text, "at most 4 alternate_queries") {
		t.Fatalf("five variants: %v %+v", err, r)
	}
}

// The schema of th_recall says the defaults the code applies, and offers
// alternate_queries with its bound.
func TestRecallSchemaSaysItsDefaults(t *testing.T) {
	b, secret := pyramidBench(t)
	tools, err := b.federated(t, secret).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		if tool.Name != "th_recall" {
			continue
		}
		raw, _ := json.Marshal(tool.InputSchema)
		for _, want := range []string{`"passages to return, default 10"`, `default 4, 0 for none`, `"maxItems":4`, `at most 4; each is searched apart`} {
			if !strings.Contains(string(raw), want) {
				t.Fatalf("schema without %s: %s", want, raw)
			}
		}
		return
	}
	t.Fatal("no th_recall")
}

// th_descend opens the sources a synthesis cites: it reads the synthesis,
// takes the ids listed under the pyramid's sources key (a JSON list, or one
// id), and reads each, up to limit (5 by default). A source that cannot be
// read is listed in missing. It needs a brain with a pyramid and a read
// right, and is audited as a read of the brain.
func TestThDescendOpensTheSourcesOfASynthesis(t *testing.T) {
	b, secret := pyramidBench(t)
	b.brain.docs = map[string]map[string]any{
		"pyramide/ledger": {"id": "pyramide/ledger", "title": "Ledger", "content": "The ledger moves on Sunday.", "metadata": map[string]string{"preuves": `["plan.md","rollback.md","gone.md"]`}},
		"pyramide/one":    {"id": "pyramide/one", "title": "One", "content": "One source.", "metadata": map[string]string{"preuves": "plan.md"}},
		"pyramide/none":   {"id": "pyramide/none", "title": "None", "content": "No source."},
		"plan.md":         {"id": "plan.md", "title": "Ledger plan", "author": "mdupont", "source_url": "https://wiki/plan", "content": "The code word is MARMOT."},
		"rollback.md":     {"id": "rollback.md", "title": "Rollback", "content": "Two hours."},
	}
	b.brain.failID = "gone.md"
	s := b.federated(t, secret)
	type descent struct {
		Synthesis struct {
			KnowledgeID string `json:"knowledge_id"`
			Title       string `json:"title"`
		} `json:"synthesis"`
		Sources []struct {
			KnowledgeID string `json:"knowledge_id"`
			Title       string `json:"title"`
			Author      string `json:"author"`
			SourceURL   string `json:"source_url"`
			Content     string `json:"content"`
		} `json:"sources"`
		Missing []struct {
			KnowledgeID string `json:"knowledge_id"`
			Reason      string `json:"reason"`
		} `json:"missing"`
		More int `json:"more"`
	}
	got := decode[descent](t, mustCall(t, s, "th_descend", map[string]any{"brain": "team/acme", "knowledge_id": "pyramide/ledger"}))
	if got.Synthesis.KnowledgeID != "pyramide/ledger" || len(got.Sources) != 2 || got.Sources[0].KnowledgeID != "plan.md" || got.Sources[0].Content != "The code word is MARMOT." || got.Sources[0].Author != "mdupont" || got.Sources[0].SourceURL != "https://wiki/plan" || got.Sources[1].Title != "Rollback" {
		t.Fatalf("descent: %+v", got)
	}
	if len(got.Missing) != 1 || got.Missing[0].KnowledgeID != "gone.md" || !strings.Contains(got.Missing[0].Reason, "no such document") || got.More != 0 {
		t.Fatalf("missing: %+v", got)
	}
	for _, c := range b.brain.calls {
		var args map[string]any
		_ = json.Unmarshal([]byte(c.ArgsJson), &args)
		if _, has := args["collection"]; c.Name != "knowledge_get" || has {
			t.Fatalf("call: %+v", c)
		}
	}
	if l := b.auditLines(t)[0]; l["tool"] != "th_descend" || l["brain"] != "team/acme" || l["ok"] != true || fmt.Sprint(l["ids"].(map[string]any)["sources"]) != "[plan.md rollback.md]" || l["ids"].(map[string]any)["knowledge_id"] != "pyramide/ledger" {
		t.Fatalf("audit: %+v", l)
	}
	got = decode[descent](t, mustCall(t, s, "th_descend", map[string]any{"brain": "team/acme", "knowledge_id": "pyramide/ledger", "limit": 1}))
	if len(got.Sources) != 1 || got.More != 2 || len(got.Missing) != 0 {
		t.Fatalf("limit: %+v", got)
	}
	got = decode[descent](t, mustCall(t, s, "th_descend", map[string]any{"brain": "team/acme", "knowledge_id": "pyramide/one"}))
	if len(got.Sources) != 1 || got.Sources[0].KnowledgeID != "plan.md" {
		t.Fatalf("one id: %+v", got)
	}
	for _, c := range []struct {
		args map[string]any
		want string
	}{
		{map[string]any{"brain": "team/acme", "knowledge_id": "pyramide/none"}, "cites no source under preuves"},
		{map[string]any{"brain": "team/other", "knowledge_id": "x"}, "team/other has no pyramid"},
		{map[string]any{"brain": "team/acme"}, "knowledge_id is required"},
		{map[string]any{"brain": "org/none", "knowledge_id": "x"}, "unknown brain"},
		{map[string]any{"brain": "team/acme", "knowledge_id": "gone.md"}, "no such document"},
	} {
		res := mustCall(t, s, "th_descend", c.args)
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, c.want) {
			t.Fatalf("%v: %+v", c.args, res)
		}
	}
	reader, err := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Read)
	if err != nil {
		t.Fatal(err)
	}
	if res := mustCall(t, b.federated(t, reader.Secret), "th_descend", map[string]any{"brain": "team/other", "knowledge_id": "x"}); !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "may not read") {
		t.Fatalf("no read right: %+v", res)
	}
	b.brain.docs["pyramide/bad"] = map[string]any{"id": "pyramide/bad", "metadata": map[string]string{"preuves": `["unclosed`}}
	if res := mustCall(t, s, "th_descend", map[string]any{"brain": "team/acme", "knowledge_id": "pyramide/bad"}); !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "cites no source") {
		t.Fatalf("bad list: %+v", res)
	}
	b.brain.badTool = "knowledge_get"
	if res := mustCall(t, s, "th_descend", map[string]any{"brain": "team/acme", "knowledge_id": "pyramide/ledger"}); !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "decode") {
		t.Fatalf("bad answer: %+v", res)
	}
	if res, _ := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_descend", Arguments: json.RawMessage(`{"brain": 5}`)}); !res.IsError {
		t.Fatal("bad arguments passed")
	}
	if l := b.auditLines(t); l[len(l)-1]["tool"] != "th_descend" || l[len(l)-1]["ok"] != false {
		t.Fatalf("a failed descent is audited: %+v", l[len(l)-1])
	}
}

// th_add replaces a document of the same id. A writer may replace their own
// document, never another author's: the proposal would take the accepted
// text out of the brain until a curator reads it, under the writer's name.
// The gateway reads the document first (knowledge_get, no collection) and
// refuses before writing; an id the brain does not have is free, and a
// curator may replace any document. The refusal is audited; a read that
// fails or does not decode refuses too.
func TestGcAddDoesNotReplaceAnotherAuthorsDocument(t *testing.T) {
	b := start(t)
	withCurator(t, b)
	b.brain.missing = true
	b.brain.docs = map[string]map[string]any{
		"spec-1": {"id": "spec-1", "author": "kmlead", "metadata": map[string]string{"_review_status": "accepted"}},
		"spec-2": {"id": "spec-2", "author": "mdupont", "metadata": map[string]string{"_review_status": "accepted"}},
		"spec-3": {"id": "spec-3"},
	}
	k, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	for _, id := range []string{"spec-1", "spec-3"} {
		res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": id, "content": "Taken over."})
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "belongs to") {
			t.Fatalf("%s: %+v", id, res)
		}
	}
	if b.brain.named("ingest_document") != nil {
		t.Fatalf("written: %+v", b.brain.calls)
	}
	get := b.brain.named("knowledge_get")
	if _, has := get["collection"]; get["knowledge_id"] != "spec-1" || has {
		t.Fatalf("get: %+v", get)
	}
	if l := b.auditLines(t)[0]; l["tool"] != "th_add" || l["ok"] != false || l["ids"].(map[string]any)["document_id"] != "spec-1" || !strings.Contains(l["error"].(string), "kmlead") {
		t.Fatalf("audit: %+v", l)
	}
	for _, id := range []string{"spec-2", "spec-new"} {
		b.brain.calls = nil
		if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": id, "content": "Mine."}); res.IsError {
			t.Fatalf("%s: %s", id, res.Content[0].(*mcp.TextContent).Text)
		}
		if args := b.brain.named("ingest_document"); args["document_id"] != id || args["review_status"] != "proposed" {
			t.Fatalf("%s: %+v", id, b.brain.calls)
		}
	}
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	b.brain.calls = nil
	if res := mustCall(t, b.federated(t, kc.Secret), "th_add", map[string]any{"brain": "team/acme", "document_id": "spec-2", "content": "Edited."}); res.IsError || len(b.brain.calls) != 1 || b.brain.calls[0].Name != "ingest_document" {
		t.Fatalf("curator: %+v %+v", res, b.brain.calls)
	}
	for _, c := range []struct{ failTool, badTool, want string }{
		{"knowledge_get", "", "brain says no to knowledge_get"},
		{"", "knowledge_get", "decode knowledge_get"},
	} {
		b.brain.failTool, b.brain.badTool, b.brain.calls = c.failTool, c.badTool, nil
		res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "spec-2", "content": "Mine."})
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, c.want) || b.brain.named("ingest_document") != nil {
			t.Fatalf("%+v: %+v", c, res)
		}
	}
}

// th_get serves a document that waits for review, or was rejected, only to
// its author and to the curators: the review gate keeps it out of the
// searches, and a reader who knows its id must not read it whole. The
// reader is told it is not found; the audit keeps the reason.
func TestThGetServesAProposalToItsAuthorAndTheCuratorsOnly(t *testing.T) {
	b := start(t)
	withCurator(t, b)
	b.brain.docs = map[string]map[string]any{
		"p-1": {"id": "p-1", "author": "lfoletti", "content": "Draft MARMOT.", "metadata": map[string]string{"_review_status": "proposed"}},
		"r-1": {"id": "r-1", "author": "lfoletti", "content": "Refused MARMOT.", "metadata": map[string]string{"_review_status": "rejected"}},
		"a-1": {"id": "a-1", "author": "lfoletti", "content": "Accepted MARMOT.", "metadata": map[string]string{"_review_status": "accepted"}},
		"n-1": {"id": "n-1", "content": "Loaded MARMOT."},
	}
	who := map[string]rights.Access{"mdupont": rights.Write, "lfoletti": rights.Write, "kmlead": rights.Curate}
	sessions := map[string]*mcp.ClientSession{}
	for u, a := range who {
		k, _ := keys.Issue(b.keysFile, u, "team/acme", "team-acme", a)
		sessions[u] = b.federated(t, k.Secret)
	}
	for _, c := range []struct {
		user, id string
		served   bool
	}{
		{"mdupont", "p-1", false}, {"mdupont", "r-1", false}, {"mdupont", "a-1", true}, {"mdupont", "n-1", true},
		{"lfoletti", "p-1", true}, {"lfoletti", "r-1", true}, {"kmlead", "p-1", true}, {"kmlead", "r-1", true},
	} {
		res := mustCall(t, sessions[c.user], "th_get", map[string]any{"brain": "team/acme", "knowledge_id": c.id})
		text := res.Content[0].(*mcp.TextContent).Text
		if c.served != !res.IsError || c.served != strings.Contains(text, "MARMOT") {
			t.Fatalf("%s %s: %+v", c.user, c.id, res)
		}
		if !c.served && !strings.Contains(text, "not found") {
			t.Fatalf("%s %s: %s", c.user, c.id, text)
		}
	}
	var refused []map[string]any
	for _, l := range b.auditLines(t) {
		if l["ok"] == false {
			refused = append(refused, l)
		}
	}
	if len(refused) != 2 || refused[0]["user"] != "mdupont" || !strings.Contains(refused[0]["error"].(string), "proposed") {
		t.Fatalf("audit: %+v", refused)
	}
	b.brain.badTool = "knowledge_get"
	if res := mustCall(t, sessions["mdupont"], "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "a-1"}); !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "decode knowledge_get") {
		t.Fatalf("bad JSON: %+v", res)
	}
}
