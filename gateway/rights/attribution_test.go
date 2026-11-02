package rights

import (
	"strings"
	"testing"
)

func withAttribution(block string) string {
	return strings.Replace(sample, "    collection: team-acme\n", "    collection: team-acme\n"+block, 1)
}

// A brain may carry an attribution: the notice that ends every context
// served from it, one line per language, English required as the
// fallback. A brain without one serves no notice.
func TestABrainCarriesItsAttribution(t *testing.T) {
	r := parse(t, withAttribution("    attribution:\n      en: Information provided by Example.\n      fr: \"Information fournie par Example.\"\n"))
	a := r.Brains["team/acme"].Attribution
	for _, c := range []struct{ lang, want string }{
		{"fr", "Information fournie par Example."},
		{"en", "Information provided by Example."},
		{"de", "Information provided by Example."}, // no German line: the English one
		{"", "Information provided by Example."},
	} {
		if got := a.Notice(c.lang); got != c.want {
			t.Fatalf("Notice(%q) = %q, want %q", c.lang, got, c.want)
		}
	}
	if got := r.Brains["org/reference"].Attribution.Notice("fr"); got != "" {
		t.Fatalf("a brain without attribution says %q", got)
	}
}

// rights check refuses an attribution a reader could not rely on: a
// language that is not a two-letter code, an empty or multi-line notice,
// one longer than MaxNotice, and one without its English fallback.
func TestAnAttributionIsChecked(t *testing.T) {
	long := strings.Repeat("x", MaxNotice+1)
	for _, c := range []struct{ block, want string }{
		{"    attribution:\n      fr: Information fournie par Example.\n", `needs an "en" notice`},
		{"    attribution:\n      en: Provided by Example.\n      FR: Fourni par Example.\n", `language "FR" must be a two-letter lower case code`},
		{"    attribution:\n      en: Provided by Example.\n      fra: Fourni par Example.\n", `language "fra" must be a two-letter lower case code`},
		{"    attribution:\n      en: \"  \"\n", `the "en" notice is empty`},
		{"    attribution:\n      en: \"Provided\\nby Example.\"\n", `the "en" notice must be one line`},
		{"    attribution:\n      en: \"Provided\\rby Example.\"\n", `the "en" notice must be one line`},
		{"    attribution:\n      en: " + long + "\n", "is longer than"},
		{"    attribution: {}\n", `needs an "en" notice`},
	} {
		_, err := Parse([]byte(withAttribution(c.block)))
		if err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "team/acme") {
			t.Fatalf("%q: got %v, want %q", c.block, err, c.want)
		}
	}
	// At the bound, it passes.
	r := parse(t, withAttribution("    attribution:\n      en: "+strings.Repeat("x", MaxNotice)+"\n"))
	if got := r.Brains["team/acme"].Attribution.Notice("en"); len(got) != MaxNotice {
		t.Fatalf("notice at the bound: %d", len(got))
	}
	// The bound counts characters, not bytes.
	parse(t, withAttribution("    attribution:\n      en: "+strings.Repeat("é", MaxNotice)+"\n"))
}

// The notice is served trimmed.
func TestANoticeIsTrimmed(t *testing.T) {
	r := parse(t, withAttribution("    attribution:\n      en: \"  Provided by Example.  \"\n"))
	if got := r.Brains["team/acme"].Attribution.Notice("en"); got != "Provided by Example." {
		t.Fatalf("%q", got)
	}
}
