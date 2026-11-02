package rights

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Attribution is the notice that ends every context served from a brain:
// who provides the information, one line per language, keyed by a
// two-letter language code. The English line is required: it is served
// when the question's language has no line of its own. A brain's publisher
// sets it when its terms ask the readers to be told where the information
// comes from.
type Attribution map[string]string

// FallbackLanguage is the language served when the question's language
// has no notice.
const FallbackLanguage = "en"

// MaxNotice bounds a notice, in characters: it is one line at the end of a
// context, not a licence.
const MaxNotice = 200

// Languages are the languages the gateway serves a notice in: those it
// reads in a question, English, the fallback, first. A notice in another
// language could never be chosen, so rights check refuses it.
var Languages = []string{FallbackLanguage, "fr", "de", "es", "it", "pt", "nl"}

// Served reports whether the gateway serves notices in the language.
func Served(lang string) bool { return slices.Contains(Languages, lang) }

var languageCode = regexp.MustCompile(`^[a-z]{2}$`)

// printable reports whether a notice is one line a reader sees as it is:
// no control character (tabs, escapes, NEL included), no format character
// (bidirectional overrides and isolates), no Unicode line or paragraph
// separator.
func printable(text string) bool {
	return !strings.ContainsFunc(text, func(r rune) bool {
		return unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp)
	})
}

// Notice is the notice in the language, or in the fallback language when
// the language has none; "" for a brain without an attribution.
func (a Attribution) Notice(lang string) string {
	if text, ok := a[lang]; ok {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(a[FallbackLanguage])
}

func (a Attribution) check() error {
	langs := make([]string, 0, len(a))
	for lang := range a {
		langs = append(langs, lang)
	}
	slices.Sort(langs)
	for _, lang := range langs {
		text := strings.TrimSpace(a[lang])
		switch {
		case !languageCode.MatchString(lang):
			return fmt.Errorf("attribution: language %q must be a two-letter lower case code", lang)
		case !Served(lang):
			return fmt.Errorf("attribution: language %q is not one the gateway serves (%s)", lang, strings.Join(Languages, ", "))
		case text == "":
			return fmt.Errorf("attribution: the %q notice is empty", lang)
		case !printable(text):
			return fmt.Errorf("attribution: the %q notice must be one line of printable text", lang)
		case utf8.RuneCountInString(text) > MaxNotice:
			return fmt.Errorf("attribution: the %q notice is longer than %d characters", lang, MaxNotice)
		}
	}
	if _, ok := a[FallbackLanguage]; !ok {
		return fmt.Errorf("attribution: needs an %q notice, served when the question's language has none", FallbackLanguage)
	}
	return nil
}
