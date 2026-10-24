package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// entityOut is what th_entity returns.
type entityOut struct {
	Entity    string   `json:"entity"`
	NodeID    string   `json:"node_id"`
	KnownIn   []string `json:"known_in"`
	Neighbors []struct {
		ID       string         `json:"id"`
		Name     string         `json:"name"`
		NodeType string         `json:"node_type"`
		Brains   []string       `json:"brains"`
		Vector   any            `json:"vector"`
		Props    map[string]any `json:"properties"`
	} `json:"neighbors"`
	Skipped []struct {
		Brain  string `json:"brain"`
		Reason string `json:"reason"`
	} `json:"skipped"`
}

// entityBench issues lfoletti a key for both brains and gives each its part
// of the graph around Acme.
func entityBench(t *testing.T, edit func(string) string) (*bench, string, string) {
	t.Helper()
	b := startWith(t, edit)
	acme, err := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	if err != nil {
		t.Fatal(err)
	}
	other, err := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	if err != nil {
		t.Fatal(err)
	}
	b.brain.graph = map[string]fakeGraph{
		acme.Secret: {Known: true, Neighbors: []map[string]any{
			{"id": "entity:zeta", "vector": nil, "content": "Zeta", "properties": map[string]any{"name": "Zeta"}},
			{"id": "entity:ledger", "vector": []float64{0, 1}, "content": "Ledger", "node_type": "System", "properties": map[string]any{"name": "Ledger", "collections": []string{"team-acme"}}},
		}},
		other.Secret: {Known: true, Neighbors: []map[string]any{
			{"id": "entity:bob", "vector": nil, "content": "Bob"},
			{"id": "entity:ledger", "vector": nil, "content": "Ledger"},
		}},
	}
	return b, acme.Secret, other.Secret
}

// th_entity asks every readable brain for the neighbors of one entity with
// the caller's own key for it, and fuses them by node id: a node two brains
// serve is one neighbor that names both, and comes first. known_in lists the
// brains that serve the entity itself. Each brain is asked without a
// collection, as the engine's graph tools take none; the key's scope decides
// what it serves. A neighbor carries its id, name and type, never a vector
// or the properties the engine keeps.
func TestThEntityFusesTheNeighborsAcrossBrains(t *testing.T) {
	b, acme, _ := entityBench(t, func(y string) string { return y })
	s := b.federated(t, acme)
	got := decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme"}))
	if got.Entity != "Acme" || got.NodeID != "entity:acme" || fmt.Sprint(got.KnownIn) != "[team/acme team/other]" || len(got.Skipped) != 0 {
		t.Fatalf("entity: %+v", got)
	}
	var ids []string
	for _, n := range got.Neighbors {
		ids = append(ids, n.ID+"="+strings.Join(n.Brains, "+"))
	}
	if strings.Join(ids, " ") != "entity:ledger=team/acme+team/other entity:bob=team/other entity:zeta=team/acme" {
		t.Fatalf("neighbors: %s", ids)
	}
	if l := got.Neighbors[0]; l.Name != "Ledger" || l.NodeType != "System" || l.Vector != nil || l.Props != nil {
		t.Fatalf("ledger: %+v", l)
	}
	for _, c := range b.brain.calls {
		var args map[string]any
		_ = json.Unmarshal([]byte(c.ArgsJson), &args)
		if _, has := args["collection"]; has {
			t.Fatalf("a graph tool got a collection: %+v", c)
		}
		switch c.Name {
		case "knowledge_memory_neighbors":
			if args["entity_name"] != "Acme" || args["max_depth"] != 1.0 || args["limit"] != 20.0 {
				t.Fatalf("neighbors call: %+v", args)
			}
		case "get_nodes":
			if fmt.Sprint(args["node_ids"]) != "[entity:acme]" {
				t.Fatalf("get_nodes call: %+v", args)
			}
		default:
			t.Fatalf("call: %+v", c)
		}
	}
	if len(b.brain.calls) != 4 {
		t.Fatalf("calls: %+v", b.brain.calls)
	}
	l := b.auditLines(t)[0]
	ids2, _ := l["ids"].(map[string]any)
	if l["tool"] != "th_entity" || l["ok"] != true || fmt.Sprint(ids2["brains"]) != "[team/acme team/other]" || ids2["entity"] != "Acme" ||
		fmt.Sprint(ids2["knowledge_ids"]) != "[team/acme:entity:ledger team/acme:entity:zeta team/other:entity:bob team/other:entity:ledger]" {
		t.Fatalf("audit: %+v", l)
	}
}

