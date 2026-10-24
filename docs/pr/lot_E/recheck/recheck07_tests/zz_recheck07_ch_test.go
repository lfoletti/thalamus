package cortexdb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"unicode"
)

func stripAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, _ := transform.String(t, s)
	return out
}

func TestRecheck07SwissMonolingual(t *testing.T) {
	dir := "/Users/mini/Documents/granular_cortex/tests/eval_ch"
	var arts map[string]map[string]struct{ Title, Text string }
	var qs []struct {
		ID, Article, Lang, Q string
		V                    map[string]string
	}
	b, _ := os.ReadFile(filepath.Join(dir, "articles.json"))
	if err := json.Unmarshal(b, &arts); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(filepath.Join(dir, "questions.json"))
	if err := json.Unmarshal(b, &qs); err != nil {
		t.Fatal(err)
	}
	for _, floors := range []bool{true, false} {
		var opts []Option
		if !floors {
			opts = append(opts, WithKnowledgeFloors(0, 0))
		}
		db, err := Open(DefaultConfig(filepath.Join(t.TempDir(), fmt.Sprintf("ch%v.db", floors))), opts...)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		for id, langs := range arts {
			for l, a := range langs {
				if strings.TrimSpace(a.Text) == "" {
					continue
				}
				recheckSave(t, db, id+"__"+l, a.Title+"\n"+a.Text)
			}
		}
		// each question and each of its variants, asked in its own language,
		// against the article in that same language.
		type probe struct{ lang, q, want string }
		var ps []probe
		for _, q := range qs {
			ps = append(ps, probe{q.Lang, q.Q, q.Article + "__" + q.Lang})
			for l, v := range q.V {
				ps = append(ps, probe{l, v, q.Article + "__" + l})
			}
		}
		for _, accent := range []bool{true, false} {
			hit5, hit10, n := 0, 0, 0
			for _, p := range ps {
				query := p.q
				if !accent {
					query = stripAccents(query)
					if query == p.q {
						continue
					}
				}
				n++
				ids := recheckIDs(t, db, KnowledgeSearchRequest{Query: query, TopK: 10})
				for i, id := range ids {
					if id == p.want {
						if i < 5 {
							hit5++
						}
						hit10++
					}
				}
			}
			t.Logf("floors=%v accents=%v n=%d @5=%.2f @10=%.2f", floors, accent, n, float64(hit5)/float64(n), float64(hit10)/float64(n))
		}
	}
}
