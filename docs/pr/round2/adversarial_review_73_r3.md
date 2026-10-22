# Relecture adverse de #73, tour 3 (21 octobre 2026, 23 h 43)

Tête `4f71850` (`gc-04b-grpc-confined-tools`), base `d805527`. Agent `upstream-reviewer`, kit `adv-73-r3`, invite fixe de `docs/pipeline_pr.md`, section 1 bis, avec les rapports des tours 1 et 2 en matériel. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 55, contexte 19259 -> 156749 tokens ; écriture cache 156747, lecture cache 4877483, entrée neuve 110 ; équivalent entrée (lecture x0,1, écriture x1,25) : 683682 ; outils : {'Bash': 55, 'SubagentHandback': 1}. Durée 13 min 10 s.

Note de l'auteur : le point L2 vient de la session, pas de la PR. Pendant la relecture, B1 a été sorti de #73 vers `gc-12-grpc-ingest-guard` (décision de Lauro), et `failing_first.md`, matériel de la relecture, a été modifié en cours de route. Leçon : ne toucher à aucun fichier du matériel pendant une relecture.

---

## Review of #73 (gc-04b-grpc-confined-tools, head 4f71850, on #72 at d805527), round 3

**Verdict: no blocker of this PR's own. Four likely comments and six nits.** The round-2 blocker B1 (a typed `InsertGraphDocument` over another collection's document) is closed on this head, and so are L1 (chunkless typed upserts) and L3 (the bucket wording). I checked each one myself (scenarios S1, S2 and S6 below). The most serious open point is L1: a hole for blank ids, which #82 already fixes and which this PR does not mention.

**What I ran**
- `both.sh -race ./pkg/rpcserver/ ./pkg/authz/`: green on both trees.
- I copied the head's new and changed test files (all but `tools_guard_test.go`, which does not compile without the code) onto a private copy of the base. 12 tests fail there. `TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks` passes, by design, and the body says so.
- No new test file sleeps or reads the environment (grep for `Sleep`, `Getenv`, `Setenv`, `LookupEnv`, `UserHomeDir`, `time.After`, `os.Environ`).
- Outside the tests, the diff adds 190 non-blank lines and removes 12.
- The kit's head matches commit 4f71850 file for file. The worktree at `.../wt/gc-04b-grpc-confined-tools` has uncommitted changes that remove the `InsertGraphDocument` guard (see L2). I reviewed the commit, not the worktree.
- Probe sources are kept at `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-73-r3/zz_probe73_test.go` and `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-73-r3/zz_enum_test.go`. I removed them from `head` and `main`, and `mut` matches `head`.

### Class check, from the code
- **Typed RPCs a confined key can reach.** I walked every method of `cortexdb.v1` through `protoregistry` and kept those whose request has a top-level string `collection`, `user_id`, `scope` or `namespace`, or that are Rowless.
  - Writes: `InsertGraphDocument`, `SaveKnowledge`, `UpdateKnowledge`, `SaveMemory`.
  - Searches: `SearchGraphRag`, `SearchText`, `SearchKnowledge`, `SearchMemory`.
  - Rowless: `ListTools`, `Health`, `Info`.
  - Plus the five id-addressed RPCs and `CallTool`.
  - The five writes are all in `authorizeTypedArgs`, and all five have a row guard. The typed searches have no field like `return_paths`, so the path class of #72 does not reach them.
- **Row-id fields of the allowlist.** I count 11: `document_id` ×1, `knowledge_id` ×4, `memory_id` ×5, `supersedes` ×1. `TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC` exercises all 11 end to end, so the "eleven" claim holds.
- **Id values.** This is where the class is not complete: blank ids, L1 below.

### Findings

