package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/convention"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// declaredThemes are the themes the brain's accepted theme documents
// declare, read from their metadata declares_theme: only a document
// theme:<subject> declares its subject, whatever another row carries. A
// proposed theme document is not served under the review gate, so it
// declares nothing until a curator accepts it.
func (s *Server) declaredThemes(ctx context.Context, b rights.Brain, user, name string) ([]string, error) {
	groups, err := s.groups(ctx, b, user, name, []string{"declares_theme", "document_id"}, nil)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, g := range groups {
		if v := g.key("declares_theme"); g.key("document_id") == convention.ThemePrefix+v {
			out = append(out, v)
		}
	}
	return out, nil
}

// group is one row of aggregate_metadata.
type group struct {
	Keys  map[string]any `json:"keys"`
	Count int            `json:"count"`
}

// key is a grouped value, "" when the records do not carry it. The engine
// groups a value that reads as a number as a number: it is read back as
// text.
func (g group) key(k string) string {
	switch v := g.Keys[k].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return ""
}

// groupLimit bounds the groups one aggregate_metadata call returns.
const groupLimit = 100000

func (s *Server) groups(ctx context.Context, b rights.Brain, user, name string, by []string, filters map[string]string) ([]group, error) {
	req := map[string]any{"type": "group_by", "group_by": by, "limit": groupLimit}
	if filters != nil {
		req["filters"] = filters
	}
	result, err := s.callBrain(ctx, b, user, name, "aggregate_metadata", req)
	if err != nil {
		return nil, err
	}
	var out struct {
		Results []group `json:"results"`
	}
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("decode aggregate_metadata: %w", err)
	}
	return out.Results, nil
}

type themesArgs struct {
	Brain string `json:"brain"`
}

type rethemeArgs struct {
	Brain    string `json:"brain"`
	Proposed string `json:"proposed"`
	Theme    string `json:"theme"`
}

type themeCount struct {
	Theme     string `json:"theme"`
	Documents int    `json:"documents"`
	// Declared says where the theme is declared: convention or document.
	Declared string `json:"declared"`
}

type proposal struct {
	Theme     string `json:"theme"`
	Documents int    `json:"documents"`
	Candidate bool   `json:"candidate"`
	CloseTo   string `json:"close_to,omitempty"`
}

type unsortedCount struct {
	Theme           string `json:"theme"`
	Documents       int    `json:"documents"`
	WithoutProposal int    `json:"without_proposal"`
}

// looseCount is a theme the documents carry that the brain does not know:
// dropped from the convention, or whose theme document is gone.
type looseCount struct {
	Theme     string `json:"theme"`
	Documents int    `json:"documents"`
}

type themesReport struct {
	Brain string `json:"brain"`
	// Documents counts every document but the theme documents.
	Documents    int           `json:"documents"`
	Themes       []themeCount  `json:"themes"`
	Unsorted     unsortedCount `json:"unsorted"`
	Undeclared   []looseCount  `json:"undeclared"`
	WithoutTheme int           `json:"without_theme"`
	Proposed     []proposal    `json:"proposed"`
	Broad        []string      `json:"broad"`
	Dormant      []string      `json:"dormant"`
}

func (s *Server) addThemeTools(srv *mcp.Server, user string) {
	srv.AddTool(&mcp.Tool{Name: "th_themes", Description: "Tells how a brain with a convention files its documents: the documents of each theme, the unsorted ones and the themes they propose, the candidates (a theme proposed by enough documents), a proposal close to an existing theme, the themes that carry more than a third of the filed documents, the themes nothing carries, and apart the documents under a theme the brain no longer knows or without a theme. A curator declares a candidate with th_add of a document theme:<subject>, then files its documents with th_retheme.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"brain": map[string]any{"type": "string"},
		}, "required": []string{"brain"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args themesArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_themes: %w", err)), nil
			}
			return s.themes(ctx, user, args), nil
		})
	srv.AddTool(&mcp.Tool{Name: "th_retheme", Description: "Files under a theme the brain knows the unsorted documents that proposed a theme, as a curator: each is written again with its own author, title and source, under the theme, without theme_proposed. To accept a proposal, declare its theme first (th_add theme:<subject>) and file it there; to refuse it, file it under another theme.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"brain":    map[string]any{"type": "string"},
			"proposed": map[string]any{"type": "string", "description": "the theme the documents propose (metadata theme_proposed)"},
			"theme":    map[string]any{"type": "string", "description": "the theme they are filed under"},
		}, "required": []string{"brain", "proposed", "theme"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args rethemeArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_retheme: %w", err)), nil
			}
			return s.retheme(ctx, user, args), nil
		})
}

