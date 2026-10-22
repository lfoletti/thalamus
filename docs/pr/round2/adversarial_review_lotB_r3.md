# Relecture adverse du lot B, tour 3 (21 octobre 2026)

Rapport de l'agent neuf (invite fixe, `docs/pipeline_pr.md`, section 1 bis), recopié tel quel. Suite donnée : #72 corrigé (B1, B2, B3, L5, N1, N6, L2 ; L1, L4, N5 nommés dans le corps), tête `f1b24eb` ; L3, L6, N4 vont à #73 et #74.

## Review of #72 / #73 / #74, round 3, read-only

**Heads reviewed:** 842995b (#72), abbeeee (#73), 086bbc9 (#74), with main at cc98c6c.

The heads moved while I was reviewing. They are now c4ede28, 12785e3 and f8457c0. The only change is 5 test lines in `pkg/cortexdb/memory_get_scope_test.go`, which close survivor M2 below. Nothing else here is affected.

**How I worked:**
- I ran nothing in your worktrees; they show 0 changes. I exported each head with `git archive` into my scratchpad (`.../d0bc1076-.../scratchpad/r04b`, `r04c`, `rmain`) and ran probe tests and mutations there.
- Every run used the canary environment you gave.
- The PR suites are green under the canary: `pkg/authz`, `pkg/rpcserver -race`, `pkg/httpapi -race`, and the two new `pkg/cortexdb` tests.

### Blockers

**B1. `search_graphrag_lexical.return_paths` lets a confined key do what the refused `search_paths` does.**
- The allowlist line (allowlist.go:165-166) allows `return_paths`, `max_paths`, `path_max_depth` and `relation_policies.*`.
- `attachGraphRAGPaths` (pkg/cortexdb/graphrag_paths.go:293-327) calls `SearchPaths` with no collection.
- Probe on a clean brain: the operator writes `hr` in `secrets` with a relation. A key confined to `notes` calls `{"collection":"notes","entity_names":["Project Nightshade"],"return_paths":true}`.
- It gets back `Paths[0]`: sentence "Project Nightshade merges with Acme.", edge `document_id:"hr"`.
- That is a fact from another collection plus a foreign document id. The same line reaches REST in #74.

**B2. `entities[]` and `relations[]` on `knowledge_save`, `knowledge_update` and `memory_save` write the shared graph.**
- Entity ids are normalised names, one node across all collections (entity_collections.go:6-12). This is the same reason 60+ tools sit in the refused table.
- Probe results:
  - The `notes` key rewrote `entity:project_nightshade` (a `secrets` entity) to description "DEFACED by notes" and added `owner`.
  - `hermes`, a user-confined key, rewrote `entity:acme` through `memory_save`.
  - `notes` added an edge `Project Nightshade -funds-> Orpheus`, which `return_paths` then serves.
- The `collections` property is missing from `reservedMetadataKeys` (allowlist.go:131-136):
  - The engine writes it (entity_collections.go:22,38) and the walk gate reads it (`entityKnownIn`; graph_runtime_sql.go:56).
  - `hermes` set `entities[].metadata.collections=""` on two `secrets` entities, which makes them "known everywhere".
  - `notes` set `"collections":"notes"` on Acme, which dropped `secrets` from the list, so walks confined to `secrets` no longer pass through Acme.
  - inventory_lotB.md:195 files `collections` as "hors classe" (out of class).
- A relation to an entity that does not exist fails with Internal "FOREIGN KEY constraint failed"; to an existing one it succeeds. That tells a confined key which entity names exist in other collections.
- adversarial_review.md:114 already notes "shared-node rewrite by name stays open", but the body does not say so.

**B3. "One answer for a foreign id and a missing id" fails on `memory_get`.**
- The guard reads `memory_id` (allowlist.go:196, rows.go:89-99). `GetMemory` ignores `memory_id` once `memory_ids` is set (memory_api.go, `if len(req.MemoryIDs) > 0`).
- Probe:
  - `{"memory_id":"zeus-1","memory_ids":["hermes-1"]}` returns NOT_FOUND "not found".
  - `{"memory_id":"nope","memory_ids":["hermes-1"]}` returns OK with hermes-1.
- That is a one-call existence oracle on every memory id. Fix: refuse both arguments together, or skip the single-id guard when a list is present.

### Likely comments

**L1. The body's account of the four create tools is wrong.**
- The body says "only the owner of the foreign row could tell the two apart". In fact the caller sees it directly:
  - `ingest_document` with `document_id=hr` returns NOT_FOUND; with an unused id it creates the document.
  - `knowledge_memory_remember` behaves the same way.
- `memory_save.session_id` is classified as data but is a fifth door: a foreign bucket answers "memory bucket not found" (memory_owner.go:92-94), an unused one is created.
- The commit message's "the id space is not an oracle" is false. This may be inherent to global ids, but it has to be stated, not denied.

**L2. `TestAForeignIdAndAMissingIdAnswerTheSameOnEveryRowIdField` (rows_test.go:62-94) does not test what its name says.**
- It expects `ErrWithheld` for a foreign id and `nil` for a missing one: different answers, at the guard only.
- End to end, the same wording is tested for 5 of the 11 fields over gRPC and REST: `memory_get`, `memory_delete`, `supersedes`, `knowledge_get`, `knowledge_delete`.
- I checked `knowledge_update` and `memory_update` end to end myself: they do match. The body and the reply still overstate what the test holds.

**L3. Typed `SaveKnowledge` accepts what the allowlist refuses on `CallTool`.**
- A confined key can send `entities[].chunk_ids`, `relations[].chunk_ids`, `entities[].metadata.chunk_ids`, `_`-prefixed keys at the top level and in entity metadata, and `collections` (knowledge.pb.go:160-161; `UpdateKnowledgeRequest` has the same entity and relation fields at 352-353; I probed only `SaveKnowledge`).
- Same result on main and on abbeeee: no error, and `entity:acme` gained `_x`, `chunk_ids` and the new description.
- Not a regression, but #73 says "the two paths cannot disagree", and they disagree on arguments.

**L4. #72 says it "changes no behaviour on either port", with Breaking Changes: None.**
- Scenario run on main and on the branch, unconfined caller, `memory_save` of a new row with `supersedes:["old","missing"]`:
  - Main: the row is written and `old` gets `superseded_by=new`.
  - Branch: no row, and `old` is untouched.
  - Both return the same error: `supersede "missing": embedding not found`.
- This is visible on every port for every key, so it should be named as a behaviour change.

**L5. `isNotFound` (rows.go:133-136) treats any error text containing "not found" as "row does not exist", and lets the call through.**
- Removing the substring clause survives the whole suite, because `core.ErrNotFound` already covers missing rows.
- I suggest dropping it. As written, it contradicts the body's "fails closed on a read that fails".

**L6. Untested wiring.**
- Broadening `oneNotFound` to every error code for a confined key survives, on gRPC (ownership.go) and on REST (policy.go). No test checks that a confined key keeps a non-404 error from a tool.
- Dropping `json.Valid` from the REST 400 check survives. A body like `{"a":` then gets 403, against the body's "not a JSON object is 400".

### Nits

- **N1.** `gate == "" ||` at allowlist.go:313 is redundant. The mutant is equivalent, because `checkPaths` refuses an empty key first.
- **N2.** Dropping `r.ContentLength` in REST survives; probably equivalent.
- **N3.** At the reviewed heads nothing tested the self/blank skip in `checkSupersedeTargets` (mutant survived). Fixed by c4ede28.
- **N4.** The round-2 tests on both ports check that a list serves nothing foreign, but never that `hermes-1` is actually served (gRPC `confined_tools_round2_test.go`; REST same file, lines 92-95).
- **N5.** Between `GuardToolRows` and the tool's upsert there is a window (check then write) that the body does not mention.
- **N6.** `ingest_document` copies caller metadata over the engine's chunk keys `document_id`, `graph_kind`, `chunk_index` and `title` (graphrag_tool_ingest.go:105). Same class as round-2 point 2. I found no reader of the chunk's `document_id`, so impact looks low.

### What holds

- Graph walks without `return_paths`, checked with a shared entity across `secrets` and `notes`: `knowledge_search`, `search_graphrag_lexical`, `recall`, `build_context_pack` and `search_text` returned no `secrets` text, relation or id.
- A user-confined key's `memory_search` (including PPR), `recall` and `context_pack` did not reach Zeus's memories.
- `aggregate_metadata`, `representative_records` and `search_vector_range` apply `Collection`.
- The added tests contain no sleeps and no environment reads. The only opens are three `DefaultConfig(t.TempDir())` calls.

### Counts I computed

- Lines outside tests: #72 +814/−12 (matches the body), #73 +76/−11, #74 +69/−7.
- 20 tools on the allowlist.
- 11 row-id fields: `document_id` ×1, `knowledge_id` ×4, `memory_id` ×5, `supersedes` ×1 (matches the body and the reply).
- My hand mutations: 14 run, 7 killed (M4, M7, M9, M11–M14), 7 survived (M1, M2, M3, M5, M6, M8, M10). Of the survivors, M1 and M8 are equivalent and M2 is fixed on the new head.

The probe and mutation files are in my scratchpad: `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/d0bc1076-b986-44ac-b9bd-ec272ada2308/scratchpad/mut.py` and `.../scratchpad/zz_main_test.go`.
