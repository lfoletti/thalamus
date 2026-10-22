# Relecture adverse de #75, tour 3 partiel (22 octobre 2026)

Tête `53121d1` (`gc-08-content-hash`), base `0328e0a` (v2.136.0). Tour partiel, seulement la différence `dcdccbf..53121d1` (correctifs du tour 2), plus la clôture des points du tour 2. Agent `upstream-reviewer`, kit `adv-75-r3` avec `interdiff.diff`, invite réduite. Texte de l'agent ci-dessous, condensé (sondes et chemins retirés, constats et mutations gardés).

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 23, contexte 19010 -> 100536 tokens ; écriture cache 100534, lecture cache 1472750, entrée neuve 46 ; équivalent entrée (lecture x0,1, écriture x1,25) : 272942 ; outils : {'Bash': 23, 'Read': 1, 'SubagentHandback': 1}. Durée 7 min 05 s.

---

# Partial review of #75, dcdccbf to 53121d1 (base 0328e0a, v2.136.0)

**Summary.** Last round's blocker (A) is fixed: a document written by `ingest_document`, read back and sent again through the tool route now gets 200 on both ports, as on `main`. No new blocker. One regression of the same class remains, on the typed `InsertGraphDocument`, and the body does not name it. Three of my nine mutations survived; each survivor is one key of three left untested where its neighbours are tested.

## 1. Findings of the previous report

- A: **closed** (`rewritten: true` on the three document tools; REST and gRPC probes 200 on both trees).
- B: **closed** (both port tests forge every key of `EngineDocumentKeys()`).
- C: **mostly closed**; each chunking key alone is still untested (G2).
- D: **closed** (comment narrowed). Carry-over for #76 and #77: `_expires_at` and the review keys must join the list.
- F: **closed**.
- Round-1 point 6 (`chunk_node_ids: null`): **closed**.
- Nit, Codecov placeholder: **not closed** (body line 51, "CODECOV_75").
- Other nits: closed.

## 2. Rules the interdiff adds

**R1. `rewritten` exemption.** No missed member among the tools.

**R2. `InsertGraphDocument` refuses the exact marks** (`pkg/rpcserver/ownership.go:191-200`).
- **G1. Likely comment: a regression against main for a read-back document.** A confined key writes `d1` through `CallTool ingest_document`, reads it with `GetKnowledge`, and sends that metadata to `InsertGraphDocument`: main passes authorization, head answers `PermissionDenied ... InsertGraphDocument takes no metadata._content_sha256 from a confined key`. The body does not name it. The body also disagrees with itself: point 3 says an `InsertGraphDocument` record "has no metadata", point 6 says it "stores the metadata it is sent"; from the code, the record gets nil metadata and the sent metadata lands on the chunk rows only. Fix: drop the three keys on that port and accept them, or name the difference.

## 3. Claims and my mutations

9 distinct mutations: 6 killed, 3 survived.
- N1: `InsertGraphDocument` refusal checks only the first key: **survived** (`pkg/rpcserver`).
- N4: exemption for the first mark only: **survived** in `pkg/authz`, killed in `pkg/httpapi`.
- N5: `sameIngest` ignores `_chunk_overlap` only: **survived**, not equivalent.
- N9: `sameIngest` ignores `_chunk_size` only: **survived**.
- N2, N3, N6, N7, N8: killed.

"420 lines outside tests over 11 files, 71 removed" holds. "Thirty-four mutations": the totals add up.

- **G2. Likely comment.** The same-count test changes size and overlap together; add one case per key.
- **G3. Nit.** The `InsertGraphDocument` assert and `TestTheRewrittenMarksPassOnTheDocumentWritesOnly` forge only the first key; loop over `EngineDocumentKeys()`.

## 4. Main against head

Read-back record on `/v1/tools/ingest_document`, `knowledge_save`, `knowledge_update`, `POST /v1/knowledge`, gRPC `CallTool`: 200/OK on both. gRPC `InsertGraphDocument`: passes on main, **PermissionDenied** on head (G1). Signed record re-sent: 403 `written_by` on both. Overlap-only re-ingest: untested (G2).

## 5. The new tests

No sleeps, no environment, temp dirs. Weak asserts: G2, G3.

## Ranking
- Blocker: none.
- Likely comments: G1, G2.
- Nits: G3, the Codecov placeholder, the #76/#77 carry-over.
