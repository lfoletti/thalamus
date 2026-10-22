# Relecture adverse de #76, tour 3 partiel (22 octobre 2026)

Tête `e4c6cf0` (`gc-05-expires-at`), base `53121d1` (#75). Tour partiel, seulement la différence `8cfbc13..e4c6cf0` (correctifs du tour 2), plus la clôture des points du tour 2. Agent `upstream-reviewer`, kit `adv-76-r3` avec `interdiff.diff`, invite réduite. Texte de l'agent ci-dessous, condensé (sondes et chemins retirés, constats et mutations gardés).

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 21, contexte 19112 -> 78271 tokens ; écriture cache 78269, lecture cache 1195624, entrée neuve 42 ; équivalent entrée (lecture x0,1, écriture x1,25) : 217398 ; outils : {'Bash': 23, 'SubagentHandback': 1}. Durée 7 min 03 s.

---

## Partial review of #76, round 3: gc-05-expires-at e4c6cf0 against base 53121d1, on interdiff 8cfbc13..e4c6cf0

**Verdict: comment, nothing blocks.** Every finding of round 2 is closed, and each closure has a test that a mutation of mine makes fail. The new rule ("an edge only expired documents back is left out") is applied to the edges a tool returns. It is not applied to the walk that produced them. That gives two visible effects (L1, L2), and the reply and the body claim more than the code does (L3). Two branches of the new keep/drop condition have no test (L4).

### 1. Findings of round 2

- **N1, closed.** `SearchPaths` now gates the nodes and edges of every path in one call (`graphrag_paths.go:148-170`). It drops a path when one of its edges is left out. The class test now sets `include_bookkeeping` and `direction: both`. My mutation M6 (no nodes passed to the gate) is killed by `TestEveryReadToolIsGatedOrNamed` and by `TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut`.
- **N2, closed.** `withoutExpiredProvenance` drops an edge that only expired provenance backs (`graphrag_tool_search.go:347-349`).
- **N3, closed.** My old mutation A (`path.Edges = edges` → `_ = edges`, `brain.go:355`), replayed as M10, is now killed.
- **N4, closed.** My old mutation I (`> ''` → `>= ''`, `expiry.go:80`), replayed as M7, is killed by `TestAnEmptyExpiryWrittenByHandIsNoExpiry`.
- **N5, closed.** The `filters` parameter of `gatedSearch` is gone; the ten call sites and the doc comment are updated.
- **N6, closed.** The gate runs once per call, not once per path, and the body says so.
- **Weak positive check, closed.** It now looks for `liveChunk` or "teaches in Uppsala", which the request does not echo.
- **Dead line, closed.**

### 2. New rules, their class, and the members they miss

Class: every caller of `withoutExpiredDocuments` or `withoutExpiredProvenance` (`expand_graph`, `get_nodes`, `knowledge_memory_neighbors`, `knowledge_memory_shortest_path`, `search_paths`). The rules filter what each tool returns. The traversal still runs over the stored graph, expired-only edges included.

**L1. Likely comment. `knowledge_memory_shortest_path` answers "no path" when a live path exists.** Arvid→Malmo returns "no path" while `search_paths` returns the live 2-hop path. `graph.ShortestPath` picks the 1-hop `born_in` edge, Rule B rejects it, and nothing searches again (`brain.go:348-354`). The same false negative applied in round 2 to a path through an expired chunk node. Fix: run the walk without the excluded nodes and edges, or retry; otherwise document that "no path" can be wrong until the sweep.

**L2. Likely comment. In `search_paths`, an expired path takes a place in `max_paths`.** `max_paths: 1` returns `"paths":[]` with no `truncated` flag; without `max_paths` the live path comes back. Fix: refill, or say in the body that it does not.

**L3. Likely comment. "Out of the walk" is claimed but not true.** `knowledge_memory_neighbors` from Arvid at depth 1 lists `entity:malmo`, adjacent only through the expired-only edge; `expand_graph` keeps the node it reached through it and drops only the edge. Fix: filter the traversal, or reword to "left out of the edges returned" and name the neighbour case.

**L4. Likely comment. Two branches of the drop condition have no test** (`graphrag_tool_search.go:347`). M12: `(doc == "" || docGone)` → `… || true` survived the whole package; an edge whose `document_id` is live but whose `chunk_ids` are all expired is dropped instead of kept. Reachable through `upsert_relations`. M13: `(src == "" || srcGone)` → `true` survived; near-equivalent, since both engine writers set `document_id` with `source_chunk_id`.

**Nit.** `expiry_r3_test.go:303` passes `path` twice to the format string.

### 3. Claims and my mutations

8 mutations, 6 killed (M3, M5, M6, M7, M10), 2 survived (M12, M13).

### 4. Main vs head

Nine scenarios with no expired document, byte-identical on both trees. With an expired document, the differences L1 to L3, none named in the body.

### 5. The new tests

No sleeps, no machine environment, safe clock use. Weak asserts: none new apart from L4.
