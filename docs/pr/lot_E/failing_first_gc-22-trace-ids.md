# Failing first: gc-22-trace-ids

Head `a2dbb8a` on `origin/main` `a591ae1` (23 October 2026). Worktree:
`/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-22-trace-ids`.

## Tests on the base (tests posed, code absent)

```sh
W=<worktree>; B=<scratch dir>
git -C $W worktree add --detach $B a591ae1
git -C $W diff a591ae1 HEAD -- '*_test.go' | git -C $B apply
cd $B
go test ./pkg/rpcserver/ -run 'TestTheRecordNamesTheRowsTheCallNames|TestAnotherMethodLogsNoIDs'
go test ./pkg/authz/ -run 'RowIDs|NamesARow'
```

The `pkg/rpcserver` test decodes the trace file as a plain JSON object, so it compiles on the base and fails there:

```
--- FAIL: TestTheRecordNamesTheRowsTheCallNames
    knowledge_get {"knowledge_id":"k1"}: ids <nil>, want map[knowledge_id:[k1]]
    knowledge_get {"knowledge_id":"k1"}: the line "cortexdb-trace rpc=CallTool dur=0.00s key=operator tool=knowledge_get" lacks " ids=1" after the tool
    memory_get {"memory_ids":["m2","m1"]}: ids <nil>, want map[memory_ids:[m1 m2]]
    knowledge_search {...plan.filters.document_ids...}: ids <nil>, want map[plan.filters.document_ids:[d1]]
    ingest_document {"document_id":"d9","content":"x"}: ids <nil>, want map[document_id:[d9]]
```

The three `pkg/authz` tests call the new `authz.RowIDsOf` and do not compile on the base (`undefined: RowIDsOf`); the mutations below disable each rule instead. `TestAnotherMethodLogsNoIDs` passes on the base: it holds what must not change.

## Hand mutations (one per rule)

Specs in `docs/pr/lot_E/mut_02_03_05/m22/`. Replay:

```sh
for f in docs/pr/lot_E/mut_02_03_05/m22/N*.json; do
  python3 -I docs/pr/lot_E/mut_02_03_05/mutmulti.py --repo $W --spec $f -- -count=1 -run 'RowIDs|NamesARow|TheRecordNames|AnotherMethodLogsNoIDs' ./pkg/authz/ ./pkg/rpcserver/
done
```

| # | Mutation | Tests that fall |
| :- | :- | :- |
| N1 | `rowData` not read (`rowData[:0]`) | `TestRowIDsOfReadsEveryDeclaredPath`, `TestEveryArgumentThatNamesARowIsLogged`, `TestTheRecordNamesTheRowsTheCallNames` |
| N2 | `memory_ids` taken out of `rowData` | the same three |
| N3 | `plan.filters.document_ids` taken out of `rowData` | the same three |
| N4 | the record's `IDs` never filled | `TestTheRecordNamesTheRowsTheCallNames` |
| N5 | a tool off the allowlist read by a `document_id` guess | `TestRowIDsOfTakesNothingOffTheAllowlist`, `TestTheRecordNamesTheRowsTheCallNames` |
| N6 | the guarded `ids` not read (`r.ids[:0]`) | `TestRowIDsOfReadsEveryDeclaredPath`, `TestRowIDsOfTakesNothingOffTheAllowlist`, `TestEveryArgumentThatNamesARowIsLogged`, `TestTheRecordNamesTheRowsTheCallNames` |
| N7 | no ` ids=<n>` on the text line | `TestTheRecordNamesTheRowsTheCallNames` |
| N8 | the count counts paths, not ids (`len(v[:1])`) | `TestTheRecordNamesTheRowsTheCallNames` |
| N9 | a path with no id kept as an empty entry | `TestRowIDsOfTakesNothingOffTheAllowlist`, `TestEveryArgumentThatNamesARowIsLogged`, `TestTheRecordNamesTheRowsTheCallNames` |

9 mutations, 9 killed. Output: `m22/results.txt` (its N8 entry is a first spelling that did not compile; the valid run is `m22/N8.txt`; its N10 entry is below).

A tenth mutation survived and showed dead code: `RowIDsOf` returned nil when the arguments were not a JSON object, but `argsObject` gives a nil map then, which names no row either. The check was removed (pipeline rule: a guard no input reaches is removed, not declared); `TestRowIDsOfTakesNothingOffTheAllowlist` keeps the `not json` case.

## Class inventory

Every tool of the allowlist is in the table `loggedPaths` of `pkg/authz/row_ids_test.go`, and `TestRowIDsOfReadsEveryDeclaredPath` fails for a tool added to the allowlist without a row there. Every argument whose name ends in `_id` or `_ids`, or is `supersedes`, is logged or excused with a reason (`session_id`, `plan.filters.session_id`: a bucket; `plan.filters.user_id`: a scope field), in `TestEveryArgumentThatNamesARowIsLogged`.
