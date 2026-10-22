package relay

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lfoletti/thalamus/gateway/rights"
)

// policy has one brain whose reads are named and one whose reads are
// counted.
func policy(t *testing.T, retention string) *rights.Rights {
	t.Helper()
	r, err := rights.Parse([]byte(`
brains:
  org/secrets:
    addr: 127.0.0.1:1
    keys_file: keys.json
    collection: secrets
    readers: [user:lfoletti]
    audit_reads: true
  team/acme:
    addr: 127.0.0.1:1
    keys_file: keys.json
    collection: team-acme
    readers: [user:lfoletti]
` + retention))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// The writes, the reviews, the logins, and the tools the engine does not
// call reads keep a named line. A read keeps a named line only on a brain
// whose reads the rights name; elsewhere it is counted for the brain,
// without a name. th_recall follows the rule brain by brain, and th_brains,
// which reads no brain, leaves nothing.
func TestTheAuditNamesWritesAndCountsReads(t *testing.T) {
	r := policy(t, "")
	recall := Entry{User: "lfoletti", Tool: "th_recall", OK: true, IDs: map[string]any{
		"brains":        []string{"org/secrets", "team/acme"},
		"knowledge_ids": []string{"org/secrets:s-1", "team/acme:a-1", "team/acme:a-2"},
	}}
	for _, c := range []struct {
		name    string
		e       Entry
		named   string
		counted string
	}{
		{"th_add", Entry{User: "lfoletti", Brain: "team/acme", Tool: "th_add", OK: true}, "lfoletti team/acme th_add", ""},
		{"th_review", Entry{User: "lfoletti", Brain: "team/acme", Tool: "th_review", OK: true}, "lfoletti team/acme th_review", ""},
		{"login", Entry{User: "lfoletti", Tool: "login", OK: true}, "lfoletti  login", ""},
		{"a write tool", Entry{User: "lfoletti", Brain: "team/acme", Tool: "ingest_document", OK: true}, "lfoletti team/acme ingest_document", ""},
		{"a tool the engine does not know", Entry{User: "lfoletti", Brain: "team/acme", Tool: "no_such_tool"}, "lfoletti team/acme no_such_tool", ""},
		{"a read tool", Entry{User: "lfoletti", Brain: "team/acme", Tool: "knowledge_get", OK: true}, "", "team/acme ok"},
		{"a failed read", Entry{User: "lfoletti", Brain: "team/acme", Tool: "knowledge_search", Error: "down"}, "", "team/acme failed"},
		{"a refused read", Entry{User: "lfoletti", Brain: "team/acme", Tool: "knowledge_search", Error: "no", Denied: true}, "lfoletti team/acme knowledge_search", ""},
		{"th_get refused", Entry{User: "lfoletti", Brain: "team/acme", Tool: "th_get", Error: "no", Denied: true}, "lfoletti team/acme th_get", ""},
		{"th_get", Entry{User: "lfoletti", Brain: "team/acme", Tool: "th_get", OK: true}, "", "team/acme ok"},
		{"th_pending", Entry{User: "lfoletti", Brain: "team/acme", Tool: "th_pending", OK: true}, "", "team/acme ok"},
		{"a read on a named brain", Entry{User: "lfoletti", Brain: "org/secrets", Tool: "knowledge_get", OK: true}, "lfoletti org/secrets knowledge_get", ""},
		{"th_get on a named brain", Entry{User: "lfoletti", Brain: "org/secrets", Tool: "th_get", OK: true}, "lfoletti org/secrets th_get", ""},
		{"th_brains", Entry{User: "lfoletti", Tool: "th_brains", OK: true}, "", ""},
		{"th_recall", recall, "lfoletti  th_recall [org/secrets] [org/secrets:s-1]", "team/acme ok"},
	} {
		named, counted := decide(c.e, r)
		var got []string
		for _, e := range named {
			s := e.User + " " + e.Brain + " " + e.Tool
			if e.Tool == "th_recall" {
				s += fmt.Sprint(" ", e.IDs["brains"], " ", e.IDs["knowledge_ids"])
			}
			got = append(got, s)
		}
		var tally []string
		for _, n := range counted {
			s := n.Brain + " ok"
			if n.Failed {
				s = n.Brain + " failed"
			}
			tally = append(tally, s)
		}
		if strings.Join(got, ";") != c.named || strings.Join(tally, ";") != c.counted {
			t.Errorf("%s: named %q, counted %q", c.name, got, tally)
		}
	}
	// A recall over brains whose reads are all counted leaves no named line.
	plain := Entry{User: "lfoletti", Tool: "th_recall", OK: true, IDs: map[string]any{"brains": []string{"team/acme"}, "knowledge_ids": []string{"team/acme:a-1"}}}
	if named, counted := decide(plain, r); len(named) != 0 || len(counted) != 1 {
		t.Fatalf("plain recall: %+v, %+v", named, counted)
	}
}

// clock is a settable time for the audit.
type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func openAt(t *testing.T, path string, c *clock) *Audit {
	t.Helper()
	a, err := OpenAudit(path)
	if err != nil {
		t.Fatal(err)
	}
	a.now = c.now
	return a
}

func lines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

// The counted reads of a day are written as one line per brain, without a
// name, when the first call of the next day arrives and when the audit
// closes.
func TestReadCountsAreWrittenWhenTheDayChangesAndAtClose(t *testing.T) {
	r := policy(t, "")
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	c := &clock{time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)}
	a := openAt(t, path, c)
	read := Entry{User: "lfoletti", Brain: "team/acme", Tool: "knowledge_get", OK: true}
	for _, e := range []Entry{read, read, {User: "lfoletti", Brain: "team/acme", Tool: "knowledge_get", Error: "down"}} {
		if err := a.Note(e, r); err != nil {
			t.Fatal(err)
		}
	}
	if data, _ := os.ReadFile(path); len(data) != 0 {
		t.Fatalf("reads written at once: %q", data)
	}
	c.t = c.t.Add(24 * time.Hour)
	if err := a.Note(read, r); err != nil {
		t.Fatal(err)
	}
	got := lines(t, path)
	if len(got) != 1 || !strings.Contains(got[0], `"brain":"team/acme","tool":"reads"`) || !strings.Contains(got[0], `"day":"2026-10-22","calls":3,"failed":1`) || strings.Contains(got[0], "lfoletti") {
		t.Fatalf("day line: %q", got)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	got = lines(t, path)
	if len(got) != 2 || !strings.Contains(got[1], `"day":"2026-10-23","calls":1`) || strings.Contains(got[1], "failed") {
		t.Fatalf("close line: %q", got)
	}
	if err := a.Note(read, r); err == nil {
		t.Fatal("a closed audit accepted a call")
	}
}

// keep drops the lines older than the cutoff. A line whose time cannot be
// read is kept: the audit never loses what it cannot date.
func TestKeepDropsTheLinesBeforeTheCutoff(t *testing.T) {
	data := []byte(`{"ts":"2026-01-01T00:00:00Z","tool":"old"}
{"ts":"2026-10-01T00:00:00Z","tool":"new"}
not json
{"tool":"no time"}
`)
	got := string(keep(data, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)))
	want := `{"ts":"2026-10-01T00:00:00Z","tool":"new"}
not json
{"tool":"no time"}
`
	if got != want {
		t.Fatalf("kept %q", got)
	}
}

