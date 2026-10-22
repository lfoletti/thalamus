## Description

First of the three PRs for point 2 of discussion #66 (scoped keys and the toolbox), cut as agreed: `authz` rules unplugged, then gRPC wiring (#73), then REST wiring (#74). This one wires nothing into either port. It changes two things in the facade, both named below: `memory_get` by list hides rows outside the caller's scope, which only a caller scope sets; and `memory_save` checks its `supersedes` targets before it writes, which every caller on every port sees (Breaking Changes).

### Changes since the second review (21 October 2026)

Amended in place, rebased on v2.135.0. Each point has a test written first; the hand mutations are listed in the "Failing first" section.

1. **One answer for a foreign id and a missing id, on every row-id field.** `authz.ErrWithheld` is the bare `not found`, and the row guard answers it for every single-id field of the allowlist (`memory_id`, `knowledge_id`, `document_id`, `supersedes`): the wording says nothing about whether the row exists. The ports (#73 over gRPC, #74 over REST) give a confined key that same wording for a row the tool itself did not find, so the two cases cannot be told apart from outside. `memory_save` now checks every `supersedes` target before it writes, so a missing target leaves no new memory behind, as a foreign one does: before, the new row was written and the error came after, and a `memory_get` of the new id told the two apart (`TestAMissingSupersedeTargetLeavesNoNewMemory`). One asymmetry stays, and it is an oracle: on the four tools that create the row they name (`ingest_document.document_id`, `knowledge_save.knowledge_id`, `memory_save.memory_id`, `knowledge_memory_remember.memory_id`), a foreign id is `not found` and a missing id is created inside the caller's scope, so the caller learns that a foreign id is taken. The same holds for the bucket a `session_id` names on `memory_save` and `knowledge_memory_remember`: a foreign session answers `memory bucket not found`, a new one is created. Ids are global, and the alternative is to let the write move a row out of another scope; the guard refuses instead. It learns that the id exists, nothing of the row. Between the guard's check and the tool's write there is no lock: a row created under the same id by another caller in that window is written by the tool. A list drops a foreign id as it drops a missing one: `memory_get.memory_ids` lists it under `missing`, at its place in the order asked (the facade reads the caller's scope, `cortexdb.CallerScope`, which both ports set through `Key.CallerContext`), and `plan.filters.document_ids` is data, not a row id, because the search it pins stays inside the collection and a foreign id yields nothing. `TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField` walks every row-id field the allowlist names: the guard withholds a foreign id with the bare wording and lets a missing one through to the tool (the ports' tests in #73 and #74 hold the tool's answer end to end); `TestTheRowGuardSeesForeignRowsUnderACallerScope` also lets a missing row through on a real brain, and the guard takes only the facade's not-found sentinel as a missing row, never an error whose text says "not found" (`TestTheRowGuardFailsClosedOnWhatItCannotDecide`); `TestMemoryGetByListHidesRowsOutsideTheCallerScopeAsMissing` holds the list; `TestTheRowGuardSeesForeignRowsUnderACallerScope` holds the design that the facade hides nothing by a single id on its own, so the guard is the one place that decides.
2. **A confined key declares no entity and no relation.** Walking the class behind your point (every argument that writes the shared graph), metadata was the smaller door: an entity a caller declares is a node of the shared graph, keyed by name across every collection, so declaring one rewrote its description and properties for every collection, added the key's collection to the ones it is known in (which the walk gate reads), and a relation to a missing entity answered with a foreign-key error, an oracle. So `entities` and `relations` are refused whole on `knowledge_save`, `knowledge_update` and `memory_save`, with their reason. The extraction the engine runs on the saved text stays, and it still touches an entity it finds, for every collection: it adds the key's collection to the entity's list (and does not remove it when the document is deleted), sets a type, recomputes the entity's embedding and records a new version in its history. It changes neither the entity's description nor its edges. That is `main`'s extraction, the same for every key; confining it is a change to the engine, not to this policy, and I left it out. A document's `metadata` takes none of the keys that place a chunk (`document_id`, `chunk_index`, `graph_kind`), which it is copied over; `title` is copied the same way and stays, since it is the caller's own title. A memory's metadata keeps those three, since nothing reads them there. In every `metadata` map, `written_by` and `updated_by` are refused: the facade stamps them from the key on four writes, but not on `ingest_document`, so a confined key could sign a document as another key. A key that starts with `_` is refused too: that prefix is the engine's. A memory's own `superseded_by` and `recall_count` stay writable by the key on its own rows; they change nothing outside its scope. `TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey` holds each on each tool. A filter map (`filters.*` of `aggregate_metadata`, `filter.*` of `representative_records` and `search_vector_range`) reads and writes nothing, so it may name an engine key: that is how a curator lists what waits on `_review_status` (#77). `TestAFilterMayNameAnEngineKey` holds both sides.
3. **Two more doors of the same walk.** `search_graphrag_lexical` with `return_paths` ran path retrieval over the shared graph with no collection (`attachGraphRAGPaths` calls `SearchPaths`), the walk `search_paths` does and a confined key is refused: `return_paths`, `max_paths`, `path_max_depth` and `relation_policies` (used by path retrieval only) are refused; the search itself, whose walk keeps to the collection, stays (`TestPathRetrievalIsRefusedToAConfinedKey`). `memory_get` reads `memory_ids` and ignores `memory_id` when both are set, while the guard read `memory_id`: a confined key now sends one form or the other (`TestMemoryGetTakesOneFormFromAConfinedKey`).
4. **What the guard cannot decide, it refuses.** Arguments that are `null` decode to no object and are refused like any non-object, and `ConfineToolArgs` hands them back unchanged instead of failing on them. A gate sent under two spellings must be `true` under both, since `encoding/json` may keep either. A document whose chunks are gone has no collection the facade can read (it names the default one), so the guard withholds it whatever the key (`TestTheRowGuardWithholdsADocumentItCannotPlace`). `TestAToolThatReadsBothStoresIsGatedForAConfinedKey` holds the first two.
5. **`supersedes` under a caller scope.** The facade now holds a target outside the caller's scope as not found, as a missing one, the way `memory_get` lists it, so a library caller that sets `CallerScope` gets that without a port in front (`TestASupersedeTargetOutsideTheCallerScopeIsNotFound`). `CallerScope` in the facade covers searches, bucket writes, the list form of `memory_get` and now `supersedes` targets. It does not cover a row named by id: under `CallerScope` alone, `memory_save` on another user's `memory_id`, `memory_update` and `memory_delete` still reach that row, as on `main`. The row guard of this PR is what holds those, so a library caller that wants them confined calls `GuardToolRows` itself, as the ports do in #73 and #74. A target deleted between the check and the stamp still fails after the new row is written, as on `main`; the window is the same as the guard's.
6. `plan.filters.document_ids` stays data: on a real brain, `knowledge_search`, `search_text` and `search_graphrag_lexical` pinned to another collection's document find nothing, and pinned to the key's own find it (`TestAPinToAForeignDocumentFindsNothingOnARealBrain`). This holds `main`'s behaviour; no code here changes it.
7. The comments in `allowlist.go` and `rows.go` name `CallerScope` and `Key.AuthorizeToolArgs`; `WithMemoryOwners` is gone with the rebase. The commit message describes the allowlist, the design since the first review.

Integration with #75–#77: this head classifies nothing of theirs, because their arguments are not on `main`. Whichever stack lands second adds its lines to the allowlist at the rebase, and `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys` says which.

**Failing first (second review).** Each test above fell on the tree without its code: the allowlist table still named `plan.filters.document_ids` as an id, a pinned document id was guarded, the eleven row-id fields said more than "not found", a list served a foreign row, a missing `supersedes` target left a new memory, declared entities and relations and the chunk keys were accepted, the path arguments were accepted, both forms of `memory_get` were accepted, an error whose text said "not found" passed the guard, `null` arguments made `ConfineToolArgs` panic, `written_by` passed in metadata, a gate switched off under a second spelling passed, a document without chunks passed the guard of a key confined to the default collection, and a foreign `supersedes` target under a caller scope was retired. Three tests hold rules that were already in the code but had none: `TestTheRowRulesHoldEveryConfinedField` (scope and namespace on both row rules), `TestACopyOfTheScopeFieldInsideAListIsRead`, and the named reasons in the completeness test. Thirty-seven hand mutations, thirty-seven killed: the refusal naming the row (memory, then document), `document_ids` and `memory_ids` back as ids, an id field lost from the list, declared entities allowed on `knowledge_save` and on `memory_save`, path retrieval allowed, both forms of `memory_get` allowed, the chunk keys allowed in a document's metadata or refused in a memory's, the `_` prefix accepted, the `_` prefix refused in a filter too, a text "not found" taken as a missing row, a missing row taken as a failed read, an object inside a list of ids not walked, the one-form rule left out of the table, a list without scope, an unreadable bucket taken as readable, an empty or self target taken as missing, the supersede check after the write, `null` taken as an object, `written_by` allowed, a gate read under one spelling, a gate not sent taken as on, a document without chunks placed, a foreign target accepted by the facade, a foreign target refused with another wording than a missing one, a target's unreadable bucket taken as not found, each confinable field left out of one row rule (four), a refused tool's own reason dropped, a new tool allowed instead of refused by default, a copy of the scope field inside a list not read, a list that spends the depth bound.

### Changes since review (20 October 2026)

The rule "the tool declares the field" is gone. In its place, an explicit allowlist, one line per tool, closed by default: a tool off the list, or an argument the line does not name, is refused to a confined key. The five ways out you found are each closed by the list and each has a regression test here, in #73 over gRPC and in #74 over REST:

1. `knowledge_memory_consolidate` (a promotion under `promotion.*` wrote another user's memories into a document): refused, like `promote_to_knowledge` and `reflect`, which read both stores without a gate of their own.
2. Recall and context packs read both stores: a two-store tool is allowed only with a gate per confined field, which `ConfineToolArgs` forces into the arguments before the call (`disable_memory=true` for a collection key, `disable_knowledge=true` for a user, scope or namespace key), whatever the caller sent.
3. `cortex_query` (its graph prefetch searched chunks of every collection): refused here. The prefetch itself is a separate fix in `query_api.go`; I will open it on its own, and add the tool to the list once it lands.
4. `extract_conversation`: `session_id` and `persist` are refused arguments for a confined key; text in, entities out.
5. `memory_save.supersedes` is a row-id field: a foreign id and a missing id both come back "not found", so the list is not an oracle.

Nested paths are named in full (`plan.filters.document_ids`, data since the second review). `entities` and `relations` are refused whole since the second review (point 2 above). `CallerContext` is kept as it is; the allowlist sits in front of it.

The table below is generated from the list (`authz.ConfinedToolsTable`) and a test holds the list to `ToolDefinitions()`: every tool is either on it, with every schema property classified (scope field, data argument, row id, gate or refused), or in the refused table with a reason. A new property of an allowed tool fails that test until it is classified. A new tool is refused with the default reason until someone puts it on the list, so the list fails closed; the thirteen tools refused for a reason of their own are named in the test, so a reason lost or added shows there (`TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys`).

Size: 898 lines outside tests, 12 removed, above the 400 you asked for. About 200 of them are the list itself, one entry per tool with its reason, and I would rather you read the list whole than in two halves.

### Problem
A key with a scope is refused every `CallTool` (gRPC) and every `POST /v1/tools/<name>` (REST), because the tool arguments are opaque JSON the policy cannot see into. The typed RPCs honour the scope row by row. The MCP stdio client in remote mode only speaks `CallTool`, so a confined key cannot be handed to a developer's machine at all: it is the unconfined key or nothing.

### Fix
Four rules, in `pkg/authz` (`allowlist.go`, `rows.go`), that let a confined key call a tool and refuse otherwise. No port calls them yet: `AuthorizeRows` still refuses a confined `CallTool`, so a server that forgets to route through the new check gets a denial, not an open door.

1. **`Key.AuthorizeToolArgs`: the tool is on the allowlist, and every top-level argument is on its line** (scope field, data argument, row id or gate). An argument the line does not name is refused with the reason, as is an argument the line refuses (`entities`, `return_paths`). Every spelling counts (`COLLECTION`, which `encoding/json` reads as `collection`).
2. **The confined field is one the tool honours**, set, equal to one of the key's values, in every nested copy (a plan's `filters.collection`).
3. **A tool that reads both stores is gated.** `Key.ConfineToolArgs` sets the gate the key needs in the arguments, and `AuthorizeToolArgs` refuses a call whose gate is off; a port calls the first and hands the tool what comes back.
4. **`Scope.GuardToolRows`: a row the call names by id already lies within the scope.** The id paths come off the line (`knowledge_id`, `memory_id`, `supersedes`, `document_id`), the row each one names is loaded, and one outside the scope is `ErrWithheld`, for the ports to word like a missing row. A row that does not exist yet passes: the write creates it inside the scope rule 2 forced. The row rules (`Scope.PermitsKnowledge`, `Scope.PermitsMemory`) are the ones the typed RPCs use in #73.

### Tools a confined key may call

| Tool | Honours | Touches | Row ids it names | Gate | Refused arguments |
|---|---|---|---|---|---|
| `aggregate_metadata` | collection | knowledge |  |  |  |
| `extract_conversation` | collection | knowledge |  |  | `persist`, `session_id` |
| `ingest_document` | collection | knowledge | `document_id` (knowledge) |  |  |
| `knowledge_delete` | (by id only) | knowledge | `knowledge_id` (knowledge) |  |  |
| `knowledge_get` | (by id only) | knowledge | `knowledge_id` (knowledge) |  |  |
| `knowledge_memory_build_context_pack` | user_id, scope, namespace, collection | knowledge and memory |  | collection → `disable_memory=true`, user_id → `disable_knowledge=true`, scope → `disable_knowledge=true`, namespace → `disable_knowledge=true` |  |
| `knowledge_memory_recall` | user_id, scope, namespace, collection | knowledge and memory |  | collection → `disable_memory=true`, user_id → `disable_knowledge=true`, scope → `disable_knowledge=true`, namespace → `disable_knowledge=true` |  |
| `knowledge_memory_remember` | user_id, scope, namespace | memory | `memory_id` (memory) |  |  |
| `knowledge_save` | collection | knowledge | `knowledge_id` (knowledge) |  | `entities`, `relations` |
| `knowledge_search` | collection | knowledge |  |  |  |
| `knowledge_update` | collection | knowledge | `knowledge_id` (knowledge) |  | `entities`, `relations` |
| `memory_delete` | (by id only) | memory | `memory_id` (memory) |  |  |
| `memory_get` | (by id only) | memory | `memory_id` (memory) |  | `memory_id` with `memory_ids` |
| `memory_save` | user_id, scope, namespace | memory | `memory_id` (memory), `supersedes` (memory) |  | `entities`, `relations` |
| `memory_search` | user_id, scope, namespace | memory |  |  |  |
| `memory_update` | (by id only) | memory | `memory_id` (memory) |  |  |
| `representative_records` | collection | knowledge |  |  |  |
| `search_graphrag_lexical` | collection | knowledge |  |  | `max_paths`, `path_max_depth`, `relation_policies`, `return_paths` |
| `search_text` | collection | knowledge |  |  |  |
| `search_vector_range` | collection | knowledge |  |  |  |

Refused to a confined key:

| Tool | Why |
|---|---|
| `apply_inference` | reaches state that no scope field confines |
| `build_context` | names chunks, whose document the guard does not load |
| `changes_since` | feeds changes of every store and every collection |
| `contract_needs_attention` | reaches state that no scope field confines |
| `contract_tally` | reaches state that no scope field confines |
| `cortex_query` | its graph prefetch searches chunks of every collection |
| `decision_chain` | reaches state that no scope field confines |
| `decision_precedents` | reaches state that no scope field confines |
| `decision_record` | reaches state that no scope field confines |
| `delete_document_graph` | reaches state that no scope field confines |
| `delete_entities` | reaches state that no scope field confines |
| `disambiguate_mentions` | reaches state that no scope field confines |
| `execution_run_finish` | reaches state that no scope field confines |
| `execution_run_get` | reaches state that no scope field confines |
| `execution_run_replay` | reaches state that no scope field confines |
| `execution_run_start` | reaches state that no scope field confines |
| `execution_runs_list` | reaches state that no scope field confines |
| `execution_step_begin` | reaches state that no scope field confines |
| `execution_step_end` | reaches state that no scope field confines |
| `execution_step_lineage` | reaches state that no scope field confines |
| `execution_step_record` | reaches state that no scope field confines |
| `expand_graph` | reaches state that no scope field confines |
| `fact_provenance` | reaches state that no scope field confines |
| `find_nodes` | reaches state that no scope field confines |
| `get_chunks` | names chunks, whose document the guard does not load |
| `get_nodes` | reaches state that no scope field confines |
| `graph_cypher_query` | reaches state that no scope field confines |
| `graph_diff` | reaches state that no scope field confines |
| `graph_health` | reports on every store |
| `graph_list_all` | reaches state that no scope field confines |
| `graph_property_values` | reaches state that no scope field confines |
| `graph_schema` | reaches state that no scope field confines |
| `graph_snapshot` | reaches state that no scope field confines |
| `graph_statistics` | reaches state that no scope field confines |
| `inference_explain` | reaches state that no scope field confines |
| `knowledge_graph_delete` | reaches state that no scope field confines |
| `knowledge_graph_export` | reaches state that no scope field confines |
| `knowledge_graph_find` | reaches state that no scope field confines |
| `knowledge_graph_import` | reaches state that no scope field confines |
| `knowledge_graph_infer_explain` | reaches state that no scope field confines |
| `knowledge_graph_infer_explain_match` | reaches state that no scope field confines |
| `knowledge_graph_infer_refresh` | reaches state that no scope field confines |
| `knowledge_graph_infer_summary` | reaches state that no scope field confines |
| `knowledge_graph_namespace_list` | reaches state that no scope field confines |
| `knowledge_graph_namespace_upsert` | reaches state that no scope field confines |
| `knowledge_graph_query` | reaches state that no scope field confines |
| `knowledge_graph_shacl_rules` | reaches state that no scope field confines |
| `knowledge_graph_shacl_validate` | reaches state that no scope field confines |
| `knowledge_graph_upsert` | reaches state that no scope field confines |
| `knowledge_memory_consolidate` | reads memories and writes a document, under promotion.*, in one call |
| `knowledge_memory_expand_entity_context` | returns chunk text of any collection |
| `knowledge_memory_neighbors` | reaches state that no scope field confines |
| `knowledge_memory_promote_to_knowledge` | reads memories and writes a document in one call |
| `knowledge_memory_reflect` | reads both stores through recall.* with no gate of its own |
| `knowledge_memory_shortest_path` | reaches state that no scope field confines |
| `memory_list_all` | lists every bucket |
| `object_set_resolve` | reaches state that no scope field confines |
| `ontology_action_apply` | reaches state that no scope field confines |
| `ontology_action_list` | reaches state that no scope field confines |
| `ontology_delete` | reaches state that no scope field confines |
| `ontology_diff` | reaches state that no scope field confines |
| `ontology_draft` | reaches state that no scope field confines |
| `ontology_get` | reaches state that no scope field confines |
| `ontology_list` | reaches state that no scope field confines |
| `ontology_save` | reaches state that no scope field confines |
| `predict_graph_edges` | reaches state that no scope field confines |
| `rank_graph_nodes` | reaches state that no scope field confines |
| `rules_apply` | reaches state that no scope field confines |
| `rules_delete` | reaches state that no scope field confines |
| `rules_list` | reaches state that no scope field confines |
| `rules_save` | reaches state that no scope field confines |
| `search_chunks_by_entities` | searches chunks of every collection |
| `search_paths` | walks the shared graph and may return chunk text of any collection |
| `uncited_facts` | reaches state that no scope field confines |
| `upsert_entities` | reaches state that no scope field confines |
| `upsert_relations` | reaches state that no scope field confines |
| `vacuum_graph` | reaches state that no scope field confines |
| `vector_dimension_repair` | rewrites every store |
| `verify_claims` | reaches state that no scope field confines |

### Tests
Each rule has a test that fails without it.
- `pkg/authz/allowlist_test.go`: the list is complete against `ToolDefinitions()` (every tool, every property, gates only on two-store tools and one per honoured field); a tool off the list and an argument off the line are refused with their reason; a two-store tool is gated, the gate is forced and an off gate is refused; the scope rule holds on the list (unset, wrong, nested, other spelling, depth bound, nil arguments); the table names every tool and every id field.
- `pkg/authz/rows_test.go`: the ids are read off the allowlisted paths (`supersedes` and the single-id fields; `memory_ids` and `plan.filters.document_ids` are data); a new row passes, on a fake loader and on a real brain; a row in another collection, or another user's memory, named alone or in a list, is withheld without naming the other collection; a collection-confined key reaches no memory row and a user-confined key no knowledge row; the guard fails closed on non-object arguments, on an unknown tool, on a read that fails and on an error whose text says "not found".
- `pkg/cortexdb/memory_get_scope_test.go`: `memory_get` by list under a caller scope lists a foreign id under `missing`, at its place; a missing `supersedes` target leaves no new memory.
- `go vet` and `go test -race ./...` pass.

**Failing first.** The rules are new functions, so the proof is per rule: disable the rule, keep the package compiling, and a test must fail. The mutations of this version are listed under "Failing first (second review)" above; the automatic pass over the diff (59 mutants) leaves none alive. Examples: admitting a tool off the list fails `TestAConfinedKeyIsRefusedAToolOffTheAllowlist`; `&& false` on the unknown-argument check or on the refused-path check fails `TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover`; a gate not forced, or not required, fails `TestAToolThatReadsBothStoresIsGatedForAConfinedKey`; `supersedes` dropped from a line fails the completeness test and `TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths`; a tool left neither allowed nor refused fails the completeness test. Every line of the patch has a test.

Part of a small series on running one CortexDB server for several teams (see discussion #66, item GC-04). The gRPC and REST wiring follow as two separate PRs.

## Type of Change
- [x] New feature (non-breaking change which adds functionality)

## Testing
- [x] Unit tests pass
- [x] New tests have been added

## Performance Impact
None on the ports: no port calls the new functions yet. In the facade, `memory_get` by list under a caller scope reads each id's bucket once (one `GetSession` per id), and `memory_save` with `supersedes` reads each target once before the write. Without a caller scope and without `supersedes`, nothing changes.

## Breaking Changes
None on the ports. One behaviour change in the facade for every caller: `memory_save` with a `supersedes` target that does not exist now fails before it writes, where `main` wrote the new memory, marked the targets listed before the missing one, and then failed. A target equal to the new `memory_id`, or blank, is skipped as before.

