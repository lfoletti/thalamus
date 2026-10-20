package relay

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/granular-cortex/gateway/keys"
	"github.com/lfoletti/granular-cortex/gateway/rights"
)

// recallResult is what gc_recall returns.
type recallOut struct {
	Results []struct {
		Brain       string   `json:"brain"`
		KnowledgeID string   `json:"knowledge_id"`
		Title       string   `json:"title"`
		Snippet     string   `json:"snippet"`
		Score       float64  `json:"score"`
		Provenance  string   `json:"provenance"`
		AlsoIn      []string `json:"also_in,omitempty"`
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

// gc_brains lists the caller's brains with the access and whether a key
// was issued; the caller is known from any one of their gateway keys.
func TestGcBrainsListsTheCallersBrains(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	tools, err := s.ListTools(context.Background(), nil)
	if err != nil || len(tools.Tools) != 2 || tools.Tools[0].Name != "gc_brains" || tools.Tools[1].Name != "gc_recall" {
		t.Fatalf("tools: %+v %v", tools, err)
	}
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "gc_brains"})
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
	l := b.auditLines(t)[0]
	if l["tool"] != "gc_brains" || l["user"] != "lfoletti" || l["brain"] != "" || l["ok"] != true {
		t.Fatalf("audit: %+v", l)
	}
}

// gc_recall asks every readable brain that has a key, with that key, and
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
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "gc_recall", Arguments: map[string]any{"query": "ledger", "top_k": 3}})
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
	if l["tool"] != "gc_recall" || l["ok"] != true || len(ids["brains"].([]any)) != 2 || len(ids["knowledge_ids"].([]any)) != 4 {
		t.Fatalf("audit: %+v", l)
	}

	// Without a key on team/other, it is skipped and said so.
	if err := keys.Revoke(b.keysFile, "lfoletti", "team/other"); err != nil {
		t.Fatal(err)
	}
	res, err = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "gc_recall", Arguments: map[string]any{"query": "ledger"}})
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
	res, err = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "gc_recall", Arguments: map[string]any{"query": "ledger"}})
	if err != nil {
		t.Fatal(err)
	}
	got = decode[recallOut](t, res)
	if len(got.Results) != 2 || len(got.Skipped) != 1 || !strings.Contains(got.Skipped[0].Reason, "brain says no") {
		t.Fatalf("got %+v", got)
	}
	// A query is required; a brain the caller names but may not read is refused.
	res, _ = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "gc_recall", Arguments: map[string]any{"query": ""}})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "query") {
		t.Fatalf("empty query: %+v", res)
	}
	res, _ = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "gc_recall", Arguments: map[string]any{"query": "x", "brains": []string{"org/none"}}})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "org/none") {
		t.Fatalf("unknown brain: %+v", res)
	}
	res, _ = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "gc_recall", Arguments: json.RawMessage(`{"query": 5}`)})
	if !res.IsError {
		t.Fatalf("bad arguments: %+v", res)
	}
}

// gc_recall restricted to named brains asks only those; a brain that is
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
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "gc_recall", Arguments: map[string]any{"query": "x", "brains": []string{"team/other"}}})
	if err != nil {
		t.Fatal(err)
	}
	got := decode[recallOut](t, res)
	if len(got.Results) != 1 || got.Results[0].Brain != "team/other" || len(b.brain.calls) != 1 || strings.Contains(b.brain.calls[0].ArgsJson, "collection") {
		t.Fatalf("got %+v, calls %+v", got, b.brain.calls)
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
	got := decode[recallOut](t, mustCall(t, s, "gc_recall", map[string]any{"query": "x"}))
	if len(got.Results) != 0 || len(got.Skipped) != 2 || got.Skipped[0].Brain != "team/acme" || got.Skipped[1].Brain != "team/other" {
		t.Fatalf("got %+v", got)
	}
	b.brain.fail = false
	other := b.rights.Brains["team/other"]
	other.Addr = "%zz"
	b.rights.Brains["team/other"] = other
	got = decode[recallOut](t, mustCall(t, s, "gc_recall", map[string]any{"query": "x"}))
	if len(got.Skipped) != 1 || got.Skipped[0].Brain != "team/other" || !strings.Contains(got.Skipped[0].Reason, "connect to %zz") {
		t.Fatalf("got %+v", got)
	}
	b.brain.badJSON = true
	got = decode[recallOut](t, mustCall(t, s, "gc_recall", map[string]any{"query": "x", "brains": []string{"team/acme"}}))
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
		if got := provenance("b", c.h); got != c.want {
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
	out := fuse(lists, 10)
	if len(out) != 3 || out[0].KnowledgeID != "a-1" || strings.Join(out[0].AlsoIn, ",") != "b" {
		t.Fatalf("%+v", out)
	}
	if want := 1.0/61 + 0.5/61; out[0].Score < want-1e-9 || out[0].Score > want+1e-9 {
		t.Fatalf("score %v", out[0].Score)
	}
	if out[1].KnowledgeID != "a-2" || out[2].KnowledgeID != "b-2" {
		t.Fatalf("%+v", out)
	}
	if got := fuse(lists, 1); len(got) != 1 {
		t.Fatalf("top_k: %+v", got)
	}
	// Rank 1 of the second brain comes before rank 2 of the first: the
	// lists are interleaved by score, not concatenated.
	equal := []brainHits{
		{Brain: "a", Weight: 1, Hits: []hit{{KnowledgeID: "a-1", Snippet: "1"}, {KnowledgeID: "a-2", Snippet: "2"}}},
		{Brain: "b", Weight: 1, Hits: []hit{{KnowledgeID: "b-1", Snippet: "3"}, {KnowledgeID: "b-2", Snippet: "4"}}},
	}
	var ids []string
	for _, f := range fuse(equal, 10) {
		ids = append(ids, f.KnowledgeID)
	}
	if strings.Join(ids, ",") != "a-1,b-1,a-2,b-2" {
		t.Fatalf("order %v", ids)
	}
}
