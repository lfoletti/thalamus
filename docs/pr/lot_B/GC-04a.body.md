## Description

First of the three PRs for point 2 of discussion #66 (scoped keys and the toolbox), cut as agreed: `authz` rules unplugged, then gRPC wiring (#73), then REST wiring (#74). This one changes no behaviour on either port.

### Changes since review (20 October 2026)

The rule "the tool declares the field" is gone. In its place, an explicit allowlist, one line per tool, closed by default: a tool off the list, or an argument the line does not name, is refused to a confined key. The five ways out you found are each closed by the list and each has a regression test here, in #73 over gRPC and in #74 over REST:

1. `knowledge_memory_consolidate` (a promotion under `promotion.*` wrote another user's memories into a document): refused, like `promote_to_knowledge` and `reflect`, which read both stores without a gate of their own.
2. Recall and context packs read both stores: a two-store tool is allowed only with a gate per confined field, which `ConfineToolArgs` forces into the arguments before the call (`disable_memory=true` for a collection key, `disable_knowledge=true` for a user, scope or namespace key), whatever the caller sent.
3. `cortex_query` (its graph prefetch searched chunks of every collection): refused here. The prefetch itself is a separate fix in `query_api.go`; I will open it on its own, and add the tool to the list once it lands.
4. `extract_conversation`: `session_id` and `persist` are refused arguments for a confined key; text in, entities out.
5. `memory_save.supersedes` is a row-id field: a foreign id and a missing id both come back "not found", so the list is not an oracle.

Nested ids are named with their full path (`plan.filters.document_ids`, `entities[].chunk_ids`, `relations[].chunk_ids`, `relations[].support_edge_ids`); the chunk and edge ones are refused, because a chunk id links the shared graph to any collection and the guard does not load chunks. Rebased on v2.134.1; `CallerContext` and `WithMemoryOwners` are kept as they are, the allowlist sits in front of them.

The table below is generated from the list (`authz.ConfinedToolsTable`) and a test holds the list to `ToolDefinitions()`: every tool is either on it, with every schema property classified (scope field, data argument, row id, gate or refused), or in the refused table with a reason. A new tool or a new property fails that test until it is classified.

Size: 727 lines outside tests, above the 400 you asked for. About 200 of them are the list itself, one entry per tool with its reason, and I would rather you read the list whole than in two halves.

### Problem
A key with a scope is refused every `CallTool` (gRPC) and every `POST /v1/tools/<name>` (REST), because the tool arguments are opaque JSON the policy cannot see into. The typed RPCs honour the scope row by row. The MCP stdio client in remote mode only speaks `CallTool`, so a confined key cannot be handed to a developer's machine at all: it is the unconfined key or nothing.

### Fix
Four rules, in `pkg/authz` (`allowlist.go`, `rows.go`), that let a confined key call a tool and refuse otherwise. Nothing calls them yet: `AuthorizeRows` still refuses a confined `CallTool`, so a server that forgets to route through the new check gets a denial, not an open door.

1. **`Key.AuthorizeToolArgs`: the tool is on the allowlist, and every top-level argument is on its line** (scope field, data argument, row id or gate). An argument the line does not name is refused with the reason, as is a path the line refuses (`entities[].chunk_ids`). Every spelling counts (`COLLECTION`, which `encoding/json` reads as `collection`).
2. **The confined field is one the tool honours**, set, equal to one of the key's values, in every nested copy (a plan's `filters.collection`).
3. **A tool that reads both stores is gated.** `Key.ConfineToolArgs` sets the gate the key needs in the arguments, and `AuthorizeToolArgs` refuses a call whose gate is off; a port calls the first and hands the tool what comes back.
4. **`Scope.GuardToolRows`: a row the call names by id already lies within the scope.** The id paths come off the line (`knowledge_id`, `memory_id`, `memory_ids`, `supersedes`, `plan.filters.document_ids`, `document_id`), the row each one names is loaded, and one outside the scope is `ErrWithheld`, for the ports to word like a missing row. A row that does not exist yet passes: the write creates it inside the scope rule 2 forced. The row rules (`Scope.PermitsKnowledge`, `Scope.PermitsMemory`) are the ones the typed RPCs use in #73.

### Tools a confined key may call

