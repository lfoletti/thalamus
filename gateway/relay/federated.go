package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	rpcv1 "github.com/liliang-cn/cortexdb/v2/pkg/rpc/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/granular-cortex/gateway/keys"
	"github.com/lfoletti/granular-cortex/gateway/rights"
)

// The federated endpoint, /mcp: the caller presents any key the gateway
// issued, and gets two tools that work across every brain the caller may
// read. gc_brains lists them; gc_recall asks each one with the caller's own
// key for it and fuses the answers. The engine still enforces each key.

// rrfK is the constant of reciprocal rank fusion: score = weight / (k + rank).
const rrfK = 60

// admitFederated finds the caller behind a bearer, looking through the key
// files of every brain, and refuses a caller with no right left.
func (s *Server) admitFederated(r *http.Request) (*mcp.Server, int, error) {
	secret, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found || secret == "" {
		return nil, http.StatusUnauthorized, errors.New("a key issued by the gateway is required")
	}
	user := ""
	for _, name := range sortedBrains(s.rights) {
		u, _, err := keys.Holder(s.rights.Brains[name].KeysFile, secret)
		if err == nil {
			user = u
			break
		}
		if !strings.Contains(err.Error(), "unknown key") && !strings.Contains(err.Error(), "not issued") {
			return nil, http.StatusInternalServerError, fmt.Errorf("key file of %s: %w", name, err)
		}
	}
	if user == "" {
		return nil, http.StatusUnauthorized, errors.New("refused: unknown key")
	}
	if len(s.rights.BrainsOf(user)) == 0 {
		return nil, http.StatusForbidden, fmt.Errorf("%s no longer has access to any brain", user)
	}
	s.mu.Lock()
	srv, cached := s.sessions["federated:"+secret]
	s.mu.Unlock()
	if cached {
		return srv, 0, nil
	}
	srv = s.federatedServer(user)
	s.mu.Lock()
	s.sessions["federated:"+secret] = srv
	s.mu.Unlock()
	return srv, 0, nil
}

func sortedBrains(r *rights.Rights) []string {
	names := make([]string, 0, len(r.Brains))
	for n := range r.Brains {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

type brainEntry struct {
	Brain  string `json:"brain"`
	Access string `json:"access"`
	Key    bool   `json:"key"`
}

type recallArgs struct {
	Query  string   `json:"query"`
	TopK   int      `json:"top_k,omitempty"`
	Brains []string `json:"brains,omitempty"`
}

type fused struct {
	Brain       string   `json:"brain"`
	KnowledgeID string   `json:"knowledge_id"`
	Title       string   `json:"title,omitempty"`
	SourceURL   string   `json:"source_url,omitempty"`
	Author      string   `json:"author,omitempty"`
	Snippet     string   `json:"snippet,omitempty"`
	Score       float64  `json:"score"`
	Provenance  string   `json:"provenance"`
	AlsoIn      []string `json:"also_in,omitempty"`
}

type skipped struct {
	Brain  string `json:"brain"`
	Reason string `json:"reason"`
}

type recallResult struct {
	Results []fused   `json:"results"`
	Skipped []skipped `json:"skipped,omitempty"`
}

// hit is one answer of knowledge_search, as the engine returns it.
type hit struct {
	KnowledgeID string  `json:"knowledge_id"`
	Title       string  `json:"title"`
	SourceURL   string  `json:"source_url"`
	Author      string  `json:"author"`
	Snippet     string  `json:"snippet"`
	Score       float64 `json:"score"`
}

// brainHits is one brain's answer, with its weight.
type brainHits struct {
	Brain  string
	Weight float64
	Hits   []hit
}

func (s *Server) federatedServer(user string) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "gc-gateway", Title: "Granular Cortex gateway: federated", Version: "0"}, nil)
	srv.AddTool(&mcp.Tool{Name: "gc_brains", Description: "Lists the brains the caller may read, with the access (read or write) and whether a key was issued for it.", InputSchema: map[string]any{"type": "object"}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var out []brainEntry
			for _, g := range s.rights.BrainsOf(user) {
				_, err := keys.SecretOf(s.rights.Brains[g.Brain].KeysFile, user, g.Brain)
				out = append(out, brainEntry{Brain: g.Brain, Access: g.Access.String(), Key: err == nil})
			}
			_ = s.audit.Record(Entry{User: user, Tool: "gc_brains", OK: true})
			return jsonResult(out), nil
		})
	srv.AddTool(&mcp.Tool{Name: "gc_recall", Description: "Asks every brain the caller may read (or only those named in brains) with the caller's own key, and fuses the answers by weighted reciprocal rank. Returns the results with their provenance and the brains that were skipped.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"query":  map[string]any{"type": "string", "description": "the question"},
			"top_k":  map[string]any{"type": "integer", "description": "results to return, default 5"},
			"brains": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "restrict to these brains"},
		}, "required": []string{"query"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args recallArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("gc_recall: %w", err)), nil
			}
			res, err := s.recall(ctx, user, args)
			if err != nil {
				return errResult(err), nil
			}
			return jsonResult(res), nil
		})
	s.addWriteTools(srv, user)
	return srv
}

