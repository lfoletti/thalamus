package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	rpcv1 "github.com/liliang-cn/cortexdb/v2/pkg/rpc/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// The federated endpoint, /mcp: the caller presents any key the gateway
// issued, and gets two tools that work across every brain the caller may
// read. th_brains lists them; th_recall asks each one with the caller's own
// key for it and fuses the answers. The engine still enforces each key.

// rrfK is the constant of reciprocal rank fusion: score = weight / (k + rank).
const rrfK = 60

// defaultTopK is how many chunks th_recall asks of each brain, and
// chunkChars the room it gives each one. Measured on the ai-audio pilot
// (22 October 2026): 4 syntheses then 10 chunks hold 81 % of the facts in
// 9 000 characters, where the snippets of 10 documents hold 54 %.
const (
	defaultTopK = 10
	chunkChars  = 1000
)

// maxVariants bounds alternate_queries: each one is a search of its own on
// every brain, in the passages and in the pyramid.
const maxVariants = 4

// holderOf finds the caller behind a bearer, looking through the key files
// of every brain, and refuses a caller with no right left.
func (s *Server) holderOf(r *http.Request) (string, int, error) {
	secret, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found || secret == "" {
		return "", http.StatusUnauthorized, errors.New("a key issued by the gateway is required")
	}
	user := ""
	for _, name := range sortedBrains(s.rights()) {
		u, _, err := keys.Holder(s.rights().Brains[name].KeysFile, secret)
		if err == nil {
			user = u
			break
		}
		if !strings.Contains(err.Error(), "unknown key") && !strings.Contains(err.Error(), "not issued") {
			return "", http.StatusInternalServerError, fmt.Errorf("key file of %s: %w", name, err)
		}
	}
	if user == "" {
		return "", http.StatusUnauthorized, errors.New("refused: unknown key")
	}
	if len(s.rights().BrainsOf(user)) == 0 {
		return "", http.StatusForbidden, fmt.Errorf("%s no longer has access to any brain", user)
	}
	return user, 0, nil
}

// admitFederated serves the federated MCP server of the caller, one per key.
func (s *Server) admitFederated(r *http.Request) (*mcp.Server, int, error) {
	user, status, err := s.holderOf(r)
	if err != nil {
		return nil, status, err
	}
	secret := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
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

// brainsOf lists the brains the user may read, with the access and whether
// a key exists, and audits the listing.
func (s *Server) brainsOf(user string) []brainEntry {
	var out []brainEntry
	for _, g := range s.rights().BrainsOf(user) {
		_, err := keys.SecretOf(s.rights().Brains[g.Brain].KeysFile, user, g.Brain)
		out = append(out, brainEntry{Brain: g.Brain, Access: g.Access.String(), Key: err == nil})
	}
	s.note(Entry{User: user, Tool: "th_brains", OK: true})
	return out
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
	Skip   []string `json:"skip_brains,omitempty"`
	// Variants are the query in other words or languages, each searched
	// apart and fused with it.
	Variants []string `json:"alternate_queries,omitempty"`
	// Syntheses is how many syntheses of the pyramids to serve first; nil
	// means defaultSyntheses, 0 none.
	Syntheses *int `json:"syntheses,omitempty"`
	// DeadlineMS is how long th_recall waits for the brains, in
	// milliseconds; 0 means it waits for each one.
	DeadlineMS int `json:"deadline_ms,omitempty"`
}

// maxDeadlineMS bounds deadline_ms: a minute.
const maxDeadlineMS = 60000

// recallDeadline reads deadline_ms: 0 is no deadline, and a value out of
// 0 to maxDeadlineMS is an error.
func recallDeadline(ms int) (time.Duration, error) {
	if ms < 0 || ms > maxDeadlineMS {
		return 0, fmt.Errorf("th_recall: deadline_ms must be between 0 (none) and %d", maxDeadlineMS)
	}
	return time.Duration(ms) * time.Millisecond, nil
}

// brainAnswer is what one brain of a recall sends: its lists, its skipped
// lines, and the refusal of its key.
type brainAnswer struct {
	i            int
	name         string
	mine, theirs []brainHits
	skips        []skipped
	refusal      string
}

// gather reads the answers of n brains, until each one answered or late is
// closed. Past late, it keeps the answers already sent and waits for no
// other.
func gather(answers <-chan brainAnswer, n int, late <-chan struct{}) []brainAnswer {
	var got []brainAnswer
	for len(got) < n {
		select {
		case a := <-answers:
			got = append(got, a)
		case <-late:
			select {
			case a := <-answers:
				got = append(got, a)
			default:
				return got
			}
		}
	}
	return got
}

// lateBrains is a skipped line for each brain asked that did not answer
// within the deadline, in the order asked.
func lateBrains(asked []string, answered map[string]bool, deadline time.Duration) []skipped {
	var late []skipped
	for _, n := range asked {
		if !answered[n] {
			late = append(late, skipped{Brain: n, Reason: fmt.Sprintf("no answer within %d ms", deadline.Milliseconds())})
		}
	}
	return late
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
	// Level is "synthesis" for a synthesis of a brain's pyramid.
	Level string `json:"level,omitempty"`
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
	srv := mcp.NewServer(&mcp.Implementation{Name: "th-gateway", Title: "Thalamus gateway: federated", Version: "0"}, nil)
	srv.AddTool(&mcp.Tool{Name: "th_brains", Description: "Lists the brains the caller may read, with the access (read or write) and whether a key was issued for it.", InputSchema: map[string]any{"type": "object"}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(s.brainsOf(user)), nil
		})
	srv.AddTool(&mcp.Tool{Name: "th_recall", Description: "Asks every brain the caller may read (or only those named in brains, less those in skip_brains) with the caller's own key, and fuses the answers by weighted reciprocal rank. Each result is a passage of a document. Returns the results with their provenance and the brains that were skipped.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"query":       map[string]any{"type": "string", "description": "the question"},
			"top_k":       map[string]any{"type": "integer", "description": fmt.Sprintf("passages to return, default %d", defaultTopK)},
			"brains":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "restrict to these brains"},
			"skip_brains": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "leave out these brains"},
			"syntheses":   map[string]any{"type": "integer", "description": fmt.Sprintf("passages of the brains' pyramids of syntheses to serve first, default %d, 0 for none", defaultSyntheses)},
			"alternate_queries": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": maxVariants,
				"description": fmt.Sprintf("the question in other words or other languages of the documents, at most %d; each is searched apart and fused with the query", maxVariants)},
			"deadline_ms": map[string]any{"type": "integer", "minimum": 0, "maximum": maxDeadlineMS,
				"description": fmt.Sprintf("how long to wait for the brains, in milliseconds, at most %d; at the deadline the answers received are fused and each brain still asked is skipped; default 0, wait for each brain", maxDeadlineMS)},
		}, "required": []string{"query"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args recallArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_recall: %w", err)), nil
			}
			res, err := s.recall(ctx, user, args)
			if err != nil {
				return errResult(err), nil
			}
			return jsonResult(res), nil
		})
	s.addWriteTools(srv, user)
	s.addPyramidTools(srv, user)
	s.addEntityTools(srv, user)
	return srv
}

