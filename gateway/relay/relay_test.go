package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	rpcv1 "github.com/liliang-cn/cortexdb/v2/pkg/rpc/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
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
	failOn            string                    // a collection whose calls fail
	failQuery         string                    // a query whose calls fail
	hits              map[string][]fakeHit      // knowledge_search answers by collection
	chunks            map[string][]fakeChunk    // the chunks knowledge_search adds, by collection
	badJSON           bool                      // answer knowledge_search with text that is not JSON
	pending           map[string][]fakeHit      // what aggregate_metadata and knowledge_get serve, by collection
	failTool, badTool string                    // one tool that fails, one that answers text that is not JSON
	denyTool          string                    // one tool the brain refuses to the key (PermissionDenied)
	docs              map[string]map[string]any // what knowledge_get serves by id, in any collection
	failID            string                    // a knowledge_get id whose call fails
	missing           bool                      // knowledge_get answers NotFound for an id it does not serve, as the engine does
	graph             map[string]fakeGraph      // what get_nodes, knowledge_memory_neighbors and expand_graph serve, by the caller's secret
	slow              map[string]time.Duration  // how long knowledge_search waits, by collection
	cut               atomic.Bool               // a slow call saw its caller go away
}

// fakeGraph is one key's part of the graph around one entity.
type fakeGraph struct {
	Known     bool             // get_nodes serves the entity, and expand_graph returns it with its neighbors
	Neighbors []map[string]any // what expand_graph serves around the entity
	Walked    []map[string]any // what knowledge_memory_neighbors serves, which follows the mentions
	Resolved  string           // the node id the brain resolves the name to, if not the one derived from it
	Fail      bool
	Deny      bool // the brain refuses the key (PermissionDenied)
}

type fakeHit struct{ ID, Title, Snippet string }

type fakeChunk struct{ DocumentID, Content string }

type call struct{ Name, ArgsJson string }