// With retention_days, the audit drops the lines older than the retention
// at its first call and at each change of day, and keeps appending after.
// Without it, every line stays.
func TestTheAuditDropsTheLinesOlderThanTheRetention(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	old := `{"ts":"2026-01-01T00:00:00Z","user":"x","key":"","brain":"","tool":"login","ok":true}` + "\n"
	if err := os.WriteFile(path, []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}
	c := &clock{time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)}
	write := Entry{User: "lfoletti", Brain: "team/acme", Tool: "th_add", OK: true}

	a := openAt(t, path, c)
	if err := a.Note(write, policy(t, "")); err != nil {
		t.Fatal(err)
	}
	if got := lines(t, path); len(got) != 2 {
		t.Fatalf("without retention: %q", got)
	}
	_ = a.Close()

	a = openAt(t, path, c)
	r := policy(t, "audit:\n  retention_days: 30\n")
	if err := a.Note(write, r); err != nil {
		t.Fatal(err)
	}
	got := lines(t, path)
	if len(got) != 2 || strings.Contains(strings.Join(got, ""), "2026-01-01") {
		t.Fatalf("first call: %q", got)
	}
	c.t = c.t.Add(31 * 24 * time.Hour)
	if err := a.Note(write, r); err != nil {
		t.Fatal(err)
	}
	if got := lines(t, path); len(got) != 1 || !strings.Contains(got[0], c.t.Format("2006-01-02")) {
		t.Fatalf("next day: %q", got)
	}
	if err := a.Note(write, r); err != nil {
		t.Fatal(err)
	}
	if got := lines(t, path); len(got) != 2 {
		t.Fatalf("append after the drop: %q", got)
	}
	_ = a.Close()
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode().Perm())
	}
}