| Tool | Honours | Touches | Row ids it names | Gate | Refused arguments |
|---|---|---|---|---|---|
| `aggregate_metadata` | collection | knowledge |  |  |  |
| `extract_conversation` | collection | knowledge |  |  | `persist`, `session_id` |
| `ingest_document` | collection | knowledge | `document_id` (knowledge) |  |  |
| `knowledge_delete` | (by id only) | knowledge | `knowledge_id` (knowledge) |  |  |
| `knowledge_get` | (by id only) | knowledge | `knowledge_id` (knowledge) |  |  |
| `knowledge_memory_build_context_pack` | user_id, scope, namespace, collection | knowledge and memory | `plan.filters.document_ids` (knowledge) | collection → `disable_memory=true`, user_id → `disable_knowledge=true`, scope → `disable_knowledge=true`, namespace → `disable_knowledge=true` |  |
| `knowledge_memory_recall` | user_id, scope, namespace, collection | knowledge and memory | `plan.filters.document_ids` (knowledge) | collection → `disable_memory=true`, user_id → `disable_knowledge=true`, scope → `disable_knowledge=true`, namespace → `disable_knowledge=true` |  |
| `knowledge_memory_remember` | user_id, scope, namespace | memory | `memory_id` (memory) |  |  |
| `knowledge_save` | collection | knowledge | `knowledge_id` (knowledge) |  | `entities[].chunk_ids`, `relations[].chunk_ids`, `relations[].support_edge_ids` |
| `knowledge_search` | collection | knowledge | `plan.filters.document_ids` (knowledge) |  |  |
| `knowledge_update` | collection | knowledge | `knowledge_id` (knowledge) |  | `entities[].chunk_ids`, `relations[].chunk_ids`, `relations[].support_edge_ids` |
| `memory_delete` | (by id only) | memory | `memory_id` (memory) |  |  |
| `memory_get` | (by id only) | memory | `memory_id` (memory), `memory_ids` (memory) |  |  |
| `memory_save` | user_id, scope, namespace | memory | `memory_id` (memory), `supersedes` (memory) |  | `entities[].chunk_ids`, `relations[].chunk_ids`, `relations[].support_edge_ids` |
| `memory_search` | user_id, scope, namespace | memory | `plan.filters.document_ids` (knowledge) |  |  |
| `memory_update` | (by id only) | memory | `memory_id` (memory) |  |  |
| `representative_records` | collection | knowledge |  |  |  |
| `search_graphrag_lexical` | collection | knowledge | `plan.filters.document_ids` (knowledge) |  |  |
| `search_text` | collection | knowledge | `plan.filters.document_ids` (knowledge) |  |  |
| `search_vector_range` | collection | knowledge |  |  |  |

Refused to a confined key:

