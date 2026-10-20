package relay

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	rpcv1 "github.com/liliang-cn/cortexdb/v2/pkg/rpc/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/lfoletti/granular-cortex/gateway/keys"
	"github.com/lfoletti/granular-cortex/gateway/rights"
)

// brain is a fake ToolsService: it records the bearer and the arguments of
// every call, and answers like a CortexDB server would.
type brain struct {
	rpcv1.UnimplementedToolsServiceServer
	mu                sync.Mutex
	bearer            string
	bearers           map[string]string // by collection
	calls             []call
	fail              bool
	failOn            string               // a collection whose calls fail
	hits              map[string][]fakeHit // knowledge_search answers by collection
	badJSON           bool                 // answer knowledge_search with text that is not JSON
	pending           map[string][]fakeHit // what aggregate_metadata and knowledge_get serve, by collection
	failTool, badTool string               // one tool that fails, one that answers text that is not JSON
}

type fakeHit struct{ ID, Title, Snippet string }

type call struct{ Name, ArgsJson string }

func (b *brain) seen(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	if v := md.Get("authorization"); len(v) > 0 {
		b.bearer = v[0]
		return v[0]
	}
	return ""
}

func (b *brain) ListTools(ctx context.Context, _ *rpcv1.ListToolsRequest) (*rpcv1.ListToolsResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.seen(ctx)
	return &rpcv1.ListToolsResponse{Tools: []*rpcv1.ToolDefinition{
		{Name: "knowledge_get", Description: "reads a document", InputSchemaJson: `{"type":"object","properties":{"knowledge_id":{"type":"string"}}}`},
		{Name: "ingest_document", Description: "writes a document", InputSchemaJson: `{"type":"object"}`},
	}}, nil
}

func (b *brain) CallTool(ctx context.Context, req *rpcv1.CallToolRequest) (*rpcv1.CallToolResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	bearer := b.seen(ctx)
	b.calls = append(b.calls, call{Name: req.GetName(), ArgsJson: req.GetArgsJson()})
	var args struct {
		Collection string `json:"collection"`
		TopK       int    `json:"top_k"`
	}
	_ = json.Unmarshal([]byte(req.GetArgsJson()), &args)
	if b.bearers == nil {
		b.bearers = map[string]string{}
	}
	b.bearers[args.Collection] = bearer
	if b.fail || (b.failOn != "" && b.failOn == args.Collection) {
		return nil, errors.New("brain says no")
	}
	if b.badJSON || b.badTool == req.GetName() {
		return &rpcv1.CallToolResponse{ResultJson: "not json"}, nil
	}
	if b.failTool == req.GetName() {
		return nil, errors.New("brain says no to " + req.GetName())
	}
	if req.GetName() == "aggregate_metadata" {
		var lim struct {
			Limit int `json:"limit"`
		}
		_ = json.Unmarshal([]byte(req.GetArgsJson()), &lim)
		groups := []map[string]any{}
		for i, x := range b.pending[args.Collection] {
			if i >= lim.Limit {
				break
			}
			groups = append(groups, map[string]any{"keys": map[string]string{"document_id": x.ID}, "count": 1})
		}
		out, _ := json.Marshal(map[string]any{"results": groups})
		return &rpcv1.CallToolResponse{ResultJson: string(out)}, nil
	}
	if req.GetName() == "knowledge_get" {
		var get struct {
			KnowledgeID string `json:"knowledge_id"`
		}
		_ = json.Unmarshal([]byte(req.GetArgsJson()), &get)
		for _, x := range b.pending["team-acme"] {
			if x.ID == get.KnowledgeID {
				out, _ := json.Marshal(map[string]any{"knowledge": map[string]any{"id": x.ID, "title": x.Title, "author": "mdupont", "content": x.Snippet}})
				return &rpcv1.CallToolResponse{ResultJson: string(out)}, nil
			}
		}
	}
	if req.GetName() == "knowledge_search" && b.hits != nil {
		type h struct {
			KnowledgeID string  `json:"knowledge_id"`
			Title       string  `json:"title"`
			Snippet     string  `json:"snippet"`
			Score       float64 `json:"score"`
		}
		var hs []h
		for i, x := range b.hits[args.Collection] {
			if args.TopK > 0 && i >= args.TopK {
				break
			}
			hs = append(hs, h{x.ID, x.Title, x.Snippet, 1 / float64(i+1)})
		}
		out, _ := json.Marshal(map[string]any{"query": "q", "results": hs})
		return &rpcv1.CallToolResponse{ResultJson: string(out)}, nil
	}
	return &rpcv1.CallToolResponse{ResultJson: `{"echo":` + req.GetArgsJson() + `}`}, nil
}