// A drop that cannot read or write the file is the call's error, and
// names the file. The lines stay as they were.
func TestADropThatCannotReadOrWriteIsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	old := `{"ts":"2026-01-01T00:00:00Z","tool":"login"}` + "\n"
	r := policy(t, "audit:\n  retention_days: 30\n")
	for _, mode := range []os.FileMode{0o200, 0o400} {
		if err := os.WriteFile(path, []byte(old), 0o600); err != nil {
			t.Fatal(err)
		}
		c := &clock{time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)}
		a := openAt(t, path, c)
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
		err := a.Note(Entry{User: "lfoletti", Tool: "login", OK: true}, r)
		_ = a.Close()
		_ = os.Chmod(path, 0o600)
		if err == nil || !strings.Contains(err.Error(), "audit.jsonl") {
			t.Fatalf("mode %v: %v", mode, err)
		}
		if got := lines(t, path); len(got) != 1 || !strings.Contains(got[0], "2026-01-01") {
			t.Fatalf("mode %v: %q", mode, got)
		}
	}
}

// A line the audit cannot write is the call's error: a named line, the
// counts of a day, and the counts written at close.
func TestALineThatCannotBeWrittenIsAnError(t *testing.T) {
	r := policy(t, "")
	read := Entry{User: "lfoletti", Brain: "team/acme", Tool: "knowledge_get", OK: true}
	write := Entry{User: "lfoletti", Brain: "team/acme", Tool: "th_add", OK: true}
	for _, c := range []struct {
		name string
		run  func(a *Audit, c *clock) error
	}{
		{"named", func(a *Audit, c *clock) error { return a.Note(write, r) }},
		{"day", func(a *Audit, c *clock) error { c.t = c.t.Add(24 * time.Hour); return a.Note(read, r) }},
		{"close", func(a *Audit, c *clock) error { return a.Close() }},
	} {
		c0 := &clock{time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)}
		a := openAt(t, filepath.Join(t.TempDir(), "audit.jsonl"), c0)
		if err := a.Note(read, r); err != nil {
			t.Fatal(err)
		}
		_ = a.file.Close() // the descriptor fails under the audit
		if err := c.run(a, c0); err == nil {
			t.Errorf("%s: no error", c.name)
		}
	}
}

// When no line is old enough to go, the drop leaves the file as it is: a
// file the gateway may not rewrite is no error then.
func TestADropWithNothingToDropLeavesTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	recent := `{"ts":"2026-10-21T00:00:00Z","tool":"login"}` + "\n"
	if err := os.WriteFile(path, []byte(recent), 0o600); err != nil {
		t.Fatal(err)
	}
	a := openAt(t, path, &clock{time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)})
	defer func() { _ = a.Close() }()
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(path, 0o600) }()
	if err := a.Note(Entry{Brain: "team/acme", Tool: "knowledge_get", OK: true}, policy(t, "audit:\n  retention_days: 30\n")); err != nil {
		t.Fatalf("drop with nothing to drop: %v", err)
	}
}

// denied is true for the refusals of a key, wrapped or not, and false for
// other failures.
func TestDeniedKnowsTheRefusalsOfAKey(t *testing.T) {
	for _, c := range []struct {
		err  error
		want bool
	}{
		{status.Error(codes.PermissionDenied, "no"), true},
		{status.Error(codes.Unauthenticated, "no"), true},
		{fmt.Errorf("th_get: %w", status.Error(codes.PermissionDenied, "no")), true},
		{status.Error(codes.Unavailable, "down"), false},
		{errors.New("plain"), false},
	} {
		if got := denied(c.err); got != c.want {
			t.Errorf("%v: %v", c.err, got)
		}
	}
}

// Reads is the classification that decide uses, for the audit summary.
func TestReadsClassesTheTools(t *testing.T) {
	for tool, want := range map[string]bool{"th_get": true, "th_descend": true, "knowledge_search": true, "th_add": false, "ingest_document": false, "no_such_tool": false} {
		if Reads(tool) != want {
			t.Errorf("%s: %v", tool, !want)
		}
	}
}