// The caller sets the depth and the limit of the walk; the depth is at most
// maxEntityDepth, because each hop multiplies what every brain walks.
func TestThEntityTakesTheDepthAndTheLimit(t *testing.T) {
	b, acme, _ := entityBench(t, func(y string) string { return y })
	s := b.federated(t, acme)
	decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme", "max_depth": 2, "limit": 5}))
	args := b.brain.named("knowledge_memory_neighbors")
	if args["max_depth"] != 2.0 || args["limit"] != 5.0 {
		t.Fatalf("args: %+v", args)
	}
	for _, c := range []struct {
		args map[string]any
		want string
	}{
		{map[string]any{"entity": "Acme", "max_depth": maxEntityDepth + 1}, fmt.Sprintf("max_depth is at most %d", maxEntityDepth)},
		{map[string]any{"entity": "  "}, "entity is required"},
		{map[string]any{}, "entity is required"},
		{map[string]any{"entity": "Acme", "brains": []string{"org/none"}}, "org/none is not a brain lfoletti may read"},
	} {
		res := mustCall(t, s, "th_entity", c.args)
		if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, c.want) {
			t.Fatalf("%v: %+v", c.args, res.Content)
		}
	}
}

// A brain that does not serve the entity is not in known_in, and is not
// skipped either: the entity is unknown there, or hidden from the key. A
// brain whose call fails, or with no key issued, is skipped with why.
func TestThEntitySaysWhereTheEntityIsUnknownOrTheBrainFailed(t *testing.T) {
	b, acme, other := entityBench(t, func(y string) string { return y })
	g := b.brain.graph[other]
	g.Known, g.Neighbors = false, nil
	b.brain.graph[other] = g
	s := b.federated(t, acme)
	got := decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme"}))
	if fmt.Sprint(got.KnownIn) != "[team/acme]" || len(got.Neighbors) != 2 || len(got.Skipped) != 0 {
		t.Fatalf("unknown in other: %+v", got)
	}
	b.brain.graph[other] = fakeGraph{Fail: true}
	got = decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme"}))
	if fmt.Sprint(got.KnownIn) != "[team/acme]" || len(got.Skipped) != 1 || got.Skipped[0].Brain != "team/other" || !strings.Contains(got.Skipped[0].Reason, "graph says no") {
		t.Fatalf("failed other: %+v", got)
	}
	if err := keys.Revoke(b.keysFile, "lfoletti", "team/other"); err != nil {
		t.Fatal(err)
	}
	got = decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme"}))
	if len(got.Skipped) != 1 || got.Skipped[0].Brain != "team/other" || got.Skipped[0].Reason == "" {
		t.Fatalf("no key for other: %+v", got)
	}
}

// brains and skip_brains choose the brains, as for th_recall.
func TestThEntityCanBeRestrictedToSomeBrains(t *testing.T) {
	b, acme, _ := entityBench(t, func(y string) string { return y })
	s := b.federated(t, acme)
	got := decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme", "brains": []string{"team/other"}}))
	if fmt.Sprint(got.KnownIn) != "[team/other]" || len(got.Neighbors) != 2 {
		t.Fatalf("brains: %+v", got)
	}
	got = decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme", "skip_brains": []string{"team/other"}}))
	if fmt.Sprint(got.KnownIn) != "[team/acme]" {
		t.Fatalf("skip_brains: %+v", got)
	}
}

// th_entity reads: a brain without audit_reads is counted, not named, as
// for th_recall.
func TestThEntityIsAuditedAsARead(t *testing.T) {
	b, acme, _ := entityBench(t, func(y string) string {
		return strings.Replace(y, "    weight: 0.5\n    audit_reads: true\n", "    weight: 0.5\n", 1)
	})
	s := b.federated(t, acme)
	decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme"}))
	lines := b.auditLines(t)
	if len(lines) != 1 {
		t.Fatalf("audit: %+v", lines)
	}
	ids, _ := lines[0]["ids"].(map[string]any)
	if fmt.Sprint(ids["brains"]) != "[team/acme]" || fmt.Sprint(ids["knowledge_ids"]) != "[team/acme:entity:ledger team/acme:entity:zeta]" {
		t.Fatalf("audit: %+v", lines[0])
	}
}

// A brain whose answer cannot be read is skipped with why, at either call,
// and the skipped brains come in order of name; arguments that are not an
// object of the schema are refused.
func TestThEntitySkipsABrainItCannotRead(t *testing.T) {
	b, acme, _ := entityBench(t, func(y string) string { return y })
	s := b.federated(t, acme)
	for _, c := range []struct{ bad, fail, want string }{
		{bad: "knowledge_memory_neighbors", want: "decode knowledge_memory_neighbors"},
		{bad: "get_nodes", want: "decode get_nodes"},
		{fail: "get_nodes", want: "brain says no to get_nodes"},
	} {
		b.brain.badTool, b.brain.failTool = c.bad, c.fail
		got := decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme"}))
		if len(got.Skipped) != 2 || got.Skipped[0].Brain != "team/acme" || got.Skipped[1].Brain != "team/other" || !strings.Contains(got.Skipped[0].Reason, c.want) || len(got.KnownIn) != 0 || len(got.Neighbors) != 0 {
			t.Fatalf("%+v: %+v", c, got)
		}
	}
	res, _ := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_entity", Arguments: json.RawMessage(`{"entity": 5}`)})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "th_entity: ") {
		t.Fatalf("bad arguments: %+v", res)
	}
}

