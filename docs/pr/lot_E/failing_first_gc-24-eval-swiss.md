# Failing first, gc-24-eval-swiss (head bfe2037, base a591ae1)

No rule of code: a dataset and a test that holds its numbers as floors. The check is that the floors bite.

Generator of the dataset: `docs/pr/lot_E/gc24/make_dataset_ch.py` (input `tests/eval_ch/articles.json` and `questions.json`).

| Mutation | Test | Result |
| :- | :- | :- |
| `containsWord` always false (`if word == "" || true {`, `pkg/cortexdb/lexical_rarity.go`) | `TestSwissLexicalRetrievalQuality` | `--- FAIL: TestSwissLexicalRetrievalQuality (0.65s)`, killed |

Replay:

```
make handmut REPO=<wt>/gc-24-eval-swiss FILE=pkg/cortexdb/lexical_rarity.go OLD='	if word == "" {' NEW='	if word == "" || true {' GOTEST='-count=1 -run TestSwissLexicalRetrievalQuality ./pkg/eval/'
```

Stability: `go test ./pkg/eval/ -run TestSwissLexical -count=3` green.