| Tool | Why |
|---|---|
| `apply_inference` | reads or writes the shared graph, which no scope field confines |
| `build_context` | names chunks, whose document the guard does not load |
| `changes_since` | feeds changes of every store and every collection |
| `contract_needs_attention` | reads or writes the shared graph, which no scope field confines |
| `contract_tally` | reads or writes the shared graph, which no scope field confines |
| `cortex_query` | its graph prefetch searches chunks of every collection |
| `decision_chain` | reads or writes the shared graph, which no scope field confines |
| `decision_precedents` | reads or writes the shared graph, which no scope field confines |
| `decision_record` | reads or writes the shared graph, which no scope field confines |
| `delete_document_graph` | reads or writes the shared graph, which no scope field confines |
| `delete_entities` | reads or writes the shared graph, which no scope field confines |
| `disambiguate_mentions` | reads or writes the shared graph, which no scope field confines |
| `execution_run_finish` | reads or writes the shared graph, which no scope field confines |
| `execution_run_get` | reads or writes the shared graph, which no scope field confines |
| `execution_run_replay` | reads or writes the shared graph, which no scope field confines |
| `execution_run_start` | reads or writes the shared graph, which no scope field confines |
| `execution_runs_list` | reads or writes the shared graph, which no scope field confines |
| `execution_step_begin` | reads or writes the shared graph, which no scope field confines |
| `execution_step_end` | reads or writes the shared graph, which no scope field confines |
| `execution_step_lineage` | reads or writes the shared graph, which no scope field confines |
| `execution_step_record` | reads or writes the shared graph, which no scope field confines |
| `expand_graph` | reads or writes the shared graph, which no scope field confines |
| `fact_provenance` | reads or writes the shared graph, which no scope field confines |
| `find_nodes` | reads or writes the shared graph, which no scope field confines |
| `get_chunks` | names chunks, whose document the guard does not load |
| `get_nodes` | reads or writes the shared graph, which no scope field confines |
| `graph_cypher_query` | reads or writes the shared graph, which no scope field confines |
| `graph_diff` | reads or writes the shared graph, which no scope field confines |
| `graph_health` | reports on every store |
| `graph_list_all` | reads or writes the shared graph, which no scope field confines |
| `graph_property_values` | reads or writes the shared graph, which no scope field confines |
| `graph_schema` | reads or writes the shared graph, which no scope field confines |
| `graph_snapshot` | reads or writes the shared graph, which no scope field confines |
| `graph_statistics` | reads or writes the shared graph, which no scope field confines |
| `inference_explain` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_delete` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_export` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_find` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_import` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_infer_explain` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_infer_explain_match` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_infer_refresh` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_infer_summary` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_namespace_list` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_namespace_upsert` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_query` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_shacl_rules` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_shacl_validate` | reads or writes the shared graph, which no scope field confines |
| `knowledge_graph_upsert` | reads or writes the shared graph, which no scope field confines |
| `knowledge_memory_consolidate` | reads memories and writes a document, under promotion.*, in one call |
| `knowledge_memory_expand_entity_context` | returns chunk text of any collection |
| `knowledge_memory_neighbors` | reads or writes the shared graph, which no scope field confines |
| `knowledge_memory_promote_to_knowledge` | reads memories and writes a document in one call |
| `knowledge_memory_reflect` | reads both stores through recall.* with no gate of its own |
| `knowledge_memory_shortest_path` | reads or writes the shared graph, which no scope field confines |
| `memory_list_all` | lists every bucket |
| `object_set_resolve` | reads or writes the shared graph, which no scope field confines |
| `ontology_action_apply` | reads or writes the shared graph, which no scope field confines |
| `ontology_action_list` | reads or writes the shared graph, which no scope field confines |
| `ontology_delete` | reads or writes the shared graph, which no scope field confines |
| `ontology_diff` | reads or writes the shared graph, which no scope field confines |
| `ontology_draft` | reads or writes the shared graph, which no scope field confines |
| `ontology_get` | reads or writes the shared graph, which no scope field confines |
| `ontology_list` | reads or writes the shared graph, which no scope field confines |
| `ontology_save` | reads or writes the shared graph, which no scope field confines |
| `predict_graph_edges` | reads or writes the shared graph, which no scope field confines |
| `rank_graph_nodes` | reads or writes the shared graph, which no scope field confines |
| `rules_apply` | reads or writes the shared graph, which no scope field confines |
| `rules_delete` | reads or writes the shared graph, which no scope field confines |
| `rules_list` | reads or writes the shared graph, which no scope field confines |
| `rules_save` | reads or writes the shared graph, which no scope field confines |
| `search_chunks_by_entities` | searches chunks of every collection |
| `search_paths` | walks the shared graph and may return chunk text of any collection |
| `uncited_facts` | reads or writes the shared graph, which no scope field confines |
| `upsert_entities` | reads or writes the shared graph, which no scope field confines |
| `upsert_relations` | reads or writes the shared graph, which no scope field confines |
| `vacuum_graph` | reads or writes the shared graph, which no scope field confines |
| `vector_dimension_repair` | rewrites every store |
| `verify_claims` | reads or writes the shared graph, which no scope field confines |

### Tests
Each rule has a test that fails without it.
- `pkg/authz/allowlist_test.go`: the list is complete against `ToolDefinitions()` (every tool, every property, gates only on two-store tools and one per honoured field); a tool off the list and an argument off the line are refused with their reason; a two-store tool is gated, the gate is forced and an off gate is refused; the scope rule holds on the list (unset, wrong, nested, other spelling, depth bound, nil arguments); the table names every tool and every id field.
- `pkg/authz/rows_test.go`: the ids are read off the allowlisted paths (`supersedes`, `memory_ids`, `plan.filters.document_ids`); a new row passes; a row in another collection, or another user's memory, named alone or in a list, is withheld without naming the other collection; a collection-confined key reaches no memory row and a user-confined key no knowledge row; the guard fails closed on non-object arguments, on an unknown tool and on a read that fails.
- `go vet` and `go test -race ./...` pass.

**Failing first.** The rules are new functions, so the proof is per rule: disable the rule, keep the package compiling, and a test must fail. Fifteen such mutations, fifteen killed; the automatic pass over the diff (42 mutants) leaves none alive. Examples: admitting a tool off the list fails `TestAConfinedKeyIsRefusedAToolOffTheAllowlist`; `&& false` on the unknown-argument check or on the refused-path check fails `TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover`; a gate not forced, or not required, fails `TestAToolThatReadsBothStoresIsGatedForAConfinedKey`; `supersedes` or `plan.filters.document_ids` dropped from a line fails the completeness test and `TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths`; a tool left neither allowed nor refused fails the completeness test. Every line of the patch has a test.

Part of a small series on running one CortexDB server for several teams (see discussion #66, item GC-04). The gRPC and REST wiring follow as two separate PRs.

## Type of Change
- [x] New feature (non-breaking change which adds functionality)

## Testing
- [x] Unit tests pass
- [x] New tests have been added

## Performance Impact
None: nothing calls the new functions yet.

## Breaking Changes
None.
