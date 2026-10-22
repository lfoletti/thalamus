<!-- Relecture adverse de #75 (agent upstream-reviewer, contexte neuf), 22 octobre 2026, tête ebe5a42 sur 0328e0a. Texte de l'agent, non retouché. Coût : 53 tours, contexte 18 972 → 168 351 tokens, environ 790 000 tokens équivalent entrée. -->

# Review of #75 (gc-08-content-hash, head ebe5a42, base 0328e0a / v2.136.0)

Read-only. Everything below was run under the canary environment through the kit's scripts. PostgreSQL probes used `CORTEXDB_TEST_POSTGRES`. All probe files have been removed from `main`, `head` and `mut`.

The round-2 points now hold. A title-only re-ingest keeps the attached graph. After `delete_document_graph`, the same text rebuilds the nodes. The dedupe lookup runs in the write transaction, is scoped to the collection, has no row cap and uses the index. The marks live under `_` keys. The PR's own tests pass with `-race` on SQLite and on PostgreSQL. I found one regression that the body does not name, plus a set of claims that the tests do not hold.

## Findings

### 1. Blocker: the new `_` marks break the read-back-and-resend path that v2.136.0 keeps open for confined keys
v2.136.0 lets a confined key send back a record it read through the typed writes. The ports strip `written_by`/`updated_by` for that reason (`pkg/rpcserver/ownership.go:136-141`, "a row read back and sent again is not refused"; `pkg/httpapi/policy.go:229,265`). The allowlist refuses any metadata key that starts with `_` (`pkg/authz/allowlist.go:482`). With this PR, every record written by `ingest_document`, `knowledge_save` and `knowledge_update` carries `_content_sha256`, `_chunk_size` and `_chunk_overlap` (`knowledge_api.go:39,147`, `graphrag_tool_ingest.go:248`). So a record read back now has three keys the policy refuses.

Probe: a key confined to `notes` saves `k1`, reads it with `knowledge_get`, then sends the same record to `POST /v1/knowledge` with new content.

| step | main | head |
| :- | :- | :- |
| `knowledge_get` metadata | `{team, written_by}` | `{_chunk_overlap, _chunk_size, _content_sha256, team, written_by}` |
| `POST /v1/knowledge` with the read-back record | 200 | 403 "takes no metadata._chunk_size from a confined key" |

The gRPC side strips the same two writer keys only (`ownership.go:160-210`), so `SaveKnowledge`/`UpdateKnowledge` should refuse in the same way. I checked that by reading the code, not with a probe. Nothing in the body or the tests names this. The facade already overwrites the marks (`stampIngestMarks`), so the ports could drop them as they drop `written_by`, or the allowlist could exempt them. A test over REST and gRPC should hold whichever you choose.

### 2. Likely comment: the schema text promises a cross-collection dedupe that the code no longer does
The `dedupe` description reads "content already stored under another document_id is not ingested again" (`graphrag_tool_defs.go:40`). The `Dedupe` field comment reads "when another document already holds this exact content" (`graphrag_tool_types.go:58`). The lookup is limited to the request's collection (`knowledge_tx.go:840-844`). A copy in another collection is stored again: my probe `dedupe other collection` on head stored `b` in `c2` next to `a` in `c1`. This is the round-1 point 1 behaviour, and it is correct, but the schema an agent reads says the opposite. Say "in the same collection".

### 3. Likely comment: the concurrency test does not reliably catch the rule it is cited for
`TestTwoConcurrentDedupeIngestsStoreOneCopy` is the test cited for "the lookup runs inside the write transaction". I moved the lookup into its own transaction just before the write transaction (the "check outside the transaction" design from round 2) and ran it with mutone 5 times: **killed 2, survived 3**. The test races two goroutines and depends on scheduling. A deterministic version would hold the write turn, or the PostgreSQL lock, from the test while the second ingest starts.

### 4. Likely comment: claims with no test (my mutations, survivors)
I wrote 11 mutations of my own (M1 to M12, with no M6). M1 is the concurrency mutation of point 3. M2 first failed to build and was rewritten. Results: 3 killed and 7 survived, plus M1, which was killed in 2 of 5 runs. M5, M2 and M7 were killed. These survived:

| # | Mutation | Claim it aims at | Result |
| :- | :- | :- | :- |
| M4 | drop `&& len(chunkIDs) == len(chunks)` (`graphrag_tool_ingest.go:75`) | "Unchanged is said only when … the chunk rows … are all there" | survived. No test deletes one chunk row out of several. The only missing-rows test deletes all of them, which `storedIn == ""` already catches. |
| M10 | `knowledge_update` stamps chunk size 0 (`knowledge_api.go:147`) | "the marks describe the new [chunks]" | survived |
| M11 | `knowledge_save` stamps chunk size 0 (`knowledge_api.go:39`) | same, on `knowledge_save` | survived |
| M12 | drop `docNode.Properties[contentHashKey] = hash` (`knowledge_tx.go:251`) | "knowledge_save and knowledge_update keep the hash on the record and on the document node" | survived |
| M8 | `MIN(d.id)` → `MAX(d.id)` (`knowledge_tx.go:840`) | "answers the smallest id, so the answer does not depend on the order" | survived. The 17-holder test has one holder in the collection. |
| M9 | producer `document_reingest` → `document_cleanup` (`graph_cleanup.go:97`) | history rows name the re-ingest | survived (nit) |
| M3 | `sameText` → `true` (`graphrag_tool_ingest.go:256`) | no claim | survived, and it is equivalent: `sameIngest` compares the whole metadata, marks included, so `sameText` is dead code. Remove it or make it the only mark comparison. |

