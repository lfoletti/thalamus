package relay

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// cut keeps the text up to the end of its n-th word, spacing included, and
// says how many words the whole text has. n <= 0, or a text of n words or
// fewer, is kept whole.
func TestCutKeepsTheFirstWords(t *testing.T) {
	for _, c := range []struct {
		text  string
		n     int
		want  string
		words int
		cut   bool
	}{
		{"one two  three\nfour five", 3, "one two  three", 5, true},
		{"  one two three", 2, "  one two", 3, true},
		{"one two three", 3, "one two three", 3, false},
		{"one two", 5, "one two", 2, false},
		{"one two three", 0, "one two three", 3, false},
		{"", 3, "", 0, false},
	} {
		got, words, cut := cut(c.text, c.n)
		if got != c.want || words != c.words || cut != c.cut {
			t.Fatalf("cut(%q, %d) = %q %d %v", c.text, c.n, got, words, cut)
		}
	}
}

// excerptBench is the pyramid bench with team/acme served in excerpts of 3
// words: the rights of a publisher whose sources are not the gateway's to
// hand out whole.
func excerptBench(t *testing.T) (*bench, string) {
	t.Helper()
	b := startWith(t, func(y string) string {
		return strings.Replace(y, "    collection: team-acme\n", "    collection: team-acme\n    excerpt_words: 3\n    pyramid: {collection: acme-pyramid, sources: preuves}\n", 1)
	})
	k, err := keys.IssueScoped(b.keysFile, "lfoletti", "team/acme", []string{"team-acme", "acme-pyramid"}, rights.Write)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write); err != nil {
		t.Fatal(err)
	}
	b.brain.docs = map[string]map[string]any{
		"plan.md":         {"id": "plan.md", "collection": "team-acme", "title": "Ledger plan", "source_url": "https://wiki/plan", "content": "The code word is MARMOT."},
		"short.md":        {"id": "short.md", "collection": "team-acme", "title": "Short", "content": "Two words."},
		"pyramide/ledger": {"id": "pyramide/ledger", "collection": "acme-pyramid", "title": "Ledger", "content": "The ledger moves on Sunday.", "metadata": map[string]string{"preuves": `["plan.md","short.md"]`}},
		"bare.md":         {"id": "bare.md", "title": "No collection", "content": "Said by an engine that names no collection."},
	}
	return b, k.Secret
}

type excerptNote struct {
	Words    int    `json:"words"`
	Of       int    `json:"of"`
	Original string `json:"original"`
}

type gotKnowledge struct {
	Knowledge struct {
		Title   string       `json:"title"`
		Content string       `json:"content"`
		Excerpt *excerptNote `json:"excerpt"`
	} `json:"knowledge"`
}

// In a brain served in excerpts, th_get cuts a source document to the
// brain's excerpt_words and says so, with the address of the original. A
// synthesis of the brain's pyramid is the publisher's own text: it is served
// whole. A document the engine places in no collection is cut. A short
// document is whole, and a brain without excerpt_words is read whole.
func TestThGetServesAnExcerptInABrainServedInExcerpts(t *testing.T) {
	b, secret := excerptBench(t)
	s := b.federated(t, secret)
	got := decode[gotKnowledge](t, mustCall(t, s, "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "plan.md"}))
	if k := got.Knowledge; k.Content != "The code word" || k.Title != "Ledger plan" || k.Excerpt == nil || *k.Excerpt != (excerptNote{3, 5, "https://wiki/plan"}) {
		t.Fatalf("source: %+v %+v", k, k.Excerpt)
	}
	got = decode[gotKnowledge](t, mustCall(t, s, "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "pyramide/ledger"}))
	if k := got.Knowledge; k.Content != "The ledger moves on Sunday." || k.Excerpt != nil {
		t.Fatalf("synthesis: %+v", k)
	}
	got = decode[gotKnowledge](t, mustCall(t, s, "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "bare.md"}))
	if k := got.Knowledge; k.Content != "Said by an" || k.Excerpt == nil || k.Excerpt.Original != "" {
		t.Fatalf("no collection: %+v", k)
	}
	res := mustCall(t, s, "th_get", map[string]any{"brain": "team/acme", "knowledge_id": "short.md"})
	got = decode[gotKnowledge](t, res)
	if k := got.Knowledge; k.Content != "Two words." || k.Excerpt != nil || strings.Contains(res.Content[0].(*mcp.TextContent).Text, "excerpt") {
		t.Fatalf("short: %+v", k)
	}
	got = decode[gotKnowledge](t, mustCall(t, s, "th_get", map[string]any{"brain": "team/other", "knowledge_id": "plan.md"}))
	if k := got.Knowledge; k.Content != "The code word is MARMOT." || k.Excerpt != nil {
		t.Fatalf("a brain without excerpts: %+v", k)
	}
}

// th_descend cuts the sources it opens the same way.
func TestThDescendServesExcerptsInABrainServedInExcerpts(t *testing.T) {
	b, secret := excerptBench(t)
	type descent struct {
		Sources []struct {
			KnowledgeID string       `json:"knowledge_id"`
			Content     string       `json:"content"`
			Excerpt     *excerptNote `json:"excerpt"`
		} `json:"sources"`
	}
	got := decode[descent](t, mustCall(t, b.federated(t, secret), "th_descend", map[string]any{"brain": "team/acme", "knowledge_id": "pyramide/ledger"}))
	if len(got.Sources) != 2 || got.Sources[0].Content != "The code word" || got.Sources[0].Excerpt == nil || *got.Sources[0].Excerpt != (excerptNote{3, 5, "https://wiki/plan"}) {
		t.Fatalf("cut source: %+v", got)
	}
	if got.Sources[1].Content != "Two words." || got.Sources[1].Excerpt != nil {
		t.Fatalf("short source: %+v", got.Sources[1])
	}
}

// The per-brain endpoint hands out the engine's raw tools, knowledge_get
// among them, which would serve a source whole: a brain served in excerpts
// is refused there to every key, before the brain is reached.
func TestABrainServedInExcerptsIsServedThroughTheFederatedEndpointOnly(t *testing.T) {
	b, secret := excerptBench(t)
	req, _ := http.NewRequest(http.MethodPost, b.url+"/mcp/team/acme", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden || !strings.Contains(string(body), "served in excerpts") {
		t.Fatalf("%d %s", resp.StatusCode, body)
	}
	if len(b.brain.calls) != 0 {
		t.Fatalf("the brain was reached: %+v", b.brain.calls)
	}
}
