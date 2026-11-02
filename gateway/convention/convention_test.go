package convention

import (
	"strings"
	"testing"

	"github.com/lfoletti/thalamus/gateway/rights"
)

var conv = rights.Convention{Themes: []string{"tests", "git"}, Natures: []string{"rule", "pitfall"}, Dated: []string{"pitfall"}, Near: 0.8}

// Every rule of a write in a brain with a convention, one case each: the
// document's theme and nature come from the brain's closed lists, a dated
// nature carries its day, the status is current or replaced, and a new
// document's id says its nature and theme. A document the brain already has
// keeps its id, whatever its form.
func TestEachRuleOfTheConvention(t *testing.T) {
	ok := map[string]string{"theme": "tests", "nature": "rule"}
	with := func(k, v string) map[string]string {
		m := map[string]string{}
		for a, b := range ok {
			m[a] = b
		}
		if v == "" {
			delete(m, k)
		} else {
			m[k] = v
		}
		return m
	}
	for _, tc := range []struct {
		name, id string
		md       map[string]string
		exists   bool
		want     string // "" for accepted
	}{
		{"a well formed new document", "rule:tests:fail-first", ok, false, ""},
		{"no theme", "rule:tests:x", with("theme", ""), false, "metadata theme, one of: tests, git, or unsorted with metadata theme_proposed"},
		{"a theme that is no key", "rule:tests:x", with("theme", "Misc stuff"), false, "metadata theme, one of: tests, git"},
		{"no nature", "rule:tests:x", with("nature", ""), false, "metadata nature, one of: rule, pitfall"},
		{"a nature outside the list", "rule:tests:x", with("nature", "note"), false, "metadata nature, one of: rule, pitfall"},
		{"a dated nature without its date", "pitfall:tests:x", map[string]string{"theme": "tests", "nature": "pitfall"}, false, "metadata date, as YYYY-MM-DD"},
		{"a dated nature with a bad date", "pitfall:tests:x", map[string]string{"theme": "tests", "nature": "pitfall", "date": "2 Nov 2026"}, false, "metadata date, as YYYY-MM-DD"},
		{"a date that is not a day", "rule:tests:x", with("date", "2026-13-40"), false, "metadata date, as YYYY-MM-DD"},
		{"a dated nature with its date", "pitfall:tests:x", map[string]string{"theme": "tests", "nature": "pitfall", "date": "2026-11-02"}, false, ""},
		{"a status outside current and replaced", "rule:tests:x", with("status", "old"), false, "metadata status, current or replaced"},
		{"a replaced status", "rule:tests:x", with("status", "replaced"), false, ""},
		{"a new document without an id", "", ok, false, "a document_id <nature>:<theme>:<subject>"},
		{"a new id of another form", "fail-first", ok, false, "a document_id <nature>:<theme>:<subject>"},
		{"a new id in upper case", "rule:tests:Fail-First", ok, false, "a document_id <nature>:<theme>:<subject>"},
		{"a new id whose nature differs", "pitfall:tests:x", ok, false, `the document_id starts with "rule:tests:"`},
		{"a new id whose theme differs", "rule:git:x", ok, false, `the document_id starts with "rule:tests:"`},
		{"an existing document keeps its id", "state:project:current", ok, true, ""},
		{"an existing document still needs a theme", "state:project:current", with("theme", ""), true, "metadata theme"},
	} {
		got, err := Check(conv, tc.id, tc.md, tc.exists, nil)
		switch {
		case tc.want == "" && err != nil:
			t.Errorf("%s: refused: %v", tc.name, err)
		case tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)):
			t.Errorf("%s: %v, want %q", tc.name, err, tc.want)
		case tc.want == "" && got["status"] == "":
			t.Errorf("%s: no status", tc.name)
		}
	}
}

// An accepted document without a status is current; Check does not change
// the caller's map, and keeps every other key.
func TestAnAcceptedDocumentIsCurrentUnlessSaidOtherwise(t *testing.T) {
	md := map[string]string{"theme": "tests", "nature": "rule", "source": "doc:x"}
	got, err := Check(conv, "rule:tests:x", md, false, nil)
	if err != nil || got["status"] != "current" || got["source"] != "doc:x" || md["status"] != "" {
		t.Fatalf("%v %v %v", got, err, md)
	}
	got, _ = Check(conv, "rule:tests:x", map[string]string{"theme": "tests", "nature": "rule", "status": "replaced"}, false, nil)
	if got["status"] != "replaced" {
		t.Fatalf("replaced: %v", got)
	}
}

// Likeness is the Jaccard index of the word trigrams, blind to case and
// punctuation; a text of fewer than three words is compared by its words.
func TestLikenessComparesWordTrigrams(t *testing.T) {
	a := "A mutation script checks each pattern once before it mutates."
	for _, tc := range []struct {
		b    string
		want float64
	}{
		{a, 1},
		{"a MUTATION script, checks each pattern once before it mutates", 1},
		{"Nothing in common here at all.", 0},
		{"A mutation script checks each pattern once before it runs.", 7.0 / 9},
	} {
		if got := Likeness(a, tc.b); got < tc.want-1e-9 || got > tc.want+1e-9 {
			t.Errorf("%q: %v, want %v", tc.b, got, tc.want)
		}
	}
	if Likeness("two words", "two words") != 1 || Likeness("two words", "other words") != 1.0/3 {
		t.Fatalf("short texts: %v %v", Likeness("two words", "two words"), Likeness("two words", "other words"))
	}
	if Likeness("", "") != 0 || Likeness("", a) != 0 {
		t.Fatal("an empty text is like nothing")
	}
}

