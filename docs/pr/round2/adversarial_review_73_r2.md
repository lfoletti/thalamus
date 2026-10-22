# Relecture adverse de #73, tour 2 (21 octobre 2026, 23 h 17)

Tête `1c149f7` (`gc-04b-grpc-confined-tools`), base `d805527`. Agent `upstream-reviewer`, kit `adv-73-r2`, invite fixe de `docs/pipeline_pr.md`, section 1 bis, avec le rapport du tour 1 en matériel. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 34, contexte 19219 -> 112667 tokens ; écriture cache 112665, lecture cache 2229660, entrée neuve 68 ; équivalent entrée (lecture x0,1, écriture x1,25) : 363797 ; outils : {'Bash': 33, 'SubagentHandback': 1}. Durée 7 min 55 s.

---

## Review of #73 (gc-04b-grpc-confined-tools, head 1c149f7, on #72 at d805527), round 2

**Verdict: one blocker, three likely comments, five nits.** The round-1 points L1, L2, L3, L4 and N1 are closed, and I checked each one myself (scenarios S3, S7 and S8 below, mutations M6 and M7). The blocker is a write path the PR claims to cover and does not: the typed `InsertGraphDocument` has no row guard.

**What I ran**
- `go test -race ./pkg/rpcserver/ ./pkg/authz/` passes on both trees (both.sh).
- One probe file was placed in both copies and then removed. No `zz_*` file is left in head, main or mut, and mut matches head.
- Outside the tests the diff is about 200 lines, in `ownership.go`, `tools.go` and `knowledge.go`.
- The author's mutation table for #73 has 31 rows; I counted them myself. Of the six new or renamed tests, five fail on main once their files are copied there. `TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks` passes on main, which is expected: it pins main's answer. `failing_first.md` records that it failed on the previous head `ff24c70`.

### Class check (from the code)
- **Typed RPCs a confined key can reach.**
  - `AuthorizeRows` waives only Health, Info and ListTools as Rowless.
  - From the protos, a confined key can reach these request messages: the five writes the PR covers, `DeleteKnowledge` and `DeleteMemory` (id only, waived to the handlers), and the searches.
  - `InsertText`, `InsertTextBatch`, `HybridSearchText` and every graph, decision and contract RPC have no scope field, so a confined key is refused them.
  - So the five writes are complete for *arguments*.
- **Row ids of those writes.** This is the class the PR leaves open:

| Write | Row guard |
|---|---|
| SaveKnowledge | `guardKnowledgeUpsert` |
| UpdateKnowledge | `guardKnowledgeUpsert` |
| SaveMemory | `guardMemoryUpsert` |
| UpdateMemory | `guardMemory` |
| InsertGraphDocument (`id`) | **none** (`graphrag.go:31-41`). The matching tool line, `ingest_document.document_id`, is guarded. |

- **The chunkless-document rule.** It applies to GetKnowledge, DeleteKnowledge and, through `guardKnowledgeUpsert`, to SaveKnowledge and UpdateKnowledge. `InsertGraphDocument` again gets nothing.

### Findings

**B1 (blocker): a collection-confined key overwrites and takes over another collection's document through typed `InsertGraphDocument`.**
- Probe: the operator saves `sec` in `secrets`. The `notes` key then calls `InsertGraphDocument{Id:"sec", Collection:"notes", Content:"pwned by notes"}`.
  - The call returns OK on main and on head.
  - The operator then reads `sec` back as `collection="notes"`, `content="pwned by notes"`.
- The same write through `CallTool` (`ingest_document`, `document_id: sec2`) is `NOT_FOUND "not found"` on head.
- The same probe also overwrote the chunkless document `orph` through `InsertGraphDocument` from the `graphrag_chunks` key (S6c: OK on head too).
- This has been on main since before the PR. But this round the PR's claim is that the typed writes and `CallTool` "cannot disagree" (`ownership.go` comment on `authorizeTypedArgs`, and "Typed writes take what the tool takes" in the body). `guardKnowledgeUpsert`'s own comment describes exactly this attack, and that guard is not on this RPC.
- I would not merge confinement over gRPC with a cross-collection overwrite left on a typed write.
- Fix: call `guardKnowledgeUpsert(ctx, s.db, req.GetId())` in `InsertGraphDocument`, and add a test with an embedder: `fakeEmbedder` is already in `harness_test.go:17`.

**L1 (likely comment): the chunkless rule on the typed upserts has no test.**
- On head, SaveKnowledge and UpdateKnowledge of a chunkless foreign document from a key confined to the default collection now answer `NOT_FOUND knowledge "orph" not found`. On main both were OK (S6, S6b).
- `TestTheTypedRPCsWithholdADocumentTheyCannotPlace` covers only Get and Delete.
- Own mutation M1 survived the whole package: let `guardKnowledgeUpsert` (`ownership.go:437`) through when the document has no chunks.

**L2 (likely comment): the body contradicts itself.**
- The Fix section still says "Unconfined keys, **the typed RPCs**, `ListTools` and the read-only clearance are unchanged." The new section and Breaking Changes say the typed RPCs changed for confined keys; so does the chunkless withholding.
- "Type of Change: New feature (non-breaking)" is ticked while Breaking Changes names a break.
- The Tests paragraph still says "Three mutations, three killed" next to the new "Thirty-one".

