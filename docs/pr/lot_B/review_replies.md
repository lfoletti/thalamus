# Réponses au relecteur, lot B (brouillons du 20 octobre 2026)

À poster **sur commande de Lauro**, après la poussée du lot. Un commentaire par PR, factuel, point par point. Pas de commentaire sur ses outils.

## #72 (`gc-04a-authz-tool-rules`)

Thanks for the review. Amended in place, rebased on v2.134.1.

- The schema-derived rule is replaced by an explicit allowlist, one line per tool, closed by default (`pkg/authz/allowlist.go`). A completeness test holds it to `ToolDefinitions()`: every tool is on the list with every property classified, or in the refused table with a reason. The table in the body is generated from it.
- The five ways out, each with a regression test here and over both ports (#73, #74):
  1. `knowledge_memory_consolidate`: refused to a confined key, with `promote_to_knowledge` and `reflect`.
  2. Two-store tools (`recall`, `build_context_pack`): allowed only with a gate per confined field, forced into the arguments before the call (`ConfineToolArgs`), whatever the caller sent.
  3. `cortex_query`: refused here. The prefetch fix in `query_api.go` is a separate small PR, opened in draft with this batch; the tool joins the list once it lands, argument by argument.
  4. `extract_conversation`: `session_id` and `persist` refused; text in, entities out.
  5. `memory_save.supersedes`: a row-id field; a foreign id and a missing id both answer "not found".
- Nested ids are named by full path; `entities[].chunk_ids`, `relations[].chunk_ids` and `relations[].support_edge_ids` are refused, since the guard does not load chunks.
- Size: 727 lines outside tests, above your 400. About 200 are the list itself (one entry per tool with its reason). I kept it whole so it can be read whole; say so if you would rather have it in two PRs.

## #73 (`gc-04b-grpc-confined-tools`)

Amended in place and rebased on #72's new head. `CallTool` now forces the gates (`ConfineToolArgs`), checks the arguments against the tool's allowlist line (`AuthorizeToolArgs`) and hands the tool the arguments it checked. The five cases of your review are held over gRPC in `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC`. Three wiring mutations, three killed (body).

## #74 (`gc-04c-rest-confined-tools`)

Amended in place and rebased on #72's new head.
- `db == nil` fails closed: the route refuses a confined key with 500 and says it has no database to guard rows against (`TestTheToolRouteRefusesAConfinedKeyWhenItHasNoDatabaseToGuardWith`).
- Same wiring as #73 over REST; the five cases are held in `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST`. Seven mutations, seven killed (body).

## Discussion #66

Noted, and applied to the six open PRs: fixed in place, rebased on v2.134.1, about 400 lines outside tests per PR, except #72, where the allowlist itself is the bulk, one entry per tool, and #76 and #77, which say their count. The allowlist table is in #72's body. `cortex_query`'s prefetch is its own small PR, #78, in draft. One note on the two stacks: the allowlist is held to the tool schema by a completeness test, so whichever lot lands second adds the other's new arguments to the list (`dedupe`, `expires_at`, the `review_status` filter, and the `sweep_expired` tool, refused); I will do that at the rebase.
