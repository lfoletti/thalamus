package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeGateway is the federated endpoint of a gateway: it checks the bearer
// and answers th_recall with canned results.
type fakeGateway struct {
	secret  string
	results []map[string]any
	skipped []map[string]any
	fail    bool
	badJSON bool // answer with text that is not JSON
	noTool  bool // a gateway without th_recall
	odd     int  // 1: an answer with no content, 2: an answer that is not text
	calls   []map[string]any
	listing []map[string]any // what th_brains answers
}

func (g *fakeGateway) start(t *testing.T) string {
	t.Helper()
	srv := mcp.NewServer(&mcp.Implementation{Name: "fake"}, nil)
	name := "th_recall"
	if g.noTool {
		name = "th_other"
	}
	srv.AddTool(&mcp.Tool{Name: name, InputSchema: map[string]any{"type": "object"}}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args map[string]any
		_ = json.Unmarshal(req.Params.Arguments, &args)
		g.calls = append(g.calls, args)
		if g.fail {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "th_recall: boom"}}}, nil
		}
		if g.badJSON {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "nope"}}}, nil
		}
		switch g.odd {
		case 1:
			return &mcp.CallToolResult{}, nil
		case 2:
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.ImageContent{Data: []byte("x"), MIMEType: "image/png"}}}, nil
		}
		out, _ := json.Marshal(map[string]any{"results": g.results, "skipped": g.skipped})
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(out)}}}, nil
	})
	srv.AddTool(&mcp.Tool{Name: "th_brains", InputSchema: map[string]any{"type": "object"}}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		out, _ := json.Marshal(g.listing)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(out)}}}, nil
	})
	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, &mcp.StreamableHTTPOptions{JSONResponse: true})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp" || r.Header.Get("Authorization") != "Bearer "+g.secret {
			http.Error(w, "refused", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(ts.Close)
	return ts.URL
}

func hook(t *testing.T, env map[string]string, stdin string, args ...string) (code int, out, errOut string) {
	t.Helper()
	var o, e bytes.Buffer
	if args == nil {
		args = []string{"recall"}
	}
	code = run(args, strings.NewReader(stdin), func(k string) string { return env[k] }, &o, &e)
	return code, o.String(), e.String()
}