**L1 (likely comment; a blocker if #82 does not land first): an id made only of blanks skips every typed upsert guard, so a confined key overwrites and re-homes another scope's row.**
- Cause: `guardKnowledgeUpsert` and `guardMemoryUpsert` return early on `strings.TrimSpace(id) == ""` (`pkg/rpcserver/ownership.go:405`, `:429`). The facade refuses only `""`, and the tool guard (`authz.Scope.guardRow`) loads `" "` like any other id.
- Probes, the same on both trees:

| Call | Result | Row afterwards |
|---|---|---|
| notes key, `InsertGraphDocument{Id:"  "}`, over an operator document in `secrets` | OK | `coll="notes"`, `content="pwned by IGD"` |
| notes key, `SaveKnowledge{KnowledgeId:" "}` | OK | moved to `notes` |
| notes key, `UpdateKnowledge{KnowledgeId:"\t"}` | OK | moved to `notes` |
| hermes key, `SaveMemory{MemoryId:" "}`, over zeus's memory | OK | `user="hermes"` |
| notes key, `CallTool ingest_document {"document_id":"  "}` (head) | NOT_FOUND `not found` | unchanged |

- So the typed path and `CallTool` still disagree, against the body ("so the two paths cannot disagree") and the commit message ("InsertGraphDocument is guarded by its id as SaveKnowledge is"). It does hold, but only for ids that are not blank.
- Any confined key can create a blank-id row inside its own scope, so the precondition is not limited to the operator.
- My mutation N (`TrimSpace(id) == ""` → `id == ""` in `guardKnowledgeUpsert`) survives the whole package.
- #82 (`gc-11-rest-upsert-rows`) makes exactly that change, with `TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC`. Ask: name the dependency in this body, or merge #82 first.

**L2 (likely comment): the proof material and the PR disagree on whether the `InsertGraphDocument` guard is part of this PR.**
- `docs/pr/round2/failing_first.md` (seventh pass) says B1 "est sorti de #73 vers la PR à part `gc-12-grpc-ingest-guard`", and that set `73` now counts 33 mutations.
- But head 4f71850 still has the guard (`graphrag.go:35`), its test, and its case in `TestTheTypedRPCsWithholdADocumentTheyCannotPlace`.
- The body still says "Thirty-four" and claims the guard three times (its own paragraph, the reply, Breaking Changes), and so does the commit message.
- The worktree holds the removal, uncommitted.
- Decide which PR carries it. Then make the body, the commit message, the Breaking Changes section and the count agree, and expect a conflict on `graphrag.go` if both PRs land.

**L3 (likely comment): "UpdateKnowledge answers a missing id as a foreign one" has no test showing that other errors keep their code.**
- My mutation V made `withheldIfMissing` (`ownership.go:325`) rewrite every error of a confined caller (`st != nil` instead of `codes.NotFound`). It survives the package.
- Probe S4: a confined `UpdateKnowledge` of its own document with empty content is InvalidArgument `empty text provided` on both trees, which is correct but untested.
- `CallTool` has this case (the empty `memory_save`); the typed route needs the same one.

**L4 (likely comment): the rewritten `scopePermitsMemory` (`ownership.go:260`) has no rpcserver test for scope or namespace confinement.**
- Mutation X (pass `""` instead of `record.Namespace`) survives the package, and so does mutation Y (pass `""` instead of `record.Scope`).
- No rpcserver test uses a key confined to a namespace or a memory scope against a stored row. The `lab` and `mine` keys only call `SaveMemory`.
- Behaviour is correct on both trees: S5 (`GetMemory`) and S5b (`UpdateMemory`) under `lab`, and S5c (`GetMemory`) under `mine`, are all NOT_FOUND. But "the typed RPCs' own row checks call the rules the policy exports, so the two paths cannot disagree" is held only by `pkg/authz` tests of `PermitsMemory`, not by the call that passes it the fields.

**N1 (nit, pre-existing): the typed guards and the tool guard sort "missing" differently.**
- The typed guards decide "missing" by text: `toStatus` maps any message containing "not found" to NOT_FOUND (`errors.go:34`).
- The second read of each upsert guard then lets any error through as a create (`ownership.go:417`, `:437`, `gerr != nil` → `return nil`).
- `authz.isNotFound` accepts only `core.ErrNotFound` and `sql.ErrNoRows` (`rows.go:143`). Either name this or align the two.

**N2 (nit):** Performance Impact says "one indexed read per id argument".
- `GetKnowledge` also loads every chunk and the entity names of each chunk (`loadKnowledgeRecord`).
- The typed upsert guards read twice when the call creates a new row.

**N3 (nit):** the commit subject is 79 columns.

**N4 (nit):** under Type of Change, both "New feature (non-breaking …)" and "Breaking change" are ticked.

**N5 (nit):** "Two things this PR does not change, so you can see them" is followed by more than two items.

**N6 (nit, weak assert):** the `InsertGraphDocument` case of `TestTheTypedRPCsWithholdADocumentTheyCannotPlace` (`confined_chunkless_test.go:85`) runs without an embedder. Without the guard the call fails Internal, so the "content unchanged" check below cannot fail for that case, and only the code check holds it. Use `fakeEmbedder`, as `TestATypedIngestCannotTakeADocumentOfAnotherCollection` does.

### My mutations (`mutone.py`, `-run .`, `./pkg/rpcserver/`)

| # | Mutation (claim aimed at) | Result |
|---|---|---|
| A | `SaveKnowledge` metadata not handed to the allowlist | killed: TypedWriteTakesWhat…, …Signature… |
| B | `UpdateKnowledge` metadata not handed | killed: TypedWriteTakesWhat… |
| C | `guardToolRows`: ErrDenied → NOT_FOUND | killed: TheCallToolChecksFailClosedOnTheirOwn |
| D | `authorizeToolArgs`: a message that is not a `CallToolRequest` passes | killed: same |
| F | `InsertGraphDocument` uses the read guard, so creates are refused | killed: TypedWriteTakesWhat…, TypedIngest… |
| G | `scopePermitsKnowledge` checks the collection only | killed: TestAConfinedKeyIsRefusedARowWhoseKindCannotExpressItsConfinement |
| N | upsert guard: blanks count as an id | **survived** (L1) |
| V | `withheldIfMissing` rewrites every code | **survived** (L3) |
| X | `scopePermitsMemory` drops the namespace | **survived** (L4) |
| Y | `scopePermitsMemory` drops the scope | **survived** (L4) |

Ten mutations: six killed, four survived, none of them equivalent.

### Scenarios, same probe on both trees (fake embedder; keys operator, hermes (user_id), notes (collection), chunks (graphrag_chunks), lab (namespace), mine (memory scope))

| Scenario | base (#72) | head |
|---|---|---|
| S1 notes `InsertGraphDocument id=sec` (secrets doc) | OK, moved to notes, overwritten | NOT_FOUND `knowledge "sec" not found`, unchanged |
| S2 chunks `InsertGraphDocument` of a chunkless doc | OK | NOT_FOUND |
| S3a notes tool `ingest_document document_id="  "` | PermissionDenied (CallTool refused) | NOT_FOUND `not found`, unchanged |
| S3b notes `InsertGraphDocument id="  "` | OK, moved to notes | **OK, moved to notes** (L1) |
| S3c notes `SaveKnowledge id=" "` | OK, moved | **same** (L1) |
| S17 hermes `SaveMemory id=" "` over zeus's memory | OK, re-homed to hermes | **same** (L1) |
| S18 notes `UpdateKnowledge id="\t"` | OK, moved | **same** (L1) |
| S4 notes `UpdateKnowledge` own doc, empty content | InvalidArgument | same (untested, L3) |
| S6 / S6b notes `UpdateKnowledge` missing / foreign | `…embedding not found` / OK (foreign overwritten) | `knowledge "nosuch" not found` / `knowledge "sec" not found` |
| S5 / S5b / S5c lab Get / Update other namespace; mine Get scope=global | NOT_FOUND | same (untested, L4) |
| S7 hermes `SaveMemory metadata._x` | OK | PermissionDenied, named |
| S8 notes `InsertGraphDocument metadata.written_by` | OK | PermissionDenied |
| S9 notes `SaveKnowledge metadata.Written_By` | OK | PermissionDenied |
| S10 notes tool `knowledge_search` | PermissionDenied | OK |
| S11 operator `InsertGraphDocument` over sec | OK | OK |
| S12 notes `UpdateKnowledge` n1 → secrets | PermissionDenied (RPC wording) | PermissionDenied (tool wording, as r2 N1) |
| S13 / S16 hermes `SaveMemory` / `UpdateMemory` zeus-1 | NOT_FOUND | same |
| S14 notes `InsertGraphDocument` empty id | Internal | same |
| S15 hermes tool `memory_update metadata.written_by` | PermissionDenied (CallTool) | PermissionDenied, named |

### Before approval
- L1: land #82 first, or name the blank-id dependency in the body. The `InsertGraphDocument` claim holds only after it.
- L2: settle which PR carries the `InsertGraphDocument` guard and update the body, the commit message and the count to match.
- L3 and L4: add the missing test cases (a confined `UpdateKnowledge` that keeps a non-NOT_FOUND code; a namespace key and a memory-scope key against a stored row through the typed RPCs).
