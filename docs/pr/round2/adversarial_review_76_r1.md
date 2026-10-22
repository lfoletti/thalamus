# Relecture adverse de #76, tour 1 complet (22 octobre 2026)

Tête `9dc1016` (`gc-05-expires-at`), base `dcdccbf` (#75, sur v2.136.0). Agent `upstream-reviewer`, kit `adv-76-r1`, invite fixe de `docs/pipeline_pr.md`, section 1 bis. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 79, contexte 18933 -> 175162 tokens ; écriture cache 175160, lecture cache 8942008, entrée neuve 158 ; équivalent entrée (lecture x0,1, écriture x1,25) : 1113150 ; outils : {'Bash': 78, 'Read': 1, 'SubagentHandback': 1}. Durée 16 min 15 s.

---

I reviewed #76 at gc-05-expires-at 9dc1016 against its base dcdccbf, as the upstream maintainer would in round 3. Verdict: changes requested. Most of round 2 holds, but I found two blockers. One write path lifts an expiry. Two tools the class table marks "gated" serve the expired document, and the class test only looks at them in ways that cannot fail.

Paths below are relative to the kit's head copy, `$K/head`, with `K=/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-76-r1`.

**Method.** I read the whole code diff, plus `expiry.go`, `sweep_expired.go` and `expiry_tool.go`. I listed each class from the code:
- every `store.*` read call;
- every exported search method of `DB`, `Quick`, `GraphRAGToolbox` and `KnowledgeMemory`;
- every non-mutating tool (61 on head, the same as the table);
- every typed RPC handler in `pkg/rpcserver`;
- every write to the `documents` table.

All test names cited in the body and replies exist. The suites I ran are green on head under the canary: the expiry tests in `pkg/cortexdb` (32 tests, 1.3 s), and `pkg/httpapi`, `pkg/rpcserver`, `cmd/cortexdb-mcp-stdio` and `pkg/authz`. `TestBothStoresLeaveExcludedDocumentsOutOfAggregates` passes on SQLite and on PostgreSQL. All probe files are removed and `mut` is identical to head.

## Blockers

**B1. `InsertGraphDocument` lifts a stored expiry, so the write class is incomplete.**
- `pkg/cortexdb/graphrag.go:197-203` builds the record with no metadata.
- `upsertGraphRAGDocumentRecord` (`pkg/cortexdb/graphrag_helpers.go:289`) then sets `existing.Metadata = doc.Metadata`, which is nil.
- The library call and the typed RPC (`pkg/rpcserver/graphrag.go:31`) re-ingest over an expiring document and remove `_expires_at`.
- Probe: `SaveKnowledge` with a past expiry, then `InsertGraphDocument` with the same id. `SearchText` went from 0 hits to 1, and the stored `_expires_at` was empty.
- This contradicts the body ("absent keeps the stored expiry … so a sync job that re-ingests from the source does not lift a retention rule"). That promise is only made for three tools, and the fourth write path to the record is neither handled nor named.

**B2. Two "gated" tools are not gated, and the class test cannot fail for them.**
- `knowledge_memory_shortest_path` has no gate (`pkg/cortexdb/brain.go:344` returns `Graph().ShortestPath` as is).
  - Probe: `from_node_id: chunk:old:000` (or `doc:old`) to the entity `Lund` returned both nodes with `content: "Arvid Lund was raised in Lund."`.
  - `TestEveryReadToolIsGatedOrNamed` asks for entity → `doc:old`. The edges are directed, so that answers "no path", and the test skips that error (`expiry_class_test.go:136`). The check can never fail.
- `search_paths` returns the expired chunk ids and `document_id: "old"` in `edges[].chunk_ids` and in `chunk_ids` (`pkg/cortexdb/graphrag_paths.go:164-167`), as soon as the expired document carries a relation.
  - The test's leak list includes `oldChunk`. But its fixture has no relations, and it drops `node_ids` for this tool (`:111`), so the response is empty.
- On the 18 tools the table marks gated, I added a positive check ("the live document is returned"). Two were not called because they have their own tests. 14 of the remaining 16 returned the live document; these two return nothing, so their check proves nothing.

## Likely comments

**L1. Unnamed difference from main with no expired document.**
- `searchGraphRAG` passes the plan filters into `gatedSearch` (`graphrag.go:514`). A `plan.filters.document_ids` filter now refills the budget ×4 until the store runs out.
- Scenario S1: `SearchGraphRAG` with `document_ids=[d11]`, `top_k=2`, where d11 ranks last of 12. main returns 0 chunks, head returns 1.
- The body says "with no expired document it asks once, as before". That is false for this input, and on a large corpus a selective filter now scans until the store is exhausted. Either name it, or pass nil filters and filter afterwards as main did.

**L2. Two rules with no test.** Both mutants below survive the whole `pkg/cortexdb` package.
- `query_api.go:166`: the post-filter after each prefetch. The comment says it is there for the graph lane, but the graph lane already goes through `SearchChunksByEntities` → `loadToolChunks`, which is gated. The lane it actually protects is the `source` lane: `runQuerySourcePrefetch` reads rows with an ungated `store.GetByID`.
  - My probe, a `QuerySource` returning the expired chunk id, kills the mutant. No test in the suite does. Fix the comment and add the test.
- `graphrag_tool_search.go:267` `delete(scores, node.ID)`: scores only exist with `relation_policies`, which no expiry test sets. My probe, with `relation_policies` from the expired chunk, kills it.

**L3. A far-future expiry is expired at once.**
- `normalizeExpiresAt` accepts `9999-12-31T23:30:00-01:00` and stores `10000-01-01T00:30:00Z`. As text, that sorts before `2026…`.
- Probe: such a document is hidden immediately, and a dry-run sweep lists it for deletion. `0000-…+01:00` gives `-0001-…`.
- The comment at `expiry.go:72` ("UTC RFC 3339 of one width") does not hold. Refuse years outside 0001–9999 after converting to UTC.

**L4. "Every typed read is gated" (body, "leaves as it is" list) is false.**
- The typed `GetKnowledge` (`pkg/rpcserver/knowledge.go:95`) serves the expired record.
- `FindKnowledgeGraph`, `QuerySparql` and `ExportKnowledgeGraph` serve the property graph.
- The tool equivalents are named in the body; the typed ones are not.

**L5. The PostgreSQL parity claim is weaker than written.**
- The parity test uses a stand-in condition (`ExcludeDocuments: "id = ?"`, `aggregate_exclude_parity_test.go:36`), not `expiredPredicate`. No facade expiry test opens a PostgreSQL brain.
- My PostgreSQL probe found the real path correct: search gated, aggregate count 1, dedupe ignores the expired holder, the sweep removes `old`, and `EXPLAIN` uses `idx_documents_expires_at`.
- So it is a coverage gap, not a defect. Say so, or add one facade test through `openPostgresBrain`.

## Nits

- `knowledge_api.go:124`: `|| existing.Metadata[expiresAtKey] != nil` is dead code. When the record has the key, `anyMapToStringMap` already returns a non-nil map. The mutant removing it survives the whole package (equivalent).
- `expiry.go:84`: the `<= ''` branch of `visiblePredicate` has no test, and its mutant survives the whole package. The engine never stores `''`, so either drop the branch or test a hand-written row.
- `knowledge_helpers.go:60`: the unconditional path (`knowledge_delete`) now returns `errKept` ("record kept: its condition no longer holds") when two deletes race on one id. main returned success. The message is wrong for that caller.
- Six tools the table marks "entities and relations" are not in the body's list: `disambiguate_mentions`, `fact_provenance`, `find_nodes`, `inference_explain`, `uncited_facts`, `verify_claims`.

## My mutations (14, none from the author's list)

10 killed, 4 survived. M8 failed to build on the first try and was rewritten; it is counted once.

| # | Mutation | Result |
|---|---|---|
| M1 | edges to dropped nodes kept in `withoutExpiredDocuments` | killed (`TestEveryReadToolIsGatedOrNamed`, `TestTheChunkToolsLeaveAnExpiredDocumentOut`) |
| M2 | `delete(scores, …)` removed | **survived** (whole package) |
| M4 | `.UTC()` dropped in `normalizeExpiresAt` | killed (`TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly`, `TestABadExpiresAtIsRefused`) |
| M5 | `RowsAffected == 0` → `< 0` (no rollback) | killed (`TestTheSweepDeletesOnlyWhatIsStillExpiredWhenItDeletes`) |
| M8 | `SearchWithFacets` facets → nil | killed (`TestTheFacadeCanAggregate`) |
| M9 | dead condition in `UpdateKnowledge` removed | **survived** (equivalent) |
| M10 | `recordOnlyKey` bypassed in `knowledge_tx.go` chunks | killed (`TestTheExpiryIsKeptOnTheRecordOnly`) |
| M11 | `Query` post-gate removed | **survived** (whole package) |
| M14 | PPR `loadChunks` gate emptied | killed (3 tests) |
| M19 | `report.Memories = nil` on partial failure removed | killed (`TestASweepStopsAtTheFirstFailure`) |
| M20 | read gate fails closed instead of advisory | killed (`TestAnUnreadableBrainIsNotAnEmptyBrain`) |
| M21 | `<= ''` branch removed from `visiblePredicate` | **survived** (whole package) |
| M22 | document nodes not dropped, only chunks | killed (2 tests) |
| M23 | dry run lists no documents | killed (3 tests) |

## Main vs head (same input on both)

| Scenario | main | head | Named? |
|---|---|---|---|
| S1 `SearchGraphRAG`, `plan.document_ids=[low-ranked]`, `top_k=2` | 0 chunks | 1 chunk | no (L1) |
| S2 `ingest_document`, metadata `_expires_at=2000…` | stored, served | dropped, served | yes |
| S3 `ingest_document`, metadata `expires_at=2000…` | stored, served | stored, served | same |
| S4 `SearchWithFacets` `top_k=3`, no expiry | 3 hits | 3 hits | same |
| S5 `Aggregate` count, no expiry | 14 | 14 | same |
| S6 `DeleteKnowledge` twice | ok, then not found | ok, then not found | same |
| S7 `SearchText` top 3, no expiry | 3 | 3 | same |

## Tests: sleeps, clock and environment

- No new test sleeps. Tests take the instant as an argument or put expiries an hour in the past or future.
- `sweep_expired_test.go` sets or clears every variable the command and `openBrainDB` read: `CORTEXDB_PATH`, `CORTEXDB_REMOTE`, `CORTEXDB_GRPC_TOKEN`, `HOME`, and the embedder, LLM and recall-floor variables via `withoutTheMachinesModels`. The remote tests use an in-process server on 127.0.0.1:0.
- Pre-existing, not from this PR: `CORTEXDB_LEXICAL_DIM` (`graphrag_tool_helpers.go:25`) is read by the library at ingest, and no test clears it.
