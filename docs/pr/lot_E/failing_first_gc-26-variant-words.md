# Failing first, gc-26-variant-words (head 305be9b, base a591ae1)

## Tests on the base

`env -u CORTEXDB_TEST_POSTGRES go test ./pkg/cortexdb/ -run 'ThroughAlternateSentence|TakeAlternateWords' -count=1`:

```
--- FAIL: TestLexicalSearchQueriesTakeAlternateWords   OR query "qui OR règle OR la OR procédure OR des OR cantons" lacks "verfahren"
--- FAIL: TestKnowledgeSearchFindsDocumentThroughAlternateSentence   results [], want the German answer first
--- FAIL: TestSearchTextFindsDocumentThroughAlternateSentence   results [fr-filler-00 … fr-filler-04], want the German answer among them
--- FAIL: TestMemorySearchFindsMemoryThroughAlternateSentence   results [], want the German answer first
```

## Hand mutations (`pkg/cortexdb/graphrag_tool_helpers.go`)

| Mutation | Tests that fail |
| :- | :- |
| variant words empty: `alternateKeywords := lexicalQueryKeywords("")` | all four |
| question words dropped from the merge: `mergeLexicalKeywords(plannedKeywords, alternateKeywords)` | TestLexicalSearchQueriesTakeAlternateWords |

Replay: `make handmut REPO=<wt>/gc-26-variant-words FILE=pkg/cortexdb/graphrag_tool_helpers.go OLD='alternateKeywords := lexicalQueryKeywords(strings.Join(alternateQueries, " "))' NEW='alternateKeywords := lexicalQueryKeywords("")' GOTEST="-count=1 -run 'ThroughAlternateSentence|TakeAlternateWords' ./pkg/cortexdb/"`
