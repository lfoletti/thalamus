package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

const withConvention = "    collection: team-acme\n    convention:\n      themes: [tests, git]\n      natures: [rule, pitfall]\n      dated: [pitfall]\n"

func conventionBench(t *testing.T) *bench {
	t.Helper()
	b := startWith(t, func(y string) string { return strings.Replace(y, "    collection: team-acme\n", withConvention, 1) })
	b.brain.missing = true
	return b
}

func text(res *mcp.CallToolResult) string { return res.Content[0].(*mcp.TextContent).Text }

// In a brain with a convention, th_add refuses a document whose metadata or
// id break it, before anything is written; the refusal says what is
// missing and the audit line says it was refused. A well formed document is
// written with its status, current unless said otherwise. A brain without a
// convention takes any document, as before.
func TestAddFollowsTheBrainsConvention(t *testing.T) {
	b := conventionBench(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:fail-first", "content": "A test fails without its code.", "metadata": map[string]any{"theme": "Misc stuff", "nature": "rule"}})
	if !res.IsError || !strings.Contains(text(res), "metadata theme, one of: tests, git") {
		t.Fatalf("bad theme: %s", text(res))
	}
	for _, c := range b.brain.calls {
		if c.Name == "ingest_document" {
			t.Fatal("a refused document reached the brain")
		}
	}
	if l := b.auditLines(t); len(l) == 0 || l[len(l)-1]["tool"] != "th_add" || l[len(l)-1]["ok"] != false || !strings.Contains(l[len(l)-1]["error"].(string), "metadata theme") {
		t.Fatalf("audit: %+v", l)
	}
	res = mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "content": "A test fails without its code.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}})
	if !res.IsError || !strings.Contains(text(res), "<nature>:<theme>:<subject>") {
		t.Fatalf("no id: %s", text(res))
	}
	res = mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:fail-first", "content": "A test fails without its code.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}})
	if res.IsError {
		t.Fatalf("well formed: %s", text(res))
	}
	args := b.brain.named("ingest_document")
	md, _ := args["metadata"].(map[string]any)
	if md["theme"] != "tests" || md["nature"] != "rule" || md["status"] != "current" {
		t.Fatalf("metadata: %v", md)
	}
	// Without a convention, any document goes in, with no status added.
	free := start(t)
	free.brain.missing = true
	kf, _ := keys.Issue(free.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	if res := mustCall(t, free.federated(t, kf.Secret), "th_add", map[string]any{"brain": "team/acme", "content": "Anything."}); res.IsError {
		t.Fatalf("free brain: %s", text(res))
	}
	if md, _ := free.brain.named("ingest_document")["metadata"].(map[string]any); len(md) != 0 {
		t.Fatalf("free metadata: %v", md)
	}
}

// A document the brain already has keeps its id, whatever its form: the
// convention binds new ids, and a curator rewrites an old document under
// its old name.
func TestAnExistingDocumentKeepsItsId(t *testing.T) {
	b := conventionBench(t)
	b.brain.docs = map[string]map[string]any{"state:project:current": {"id": "state:project:current", "author": "lfoletti", "content": "Old state."}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_add", map[string]any{"brain": "team/acme", "document_id": "state:project:current", "content": "New state.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}})
	if res.IsError {
		t.Fatalf("existing id: %s", text(res))
	}
}

