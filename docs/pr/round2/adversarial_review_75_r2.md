# Relecture adverse de #75, tour 2 partiel (22 octobre 2026)

Tête `dcdccbf` (`gc-08-content-hash`), base `0328e0a` (v2.136.0). Tour partiel, seulement la différence `ebe5a42..dcdccbf` (correctifs du tour 1), selon la règle acceptée par Lauro le 21 octobre 2026. Agent `upstream-reviewer`, kit `adv-75-r2` avec `interdiff.diff`, invite réduite. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 32, contexte 18901 -> 105028 tokens ; écriture cache 105026, lecture cache 2173989, entrée neuve 64 ; équivalent entrée (lecture x0,1, écriture x1,25) : 348681 ; outils : {'Bash': 31, 'SubagentHandback': 1}. Durée 7 min 32 s.

---

# Partial review of #75, ebe5a42 to dcdccbf (base 0328e0a, v2.136.0)

This was read-only. Every run used the canary environment, through `both.sh`, `mutone.py` or the same environment set by hand. I removed all probe files. `main`, `head` and `mut` match the export again (I checked with `diff -rq`). The worktree is clean.

The round-1 points on the typed routes are fixed, and the concurrency test is now reliable. I found one new regression of the same kind as round-1 point 1, on the tool route, and the body states it the wrong way round. I also found two rules the interdiff relies on that no test holds.

## 1. Round-1 findings