// recall fans the query out to the caller's brains and fuses the answers.
func (s *Server) recall(ctx context.Context, user string, args recallArgs) (recallResult, error) {
	if strings.TrimSpace(args.Query) == "" {
		return recallResult{}, errors.New("th_recall: query is required")
	}
	if args.TopK <= 0 {
		args.TopK = defaultTopK
	}
	if len(args.Variants) > maxVariants {
		return recallResult{}, fmt.Errorf("th_recall: at most %d alternate_queries", maxVariants)
	}
	queries := append([]string{args.Query}, args.Variants...)
	syntheses := defaultSyntheses
	if args.Syntheses != nil {
		syntheses = *args.Syntheses
	}
	deadline, err := recallDeadline(args.DeadlineMS)
	if err != nil {
		return recallResult{}, err
	}
	grants, err := s.grantsFor(user, args.Brains, args.Skip, "th_recall")
	if err != nil {
		return recallResult{}, err
	}

	// Each brain sends its answer on a channel with room for all: a brain
	// that answers after the deadline writes nothing that is read, and its
	// call ends with ctx.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	answers := make(chan brainAnswer, len(grants))
	asked := make([]string, len(grants))
	for i, g := range grants {
		asked[i] = g.Brain
		go func(i int, name string) {
			a := brainAnswer{i: i, name: name}
			defer func() { answers <- a }()
			b := s.rights().Brains[name]
			for q, query := range queries {
				hits, err := s.askBrain(ctx, b, user, name, query, args.TopK, b.Collection)
				var syn []hit
				var synErr error
				if err == nil && b.Pyramid != nil && syntheses > 0 {
					syn, synErr = s.askBrain(ctx, b, user, name, query, syntheses, b.Pyramid.Collection)
				}
				prefix := ""
				if q > 0 {
					prefix = fmt.Sprintf("alternate query %d: ", q)
				}
				if err != nil {
					if q == 0 {
						a.skips = []skipped{{Brain: name, Reason: err.Error()}}
						if denied(err) {
							a.refusal = err.Error()
						}
						return
					}
					a.skips = append(a.skips, skipped{Brain: name, Reason: prefix + err.Error()})
					continue
				}
				if synErr != nil {
					a.skips = append(a.skips, skipped{Brain: name, Reason: prefix + "pyramid: " + synErr.Error()})
				}
				a.mine = append(a.mine, brainHits{Brain: name, Weight: b.Weight(), Hits: hits})
				a.theirs = append(a.theirs, brainHits{Brain: name, Weight: b.Weight(), Hits: syn})
			}
		}(i, g.Brain)
	}

	// late is closed at the deadline and stays closed.
	var late <-chan struct{}
	if deadline > 0 {
		lateCtx, stop := context.WithTimeout(context.Background(), deadline)
		defer stop()
		late = lateCtx.Done()
	}
	var (
		lists    = make([][]brainHits, len(grants))
		synths   = make([][]brainHits, len(grants))
		skips    []skipped
		read     []string
		brains   []string
		refusals = map[string]string{}
		answered = map[string]bool{}
	)
	for _, a := range gather(answers, len(grants), late) {
		answered[a.name] = true
		skips = append(skips, a.skips...)
		if a.refusal != "" {
			refusals[a.name] = a.refusal
		}
		if a.mine == nil {
			continue
		}
		lists[a.i], synths[a.i] = a.mine, a.theirs
		brains = append(brains, a.name)
		for _, l := range append(a.mine, a.theirs...) {
			for _, h := range l.Hits {
				read = append(read, a.name+":"+h.KnowledgeID)
			}
		}
	}
	skips = append(skips, lateBrains(asked, answered, deadline)...)
	sort.SliceStable(skips, func(i, j int) bool { return skips[i].Brain < skips[j].Brain })
	sort.Strings(brains)
	sort.Strings(read)
	read = slices.Compact(read)
	entry := Entry{User: user, Tool: "th_recall", OK: true, IDs: map[string]any{"brains": brains, "knowledge_ids": read, "denied": refusals}}
	s.note(entry)
	results := append(fuse(slices.Concat(synths...), syntheses, "synthesis"), fuse(slices.Concat(lists...), args.TopK, "")...)
	return recallResult{Results: results, Skipped: skips}, nil
}

