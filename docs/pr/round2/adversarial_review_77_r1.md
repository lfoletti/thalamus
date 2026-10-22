# Relecture adverse de #77, tour 1 complet (22 octobre 2026)

Tête `cc8716c` (`gc-06-review-status`), base `2384236` (#76, sur #75 `f4c5aa7`, sur v2.136.0). Invite fixe de `docs/pipeline_pr.md`, section 1 bis, sans PostgreSQL. Agent `upstream-reviewer`, kit `adv-77-r1`. Texte de l'agent ci-dessous, condensé (sondes et chemins retirés, constats, mutations et tableaux gardés).

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 52, contexte 18995 -> 158716 tokens ; écriture cache 447644, lecture cache 5147504, entrée neuve 104 ; équivalent entrée (lecture x0,1, écriture x1,25) : 1074305 ; outils : {'Bash': 50, 'Read': 2, 'SubagentHandback': 1}. Durée 27 min 38 s.

Suite donnée (décisions de Lauro du 22 octobre 2026 pour les points 1 et 6) : voir `failing_first.md`, section #77, « Relecture adverse r1 ».

---

**Verdict:** no blocker. The core rules hold under my probes and mutations. What is left: overclaims in the docs and the body, two filters with two meanings, and four survivors that are test gaps on rules the body states.

## Likely comments

1. The GUIDE and the commit message overclaim for memories in the graph walks. Gate on, a proposed memory with an entity: `knowledge_memory_neighbors`, `knowledge_memory_expand_entity_context` and `search_paths` still return its text. Recall, context pack and `memory_search` hide it.
2. The Fix section says the decision is kept "on the record, on the document node and on the chunks"; `recordOnlyKey` keeps it on the record only. Side effect: a curator cannot `group_by` the status.
3. Two status filters, two meanings: `plan.filters.review_status=accepted` keeps documents with no status; `_review_status: accepted` on the aggregates leaves them out. Neither description says so.
4. Survivor M7: `graphrag_ppr.go:770`, `hiddenAmong(ctx, filters, …)` to `nil`, survives the package. The class is every lane of `knowledge_search`; only some lanes are tested.
5. Survivor M21: a rejected memory under the gate has no test (`reviewHides` without `ReviewRejected`).
6. Promotion launders a proposed memory into served knowledge (`brain.go:138-190` reads by id and saves with no review field).
7. The Codecov section does not name this PR's own PostgreSQL lines: `store_postgres_aggregate.go` 62–71 and 76–85 (the `OnlyDocuments` branch and `pgBindEach`), plus the new index line. Check the "13 lines" against the head's report.

## Nits

8. `knowledge_memory_remember` is a sixth tool that decides (`KnowledgeMemoryRememberRequest = MemorySaveRequest`); `review_status: "bogus"` is accepted on main and refused on head, not named.
9. Survivor M25: `strings.TrimSpace(f.By)` to `f.By` in `decideReview`.
10. Survivor M8: the `errMemoryBucketWithheld` branch of the new `SaveMemory` read has no test.
11. With the gate off, the plan filter fails open on an unreadable brain (`hiddenAmong`).
12. Stale comment on `MemoryRecord.ReviewStatus`/`ReviewedBy` (old key names; "left out" true only under the gate).
13. `TestTheSchemasPromiseNothingAboutEntitiesAndRelations` and `TestAMetadataKeyNamedReviewStatusIsTheCallers` pass on the base by construction: guards, not failing-first proofs.

No added test sleeps or reads the machine's environment.

## Mutations (22 run, 18 killed, 4 survived)

M7 (PPR chunks ignore the plan status), M8 (withheld bucket), M21 (rejected memory served), M25 (reviewer not trimmed) survived. M1 to M6, M9 to M13, M16, M17, M19, M22 to M24, M26 killed.

## Main against head, gate off

| Scenario | main | head | Named? |
|:-|:-|:-|:-|
| `ingest_document` metadata `review_status=approved` | stored | stored | yes |
| `knowledge_save` metadata `_review_status=proposed` | key stored | key dropped | yes |
| `knowledge_save review_status=bogus` | ok | error | yes |
| `knowledge_save accepted`, no reviewer | ok | error | yes |
| `knowledge_search`, no filter | all four | all four | yes |
| `knowledge_search` plan `proposed` | all four | kp | yes |
| aggregate count `_review_status=accepted` | 0 | 0 (unlike the plan) | no (3) |
| aggregate `group_by _review_status` | null:4, proposed:1 | null:4 | no (2) |
| `memory_save` metadata `_review_status=proposed` | key stored | key dropped | yes |
| `knowledge_memory_remember review_status=bogus` | ok | error | no (8) |

## Head only, gate on

| Scenario | head |
|:-|:-|
| plan proposed on the five plan tools | kp |
| proposed memory, then promotion, then `knowledge_search` | promoted document served (6) |
| proposed memory with entity: neighbors, expand_entity_context, search_paths | text served (1) |
| gate off, `documents.metadata` unreadable, plan proposed | kp, ka (11) |
