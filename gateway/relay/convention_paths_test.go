package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// The per-brain endpoint hands out the engine's raw tools, ingest_document
// among them, which would write past the convention: a brain with a
// convention is refused there to every key, before the brain is reached.
func TestTheRawRelayRefusesABrainWithAConvention(t *testing.T) {
	b := conventionBench(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	req, _ := http.NewRequest(http.MethodPost, b.url+"/mcp/team/acme", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	req.Header.Set("Authorization", "Bearer "+k.Secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden || !strings.Contains(string(body), "has a convention") {
		t.Fatalf("%d %s", resp.StatusCode, body)
	}
	if len(b.brain.calls) != 0 {
		t.Fatalf("the brain was reached: %+v", b.brain.calls)
	}
}

// Only a document theme:<subject> declares a theme: a declares_theme the
// caller puts on another document is dropped before it is written, and a
// row that carries one under another id declares nothing.
func TestOnlyAThemeDocumentDeclaresATheme(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = []map[string]string{{"document_id": "rule:tests:sneaky", "declares_theme": "test"}, {"document_id": "theme:ops", "declares_theme": "ops"}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:y", "content": "A rule.", "metadata": map[string]any{"theme": "tests", "nature": "rule", "declares_theme": "x"}}); res.IsError {
		t.Fatalf("add: %s", text(res))
	}
	if md := b.brain.named("ingest_document")["metadata"].(map[string]any); md["declares_theme"] != nil {
		t.Fatalf("the caller declared a theme: %v", md)
	}
	res := mustCall(t, s, "th_themes", map[string]any{"brain": "team/acme"})
	if res.IsError || strings.Contains(text(res), `"theme":"test"`) || !strings.Contains(text(res), `{"theme":"ops","documents":0,"declared":"document"}`) {
		t.Fatalf("themes: %s", text(res))
	}
	res = mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:test:z", "content": "Another rule.", "metadata": map[string]any{"theme": "test", "nature": "rule"}})
	if res.IsError || len(res.Content) != 2 {
		t.Fatalf("test is no theme: %+v", res.Content)
	}
}

// POST /v1/add says whose fault a refusal is by its kind, not its words: a
// theme that exists or has a close name is a 400, a near copy a 409 even
// when the other id says "convention". A document filed unsorted is a 200
// whose body says so.
func TestRestStatusesFollowTheKindOfRefusal(t *testing.T) {
	b := conventionBench(t)
	b.brain.hits = map[string][]fakeHit{"team-acme": {{ID: "rule:git:commit-convention", Snippet: "x"}}}
	b.brain.docs = map[string]map[string]any{"rule:git:commit-convention": {"id": "rule:git:commit-convention", "author": "lfoletti", "content": "Each commit message names its rule and its test."}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	post := func(v map[string]any) (int, string) {
		body, _ := json.Marshal(v)
		return b.rest(t, http.MethodPost, "/v1/add", k.Secret, string(body))
	}
	for _, id := range []string{"theme:test", "theme:tests"} {
		if status, out := post(map[string]any{"brain": "team/acme", "document_id": id, "content": "Testing."}); status != http.StatusBadRequest {
			t.Errorf("%s: %d %s", id, status, out)
		}
	}
	if status, out := post(map[string]any{"brain": "team/acme", "document_id": "rule:git:commit-names", "content": "Each commit message names its rule and its test.", "metadata": map[string]any{"theme": "git", "nature": "rule"}}); status != http.StatusConflict || !strings.Contains(out, "near copy of rule:git:commit-convention") {
		t.Fatalf("near copy: %d %s", status, out)
	}
	b.brain.hits = nil
	status, out := post(map[string]any{"brain": "team/acme", "document_id": "rule:caching:ttl", "content": "A cache entry lives an hour.", "metadata": map[string]any{"theme": "caching", "nature": "rule"}})
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); status != http.StatusOK || err != nil || !strings.Contains(fmt.Sprint(got["note"]), "filed under unsorted, theme_proposed caching") || got["echo"] == nil {
		t.Fatalf("unsorted: %d %s", status, out)
	}
	if n := withNote("not json", "a note"); !strings.Contains(n, `"note":"a note"`) || !strings.Contains(n, `"result":"not json"`) {
		t.Fatalf("note on a text answer: %s", n)
	}
}

// th_themes counts every document of the brain: a theme the convention
// lost, a theme whose declaring document is gone, and a document without a
// theme are listed apart, and counted in the total. Theme documents are
// not counted: they file nothing. Undeclared themes come by count, then
// by name.
func TestThemesCountsEveryDocument(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = []map[string]string{
		{"document_id": "theme:sec", "declares_theme": "sec"},
		{"document_id": "rule:tests:a", "theme": "tests"},
		{"document_id": "rule:legacy:a", "theme": "legacy"},
		{"document_id": "rule:archive:a", "theme": "archive"},
		{"document_id": "rule:ops:a", "theme": "ops"},
		{"document_id": "rule:ops:b", "theme": "ops"},
		{"document_id": "old-note"},
	}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"})
	var got struct {
		Documents    int `json:"documents"`
		WithoutTheme int `json:"without_theme"`
		Undeclared   []struct {
			Theme     string `json:"theme"`
			Documents int    `json:"documents"`
		} `json:"undeclared"`
	}
	if err := json.Unmarshal([]byte(text(res)), &got); err != nil || got.Documents != 6 || got.WithoutTheme != 1 || fmt.Sprint(got.Undeclared) != "[{ops 2} {archive 1} {legacy 1}]" {
		t.Fatalf("%v %s", err, text(res))
	}
}

// The engine groups a value that reads as a number as a number: the
// gateway reads it back as text, so a document of such an id is found.
func TestANumericGroupedValueIsReadAsText(t *testing.T) {
	b := conventionBench(t)
	withCurator(t, b)
	b.brain.numeric = true
	b.brain.rows = []map[string]string{{"document_id": "2026", "theme": "unsorted", "theme_proposed": "sec"}}
	b.brain.docs = map[string]map[string]any{"2026": {"id": "2026", "author": "mdupont", "content": "Old.", "metadata": map[string]any{"theme": "unsorted", "theme_proposed": "sec"}}}
	kc, _ := keys.Issue(b.keysFile, "kmlead", "team/acme", "team-acme", rights.Curate)
	if res := mustCall(t, b.federated(t, kc.Secret), "th_retheme", map[string]any{"brain": "team/acme", "proposed": "sec", "theme": "git"}); res.IsError || !strings.Contains(text(res), `"moved":["2026"]`) {
		t.Fatalf("retheme: %s", text(res))
	}
}

// A candidate whose answer does not decode refuses the write: the gate
// does not open because a copy could not be read.
func TestAnUnreadableCandidateRefusesTheWrite(t *testing.T) {
	b := conventionBench(t)
	b.brain.hits = map[string][]fakeHit{"team-acme": {{ID: "rule:tests:a", Snippet: "x"}}}
	b.brain.badID = "rule:tests:a"
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:b", "content": "A rule.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}})
	if !res.IsError || !strings.Contains(text(res), "cannot look for near copies") || !strings.Contains(text(res), "rule:tests:a") {
		t.Fatalf("unreadable candidate: %s", text(res))
	}
}

