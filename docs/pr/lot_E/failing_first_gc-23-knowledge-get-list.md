# Failing first: gc-23-knowledge-get-list

Head `5b65970` on `origin/main` `a591ae1` (23 October 2026). Worktree:
`/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-23-knowledge-get-list`.

## Tests on the base (tests posed, code absent)

Every test goes through the tool surface with JSON arguments, so all compile on `a591ae1`.

```sh
W=<worktree>; B=<scratch dir>
git -C $W worktree add --detach $B a591ae1
git -C $W diff a591ae1 HEAD -- '*_test.go' | git -C $B apply
cd $B
go test ./pkg/cortexdb/ -run 'TestKnowledgeGetList'
go test ./pkg/authz/ -run 'TestKnowledgeGetTakesOneForm|TestEveryToolAndEveryArgument'
go test ./pkg/rpcserver/ -run TestAConfinedKeyListsOnlyItsCollection
go test ./pkg/httpapi/ -run TestAConfinedKeyListsOnlyItsCollection
```

Observed on the base:

```
--- FAIL: TestKnowledgeGetListsSeveralInTheOrderAsked
    knowledge_id is required
--- FAIL: TestKnowledgeGetListHidesAnotherCollectionAsMissing
    [ops-1 absent notes-1]: knowledge_id is required
--- FAIL: TestKnowledgeGetListServesWhatTheSingleFormServes
    served [], missing [] (knowledge_id is required), want both, as the single form serves them
--- FAIL: TestKnowledgeGetTakesOneFormFromAConfinedKey
    {"knowledge_ids":["n1","n2"]}: one form was refused: denied: key "notes" is confined and knowledge_get takes no knowledge_ids argument a confined key may send
--- FAIL: TestAConfinedKeyListsOnlyItsCollectionWithKnowledgeGet
    notes lists [hr-1 absent n1]: ... PermissionDenied desc = denied: key "notes" is confined and knowledge_get takes no knowledge_ids argument ...
    the operator lists [hr-1 n1]: ... InvalidArgument desc = knowledge_id is required
--- FAIL: TestAConfinedKeyListsOnlyItsCollectionWithKnowledgeGetOverREST
    notes lists: 403 {"error":"denied: key \"notes\" is confined and knowledge_get takes no knowledge_ids argument a confined key may send"}, want 200
```

The existing `TestTheAllowlistTableNamesEveryToolAndEveryIDField` has its `knowledge_get` row updated (the refused pair is now listed).

## Hand mutations (one per rule)

Specs in `docs/pr/lot_E/mut_02_03_05/m23/`. Replay:

```sh
for f in docs/pr/lot_E/mut_02_03_05/m23/K*.json; do
  python3 -I docs/pr/lot_E/mut_02_03_05/mutmulti.py --repo $W --spec $f -- -count=1 -run 'KnowledgeGet|ListsOnlyItsCollection|TestEveryToolAndEveryArgument|Table' ./pkg/cortexdb/ ./pkg/authz/ ./pkg/rpcserver/ ./pkg/httpapi/
done
```

| # | Mutation | Tests that fall |
| :- | :- | :- |
| K01 | `knowledgeListed` always true | `TestKnowledgeGetListHidesAnotherCollectionAsMissing`, `TestAConfinedKeyListsOnlyItsCollectionWithKnowledgeGet`, `...OverREST` |
| K02 | a document without chunks listed under a scope | `TestKnowledgeGetListHidesAnotherCollectionAsMissing` |
| K03 | an absent id dropped in silence | `TestKnowledgeGetListsSeveralInTheOrderAsked`, `TestKnowledgeGetListHidesAnotherCollectionAsMissing`, both port tests |
| K04 | order lost (ids sorted; two sites) | `TestKnowledgeGetListsSeveralInTheOrderAsked`, `TestKnowledgeGetListHidesAnotherCollectionAsMissing`, `TestKnowledgeGetListServesWhatTheSingleFormServes`, `TestAConfinedKeyListsOnlyItsCollectionWithKnowledgeGet` |
| K05 | a read error taken as missing | `TestKnowledgeGetListHidesAnotherCollectionAsMissing` |
| K06 | `oneOf` taken out of the allowlist line | `TestKnowledgeGetTakesOneFormFromAConfinedKey`, `TestTheAllowlistTableNamesEveryToolAndEveryIDField`, both port tests |
| K07 | the single form wins when both are set | `TestKnowledgeGetListsSeveralInTheOrderAsked` |
| K08 | `knowledge_ids` not classified | `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys`, `TestKnowledgeGetTakesOneFormFromAConfinedKey`, both port tests |
| K09 | an empty id not taken as missing | `TestKnowledgeGetListsSeveralInTheOrderAsked` |
| K10 | `knowledge` not set to the first one found | `TestKnowledgeGetListsSeveralInTheOrderAsked` |
| K11 | the collection not checked | `TestKnowledgeGetListHidesAnotherCollectionAsMissing`, both port tests |

11 mutations, 11 killed. Output: `m23/results.txt`.

K02 survived a first time: the facade places a document without chunks in the default collection, which the test's scope did not reach, so the collection check alone hid it. The test's scope now reaches the default collection too, and K02 falls.

A first run of these mutations hit a full disk ("no space left on device"): the build failures were counted as kills. The tool now reports `BUILD-FAILED` apart, and the whole set was run again after space was freed; the table is that second run.

## Class inventory

- Read paths of `knowledge_get`: the list form (new), the single form (unchanged, unscoped, used by the row guard: `TestKnowledgeGetListHidesAnotherCollectionAsMissing`, last part), typed `GetKnowledge` and `GET /v1/knowledge` (unchanged, no list).
- What the single form serves by id and the list serves the same: expired, proposed under the gate (`TestKnowledgeGetListServesWhatTheSingleFormServes`).
- What a scoped list hides: another collection, no chunks (K01, K02, K11); an absent and an empty id are missing for everyone (K03, K09).