// A new document almost identical to one the brain has (here a likeness of
// 9/11, above the default bound of 0.8) is refused, naming the other, unless it says it replaces it or is distinct from it. The
// brain is searched by words for the candidates and each is read whole.
func TestANearCopyIsRefused(t *testing.T) {
	b := conventionBench(t)
	old := "A mutation script checks each pattern once before it mutates the file."
	b.brain.hits = map[string][]fakeHit{"team-acme": {{ID: "rule:tests:mutation-patterns", Title: "Patterns", Snippet: "A mutation script"}, {ID: "rule:tests:other", Title: "Other", Snippet: "x"}}}
	b.brain.docs = map[string]map[string]any{
		"rule:tests:mutation-patterns": {"id": "rule:tests:mutation-patterns", "author": "lfoletti", "content": old},
		"rule:tests:other":             {"id": "rule:tests:other", "author": "lfoletti", "content": "Nothing in common with the new one."},
	}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	md := map[string]any{"theme": "tests", "nature": "rule"}
	add := func(md map[string]any) *mcp.CallToolResult {
		return mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:mutation-check", "content": "A mutation script checks each pattern once before it mutates the files.", "metadata": md})
	}
	res := add(md)
	if !res.IsError || !strings.Contains(text(res), "near copy of rule:tests:mutation-patterns") || strings.Contains(text(res), "rule:tests:other") {
		t.Fatalf("near copy: %s", text(res))
	}
	search := b.brain.named("knowledge_search")
	if search["collection"] != "team-acme" || search["retrieval_mode"] != "lexical" {
		t.Fatalf("search: %v", search)
	}
	for _, key := range []string{"replaces", "distinct_from"} {
		if res := add(map[string]any{"theme": "tests", "nature": "rule", key: "rule:tests:mutation-patterns"}); res.IsError {
			t.Fatalf("%s: %s", key, text(res))
		}
	}
	// Rewriting the same document is not a copy of itself.
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:mutation-patterns", "content": old, "metadata": md}); res.IsError {
		t.Fatalf("rewrite: %s", text(res))
	}
}