// A write reads its id once, and checks the owner before it searches the
// brain: a writer who names another author's document reaches only that
// read; a new document costs the read, the search and the write.
func TestAWriteReadsItsIdOnceAndChecksTheOwnerFirst(t *testing.T) {
	b := conventionBench(t)
	b.brain.hits = map[string][]fakeHit{"team-acme": {}}
	b.brain.docs = map[string]map[string]any{"rule:tests:theirs": {"id": "rule:tests:theirs", "author": "mdupont", "content": "Theirs."}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	names := func() string {
		var n []string
		for _, c := range b.brain.calls {
			n = append(n, c.Name)
		}
		b.brain.calls = nil
		return strings.Join(n, ",")
	}
	res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:theirs", "content": "Mine now.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}})
	if !res.IsError || !strings.Contains(text(res), "belongs to mdupont") || names() != "knowledge_get" {
		t.Fatalf("another's: %s", text(res))
	}
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:mine", "content": "Mine.", "metadata": map[string]any{"theme": "tests", "nature": "rule"}}); res.IsError {
		t.Fatalf("new: %s", text(res))
	}
	if n := names(); n != "knowledge_get,knowledge_search,ingest_document" {
		t.Fatalf("calls: %s", n)
	}
}

// The near-copy search asks for the first forty words of the document and
// reads five candidates; a likeness that reaches the bound, here exactly
// 4/5, refuses the write.
func TestTheNearCopySearchKeepsItsBounds(t *testing.T) {
	b := conventionBench(t)
	old := "alpha beta gamma delta epsilon zeta eta"
	b.brain.hits = map[string][]fakeHit{"team-acme": {{ID: "rule:tests:h1", Snippet: "x"}, {ID: "rule:tests:h2", Snippet: "x"}, {ID: "rule:tests:h3", Snippet: "x"}, {ID: "rule:tests:h4", Snippet: "x"}, {ID: "rule:tests:old", Snippet: "x"}}}
	b.brain.docs = map[string]map[string]any{"rule:tests:old": {"id": "rule:tests:old", "author": "lfoletti", "content": old}}
	for i := 1; i <= 4; i++ {
		id := fmt.Sprintf("rule:tests:h%d", i)
		b.brain.docs[id] = map[string]any{"id": id, "author": "lfoletti", "content": "Nothing alike."}
	}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:new", "content": "alpha beta gamma delta epsilon zeta", "metadata": map[string]any{"theme": "tests", "nature": "rule"}})
	if !res.IsError || !strings.Contains(text(res), "near copy of rule:tests:old (likeness 0.80)") {
		t.Fatalf("at the bound: %s", text(res))
	}
	var w []string
	for i := range 50 {
		w = append(w, fmt.Sprintf("w%d", i))
	}
	b.brain.hits = map[string][]fakeHit{"team-acme": {}}
	b.brain.calls = nil
	if res := mustCall(t, s, "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:tests:long", "content": strings.Join(w, " "), "metadata": map[string]any{"theme": "tests", "nature": "rule"}}); res.IsError {
		t.Fatalf("long: %s", text(res))
	}
	if q := b.brain.named("knowledge_search")["query"]; q != strings.Join(w[:40], " ") {
		t.Fatalf("query: %v", q)
	}
}

