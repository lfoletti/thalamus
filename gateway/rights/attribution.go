package rights

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
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

var languageCode = regexp.MustCompile(`^[a-z]{2}$`)

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
		case text == "":
			return fmt.Errorf("attribution: the %q notice is empty", lang)
		case strings.ContainsAny(text, "\r\n"):
			return fmt.Errorf("attribution: the %q notice must be one line", lang)
		case utf8.RuneCountInString(text) > MaxNotice:
			return fmt.Errorf("attribution: the %q notice is longer than %d characters", lang, MaxNotice)
		}
	}
	if _, ok := a[FallbackLanguage]; !ok {
		return fmt.Errorf("attribution: needs an %q notice, served when the question's language has none", FallbackLanguage)
	}
	return nil
}