// A bound declared in the convention lets a lighter rewrite in; a failed
// search or read refuses the write rather than let a copy through.
func TestNearIsTheBrainsBoundAndAFailedSearchRefuses(t *testing.T) {
	b := startWith(t, func(y string) string {
		return strings.Replace(y, "    collection: team-acme\n", withConvention+"      near: 0.99\n", 1)
	})
	b.brain.missing = true
	b.brain.hits = map[string][]fakeHit{"team-acme": {{ID: "rule:tests:a", Snippet: "x"}}}
	b.brain.docs = map[string]map[string]any{"rule:tests:a": {"id": "rule:tests:a", "author": "lfoletti", "content": "A mutation script checks each pattern once before it mutates the file."}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	args := map[string]any{"brain": "team/acme", "document_id": "rule:tests:b", "content": "A mutation script checks each pattern once before it mutates the code.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}}
	if res := mustCall(t, s, "th_add", args); res.IsError {
		t.Fatalf("below the bound: %s", text(res))
	}
	b.brain.failTool = "knowledge_search"
	if res := mustCall(t, s, "th_add", args); !res.IsError || !strings.Contains(text(res), "near copies") {
		t.Fatalf("failed search: %s", text(res))
	}
	b.brain.failTool, b.brain.badTool = "", "knowledge_search"
	if res := mustCall(t, s, "th_add", args); !res.IsError || !strings.Contains(text(res), "decode knowledge_search") {
		t.Fatalf("bad search answer: %s", text(res))
	}
	b.brain.badTool, b.brain.failID = "", "rule:tests:a"
	if res := mustCall(t, s, "th_add", args); !res.IsError || !strings.Contains(text(res), "near copies") || !strings.Contains(text(res), "no such document rule:tests:a") {
		t.Fatalf("failed candidate read: %s", text(res))
	}
	b.brain.failID = ""
	// A read of the id that fails for another reason than a missing
	// document refuses the write too, for a curator, who skips the owner check.
	withCurator(t, b)
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	b.brain.failTool, b.brain.failID = "", "rule:tests:b"
	if res := mustCall(t, b.federated(t, kc.Secret), "th_add", args); !res.IsError || !strings.Contains(text(res), "no such document") {
		t.Fatalf("failed read: %s", text(res))
	}
}

// POST /v1/add applies the same convention: 400 for a document that breaks
// it, 409 for a near copy.
func TestRestAddFollowsTheConvention(t *testing.T) {
	b := conventionBench(t)
	b.brain.hits = map[string][]fakeHit{"team-acme": {{ID: "rule:tests:a", Snippet: "x"}}}
	b.brain.docs = map[string]map[string]any{"rule:tests:a": {"id": "rule:tests:a", "author": "lfoletti", "content": "The ledger moves on Sunday at noon."}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	body, _ := json.Marshal(map[string]any{"brain": "team/acme", "document_id": "rule:tests:b", "content": "The ledger moves on Sunday at noon.", "metadata": map[string]any{"theme": "tests"}})
	if status, out := b.rest(t, http.MethodPost, "/v1/add", k.Secret, string(body)); status != http.StatusBadRequest || !strings.Contains(out, "metadata nature") {
		t.Fatalf("convention: %d %s", status, out)
	}
	body, _ = json.Marshal(map[string]any{"brain": "team/acme", "document_id": "rule:tests:b", "content": "The ledger moves on Sunday at noon.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}})
	if status, out := b.rest(t, http.MethodPost, "/v1/add", k.Secret, string(body)); status != http.StatusConflict || !strings.Contains(out, "near copy of rule:tests:a") {
		t.Fatalf("near copy: %d %s", status, out)
	}
}

// A theme declared by a document of the brain is a theme like the others;
// a theme the brain does not know yet files the document under unsorted,
// and says so. The brain's themes are read only when the document names a
// theme outside the convention's list, and a failed read refuses the write.
func TestThemesMayBeDeclaredInTheBrain(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = []map[string]string{{"document_id": "theme:ops", "declares_theme": "ops"}, {"document_id": "theme:draft", "declares_theme": "draft", "_review_status": "proposed"}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:x", "content": "A test fails first.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}}); res.IsError {
		t.Fatalf("listed theme: %s", text(res))
	}
	if b.brain.named("aggregate_metadata") != nil {
		t.Fatal("a listed theme read the brain's themes")
	}
	res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:ops:x", "content": "Deploy on Monday.", "metadata": map[string]any{"theme": "ops", "nature": "rule"}})
	if res.IsError || len(res.Content) != 1 {
		t.Fatalf("declared theme: %s", text(res))
	}
	agg := b.brain.named("aggregate_metadata")
	if agg["collection"] != "team-acme" || agg["type"] != "group_by" || fmt.Sprint(agg["group_by"]) != "[declares_theme document_id]" {
		t.Fatalf("themes read: %v", agg)
	}
	// A theme only proposed, not yet accepted, is not a theme.
	res = mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:draft:x", "content": "A draft rule.", "metadata": map[string]any{"theme": "draft", "nature": "rule"}})
	if res.IsError || len(res.Content) != 2 || !strings.Contains(res.Content[1].(*mcp.TextContent).Text, `theme "draft" is not declared: filed under unsorted, theme_proposed draft`) {
		t.Fatalf("undeclared theme: %+v", res.Content)
	}
	var last map[string]any
	for _, c := range b.brain.calls {
		if c.Name == "ingest_document" {
			_ = json.Unmarshal([]byte(c.ArgsJson), &last)
		}
	}
	if md := last["metadata"].(map[string]any); md["theme"] != "unsorted" || md["theme_proposed"] != "draft" {
		t.Fatalf("filed: %v", md)
	}
	b.brain.failTool = "aggregate_metadata"
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:ops:y", "content": "Other.", "metadata": map[string]any{"theme": "ops", "nature": "rule"}}); !res.IsError || !strings.Contains(text(res), "cannot read the brain's themes") {
		t.Fatalf("failed read: %s", text(res))
	}
	b.brain.failTool, b.brain.badTool = "", "aggregate_metadata"
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:ops:y", "content": "Other.", "metadata": map[string]any{"theme": "ops", "nature": "rule"}}); !res.IsError || !strings.Contains(text(res), "decode aggregate_metadata") {
		t.Fatalf("bad read: %s", text(res))
	}
}

// A document theme:<subject> declares a theme: written with declares_theme,
// refused when the brain has the theme or one of a close name.
func TestAThemeDocumentIsWrittenWithItsTheme(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = []map[string]string{{"document_id": "theme:ops", "declares_theme": "ops"}, {"document_id": "theme:deploy", "declares_theme": "deploy"}}
	b.brain.docs = map[string]map[string]any{"theme:ops": {"id": "theme:ops", "author": "lfoletti", "content": "Operations."}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "theme:security", "content": "Attacks on models and their defences."}); res.IsError {
		t.Fatalf("declaration: %s", text(res))
	}
	if md := b.brain.named("ingest_document")["metadata"].(map[string]any); md["declares_theme"] != "security" {
		t.Fatalf("metadata: %v", md)
	}
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "theme:opss", "content": "Operations."}); res.IsError {
		t.Fatalf("three letters are never close: %s", text(res))
	}
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "theme:test", "content": "Testing."}); !res.IsError || !strings.Contains(text(res), `close to the theme "tests"`) {
		t.Fatalf("close: %s", text(res))
	}
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "theme:deploi", "content": "Deploying."}); !res.IsError || !strings.Contains(text(res), `close to the theme "deploy"`) {
		t.Fatalf("close to a declared theme: %s", text(res))
	}
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "theme:ops", "content": "Operations again."}); res.IsError {
		t.Fatalf("rewriting a theme document: %s", text(res))
	}
}