// named is the arguments of the first call to a tool, nil when there is none.
func (b *brain) named(tool string) map[string]any {
	for _, c := range b.calls {
		if c.Name == tool {
			var args map[string]any
			_ = json.Unmarshal([]byte(c.ArgsJson), &args)
			return args
		}
	}
	return nil
}

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
	// A slow collection waits before the lock, so the other brains answer.
	var where struct {
		Collection string `json:"collection"`
	}
	_ = json.Unmarshal([]byte(req.GetArgsJson()), &where)
	if d := b.slow[where.Collection]; d > 0 && req.GetName() == "knowledge_search" {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			b.cut.Store(true)
			return nil, ctx.Err()
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	bearer := b.seen(ctx)
	b.calls = append(b.calls, call{Name: req.GetName(), ArgsJson: req.GetArgsJson()})
	var args struct {
		Collection string `json:"collection"`
		TopK       int    `json:"top_k"`
		Query      string `json:"query"`
	}
	_ = json.Unmarshal([]byte(req.GetArgsJson()), &args)
	if b.bearers == nil {
		b.bearers = map[string]string{}
	}
	b.bearers[args.Collection] = bearer
	if b.fail || (b.failOn != "" && b.failOn == args.Collection) || (b.failQuery != "" && b.failQuery == args.Query) {
		return nil, errors.New("brain says no")
	}
	if b.badJSON || b.badTool == req.GetName() {
		return &rpcv1.CallToolResponse{ResultJson: "not json"}, nil
	}
	if b.failTool == req.GetName() {
		return nil, errors.New("brain says no to " + req.GetName())
	}
	if b.denyTool == req.GetName() {
		return nil, status.Error(codes.PermissionDenied, "not for this key")
	}
	if g, ok := b.graph[strings.TrimPrefix(bearer, "Bearer ")]; ok && (req.GetName() == "get_nodes" || req.GetName() == "knowledge_memory_neighbors" || req.GetName() == "expand_graph") {
		if g.Fail {
			return nil, errors.New("graph says no")
		}
		if g.Deny {
			return nil, status.Error(codes.PermissionDenied, "not for this key")
		}
		var walk struct {
			EntityName string   `json:"entity_name"`
			NodeIDs    []string `json:"node_ids"`
		}
		_ = json.Unmarshal([]byte(req.GetArgsJson()), &walk)
		id := "entity:" + strings.ToLower(walk.EntityName)
		if g.Resolved != "" {
			id = g.Resolved
		}
		reply := map[string]any{"resolved_node_id": id, "neighbors": g.Walked}
		if req.GetName() == "expand_graph" {
			nodes := []map[string]any{}
			if g.Known {
				nodes = append(nodes, map[string]any{"id": walk.NodeIDs[0], "vector": nil, "content": walk.NodeIDs[0]})
			}
			// The engine returns the subgraph in order of node id, the seed among it.
			nodes = append(nodes, g.Neighbors...)
			sort.SliceStable(nodes, func(i, j int) bool { return fmt.Sprint(nodes[i]["id"]) < fmt.Sprint(nodes[j]["id"]) })
			reply = map[string]any{"nodes": nodes, "edges": []any{}, "scores": map[string]float64{}}
		}
		if req.GetName() == "get_nodes" {
			nodes := []map[string]any{}
			if g.Known {
				nodes = append(nodes, map[string]any{"id": walk.NodeIDs[0], "vector": nil, "content": walk.NodeIDs[0]})
			}
			reply = map[string]any{"nodes": nodes}
		}
		out, _ := json.Marshal(reply)
		return &rpcv1.CallToolResponse{ResultJson: string(out)}, nil
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
		if get.KnowledgeID == b.failID && b.failID != "" {
			return nil, errors.New("no such document " + get.KnowledgeID)
		}
		if d, ok := b.docs[get.KnowledgeID]; ok {
			out, _ := json.Marshal(map[string]any{"knowledge": d})
			return &rpcv1.CallToolResponse{ResultJson: string(out)}, nil
		}
		for _, x := range b.pending["team-acme"] {
			if x.ID == get.KnowledgeID {
				out, _ := json.Marshal(map[string]any{"knowledge": map[string]any{"id": x.ID, "title": x.Title, "author": "mdupont", "content": x.Snippet}})
				return &rpcv1.CallToolResponse{ResultJson: string(out)}, nil
			}
		}
		if b.missing {
			return nil, status.Error(codes.NotFound, "not found")
		}
	}
	if req.GetName() == "knowledge_search" && b.hits != nil {
		type h struct {
			KnowledgeID string  `json:"knowledge_id"`
			Title       string  `json:"title"`
			SourceURL   string  `json:"source_url,omitempty"`
			Author      string  `json:"author,omitempty"`
			Snippet     string  `json:"snippet"`
			Score       float64 `json:"score"`
		}
		var hs []h
		for i, x := range b.hits[args.Collection] {
			if args.TopK > 0 && i >= args.TopK {
				break
			}
			hit := h{KnowledgeID: x.ID, Title: x.Title, Snippet: x.Snippet, Score: 1 / float64(i+1)}
			if b.chunks[args.Collection] != nil {
				hit.SourceURL, hit.Author = "https://wiki/"+x.ID, "mdupont"
			}
			hs = append(hs, hit)
		}
		reply := map[string]any{"query": "q", "results": hs}
		var max struct {
			Chunks int `json:"max_context_chunks"`
		}
		_ = json.Unmarshal([]byte(req.GetArgsJson()), &max)
		var cs []map[string]any
		chunks, byQuery := b.chunks[args.Collection+"|"+args.Query]
		if !byQuery {
			chunks = b.chunks[args.Collection]
		}
		for i, c := range chunks {
			if max.Chunks > 0 && i >= max.Chunks {
				break
			}
			cs = append(cs, map[string]any{"ID": fmt.Sprintf("chunk:%s:%d", c.DocumentID, i), "DocumentID": c.DocumentID, "Content": c.Content, "Score": 1 / float64(i+1)})
		}
		if cs != nil {
			reply["chunks"] = cs
		}
		out, _ := json.Marshal(reply)
		return &rpcv1.CallToolResponse{ResultJson: string(out)}, nil
	}
	return &rpcv1.CallToolResponse{ResultJson: `{"echo":` + req.GetArgsJson() + `}`}, nil
}