| # | Closed? | How I checked |
| :- | :- | :- |
| 1 (read-back record refused on the typed routes) | **Typed routes: closed. Tool route: not closed, see A.** | REST probe on both trees. A confined key writes `d1` with `ingest_document`, reads it with `knowledge_get`, then sends the same metadata to `POST /v1/knowledge`. Main answers 200 and head answers 200. My mutations A, B and F are killed by `TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces`. |
| 2 (the schema promised dedupe across collections) | Closed | `graphrag_tool_defs.go:40` and `graphrag_tool_types.go:58-61` now say "of the same collection". I grepped `pkg`, `cmd` and `docs` for other wordings and found none. |
| 3 (concurrency test depended on luck) | Closed | I re-ran my round-1 mutation M1 (the lookup in its own transaction before the write transaction) 5 times: **killed 5 of 5**, against 2 of 5 before. On head, the new tests pass with `-race -count=5` in 17.7 s. |
| 4 (survivors M3, M4, M8 to M12) | Closed, by reading (I did not replay the author's mutations) | `TestUnchangedNeedsEveryChunkRow` deletes 1 row of 3 (M4). `TestDedupeNamesTheSmallestHolder` inserts in the order b, a, c, so taking the first row inserted or `MAX` both fail (M8). The producer is asserted exactly (M9). Chunk sizes 4 and 5 differ from the default 120 (M10, M11). The document node's hash is checked (M12). `sameText` is removed (M3). |
| 5 (weak `strings.Contains` asserts) | Closed | `hasCount` splits `graphCounts` output on `,`. Every converted call now matches exactly. |
| 6 (`chunk_node_ids: null` on `duplicate_of`) | **Not addressed, and the reply does not mention it** | The interdiff does not change the response. |
| 7 (reads before `sameIngest`) | Closed in code, but the body text is stale (see F) | `graphrag_tool_ingest.go:71-91`: `storedChunks` now runs only after `sameIngest`. |
| 8 (proof document stale) | Closed | Jeu `75` in `mutate_round2.py` (lines 262-308) has 12 entries, including the index one, and no `sameText` entry. The round-2 table in `failing_first.md` is kept as history, and the fourth pass explains the change. |

## 2. Rules the interdiff adds

### A. Blocker: the tool route now refuses documents written by `ingest_document`, and main accepts them
- **Rule.** The marks are exempted from the allowlist in `knowledgeToolArgs` (`policy.go:244-246`) and in `authorizeTypedArgs`, for `SaveKnowledge` and `UpdateKnowledge` only (`ownership.go:166-197`).
- **Class.** Every confined-key write that can carry a document's metadata read back:
  - typed `SaveKnowledge` and `UpdateKnowledge`, on gRPC and REST
  - typed `InsertGraphDocument`
  - tool route (`CallTool` and `POST /v1/tools/<name>`): `knowledge_save`, `knowledge_update`, `ingest_document`
- **What the body says.** Point 6: "The tool route keeps the refusal, as it does for `written_by`."
- **Why that is wrong.** It holds only for signed records. `IngestDocument` never calls `stampWriter`: the only callers are `knowledge_api.go:38,120` and `memory_api.go`. So a document written by `ingest_document`, by any key or by none, has no `written_by`, and main does not refuse it on the tool route.
- **Probe on REST, both trees.** A confined key ingests `d1` with `{team: a}` through `/v1/tools/ingest_document`. It reads `d1` back through `/v1/tools/knowledge_get`. Then it sends the metadata it read with new content:

| resend | main | head |
| :- | :- | :- |
| `/v1/tools/ingest_document` | 200 | **403** "takes no metadata._chunk_overlap from a confined key" |
| `/v1/tools/knowledge_update` | 200 | **403**, same message |
| `POST /v1/knowledge` (typed) | 200 | 200 |

- **gRPC.** `CallTool` goes through the same allowlist, and `EngineDocumentKeys` is used only in `authorizeTypedArgs`. So I expect the same 403 there, but I read the code and did not run a probe.
- **Why it matters.** For an MCP agent this is the main path, and `ingest_document` is the sync job's tool.
- **Fix.** Both facade paths overwrite all three marks: `IngestDocument` always calls `stampIngestMarks`, and `UpdateKnowledge` stamps whenever metadata is sent. So the tool route can hand them over as the typed routes do. Add a test on both ports with a record written by `ingest_document`. Otherwise, name this regression in the body.

### B. Likely comment: the port exemption is safe only if the facade overwrites all three marks, and two of them have no test
- **Claim.** The `EngineDocumentKeys` comment (`graphrag_tool_ingest.go:241-245`) says every write "replaces a caller's copy of them". The port tests forge only `_content_sha256`: `confined_ingest_marks_test.go` on REST line ~49 and on gRPC line ~57.
- **E (survived, in `pkg/cortexdb` and in `pkg/rpcserver`).** After `UpdateKnowledge`'s `stampIngestMarks` (`knowledge_api.go:147`), keep the caller's `_chunk_size`.
- **G (survived, in `pkg/cortexdb` and in `pkg/httpapi`).** In `SaveKnowledge` (`knowledge_api.go:39`), keep the caller's `_chunk_overlap`.
- **Effect.** Under either mutant, a confined key can now write a forged engine mark through the typed routes, and the allowlist used to stop exactly that.
- **Fix.** Forge all three marks in the port tests, or loop over `EngineDocumentKeys()` there.

### C. Likely comment: with `sameText` removed, "same chunking" has no test when the chunk count stays the same
- **Rule.** `sameIngest` is now the only place the chunking marks are compared (`graphrag_tool_ingest.go:278-292`).
- **D (survived, jeu-75 filter).** `sameIngest` skips `_chunk_size` and `_chunk_overlap`. It is not equivalent: a probe that ingests 12 words with size 4 and overlap 0, then size 5 and overlap 1 (3 chunks both times), passes on head and **fails under D** (`unchanged` reported, while the stored chunks are the size-4 ones).
- **Why nothing catches it.** The existing chunking test changes the chunk count, so `len(chunkIDs) == len(chunks)` hides the mutant.
- **H (survived).** `SaveKnowledge` stamps overlap 0 (`resolvedIngestOptions(req.Collection, req.ChunkSize, 0)`). `TestTheMarksDescribeTheStoredChunks` never sets an overlap, so "the marks describe the chunks" holds for the size only.

### D. Nit: `EngineDocumentKeys` claims more than its test checks
- The comment says "the metadata keys the engine writes on a document". Importflow's `_table` is also an engine key, and it is not replaced by the facade, so it cannot be in this list.
- `TestEngineDocumentKeysNameEveryKeyTheEngineWrites` checks a single writer, `ingest_document`.
- My mutation C (adding `_table` to the list) is killed, so "no more" holds.
- Narrow the comment to "the keys the facade rewrites on every document write".
- For #76 and #77: `_expires_at` and the review keys are rewritten by the facade too, so they must join this list, or a read-back record is refused again there.

### E. The new check order, and the removed `sameText`
The order is a performance change, so I accept it without a test. The removal is equivalent, as the author says, but it leaves the survivor of C.

## 3. Claims in the texts
- "407 lines outside tests over 12 files, 70 removed": I counted 12 files, +407 −70 (`git diff --numstat`, test files excluded). It matches.
- "Twenty-five mutations, twenty-five killed": the body's list has 25 items, and the totals in `failing_first.md` (12 + 2 + 1 + 1 + 9) also give 25. I did not replay them.
- "The tool route keeps the refusal, as it does for `written_by`": **false** for unsigned documents (A).
- "The marks … describe the chunks `knowledge_save` and `knowledge_update` stored": holds for the size and the hash, not for the overlap (H).
- **F. Nit:** the Performance Impact paragraph still says "A re-ingest over an existing id reads … four reads". Since this round, the extra three reads happen only when `sameIngest` holds.
- **Nit:** the `## Codecov` section still contains the placeholder "The CODECOV_75 lines".
- **Nit:** `ownership.go:139-141` reflows the comment into one line of about 120 characters.

## 4. My mutations (8 distinct; E and G each run against two packages)

| # | Target | Result |
| :- | :- | :- |
| A | gRPC `UpdateKnowledge` without `marked(m)` | killed by `TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces` |
| B | REST deletes only `_content_sha256` | killed by the same test |
| C | `_table` added to `EngineDocumentKeys` | killed by `TestEngineDocumentKeysNameEveryKeyTheEngineWrites` |
| F | REST `memoryToolArgs` also exempts `_content_sha256` | killed by `TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces` |
| E | `UpdateKnowledge` keeps the caller's `_chunk_size` | **survived** (`pkg/cortexdb` and `pkg/rpcserver`) |
| G | `SaveKnowledge` keeps the caller's `_chunk_overlap` | **survived** (`pkg/cortexdb` and `pkg/httpapi`) |
| D | `sameIngest` skips the chunking marks | **survived** jeu-75 filter, not equivalent (my probe kills it) |
| H | `SaveKnowledge` stamps overlap 0 | **survived** |
| M1 (round-1 re-run) | lookup outside the write transaction | killed 5 of 5 |

Total: 4 killed, 4 survived, plus M1 at 5 of 5.

## 5. The new tests: asserts, environment, sleeps
- No `Sleep`, no `Getenv`, no `time.` in `graphrag_ingest_round3_test.go` or in the two `confined_ingest_marks_test.go`. `openIngestBrain` uses `t.TempDir()`.
- The new concurrency test starts its goroutines on a closed channel and sleeps nowhere.
- Weak asserts:
  - the port tests forge one mark of three (B);
  - `TestTheMarksDescribeTheStoredChunks` asserts overlap `"0"` with no overlap set (C/H);
  - the producer assert in `TestARetiredChunkIsRecordedAsAReingest` (`strings.Repeat`) is exact but hard to read. A loop with `!=` would say the same thing.
- Green on head with `-race`: `pkg/cortexdb` (the new tests, `-count=5`), and `pkg/httpapi` and `pkg/rpcserver` (`-run Marks`).

## Ranking
- **Blocker:** A, the tool route regression for documents written by `ingest_document` (`policy.go:244`, `ownership.go:166-197`; the body's point 6 says the opposite).
- **Likely comments:** B (E, G), C (D, H), round-1 point 6 left without an answer.
- **Nits:** D, F, the Codecov placeholder, the long comment line, the `strings.Repeat` assert.

Main file of the blocker: `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-75-r2/head/pkg/httpapi/policy.go`.
