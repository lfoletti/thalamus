package relay

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// The language of a question is read from its function words, without a
// model: the language with the most of them, and "" (the default) when no
// language has more than the others.
func TestTheLanguageOfAQuestionIsReadFromItsFunctionWords(t *testing.T) {
	for _, c := range []struct{ query, want string }{
		{"What does the regulation require of the providers?", "en"},
		{"Quelles sont les obligations des fournisseurs pour les modèles ?", "fr"},
		{"Was verlangt die Verordnung von den Anbietern und wie ist das geregelt?", "de"},
		{"¿Cuáles son las obligaciones de los proveedores y cómo se aplican?", "es"},
		{"Quali sono gli obblighi dei fornitori e come si applicano nella pratica?", "it"},
		{"Quais são as obrigações dos fornecedores e como elas se aplicam?", "pt"},
		{"Wat zijn de verplichtingen van de aanbieders en hoe worden ze toegepast?", "nl"},
		{"AI Act", ""},                  // no function word
		{"", ""},                        // nothing
		{"de la", ""},                   // as French as Spanish: a tie
		{"C'est l'IA ?", "fr"},          // an elision is split off its word
		{"WHAT IS THE RULE", "en"},      // case does not matter
		{"Où en est la décision", "fr"}, // accented function words count
	} {
		if got := language(c.query); got != c.want {
			t.Fatalf("language(%q) = %q, want %q", c.query, got, c.want)
		}
	}
}

// attributed adds the notices to a JSON object under "attribution",
// leaving the rest of the text as it was; an answer that is not an object,
// or already has the key, gets them as a second text instead. No notice,
// or an error, leaves the result untouched.
func TestAttributedAddsTheNoticesWithoutTouchingTheAnswer(t *testing.T) {
	ns := []notice{{Brain: "b/one", Text: "By One."}, {Brain: "b/two", Text: "By Two."}}
	text := func(r *mcp.CallToolResult, i int) string { return r.Content[i].(*mcp.TextContent).Text }
	res := attributed(&mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: ` {"id": 12345678901234567890, "z": 1} `}}}, ns)
	if got := text(res, 0); got != ` {"id": 12345678901234567890, "z": 1,"attribution":[{"brain":"b/one","text":"By One."},{"brain":"b/two","text":"By Two."}]} ` || len(res.Content) != 1 {
		t.Fatalf("object: %s", got)
	}
	res = attributed(&mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: `{}`}}}, ns[:1])
	if got := text(res, 0); got != `{"attribution":[{"brain":"b/one","text":"By One."}]}` {
		t.Fatalf("empty object: %s", got)
	}
	for _, body := range []string{`[1, 2]`, `null`, `not json`, `{"attribution": "theirs"}`, `{"a": 1} {"b": 2}`} {
		res = attributed(&mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: body}}}, ns)
		if text(res, 0) != body || len(res.Content) != 2 || text(res, 1) != "By One.\nBy Two." {
			t.Fatalf("%s: %+v", body, res.Content)
		}
	}
	plain := &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: `{"a":1}`}}}
	if res := attributed(plain, nil); text(res, 0) != `{"a":1}` || len(res.Content) != 1 {
		t.Fatalf("no notice: %+v", res.Content)
	}
	failed := errResult(errorString("no"))
	if res := attributed(failed, ns); len(res.Content) != 1 || text(res, 0) != "no" {
		t.Fatalf("error: %+v", res.Content)
	}
}

type errorString string

func (e errorString) Error() string { return string(e) }

// attributionBench is the bench with an attribution on team/acme, and none
// on team/other.
func attributionBench(t *testing.T) *bench {
	t.Helper()
	return startWith(t, func(y string) string {
		return strings.Replace(y, "    collection: team-acme\n", "    collection: team-acme\n    attribution:\n      en: Information provided by Example.\n      fr: Information fournie par Example.\n", 1)
	})
}

type attributionOut struct {
	Attribution []struct {
		Brain string `json:"brain"`
		Text  string `json:"text"`
	} `json:"attribution"`
}

func noticesOf(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	var parts []string
	for _, n := range decode[attributionOut](t, res).Attribution {
		parts = append(parts, n.Brain+"="+n.Text)
	}
	return strings.Join(parts, "|")
}

