# Relecture adverse de #76, tour 4 partiel (22 octobre 2026)

Tête `e100a34` (`gc-05-expires-at`), base `025c490` (#75). Tour partiel, seulement la différence `git range-diff 53121d1..e4c6cf0 025c490..e100a34` (correctifs du tour 3 et rebasage sur #75 amendée), plus la clôture des points du tour 3. Agent `upstream-reviewer`, kit `adv-76-r4`, invite réduite, sans PostgreSQL. Texte de l'agent ci-dessous, condensé (sondes et chemins retirés, constats et mutations gardés).

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 35, contexte 19237 -> 87049 tokens ; écriture cache 87047, lecture cache 1812786, entrée neuve 70 ; équivalent entrée (lecture x0,1, écriture x1,25) : 290087 ; outils : {'Bash': 34, 'SubagentHandback': 1}. Durée 7 min 03 s.

---

**Verdict: comment. Nothing blocks.** The interdiff changes no production logic in this PR's own commit: it rewords the `brain.go` comment, the test name and the `EngineDocumentKeys` doc, adds two test cases (L4), fixes the nit, and rebases onto #75's drop of a caller's marks in `InsertGraphDocument`, which now also drops `_expires_at`. The body names L1 to L3 for the edge case only; the expired chunk node case is missing (F1). One real survivor of 8 mutations (M4) shows a claim with no test (F2).

## 1. Findings of round 3

- L1, L2: closed by naming, not exact (F1).
- L3: closed (body, reply, comment, test name).
- L4: closed; my M5 is killed by `TestAnEdgeDoesNotNameAnExpiredChunkOrDocument`.
- Nit: closed.

## 2. Rules and class

- A caller's `_expires_at` is an engine key: every writer of a document record drops the caller's copy; `rewritten` covers the three tools; typed `InsertGraphDocument` maps to the `ingest_document` line. No member missed.
- `InsertGraphDocument` keeps the stored expiry on the record: verified (stored past expiry survives a plain re-insert, a caller's future value and a caller's empty value; the document stays hidden; no chunk carries the key).

## 3. Claims and mutations

8 mutations: 6 killed, M8 killed in two ports (survives `pkg/authz` by design), **M4 survived** (the record takes the caller's metadata on a new document; no test reads the record after a forged key).

## 4. Main vs head

`InsertGraphDocument` with a caller's `_expires_at`: main stores it on the chunk; head drops it and keeps the stored expiry on the record. Covered by Breaking Changes and #75's drop.

## 5. Findings

- **F1, likely comment.** The body ties the "no path" and `max_paths` effects to an expired-only relation; the same holds for a path through a chunk node of an expired document (probe with no relation edges: shortest_path Ada→Bo "no path", search_paths `max_paths: 1` empty). Say both.
- **F2, nit.** No test reads the record after a forged key is sent to `InsertGraphDocument` (M4).
- **F3, nit.** The marks tests over the transports skip `_expires_at`; authorization and the facade drop are tested apart.

New tests: no sleeps, no environment, safe clock.
