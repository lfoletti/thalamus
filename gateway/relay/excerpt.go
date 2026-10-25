package relay

import (
	"unicode"

	"github.com/lfoletti/thalamus/gateway/rights"
)

// A brain served in excerpts (excerpt_words in the rights file) hands out
// the first words of a source document, never the whole: the reader follows
// the address of the original. The syntheses of its pyramid are the
// publisher's own text and stay whole. The per-brain endpoint, whose raw
// tools would serve a document whole, is refused to such a brain (admit).

// excerpt says what was cut: the words served, the words of the whole, and
// the address of the original.
type excerpt struct {
	Words    int    `json:"words"`
	Of       int    `json:"of"`
	Original string `json:"original,omitempty"`
}

// cut keeps the text up to the end of its n-th word, with its spacing, and
// counts the words of the whole. n <= 0 keeps the text whole.
func cut(text string, n int) (string, int, bool) {
	words, end, in := 0, len(text), false
	for i, r := range text {
		space := unicode.IsSpace(r)
		if !space && !in {
			words++
		}
		if space && in && words == n {
			end = i
		}
		in = !space
	}
	if n <= 0 || words <= n {
		return text, words, false
	}
	return text[:end], words, true
}

// bound is the one place a brain's excerpt rule applies: the content of a
// document of that collection as served, and the note of the cut, nil when
// it is whole. th_descend opens sources only, and names no collection.
func bound(b rights.Brain, collection, content, url string) (string, *excerpt) {
	if b.Pyramid != nil && collection == b.Pyramid.Collection {
		return content, nil
	}
	text, words, cutOff := cut(content, b.ExcerptWords)
	if !cutOff {
		return content, nil
	}
	return text, &excerpt{Words: b.ExcerptWords, Of: words, Original: url}
}
