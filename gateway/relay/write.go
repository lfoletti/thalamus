package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lfoletti/thalamus/gateway/convention"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// The write circuit of the federated endpoint. th_add proposes a document
// into a brain the caller may write: it lands with review_status proposed
// (the explicit field of ingest_document, never a metadata key),
// and is served only once a curator accepts it, when the brain's server
// runs with the review gate. A curator's own document is accepted at once.
// th_pending lists what waits in a brain the caller curates; th_review
// decides. The engine records the calling key as the reviewer whatever the
// client sends; the gateway sends the user's name, which is what the key
// stands for.

type addArgs struct {
	Brain      string            `json:"brain"`
	DocumentID string            `json:"document_id,omitempty"`
	Title      string            `json:"title,omitempty"`
	Content    string            `json:"content"`
	SourceURL  string            `json:"source_url,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type pendingArgs struct {
	Brain string `json:"brain"`
	TopK  int    `json:"top_k,omitempty"`
}

type getArgs struct {
	Brain       string `json:"brain"`
	KnowledgeID string `json:"knowledge_id"`
}

type reviewArgs struct {
	Brain       string `json:"brain"`
	KnowledgeID string `json:"knowledge_id"`
	Decision    string `json:"decision"`
}

func (s *Server) addWriteTools(srv *mcp.Server, user string) {
	srv.AddTool(&mcp.Tool{Name: "th_add", Description: "Proposes a document into a brain the caller may write. It waits for a curator's review; a curator's own document is accepted at once. In a brain with a convention, metadata theme and nature come from the brain's lists, a new document_id is <nature>:<theme>:<subject>, and a near copy of a document the brain has is refused, naming it. Returns the engine's ingest report.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"brain":       map[string]any{"type": "string"},
			"document_id": map[string]any{"type": "string", "description": "stable id; derived from the time when absent, except in a brain with a convention, which needs one"},
			"title":       map[string]any{"type": "string"},
			"content":     map[string]any{"type": "string"},
			"source_url":  map[string]any{"type": "string"},
			"metadata":    map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}},
		}, "required": []string{"brain", "content"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args addArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_add: %w", err)), nil
			}
			return s.add(ctx, user, args), nil
		})
	srv.AddTool(&mcp.Tool{Name: "th_get", Description: "Reads one document, by brain and knowledge_id (both are on the provenance line of a th_recall result or of the hook's extract), in a brain the caller may read. In a brain served in excerpts, a source document is cut to its first words, with the address of the original.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"brain":        map[string]any{"type": "string"},
			"knowledge_id": map[string]any{"type": "string"},
		}, "required": []string{"brain", "knowledge_id"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args getArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_get: %w", err)), nil
			}
			return s.get(ctx, user, args), nil
		})
	srv.AddTool(&mcp.Tool{Name: "th_pending", Description: "Lists the documents that wait for review in a brain the caller curates.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"brain": map[string]any{"type": "string"},
			"top_k": map[string]any{"type": "integer", "description": "default 20"},
		}, "required": []string{"brain"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args pendingArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_pending: %w", err)), nil
			}
			return s.pending(ctx, user, args), nil
		})
	srv.AddTool(&mcp.Tool{Name: "th_review", Description: "Accepts or rejects a proposed document, as a curator of the brain.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"brain":        map[string]any{"type": "string"},
			"knowledge_id": map[string]any{"type": "string"},
			"decision":     map[string]any{"type": "string", "enum": []string{"accepted", "rejected"}},
		}, "required": []string{"brain", "knowledge_id", "decision"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args reviewArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_review: %w", err)), nil
			}
			return s.review(ctx, user, args), nil
		})
}

// brainFor checks the brain exists and the caller has at least the access.
func (s *Server) brainFor(user, name string, need rights.Access, tool string) (rights.Brain, error) {
	b, ok := s.rights().Brains[name]
	if !ok {
		return rights.Brain{}, fmt.Errorf("%s: unknown brain %q", tool, name)
	}
	access, _ := s.rights().Access(user, name)
	switch {
	case need == rights.Read && !access.AllowsRead():
		return rights.Brain{}, fmt.Errorf("%s: %s may not read %s", tool, user, name)
	case need == rights.Curate && access < rights.Curate:
		return rights.Brain{}, fmt.Errorf("%s: %s is not a curator of %s", tool, user, name)
	case need == rights.Write && !access.AllowsWrite():
		return rights.Brain{}, fmt.Errorf("%s: %s may not write to %s", tool, user, name)
	}
	return b, nil
}

// audited calls the brain and writes the audit line, success or failure.
func (s *Server) audited(ctx context.Context, b rights.Brain, user, name, tool, engineTool string, req map[string]any, ids map[string]any) *mcp.CallToolResult {
	entry := Entry{User: user, Brain: name, Tool: tool, IDs: ids}
	result, err := s.callBrain(ctx, b, user, name, engineTool, req)
	if err != nil {
		entry.Error = err.Error()
		entry.Denied = denied(err)
		s.note(entry)
		return errResult(fmt.Errorf("%s: %w", tool, err))
	}
	entry.OK = true
	s.note(entry)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}
}

func (s *Server) add(ctx context.Context, user string, args addArgs) *mcp.CallToolResult {
	res, _ := s.write(ctx, user, args)
	return res
}

// write is th_add, and POST /v1/add: it returns the result and, when the
// write is refused, why, so that REST can tell whose fault it is by its
// kind (a *convention.Refusal, a *nearCopy) rather than its words. The id
// is read once, for the owner check and the convention both, and the owner
// is checked before the brain is searched.
func (s *Server) write(ctx context.Context, user string, args addArgs) (*mcp.CallToolResult, error) {
	if strings.TrimSpace(args.Content) == "" {
		err := errors.New("th_add: content is required")
		return errResult(err), err
	}
	b, err := s.brainFor(user, args.Brain, rights.Write, "th_add")
	if err != nil {
		return errResult(err), err
	}
	access, _ := s.rights().Access(user, args.Brain)
	refused := func(err error, denied bool) (*mcp.CallToolResult, error) {
		// The cause goes back as it is, for REST to tell its kind.
		s.note(Entry{User: user, Brain: args.Brain, Tool: "th_add", IDs: map[string]any{"document_id": args.DocumentID}, Error: err.Error(), Denied: denied})
		return errResult(fmt.Errorf("th_add: %w", err)), err
	}
	exists := false
	var stored map[string]string
	if args.DocumentID != "" && (b.Convention != nil || access != rights.Curate) {
		author, md, found, err := s.authorOf(ctx, b, user, args.Brain, args.DocumentID)
		stored = md
		if err != nil {
			return refused(err, access != rights.Curate)
		}
		exists = found
		if found && access != rights.Curate && author != user {
			if author == "" {
				author = "no author"
			}
			return refused(&writeRefusal{fmt.Sprintf("document %q belongs to %s: propose it under another document_id", args.DocumentID, author), http.StatusConflict}, true)
		}
		if found && access != rights.Curate && b.Convention != nil && strings.HasPrefix(args.DocumentID, convention.ThemePrefix) {
			// A writer's rewrite would wait for review, hidden, and take
			// the theme from the brain until a curator accepts it again.
			return refused(&writeRefusal{fmt.Sprintf("document %q declares a theme: only a curator of %s rewrites it", args.DocumentID, args.Brain), http.StatusForbidden}, true)
		}
	}
	note := ""
	if c := b.Convention; c != nil {
		// A rewrite keeps the stored metadata the caller does not send: a
		// session that rewrites a state without metadata keeps its theme
		// and nature. A key the caller sends wins.
		args.Metadata = keepStored(stored, args.Metadata)
		md, err := s.followsConvention(ctx, b, user, *c, args, exists)
		if err != nil {
			return refused(err, false)
		}
		if md["theme_proposed"] != "" && args.Metadata["theme"] != md["theme"] {
			note = fmt.Sprintf("theme %q is not declared: filed under %s, theme_proposed %s", args.Metadata["theme"], md["theme"], md["theme_proposed"])
		}
		args.Metadata = md
	}
	if args.DocumentID == "" {
		args.DocumentID = user + "-" + time.Now().UTC().Format("20060102T150405.000000000Z")
	}
	md := map[string]string{}
	for k, v := range args.Metadata {
		md[k] = v
	}
	req := map[string]any{"document_id": args.DocumentID, "content": args.Content, "author": user, "metadata": md}
	if access == rights.Curate {
		// The engine names the calling key as reviewer and refuses another
		// name, so the gateway sends none.
		req["review_status"] = "accepted"
	} else {
		req["review_status"] = "proposed"
	}
	if args.Title != "" {
		req["title"] = args.Title
	}
	if args.SourceURL != "" {
		req["source_url"] = args.SourceURL
	}
	res := s.audited(ctx, b, user, args.Brain, "th_add", "ingest_document", req, map[string]any{"document_id": args.DocumentID})
	if note != "" && !res.IsError {
		res.Content = append(res.Content, &mcp.TextContent{Text: note})
	}
	return res, nil
}

// writeRefusal is a write refused for the caller's right over the
// document, with the REST status that says so.
type writeRefusal struct {
	msg    string
	status int
}

func (e *writeRefusal) Error() string { return e.msg }

// nearCopy is the refusal of a document almost identical to one the brain
// has.
type nearCopy struct {
	id       string
	likeness float64
}

func (e *nearCopy) Error() string {
	return fmt.Sprintf("near copy of %s (likeness %.2f): rewrite %s under its id, or name it in metadata replaces or distinct_from", e.id, e.likeness, e.id)
}

// followsConvention applies the brain's convention to a document before it
// is written: its metadata and id (convention.Check), then the near copies.
// The brain is searched by words for the document's text, each candidate is
// read whole, and one whose likeness reaches the brain's bound refuses the
// write, unless it is the document itself or the caller names it in
// replaces or distinct_from. A search or a read that fails, or does not
// decode, refuses the write too: the gate does not open because the brain
// could not be read. exists says the brain has a document of this id.
func (s *Server) followsConvention(ctx context.Context, b rights.Brain, user string, c rights.Convention, args addArgs, exists bool) (map[string]string, error) {
	var declared []string
	// The brain's themes are read for a new theme document, and for a
	// document whose theme, or proposed theme, the convention does not list.
	theme, unsorted := args.Metadata["theme"], c.UnsortedTheme()
	if p := args.Metadata["theme_proposed"]; theme == unsorted && p != "" {
		theme = p
	}
	themeDoc := strings.HasPrefix(args.DocumentID, convention.ThemePrefix)
	if (themeDoc && !exists) || (!themeDoc && theme != unsorted && !slices.Contains(c.Themes, theme)) {
		var err error
		if declared, err = s.declaredThemes(ctx, b, user, args.Brain); err != nil {
			return nil, fmt.Errorf("cannot read the brain's themes: %w", err)
		}
	}
	md, err := convention.Check(c, args.DocumentID, args.Metadata, exists, declared)
	if err != nil {
		return nil, err
	}
	query := strings.Join(strings.Fields(args.Content)[:min(len(strings.Fields(args.Content)), nearQueryWords)], " ")
	result, err := s.callBrain(ctx, b, user, args.Brain, "knowledge_search", map[string]any{"query": query, "top_k": nearCandidates, "retrieval_mode": "lexical"})
	if err != nil {
		return nil, fmt.Errorf("cannot look for near copies: %w", err)
	}
	var found struct {
		Results []hit `json:"results"`
	}
	if err := json.Unmarshal([]byte(result), &found); err != nil {
		return nil, fmt.Errorf("cannot look for near copies: decode knowledge_search: %w", err)
	}
	for _, h := range found.Results {
		id := h.KnowledgeID
		if id == args.DocumentID || id == md["replaces"] || id == md["distinct_from"] {
			continue
		}
		got, err := s.callBrain(ctx, b, user, args.Brain, "knowledge_get", map[string]any{"knowledge_id": id, "collection": nil})
		if err != nil {
			return nil, fmt.Errorf("cannot look for near copies: %w", err)
		}
		var doc struct {
			Knowledge struct {
				Content string `json:"content"`
			} `json:"knowledge"`
		}
		if err := json.Unmarshal([]byte(got), &doc); err != nil {
			return nil, fmt.Errorf("cannot look for near copies: decode knowledge_get %s: %w", id, err)
		}
		if l := convention.Likeness(args.Content, doc.Knowledge.Content); l >= c.NearBound() {
			return nil, &nearCopy{id, l}
		}
	}
	return md, nil
}

// The near-copy search asks the brain, by its words, for the first
// nearQueryWords of the document, and reads the nearCandidates first hits.
const (
	nearQueryWords = 40
	nearCandidates = 5
)

// keepStored returns the stored metadata of a document, without the
// engine's own keys, under the keys the caller sends.
func keepStored(stored, sent map[string]string) map[string]string {
	out := make(map[string]string, len(stored)+len(sent))
	for k, v := range stored {
		if !strings.HasPrefix(k, "_") && k != "written_by" {
			out[k] = v
		}
	}
	for k, v := range sent {
		out[k] = v
	}
	return out
}

// authorOf reads a document's author and stored metadata, found false when the brain has no
// document of that id. th_add replaces the document of that id: a writer
// may not take over another author's document, since the proposal would
// take the accepted text out of the brain until review, under the
// writer's name.
func (s *Server) authorOf(ctx context.Context, b rights.Brain, user, name, id string) (author string, md map[string]string, found bool, err error) {
	result, err := s.callBrain(ctx, b, user, name, "knowledge_get", map[string]any{"knowledge_id": id, "collection": nil})
	if status.Code(err) == codes.NotFound {
		return "", nil, false, nil
	}
	if err != nil {
		return "", nil, false, err
	}
	var got struct {
		Knowledge struct {
			Author   string            `json:"author"`
			Metadata map[string]string `json:"metadata"`
		} `json:"knowledge"`
	}
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		return "", nil, false, fmt.Errorf("decode knowledge_get: %w", err)
	}
	return got.Knowledge.Author, got.Knowledge.Metadata, true, nil
}

// pendingHit is one document that waits for review.
type pendingHit struct {
	KnowledgeID string `json:"knowledge_id"`
	Title       string `json:"title,omitempty"`
	Author      string `json:"author,omitempty"`
	Snippet     string `json:"snippet"`
}

// pending lists the proposed documents of the brain: aggregate_metadata
// groups the chunks by document under a filter on the engine's key
// _review_status, which lifts the review gate for that status (a search
// needs a text, a listing does not), then knowledge_get reads each
// document. One audit line names them all.
// get reads one document, whole or as the brain's excerpt. knowledge_get
// takes no collection: the brain's row guard checks the document against
// the key.
func (s *Server) get(ctx context.Context, user string, args getArgs) *mcp.CallToolResult {
	if strings.TrimSpace(args.KnowledgeID) == "" {
		return errResult(errors.New("th_get: knowledge_id is required"))
	}
	b, err := s.brainFor(user, args.Brain, rights.Read, "th_get")
	if err != nil {
		return errResult(err)
	}
	ids := map[string]any{"knowledge_id": args.KnowledgeID}
	result, err := s.callBrain(ctx, b, user, args.Brain, "knowledge_get", map[string]any{"knowledge_id": args.KnowledgeID, "collection": nil})
	entry := Entry{User: user, Brain: args.Brain, Tool: "th_get", IDs: ids}
	if err != nil {
		entry.Error, entry.Denied = err.Error(), denied(err)
		s.note(entry)
		return errResult(fmt.Errorf("th_get: %w", err))
	}
	var got struct {
		Knowledge struct {
			Author   string `json:"author"`
			Metadata struct {
				ReviewStatus string `json:"_review_status"`
			} `json:"metadata"`
		} `json:"knowledge"`
	}
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		entry.Error = err.Error()
		s.note(entry)
		return errResult(fmt.Errorf("th_get: decode knowledge_get: %w", err))
	}
	// The review gate keeps a proposal, or a rejected document, out of the
	// searches; read whole by its id, it is served to its author and the
	// curators only. The others are told it is not found.
	access, _ := s.rights().Access(user, args.Brain)
	if st := got.Knowledge.Metadata.ReviewStatus; st != "" && st != "accepted" && access != rights.Curate && got.Knowledge.Author != user {
		entry.Error, entry.Denied = "document "+st+", not accepted", true
		s.note(entry)
		return errResult(fmt.Errorf("th_get: document %q not found in %s", args.KnowledgeID, args.Brain))
	}
	if b.ExcerptWords > 0 {
		result = excerptOf(b, result)
	}
	entry.OK = true
	s.note(entry)
	return attributed(&mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, s.noticesFor([]string{args.Brain}, ""))
}

// excerptOf rewrites a knowledge_get result with the brain's excerpt of its
// content; the other fields pass as the engine wrote them. get decoded the
// result once already, so it decodes here.
func excerptOf(b rights.Brain, result string) string {
	var doc struct {
		Knowledge map[string]any `json:"knowledge"`
	}
	_ = json.Unmarshal([]byte(result), &doc)
	k := doc.Knowledge
	str := func(key string) string { v, _ := k[key].(string); return v }
	content, note := bound(b, str("collection"), str("content"), str("source_url"))
	if note == nil {
		return result
	}
	k["content"], k["excerpt"] = content, note
	out, _ := json.Marshal(doc)
	return string(out)
}

func (s *Server) pending(ctx context.Context, user string, args pendingArgs) *mcp.CallToolResult {
	b, err := s.brainFor(user, args.Brain, rights.Curate, "th_pending")
	if err != nil {
		return errResult(err)
	}
	if args.TopK <= 0 {
		args.TopK = 20
	}
	hits, err := s.listPending(ctx, b, user, args.Brain, args.TopK)
	entry := Entry{User: user, Brain: args.Brain, Tool: "th_pending"}
	if err != nil {
		entry.Error = err.Error()
		entry.Denied = denied(err)
		s.note(entry)
		return errResult(fmt.Errorf("th_pending: %w", err))
	}
	var ids []string
	for _, h := range hits {
		ids = append(ids, h.KnowledgeID)
	}
	entry.OK, entry.IDs = true, map[string]any{"knowledge_ids": ids}
	s.note(entry)
	return jsonResult(map[string]any{"brain": args.Brain, "results": hits})
}

func (s *Server) listPending(ctx context.Context, b rights.Brain, user, name string, topK int) ([]pendingHit, error) {
	req := map[string]any{"type": "group_by", "group_by": []string{"document_id"}, "filters": map[string]any{"_review_status": "proposed"}, "limit": topK}
	result, err := s.callBrain(ctx, b, user, name, "aggregate_metadata", req)
	if err != nil {
		return nil, err
	}
	var groups struct {
		Results []struct {
			Keys map[string]string `json:"keys"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(result), &groups); err != nil {
		return nil, fmt.Errorf("decode aggregate_metadata: %w", err)
	}
	hits := []pendingHit{}
	for _, g := range groups.Results {
		id := g.Keys["document_id"]
		// knowledge_get takes no collection: the row guard of the brain
		// checks the document against the key.
		result, err := s.callBrain(ctx, b, user, name, "knowledge_get", map[string]any{"knowledge_id": id, "collection": nil})
		if err != nil {
			return nil, err
		}
		var got struct {
			Knowledge struct {
				Title   string `json:"title"`
				Author  string `json:"author"`
				Content string `json:"content"`
			} `json:"knowledge"`
		}
		if err := json.Unmarshal([]byte(result), &got); err != nil {
			return nil, fmt.Errorf("decode knowledge_get: %w", err)
		}
		hits = append(hits, pendingHit{KnowledgeID: id, Title: got.Knowledge.Title, Author: got.Knowledge.Author, Snippet: snippet(got.Knowledge.Content)})
	}
	return hits, nil
}

