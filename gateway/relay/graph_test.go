package relay

import (
	"encoding/json"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// The graph tools are a closed list of the engine's read-only graph tools,
// each served on /mcp under th_ with the engine's own arguments and the
// brain. The schemas below are those of CortexDB v2.141.0.
func TestTheGraphToolsAreListedWithTheEnginesArguments(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Read)
	list, err := b.federated(t, k.Secret).ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	tools := map[string]*mcp.Tool{}
	var names []string
	for _, tool := range list.Tools {
		tools[tool.Name] = tool
		names = append(names, tool.Name)
	}
	sort.Strings(names)
	want := []string{"th_add", "th_brains", "th_descend", "th_entity", "th_expand_graph", "th_find_nodes", "th_get", "th_get_nodes", "th_pending", "th_recall", "th_review", "th_search_chunks_by_entities", "th_search_paths"}
	if !slices.Equal(names, want) {
		t.Fatalf("tools: %v", names)
	}
	for name, c := range map[string]struct{ properties, required []string }{
		"th_find_nodes":                {[]string{"brain", "limit", "names", "node_types"}, []string{"brain", "names"}},
		"th_expand_graph":              {[]string{"brain", "edge_types", "limit", "max_hops", "node_ids", "node_types", "relation_policies"}, []string{"brain", "node_ids"}},
		"th_get_nodes":                 {[]string{"brain", "node_ids"}, []string{"brain", "node_ids"}},
		"th_search_chunks_by_entities": {[]string{"brain", "entity_names", "max_hops", "top_k"}, []string{"brain", "entity_names"}},
		"th_search_paths":              {[]string{"brain", "decay", "direction", "edge_types", "entity_names", "include_bookkeeping", "max_depth", "max_expansions", "max_paths", "node_ids", "relation_policies", "use_edge_weights", "with_text"}, []string{"brain"}},
	} {
		raw, _ := json.Marshal(tools[name].InputSchema)
		var schema struct {
			Properties map[string]any `json:"properties"`
			Required   []string       `json:"required"`
		}
		_ = json.Unmarshal(raw, &schema)
		var props []string
		for p := range schema.Properties {
			props = append(props, p)
		}
		sort.Strings(props)
		sort.Strings(schema.Required)
		if !slices.Equal(props, c.properties) || !slices.Equal(schema.Required, c.required) {
			t.Fatalf("%s: %v %v", name, props, schema.Required)
		}
		if !strings.Contains(tools[name].Description, "brain served in excerpts") {
			t.Fatalf("%s says nothing of the excerpts: %s", name, tools[name].Description)
		}
	}
}

// A graph tool reaches the brain under the engine's name, with the caller's
// own key for that brain, which confines the walk to the key's part of the
// graph. Only the engine's arguments go on: never the brain, never a
// collection (the key holds it), never an argument the engine does not
// declare. The call is audited as a read, with the ids it names.
func TestAGraphToolIsRelayedWithTheCallersKey(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Read)
	b.brain.replies = map[string]string{"expand_graph": `{"nodes":[{"id":"entity:ledger","content":"Ledger","node_type":"concept"}],"edges":[]}`}
	res := mustCall(t, b.federated(t, k.Secret), "th_expand_graph", map[string]any{
		"brain": "team/acme", "node_ids": []string{"entity:ledger"}, "max_hops": 2, "collection": "other", "stray": "x"})
	got := decode[map[string]any](t, res)
	if nodes, _ := got["nodes"].([]any); len(nodes) != 1 || nodes[0].(map[string]any)["content"] != "Ledger" {
		t.Fatalf("answer: %+v", got)
	}
	if len(b.brain.calls) != 1 || b.brain.calls[0].Name != "expand_graph" {
		t.Fatalf("calls: %+v", b.brain.calls)
	}
	args := b.brain.named("expand_graph")
	if len(args) != 2 || args["max_hops"] != 2.0 || len(args["node_ids"].([]any)) != 1 || b.brain.bearer != "Bearer "+k.Secret {
		t.Fatalf("relayed: %+v with %s", args, b.brain.bearer)
	}
	l := b.auditLines(t)[0]
	if l["tool"] != "th_expand_graph" || l["brain"] != "team/acme" || l["user"] != "mdupont" || l["ok"] != true || l["ids"].(map[string]any)["node_ids"] == nil {
		t.Fatalf("audit: %+v", l)
	}
}

// Every one of the five reaches the engine tool of its name.
func TestEachGraphToolReachesItsEngineTool(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Read)
	s := b.federated(t, k.Secret)
	for _, tool := range []string{"find_nodes", "expand_graph", "get_nodes", "search_chunks_by_entities", "search_paths"} {
		if res := mustCall(t, s, "th_"+tool, map[string]any{"brain": "team/acme", "node_ids": []string{"x"}, "names": []string{"x"}, "entity_names": []string{"x"}}); res.IsError {
			t.Fatalf("%s: %+v", tool, res.Content[0])
		}
		if c := b.brain.calls[len(b.brain.calls)-1]; c.Name != tool {
			t.Fatalf("th_%s reached %s", tool, c.Name)
		}
	}
}