// conventionOf checks the access and that the brain has a convention.
func (s *Server) conventionOf(user, name string, need rights.Access, tool string) (rights.Brain, rights.Convention, error) {
	b, err := s.brainFor(user, name, need, tool)
	if err != nil {
		return b, rights.Convention{}, err
	}
	if b.Convention == nil {
		return b, rights.Convention{}, fmt.Errorf("%s: %s has no convention", tool, name)
	}
	return b, *b.Convention, nil
}

// documentsBy counts the documents, not the records, of each value of key.
func (s *Server) documentsBy(ctx context.Context, b rights.Brain, user, name, key string, filters map[string]string) (map[string]int, error) {
	groups, err := s.groups(ctx, b, user, name, []string{key, "document_id"}, filters)
	if err != nil {
		return nil, err
	}
	// The engine groups by the pair: one group is one document. A theme
	// document files nothing, and is not counted.
	out := map[string]int{}
	for _, g := range groups {
		if !strings.HasPrefix(g.key("document_id"), convention.ThemePrefix) {
			out[g.key(key)]++
		}
	}
	return out, nil
}

func (s *Server) themes(ctx context.Context, user string, args themesArgs) *mcp.CallToolResult {
	b, c, err := s.conventionOf(user, args.Brain, rights.Read, "th_themes")
	if err != nil {
		return errResult(err)
	}
	report, err := s.themesReport(ctx, b, user, args.Brain, c)
	entry := Entry{User: user, Brain: args.Brain, Tool: "th_themes", OK: err == nil}
	if err != nil {
		entry.Error, entry.Denied = err.Error(), denied(err)
		s.note(entry)
		return errResult(fmt.Errorf("th_themes: %w", err))
	}
	s.note(entry)
	return jsonResult(report)
}

// broadShare is the share of the filed documents above which a theme is
// broad, a hint that it holds two themes.
const broadShare = 1.0 / 3

func (s *Server) themesReport(ctx context.Context, b rights.Brain, user, name string, c rights.Convention) (themesReport, error) {
	r := themesReport{Brain: name, Themes: []themeCount{}, Undeclared: []looseCount{}, Proposed: []proposal{}, Broad: []string{}, Dormant: []string{}}
	declared, err := s.declaredThemes(ctx, b, user, name)
	if err != nil {
		return r, err
	}
	byTheme, err := s.documentsBy(ctx, b, user, name, "theme", nil)
	if err != nil {
		return r, err
	}
	unsorted := c.UnsortedTheme()
	byProposal, err := s.documentsBy(ctx, b, user, name, "theme_proposed", map[string]string{"theme": unsorted})
	if err != nil {
		return r, err
	}
	known := append(slices.Clone(c.Themes), declared...)
	filed := 0
	for i, th := range known {
		where := "convention"
		if i >= len(c.Themes) {
			where = "document"
		}
		r.Themes = append(r.Themes, themeCount{Theme: th, Documents: byTheme[th], Declared: where})
		filed += byTheme[th]
		if byTheme[th] == 0 {
			r.Dormant = append(r.Dormant, th)
		}
	}
	for _, t := range r.Themes {
		if float64(t.Documents) > broadShare*float64(filed) {
			r.Broad = append(r.Broad, t.Theme)
		}
	}
	r.Unsorted = unsortedCount{Theme: unsorted, Documents: byTheme[unsorted], WithoutProposal: byProposal[""]}
	for th, n := range byTheme {
		r.Documents += n
		switch {
		case th == "":
			r.WithoutTheme = n
		case th != unsorted && !slices.Contains(known, th):
			r.Undeclared = append(r.Undeclared, looseCount{th, n})
		}
	}
	sort.Slice(r.Undeclared, func(i, j int) bool {
		a, b := r.Undeclared[i], r.Undeclared[j]
		if a.Documents != b.Documents {
			return a.Documents > b.Documents
		}
		return a.Theme < b.Theme
	})
	for p, n := range byProposal {
		if p == "" {
			continue
		}
		pr := proposal{Theme: p, Documents: n, Candidate: n >= c.EmergeAt()}
		for _, k := range known {
			if convention.CloseNames(k, p) {
				pr.CloseTo = k
				break
			}
		}
		r.Proposed = append(r.Proposed, pr)
	}
	sort.Slice(r.Proposed, func(i, j int) bool {
		a, b := r.Proposed[i], r.Proposed[j]
		if a.Documents != b.Documents {
			return a.Documents > b.Documents
		}
		return a.Theme < b.Theme
	})
	return r, nil
}

