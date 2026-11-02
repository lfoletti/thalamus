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
// language the gateway does not serve, an empty or multi-line notice,
// one longer than MaxNotice, and one without its English fallback.
func TestAnAttributionIsChecked(t *testing.T) {
	long := strings.Repeat("x", MaxNotice+1)
	for _, c := range []struct{ block, want string }{
		{"    attribution:\n      fr: Information fournie par Example.\n", `needs an "en" notice`},
		{"    attribution:\n      en: Provided by Example.\n      FR: Fourni par Example.\n", `language "FR" is not one the gateway serves (en, fr, de, es, it, pt, nl)`},
		{"    attribution:\n      en: Provided by Example.\n      fra: Fourni par Example.\n", `language "fra" is not one the gateway serves`},
		{"    attribution:\n      en: \"  \"\n", `the "en" notice is empty`},
		{"    attribution:\n      en: \"Provided\\nby Example.\"\n", `the "en" notice must be one line`},
		{"    attribution:\n      en: \"Provided\\rby Example.\"\n", `the "en" notice must be one line`},
		{"    attribution:\n      en: \"Provided\\x7fby Example.\"\n", `the "en" notice must be one line`}, // DEL
		{"    attribution:\n      en: " + long + "\n", "is longer than"},
		{"    attribution: {}\n", `needs an "en" notice`},
		// Unicode line breaks, control and format characters: the notice
		// goes raw into a model's context and a terminal.
		{"    attribution:\n      en: \"a\\Lb\"\n", `the "en" notice must be one line of printable text`},       // U+2028
		{"    attribution:\n      en: \"a\\Pb\"\n", `the "en" notice must be one line of printable text`},       // U+2029
		{"    attribution:\n      en: \"a\\Nb\"\n", `the "en" notice must be one line of printable text`},       // NEL
		{"    attribution:\n      en: \"a\\0b\"\n", `the "en" notice must be one line of printable text`},       // NUL
		{"    attribution:\n      en: \"a\\tb\"\n", `the "en" notice must be one line of printable text`},       // tab
		{"    attribution:\n      en: \"a\\e[31mred\"\n", `the "en" notice must be one line of printable text`}, // ANSI
		{"    attribution:\n      en: \"a\\u202Eb\"\n", `the "en" notice must be one line of printable text`},   // bidi override
		{"    attribution:\n      en: \"a\\u2066b\"\n", `the "en" notice must be one line of printable text`},   // bidi isolate
		{"    attribution:\n      en: \"a\\ue000b\"\n", `the "en" notice must be one line of printable text`},   // private use
		// A language the gateway never serves is dead configuration.
		{"    attribution:\n      en: By Example.\n      ja: By Example.\n", `language "ja" is not one the gateway serves`},
		{"    attribution:\n      en: By Example.\n      zz: By Example.\n", `language "zz" is not one the gateway serves`},
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
	// The bound is on the trimmed notice.
	parse(t, withAttribution("    attribution:\n      en: \"  "+strings.Repeat("x", MaxNotice)+"  \"\n"))
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

// Languages are the languages the gateway serves a notice in: those it
// detects in a question, English first.
func TestTheLanguagesServedStartWithEnglish(t *testing.T) {
	if Languages[0] != FallbackLanguage || len(Languages) != 7 {
		t.Fatalf("%v", Languages)
	}
	for _, l := range Languages {
		if !Served(l) {
			t.Fatalf("%s not served", l)
		}
	}
	for _, l := range []string{"ja", "", "EN", "eng"} {
		if Served(l) {
			t.Fatalf("%q served", l)
		}
	}
}

// LanguageOf reads the language a code or a locale names, as a person sets
// it: any case, quoted or not, with a region, an encoding or a modifier,
// or first in a GNU LANGUAGE-style list ("fr:en").
// "" for a language the gateway does not serve.
func TestALanguageIsReadFromACodeOrALocale(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"fr", "fr"}, {"FR", "fr"}, {"fr_CH.UTF-8", "fr"}, {"de-CH", "de"}, {"pt_BR", "pt"},
		{"en_US.UTF-8@euro", "en"}, {"it@euro", "it"}, {"fr:en", "fr"}, {"de_CH:fr", "de"}, {"nl.UTF-8", "nl"}, {` "es" `, "es"}, {"'de'", "de"},
		{"", ""}, {"C", ""}, {"C.UTF-8", ""}, {"POSIX", ""}, {"ja_JP", ""}, {"fra", ""}, {"f", ""}, {`"`, ""},
	} {
		if got := LanguageOf(c.in); got != c.want {
			t.Fatalf("LanguageOf(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
