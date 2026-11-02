package relay

import (
	"encoding/json"
	"slices"
	"strings"
	"unicode"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/text/unicode/norm"

	"github.com/lfoletti/thalamus/gateway/rights"
)

// A brain's attribution (rights.Attribution) ends every context served
// from it: th_recall names the notice of each brain that contributed a
// result, in the language of the question; th_get, th_descend, th_entity,
// the graph tools and the per-brain endpoint carry the notice of the brain
// they read, in the fallback language, since they have no question to read
// it from; the per-brain endpoint credits every answer but those of its
// pure writes (pureWrites).
// th_add, th_pending, th_review and th_retheme are a writer's or a
// curator's tools, and th_themes serves counts and theme names, not
// documents: they carry none.

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
	"de": {"was", "die", "der", "das", "den", "dem", "und", "ist", "wie", "von", "mit", "für", "auf", "ein", "eine", "nicht", "zu", "im", "sind", "werden", "wer", "warum", "wo", "ich", "wir"},
	"es": {"el", "la", "los", "las", "de", "del", "y", "que", "qué", "cuál", "cuáles", "son", "cómo", "se", "en", "por", "para", "con", "una", "un", "es", "lo", "dónde", "está", "cuándo"},
	"it": {"il", "lo", "la", "gli", "dei", "delle", "della", "di", "e", "è", "che", "quali", "quale", "sono", "come", "si", "nella", "nel", "per", "con", "una", "un"},
	"pt": {"o", "os", "as", "da", "do", "dos", "das", "de", "e", "que", "quais", "qual", "são", "como", "se", "em", "para", "com", "uma", "um", "elas", "eles", "não"},
	"nl": {"de", "het", "een", "en", "van", "zijn", "wat", "hoe", "worden", "ze", "dat", "die", "voor", "met", "op", "niet", "te"},
}

// accents are the accented letters and the marks (Spanish ¿ and ¡) that
// mark a language, to break a tie between languages other than English.
var accents = map[string]string{
	"fr": "éèêëàâçîïôûùœ",
	"de": "äöüß",
	"es": "áéíóúñü¿¡",
	"it": "àèéìòù",
	"pt": "ãõáâêôçéíóúà",
}

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
// language with the most of them. "" (the fallback, English) when no word
// is found, or when English shares the lead. A tie between other
// languages goes to the one with the most of its accented letters, then
// to the first in rights.Languages: English never wins on zero words.
func language(text string) string {
	// Composed (NFC), a decomposed "où" is one word and its accent counts.
	lower := strings.ToLower(norm.NFC.String(text))
	counts := map[string]int{}
	for _, w := range strings.FieldsFunc(lower, func(r rune) bool { return !unicode.IsLetter(r) }) {
		for _, lang := range rights.Languages {
			if wordsOf[lang][w] {
				counts[lang]++
			}
		}
	}
	top := 0
	for _, n := range counts {
		top = max(top, n)
	}
	var lead []string
	for _, lang := range rights.Languages {
		if counts[lang] == top {
			lead = append(lead, lang)
		}
	}
	// With no word at all, every language leads, English first.
	switch {
	case len(lead) > 1 && lead[0] == rights.FallbackLanguage:
		return ""
	case len(lead) == 1:
		return lead[0]
	}
	best, marks := lead[0], -1
	for _, lang := range lead {
		n := 0
		for _, r := range lower {
			if strings.ContainsRune(accents[lang], r) {
				n++
			}
		}
		if n > marks {
			best, marks = lang, n
		}
	}
	return best
}

// pureWrites are the engine's writes and deletes that answer with ids and
// counts only, never the brain's content: the per-brain endpoint credits
// every other answer. A tool the engine classes as a write but which
// serves content (knowledge_graph_query runs SPARQL SELECT,
// knowledge_memory_reflect recalls), a write that returns its record
// (knowledge_save), and a tool of a newer engine are credited: a notice too
// many costs a line, a notice missing breaks the publisher's terms.
var pureWrites = []string{
	"ingest_document", "upsert_entities", "upsert_relations", "delete_entities", "delete_document_graph",
	"knowledge_delete", "memory_delete", "knowledge_graph_upsert", "knowledge_graph_delete",
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
	// An answer without a first text has "" here, which is not an object.
	text := ""
	if len(res.Content) > 0 {
		if tc, ok := res.Content[0].(*mcp.TextContent); ok {
			text = tc.Text
		}
	}
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
