package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/rights"
)

// The pyramid of a brain: syntheses that a producer (contaixt, for one)
// writes with a model and stores in a collection of their own, each listing
// the documents it cites. th_recall serves them first; th_descend opens what
// one of them cites. The gateway neither writes nor checks a synthesis.

// defaultSyntheses is how many syntheses th_recall serves when the caller
// says nothing; defaultDescent is how many sources th_descend reads.
const (
	defaultSyntheses = 4
	defaultDescent   = 5
)

type descendArgs struct {
	Brain       string `json:"brain"`
	KnowledgeID string `json:"knowledge_id"`
	Limit       int    `json:"limit,omitempty"`
}

// record is one document as knowledge_get returns it.
type record struct {
	KnowledgeID string            `json:"knowledge_id"`
	Title       string            `json:"title,omitempty"`
	SourceURL   string            `json:"source_url,omitempty"`
	Author      string            `json:"author,omitempty"`
	Content     string            `json:"content,omitempty"`
	Excerpt     *excerpt          `json:"excerpt,omitempty"`
	Metadata    map[string]string `json:"-"`
}

type missingSource struct {
	KnowledgeID string `json:"knowledge_id"`
	Reason      string `json:"reason"`
}

type descent struct {
	Synthesis struct {
		KnowledgeID string `json:"knowledge_id"`
		Title       string `json:"title,omitempty"`
	} `json:"synthesis"`
	Sources []record        `json:"sources"`
	Missing []missingSource `json:"missing,omitempty"`
	// More counts the cited sources left out by the limit.
	More int `json:"more,omitempty"`
}

// sourcesOf reads the ids a synthesis cites: a JSON list of strings, or one
// id. A value that opens a list it does not close cites nothing.
func sourcesOf(value string) []string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") {
		if value == "" {
			return nil
		}
		return []string{value}
	}
	var ids []string
	_ = json.Unmarshal([]byte(value), &ids)
	return ids
}

// readDoc reads one document whole. knowledge_get takes no collection: the
// row guard of the brain checks the document against the key.
func (s *Server) readDoc(ctx context.Context, b rights.Brain, user, name, id string) (record, error) {
	result, err := s.callBrain(ctx, b, user, name, "knowledge_get", map[string]any{"knowledge_id": id, "collection": nil})
	if err != nil {
		return record{}, err
	}
	var got struct {
		Knowledge struct {
			Title     string            `json:"title"`
			SourceURL string            `json:"source_url"`
			Author    string            `json:"author"`
			Content   string            `json:"content"`
			Metadata  map[string]string `json:"metadata"`
		} `json:"knowledge"`
	}
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		return record{}, fmt.Errorf("decode knowledge_get: %w", err)
	}
	k := got.Knowledge
	return record{KnowledgeID: id, Title: k.Title, SourceURL: k.SourceURL, Author: k.Author, Content: k.Content, Metadata: k.Metadata}, nil
}

// descend reads a synthesis of the brain's pyramid and the sources it cites,
// up to the limit. A source that cannot be read is listed in missing.
func (s *Server) descend(ctx context.Context, user string, args descendArgs) *mcp.CallToolResult {
	if strings.TrimSpace(args.KnowledgeID) == "" {
		return errResult(errors.New("th_descend: knowledge_id is required"))
	}
	b, err := s.brainFor(user, args.Brain, rights.Read, "th_descend")
	if err != nil {
		return errResult(err)
	}
	if b.Pyramid == nil {
		return errResult(fmt.Errorf("th_descend: %s has no pyramid", args.Brain))
	}
	if args.Limit <= 0 {
		args.Limit = defaultDescent
	}
	entry := Entry{User: user, Brain: args.Brain, Tool: "th_descend", IDs: map[string]any{"knowledge_id": args.KnowledgeID}}
	fail := func(err error) *mcp.CallToolResult {
		entry.Error = err.Error()
		entry.Denied = denied(err)
		s.note(entry)
		return errResult(fmt.Errorf("th_descend: %w", err))
	}
	syn, err := s.readDoc(ctx, b, user, args.Brain, args.KnowledgeID)
	if err != nil {
		return fail(err)
	}
	ids := sourcesOf(syn.Metadata[b.Pyramid.SourcesKey()])
	if len(ids) == 0 {
		return fail(fmt.Errorf("%s cites no source under %s", args.KnowledgeID, b.Pyramid.SourcesKey()))
	}
	var out descent
	out.Synthesis.KnowledgeID, out.Synthesis.Title = args.KnowledgeID, syn.Title
	if len(ids) > args.Limit {
		out.More = len(ids) - args.Limit
		ids = ids[:args.Limit]
	}
	read := []string{}
	out.Sources = []record{}
	for _, id := range ids {
		doc, err := s.readDoc(ctx, b, user, args.Brain, id)
		if err != nil {
			out.Missing = append(out.Missing, missingSource{KnowledgeID: id, Reason: err.Error()})
			continue
		}
		doc.Content, doc.Excerpt = bound(b, "", doc.Content, doc.SourceURL)
		out.Sources = append(out.Sources, doc)
		read = append(read, id)
	}
	entry.IDs["sources"] = read
	entry.OK = true
	s.note(entry)
	return jsonResult(out)
}

func (s *Server) addPyramidTools(srv *mcp.Server, user string) {
	srv.AddTool(&mcp.Tool{Name: "th_descend", Description: "Opens the sources a synthesis cites: give the brain and the knowledge_id of a th_recall result marked level synthesis. Returns the synthesis, each cited document (up to limit, 5 by default), and the sources that could not be read. In a brain served in excerpts, a document is cut to its first words, with the address of the original.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"brain":        map[string]any{"type": "string"},
			"knowledge_id": map[string]any{"type": "string"},
			"limit":        map[string]any{"type": "integer", "description": "sources to read, default 5"},
		}, "required": []string{"brain", "knowledge_id"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args descendArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_descend: %w", err)), nil
			}
			return s.descend(ctx, user, args), nil
		})
}