M10 and M11 do no harm today, because the chunk-count check rejects a wrong mark. That makes the marks a lie that only another check hides, so they need a test.

### 5. Nit: weak assertions
`TestAContentChangeReplacesTheChunksAndKeepsTheEntities` (`graphrag_ingest_round2_test.go:158`) uses `strings.Contains(edges, "mentions=2")` and `"has_chunk=1"`. These also match `mentions=20` and `has_chunk=10`, and the node checks above it do the same with `chunk=1`. Compare the whole string.

### 6. Nit: the duplicate response
On `duplicate_of` the response is `{"document_node_id":"","chunk_node_ids":null,…}` (head probe `dedupe json`). Use `[]` or omit the fields, and say so in the schema.

### 7. Nit: reads on every re-ingest
`storedChunks` (two reads) runs before the cheap `sameIngest` comparison (`graphrag_tool_ingest.go:69-75`). A re-ingest with other text pays for the chunk listing it then discards. Check `sameIngest` first.

### 8. Nit: the proof document is stale
In `failing_first.md` §#75, the table still lists the removed "same text treated as other text" mutation. It does not list the index mutation that `mutate_round2.py` (jeu 75, line 198) has and that the body counts. The body says fifteen. I count 11 live rows in the table plus 3 in the third pass, which is 14, plus the index row in the script, which makes 15.

## Checked and found to hold
- Diff size: I counted 10 non-test files, +385 −66. That matches the body.
- PostgreSQL index: `idx_documents_content_sha256` is created on head and not on main. `EXPLAIN` of the hash predicate with `enable_seqscan=off` shows `Bitmap Index Scan on idx_documents_content_sha256`, so the expression matches `JSONTextGuarded`. No test asserts the PostgreSQL plan; the lines are declared under Codecov.
- The advisory lock uses the one-`bigint` key space. `pkg/graph/changefeed.go:609` uses the two-`int` key space, so the two do not collide.
- On SQLite, `BeginTx` takes the in-process write turn (`core/sqlite_writelock.go:136-150`), so the in-transaction lookup is serialized inside one process. The two-process limit is named.
- Writers of chunk ids (`graphChunkNodeID`): `ingest_document`, the knowledge plan and `InsertGraphDocument`. The last one drops the marks, which is named and tested.
- The PR's tests on head: `-race` green on SQLite and PostgreSQL. No `Sleep` anywhere. The only environment read is `CORTEXDB_TEST_POSTGRES`, with a skip.

## Scenarios run on main and head (same input, `toolbox.Call` with JSON)

| Scenario | main | head | Named? |
| :- | :- | :- | :- |
| identical re-ingest | rewrite, v=2 | `unchanged`, v=1 | yes |
| same text into another collection | chunks move to `b`, v=2 | same, not unchanged | yes (same as main) |
| shorter text | 3 chunk rows and nodes kept | 1 row and 1 node | yes |
| attached graph, title only | 3 chunks, mentions 2, relation 1, Hades hit | same | yes |
| attached graph, shorter text | Hades still hits retired-on-head chunk 002 | Hades: 0 hits, relation `fears` kept, chunk_ids name a retired chunk | yes |
| attached graph, longer text | 5 chunks, graph kept | same | — |
| attached graph, `knowledge_update` title | graph rebuilt by extraction | same | yes |
| `knowledge_save` then identical `ingest_document` | rewrite, v=2 | `unchanged`, v=1, extraction graph kept | covered by "unchanged" |
| ingest, `knowledge_update` author, ingest with that author | v=3 | `unchanged`, v=2 | covered |
| caller metadata `_content_sha256`, `_table`, `content_sha256` | stored as given | `_content_sha256` replaced, `_table` and `content_sha256` kept | yes |
| `dedupe:true`, same collection | ignored, both stored | `duplicate_of: a`, nothing stored | yes |
| `dedupe:true`, other collection | both stored | both stored | yes in the body, **not in the schema** (finding 2) |
| `knowledge_save` without metadata | metadata empty | metadata = the three marks | yes ("three new metadata keys") |
| `knowledge_update` metadata only, then ingest with default chunking | rewrite, v=3 | `unchanged`, v=2 | covered |
| confined key, read-back record to `POST /v1/knowledge` | 200 | **403** | **no** (finding 1) |
| confined key, read-back metadata to `knowledge_update` and `ingest_document` tools | 403 (`written_by`) | 403 (`_…` or `written_by`) | already refused on main |
| PostgreSQL plan of the hash predicate | no index | index scan | yes |

## Counts (computed)
- 11 mutations of my own, plus 5 runs of M1: 3 killed, 7 survived (including 1 equivalent), M1 killed in 2 of 5 runs.
- 13 new test functions in the two round-2 files (11 + 2), all green on head.
- 17 main/head scenarios, 2 that differ without being named (finding 1, and the schema in finding 2).