type bench struct {
	a        *Audit
	brain    *brain
	rights   *rights.Rights
	keysFile string
	audit    string
	url      string
}

func start(t *testing.T) *bench { t.Helper(); return startWith(t, func(y string) string { return y }) }

// startWith starts the bench with rights the edit changes.
func startWith(t *testing.T, edit func(string) string) *bench {
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
	r, err := rights.Parse([]byte(edit(`
brains:
  team/acme:
    addr: ` + lis.Addr().String() + `
    keys_file: ` + keysFile + `
    collection: team-acme
    readers: [user:mdupont]
    writers: [user:lfoletti]
    audit_reads: true
  team/other:
    addr: ` + lis.Addr().String() + `
    keys_file: ` + keysFile + `
    collection: other
    writers: [user:lfoletti]
    weight: 0.5
    audit_reads: true
groups: {}
`)))
	if err != nil {
		t.Fatal(err)
	}
	audit := filepath.Join(dir, "audit.jsonl")
	a, err := OpenAudit(audit)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	srv := httptest.NewServer(New(rights.Static(r), a).Handler())
	t.Cleanup(srv.Close)
	return &bench{a: a, brain: b, rights: r, keysFile: keysFile, audit: audit, url: srv.URL}
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
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != c.want {
			t.Fatalf("%s: got %d, want %d", c.name, resp.StatusCode, c.want)
		}
		// A request without a key is refused before the key file is read.
		if c.token == "" && !strings.Contains(string(body), "a key issued by the gateway is required") {
			t.Fatalf("%s: %s", c.name, body)
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
	if err := a.Note(Entry{User: "u", Brain: "b", Tool: "t", OK: true}, policy(t, "")); err != nil {
		t.Fatal(err)
	}
	_ = a.Close()
	a, err = OpenAudit(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Note(Entry{User: "u", Brain: "b", Tool: "t2", OK: false, Error: "x"}, policy(t, "")); err != nil {
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
	if err := a.Note(Entry{}, policy(t, "")); err == nil {
		t.Fatal("a closed audit accepted a record")
	}
	if err := a.Close(); err != nil {
		t.Fatalf("a second close: %v", err)
	}
}

// The relay reads the rights file again when it changes: a right withdrawn
// in the file is refused at the next request, without a restart, and a
// right added is served. A broken file keeps the previous rights.
func TestTheRelayFollowsTheRightsFile(t *testing.T) {
	b := start(t)
	path := filepath.Join(t.TempDir(), "rights.yaml")
	text := "brains:\n  team/acme:\n    addr: " + b.rights.Brains["team/acme"].Addr + "\n    keys_file: " + b.keysFile + "\n    collection: team-acme\n    readers: [user:mdupont]\n    writers: [user:lfoletti]\n"
	write := func(s string, when time.Time) {
		t.Helper()
		if err := os.WriteFile(path, []byte(s), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatal(err)
		}
	}
	t0 := time.Now().Add(-time.Hour)
	write(text, t0)
	src, err := rights.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a, err := OpenAudit(b.audit)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a.Close() }()
	ts := httptest.NewServer(New(src, a).Handler())
	defer ts.Close()
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	status := func() int {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/mcp/team/acme", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}`))
		req.Header.Set("Authorization", "Bearer "+k.Secret)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res.StatusCode
	}
	if got := status(); got != http.StatusOK {
		t.Fatalf("before: %d", got)
	}
	write(strings.Replace(text, "    writers: [user:lfoletti]\n", "", 1), t0.Add(time.Second))
	if got := status(); got != http.StatusForbidden {
		t.Fatalf("withdrawn: %d", got)
	}
	write("brains: [", t0.Add(2*time.Second))
	if got := status(); got != http.StatusForbidden {
		t.Fatalf("broken: %d", got)
	}
	write(text, t0.Add(3*time.Second))
	if got := status(); got != http.StatusOK {
		t.Fatalf("restored: %d", got)
	}
}

// Through the relay, a read of a brain without audit_reads leaves no named
// line: the audit counts it, and writes the count of the day at close.
func TestARelayedReadIsCountedWithoutAName(t *testing.T) {
	b := startWith(t, func(y string) string { return strings.ReplaceAll(y, "    audit_reads: true\n", "") })
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s, err := b.session(t, "team/acme", k.Secret)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "knowledge_get", Arguments: map[string]any{"knowledge_id": "doc-1"}}); err != nil {
		t.Fatal(err)
	}
	if lines := b.auditLines(t); len(lines) != 0 {
		t.Fatalf("named read: %+v", lines)
	}
	if err := b.a.Close(); err != nil {
		t.Fatal(err)
	}
	lines := b.auditLines(t)
	if len(lines) != 1 || lines[0]["tool"] != "reads" || lines[0]["brain"] != "team/acme" || lines[0]["calls"] != 1.0 || lines[0]["user"] != "" {
		t.Fatalf("count: %+v", lines)
	}
}