// themeRows is a brain of three themes: tests carries four documents, git
// one, ops (declared by a document) none; six unsorted documents propose
// security (five, a candidate) or test (one, close to tests), and one
// proposes nothing.
func themeRows() []map[string]string {
	rows := []map[string]string{{"document_id": "theme:ops", "declares_theme": "ops"}}
	for i, th := range []string{"tests", "tests", "tests", "tests", "git"} {
		// A document has many records; the engine groups them by document.
		rows = append(rows, map[string]string{"document_id": fmt.Sprintf("rule:%s:%d", th, i), "theme": th})
		rows = append(rows, map[string]string{"document_id": fmt.Sprintf("rule:%s:%d", th, i), "theme": th})
	}
	for i, p := range []string{"security", "security", "security", "security", "security", "test", ""} {
		r := map[string]string{"document_id": fmt.Sprintf("rule:u:%d", i), "theme": "unsorted"}
		if p != "" {
			r["theme_proposed"] = p
		}
		rows = append(rows, r)
	}
	return append(rows, map[string]string{"document_id": "rule:tests:proposed", "theme": "tests", "_review_status": "proposed"})
}

// th_themes tells a reader how the brain's documents are filed: the
// documents of each theme, the unsorted ones and what they propose, the
// candidates (a proposal of at least emerge documents), a proposal close
// to a theme, the themes that carry more than a third of the brain, and
// the themes nothing carries.
func TestThemesTellsHowTheBrainIsFiled(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = themeRows()
	acme := b.rights.Brains["team/acme"]
	acme.Readers = []string{"user:guest"}
	b.rights.Brains["team/acme"] = acme
	k, _ := keys.Issue(b.keysFile, "guest", "team/acme", "team-acme", rights.Read)
	res := mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"})
	if res.IsError {
		t.Fatalf("th_themes: %s", text(res))
	}
	var got struct {
		Documents int `json:"documents"`
		Themes    []struct {
			Theme     string `json:"theme"`
			Documents int    `json:"documents"`
			Declared  string `json:"declared"`
		} `json:"themes"`
		Unsorted struct {
			Theme           string `json:"theme"`
			Documents       int    `json:"documents"`
			WithoutProposal int    `json:"without_proposal"`
		} `json:"unsorted"`
		Proposed []struct {
			Theme     string `json:"theme"`
			Documents int    `json:"documents"`
			Candidate bool   `json:"candidate"`
			CloseTo   string `json:"close_to"`
		} `json:"proposed"`
		Broad   []string `json:"broad"`
		Dormant []string `json:"dormant"`
	}
	if err := json.Unmarshal([]byte(text(res)), &got); err != nil {
		t.Fatal(err)
	}
	themes := fmt.Sprint(got.Themes)
	if got.Documents != 12 || themes != "[{tests 4 convention} {git 1 convention} {ops 0 document}]" {
		t.Fatalf("themes: %d %s", got.Documents, themes)
	}
	if got.Unsorted.Theme != "unsorted" || got.Unsorted.Documents != 7 || got.Unsorted.WithoutProposal != 1 {
		t.Fatalf("unsorted: %+v", got.Unsorted)
	}
	if p := fmt.Sprint(got.Proposed); p != "[{security 5 true } {test 1 false tests}]" {
		t.Fatalf("proposed: %s", p)
	}
	if fmt.Sprint(got.Broad) != "[tests]" || fmt.Sprint(got.Dormant) != "[ops]" {
		t.Fatalf("broad %v dormant %v", got.Broad, got.Dormant)
	}
	if l := b.auditLines(t); l[len(l)-1]["tool"] != "th_themes" || l[len(l)-1]["ok"] != true {
		t.Fatalf("audit: %+v", l[len(l)-1])
	}
}

