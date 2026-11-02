package relay

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// th_themes names the documents a curator must file: those without a
// theme, and those under a theme the brain does not know, each list in the
// order of the ids and bounded at themeIDsLimit. The counts stay whole. A
// filed or unsorted document is not named. On a brain with audit_reads,
// the named audit line carries the ids it served.
func TestThemesNamesTheDocumentsToFile(t *testing.T) {
	b := conventionBench(t)
	rows := []map[string]string{
		{"document_id": "theme:sec", "declares_theme": "sec"},
		{"document_id": "rule:tests:a", "theme": "tests"},
		{"document_id": "rule:u:a", "theme": "unsorted", "theme_proposed": "sec2"},
		{"document_id": "rule:ops:b", "theme": "ops"},
		{"document_id": "rule:ops:a", "theme": "ops"},
		{"document_id": "rule:ops:a", "theme": "ops"},
		{"document_id": "note-b"},
		{"document_id": "note-a"},
	}
	for i := 0; i < themeIDsLimit+5; i++ {
		rows = append(rows, map[string]string{"document_id": fmt.Sprintf("old:%03d", i), "theme": "legacy"})
	}
	b.brain.rows = rows
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	res := mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"})
	var got struct {
		WithoutTheme    int      `json:"without_theme"`
		WithoutThemeIDs []string `json:"without_theme_ids"`
		Undeclared      []struct {
			Theme     string   `json:"theme"`
			Documents int      `json:"documents"`
			IDs       []string `json:"ids"`
		} `json:"undeclared"`
	}
	if err := json.Unmarshal([]byte(text(res)), &got); err != nil {
		t.Fatalf("%v %s", err, text(res))
	}
	if got.WithoutTheme != 2 || fmt.Sprint(got.WithoutThemeIDs) != "[note-a note-b]" {
		t.Fatalf("without theme: %d %v", got.WithoutTheme, got.WithoutThemeIDs)
	}
	if len(got.Undeclared) != 2 {
		t.Fatalf("undeclared: %+v", got.Undeclared)
	}
	legacy, ops := got.Undeclared[0], got.Undeclared[1]
	if legacy.Theme != "legacy" || legacy.Documents != themeIDsLimit+5 || len(legacy.IDs) != themeIDsLimit || legacy.IDs[0] != "old:000" || legacy.IDs[themeIDsLimit-1] != fmt.Sprintf("old:%03d", themeIDsLimit-1) {
		t.Fatalf("legacy: %s %d %d %v", legacy.Theme, legacy.Documents, len(legacy.IDs), legacy.IDs[:2])
	}
	if ops.Theme != "ops" || ops.Documents != 2 || fmt.Sprint(ops.IDs) != "[rule:ops:a rule:ops:b]" {
		t.Fatalf("ops: %+v", ops)
	}
	l := b.auditLines(t)
	last := l[len(l)-1]
	listed, _ := last["ids"].(map[string]any)["listed"].([]any)
	if last["tool"] != "th_themes" || len(listed) != 2+themeIDsLimit+2 || listed[0] != "note-a" {
		t.Fatalf("audit: %+v", last)
	}
	// Nothing to file: the lists are empty, and the audit names nothing.
	b.brain.rows = []map[string]string{{"document_id": "rule:tests:a", "theme": "tests"}}
	res = mustCall(t, b.federated(t, k.Secret), "th_themes", map[string]any{"brain": "team/acme"})
	if err := json.Unmarshal([]byte(text(res)), &got); err != nil || got.WithoutThemeIDs == nil || len(got.WithoutThemeIDs) != 0 {
		t.Fatalf("empty list: %v %s", err, text(res))
	}
	l = b.auditLines(t)
	if _, named := l[len(l)-1]["ids"]; named {
		t.Fatalf("audit names ids: %+v", l[len(l)-1])
	}
}