func (s *Server) retheme(ctx context.Context, user string, args rethemeArgs) *mcp.CallToolResult {
	b, c, err := s.conventionOf(user, args.Brain, rights.Curate, "th_retheme")
	if err != nil {
		return errResult(err)
	}
	if args.Proposed == "" {
		return errResult(errors.New("th_retheme: proposed is required"))
	}
	moved, err := s.refile(ctx, b, user, args, c)
	entry := Entry{User: user, Brain: args.Brain, Tool: "th_retheme", IDs: map[string]any{"moved": moved}, OK: err == nil}
	if err != nil {
		entry.Error, entry.Denied = err.Error(), denied(err)
		s.note(entry)
		done := "nothing"
		if len(moved) > 0 {
			done = strings.Join(moved, ", ")
		}
		return errResult(fmt.Errorf("th_retheme: after moving %s: %w", done, err))
	}
	s.note(entry)
	return jsonResult(map[string]any{"brain": args.Brain, "theme": args.Theme, "moved": moved})
}

// refile writes again, under the theme, each document that proposed it.
func (s *Server) refile(ctx context.Context, b rights.Brain, user string, args rethemeArgs, c rights.Convention) ([]string, error) {
	moved := []string{}
	declared, err := s.declaredThemes(ctx, b, user, args.Brain)
	if err != nil {
		return moved, err
	}
	if !slices.Contains(c.Themes, args.Theme) && !slices.Contains(declared, args.Theme) {
		return moved, fmt.Errorf("theme %q is not declared: declare it with th_add theme:%s", args.Theme, args.Theme)
	}
	groups, err := s.groups(ctx, b, user, args.Brain, []string{"document_id"}, map[string]string{"theme": c.UnsortedTheme(), "theme_proposed": args.Proposed})
	if err != nil {
		return moved, err
	}
	for _, g := range groups {
		id := g.key("document_id")
		result, err := s.callBrain(ctx, b, user, args.Brain, "knowledge_get", map[string]any{"knowledge_id": id, "collection": nil})
		if err != nil {
			return moved, err
		}
		var got struct {
			Knowledge struct {
				Title     string            `json:"title"`
				Author    string            `json:"author"`
				SourceURL string            `json:"source_url"`
				Content   string            `json:"content"`
				Metadata  map[string]string `json:"metadata"`
			} `json:"knowledge"`
		}
		if err := json.Unmarshal([]byte(result), &got); err != nil {
			return moved, fmt.Errorf("decode knowledge_get: %w", err)
		}
		k := got.Knowledge
		md := map[string]string{}
		for key, v := range k.Metadata {
			md[key] = v
		}
		delete(md, "theme_proposed")
		md["theme"] = args.Theme
		req := map[string]any{"document_id": id, "content": k.Content, "author": k.Author, "metadata": md, "review_status": "accepted"}
		if k.Title != "" {
			req["title"] = k.Title
		}
		if k.SourceURL != "" {
			req["source_url"] = k.SourceURL
		}
		if _, err := s.callBrain(ctx, b, user, args.Brain, "ingest_document", req); err != nil {
			return moved, err
		}
		moved = append(moved, id)
	}
	return moved, nil
}
