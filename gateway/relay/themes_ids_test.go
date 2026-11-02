package relay

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

type namedReport struct {
	WithoutTheme    int      `json:"without_theme"`
	WithoutThemeIDs []string `json:"without_theme_ids"`
	UnnamedIDs      int      `json:"unnamed_ids"`
	Unsorted        struct {
		WithoutProposal    int      `json:"without_proposal"`
		WithoutProposalIDs []string `json:"without_proposal_ids"`
	} `json:"unsorted"`
	Undeclared []struct {
		Theme     string   `json:"theme"`
		Documents int      `json:"documents"`
		IDs       []string `json:"ids"`
	} `json:"undeclared"`
}

func themesOf(t *testing.T, b *bench, secret string) namedReport {
	t.Helper()
	res := mustCall(t, b.federated(t, secret), "th_themes", map[string]any{"brain": "team/acme"})
	var got namedReport
	if err := json.Unmarshal([]byte(text(res)), &got); err != nil {
		t.Fatalf("%v %s", err, text(res))
	}
	return got
}

// th_themes names the documents a curator must file: those without a
// theme, the unsorted ones without a proposal, and those under a theme the
// brain does not know. Each list holds its first 50 ids in byte order,
// whatever order the engine sends; the counts stay whole. A filed document
// and an unsorted one with a proposal are not named. An id the engine
// sends back as a number lost its spelling: it is counted, not named. On a
// brain with audit_reads, the named audit line carries each id once.
func TestThemesNamesTheDocumentsToFile(t *testing.T) {
	if themeIDsLimit != 50 || themeIDsTotal != 200 {
		t.Fatalf("the documented bounds are 50 per list and 200 per call: %d %d", themeIDsLimit, themeIDsTotal)
	}
	b := conventionBench(t)
	b.brain.numeric = true
	rows := []map[string]string{
		{"document_id": "theme:sec", "declares_theme": "sec"},
		{"document_id": "rule:tests:a", "theme": "tests"},
		{"document_id": "rule:u:a", "theme": "unsorted", "theme_proposed": "sec2"},
		{"document_id": "rule:u:c", "theme": "unsorted"},
		{"document_id": "rule:u:b", "theme": "unsorted"},
		{"document_id": "rule:ops:b", "theme": "ops"},
		{"document_id": "rule:ops:a", "theme": "ops"},
		{"document_id": "rule:ops:a", "theme": "ops"},
		{"document_id": "rule:ops:a"},
		{"document_id": "2026", "theme": "ops"},
		{"document_id": "note-b"},
		{"document_id": "note-a"},
		{"document_id": "007"},
		{"document_id": "1e3"},
		{"document_id": "12345678901234567890"},
	}
	// The engine sends the documents in no order: the last ones first.
	for i := 59; i >= 0; i-- {
		rows = append(rows, map[string]string{"document_id": fmt.Sprintf("old:%03d", i), "theme": "legacy"})
	}
	b.brain.rows = rows
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	got := themesOf(t, b, k.Secret)
	if got.WithoutTheme != 6 || fmt.Sprint(got.WithoutThemeIDs) != "[note-a note-b rule:ops:a]" || got.UnnamedIDs != 4 {
		t.Fatalf("without theme: %d %v unnamed %d", got.WithoutTheme, got.WithoutThemeIDs, got.UnnamedIDs)
	}
	if got.Unsorted.WithoutProposal != 2 || fmt.Sprint(got.Unsorted.WithoutProposalIDs) != "[rule:u:b rule:u:c]" {
		t.Fatalf("unsorted: %+v", got.Unsorted)
	}
	if len(got.Undeclared) != 2 {
		t.Fatalf("undeclared: %+v", got.Undeclared)
	}
	legacy, ops := got.Undeclared[0], got.Undeclared[1]
	if legacy.Theme != "legacy" || legacy.Documents != 60 || len(legacy.IDs) != 50 || legacy.IDs[0] != "old:000" || legacy.IDs[49] != "old:049" {
		t.Fatalf("legacy: %s %d %d %v", legacy.Theme, legacy.Documents, len(legacy.IDs), legacy.IDs[:2])
	}
	if ops.Theme != "ops" || ops.Documents != 3 || fmt.Sprint(ops.IDs) != "[rule:ops:a rule:ops:b]" {
		t.Fatalf("ops: %+v", ops)
	}
	l := b.auditLines(t)
	last := l[len(l)-1]
	listed, _ := last["ids"].(map[string]any)["listed"].([]any)
	if last["tool"] != "th_themes" || len(listed) != 3+2+50+1 || listed[0] != "note-a" || listed[3] != "rule:u:b" {
		t.Fatalf("audit: %+v", last)
	}
	for _, id := range listed {
		if s := id.(string); s == "7" || s == "1000" || strings.Contains(s, "e+") {
			t.Fatalf("a number is named as an id: %v", listed)
		}
	}
	// Nothing to file: the lists are empty, and the audit names nothing.
	b.brain.rows = []map[string]string{{"document_id": "rule:tests:a", "theme": "tests"}}
	res := mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"})
	for _, empty := range []string{`"without_theme_ids":[]`, `"without_proposal_ids":[]`, `"unnamed_ids":0`} {
		if !strings.Contains(text(res), empty) {
			t.Fatalf("want %s: %s", empty, text(res))
		}
	}
	l = b.auditLines(t)
	if _, named := l[len(l)-1]["ids"]; named {
		t.Fatalf("audit names ids: %+v", l[len(l)-1])
	}
}

// One call names 200 ids at most, in the order of the lists: past the
// bound, a list is empty and its count still whole.
func TestThemesNamesTwoHundredIDsAtMost(t *testing.T) {
	b := conventionBench(t)
	rows := []map[string]string{{"document_id": "note-a"}, {"document_id": "note-b"}}
	for _, th := range []string{"l1", "l2", "l3", "l4", "l5"} {
		for i := 0; i < 55; i++ {
			rows = append(rows, map[string]string{"document_id": fmt.Sprintf("%s:%02d", th, i), "theme": th})
		}
	}
	b.brain.rows = rows
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	got := themesOf(t, b, k.Secret)
	sizes := []int{len(got.WithoutThemeIDs)}
	for _, u := range got.Undeclared {
		sizes = append(sizes, len(u.IDs))
		if u.Documents != 55 || u.IDs == nil {
			t.Fatalf("%s: %d %v", u.Theme, u.Documents, u.IDs)
		}
	}
	if fmt.Sprint(sizes) != "[2 50 50 50 48 0]" {
		t.Fatalf("sizes: %v", sizes)
	}
	l := b.auditLines(t)
	if listed, _ := l[len(l)-1]["ids"].(map[string]any)["listed"].([]any); len(listed) != 200 {
		t.Fatalf("audit: %d", len(listed))
	}
}
