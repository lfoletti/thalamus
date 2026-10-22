# Relecture adverse de #77, tour 2 partiel (22 octobre 2026)

Tête `e437a25` (`gc-06-review-status`), base `2384236`. Tour partiel, seulement la différence `cc8716c..e437a25`, plus la clôture des 13 points du tour 1. Agent `upstream-reviewer`, kit `adv-77-r2`, sans PostgreSQL. Texte de l'agent condensé.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 60, contexte 19191 -> 134657 tokens écriture cache 134655, lecture cache 5080383, entrée neuve 120 équivalent entrée (lecture x0,1, écriture x1,25) : 676357 outils : {'Bash': 61, 'SubagentHandback': 1}. Durée 11 min 40 s.

Suite donnée : `failing_first.md`, section #77, « Relecture adverse r2 ». Décision de Lauro pour A : nommer la classe, sans code.

---

**Verdict:** no blocker. 9 of the 13 points closed outright, 2 in part. Open: the PPR/graph expansion survivor (point 4) and the Codecov PostgreSQL lines (point 7, the body read before its rewrite).

- **A, likely comment.** Under the gate, `find_nodes` and `disambiguate_mentions` return a hidden memory's text; `predict_graph_edges` from a visible memory returns a hidden one with its text as label; `rank_graph_nodes`, `graph_list_all`, `graph_cypher_query` return it (and `k-prop`). `expiry_class_test.go` files the first two under "entities and relations", wrong for memory nodes. Gate them, or name the exceptions; the GUIDE says "the graph walks".
- **B, likely comment.** `extract_conversation` with `persist` turns a hidden memory into a served document (`loadSessionText` reads every row of the session with no review check). The reason given for refusing promotion applies word for word.
- **Point 4 open.** `graphrag_ppr.go:770` `filters` to `nil` still survives: `TestEveryKnowledgeSearchLaneHonoursThePlanStatus` has no entities, so the expansion never reaches another document. An entity-linked probe kills it.
- **N1.** `aggregate.go:51`, the `= ''` arm of `accepted` has no test (Ma survived).
- **N2.** `review.go:59`, a proposal with `reviewed_by: "   "` has no test (Mk survived).
- **N3.** Performance Impact does not name the per-memory-node read in walks under the gate.
- **N4.** Under the gate, a node whose memory was deleted is hidden; without it, served. Not named.
- Weak asserts: the `memory_search` part of the walks test passes on an empty result; the outage walk test has no positive control (Mc kills it anyway).
- 13 mutations of the reviewer: 10 killed, 3 survived (Ma, Mf = M7, Mk).
- Main against head, gate off: every difference named, the outage with `accepted` implied by "a status filter is closed on an outage".