// recall fans the query out to the caller's brains and fuses the answers.
func (s *Server) recall(ctx context.Context, user string, args recallArgs) (recallResult, error) {
	if strings.TrimSpace(args.Query) == "" {
		return recallResult{}, errors.New("gc_recall: query is required")
	}
	if args.TopK <= 0 {
		args.TopK = 5
	}
	grants := s.rights.BrainsOf(user)
	if len(args.Brains) > 0 {
		readable := map[string]bool{}
		for _, g := range grants {
			readable[g.Brain] = true
		}
		var kept []rights.Grant
		for _, name := range args.Brains {
			if !readable[name] {
				return recallResult{}, fmt.Errorf("gc_recall: %s is not a brain %s may read", name, user)
			}
			kept = append(kept, rights.Grant{Brain: name})
		}
		grants = kept
	}

	var (
		mu     sync.Mutex
		wg     sync.WaitGroup
		lists  = make([]brainHits, len(grants))
		skips  []skipped
		read   []string
		brains []string
	)
	for i, g := range grants {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			b := s.rights.Brains[name]
			hits, err := s.askBrain(ctx, b, user, name, args)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				skips = append(skips, skipped{Brain: name, Reason: err.Error()})
				return
			}
			lists[i] = brainHits{Brain: name, Weight: b.Weight(), Hits: hits}
			brains = append(brains, name)
			for _, h := range hits {
				read = append(read, name+":"+h.KnowledgeID)
			}
		}(i, g.Brain)
	}
	wg.Wait()
	sort.Slice(skips, func(i, j int) bool { return skips[i].Brain < skips[j].Brain })
	sort.Strings(brains)
	sort.Strings(read)
	entry := Entry{User: user, Tool: "gc_recall", OK: true, IDs: map[string]any{"brains": brains, "knowledge_ids": read}}
	_ = s.audit.Record(entry)
	return recallResult{Results: fuse(lists, args.TopK), Skipped: skips}, nil
}

// callBrain calls one tool on one brain with the user's key for it. The
// brain's collection is added to the arguments when the brain has one,
// unless the caller set "collection" to nil: that tool takes none.
func (s *Server) callBrain(ctx context.Context, b rights.Brain, user, name, tool string, req map[string]any) (string, error) {
	secret, err := keys.SecretOf(b.KeysFile, user, name)
	if err != nil {
		return "", err
	}
	conn, err := dial(b.Addr, secret)
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()
	if v, set := req["collection"]; set && v == nil {
		delete(req, "collection")
	} else if b.Collection != "" {
		req["collection"] = b.Collection
	}
	raw, _ := json.Marshal(req)
	resp, err := rpcv1.NewToolsServiceClient(conn).CallTool(ctx, &rpcv1.CallToolRequest{Name: tool, ArgsJson: string(raw)})
	if err != nil {
		return "", err
	}
	return resp.GetResultJson(), nil
}

// askBrain calls knowledge_search on one brain with the user's key for it.
func (s *Server) askBrain(ctx context.Context, b rights.Brain, user, name string, args recallArgs) ([]hit, error) {
	result, err := s.callBrain(ctx, b, user, name, "knowledge_search", map[string]any{"query": args.Query, "top_k": args.TopK})
	if err != nil {
		return nil, err
	}
	var out struct {
		Results []hit `json:"results"`
	}
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("decode knowledge_search: %w", err)
	}
	return out.Results, nil
}

// fuse merges the brains' lists by weighted reciprocal rank. The same
// snippet in two brains is one result, scored by both, that names the other
// brains in also_in. Ties keep the order of the lists.
func fuse(lists []brainHits, topK int) []fused {
	var out []fused
	bySnippet := map[string]int{}
	for _, l := range lists {
		for rank, h := range l.Hits {
			score := l.Weight / float64(rrfK+rank+1)
			key := strings.TrimSpace(h.Snippet)
			if i, seen := bySnippet[key]; seen && key != "" {
				out[i].Score += score
				out[i].AlsoIn = append(out[i].AlsoIn, l.Brain)
				continue
			}
			f := fused{Brain: l.Brain, KnowledgeID: h.KnowledgeID, Title: h.Title, SourceURL: h.SourceURL, Author: h.Author, Snippet: h.Snippet, Score: score}
			f.Provenance = provenance(l.Brain, h)
			if key != "" {
				bySnippet[key] = len(out)
			}
			out = append(out, f)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > topK {
		out = out[:topK]
	}
	return out
}

// provenance is the fixed line an agent and a person both read:
// [brain] title, id, source, by author.
func provenance(brain string, h hit) string {
	parts := []string{h.Title, h.KnowledgeID}
	if h.Title == "" {
		parts = parts[1:]
	}
	if h.SourceURL != "" {
		parts = append(parts, h.SourceURL)
	}
	if h.Author != "" {
		parts = append(parts, "by "+h.Author)
	}
	return "[" + brain + "] " + strings.Join(parts, ", ")
}

func jsonResult(v any) *mcp.CallToolResult {
	data, _ := json.Marshal(v)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}
}

func errResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}
}
