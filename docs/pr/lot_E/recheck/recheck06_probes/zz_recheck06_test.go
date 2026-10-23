package cortexdb

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// Scratch probe for fiche 06 (not for upstream): two collections share the
// entity "Kestrel Vale"; only beta knows "Project Nightjar", a relation
// Kestrel Vale -funds-> Project Nightjar, and a description written from
// beta. Each graph tool is called under a CallerScope confined to alpha,
// and its JSON answer is searched for what only beta holds.
func TestRecheck06WhatAConfinedCallerSeesThroughTheGraphTools(t *testing.T) {
	db, err := Open(DefaultConfig(filepath.Join(t.TempDir(), "recheck06.db")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	bg := context.Background()
	if _, err := db.SaveKnowledge(bg, KnowledgeSaveRequest{
		KnowledgeID: "a1", Collection: "alpha", Content: "Kestrel Vale runs the alpha mill on the river.",
		Entities: []ToolEntityInput{{Name: "Kestrel Vale", Description: "alpha miller", ChunkIDs: []string{graphChunkNodeID("a1", 0)}}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SaveKnowledge(bg, KnowledgeSaveRequest{
		KnowledgeID: "b1", Collection: "beta", Content: "Kestrel Vale secretly funds Project Nightjar with BETASECRET money.",
		Entities: []ToolEntityInput{
			{Name: "Kestrel Vale", Description: "beta-only description BETADESC", ChunkIDs: []string{graphChunkNodeID("b1", 0)}},
			{Name: "Project Nightjar", Description: "covert BETADESC2", ChunkIDs: []string{graphChunkNodeID("b1", 0)}},
		},
		Relations: []ToolRelationInput{{From: "Kestrel Vale", To: "Project Nightjar", Type: "funds", ChunkIDs: []string{graphChunkNodeID("b1", 0)}}},
	}); err != nil {
		t.Fatal(err)
	}

	ctx := WithCallerScope(bg, CallerScope{Key: "alpha-key", Knowledge: func(c string) bool { return c == "alpha" }})
	tools := db.GraphRAGTools()
	km := db.KnowledgeMemory()
	kv, nj := EntityNodeID("Kestrel Vale"), EntityNodeID("Project Nightjar")
	markers := []string{"BETASECRET", "BETADESC", "Nightjar", "b1", "beta"}

	probe := func(name string, out any, err error) {
		t.Helper()
		if err != nil {
			t.Logf("%-40s error: %v", name, err)
			return
		}
		raw, _ := json.Marshal(out)
		var found []string
		for _, m := range markers {
			if strings.Contains(string(raw), m) {
				found = append(found, m)
			}
		}
		t.Logf("%-40s leaks=%v alphaText=%v", name, found, strings.Contains(string(raw), "alpha mill"))
	}

	r1, e1 := tools.SearchChunksByEntities(ctx, ToolSearchChunksByEntitiesRequest{EntityNames: []string{"Kestrel Vale"}})
	probe("search_chunks_by_entities", r1, e1)
	r2, e2 := tools.SearchPaths(ctx, ToolSearchPathsRequest{EntityNames: []string{"Kestrel Vale"}, WithText: true})
	probe("search_paths", r2, e2)
	r3, e3 := km.ExpandEntityContext(ctx, KnowledgeMemoryExpandEntityContextRequest{EntityNames: []string{"Kestrel Vale"}})
	probe("knowledge_memory_expand_entity_context", r3, e3)
	r4, e4 := tools.ExpandGraph(ctx, ToolExpandGraphRequest{NodeIDs: []string{kv}, MaxHops: 2})
	probe("expand_graph", r4, e4)
	r5, e5 := tools.GetNodes(ctx, ToolGetNodesRequest{NodeIDs: []string{kv, nj}})
	probe("get_nodes", r5, e5)
	r6, e6 := tools.FindNodes(ctx, ToolFindNodesRequest{Names: []string{"Nightjar"}})
	probe("find_nodes", r6, e6)
	r7, e7 := km.Neighbors(ctx, KnowledgeMemoryNeighborsRequest{EntityName: "Kestrel Vale", MaxDepth: 2})
	probe("knowledge_memory_neighbors", r7, e7)
	r8, e8 := km.ShortestPath(ctx, KnowledgeMemoryShortestPathRequest{FromEntityName: "Kestrel Vale", ToEntityName: "Project Nightjar"})
	probe("knowledge_memory_shortest_path", r8, e8)

	// cortex_query, with the collection the port forces, every lane.
	r9, e9 := db.Query(ctx, QueryRequest{Collection: "alpha", Query: "Kestrel Vale funds", EntityNames: []string{"Kestrel Vale", "Project Nightjar"}, Limit: 10})
	probe("cortex_query default lanes (alpha)", r9, e9)
	// A beta-only entity as the graph seed, three hops: does an alpha chunk
	// come back because beta's graph links it (structural oracle)?
	r10, e10 := db.Query(ctx, QueryRequest{Collection: "alpha", Limit: 10,
		Prefetch: []QueryPrefetch{{Kind: QueryPrefetchGraph, EntityNames: []string{"Project Nightjar"}, MaxHops: 3}}})
	probe("cortex_query graph lane, beta-only seed", r10, e10)
	for _, hops := range []int{1, 2} {
		rh, eh := db.Query(ctx, QueryRequest{Collection: "alpha", Limit: 10,
			Prefetch: []QueryPrefetch{{Kind: QueryPrefetchGraph, EntityNames: []string{"Project Nightjar"}, MaxHops: hops}}})
		probe("cortex_query graph lane, beta seed, hops", rh, eh)
		t.Logf("  hops=%d", hops)
	}
	r10b, e10b := db.Query(ctx, QueryRequest{Collection: "alpha", Limit: 10,
		Prefetch: []QueryPrefetch{{Kind: QueryPrefetchGraph, EntityNames: []string{"Nobody At All"}, MaxHops: 3}}})
	probe("cortex_query graph lane, unknown seed", r10b, e10b)
	// Positive control: knowledge_search in graph mode, collection alpha,
	// where the walk gate of #70 applies.
	r11, e11 := db.SearchKnowledge(ctx, KnowledgeSearchRequest{Collection: "alpha", Query: "Kestrel Vale", RetrievalMode: RetrievalModeGraph, EntityNames: []string{"Kestrel Vale"}})
	probe("knowledge_search graph (alpha)", r11, e11)
}