// th_recall ends with the notice of each brain that contributed a result,
// once, in the language of the question; a brain asked but without a
// result, or without an attribution, adds none.
func TestRecallEndsWithTheNoticeOfEachContributingBrain(t *testing.T) {
	b := attributionBench(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	_, _ = keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	b.brain.hits = map[string][]fakeHit{
		"team-acme": {{"a-1", "Plan", "The ledger moves on Sunday."}, {"a-2", "Plan 2", "Two hours of rollback."}},
		"other":     {{"o-1", "Other", "Something else."}},
	}
	s := b.federated(t, k.Secret)
	if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "Quand est-ce que le registre change ?"})); got != "team/acme=Information fournie par Example." {
		t.Fatalf("french: %s", got)
	}
	if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "ledger"})); got != "team/acme=Information provided by Example." {
		t.Fatalf("default: %s", got)
	}
	// Only team/other answers: no notice.
	b.brain.hits["team-acme"] = nil
	res := mustCall(t, s, "th_recall", map[string]any{"query": "ledger"})
	if strings.Contains(res.Content[0].(*mcp.TextContent).Text, "attribution") {
		t.Fatalf("no contribution, still a notice: %s", res.Content[0].(*mcp.TextContent).Text)
	}
	// A brain cut off by top_k contributed nothing: team/other, weighted 2,
	// takes the one place.
	b.rights.Brains["team/other"] = withWeight(b.rights.Brains["team/other"], 2)
	b.brain.hits = map[string][]fakeHit{"team-acme": {{"a-1", "Plan", "Low."}}, "other": {{"o-1", "Other", "High."}}}
	res = mustCall(t, s, "th_recall", map[string]any{"query": "ledger", "top_k": 1})
	if got := decode[recallOut](t, res); len(got.Results) != 1 || got.Results[0].Brain != "team/other" || strings.Contains(res.Content[0].(*mcp.TextContent).Text, "attribution") {
		t.Fatalf("top_k: %s", res.Content[0].(*mcp.TextContent).Text)
	}
	// The same passage in both brains is one result, and team/acme, which
	// has it too, is credited.
	b.brain.hits = map[string][]fakeHit{"team-acme": {{"a-1", "Plan", "Same."}}, "other": {{"o-1", "Other", "Same."}}}
	if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "ledger", "top_k": 1})); got != "team/acme=Information provided by Example." {
		t.Fatalf("also_in: %s", got)
	}
}

func withWeight(b rights.Brain, w float64) rights.Brain { b.RankWeight = &w; return b }

// contributors names each brain a result came from, or also came from,
// once and sorted.
func TestContributorsAreTheBrainsOfTheResults(t *testing.T) {
	got := contributors([]fused{{Brain: "b", AlsoIn: []string{"a"}}, {Brain: "c"}, {Brain: "c"}})
	if strings.Join(got, ",") != "a,b,c" {
		t.Fatalf("%v", got)
	}
}

// th_get and th_descend serve a document of a brain with its notice in the
// default language; a brain without attribution answers as before.
func TestReadsOfADocumentCarryTheNotice(t *testing.T) {
	b := startWith(t, func(y string) string {
		return strings.Replace(y, "    collection: team-acme\n", "    collection: team-acme\n    pyramid: {collection: acme-pyramid, sources: preuves}\n    attribution:\n      en: Information provided by Example.\n      fr: Information fournie par Example.\n", 1)
	})
	k, _ := keys.IssueScoped(b.keysFile, "lfoletti", "team/acme", []string{"team-acme", "acme-pyramid"}, rights.Write)
	other, _ := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	_ = other
	b.brain.docs = map[string]map[string]any{
		"syn":     {"id": "syn", "title": "Syn", "content": "A synthesis.", "metadata": map[string]string{"preuves": "plan.md"}},
		"plan.md": {"id": "plan.md", "title": "Plan", "content": "The code word is MARMOT."},
	}
	s := b.federated(t, k.Secret)
	if got := noticesOf(t, mustCall(t, s, "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "plan.md"})); got != "team/acme=Information provided by Example." {
		t.Fatalf("th_get: %s", got)
	}
	if got := noticesOf(t, mustCall(t, s, "th_descend", map[string]any{"brain": "team/acme", "knowledge_id": "syn"})); got != "team/acme=Information provided by Example." {
		t.Fatalf("th_descend: %s", got)
	}
	res := mustCall(t, s, "th_get", map[string]any{"brain": "team/other", "knowledge_id": "plan.md"})
	if strings.Contains(res.Content[0].(*mcp.TextContent).Text, "attribution") || len(res.Content) != 1 {
		t.Fatalf("no attribution: %+v", res.Content)
	}
}

// The graph tools of the federated endpoint, and the per-brain endpoint,
// serve a brain's notice with each answer.
func TestGraphReadsAndTheBrainsEndpointCarryTheNotice(t *testing.T) {
	b := attributionBench(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	if got := noticesOf(t, mustCall(t, s, "th_find_nodes", map[string]any{"brain": "team/acme", "query": "ledger"})); got != "team/acme=Information provided by Example." {
		t.Fatalf("th_find_nodes: %s", got)
	}
	direct, err := b.session(t, "team/acme", k.Secret)
	if err != nil {
		t.Fatal(err)
	}
	defer direct.Close()
	res, err := direct.CallTool(context.Background(), &mcp.CallToolParams{Name: "knowledge_get", Arguments: map[string]any{"knowledge_id": "x"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := noticesOf(t, res); got != "team/acme=Information provided by Example." {
		t.Fatalf("per-brain endpoint: %s", got)
	}
}

// th_entity serves the notice of each brain that knows the entity or
// serves one of its neighbors.
func TestEntityCarriesTheNoticeOfItsBrains(t *testing.T) {
	b := attributionBench(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	o, _ := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	b.brain.graph = map[string]fakeGraph{
		k.Secret: {Known: true, Neighbors: []map[string]any{{"id": "n:1", "content": "One", "node_type": "concept"}}},
		o.Secret: {Known: true},
	}
	s := b.federated(t, k.Secret)
	res := mustCall(t, s, "th_entity", map[string]any{"entity": "Ledger"})
	if got := noticesOf(t, res); got != "team/acme=Information provided by Example." {
		var raw map[string]json.RawMessage
		_ = json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &raw)
		t.Fatalf("th_entity: %s (%s)", got, res.Content[0].(*mcp.TextContent).Text)
	}
}
