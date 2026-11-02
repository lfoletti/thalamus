// Package convention holds the rules a brain's convention puts on the
// documents written to it: the metadata every document declares, the form
// of a new document's id, and the word likeness that tells a near copy.
// The rules are pure; the relay reads the brain and applies them.
package convention

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/lfoletti/thalamus/gateway/rights"
)

// idForm is <nature>:<theme>:<subject>, lower case words joined by hyphens.
var idForm = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*:[a-z0-9]+(-[a-z0-9]+)*:[a-z0-9]+(-[a-z0-9]+)*$`)

// Refusal is why the convention refuses a document: the caller's fault, to
// be fixed in the document, never the brain's.
type Refusal struct{ msg string }

func (r *Refusal) Error() string { return r.msg }

func refuse(format string, args ...any) error { return &Refusal{fmt.Sprintf(format, args...)} }

// Check returns the metadata the document is written with, or why the
// convention refuses it, as a *Refusal. exists says the brain already has a
// document of this id: an old document keeps its id, whatever its form.
// declared are the themes the brain's own theme documents declare, beside
// the convention's. A theme the brain does not know yet does not refuse
// the document: it is filed under the unsorted theme, the theme it names
// kept in theme_proposed. The returned map is a copy, with status current
// when the caller set none. declares_theme is the gateway's key: the
// caller's is dropped, and only a document theme:<subject> gets one.
func Check(c rights.Convention, id string, md map[string]string, exists bool, declared []string) (map[string]string, error) {
	out := make(map[string]string, len(md)+2)
	for k, v := range md {
		out[k] = v
	}
	delete(out, "declares_theme")
	if out["status"] == "" {
		out["status"] = "current"
	} else if s := out["status"]; s != "current" && s != "replaced" {
		return nil, refuse("the brain's convention needs metadata status, current or replaced")
	}
	known := Known(c, declared)
	unsorted := c.UnsortedTheme()
	if subject, ok := strings.CutPrefix(id, ThemePrefix); ok {
		// A theme document is filed under no theme: it declares one. Its
		// nature, when it says one, and its date follow the rules.
		delete(out, "theme")
		delete(out, "theme_proposed")
		if err := natureAndDate(c, out, out["nature"] != ""); err != nil {
			return nil, err
		}
		switch {
		case !rights.Name(subject) && exists:
			// An old theme document keeps its id, and declares nothing.
			return out, nil
		case !rights.Name(subject):
			return nil, refuse("the brain's convention needs a document_id theme:<subject>, in lower case words joined by hyphens, not a number")
		case exists:
			// A theme document rewritten keeps its theme.
			out["declares_theme"] = subject
			return out, nil
		}
		return out, declaresTheme(subject, out, append(known, unsorted))
	}
	theme, nature := out["theme"], out["nature"]
	if theme != unsorted && !slices.Contains(known, theme) {
		if !rights.Name(theme) {
			return nil, refuse("the brain's convention needs metadata theme, one of: %s, or %s with metadata theme_proposed", strings.Join(known, ", "), unsorted)
		}
		out["theme"], out["theme_proposed"], theme = unsorted, theme, unsorted
	}
	if p := out["theme_proposed"]; theme == unsorted && slices.Contains(known, p) {
		// A proposal of a theme the brain knows is filed under it.
		out["theme"], theme = p, p
	}
	idTheme := theme
	if theme != unsorted {
		// A known theme proposes nothing.
		delete(out, "theme_proposed")
	} else if p := out["theme_proposed"]; p != "" {
		if !rights.Name(p) {
			return nil, refuse("the brain's convention needs metadata theme_proposed in lower case words joined by hyphens, not a number")
		}
		idTheme = p
	}
	if err := natureAndDate(c, out, true); err != nil {
		return nil, err
	}
	if !exists {
		if !idForm.MatchString(id) {
			return nil, refuse("the brain's convention needs a document_id <nature>:<theme>:<subject>, in lower case words joined by hyphens")
		}
		if prefix := nature + ":" + idTheme + ":"; !strings.HasPrefix(id, prefix) {
			return nil, refuse("under the brain's convention, the document_id starts with %q", prefix)
		}
	}
	return out, nil
}

// natureAndDate checks the document's nature, from the brain's list, when
// it needs one or says one, and its date, as YYYY-MM-DD, when it says one
// or its nature is dated.
func natureAndDate(c rights.Convention, md map[string]string, needNature bool) error {
	nature := md["nature"]
	if needNature && !slices.Contains(c.Natures, nature) {
		return refuse("the brain's convention needs metadata nature, one of: %s", strings.Join(c.Natures, ", "))
	}
	date, dated := md["date"]
	if dated || slices.Contains(c.Dated, nature) {
		if _, err := time.Parse(time.DateOnly, date); err != nil {
			return refuse("the brain's convention needs metadata date, as YYYY-MM-DD, for a %s", nature)
		}
	}
	return nil
}

// Known are the themes the brain knows: the convention's, then each theme
// a document declares that the convention does not list already, as when
// a grown theme is promoted into the rights file.
func Known(c rights.Convention, declared []string) []string {
	out := slices.Clone(c.Themes)
	for _, d := range declared {
		if !slices.Contains(out, d) {
			out = append(out, d)
		}
	}
	return out
}

// ThemePrefix starts the id of a document that declares a theme.
const ThemePrefix = "theme:"

// declaresTheme checks a new theme document and names its theme in
// declares_theme.
func declaresTheme(subject string, md map[string]string, known []string) error {
	for _, k := range known {
		if k == subject {
			return refuse("the theme %q exists", subject)
		}
		if CloseNames(k, subject) && md["distinct_from"] != k {
			return refuse("the theme %q is close to the theme %q: use that theme, or name it in metadata distinct_from", subject, k)
		}
	}
	md["declares_theme"] = subject
	return nil
}

// CloseNames reports whether two theme names are likely the same theme:
// the same once the hyphens and a final s are dropped, or, for names of
// four letters or more, one edit apart.
func CloseNames(a, b string) bool {
	stem := func(s string) string { return strings.TrimSuffix(strings.ReplaceAll(s, "-", ""), "s") }
	if stem(a) == stem(b) {
		return true
	}
	return len(a) >= 4 && len(b) >= 4 && edits(a, b) <= 1
}

// edits is the Levenshtein distance of two words.
func edits(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}

// Likeness is the Jaccard index of the two texts' word trigrams, blind to
// case and punctuation. A text of fewer than three words is compared by its
// words; an empty text is like nothing.
func Likeness(a, b string) float64 {
	wa, wb := words(a), words(b)
	n := 3
	if len(wa) < n || len(wb) < n {
		n = 1
	}
	sa, sb := grams(wa, n), grams(wb, n)
	if len(sa) == 0 || len(sb) == 0 {
		return 0
	}
	common := 0
	for g := range sa {
		if sb[g] {
			common++
		}
	}
	return float64(common) / float64(len(sa)+len(sb)-common)
}

func words(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
}

func grams(w []string, n int) map[string]bool {
	out := map[string]bool{}
	for i := 0; i+n <= len(w); i++ {
		out[strings.Join(w[i:i+n], " ")] = true
	}
	return out
}
