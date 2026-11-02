package relay

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// The language of a question is read from its function words, without a
// model: the language with the most of them. English sharing the lead, or
// no function word at all, is "" (the fallback, English). A tie between
// other languages is broken by their accented letters, then by the order
// of rights.Languages: English never wins on zero words.
func TestTheLanguageOfAQuestionIsReadFromItsFunctionWords(t *testing.T) {
	for _, c := range []struct{ query, want string }{
		{"What does the regulation require of the providers?", "en"},
		{"Quelles sont les obligations des fournisseurs pour les modèles ?", "fr"},
		{"Was verlangt die Verordnung von den Anbietern und wie ist das geregelt?", "de"},
		{"¿Cuáles son las obligaciones de los proveedores y cómo se aplican?", "es"},
		{"Quali sono gli obblighi dei fornitori e come si applicano nella pratica?", "it"},
		{"Quais são as obrigações dos fornecedores e como elas se aplicam?", "pt"},
		{"Wat zijn de verplichtingen van de aanbieders en hoe worden ze toegepast?", "nl"},
		{"AI Act", ""},  // no function word
		{"", ""},        // nothing
		{"de la", "fr"}, // as French as Spanish: the order decides
		{"the de", ""},  // English shares the lead: the fallback
		{"RGPD : délai de conservation ?", "fr"},
		{"Wo finde ich die Regel?", "de"},
		{"Qual è la scadenza?", "it"},
		{"¿Dónde está?", "es"},
		{"de mañana", "es"},               // a tie, broken by ñ
		{"de ação", "pt"},                 // ã and ç: Portuguese over French
		{"Qu’est-ce que le RGPD ?", "fr"}, // the typographic apostrophe splits too
		{"C'est l'IA ?", "fr"},            // an elision is split off its word
		{"WHAT IS THE RULE", "en"},        // case does not matter
		{"Où en est la décision", "fr"},   // accented function words count
		{"die Regel", "de"},               // a tie of de and nl: the order decides
		{"de ça", "fr"},                   // ç, in fr and pt: the order decides
		{"¿de?", "es"},                    // ¿ marks Spanish
		{"ou\u0300 ?", "fr"},              // decomposed (NFD): one word, où
		{"de man\u0303ana", "es"},         // a decomposed ñ counts
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
	res = attributed(&mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "{ }\n"}}}, ns[:1])
	if got := text(res, 0); got != `{ "attribution":[{"brain":"b/one","text":"By One."}]}`+"\n" {
		t.Fatalf("blank object: %q", got)
	}
	// An answer without a first text gets the notices as a text.
	for _, content := range [][]mcp.Content{nil, {&mcp.ImageContent{Data: []byte("x"), MIMEType: "image/png"}}} {
		res = attributed(&mcp.CallToolResult{Content: content}, ns[:1])
		if got := text(res, len(res.Content)-1); got != "By One." || len(res.Content) != len(content)+1 {
			t.Fatalf("content %v: %+v", content, res.Content)
		}
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
		k.Secret: {Known: true, Neighbors: []map[string]any{{"id": "entity:bob", "vector": nil, "content": "Bob"}}},
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

// Two attributed brains are each credited, each in its own words: a brain
// without the question's language gives its English line, so one context
// may mix languages. A brain cut off by top_k is not credited.
func TestRecallCreditsEachAttributedBrain(t *testing.T) {
	b := startWith(t, func(y string) string {
		y = strings.Replace(y, "    collection: team-acme\n", "    collection: team-acme\n    attribution: {en: By Acme., fr: Par Acme.}\n", 1)
		return strings.Replace(y, "    collection: other\n", "    collection: other\n    attribution: {en: By Other.}\n", 1)
	})
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	_, _ = keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	b.brain.hits = map[string][]fakeHit{"team-acme": {{"a-1", "Plan", "One."}}, "other": {{"o-1", "Other", "Two."}}}
	s := b.federated(t, k.Secret)
	if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "quand est le plan ?"})); got != "team/acme=Par Acme.|team/other=By Other." {
		t.Fatalf("two brains: %s", got)
	}
	// lang overrides the detection.
	if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "plan", "lang": "fr"})); got != "team/acme=Par Acme.|team/other=By Other." {
		t.Fatalf("lang: %s", got)
	}
	// A code in capitals, or a locale, names its language.
	for _, lang := range []string{"FR", "fr_CH.UTF-8", "fr-CH"} {
		if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "plan", "lang": lang})); got != "team/acme=Par Acme.|team/other=By Other." {
			t.Fatalf("lang %q: %s", lang, got)
		}
	}
	// A lang the gateway does not serve, or not a string, is read from the
	// query instead, and the recall answers.
	for _, lang := range []any{"ja", "fra", "C", 12, nil} {
		if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "plan", "lang": lang})); got != "team/acme=By Acme.|team/other=By Other." {
			t.Fatalf("lang %v: %s", lang, got)
		}
		if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "quand est le plan ?", "lang": lang})); got != "team/acme=Par Acme.|team/other=By Other." {
			t.Fatalf("lang %v, a French query: %s", lang, got)
		}
	}
	if got := noticesOf(t, mustCall(t, s, "th_recall", map[string]any{"query": "plan", "top_k": 1})); got != "team/acme=By Acme." {
		t.Fatalf("top_k 1: %s", got)
	}
}