func contextOf(t *testing.T, out string) string {
	t.Helper()
	var v struct {
		H struct {
			Event string `json:"hookEventName"`
			Ctx   string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("output %q: %v", out, err)
	}
	if v.H.Event != "UserPromptSubmit" {
		t.Fatalf("event %q", v.H.Event)
	}
	return v.H.Ctx
}

// The hook asks the gateway with the prompt and the key, and hands Claude
// Code the extracts with their provenance, the skipped brains, and a note
// that the extracts come from the shared brains.
func TestTheHookInjectsTheExtractsWithTheirProvenance(t *testing.T) {
	g := &fakeGateway{secret: "s3", results: []map[string]any{
		{"brain": "team/acme", "knowledge_id": "spec-1", "snippet": "The ledger moves on Sunday.", "provenance": "[team/acme] Spec migration, spec-1, by mdupont", "also_in": []string{"org/reference"}},
		{"brain": "org/reference", "knowledge_id": "r-1", "snippet": "Rollback is two hours.", "provenance": "[org/reference] Rollback, r-1"},
	}, skipped: []map[string]any{{"brain": "team/other", "reason": "no key issued to lfoletti for team/other"}}}
	url := g.start(t)
	code, out, errOut := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"when does the ledger move?","cwd":"/x"}`)
	if code != 0 || errOut != "" {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	ctx := contextOf(t, out)
	for _, want := range []string{
		"Thalamus: 2 extracts from the shared brains",
		"- [team/acme] Spec migration, spec-1, by mdupont (also in org/reference): The ledger moves on Sunday.",
		"- [org/reference] Rollback, r-1: Rollback is two hours.",
		"Skipped: team/other (no key issued to lfoletti for team/other).",
		"An extract is a cut: read a document whole with th_get {brain, knowledge_id} from its provenance line. Ask for more with th_recall; list the brains with th_brains.",
	} {
		if !strings.Contains(ctx, want) {
			t.Fatalf("context lacks %q:\n%s", want, ctx)
		}
	}
	if len(g.calls) != 1 || g.calls[0]["query"] != "when does the ledger move?" || g.calls[0]["top_k"] != float64(10) {
		t.Fatalf("calls: %+v", g.calls)
	}
	// TH_RECALL_TOPK sets the number asked.
	hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_RECALL_TOPK": "7"}, `{"prompt":"ledger rollback window"}`)
	if g.calls[1]["top_k"] != float64(7) {
		t.Fatalf("top_k: %+v", g.calls[1])
	}
}

// The hook never blocks a prompt: every problem is exit 0 with no output,
// and the reason on stderr. No results is no output either.
func TestTheHookIsSilentOnEveryProblem(t *testing.T) {
	g := &fakeGateway{secret: "s3"}
	url := g.start(t)
	for _, c := range []struct {
		name  string
		env   map[string]string
		stdin string
		want  string
	}{
		{"no gateway", map[string]string{"TH_KEY": "s3"}, `{"prompt":"a long enough prompt"}`, "TH_GATEWAY"},
		{"no key", map[string]string{"TH_GATEWAY": url}, `{"prompt":"a long enough prompt"}`, "TH_KEY"},
		{"bad key", map[string]string{"TH_GATEWAY": url, "TH_KEY": "wrong"}, `{"prompt":"a long enough prompt"}`, "Unauthorized"},
		{"gateway down", map[string]string{"TH_GATEWAY": "http://127.0.0.1:1", "TH_KEY": "s3"}, `{"prompt":"a long enough prompt"}`, "connect"},
		{"bad stdin", map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{`, "hook input"},
		{"empty prompt", map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"  "}`, ""},
		{"short prompt", map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"ok go"}`, ""},
		{"slash command", map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"/th-brains please now"}`, ""},
		{"no results", map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"nothing matches this one"}`, ""},
		{"bad top_k", map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_RECALL_TOPK": "x"}, `{"prompt":"a long enough prompt"}`, "TH_RECALL_TOPK"},
	} {
		code, out, errOut := hook(t, c.env, c.stdin)
		if code != 0 || out != "" || !strings.Contains(errOut, c.want) {
			t.Fatalf("%s: code %d, stdout %q, stderr %q", c.name, code, out, errOut)
		}
	}
	// Of all these, only "no results" reached the gateway: an empty, short
	// or slash prompt asks nothing.
	if len(g.calls) != 1 {
		t.Fatalf("the gateway was asked %d times: %+v", len(g.calls), g.calls)
	}
	g.fail = true
	if code, out, errOut := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"a long enough prompt"}`); code != 0 || out != "" || !strings.Contains(errOut, "boom") {
		t.Fatalf("tool error: code %d, stdout %q, stderr %q", code, out, errOut)
	}
	g.fail, g.badJSON = false, true
	if code, out, errOut := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"a long enough prompt"}`); code != 0 || out != "" || !strings.Contains(errOut, "decode th_recall") {
		t.Fatalf("bad json: code %d, stdout %q, stderr %q", code, out, errOut)
	}
	for odd := 1; odd <= 2; odd++ {
		g.badJSON, g.odd = false, odd
		if code, out, errOut := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"a long enough prompt"}`); code != 0 || out != "" || !strings.Contains(errOut, "decode th_recall") {
			t.Fatalf("odd answer %d: code %d, stdout %q, stderr %q", odd, code, out, errOut)
		}
	}
	other := &fakeGateway{secret: "s3", noTool: true}
	if code, out, errOut := hook(t, map[string]string{"TH_GATEWAY": other.start(t), "TH_KEY": "s3"}, `{"prompt":"a long enough prompt"}`); code != 0 || out != "" || !strings.Contains(errOut, "th_recall") {
		t.Fatalf("no tool: code %d, stdout %q, stderr %q", code, out, errOut)
	}
	if code, _, errOut := hook(t, nil, "", "other"); code != 2 || !strings.Contains(errOut, "usage") {
		t.Fatalf("usage: code %d, stderr %q", code, errOut)
	}
}

// An oversized context is cut whole extracts at a time, from the end, and
// says how many it left out; one extract too long for the budget on its own
// gives no context at all.
func TestTheContextFitsTheBudget(t *testing.T) {
	long := strings.Repeat("x", 2000)
	var results []map[string]any
	for i := 0; i < 8; i++ {
		results = append(results, map[string]any{"brain": "b", "knowledge_id": fmt.Sprint("d-", i), "snippet": long, "provenance": fmt.Sprint("[b] d-", i)})
	}
	g := &fakeGateway{secret: "s3", results: results}
	url := g.start(t)
	_, out, _ := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"a long enough prompt"}`)
	ctx := contextOf(t, out)
	if len(ctx) > contextBudget || !strings.Contains(ctx, "d-3") || strings.Contains(ctx, "d-4") || !strings.Contains(ctx, "(4 more not shown; ask for them with th_recall.)") {
		t.Fatalf("len %d:\n%s", len(ctx), ctx[:200])
	}
	g.results = []map[string]any{{"brain": "b", "knowledge_id": "d", "snippet": strings.Repeat("y", contextBudget), "provenance": "[b] d"}}
	if _, out, errOut := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"a long enough prompt"}`); out != "" || !strings.Contains(errOut, "budget") {
		t.Fatalf("one too long: stdout %q, stderr %q", out, errOut)
	}
}

// Without TH_GATEWAY and TH_KEY in the environment, the hook reads the
// env file that th-gateway login writes: TH_ENV, or ~/.thalamus/env. The
// environment wins over the file, field by field.
func TestTheHookReadsTheEnvFile(t *testing.T) {
	g := &fakeGateway{secret: "s3", results: []map[string]any{{"brain": "team/acme", "knowledge_id": "spec-1", "snippet": "The code word is MARMOT.", "provenance": "[team/acme] Spec, spec-1"}}}
	url := g.start(t)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".thalamus"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".thalamus", "env"), []byte("# written by th-gateway login\nTH_GATEWAY="+url+"/\nTH_KEY = s3\nJUNK\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := hook(t, map[string]string{"HOME": dir}, `{"prompt":"when does the ledger move?","cwd":"/x"}`)
	if code != 0 || errOut != "" || !strings.Contains(contextOf(t, out), "MARMOT") {
		t.Fatalf("home file: %d %s %s", code, out, errOut)
	}
	other := filepath.Join(dir, "other.env")
	if err := os.WriteFile(other, []byte("TH_GATEWAY=http://127.0.0.1:1\nTH_KEY=s3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, errOut = hook(t, map[string]string{"TH_ENV": other, "TH_GATEWAY": url}, `{"prompt":"when does the ledger move?"}`)
	if code != 0 || errOut != "" || !strings.Contains(contextOf(t, out), "MARMOT") {
		t.Fatalf("TH_ENV and environment: %d %s %s", code, out, errOut)
	}
	if _, _, errOut = hook(t, map[string]string{"TH_ENV": filepath.Join(dir, "missing.env"), "TH_KEY": "wrong"}, `{"prompt":"when does the ledger move?"}`); !strings.Contains(errOut, "TH_GATEWAY is not set, and no env file") {
		t.Fatalf("missing gateway: %s", errOut)
	}
	if _, _, errOut = hook(t, map[string]string{"HOME": t.TempDir(), "TH_GATEWAY": url}, `{"prompt":"when does the ledger move?"}`); !strings.Contains(errOut, "TH_KEY is not set, and no env file") {
		t.Fatalf("missing key: %s", errOut)
	}
}

// A synthesis of a pyramid comes first, as the gateway orders it, and the
// footer then says how to open its sources. Without a synthesis, the footer
// does not mention th_descend.
func TestTheHookTellsHowToOpenASynthesis(t *testing.T) {
	g := &fakeGateway{secret: "s3", results: []map[string]any{
		{"brain": "team/acme", "knowledge_id": "pyramide/ledger", "level": "synthesis", "snippet": "The ledger moves on Sunday.", "provenance": "[team/acme] synthesis: Ledger, pyramide/ledger"},
		{"brain": "team/acme", "knowledge_id": "spec-1", "snippet": "Rollback is two hours.", "provenance": "[team/acme] Spec, spec-1"},
	}}
	url := g.start(t)
	_, out, _ := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"when does the ledger move?"}`)
	ctx := contextOf(t, out)
	want := "A synthesis cites its sources: open them with th_descend {brain, knowledge_id}."
	if !strings.Contains(ctx, want) || strings.Index(ctx, "pyramide/ledger") > strings.Index(ctx, "spec-1") {
		t.Fatalf("context:\n%s", ctx)
	}
	g.results = g.results[1:]
	_, out, _ = hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3"}, `{"prompt":"when does the ledger move?"}`)
	if strings.Contains(contextOf(t, out), "th_descend") {
		t.Fatalf("no synthesis, no th_descend:\n%s", out)
	}
}
