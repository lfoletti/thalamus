# Failing first, gc-25-coverage-accents (head ef9e217, base a591ae1)

## Tests on the base (code absent)

`env -u CORTEXDB_TEST_POSTGRES go test ./pkg/cortexdb/ -run 'TestKnowledgeKeywordFloor|TestMemoryKeywordFloor' -count=1`, with `lexical_accents_test.go` only (the two unit tests in `lexical_fold_test.go` call `foldAccents` and do not compile on the base):

```
--- FAIL: TestKnowledgeKeywordFloorComparesWordsWithoutAccents     results [], want the Universität document first
--- FAIL: TestKnowledgeKeywordFloorMatchesAccentedQueryOnUnaccentedText  results [], want the Conseil d'Etat document first
--- FAIL: TestKnowledgeKeywordFloorComparesKeywordsWithoutAccents  results [], want the Universität document first
--- FAIL: TestMemoryKeywordFloorComparesWordsWithoutAccents        results [], want the Universität memory first
```

`TestKnowledgeKeywordFloorStillRefusesOneSharedWordWithoutAccents` passes on the base, as it should (it guards what the floor refused).

## Hand mutations (all in `pkg/cortexdb/lexical_rarity.go`, `tools/handmut.py`)

GOTEST: `-count=1 -run 'TestKnowledgeKeywordFloor|TestMemoryKeywordFloor|TestRarity|TestFoldAccents|TestContainsWord|TestKnowledgeBelowTheFloor|TestAMemory|TestAKeyword|TestLexicalMode' ./pkg/cortexdb/`

| # | Mutation | Tests that fail |
| :- | :- | :- |
| M1 | row not folded: `folded := lower` | ComparesWordsWithoutAccents, ComparesKeywordsWithoutAccents, Memory…WithoutAccents, TestRarityCoverageComparesWordsWithoutAccents |
| M2 | question words not folded: `append(s.folded, w)` | MatchesAccentedQueryOnUnaccentedText, ComparesKeywordsWithoutAccents, TestRarityCoverage… |
| M3 | keyword compared as before: `containsWord(lower, e)` | ComparesKeywordsWithoutAccents, TestRarityCoverage… |
| M3b | keyword not folded: `containsWord(folded, e)` | TestRarityCoverage… |
| M4 | question words kept with accents for synonym dedup: `have[w]` | TestRarityCoverage… |
| M4b | synonym kept with accents for dedup: `key := term` | TestRarityCoverage… |
| M5 | combining mark kept: `return r` | TestFoldAccentsFoldsWhatTheWordIndexFolds |
| M6 | no fold at all: `if false {` (non-ASCII test) | all six accent tests |
| M7 | `ü` (U+00FC) left as is in the table | TestFoldAccentsFoldsWhatTheWordIndexFolds |
| M8 | `ſ` (U+017F) left as is in the table | TestFoldAccentsFoldsWhatTheWordIndexFolds |
| M9 | the floor takes one shared word: `return m.words >= 1 \|\| s.words <= 1` | TestKnowledgeKeywordFloorStillRefusesOneSharedWordWithoutAccents |

11 of 11 killed. Each mutant compiled (FAIL lines are test failures, not build failures).

## The table against the index

The table was checked against FTS5 itself (modernc.org/sqlite v1.38.2): each rune from U+00C0 to U+017F inserted as `x<r>x` and read back through `fts5vocab`. unicode61 folds exactly the letters that decompose to an ASCII letter plus marks, and also `ſ` to `s`; it keeps `æ ð ø þ ß đ ħ ı ĳ ĸ ŀ ł ŉ ŋ œ ŧ`. Combining marks are dropped (`café` → `cafe`). The table matches.