type bench struct {
	brain    *brain
	rights   *rights.Rights
	keysFile string
	audit    string
	url      string
}

func start(t *testing.T) *bench {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	b := &brain{}
	gs := grpc.NewServer()
	rpcv1.RegisterToolsServiceServer(gs, b)
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(gs.Stop)

	dir := t.TempDir()
	keysFile := filepath.Join(dir, "keys.json")
	r, err := rights.Parse([]byte(`
brains:
  team/acme:
    addr: ` + lis.Addr().String() + `
    keys_file: ` + keysFile + `
    collection: team-acme
    readers: [user:mdupont]
    writers: [user:lfoletti]
  team/other:
    addr: ` + lis.Addr().String() + `
    keys_file: ` + keysFile + `
    collection: other
    writers: [user:lfoletti]
    weight: 0.5
groups: {}
`))
	if err != nil {
		t.Fatal(err)
	}
	audit := filepath.Join(dir, "audit.jsonl")
	a, err := OpenAudit(audit)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	srv := httptest.NewServer(New(r, a).Handler())
	t.Cleanup(srv.Close)
	return &bench{brain: b, rights: r, keysFile: keysFile, audit: audit, url: srv.URL}
}

type bearer struct{ token string }

func (b bearer) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(req)
}

func (b *bench) session(t *testing.T, brain, token string) (*mcp.ClientSession, error) {
	t.Helper()
	tr := &mcp.StreamableClientTransport{Endpoint: b.url + "/mcp/" + brain, HTTPClient: &http.Client{Transport: bearer{token}}, MaxRetries: -1}
	return mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(context.Background(), tr, nil)
}

func (b *bench) auditLines(t *testing.T) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(b.audit)
	if err != nil {
		t.Fatal(err)
	}
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("audit line %q: %v", line, err)
		}
		out = append(out, m)
	}
	return out
}

// A call through the gateway reaches the brain with the caller's own key,
// its arguments unchanged, and leaves an audit line that says who, what,
// where and which ids.
func TestACallIsRelayedWithTheCallersKeyAndAudited(t *testing.T) {
	b := start(t)
	k, err := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	if err != nil {
		t.Fatal(err)
	}
	s, err := b.session(t, "team/acme", k.Secret)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()
	tools, err := s.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 2 {
		t.Fatalf("tools: %+v, %v", tools, err)
	}
	if names := tools.Tools[0].Name + "," + tools.Tools[1].Name; names != "ingest_document,knowledge_get" {
		t.Fatalf("tools: %s", names)
	}
	res, err := s.CallTool(ctx, &mcp.CallToolParams{Name: "knowledge_get", Arguments: map[string]any{"knowledge_id": "doc-1", "collection": "team-acme"}})
	if err != nil || res.IsError {
		t.Fatalf("call: %+v, %v", res, err)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, `"echo":`) || !strings.Contains(text, `"knowledge_id":"doc-1"`) {
		t.Fatalf("result %q", text)
	}
	if b.brain.bearer != "Bearer "+k.Secret {
		t.Fatalf("the brain saw %q", b.brain.bearer)
	}
	if len(b.brain.calls) != 1 || b.brain.calls[0].Name != "knowledge_get" || !strings.Contains(b.brain.calls[0].ArgsJson, `"collection":"team-acme"`) {
		t.Fatalf("calls: %+v", b.brain.calls)
	}
	lines := b.auditLines(t)
	if len(lines) != 1 {
		t.Fatalf("audit: %+v", lines)
	}
	l := lines[0]
	ids, _ := l["ids"].(map[string]any)
	if l["user"] != "lfoletti" || l["brain"] != "team/acme" || l["tool"] != "knowledge_get" || l["key"] != k.ID || l["ok"] != true || ids["knowledge_id"] != "doc-1" || l["ts"] == "" {
		t.Fatalf("audit line: %+v", l)
	}
}

// A failure of the brain is the tool's error for the client, and an audit
// line that says so; list ids are recorded too.
func TestAToolFailureIsAudited(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s, err := b.session(t, "team/acme", k.Secret)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	b.brain.fail = true
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "ingest_document", Arguments: map[string]any{"document_ids": []string{"a", "b"}, "supersedes": []string{"m-1"}, "content": "x"}})
	if err != nil || !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "brain says no") {
		t.Fatalf("got %+v, %v", res, err)
	}
	l := b.auditLines(t)[0]
	ids := l["ids"].(map[string]any)
	if l["ok"] != false || !strings.Contains(l["error"].(string), "brain says no") || len(ids) != 2 || ids["document_ids"] == nil || ids["supersedes"] == nil || ids["content"] != nil {
		t.Fatalf("audit line: %+v", l)
	}
}