// A theme the brain does not know yet does not stop the write: the
// document is filed under the brain's unsorted theme, with the theme it
// names kept in theme_proposed, and its id names the proposed theme. A
// theme declared by a document of the brain is a theme like the others.
func TestAnUndeclaredThemeIsFiledUnsorted(t *testing.T) {
	got, err := Check(conv, "rule:security:x", map[string]string{"theme": "security", "nature": "rule"}, false, nil)
	if err != nil || got["theme"] != "unsorted" || got["theme_proposed"] != "security" {
		t.Fatalf("undeclared: %v %v", got, err)
	}
	if _, err := Check(conv, "rule:tests:x", map[string]string{"theme": "security", "nature": "rule"}, false, nil); err == nil || !strings.Contains(err.Error(), `starts with "rule:security:"`) {
		t.Fatalf("id of the proposed theme: %v", err)
	}
	got, err = Check(conv, "rule:security:x", map[string]string{"theme": "security", "nature": "rule"}, false, []string{"security"})
	if err != nil || got["theme"] != "security" || got["theme_proposed"] != "" {
		t.Fatalf("declared by a document: %v %v", got, err)
	}
	got, err = Check(conv, "rule:unsorted:x", map[string]string{"theme": "unsorted", "nature": "rule"}, false, nil)
	if err != nil || got["theme"] != "unsorted" {
		t.Fatalf("unsorted without a proposal: %v %v", got, err)
	}
	got, err = Check(conv, "rule:ops:x", map[string]string{"theme": "unsorted", "theme_proposed": "ops", "nature": "rule"}, false, nil)
	if err != nil || got["theme_proposed"] != "ops" {
		t.Fatalf("unsorted with a proposal: %v %v", got, err)
	}
	if _, err := Check(conv, "rule:unsorted:x", map[string]string{"theme": "unsorted", "theme_proposed": "Ops!", "nature": "rule"}, false, nil); err == nil || !strings.Contains(err.Error(), "metadata theme_proposed") {
		t.Fatalf("a proposal that is no key: %v", err)
	}
	// Under a custom name for the unsorted theme.
	c := conv
	c.Unsorted = "a-classer"
	if got, err := Check(c, "rule:ops:x", map[string]string{"theme": "ops", "nature": "rule"}, false, nil); err != nil || got["theme"] != "a-classer" {
		t.Fatalf("custom unsorted: %v %v", got, err)
	}
}

// A document theme:<subject> declares a theme: it needs no theme or nature,
// and is written with declares_theme. A theme that exists already, the
// unsorted theme, or a name close to a theme the brain has is refused,
// unless the declaration says it is distinct from that theme.
func TestAThemeDocumentDeclaresATheme(t *testing.T) {
	got, err := Check(conv, "theme:security", map[string]string{}, false, []string{"ops"})
	if err != nil || got["declares_theme"] != "security" || got["status"] != "current" {
		t.Fatalf("declaration: %v %v", got, err)
	}
	for _, tc := range []struct {
		id, want string
		md       map[string]string
	}{
		{"theme:Security", "a document_id theme:<subject>", nil},
		{"theme:tests", `theme "tests" exists`, nil},
		{"theme:ops", `theme "ops" exists`, nil},
		{"theme:unsorted", `theme "unsorted" exists`, nil},
		{"theme:test", `close to the theme "tests"`, nil},
		{"theme:testss", `close to the theme "tests"`, nil},
	} {
		if _, err := Check(conv, tc.id, tc.md, false, []string{"ops"}); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v, want %q", tc.id, err, tc.want)
		}
	}
	if got, err := Check(conv, "theme:ops", map[string]string{}, true, []string{"ops"}); err != nil || got["declares_theme"] != "ops" {
		t.Fatalf("rewritten: %v %v", got, err)
	}
	if _, err := Check(conv, "theme:test", map[string]string{"distinct_from": "tests"}, false, nil); err != nil {
		t.Fatalf("distinct: %v", err)
	}
}

// Two theme names are close when they are the same once the hyphens and a
// final s are dropped, or when one letter turns one into the other, for
// names of four letters or more.
func TestCloseNames(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"tests", "test", true},
		{"data-sets", "datasets", true},
		{"deploy", "deploi", true},
		{"tests", "testing", false},
		{"git", "gif", false},
		{"git", "gitlab", false},
		{"ops", "opss", false},
		{"open-source-ai", "opensourceai", true},
		{"bugs", "bug", true},
		{"deploy", "deplois", false},
		{"tests", "testss", true},
	} {
		if got := CloseNames(tc.a, tc.b); got != tc.want {
			t.Errorf("%s %s: %v", tc.a, tc.b, got)
		}
	}
}
