package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// A key given twice in one object, at any depth, is a repeat; the same key
// in two objects, or a value given twice in an array, is not. Keys compare
// after unescaping and under the case folding of encoding/json, which maps
// "query" and "QUERY" to the same struct field.
func TestRepeatedKeyFindsAKeyGivenTwiceInOneObject(t *testing.T) {
	for _, c := range []struct {
		raw, key string
		repeated bool
	}{
		{`{}`, "", false},
		{``, "", false},
		{`{"a":1,"b":2}`, "", false},
		{`{"a":1,"a":2}`, "a", true},
		{`{"knowledge_id":"a","knowledge_id":null}`, "knowledge_id", true},
		{`{"a":{"b":1,"b":2}}`, "b", true},
		{`{"a":[{"b":1},{"c":{"d":1,"d":1}}]}`, "d", true},
		{`[{"a":1,"a":1}]`, "a", true},
		{`{"a":{"x":1},"b":{"x":1}}`, "", false},
		{`{"a":{"a":1}}`, "", false},
		{`{"a":[1,1,"a","a"],"b":["a"]}`, "", false},
		{`{"a":{"b":1},"a":2}`, "a", true},
		{`{"a":[],"a":2}`, "a", true},
		{`{"a":"b","b":"a"}`, "", false},
		{`{"a":1,"a":2}`, "a", true},
		{`{"a":1,`, "", false},
		{`not json`, "", false},
		{`"a"`, "", false},
		{`{"query":1,"QUERY":2}`, "QUERY", true},
		{`{"Knowledge_Id":1,"knowledge_ID":2}`, "knowledge_ID", true},
		{"{\"k\":1,\"\u212a\":2}", "\u212a", true},
		{"{\"s\":1,\"\u017f\":2}", "\u017f", true},
		{`{"a":{"Q":1},"q":2}`, "", false},
		{"{\"e\":1,\"\u00e9\":2,\"\u00c9\":3}", "\u00c9", true},
		{"{\"e\":1,\"\u00e9\":2}", "", false},
		{`{"a_b":1,"ab":2}`, "", false},
		{`1`, "", false},
	} {
		key, repeated := repeatedKey([]byte(c.raw))
		if key != c.key || repeated != c.repeated {
			t.Errorf("%s: got %q %v, want %q %v", c.raw, key, repeated, c.key, c.repeated)
		}
	}
}

// The per-brain relay refuses a call whose arguments repeat a key, before
// the brain is reached: two readers of the arguments must see the same
// value. The refusal is a named audit line, even on a brain whose reads are
// counted.
func TestTheRelayRefusesArgumentsThatRepeatAKey(t *testing.T) {
	b := startWith(t, func(y string) string { return strings.ReplaceAll(y, "    audit_reads: true\n", "") })
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	s, err := b.session(t, "team/acme", k.Secret)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "knowledge_get", Arguments: json.RawMessage(`{"knowledge_id":"a","knowledge_id":null}`)})
	if err != nil || !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, `"knowledge_id" is repeated`) {
		t.Fatalf("got %+v, %v", res, err)
	}
	// The same for two spellings of one key under case folding.
	res, err = s.CallTool(context.Background(), &mcp.CallToolParams{Name: "knowledge_get", Arguments: json.RawMessage(`{"knowledge_id":"a","Knowledge_ID":"b"}`)})
	if err != nil || !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, `"Knowledge_ID" is repeated`) {
		t.Fatalf("got %+v, %v", res, err)
	}
	if len(b.brain.calls) != 0 {
		t.Fatalf("the brain was reached: %+v", b.brain.calls)
	}
	lines := b.auditLines(t)
	if len(lines) != 2 {
		t.Fatalf("audit: %+v", lines)
	}
	l := lines[0]
	if l["user"] != "lfoletti" || l["brain"] != "team/acme" || l["tool"] != "knowledge_get" || l["key"] != k.ID || l["ok"] != false || !strings.Contains(l["error"].(string), `"knowledge_id" is repeated`) {
		t.Fatalf("audit line: %+v", l)
	}
}

// foldKey folds as encoding/json matches a key to a struct field: two keys
// fold the same exactly when bytes.EqualFold says they are equal.
func TestFoldKeyAgreesWithEqualFold(t *testing.T) {
	keys := []string{"query", "QUERY", "Query", "k", "K", "K", "s", "S", "ſ", "e", "é", "É", "σ", "ς", "Σ", "a_b", "ab", "", "ß", "ss"}
	for _, x := range keys {
		for _, y := range keys {
			if same, want := foldKey(x) == foldKey(y), bytes.EqualFold([]byte(x), []byte(y)); same != want {
				t.Errorf("%q %q: folded alike %v, EqualFold %v", x, y, same, want)
			}
		}
	}
}