// grantsFor chooses the brains a federated tool asks: every brain the
// caller may read, or only those named in brains, less those in skip. A
// named brain the caller may not read is refused, and a brain named twice
// is asked once.
func (s *Server) grantsFor(user string, brains, skip []string, tool string) ([]rights.Grant, error) {
	grants := s.rights().BrainsOf(user)
	if len(brains) > 0 {
		readable := map[string]bool{}
		for _, g := range grants {
			readable[g.Brain] = true
		}
		var kept []rights.Grant
		for _, name := range brains {
			if !readable[name] {
				return nil, fmt.Errorf("%s: %s is not a brain %s may read", tool, name, user)
			}
			if !slices.ContainsFunc(kept, func(g rights.Grant) bool { return g.Brain == name }) {
				kept = append(kept, rights.Grant{Brain: name})
			}
		}
		grants = kept
	}
	return without(grants, skip), nil
}

// without leaves out of grants the brains named in skip.
func without(grants []rights.Grant, skip []string) []rights.Grant {
	out := []rights.Grant{}
	for _, g := range grants {
		if !slices.Contains(skip, g.Brain) {
			out = append(out, g)
		}
	}
	return out
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
	} else if !set && b.Collection != "" {
		req["collection"] = b.Collection
	}
	// Encoded from a map, the arguments name each key once: the caller's
	// JSON is never relayed as it came.
	raw, _ := json.Marshal(req)
	resp, err := rpcv1.NewToolsServiceClient(conn).CallTool(ctx, &rpcv1.CallToolRequest{Name: tool, ArgsJson: string(raw)})
	if err != nil {
		return "", err
	}
	return resp.GetResultJson(), nil
}

// askBrain calls knowledge_search on one collection of a brain (none: the
// whole server) with the user's key for it.
func (s *Server) askBrain(ctx context.Context, b rights.Brain, user, name, query string, topK int, collection string) ([]hit, error) {
	req := map[string]any{"query": query, "top_k": topK, "retrieval_mode": "graph", "max_context_chunks": topK, "max_context_chars": chunkChars * topK}
	if collection != "" {
		req["collection"] = collection
	}
	result, err := s.callBrain(ctx, b, user, name, "knowledge_search", req)
	if err != nil {
		return nil, err
	}
	var out struct {
		Results []hit `json:"results"`
		Chunks  []struct {
			DocumentID string
			Content    string
		} `json:"chunks"`
	}
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("decode knowledge_search: %w", err)
	}
	if len(out.Chunks) == 0 {
		return out.Results, nil
	}
	docs := map[string]hit{}
	for _, h := range out.Results {
		docs[h.KnowledgeID] = h
	}
	hits := make([]hit, 0, len(out.Chunks))
	for _, c := range out.Chunks {
		h := docs[c.DocumentID]
		hits = append(hits, hit{KnowledgeID: c.DocumentID, Title: h.Title, SourceURL: h.SourceURL, Author: h.Author, Snippet: c.Content})
	}
	return hits, nil
}

// fuse merges the brains' lists by weighted reciprocal rank. The same
// snippet in two brains is one result, scored by both, that names the other
// brains in also_in. Ties keep the order of the lists. Every result gets
// the level, "" for a passage.
func fuse(lists []brainHits, topK int, level string) []fused {
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
			f := fused{Brain: l.Brain, KnowledgeID: h.KnowledgeID, Title: h.Title, SourceURL: h.SourceURL, Author: h.Author, Snippet: h.Snippet, Score: score, Level: level}
			f.Provenance = provenance(l.Brain, h, level)
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
// [brain] title, id, source, by author; "[brain] synthesis: …" for a
// synthesis.
func provenance(brain string, h hit, level string) string {
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
	if level != "" {
		return "[" + brain + "] " + level + ": " + strings.Join(parts, ", ")
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