// th_themes is refused on a brain without a convention, without a read
// right, and when the brain cannot be read.
func TestThemesIsRefusedWithoutAConventionOrARead(t *testing.T) {
	free := start(t)
	k, _ := keys.Issue(free.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	if res := mustCall(t, free.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"}); !res.IsError || !strings.Contains(text(res), "team/acme has no convention") {
		t.Fatalf("no convention: %s", text(res))
	}
	b := conventionBench(t)
	k, _ = keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	if res := mustCall(t, s, "th_themes", map[string]any{"brain": "team/nowhere"}); !res.IsError || !strings.Contains(text(res), "unknown brain") {
		t.Fatalf("unknown: %s", text(res))
	}
	for i, fail := range []string{"declares_theme", "theme", "theme_proposed"} {
		b.brain.rows, b.brain.failTool = themeRows(), ""
		b.brain.failGroup = fail
		if res := mustCall(t, s, "th_themes", map[string]any{"brain": "team/acme"}); !res.IsError || !strings.Contains(text(res), "th_themes") {
			t.Fatalf("%d: %s", i, text(res))
		}
	}
	if l := b.auditLines(t); l[len(l)-1]["tool"] != "th_themes" || l[len(l)-1]["ok"] != false {
		t.Fatalf("audit: %+v", l[len(l)-1])
	}
}

// th_retheme files under a theme the documents that proposed it, as a
// curator: each is read whole and written again with its own author,
// title and source, under the theme, without theme_proposed. The theme
// must be one the brain knows. Accepting a proposal is declaring its
// theme, then th_retheme to it; refusing it is th_retheme to another.
func TestRethemeFilesTheProposals(t *testing.T) {
	b := conventionBench(t)
	withCurator(t, b)
	b.brain.rows = themeRows()
	b.brain.docs = map[string]map[string]any{}
	for i := range 5 {
		id := fmt.Sprintf("rule:u:%d", i)
		b.brain.docs[id] = map[string]any{"id": id, "title": "T" + id, "author": "mdupont", "source_url": "https://x/" + id, "content": "Text " + id, "metadata": map[string]any{"theme": "unsorted", "theme_proposed": "security", "nature": "rule", "date": "2026-11-02"}}
	}
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	s := b.federated(t, kc.Secret)
	res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": "git"})
	if res.IsError || !strings.Contains(text(res), `"moved":["rule:u:0","rule:u:1","rule:u:2","rule:u:3","rule:u:4"]`) {
		t.Fatalf("retheme: %s", text(res))
	}
	var ingests []map[string]any
	for _, c := range b.brain.calls {
		if c.Name == "ingest_document" {
			var a map[string]any
			_ = json.Unmarshal([]byte(c.ArgsJson), &a)
			ingests = append(ingests, a)
		}
	}
	a := ingests[0]
	md := a["metadata"].(map[string]any)
	if len(ingests) != 5 || a["document_id"] != "rule:u:0" || a["author"] != "mdupont" || a["title"] != "Trule:u:0" || a["source_url"] != "https://x/rule:u:0" || a["content"] != "Text rule:u:0" || a["review_status"] != "accepted" || a["collection"] != "team-acme" {
		t.Fatalf("ingest: %v", a)
	}
	if md["theme"] != "git" || md["theme_proposed"] != nil || md["nature"] != "rule" || md["date"] != "2026-11-02" {
		t.Fatalf("metadata: %v", md)
	}
	l := b.auditLines(t)
	if last := l[len(l)-1]; last["tool"] != "th_retheme" || last["ok"] != true || fmt.Sprint(last["ids"].(map[string]any)["moved"]) != "[rule:u:0 rule:u:1 rule:u:2 rule:u:3 rule:u:4]" {
		t.Fatalf("audit: %+v", last)
	}
	// To a theme declared by a document.
	if res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": "ops"}); res.IsError {
		t.Fatalf("declared theme: %s", text(res))
	}
}