// A theme is broad when it carries more than a third of the filed
// documents: two of five is, one of three is not.
func TestABroadThemeCarriesMoreThanAThird(t *testing.T) {
	b := conventionBench(t)
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s := b.federated(t, k.Secret)
	for _, tc := range []struct {
		themes []string
		want   string
	}{
		{[]string{"tests", "tests", "git", "git", "git"}, "[tests git]"},
		{[]string{"tests", "git", "git"}, "[git]"},
	} {
		b.brain.rows = nil
		for i, th := range tc.themes {
			b.brain.rows = append(b.brain.rows, map[string]string{"document_id": fmt.Sprintf("rule:%s:%d", th, i), "theme": th})
		}
		var got struct {
			Broad []string `json:"broad"`
		}
		res := mustCall(t, s, "th_themes", map[string]any{"brain": "team/acme"})
		if err := json.Unmarshal([]byte(text(res)), &got); err != nil || fmt.Sprint(got.Broad) != tc.want {
			t.Errorf("%v: %s", tc.themes, text(res))
		}
	}
}

// A document the caller files unsorted itself, naming its proposal, gets
// no note: the gateway changed nothing.
func TestNoNoteWhenTheCallerFiledItUnsorted(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = []map[string]string{}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_add", map[string]any{"brain": "team/acme", "document_id": "rule:sec:a", "content": "A rule.", "metadata": map[string]any{"theme": "unsorted", "theme_proposed": "sec", "nature": "rule"}})
	if res.IsError || len(res.Content) != 1 {
		t.Fatalf("%+v", res.Content)
	}
}

// A proposal close to several themes names the first, in the order of the
// convention then of the declarations; th_themes reads every group, not a
// first page.
func TestAProposalIsCloseToTheFirstCloseTheme(t *testing.T) {
	b := conventionBench(t)
	b.brain.rows = []map[string]string{{"document_id": "theme:tent", "declares_theme": "tent"}, {"document_id": "u1", "theme": "unsorted", "theme_proposed": "test"}, {"document_id": "u2", "theme": "unsorted", "theme_proposed": "zeta"}, {"document_id": "u3", "theme": "unsorted", "theme_proposed": "alpha"}}
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"})
	if !strings.Contains(text(res), `{"theme":"test","documents":1,"candidate":false,"close_to":"tests"}`) || !strings.Contains(text(res), `"theme":"alpha"`) || !strings.Contains(text(res), `"theme":"zeta"`) {
		t.Fatalf("%s", text(res))
	}
}
