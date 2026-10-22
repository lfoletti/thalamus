# Relecture adverse de #76, tour 2 partiel (22 octobre 2026)

Tête `8cfbc13` (`gc-05-expires-at`), base `879d864` (#75). Tour partiel, seulement la différence `17983c4..8cfbc13` (correctifs du tour 1), plus la clôture des points du tour 1. Agent `upstream-reviewer`, kit `adv-76-r2` avec `interdiff.diff`, invite réduite. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 42, contexte 19077 -> 131469 tokens ; écriture cache 131753, lecture cache 3626015, entrée neuve 84 ; équivalent entrée (lecture x0,1, écriture x1,25) : 527292 ; outils : {'Bash': 42, 'SubagentHandback': 1}. Durée 17 min 51 s.

---

I reviewed #76 partially, at gc-05-expires-at 8cfbc13 against base 879d864, on the interdiff 17983c4..8cfbc13. Verdict: changes requested. B1, L1 to L5 and the nits are closed. B2 is only half closed: `search_paths` still serves the expired passage when it is asked to walk bookkeeping. The new positive check in the class test also cannot fail for the six tools that echo the query.

Paths are relative to `$K/head`, with K=/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-76-r2. All probe files are removed, and `$K/mut` is identical to `$K/head` (`diff -rq` is empty).

## 1. Findings of the previous report

- **B1, closed.** `upsertGraphRAGDocumentRecord` carries `recordOnlyKey` keys from the stored record (`graphrag_helpers.go:289-301`), and `TestInsertGraphDocumentKeepsTheStoredExpiry` holds it. I checked the class of writes to `documents`: `CreateDocument`/`UpdateDocument` in `graphrag_helpers.go` and `knowledge_helpers.go`, and the upsert in `knowledge_tx.go:879`. No other writer drops the expiry.
- **B2, half closed.**
  - Closed part: `knowledge_memory_shortest_path` now refuses a path through a node of an expired document (`brain.go:351-355`). The class test now calls it from the expired chunk and expects "no path", which can fail.
  - Closed part: `chunk_ids`, `source_chunk_id` and `document_id` are stripped from edges.
  - Still open: see N1 and N2 below.
- **L1, closed.** Every caller of `gatedSearch` now passes nil filters; I listed the ten call sites. The plan's `document_ids` are applied afterwards with `allowDocumentID` (`graphrag.go:518-526`), and `TestAPinWithNothingExpiredFiltersTheSeedsAsOnMain` covers it. The scenario in the table below gives the same result on both trees.
- **L2, closed.** The comment at `query_api.go:166` now names the source lane. `TestASourceLaneDoesNotServeAnExpiredRow` and `TestAScoredExpansionDoesNotScoreAnExpiredChunk` were added. I did not replay those mutants, as asked.
- **L3, closed.** `expiry.go:32-36` refuses a UTC year outside 0001 to 9999. My mutation E (`> 10000`) is killed.
- **L4, closed.** The typed `GetKnowledge`, `FindKnowledgeGraph`, `QuerySparql` and `ExportKnowledgeGraph` are now named in the body.
- **L5, closed.** `TestTheExpiryHoldsOnPostgres` goes through the facade, and it passed on my run under `CORTEXDB_TEST_POSTGRES`. Without that variable it skips, with a message.
- **Nits, closed**, with one caveat on the third:
  - The dead condition in `knowledge_api.go:124` is removed.
  - `knowledge_delete` on a gone record is no error again (`knowledge_helpers.go:60`, `TestAnUnconditionalCleanupOfAGoneRecordIsNoError`).
  - The six entity tools are named in the body.
  - The `<= ''` branch of `visiblePredicate` now has a test, but its twin in the gate has none (N4).

## 2. New rules and the members they miss

**N1. Blocker. `search_paths` with `include_bookkeeping: true` serves the expired passage.**
- `graphrag_paths.go:146` only cleans the edges' provenance.
- `p.Nodes` is not gated. It feeds `names` (`:155-161`) and so `from_name`, `to_name` and `sentence` (`:152`).
- Probe on the class-test fixture with `{"entity_names":["Arvid Lund"],"include_bookkeeping":true,"direction":"both","max_depth":2}`. The response contained a path `"to":"chunk:old:000"` with `"sentence":"Arvid Lund grew up by the harbour of Lund. mentions Arvid Lund."`.
- The class test never sets `include_bookkeeping`. The rule needs to cover the path's nodes, not only its edges: drop the paths through a node of an expired document, the same way `ShortestPath` does.

**N2. Likely comment. Edge ids still name the expired document or chunk.**
- `relationEdgeID` (`graphrag_tool_ingest.go:489`) puts `:doc:<id>` into the id of every relation written by `upsert_relations` with a `document_id`. `InsertGraphDocument` uses `edge:rel:<chunkID>:…` (`graphrag.go:324`).
- `withoutExpiredProvenance` keeps the edge and its id.
- Probe: one `upsert_relations` with `document_id: "old"`. `expand_graph`, `search_paths` and `knowledge_memory_shortest_path` (entity to entity) all returned `"id":"edge:relation:entity:arvid_lund:entity:lund:born_in:doc:old"`.
- The class test misses this because its fixture writes relations through `knowledge_save`, whose ids carry an index (`knowledge_tx.go:673`). This matters for the body's claim that edges no longer "name the expired chunk and document". Either drop an edge whose only provenance is expired, or say in the body that its id still names the document.
- The `InsertGraphDocument` case I read from the code only; I did not run it, because it needs an extractor.

**N3. Likely comment. The stripped edges of `knowledge_memory_shortest_path` have no test.**
- My mutation A replaces `path.Edges = edges` (`brain.go:355`) with `_ = edges`. It survives the whole `pkg/cortexdb` package.
- No test asks for an entity-to-entity path whose relation cites the expired chunk.

**N4. Likely comment. The gate's `> ''` (`expiry.go:80`) has no test.**
- The comment of `TestAnEmptyExpiryWrittenByHandIsNoExpiry` (`expiry_r3_test.go:165`) says "The gate reads it as no expiry, and so does the dedupe lookup". The test checks only dedupe, which uses `visiblePredicate`.
- My mutation I turns `> ''` into `>= ''`. It hides, and lets the sweep delete, a document with an empty expiry, and it survives the whole package.
- Add a search, or a dry-run sweep, on that row.

**N5. Nit. The `filters` parameter of `gatedSearch` is now dead.**
- All ten callers pass nil (`retrieval_plan.go:289`), and the doc comment still says it "keeps those the filters … allow".
- Drop the parameter, so the L1 regression cannot come back through it.

**N6. Nit. More queries per call than the body says.**
- `withoutExpiredProvenance` costs one `GetNodesBatch` and one `hiddenAmong` whenever an edge carries provenance.
- `search_paths` calls it once per path (`graphrag_paths.go:146`), so up to 2×`max_paths` extra queries, 20 at the default. `expand_graph` goes from one gate query to three.
- The Performance Impact section says "One indexed query per read call". That is now not true for these two tools. I computed this from the code; I did not measure it.

## 3. Claims and my mutations

The claims check out:
- Every test named in the new body text exists. The new tests pass on head, with PostgreSQL included.
- The gRPC typed writes drop a forged `_expires_at` on a record that has no expiry. I probed `SaveKnowledge` and `UpdateKnowledge`; both left the stored expiry empty.
- One claim does not hold: "checks that every gated tool returns the live document, so a tool that returns nothing fails it" (body and reply, point 5). The check at `expiry_class_test.go:169` accepts "Uppsala" anywhere in the body, and the query in the arguments contains "Uppsala".
  - `cortex_query`, `knowledge_search`, `search_text`, `search_graphrag_lexical`, `knowledge_memory_recall` and `knowledge_memory_build_context_pack` echo the query.
  - Mutation B makes `search_text` return no chunk. It survives `TestEveryReadToolIsGatedOrNamed`; other tests in the package do kill it (29 of them).
  - Fix: check for `liveChunk` or the live passage text only.
- Nit: in the class test, `expiry_class_test.go:145` (`HasPrefix(how, "gated: ")`) is dead, since `:142` already skipped everything that is not exactly "gated".

My mutations: 9, none from the author's list. B and I failed to build on the first try and were rewritten; each is counted once. Result: 5 killed, 4 survived.

| # | Mutation | Result |
|---|---|---|
| A | `brain.go:355` `path.Edges = edges` → `_ = edges` | **survived** (whole package), N3 |
| B | `search_text` tool returns `chunks[:0]` | **survived** the class test alone (killed by 29 other tests), weak positive check |
| C | `document_id` of a hidden document not deleted | killed (`TestEveryReadToolIsGatedOrNamed`, `TestAnEdgeDoesNotNameAnExpiredChunkOrDocument`) |
| D | the chunks' documents not added to the `hiddenAmong` lookup | killed (`TestAnEdgeDoesNotNameAnExpiredChunkOrDocument`) |
| E | year bound `> 9999` → `> 10000` | killed (`TestAnExpiryOutsideFourDigitYearsIsRefused`) |
| F | `chunk_ids` of the copy left unfiltered | killed (2 tests) |
| G | `ShortestPath` refuses only when every node is dropped | killed (`TestEveryReadToolIsGatedOrNamed`) |
| H | stored record-only key carried only when the new metadata is nil | **survived**, equivalent: `InsertGraphDocument` never passes metadata |
| I | gate `> ''` → `>= ''` | **survived** (whole package), N4 |

## 4. Main vs head, on the inputs the interdiff changes (no expired document)

| Scenario | main | head |
|---|---|---|
| `SearchGraphRAG`, `top_k=2`, pin `[d11]` (8 runs) | 0 chunks ×8 | 0 chunks ×8 |
| same, pin `[d00,d05,d11]` (8 runs) | `[]` ×5, `[d00]` ×3 | `[]` ×2, `[d00]` ×6 |
| same, no pin | ties in varying order | ties in varying order |
| `InsertGraphDocument` over a saved record with `team` metadata | keys `[]` | keys `[]` |
| `DeleteKnowledge` twice; cleanup of a gone id | ok, then not found; nil | same |
| `shortest_path` chunk→entity, entity→entity, doc→entity | three bodies | identical |
| `search_paths`, `expand_graph` (with an `upsert_relations` edge) | bodies | identical |

The pinned rows differ only by order among tied scores: the 11 copies have identical vectors, and both trees reorder them across runs. The pin and the InsertGraphDocument rows show no unnamed difference. The visible differences are the expiry ones (N1 and N2), and the body does not name them.

## 5. The new tests

- No new test sleeps.
- `TestTheExpiryHoldsOnPostgres` reads `CORTEXDB_TEST_POSTGRES` by design and skips without it. It calls `time.Now()` against an expiry an hour in the past, which is safe.
- No new test reads the server, the embedder or the models of the machine; they all pass under the canary.
- Weak asserts: the positive check of the class test (section 3) and the comment of `TestAnEmptyExpiryWrittenByHandIsNoExpiry` (N4).
- `TestAnEdgeDoesNotNameAnExpiredChunkOrDocument` makes its read failure deterministic by renaming the column, which is good. It writes its edges by hand, so it does not see the edge-id case of N2.
- The `_expires_at` skips added to the mark tests in `pkg/httpapi` and `pkg/rpcserver` are sound. `confined_expiry_roundtrip_test.go` covers a record that has an expiry, and my probe covers one that has none.
