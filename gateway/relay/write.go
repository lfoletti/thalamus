package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	srv.AddTool(&mcp.Tool{Name: "th_add", Description: "Proposes a document into a brain the caller may write. It waits for a curator's review; a curator's own document is accepted at once. Returns the engine's ingest report.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"brain":       map[string]any{"type": "string"},
			"document_id": map[string]any{"type": "string", "description": "stable id; derived from the time when absent"},
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
	if strings.TrimSpace(args.Content) == "" {
		return errResult(errors.New("th_add: content is required"))
	}
	b, err := s.brainFor(user, args.Brain, rights.Write, "th_add")
	if err != nil {
		return errResult(err)
	}
	access, _ := s.rights().Access(user, args.Brain)
	if args.DocumentID == "" {
		args.DocumentID = user + "-" + time.Now().UTC().Format("20060102T150405.000000000Z")
	} else if access != rights.Curate {
		if err := s.owns(ctx, b, user, args.Brain, args.DocumentID); err != nil {
			s.note(Entry{User: user, Brain: args.Brain, Tool: "th_add", IDs: map[string]any{"document_id": args.DocumentID}, Error: err.Error(), Denied: true})
			return errResult(fmt.Errorf("th_add: %w", err))
		}
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
	return s.audited(ctx, b, user, args.Brain, "th_add", "ingest_document", req, map[string]any{"document_id": args.DocumentID})
}

// owns checks that a writer who names a document_id does not take over
// another author's document: th_add replaces the document of that id, and the
// proposal would take the accepted text out of the brain until review, under
// the writer's name. An id the brain does not have is free.
func (s *Server) owns(ctx context.Context, b rights.Brain, user, name, id string) error {
	result, err := s.callBrain(ctx, b, user, name, "knowledge_get", map[string]any{"knowledge_id": id, "collection": nil})
	if status.Code(err) == codes.NotFound {
		return nil
	}
	if err != nil {
		return err
	}
	var got struct {
		Knowledge struct {
			Author string `json:"author"`
		} `json:"knowledge"`
	}
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		return fmt.Errorf("decode knowledge_get: %w", err)
	}
	if got.Knowledge.Author != user {
		owner := got.Knowledge.Author
		if owner == "" {
			owner = "no author"
		}
		return fmt.Errorf("document %q belongs to %s: propose it under another document_id", id, owner)
	}
	return nil
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
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}
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
