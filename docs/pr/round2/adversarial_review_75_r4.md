# Relecture adverse de #75, tour 4 partiel (22 octobre 2026)

Tête `025c490` (`gc-08-content-hash`), base `0328e0a` (v2.136.0). Tour partiel, seulement la différence `53121d1..025c490` (correctifs du tour 3), plus la clôture des points du tour 3. Agent `upstream-reviewer`, kit `adv-75-r4` avec `interdiff.diff`, invite réduite, sans PostgreSQL. Texte de l'agent ci-dessous, condensé (sondes et chemins retirés, constats et mutations gardés).

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 20, contexte 19019 -> 82268 tokens ; écriture cache 82266, lecture cache 1040574, entrée neuve 40 ; équivalent entrée (lecture x0,1, écriture x1,25) : 206889 ; outils : {'Bash': 20, 'SubagentHandback': 1}. Durée 4 min 40 s.

---

**Summary.** G1, G2 and G3 are closed. `InsertGraphDocument` no longer refuses the marks on the gRPC port; the facade drops a caller's copy of the three marks from the chunk metadata, and the body names that change against main. No blocker. Two likely comments: the commit message still describes the old refusal, and the new facade test does not pin that only the three marks are dropped (one mutation survived). The Codecov placeholder nit is still open.

## 1. Findings of the previous report

- G1 closed (`graphrag.go:246-249`, `ownership.go` back to main's mapping; `TestInsertGraphDocumentDropsACallersCopyOfTheMarks`, the gRPC loop). After `InsertGraphDocument`, the record metadata is `map[]` on both trees.
- G2 closed (size-only and overlap-only cases; M5 kills it).
- G3 closed (loops over `EngineDocumentKeys()`; M4 killed).
- Nit, Codecov placeholder: not closed.
- Nit, carry-over: any key #76 or #77 adds to `EngineDocumentKeys()` is also dropped by `InsertGraphDocument`; those PRs must decide whether that is wanted.

## 2. Rules the interdiff adds

- R1, drop in `InsertGraphDocument`: one caller outside generated code (`pkg/rpcserver/graphrag.go:38`); no REST route. No member missed.
  - **G4, likely comment.** The test comment says "Its other keys are kept, as on main", but it checks only `team`. Dropping every `_` key survives (M2). Add another `_` key and another spelling of a mark, and assert both kept.
- R2, typed `InsertGraphDocument` checked against the `ingest_document` line with the `rewritten` exemption: the four typed routes accept the exact marks and refuse other spellings and `_table`.

## 3. Claims and mutations

- Claims of the body and the reply hold. "Thirty-nine mutations": the subtotals add up.
- **G5, likely comment.** The commit message still ends with "InsertGraphDocument writes no mark, so it refuses them."
- 9 mutations: 6 killed, 1 survived (M2, drop every `_` key), 1 equivalent (M8, `meta &&` removed), 1 killed in three packages.

## 4. Main against head

Unconfined `InsertGraphDocument` with `_content_sha256`, `_chunk_size`, `_table`, `_CONTENT_SHA256`, `team`: main keeps all five on the chunk; head drops the two marks and keeps the rest (named in point 6, not in "Breaking Changes"). Read-back and re-ingest scenarios: same on both. Confined key with the exact marks: denied on main, allowed on head (named). Other spellings, `_table`, memory writes: denied on both.

Side observation from #72: with two wrong keys, the denial message of `memory_save` depends on map order.

## 5. The new tests

No sleeps, no environment reads. The gRPC assert is "not PERMISSION_DENIED" only, acceptable with the `_table` assert after it (M6 killed). Weak assert: G4.

## Ranking
- Blocker: none.
- Likely comments: G4, G5.
- Nits: the Codecov placeholder; "Breaking Changes: None" does not name the dropped keys; the carry-over to #76 and #77.