// th_retheme is refused to a writer, to a theme the brain does not know or
// to the unsorted theme, and when a read or a write fails, saying what was
// moved before.
func TestRethemeIsRefusedOutsideItsRules(t *testing.T) {
	b := conventionBench(t)
	withCurator(t, b)
	b.brain.rows = themeRows()
	b.brain.docs = map[string]map[string]any{"rule:u:0": {"id": "rule:u:0", "content": "x", "metadata": map[string]any{"theme": "unsorted"}}}
	kw, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	if res := mustCall(t, b.federated(t, kw.Secret), "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": "git"}); !res.IsError || !strings.Contains(text(res), "not a curator") {
		t.Fatalf("writer: %s", text(res))
	}
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	s := b.federated(t, kc.Secret)
	for _, c := range []struct{ theme, want string }{
		{"security", `theme "security" is not declared`},
		{"unsorted", `theme "unsorted" is not declared`},
	} {
		if res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": c.theme}); !res.IsError || !strings.Contains(text(res), c.want) {
			t.Errorf("%s: %s", c.theme, text(res))
		}
	}
	if res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "", "theme": "git"}); !res.IsError || !strings.Contains(text(res), "proposed is required") {
		t.Fatalf("no proposal: %s", text(res))
	}
	b.brain.failID = "rule:u:1"
	if res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": "git"}); !res.IsError || !strings.Contains(text(res), "after moving rule:u:0") || !strings.Contains(text(res), "no such document rule:u:1") {
		t.Fatalf("failed read: %s", text(res))
	}
	b.brain.failID, b.brain.failTool = "", "ingest_document"
	if res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": "git"}); !res.IsError || !strings.Contains(text(res), "after moving nothing") {
		t.Fatalf("failed write: %s", text(res))
	}
	b.brain.failTool = "aggregate_metadata"
	if res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": "git"}); !res.IsError || !strings.Contains(text(res), "th_retheme") {
		t.Fatalf("failed list: %s", text(res))
	}
	b.brain.failTool, b.brain.badTool = "", "knowledge_get"
	if res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": "git"}); !res.IsError || !strings.Contains(text(res), "decode knowledge_get") {
		t.Fatalf("bad read: %s", text(res))
	}
	free := start(t)
	withCurator(t, free)
	kf, _ := keys.Issue(free.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	if res := mustCall(t, free.federated(t, kf.Secret), "th_retheme", map[string]any{"brain": "team/acme", "proposed": "a", "theme": "b"}); !res.IsError || !strings.Contains(text(res), "has no convention") {
		t.Fatalf("no convention: %s", text(res))
	}
}

// th_themes reads: on a brain without audit_reads it is counted, without a
// named line, like the other reads.
func TestThemesIsARead(t *testing.T) {
	b := startWith(t, func(y string) string {
		y = strings.Replace(y, "    collection: team-acme\n", withConvention, 1)
		return strings.Replace(y, "    writers: [user:lfoletti]\n    audit_reads: true\n  team/other:", "    writers: [user:lfoletti]\n  team/other:", 1)
	})
	if b.rights.Brains["team/acme"].AuditReads {
		t.Fatal("the bench still audits the reads of team/acme")
	}
	b.brain.rows = themeRows()
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	if res := mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"}); res.IsError {
		t.Fatalf("th_themes: %s", text(res))
	}
	for _, l := range b.auditLines(t) {
		if l["tool"] == "th_themes" {
			t.Fatalf("a named line: %+v", l)
		}
	}
}

// Arguments of the wrong type reach no brain; proposals of the same count
// come in the order of their names; a failed list of the documents to file
// refuses th_retheme.
func TestThemeToolsEdges(t *testing.T) {
	b := conventionBench(t)
	withCurator(t, b)
	b.brain.rows = []map[string]string{{"document_id": "u1", "theme": "unsorted", "theme_proposed": "zeta"}, {"document_id": "u2", "theme": "unsorted", "theme_proposed": "alpha"}}
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	s := b.federated(t, kc.Secret)
	for _, name := range []string{"th_themes", "th_retheme"} {
		res, _ := s.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: json.RawMessage(`{"brain": 5}`)})
		if res == nil || !res.IsError || !strings.Contains(text(res), name+": json: cannot unmarshal") {
			t.Fatalf("%s bad arguments passed: %+v", name, res)
		}
	}
	if len(b.brain.calls) != 0 {
		t.Fatalf("calls: %v", b.brain.calls)
	}
	res := mustCall(t, s, "th_themes", map[string]any{"brain": "team/acme"})
	if !strings.Contains(text(res), `"proposed":[{"theme":"alpha","documents":1,"candidate":false},{"theme":"zeta","documents":1,"candidate":false}]`) {
		t.Fatalf("order: %s", text(res))
	}
	b.brain.failGroup = "document_id"
	if res := mustCall(t, s, "th_retheme", map[string]any{"brain": "team/acme", "proposed": "zeta", "theme": "git"}); !res.IsError || !strings.Contains(text(res), "groups of document_id") {
		t.Fatalf("failed list: %s", text(res))
	}
}
