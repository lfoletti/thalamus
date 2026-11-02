package relay

import (
	"encoding/json"
	"slices"
	"strings"
	"unicode"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// A brain's attribution (rights.Attribution) ends every context served
// from it: th_recall names the notice of each brain that contributed a
// result, in the language of the question; th_get, th_descend, th_entity,
// the graph tools and the per-brain endpoint carry the notice of the brain
// they read, in the fallback language, since they have no question to read
// it from. The curator's tools (th_add, th_pending, th_review, th_themes,
// th_retheme) serve the brain's administration, not its information, and
// carry none.

// notice is one brain's attribution, as an answer carries it.
type notice struct {
	Brain string `json:"brain"`
	Text  string `json:"text"`
}

// functionWords are the words that tell a question's language: articles,
// pronouns, prepositions, conjunctions, question words and the commonest
// verbs. A word shared by two languages counts for both.
var functionWords = map[string][]string{
	"en": {"the", "of", "what", "which", "who", "when", "where", "why", "how", "does", "do", "is", "are", "and", "to", "in", "for", "with", "on", "this", "that", "it", "a", "an"},
	"fr": {"le", "la", "les", "des", "du", "de", "un", "une", "et", "est", "que", "qu", "qui", "quoi", "quel", "quelle", "quels", "quelles", "pour", "dans", "sur", "avec", "où", "en", "c", "l", "d", "sont", "ce", "ces", "au", "aux", "comment", "quand", "pourquoi"},
	"de": {"was", "die", "der", "das", "den", "dem", "und", "ist", "wie", "von", "mit", "für", "auf", "ein", "eine", "nicht", "zu", "im", "sind", "werden", "wer", "warum"},
	"es": {"el", "la", "los", "las", "de", "del", "y", "que", "qué", "cuál", "cuáles", "son", "cómo", "se", "en", "por", "para", "con", "una", "un", "es", "lo"},
	"it": {"il", "lo", "gli", "dei", "delle", "della", "di", "e", "è", "che", "quali", "quale", "sono", "come", "si", "nella", "nel", "per", "con", "una", "un"},
	"pt": {"o", "os", "as", "da", "do", "dos", "das", "de", "e", "que", "quais", "qual", "são", "como", "se", "em", "para", "com", "uma", "um", "elas", "eles", "não"},
	"nl": {"de", "het", "een", "en", "van", "zijn", "wat", "hoe", "worden", "ze", "dat", "die", "voor", "met", "op", "niet", "te"},
}

// languages is the order in which a tie is looked for: fixed, so the
// answer does not depend on the map.
var languages = []string{"en", "fr", "de", "es", "it", "pt", "nl"}

var wordsOf = func() map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for lang, words := range functionWords {
		out[lang] = map[string]bool{}
		for _, w := range words {
			out[lang][w] = true
		}
	}
	return out
}()

// language reads the language of a question from its function words: the
// language with the most of them, or "" when none has more than every
// other, and the brain's fallback notice is served.
func language(text string) string {
	counts := map[string]int{}
	for _, w := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) }) {
		for _, lang := range languages {
			if wordsOf[lang][w] {
				counts[lang]++
			}
		}
	}
	best, top, tie := "", 0, false
	for _, lang := range languages {
		switch n := counts[lang]; {
		case n > top:
			best, top, tie = lang, n, false
		case n == top && n > 0:
			tie = true
		}
	}
	if tie {
		return ""
	}
	return best
}

// noticesFor is the notice of each brain that has one, in the language,
// in the order given.
func (s *Server) noticesFor(brains []string, lang string) []notice {
	var out []notice
	for _, name := range brains {
		if text := s.rights().Brains[name].Attribution.Notice(lang); text != "" {
			out = append(out, notice{Brain: name, Text: text})
		}
	}
	return out
}

// contributors are the brains a result came from, or also came from, once
// and sorted.
func contributors(results []fused) []string {
	var out []string
	for _, r := range results {
		out = append(append(out, r.Brain), r.AlsoIn...)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// attributed adds the notices to an answer: under "attribution" when the
// answer is a JSON object without that key, the rest of its text as it
// was; as a second text, one notice a line, otherwise. An error, or no
// notice, is left as it is.
func attributed(res *mcp.CallToolResult, notices []notice) *mcp.CallToolResult {
	if res.IsError || len(notices) == 0 {
		return res
	}
	text := res.Content[0].(*mcp.TextContent).Text
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &fields); err == nil && fields != nil {
		if _, taken := fields["attribution"]; !taken {
			data, _ := json.Marshal(notices)
			end := strings.LastIndex(text, "}")
			sep := ","
			if len(fields) == 0 {
				sep = ""
			}
			res.Content[0] = &mcp.TextContent{Text: text[:end] + sep + `"attribution":` + string(data) + text[end:]}
			return res
		}
	}
	lines := make([]string, len(notices))
	for i, n := range notices {
		lines[i] = n.Text
	}
	res.Content = append(res.Content, &mcp.TextContent{Text: strings.Join(lines, "\n")})
	return res
}