// A graph tool needs the right to read the brain, and a brain the gateway
// declares. A brain with curators is refused: its graph holds the nodes of
// the documents that wait for review, which the review gate of th_get and
// th_recall keeps from a reader. Neither reaches the brain.
func TestAGraphToolIsRefusedBeforeTheBrain(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Read)
	s := b.federated(t, k.Secret)
	for _, c := range []struct {
		brain, want string
	}{{"org/none", "unknown brain"}, {"team/other", "may not read"}} {
		res := mustCall(t, s, "th_get_nodes", map[string]any{"brain": c.brain, "node_ids": []string{"x"}})
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, c.want) {
			t.Fatalf("%s: %+v", c.brain, res.Content[0])
		}
	}
	withCurator(t, b)
	res := mustCall(t, s, "th_get_nodes", map[string]any{"brain": "team/acme", "node_ids": []string{"x"}})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "review circuit") {
		t.Fatalf("curated: %+v", res.Content[0])
	}
	if len(b.brain.calls) != 0 {
		t.Fatalf("the brain was reached: %+v", b.brain.calls)
	}
}

// A failure of the brain is the tool's error, and audited.
func TestAGraphToolFailureIsAudited(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "mdupont", "team/acme", "team-acme", rights.Read)
	b.brain.failTool = "find_nodes"
	res := mustCall(t, b.federated(t, k.Secret), "th_find_nodes", map[string]any{"brain": "team/acme", "names": []string{"x"}})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "th_find_nodes: ") {
		t.Fatalf("%+v", res.Content[0])
	}
	if l := b.auditLines(t)[0]; l["tool"] != "th_find_nodes" || l["ok"] != false || l["error"] == nil {
		t.Fatalf("audit: %+v", l)
	}
}

// In a brain served in excerpts, no text a graph tool serves is longer than
// the brain's excerpt_words: a chunk's content, a path's cited text, a node's
// content or property. The answer says how many words were kept. A brain
// without excerpts serves the text whole. No vector is served: it is the
// engine's, and it is large.
func TestAGraphToolServesExcerptsInABrainServedInExcerpts(t *testing.T) {
	b, secret := excerptBench(t)
	b.brain.replies = map[string]string{
		"search_chunks_by_entities": `{"chunks":[{"id":"chunk:plan.md:0","document_id":"plan.md","content":"The code word is MARMOT.","metadata":{"_source":"plan.md"}},{"id":"chunk:short.md:0","content":"Two words."}]}`,
		"get_nodes":                 `{"nodes":[{"id":"doc:plan.md","vector":[0.1,0.2],"content":"The code word is MARMOT.","properties":{"quote":"the ledger moves on Sunday"}}]}`,
	}
	s := b.federated(t, secret)
	chunks := decode[cutChunks](t, mustCall(t, s, "th_search_chunks_by_entities", map[string]any{"brain": "team/acme", "entity_names": []string{"ledger"}}))
	if len(chunks.Chunks) != 2 || chunks.Chunks[0].Content != "The code word" || chunks.Chunks[0].ID != "chunk:plan.md:0" || chunks.Chunks[1].Content != "Two words." || chunks.Excerpt == nil || chunks.Excerpt.Words != 3 {
		t.Fatalf("chunks: %+v %+v", chunks, chunks.Excerpt)
	}
	res := mustCall(t, s, "th_get_nodes", map[string]any{"brain": "team/acme", "node_ids": []string{"doc:plan.md"}})
	text := res.Content[0].(*mcp.TextContent).Text
	if strings.Contains(text, "MARMOT") || strings.Contains(text, "Sunday") || strings.Contains(text, "vector") || !strings.Contains(text, `"the ledger moves"`) {
		t.Fatalf("nodes: %s", text)
	}
	res = mustCall(t, s, "th_get_nodes", map[string]any{"brain": "team/other", "node_ids": []string{"doc:plan.md"}})
	text = res.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "The code word is MARMOT.") || strings.Contains(text, "vector") || strings.Contains(text, "excerpt") {
		t.Fatalf("a brain without excerpts: %s", text)
	}
}

// cutChunks is the answer of th_search_chunks_by_entities with its note.
type cutChunks struct {
	Chunks []struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	} `json:"chunks"`
	Excerpt *struct {
		Words int `json:"words"`
	} `json:"excerpt"`
}

// An answer that is not JSON is the tool's error: it cannot be cut.
func TestAGraphToolRefusesAnAnswerItCannotRead(t *testing.T) {
	b, secret := excerptBench(t)
	b.brain.badTool = "get_nodes"
	res := mustCall(t, b.federated(t, secret), "th_get_nodes", map[string]any{"brain": "team/acme", "node_ids": []string{"x"}})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "th_get_nodes: ") {
		t.Fatalf("%+v", res.Content[0])
	}
}