// A read the brain refuses to the key keeps a named line, even on a brain
// whose reads are counted: a refusal is a matter of security, and says
// nothing of the work of the person.
func TestARefusedReadIsNamed(t *testing.T) {
	b := startWith(t, func(y string) string { return strings.ReplaceAll(y, "    audit_reads: true\n", "") })
	b.brain.denyTool = "knowledge_get"
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s, err := b.session(t, "team/acme", k.Secret)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "knowledge_get", Arguments: map[string]any{"knowledge_id": "doc-1"}}); err != nil || !res.IsError {
		t.Fatalf("refused call: %+v, %v", res, err)
	}
	// The same through th_get, and through th_pending for a curator.
	withCurator(t, b)
	if res := mustCall(t, b.federated(t, k.Secret), "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "doc-1"}); !res.IsError {
		t.Fatal("th_get was not refused")
	}
	b.brain.denyTool = "aggregate_metadata"
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	if res := mustCall(t, b.federated(t, kc.Secret), "th_pending", map[string]any{"brain": "team/acme"}); !res.IsError {
		t.Fatal("th_pending was not refused")
	}
	lines := b.auditLines(t)
	if len(lines) != 3 {
		t.Fatalf("audit: %+v", lines)
	}
	for i, want := range []string{"lfoletti knowledge_get", "lfoletti th_get", "kmlead th_pending"} {
		l := lines[i]
		if fmt.Sprint(l["user"], " ", l["tool"]) != want || l["ok"] != false || !strings.Contains(l["error"].(string), "PermissionDenied") {
			t.Fatalf("line %d: %+v", i, l)
		}
	}
}

// A brain with curators has its review circuit in the gateway: th_add,
// th_pending and th_review on /mcp. Its per-brain endpoint would hand the
// engine's raw tools to a writer's key, which the engine lets accept a
// proposal, publish without review, replace or delete another author's
// document (essai du 24.10.2026 sur un cerveau local). It is refused to every
// key, before the brain is reached; a brain without curators keeps it.
func TestABrainWithCuratorsIsServedThroughTheFederatedEndpointOnly(t *testing.T) {
	b := start(t)
	withCurator(t, b)
	for _, c := range []struct {
		user   string
		access rights.Access
	}{{"mdupont", rights.Write}, {"kmlead", rights.Curate}} {
		k, _ := keys.Issue(b.keysFile, c.user, "team/acme", "team-acme", c.access)
		req, _ := http.NewRequest(http.MethodPost, b.url+"/mcp/team/acme", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
		req.Header.Set("Authorization", "Bearer "+k.Secret)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden || !strings.Contains(string(body), "review circuit") {
			t.Fatalf("%s: %d %s", c.user, resp.StatusCode, body)
		}
	}
	if len(b.brain.calls) != 0 {
		t.Fatalf("the brain was reached: %+v", b.brain.calls)
	}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	s, err := b.session(t, "team/other", k.Secret)
	if err != nil {
		t.Fatalf("a brain without curators: %v", err)
	}
	_ = s.Close()
}