// fuseNeighbors names a brain once on a neighbor, even when the brain serves
// the node twice, as a walk of two hops can reach it by two paths.
func TestFuseNeighborsNamesABrainOnce(t *testing.T) {
	known, out := fuseNeighbors([]brainGraph{{Brain: "b", Known: true, Neighbors: []node{{ID: "entity:x"}, {ID: "entity:x"}}}, {Brain: "a"}})
	if fmt.Sprint(known) != "[b]" || len(out) != 1 || fmt.Sprint(out[0].Brains) != "[b]" {
		t.Fatalf("known %v, out %+v", known, out)
	}
}

// A neighbor is an entity: the chunk and document nodes the walk reaches
// from an entity (the passages that mention it) are left out, so their
// text is never served as a name, and the audit names entities only.
func TestThEntityServesEntitiesOnly(t *testing.T) {
	b, acme, other := entityBench(t, func(y string) string { return y })
	b.brain.graph[other] = fakeGraph{Known: true, Neighbors: []map[string]any{
		{"id": "chunk:spec-1:000", "content": "Acme migrates the ledger on Sunday night."},
		{"id": "doc:spec-1", "content": "spec-1"},
		{"id": "entity:bob", "content": "Bob"},
	}}
	s := b.federated(t, acme)
	got := decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme", "brains": []string{"team/other"}}))
	if len(got.Neighbors) != 1 || got.Neighbors[0].ID != "entity:bob" {
		t.Fatalf("neighbors: %+v", got.Neighbors)
	}
	ids, _ := b.auditLines(t)[0]["ids"].(map[string]any)
	if fmt.Sprint(ids["knowledge_ids"]) != "[team/other:entity:bob]" {
		t.Fatalf("audit: %+v", ids)
	}
}

// limit is at most 100: each brain walks up to limit nodes.
func TestThEntityBoundsTheLimit(t *testing.T) {
	b, acme, _ := entityBench(t, func(y string) string { return y })
	s := b.federated(t, acme)
	res := mustCall(t, s, "th_entity", map[string]any{"entity": "Acme", "limit": 101})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "limit is at most 100") {
		t.Fatalf("limit: %+v", res.Content)
	}
	decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme", "limit": 100}))
}

// A brain named twice in brains is asked once.
func TestThEntityAsksABrainNamedTwiceOnce(t *testing.T) {
	b, acme, _ := entityBench(t, func(y string) string { return y })
	s := b.federated(t, acme)
	got := decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme", "brains": []string{"team/acme", "team/acme"}}))
	if fmt.Sprint(got.KnownIn) != "[team/acme]" || len(b.brain.calls) != 2 {
		t.Fatalf("known_in %v, calls %+v", got.KnownIn, b.brain.calls)
	}
}

// A brain that refuses the caller's key is skipped, and the audit names
// the refusal, brain by brain, as for any refused read: next to the named
// line of the brains that answered, or alone.
func TestThEntityAuditsABrainThatRefusedTheKey(t *testing.T) {
	b0, acme0, other0 := entityBench(t, func(y string) string { return y })
	g := b0.brain.graph[other0]
	g.Deny = true
	b0.brain.graph[other0] = g
	decode[entityOut](t, mustCall(t, b0.federated(t, acme0), "th_entity", map[string]any{"entity": "Acme"}))
	l0 := b0.auditLines(t)
	if len(l0) != 2 || l0[0]["brain"] != "" || l0[0]["ok"] != true || l0[1]["brain"] != "team/other" || l0[1]["ok"] != false {
		t.Fatalf("audit: %+v", l0)
	}
	if ids, _ := l0[0]["ids"].(map[string]any); ids["denied"] != nil || fmt.Sprint(ids["brains"]) != "[team/acme]" {
		t.Fatalf("named line: %+v", l0[0])
	}

	b, acme, _ := entityBench(t, func(y string) string {
		return strings.Replace(y, "    weight: 0.5\n    audit_reads: true\n", "    weight: 0.5\n", 1)
	})
	b.brain.denyTool = "knowledge_memory_neighbors"
	s := b.federated(t, acme)
	got := decode[entityOut](t, mustCall(t, s, "th_entity", map[string]any{"entity": "Acme"}))
	if len(got.Skipped) != 2 {
		t.Fatalf("skipped: %+v", got.Skipped)
	}
	var lines []string
	for _, l := range b.auditLines(t) {
		lines = append(lines, fmt.Sprint(l["brain"], " ", l["tool"], " ", l["ok"], " ", strings.Contains(fmt.Sprint(l["error"]), "not for this key")))
	}
	if strings.Join(lines, ";") != "team/acme th_entity false true;team/other th_entity false true" {
		t.Fatalf("audit: %v", lines)
	}
}