// A brain that serves neighbors of the entity without knowing the entity
// itself is credited too.
func TestEntityCreditsABrainThatServesOnlyNeighbors(t *testing.T) {
	b := startWith(t, func(y string) string {
		return strings.Replace(y, "    collection: other\n", "    collection: other\n    attribution: {en: By Other.}\n", 1)
	})
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	o, _ := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	b.brain.graph = map[string]fakeGraph{
		k.Secret: {Known: true},
		o.Secret: {Neighbors: []map[string]any{{"id": "entity:bob", "vector": nil, "content": "Bob"}}},
	}
	s := b.federated(t, k.Secret)
	res := mustCall(t, s, "th_entity", map[string]any{"entity": "Ledger"})
	if got := noticesOf(t, res); got != "team/other=By Other." {
		t.Fatalf("th_entity: %s (%s)", got, res.Content[0].(*mcp.TextContent).Text)
	}
}

// The per-brain endpoint credits every answer but those of the pure writes,
// which return ids and counts, not the brain's content: a tool the engine
// classes as a write but which serves content (a SPARQL SELECT through
// knowledge_graph_query, reflect, consolidate, extract_conversation), and a
// tool the gateway's engine does not know, carry the notice.
func TestTheBrainsEndpointCreditsAllButPureWrites(t *testing.T) {
	b := attributionBench(t)
	credited := []string{"knowledge_get", "knowledge_graph_query", "knowledge_memory_reflect", "knowledge_memory_consolidate", "extract_conversation", "a_tool_of_a_newer_engine",
		// Writes that return their record: the content itself.
		"knowledge_save", "knowledge_update", "memory_save", "memory_update", "knowledge_memory_remember"}
	// Named here, not read from pureWrites: a pure write left out of it
	// would be credited, and this test says so.
	writes := []string{"ingest_document", "upsert_entities", "upsert_relations", "delete_entities", "delete_document_graph", "knowledge_delete", "memory_delete", "knowledge_graph_upsert", "knowledge_graph_delete"}
	b.brain.tools = append(slices.Clone(credited), writes...)
	b.brain.replies = map[string]string{}
	for _, name := range b.brain.tools {
		b.brain.replies[name] = `{"rows":[{"s":"ledger"}]}`
	}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	direct, err := b.session(t, "team/acme", k.Secret)
	if err != nil {
		t.Fatal(err)
	}
	defer direct.Close()
	call := func(name string) *mcp.CallToolResult {
		res, err := direct.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: map[string]any{"query": "SELECT ?s WHERE { ?s ?p ?o }"}})
		if err != nil || res.IsError {
			t.Fatalf("%s: %v %+v", name, err, res)
		}
		return res
	}
	for _, name := range credited {
		if got := noticesOf(t, call(name)); got != "team/acme=Information provided by Example." {
			t.Fatalf("%s: %s", name, got)
		}
	}
	for _, name := range writes {
		if res := call(name); len(res.Content) != 1 || res.Content[0].(*mcp.TextContent).Text != `{"rows":[{"s":"ledger"}]}` {
			t.Fatalf("a pure write is credited: %s %+v", name, res.Content)
		}
	}
}

// pureWrites are writes of the engine: a read named there would lose its
// notice.
func TestThePureWritesAreWritesOfTheEngine(t *testing.T) {
	if len(pureWrites) == 0 {
		t.Fatal("no pure write")
	}
	for _, name := range pureWrites {
		if access, known := authz.LookupTool(name); !known || access != authz.Write {
			t.Fatalf("%s: known %v, access %v", name, known, access)
		}
	}
}

// A brain without attribution answers th_recall as main does: no
// attribution key at all, and the same bytes whatever lang says, since
// main ignores the argument.
func TestABrainWithoutAttributionAnswersAsBefore(t *testing.T) {
	b := start(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	b.brain.hits = map[string][]fakeHit{"team-acme": {{"a-1", "Plan", "The ledger moves on Sunday."}}}
	s := b.federated(t, k.Secret)
	for _, query := range []string{"ledger", "quand est le registre ?"} {
		want := mustCall(t, s, "th_recall", map[string]any{"query": query})
		if want.IsError {
			t.Fatalf("%+v", want.Content)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(want.Content[0].(*mcp.TextContent).Text), &fields); err != nil || fields == nil {
			t.Fatalf("%v %+v", err, want.Content)
		}
		if _, carried := fields["attribution"]; carried || len(want.Content) != 1 {
			t.Fatalf("an answer without attribution carries one: %+v", want.Content)
		}
		for _, lang := range []any{"fr", "FR", "ja", "fr_CH.UTF-8", 12, true} {
			got := mustCall(t, s, "th_recall", map[string]any{"query": query, "lang": lang})
			if got.IsError || len(got.Content) != 1 || got.Content[0].(*mcp.TextContent).Text != want.Content[0].(*mcp.TextContent).Text {
				t.Fatalf("lang %v: %+v, want %+v", lang, got.Content, want.Content)
			}
		}
	}
}

// The schema names the languages th_recall serves notices in, and says
// what another value does; it has no enum, so a client that checks the
// schema still sends a locale such as fr_CH, which the gateway reads.
func TestTheLangArgumentNamesTheServedLanguages(t *testing.T) {
	b := attributionBench(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	tools, err := b.federated(t, k.Secret).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		if tool.Name != "th_recall" {
			continue
		}
		raw, _ := json.Marshal(tool.InputSchema.(map[string]any)["properties"].(map[string]any)["lang"])
		if !strings.Contains(string(raw), "one of en, fr, de, es, it, pt, nl") || !strings.Contains(string(raw), "read from the query") || strings.Contains(string(raw), "enum") {
			t.Fatalf("lang: %s", raw)
		}
		return
	}
	t.Fatal("no th_recall")
}