**L3 (likely comment): "every NOT_FOUND of a tool becomes that bare `not found`, `memory bucket not found` included" has no test.**
- Own mutation M3 survived: in `oneNotFound`, keep the tool's wording when it contains "bucket".
- No test file in the package mentions a bucket outside a comment (`confined_tools_review_test.go:77`).

**N1 (nit):** typed refusals now name a tool the caller did not call.
- hermes `SaveKnowledge` and `InsertGraphDocument`, main: `.../SaveKnowledge carries no user_id field`.
- Head: `tool knowledge_save takes no user_id argument` (and `ingest_document` likewise). The code is still PermissionDenied.
- This comes from `authorizeTypedArgs` running before `AuthorizeRows`. It is not named.

**N2 (nit):** "the typed RPCs keep main's wording, `memory "X" not found`" is not the whole list. Typed `SaveMemory` into another user's session answers `NotFound memory bucket not found` on both trees (S9), where `CallTool` flattens it.

**N3 (nit, carried from r1 N2):** the weak assert is still there.
- `confined_typed_args_test.go:114-117` accepts any error except PermissionDenied for `InsertGraphDocument` "because the test brain has no embedder".
- A fake embedder is available in the package, so the test could assert OK and the stored metadata.

**N4 (nit):** commit message layout. The wraps are broken ("A document whose chunks are gone,\nwhich ..."), and one line is past 72 columns ("... A tool off the allowlist stays refused").

**N5 (nit):** equivalent mutant, so nobody reads it as a guard. M2: hand empty `user_id`, `scope` and `namespace` to the allowlist line (`ownership.go:198`). It survives, and it is equivalent: an empty confined field is refused by `AuthorizeRows` anyway.

### Own mutations (`mutone.py`, `-run .` on `./pkg/rpcserver/`)

| # | Mutation (aimed at) | Result |
|---|---|---|
| M1 | `guardKnowledgeUpsert` lets a chunkless document through (chunkless claim, typed upserts) | **survived** (L1) |
| M2 | empty SaveMemory scope fields handed over (`ownership.go:198`) | survived, equivalent (N5) |
| M3 | `oneNotFound` keeps "bucket" wording (bucket claim) | **survived** (L3) |
| M4 | guard's NOT_FOUND wording gets the tool name appended (one-wording claim) | killed: TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| M5 | `idAddressedMethods` waiver moved before `authorizeTypedArgs` (UpdateMemory checked) | killed: TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| M6 | UpdateKnowledge relations dropped (r1 L2) | killed: same test |
| M7 | SaveMemory namespace dropped (r1 N1) | killed: same test |
| M8 | `guardToolRows` read failure turned into NOT_FOUND (an outage hidden as a missing row) | killed: TestTheCallToolChecksFailClosedOnTheirOwn |

Eight mutations: five killed, two real survivors, one equivalent.

### Scenarios on main and head (same probe, fake embedder, keys operator, hermes (user_id), notes (collection), chunks (graphrag_chunks))

| Scenario | main | head |
|---|---|---|
| S1 notes `InsertGraphDocument id=sec` (secrets doc) | OK, doc moved to notes and overwritten | **same** (B1) |
| S2 notes `ingest_document document_id=sec2` | PermissionDenied (opaque JSON) | NotFound `not found` |
| S3 notes Save, Get, then Update/Save with the metadata it read back (`written_by:notes`) | OK / OK | OK / OK (r1 L3 closed) |
| S4 hermes SaveKnowledge / InsertGraphDocument | PermissionDenied "carries no user_id field" | PermissionDenied "tool … takes no user_id argument" (N1) |
| S5 hermes SaveMemory `metadata._k` top level | OK | PermissionDenied (named) |
| S5 hermes SaveMemory nested `metadata.a._k` | OK | OK (tool `memory_save` also OK: consistent) |
| S6 chunks UpdateKnowledge / SaveKnowledge of a chunkless secrets doc | OK / OK | NotFound `knowledge "orph" not found` (untested, L1) |
| S6c chunks InsertGraphDocument of the same doc | OK | **OK** (B1) |
| S7 notes UpdateKnowledge missing / foreign | `…embedding not found` / `knowledge "sec2" not found` | `knowledge "nosuch" not found` / `knowledge "sec2" not found` (r1 L1 closed) |
| S7c notes UpdateKnowledge own doc, empty content | InvalidArgument `empty text provided` | same |
| S8 hermes `memory_update` metadata.written_by (tool) / UpdateMemory (typed) | denied (opaque) / OK | PermissionDenied / OK (as the body says) |
| S9 hermes SaveMemory into zeus's session | NotFound `memory bucket not found`, no row | same (N2) |
| S10 notes InsertGraphDocument `metadata.document_id` | OK | PermissionDenied (named) |

### Tests that sleep, read the environment, or pass without the code
- No `Sleep`, `Getenv`, `Setenv`, `LookupEnv`, `UserHomeDir` or `time.After` in the new or changed test files.
- Tests that pass without the code: only `TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks`, by design. I would say so in the body.
- Weak asserts: N3, and the chunkless test checks the code only, not the wording. The wording comes from `withheld` on both sides, so that one does not matter.

### Before approval
- B1: guard the typed `InsertGraphDocument` id, with a test using the fake embedder.
- L1 and L3: add the two missing test cases.
- L2: make the body agree with itself.
