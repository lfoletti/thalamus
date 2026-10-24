package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// th_entity: one entity across the brains. No brain stores an edge to
// another brain; the same entity has the same node id in every brain (its
// normalized name), so the gateway asks each brain for the entity and its
// neighbors with the caller's own key for it, and fuses the answers by node
// id. Each key's scope decides what its brain serves: the engine walks only
// the nodes and edges of the key's collections.

// defaultEntityLimit is how many neighbors th_entity asks of each brain;
// maxEntityDepth and maxEntityLimit bound the walk, because each hop and
// each node multiply what every brain walks.
const (
	defaultEntityLimit = 20
	maxEntityDepth     = 3
	maxEntityLimit     = 100
)

type entityArgs struct {
	Entity   string   `json:"entity"`
	Brains   []string `json:"brains,omitempty"`
	Skip     []string `json:"skip_brains,omitempty"`
	MaxDepth int      `json:"max_depth,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// node is a graph node as the engine's graph tools return it. The vector
// and the properties are not served on.
type node struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	NodeType string `json:"node_type,omitempty"`
}

// brainGraph is one brain's answer: whether it serves the entity, and the
// neighbors it serves.
type brainGraph struct {
	Brain     string
	Known     bool
	Neighbors []node
}

type neighbor struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	NodeType string   `json:"node_type,omitempty"`
	Brains   []string `json:"brains"`
}

type entityResult struct {
	Entity    string     `json:"entity"`
	NodeID    string     `json:"node_id"`
	KnownIn   []string   `json:"known_in"`
	Neighbors []neighbor `json:"neighbors"`
	Skipped   []skipped  `json:"skipped,omitempty"`
}

// fuseNeighbors merges the brains' answers by node id. A neighbor is an
// entity: the chunk and document nodes that mention one are left out, so a
// passage is never served as a name. A neighbor names
// every brain that serves it, and the neighbors that more brains serve come
// first, then by id. A type one brain serves is kept.
func fuseNeighbors(graphs []brainGraph) (known []string, out []neighbor) {
	known, out = []string{}, []neighbor{}
	at := map[string]int{}
	for _, g := range graphs {
		if g.Known {
			known = append(known, g.Brain)
		}
		for _, n := range g.Neighbors {
			if !strings.HasPrefix(n.ID, "entity:") {
				continue
			}
			i, seen := at[n.ID]
			if !seen {
				i = len(out)
				at[n.ID] = i
				out = append(out, neighbor{ID: n.ID, Name: n.Content})
			}
			if out[i].NodeType == "" {
				out[i].NodeType = n.NodeType
			}
			if !slices.Contains(out[i].Brains, g.Brain) {
				out[i].Brains = append(out[i].Brains, g.Brain)
			}
		}
	}
	sort.Strings(known)
	for i := range out {
		sort.Strings(out[i].Brains)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if len(out[i].Brains) != len(out[j].Brains) {
			return len(out[i].Brains) > len(out[j].Brains)
		}
		return out[i].ID < out[j].ID
	})
	return known, out
}

// askEntity asks one brain for the entity's neighbors, then whether it
// serves the entity itself. The graph tools take no collection.
func (s *Server) askEntity(ctx context.Context, user, name string, args entityArgs) (string, brainGraph, error) {
	b := s.rights().Brains[name]
	raw, err := s.callBrain(ctx, b, user, name, "knowledge_memory_neighbors", map[string]any{
		"entity_name": args.Entity, "max_depth": args.MaxDepth, "limit": args.Limit, "collection": nil})
	if err != nil {
		return "", brainGraph{}, err
	}
	var walk struct {
		NodeID    string `json:"resolved_node_id"`
		Neighbors []node `json:"neighbors"`
	}
	if err := json.Unmarshal([]byte(raw), &walk); err != nil {
		return "", brainGraph{}, fmt.Errorf("decode knowledge_memory_neighbors: %w", err)
	}
	raw, err = s.callBrain(ctx, b, user, name, "get_nodes", map[string]any{"node_ids": []string{walk.NodeID}, "collection": nil})
	if err != nil {
		return "", brainGraph{}, err
	}
	var got struct {
		Nodes []node `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		return "", brainGraph{}, fmt.Errorf("decode get_nodes: %w", err)
	}
	return walk.NodeID, brainGraph{Brain: name, Known: len(got.Nodes) > 0, Neighbors: walk.Neighbors}, nil
}

// entity fans the entity out to the caller's brains and fuses the answers.
func (s *Server) entity(ctx context.Context, user string, args entityArgs) (entityResult, error) {
	args.Entity = strings.TrimSpace(args.Entity)
	if args.Entity == "" {
		return entityResult{}, errors.New("th_entity: entity is required")
	}
	if args.MaxDepth > maxEntityDepth {
		return entityResult{}, fmt.Errorf("th_entity: max_depth is at most %d", maxEntityDepth)
	}
	if args.MaxDepth <= 0 {
		args.MaxDepth = 1
	}
	if args.Limit > maxEntityLimit {
		return entityResult{}, fmt.Errorf("th_entity: limit is at most %d", maxEntityLimit)
	}
	if args.Limit <= 0 {
		args.Limit = defaultEntityLimit
	}
	grants, err := s.grantsFor(user, args.Brains, args.Skip, "th_entity")
	if err != nil {
		return entityResult{}, err
	}
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		graphs   = make([]brainGraph, len(grants))
		out      = entityResult{Entity: args.Entity}
		brains   []string
		read     []string
		refusals = map[string]string{}
	)
	for i, g := range grants {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			id, graph, err := s.askEntity(ctx, user, name, args)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				out.Skipped = append(out.Skipped, skipped{Brain: name, Reason: err.Error()})
				if denied(err) {
					refusals[name] = err.Error()
				}
				return
			}
			out.NodeID, graphs[i] = id, graph
			brains = append(brains, name)
		}(i, g.Brain)
	}
	wg.Wait()
	sort.Slice(out.Skipped, func(i, j int) bool { return out.Skipped[i].Brain < out.Skipped[j].Brain })
	sort.Strings(brains)
	out.KnownIn, out.Neighbors = fuseNeighbors(graphs)
	for _, n := range out.Neighbors {
		for _, b := range n.Brains {
			read = append(read, b+":"+n.ID)
		}
	}
	sort.Strings(read)
	s.note(Entry{User: user, Tool: "th_entity", OK: true, IDs: map[string]any{"entity": args.Entity, "brains": brains, "knowledge_ids": read, "denied": refusals}})
	return out, nil
}

func (s *Server) addEntityTools(srv *mcp.Server, user string) {
	srv.AddTool(&mcp.Tool{Name: "th_entity", Description: "Finds one entity across every brain the caller may read (or only those named in brains, less those in skip_brains), with the caller's own key for each, and fuses the answers by node id. Returns the brains that know the entity, its neighbors with the brains that serve each one, and the brains that were skipped.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"entity":      map[string]any{"type": "string", "description": "the entity's name"},
			"brains":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"skip_brains": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"max_depth":   map[string]any{"type": "integer", "description": fmt.Sprintf("hops of the walk, default 1, at most %d", maxEntityDepth)},
			"limit":       map[string]any{"type": "integer", "description": fmt.Sprintf("nodes each brain walks, default %d, at most %d", defaultEntityLimit, maxEntityLimit)},
		}, "required": []string{"entity"}}},
		func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args entityArgs
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return errResult(fmt.Errorf("th_entity: %w", err)), nil
			}
			out, err := s.entity(ctx, user, args)
			if err != nil {
				return errResult(err), nil
			}
			return jsonResult(out), nil
		})
}
