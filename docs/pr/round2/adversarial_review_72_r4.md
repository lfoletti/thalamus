# Relecture adverse de #72, tour 4 (21 octobre 2026)

Rapport de l'agent neuf sur la tête `f1b24eb` (invite fixe, `docs/pipeline_pr.md`, section 1 bis), recopié tel quel. Deux bloquants : #72 non poussé.

Review of #72 (`gc-04a-authz-tool-rules`, head f1b24eb, base cc98c6c). Read-only: nothing pushed, nothing posted, the worktree is untouched. All tests ran under the canary environment, in copies exported with `git archive` to `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/d0bc1076-b986-44ac-b9bd-ec272ada2308/scratchpad/adv72/{head,main}`. My probe tests are the `zz_adv*_test.go` files in that copy.

**Verdict:** request changes. There are two blockers: any confined key can crash the gRPC server, and any confined key can forge `written_by` through `ingest_document`.

## What I verified myself

- **Suites:** `go test -race ./pkg/authz/` passes. `go test -race -run 'Memory|Supersede' ./pkg/cortexdb/` passes. The new tests contain no `time.Sleep`, no `os.Getenv` and no `t.Setenv`.
- **Tool counts:** `ToolDefinitions()` has 99 tools. 20 are on the allowlist and 79 are refused.
- **Table:** the table `ConfinedToolsTable()` generates is byte-identical to the table in the body.
- **Size:** 850 lines added outside tests, 12 removed, as the body says.
- **Row-id fields:** there are 11, as the body and the test say.
- **Mutations:** the table in `failing_first.md` (#72 section) has 21 rows, matching "twenty-one".
- **Scenarios replayed on `main` and on the head:**
  - `GetMemory` by list under a `CallerScope`: `main` serves `zeus-1`. The head lists it under `missing`, in order. This is named in the body.
  - `SaveMemory` with a missing `supersedes` target: `main` writes `new-1` and marks `hermes-1`. The head writes neither. This is named under Breaking Changes.
  - `SaveMemory` with a foreign `supersedes` target under a `CallerScope`, calling the facade directly: `main` and the head both mark Zeus's row. Unchanged and not named (see L1).

## Blockers

**B1. `ConfineToolArgs` panics on `null` arguments, which kills the gRPC server.**
- `allowlist.go:355` calls `argsObject` (`:484-492`). `json.Unmarshal("null", &obj)` sets the map to nil and returns no error. Line `:370` (`obj[gate] = true`) then panics with "assignment to entry in nil map".
- Probe `TestAdvNullArgs`: a key confined to `notes`, tool `knowledge_memory_recall`, arguments `null` → recovered panic, that exact message.
- The doc comment (`:344-346`) says non-object arguments "come back unchanged". The test at `allowlist_test.go:278-291` covers `[1]` and never `null`.
- Both ports call `ConfineToolArgs` first, on the raw caller input:
  - #73: `rpcserver/ownership.go:121`, inside the auth interceptor.
  - #74: `httpapi/policy.go:143`.
- The `pkg/rpcserver/metrics.go:99` comment confirms that grpc-go does not recover a panicking handler. So once #73 lands, any confined key can take the process down with `CallTool{name:"knowledge_memory_recall", args_json:"null"}`.

**B2. A confined key can forge `written_by` / `updated_by` through `ingest_document`.** This is the same class as round-2 point 2: provenance written through metadata.
- `cortexdb/written_by.go:13-19` promises "Whatever a caller wrote under those keys itself is replaced, so a client cannot write as someone else". But only `SaveKnowledge`, `UpdateKnowledge`, `SaveMemory` and `UpdateMemory` call `stampWriter` or `stampUpdater`. The ingest path (`graphrag_tool_ingest.go` ~`:99-106`) copies `req.Metadata` with no stamp.
- `checkOpenKeys` (`allowlist.go:466-479`) refuses only `_*` keys and `document_id`, `chunk_index`, `graph_kind`.
- Probe `TestAdvWrittenByOnIngest`: `AuthorizeToolArgs` accepts `ingest_document` with `metadata:{"written_by":"zeus-key","updated_by":"zeus-key"}` from key `notes-key`. The stored document metadata is `{written_by:zeus-key, updated_by:zeus-key}`. The same call through `knowledge_save` stores `written_by:notes-key`.
- The `_` rule protects no key that exists on `main`. `inventory_72.md` does not list `written_by` or `updated_by`.

## Likely comments

**L1. The facade confines the `memory_get` list but not `supersedes`.**
- `memory_api.go:123` / `:1285`: `checkSupersedeTargets` reads through `loadMemoryRow` without `memoryReadPermitted`. The PR added that check to `GetMemory`'s list form at `:248` in the same file.
- Replayed scenario: under `CallerScope{hermes}`, `SaveMemory(supersedes:["zeus-1"])` stamps `superseded_by` on Zeus's row, on both `main` and the head.
- Only the port guard stops this. A library or gateway caller that relies on v2.135.0's `CallerScope` does not get the protection. Either confine it in the facade or say why not.

**L2. The gate check does not read the gate the way `encoding/json` does.**
- `allowlist.go:332` uses `lookupKey`, which takes the exact spelling first. `encoding/json` decoding into a struct takes the last occurrence.
- Probe `TestAdvGateDuplicateSpelling`: `{"collection":"notes","disable_memory":true,"Disable_Memory":false}` passes `AuthorizeToolArgs`, and the request then decodes with `DisableMemory=false`.
- `ConfineToolArgs` removes every spelling, so the ports are safe today, and v2.135.0 enforces this underneath anyway. But the body says "AuthorizeToolArgs refuses a call whose gate is off" and the test (`:251`) says the gate "counts under any spelling, as encoding/json reads it". Neither holds for duplicate spellings.

**L3. The row guard places a document with no chunks in `graphrag_chunks`.**
- `rows.go:125-133` trusts `GetKnowledge(...).Collection`. `knowledgeCollection` (`knowledge_helpers.go:134-136`) returns `defaultGraphRAGCollection` when no chunk exists.
- Probe `TestAdvChunklessForeignDocument`:
  1. A document is saved in `secrets` and its embeddings are removed (`Vector().DeleteByDocID`).
  2. A key confined to `graphrag_chunks` now passes the guard on `knowledge_get`.
  3. `GetKnowledge` returns `content="the retainer"`.
- Not reachable through the toolbox alone, and pre-existing in the facade. But the guard should fail closed on a collection it cannot determine.

**L4. "A foreign `document_ids` pin yields nothing" has no test in this PR.**
- The only test is `rows_test.go:47-53`, against `fakeRows`.
- My real-brain probe (`TestAdvDocumentIDsPin`) confirms the claim holds, with and without `CallerContext`, on `search_text`, `knowledge_search`, `search_graphrag_lexical`, `knowledge_memory_recall` and `knowledge_memory_build_context_pack`. The claim needs a test on a real brain.

## Nits

- **N1. Stale counts in the body.** The "Tests" section still says "Fifteen such mutations… (42 mutants)", while the round-2 section says 21. Internally, `failing_first.md` says the test requires "au moins 8 champs", but the test asserts `== 11`.
- **N2. `title` is missing from the engine chunk keys.** The engine writes `title` on every chunk (`graphrag_tool_ingest.go:103`, `knowledge_tx.go:226`), and `metadata.title` overrides it. Harmless, since the caller sets the title anyway, but the class in the body ("the keys the engine writes on every chunk") is incomplete.
- **N3. Memory metadata keys the engine reads are writable on the key's own rows.** `superseded_by` and `superseded_at` can be set by a confined key on its own memories through `memory_save` / `memory_update`, which un-retires or hides them. The impact stays within the key's scope.
- **N4. "Extraction … keeps the rest" has no test in this PR.** My probe (`TestAdvExtractionKeepsEntity`) shows a confined `knowledge_save` adds `notes` to a `secrets` entity's `collections` and also adds a `type` property. Searches under the `notes` key did not reveal the entity's description or properties.
- **N5. `checkPaths` and `encoding/json` fold keys differently.** `checkPaths` lowercases with `strings.ToLower`, while `encoding/json` folds with `SimpleFold`. For example, `memory_İd` passes `checkPaths` but is ignored by the decoder and by `JSONPathValues`. Harmless (it accepts a key the tool ignores), but the stated equivalence is not exact.
- **N6. The supersede check-then-write is not atomic.** A target deleted between `checkSupersedeTargets` and `markMemoriesSuperseded` leaves the new row written and returns an error, which is the old behaviour in a narrow window. This is not named next to the guard's own TOCTOU note.

## Held up

These all behaved as claimed:
- the 11 row-id fields answer the bare `not found` for a foreign id and let a missing one through;
- `memory_get` refuses both forms together, under any spelling;
- the path-retrieval arguments are refused;
- `entities` / `relations` are refused whole;
- `session_id` reads on `memory_search`, recall and context pack answer a foreign session and a missing one identically (both empty);
- the completeness test against `ToolDefinitions()` holds;
- Unicode spellings of scope fields are refused.