// snippet is the start of a content, whole words, about 200 characters.
func snippet(content string) string {
	content = strings.Join(strings.Fields(content), " ")
	if len(content) <= 200 {
		return content
	}
	if i := strings.LastIndex(content[:200], " "); i > 0 {
		return content[:i] + "…"
	}
	return content[:200] + "…"
}

func (s *Server) review(ctx context.Context, user string, args reviewArgs) *mcp.CallToolResult {
	if args.Decision != "accepted" && args.Decision != "rejected" {
		return errResult(errors.New("th_review: decision must be accepted or rejected"))
	}
	if strings.TrimSpace(args.KnowledgeID) == "" {
		return errResult(errors.New("th_review: knowledge_id is required"))
	}
	b, err := s.brainFor(user, args.Brain, rights.Curate, "th_review")
	if err != nil {
		return errResult(err)
	}
	// The decision goes through the explicit fields, with no metadata, so the
	// engine keeps the stored metadata and takes the calling key as reviewer;
	// a reviewed_by naming anyone else would be refused.
	req := map[string]any{"knowledge_id": args.KnowledgeID, "review_status": args.Decision}
	return s.audited(ctx, b, user, args.Brain, "th_review", "knowledge_update", req, map[string]any{"knowledge_id": args.KnowledgeID, "decision": args.Decision})
}
