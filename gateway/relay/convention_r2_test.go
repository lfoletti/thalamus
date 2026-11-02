package relay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// ingests are the arguments of every ingest_document the brain received.
func (b *brain) ingests() []map[string]any {
	var out []map[string]any
	for _, c := range b.calls {
		if c.Name == "ingest_document" {
			var a map[string]any
			_ = json.Unmarshal([]byte(c.ArgsJson), &a)
			out = append(out, a)
		}
	}
	return out
}

// th_retheme files only the accepted documents: on an engine without the
// review gate, which serves every row, a rejected or a pending proposal is
// not accepted by a move, and a theme document is never moved. A document
// without a title keeps its author.
func TestRethemeMovesOnlyTheAcceptedDocuments(t *testing.T) {
	b := conventionBench(t)
	withCurator(t, b)
	b.brain.noGate = true
	b.brain.rows = []map[string]string{
		{"document_id": "rule:u:ok", "theme": "unsorted", "theme_proposed": "security"},
		{"document_id": "rule:u:rejected", "theme": "unsorted", "theme_proposed": "security", "_review_status": "rejected"},
		{"document_id": "rule:u:pending", "theme": "unsorted", "theme_proposed": "security", "_review_status": "proposed"},
		{"document_id": "theme:security", "theme": "unsorted", "theme_proposed": "security"},
	}
	b.brain.docs = map[string]map[string]any{}
	for _, id := range []string{"rule:u:ok", "rule:u:rejected", "rule:u:pending", "theme:security"} {
		b.brain.docs[id] = map[string]any{"id": id, "author": "mdupont", "content": "Text " + id, "metadata": map[string]any{"theme": "unsorted", "theme_proposed": "security", "nature": "rule"}}
	}
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	res := mustCall(t, b.federated(t, kc.Secret), "th_retheme", map[string]any{"brain": "team/acme", "proposed": "security", "theme": "git"})
	if res.IsError || !strings.Contains(text(res), `"moved":["rule:u:ok"]`) {
		t.Fatalf("retheme: %s", text(res))
	}
	in := b.brain.ingests()
	if len(in) != 1 || in[0]["author"] != "mdupont" || in[0]["title"] != nil {
		t.Fatalf("ingests: %v", in)
	}
}

// A theme declared both in the convention and by a document of the brain,
// as when a grown theme is promoted into rights.yaml, is one theme: listed
// once, from the convention, and counted once in the share that makes a
// theme broad.
func TestThemesCountsATwiceDeclaredThemeOnce(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = []map[string]string{
		{"document_id": "theme:git", "declares_theme": "git"},
		{"document_id": "rule:git:a", "theme": "git"},
		{"document_id": "rule:tests:a", "theme": "tests"},
	}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"})
	var got themesReport
	if err := json.Unmarshal([]byte(text(res)), &got); err != nil {
		t.Fatalf("%v: %s", err, text(res))
	}
	if fmt.Sprint(got.Themes) != "[{tests 1 convention} {git 1 convention}]" || fmt.Sprint(got.Broad) != "[tests git]" || got.Documents != 2 {
		t.Fatalf("themes %v broad %v documents %d", got.Themes, got.Broad, got.Documents)
	}
}

// th_themes speaks of the brain's own unsorted theme, and does not count
// it among the themes the brain no longer knows.
func TestThemesNamesTheBrainsUnsortedTheme(t *testing.T) {
	b := startWith(t, func(y string) string {
		return strings.Replace(y, "    collection: team-acme\n", withConvention+"      unsorted: inbox\n", 1)
	})
	b.brain.rows = []map[string]string{
		{"document_id": "rule:x:a", "theme": "inbox", "theme_proposed": "x"},
		{"document_id": "rule:tests:a", "theme": "tests"},
	}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"})
	var got themesReport
	if err := json.Unmarshal([]byte(text(res)), &got); err != nil {
		t.Fatalf("%v: %s", err, text(res))
	}
	if got.Unsorted.Theme != "inbox" || got.Unsorted.Documents != 1 || len(got.Undeclared) != 0 || fmt.Sprint(got.Proposed) != "[{x 1 false }]" {
		t.Fatalf("unsorted %+v undeclared %v proposed %v", got.Unsorted, got.Undeclared, got.Proposed)
	}
}

// Only a curator rewrites a theme document: a writer's rewrite of their own
// would wait for review, hidden, and take the theme from the brain until a
// curator accepts it again. A writer still proposes a new theme document.
// REST says 403.
func TestOnlyACuratorRewritesAThemeDocument(t *testing.T) {
	b := conventionBench(t)
	withCurator(t, b)
	b.brain.rows = []map[string]string{{"document_id": "theme:ops", "declares_theme": "ops"}}
	b.brain.docs = map[string]map[string]any{"theme:ops": {"id": "theme:ops", "author": "lfoletti", "content": "Operations."}}
	kw, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, kw.Secret)
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "theme:ops", "content": "Operations, again."}); !res.IsError || !strings.Contains(text(res), "only a curator") {
		t.Fatalf("writer rewrite: %s", text(res))
	}
	body, _ := json.Marshal(map[string]any{"brain": "team/acme", "document_id": "theme:ops", "content": "Operations, again."})
	if status, out := b.rest(t, http.MethodPost, "/v1/add", kw.Secret, string(body)); status != http.StatusForbidden {
		t.Fatalf("rest: %d %s", status, out)
	}
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "theme:security", "content": "Attacks and defences."}); res.IsError {
		t.Fatalf("writer proposes: %s", text(res))
	}
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	if res := mustCall(t, b.federated(t, kc.Secret), "th_add", map[string]any{"brain": "team/acme", "document_id": "theme:ops", "content": "Operations, again."}); res.IsError {
		t.Fatalf("curator rewrite: %s", text(res))
	}
}

// A document that proposes a theme the brain already knows, declared by a
// document, is filed under that theme.
func TestAProposalOfAKnownThemeIsFiledUnderIt(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = []map[string]string{{"document_id": "theme:ops", "declares_theme": "ops"}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:ops:a", "content": "Deploy on Monday.", "metadata": map[string]any{"theme": "unsorted", "theme_proposed": "ops", "nature": "rule"}})
	if res.IsError {
		t.Fatalf("add: %s", text(res))
	}
	md := b.brain.named("ingest_document")["metadata"].(map[string]any)
	if md["theme"] != "ops" || md["theme_proposed"] != nil || len(res.Content) != 1 {
		t.Fatalf("metadata %v content %d", md, len(res.Content))
	}
}

// The owner refusal is told by its kind, not its words: over REST it is a
// 409 whatever the document_id says.
func TestTheOwnerRefusalIsAConflictWhateverItsWords(t *testing.T) {
	b := conventionBench(t)
	b.brain.docs = map[string]map[string]any{"unknown brain": {"id": "unknown brain", "author": "mdupont", "content": "x"}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	body, _ := json.Marshal(map[string]any{"brain": "team/acme", "document_id": "unknown brain", "content": "unknown brain may not write is required", "metadata": map[string]any{"theme": "tests", "nature": "rule"}})
	if status, out := b.rest(t, http.MethodPost, "/v1/add", k.Secret, string(body)); status != http.StatusConflict || !strings.Contains(out, "belongs to mdupont") {
		t.Fatalf("owner: %d %s", status, out)
	}
}
