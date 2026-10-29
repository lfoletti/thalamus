package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/liliang-cn/cortexdb/v2/pkg/cortexdb"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/rights"
)

// The graph tools: a closed list of the engine's read-only graph tools, each
// served on /mcp as th_<name> with the engine's own arguments and the brain.
// The call goes with the caller's own key for that brain, so the engine
// walks the key's part of the graph alone (CortexDB v2.141.0 opens the five
// to a scoped key). The gateway adds what the engine does not know: the
// rights file, the audit, the excerpt rule of the brain, and the review
// circuit, whose proposals are nodes of the graph.
var graphTools = []string{"find_nodes", "expand_graph", "get_nodes", "search_chunks_by_entities", "search_paths"}

// The graph tools read: the audit counts them, or names them on a brain
// whose reads are named.
func init() {
	for _, name := range graphTools {
		readsOf["th_"+name] = true
	}
}

// graphNote is said of every graph tool after the engine's description.
const graphNote = " Served for one brain the caller may read, with the caller's key for it. In a brain served in excerpts, no text is longer than the brain's excerpt_words, and the answer's excerpt says how many words were kept."

func (s *Server) addGraphTools(srv *mcp.Server, user string) {
	defs := map[string]cortexdb.ToolDefinition{}
	for _, d := range cortexdb.ToolDefinitions() {
		defs[d.Name] = d
	}
	for _, name := range graphTools {
		def := defs[name]
		props := map[string]any{"brain": map[string]any{"type": "string"}}
		declared := map[string]bool{}
		for p, schema := range def.InputSchema["properties"].(map[string]any) {
			props[p], declared[p] = schema, true
		}
		required := []string{"brain"}
		if r, ok := def.InputSchema["required"].([]string); ok {
			required = append(required, r...)
		}
		tool := "th_" + name
		srv.AddTool(&mcp.Tool{Name: tool, Description: def.Description + graphNote,
			InputSchema: map[string]any{"type": "object", "properties": props, "required": required}},
			func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				// The SDK hands the arguments on unchecked. Arguments that
				// are not an object name no brain, and are refused as such;
				// the engine checks the types of its own.
				var args map[string]any
				_ = json.Unmarshal(req.Params.Arguments, &args)
				return s.graph(ctx, user, tool, name, args, declared), nil
			})
	}
}

// graph relays one graph tool: the brain is read, the engine's arguments go
// on and nothing else, and the answer is cut to the brain's excerpts.
func (s *Server) graph(ctx context.Context, user, tool, engineTool string, args map[string]any, declared map[string]bool) *mcp.CallToolResult {
	name, _ := args["brain"].(string)
	b, err := s.brainFor(user, name, rights.Read, tool)
	if err != nil {
		return errResult(err)
	}
	if len(b.Curators) > 0 {
		return errResult(fmt.Errorf("%s: brain %s has a review circuit: its graph holds the documents that wait for review", tool, name))
	}
	req := map[string]any{"collection": nil}
	for k, v := range args {
		if declared[k] {
			req[k] = v
		}
	}
	raw, _ := json.Marshal(req)
	entry := Entry{User: user, Brain: name, Tool: tool, IDs: idsOf(string(raw))}
	result, err := s.callBrain(ctx, b, user, name, engineTool, req)
	if err == nil {
		var docs []string
		result, docs, err = excerpted(result, b.ExcerptWords)
		if len(docs) > 0 {
			if entry.IDs == nil {
				entry.IDs = map[string]any{}
			}
			entry.IDs["documents"] = docs
		}
	}
	if err != nil {
		entry.Error, entry.Denied = err.Error(), denied(err)
		s.note(entry)
		return errResult(fmt.Errorf("%s: %w", tool, err))
	}
	entry.OK = true
	s.note(entry)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}
}

// excerpted drops the vectors of an engine answer and, with n > 0, cuts
// every text in it to its first n words; the answer then says so. It also
// returns the documents the answer serves, for the audit: a chunk's
// document_id and a document node's id, sorted, once each.
func excerpted(raw string, n int) (string, []string, error) {
	var v any
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return "", nil, errors.New("the brain's answer is not JSON")
	}
	docs := map[string]bool{}
	documentsOf(v, docs)
	cutOff := false
	v = walkText(v, func(text string) string {
		kept, _, c := cut(text, n)
		cutOff = cutOff || c
		return kept
	})
	if m, ok := v.(map[string]any); ok && cutOff {
		m["excerpt"] = map[string]any{"words": n}
	}
	out, _ := json.Marshal(v)
	return string(out), slices.Sorted(maps.Keys(docs)), nil
}

// documentsOf adds to docs every document a decoded answer names.
func documentsOf(v any, docs map[string]bool) {
	switch x := v.(type) {
	case []any:
		for _, e := range x {
			documentsOf(e, docs)
		}
	case map[string]any:
		if d, ok := x["document_id"].(string); ok && d != "" {
			docs[d] = true
		}
		if id, ok := x["id"].(string); ok && strings.HasPrefix(id, "doc:") {
			docs[strings.TrimPrefix(id, "doc:")] = true
		}
		for _, e := range x {
			documentsOf(e, docs)
		}
	}
}

// walkText rewrites every string of a decoded JSON value and drops every
// "vector" field.
func walkText(v any, f func(string) string) any {
	switch x := v.(type) {
	case string:
		return f(x)
	case []any:
		for i := range x {
			x[i] = walkText(x[i], f)
		}
	case map[string]any:
		for k := range x {
			if k == "vector" {
				delete(x, k)
				continue
			}
			x[k] = walkText(x[k], f)
		}
	}
	return v
}