// The gateway refuses before it dials: a secret it does not know (401), a
// key issued for another brain of the same server (403), a user whose right
// was withdrawn from rights.yaml although the key still exists (403), and a
// brain that is not declared (404). None of these reaches the brain or the
// audit file.
func TestRefusalsHappenBeforeTheBrainIsReached(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	other, _ := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	stale, _ := keys.Issue(b.keysFile, "ghost", "team/acme", "team-acme", rights.Write)
	for _, c := range []struct {
		name, path, token string
		want              int
	}{
		{"unknown secret", "/mcp/team/acme", "nope", http.StatusUnauthorized},
		{"key of another brain", "/mcp/team/acme", other.Secret, http.StatusForbidden},
		{"right withdrawn", "/mcp/team/acme", stale.Secret, http.StatusForbidden},
		{"unknown brain", "/mcp/team/acm", k.Secret, http.StatusNotFound},
		{"no token", "/mcp/team/acme", "", http.StatusUnauthorized},
		{"a path under the brains", "/mcp/", k.Secret, http.StatusNotFound},
	} {
		req, _ := http.NewRequest(http.MethodPost, b.url+c.path, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
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
	if b.brain.bearer != "" || len(b.brain.calls) != 0 {
		t.Fatalf("the brain was reached: %q %+v", b.brain.bearer, b.brain.calls)
	}
	if data, _ := os.ReadFile(b.audit); len(data) != 0 {
		t.Fatalf("audit has %q", data)
	}
}

// A brain that cannot be reached, or whose key file cannot be read, is an
// error to the client, not a silent empty tool list.
func TestABrainThatCannotBeReachedIsAnError(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	br := b.rights.Brains["team/acme"]
	br.Addr = "127.0.0.1:1"
	b.rights.Brains["team/acme"] = br
	if _, err := b.session(t, "team/acme", k.Secret); err == nil || !strings.Contains(err.Error(), "Bad Gateway") {
		t.Fatalf("got %v", err)
	}
	for _, addr := range []string{"http://127.0.0.1:1/x", "%zz"} {
		br.Addr = addr
		b.rights.Brains["team/acme"] = br
		if _, err := b.session(t, "team/acme", k.Secret); err == nil || !strings.Contains(err.Error(), "Bad Gateway") {
			t.Fatalf("%s: got %v", addr, err)
		}
	}
	if err := os.WriteFile(b.keysFile, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := b.session(t, "team/acme", k.Secret); err == nil || !strings.Contains(err.Error(), "Internal Server Error") {
		t.Fatalf("broken key file: got %v", err)
	}
}

// A call that names no row has no ids in its audit line; arguments that are
// not an object give none either.
func TestACallWithoutIdsHasNoneInTheAudit(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s, err := b.session(t, "team/acme", k.Secret)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "ingest_document", Arguments: map[string]any{"content": "x"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "ingest_document"}); err != nil {
		t.Fatal(err)
	}
	for _, l := range b.auditLines(t) {
		if _, has := l["ids"]; has {
			t.Fatalf("audit line has ids: %+v", l)
		}
	}
	if b.brain.calls[1].ArgsJson != "{}" {
		t.Fatalf("no arguments became %q", b.brain.calls[1].ArgsJson)
	}
	if idsOf("5") != nil || idsOf(`{"content":"x"}`) != nil {
		t.Fatal("idsOf is wrong")
	}
}

// The audit file is opened for append, so a restart keeps what was written,
// and a file that cannot be opened is an error that names it.
func TestTheAuditFileAppendsAndNamesItselfOnFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	a, err := OpenAudit(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Record(Entry{User: "u", Brain: "b", Tool: "t", OK: true}); err != nil {
		t.Fatal(err)
	}
	_ = a.Close()
	a, err = OpenAudit(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Record(Entry{User: "u", Brain: "b", Tool: "t2", OK: false, Error: "x"}); err != nil {
		t.Fatal(err)
	}
	_ = a.Close()
	data, _ := os.ReadFile(path)
	if n := strings.Count(string(data), "\n"); n != 2 || !strings.Contains(string(data), `"tool":"t2"`) {
		t.Fatalf("audit: %q", data)
	}
	if _, err := OpenAudit(filepath.Join(dir, "missing", "audit.jsonl")); err == nil || !strings.Contains(err.Error(), "audit.jsonl") {
		t.Fatalf("got %v", err)
	}
	if err := a.Record(Entry{}); err == nil {
		t.Fatal("a closed audit accepted a record")
	}
	if err := a.Close(); err != nil {
		t.Fatalf("a second close: %v", err)
	}
}
