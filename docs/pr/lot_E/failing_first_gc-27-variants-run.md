# Failing first, gc-27-variants-run (head e53d564, base a591ae1)

## Tests on the base

`env -u CORTEXDB_TEST_POSTGRES go test ./pkg/cortexdb/ -run 'RunsAlternatesWhenQuestionFillsTopK' -count=1`:

```
--- FAIL: TestSearchTextRunsAlternatesWhenQuestionFillsTopK   results [fr0 fr1 fr2 fr3 fr4], want five with the row only the variant finds
--- FAIL: TestSearchTextOnlyRunsAlternatesWhenQuestionFillsTopK   results [fr0 fr1 fr2], want three with the row only the variant finds
```

## Hand mutations (each `break` put back)

| Mutation | Test that fails |
| :- | :- |
| `if idx == 0 && len(merged) >= searchOpts.TopK { break }` back in `graphrag_tool_search.go` | TestSearchTextRunsAlternatesWhenQuestionFillsTopK |
| `if idx == 0 && len(merged) >= opts.TopK { break }` back in `text_search.go` (`searchTextOnlyExpanded`) | TestSearchTextOnlyRunsAlternatesWhenQuestionFillsTopK |

2 of 2 killed, each by its own test.
