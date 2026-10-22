# Inventaire du lot B (#72 à #74), second amendement, cinquième passage

Généré par `make inventory` sur la tête de #74 (`7a2ba97`, sur #72 `d805527`, poussé ; #73 `ff24c70`), classes `tools`, `ids`, `keys`, `rpcs` ; colonne « Test » remplie par règle (`fillB.py` du 21 octobre 2026, une règle par cas, chaque nom de test vérifié dans l'arbre). Les classes `reads`, `writes` et `exports` sont hors du lot B : il ne change que l'entrée de `CallTool` et de la route REST, plus deux points de la façade, `memory_get` par liste et la vérification de `supersedes`, tenus par `TestMemoryGetByListHidesRowsOutsideTheCallerScopeAsMissing` et `TestAMissingSupersedeTargetLeavesNoNewMemory`.

## Inventaire de `7a2ba97` contre `origin/main` (`*` = fichier touché par le diff)

### tools : chaque outil déclaré (97)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `aggregate_metadata` | `pkg/cortexdb/graphrag_tool_defs.go:165` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `apply_inference` | `pkg/cortexdb/inference_tooldefs.go:6` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `build_context` | `pkg/cortexdb/graphrag_tool_defs.go:411` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : names chunks, whose document the guard does not load) |
| `contract_needs_attention` | `pkg/cortexdb/contract_tooldefs.go:52` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `contract_tally` | `pkg/cortexdb/contract_tooldefs.go:47` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `cortex_query` | `pkg/cortexdb/graphrag_tool_defs.go:146` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : its graph prefetch searches chunks of every collection) |
| `decision_chain` | `pkg/cortexdb/decision_tooldefs.go:101` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `decision_precedents` | `pkg/cortexdb/decision_tooldefs.go:113` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `decision_record` | `pkg/cortexdb/decision_tooldefs.go:75` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `delete_document_graph` | `pkg/cortexdb/graphrag_tool_defs.go:85` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `delete_entities` | `pkg/cortexdb/graphrag_tool_defs.go:69` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `disambiguate_mentions` | `pkg/cortexdb/graphrag_tool_defs.go:252` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_run_finish` | `pkg/cortexdb/execution_tooldefs.go:160` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_run_get` | `pkg/cortexdb/execution_tooldefs.go:172` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_run_replay` | `pkg/cortexdb/execution_tooldefs.go:201` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_run_start` | `pkg/cortexdb/execution_tooldefs.go:126` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_runs_list` | `pkg/cortexdb/execution_tooldefs.go:182` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_step_begin` | `pkg/cortexdb/execution_tooldefs.go:139` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_step_end` | `pkg/cortexdb/execution_tooldefs.go:147` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_step_lineage` | `pkg/cortexdb/execution_tooldefs.go:191` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `execution_step_record` | `pkg/cortexdb/execution_tooldefs.go:154` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `expand_graph` | `pkg/cortexdb/graphrag_tool_defs.go:299` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `extract_conversation` | `pkg/cortexdb/graphrag_tool_defs.go:429` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `fact_provenance` | `pkg/cortexdb/graphrag_tool_defs.go:350` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `find_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:324` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `get_chunks` | `pkg/cortexdb/graphrag_tool_defs.go:336` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : names chunks, whose document the guard does not load) |
| `get_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:314` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `graph_diff` | `pkg/cortexdb/temporal_tooldefs.go:71` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `graph_health` | `pkg/cortexdb/graphrag_tool_defs.go:394` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reports on every store) |
| `graph_list_all` | `pkg/cortexdb/knowledge_memory_tooldefs.go:310` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `graph_property_values` | `pkg/cortexdb/graphrag_tool_defs.go:209` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `graph_schema` | `pkg/cortexdb/graphrag_tool_defs.go:194` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `graph_snapshot` | `pkg/cortexdb/temporal_tooldefs.go:56` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `graph_statistics` | `pkg/cortexdb/graphrag_tool_defs.go:247` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `inference_explain` | `pkg/cortexdb/rules_tooldefs.go:77` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ingest_document` | `pkg/cortexdb/graphrag_tool_defs.go:25` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `knowledge_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:90` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `knowledge_get` | `pkg/cortexdb/knowledge_memory_tooldefs.go:50` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `knowledge_graph_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:139` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_export` | `pkg/cortexdb/knowledge_memory_tooldefs.go:164` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_find` | `pkg/cortexdb/knowledge_memory_tooldefs.go:129` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_import` | `pkg/cortexdb/knowledge_memory_tooldefs.go:152` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_infer_explain` | `pkg/cortexdb/knowledge_memory_tooldefs.go:239` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_infer_explain_match` | `pkg/cortexdb/knowledge_memory_tooldefs.go:250` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_infer_refresh` | `pkg/cortexdb/knowledge_memory_tooldefs.go:217` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_infer_summary` | `pkg/cortexdb/knowledge_memory_tooldefs.go:234` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_namespace_list` | `pkg/cortexdb/knowledge_memory_tooldefs.go:113` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_namespace_upsert` | `pkg/cortexdb/knowledge_memory_tooldefs.go:101` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_query` | `pkg/cortexdb/knowledge_memory_tooldefs.go:174` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_shacl_rules` | `pkg/cortexdb/knowledge_memory_tooldefs.go:201` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_shacl_validate` | `pkg/cortexdb/knowledge_memory_tooldefs.go:191` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_graph_upsert` | `pkg/cortexdb/knowledge_memory_tooldefs.go:118` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_memory_build_context_pack` | `pkg/cortexdb/brain_tooldefs.go:65` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche), TestAToolThatReadsBothStoresIsGatedForAConfinedKey (porte forcée) |
| `knowledge_memory_consolidate` | `pkg/cortexdb/brain_tooldefs.go:231` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads memories and writes a document, under promotion.*, in one call) |
| `knowledge_memory_expand_entity_context` | `pkg/cortexdb/brain_tooldefs.go:125` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : returns chunk text of any collection) |
| `knowledge_memory_neighbors` | `pkg/cortexdb/brain_tooldefs.go:149` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_memory_promote_to_knowledge` | `pkg/cortexdb/brain_tooldefs.go:103` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads memories and writes a document in one call) |
| `knowledge_memory_recall` | `pkg/cortexdb/brain_tooldefs.go:26` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche), TestAToolThatReadsBothStoresIsGatedForAConfinedKey (porte forcée) |
| `knowledge_memory_reflect` | `pkg/cortexdb/brain_tooldefs.go:178` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads both stores through recall.* with no gate of its own) |
| `knowledge_memory_remember` | `pkg/cortexdb/brain_tooldefs.go:6` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `knowledge_memory_shortest_path` | `pkg/cortexdb/brain_tooldefs.go:165` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `knowledge_save` | `pkg/cortexdb/knowledge_memory_tooldefs.go:8` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `knowledge_search` | `pkg/cortexdb/knowledge_memory_tooldefs.go:60` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `knowledge_update` | `pkg/cortexdb/knowledge_memory_tooldefs.go:29` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `memory_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:354` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `memory_get` | `pkg/cortexdb/knowledge_memory_tooldefs.go:322` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `memory_list_all` | `pkg/cortexdb/knowledge_memory_tooldefs.go:299` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : lists every bucket) |
| `memory_save` | `pkg/cortexdb/knowledge_memory_tooldefs.go:261` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `memory_search` | `pkg/cortexdb/knowledge_memory_tooldefs.go:333` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `memory_update` | `pkg/cortexdb/knowledge_memory_tooldefs.go:284` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `object_set_resolve` | `pkg/cortexdb/ontology_tooldefs.go:56` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ontology_action_apply` | `pkg/cortexdb/ontology_tooldefs.go:75` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ontology_action_list` | `pkg/cortexdb/ontology_tooldefs.go:70` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ontology_delete` | `pkg/cortexdb/ontology_tooldefs.go:110` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ontology_diff` | `pkg/cortexdb/ontology_tooldefs.go:42` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ontology_draft` | `pkg/cortexdb/ontology_tooldefs.go:93` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ontology_get` | `pkg/cortexdb/ontology_tooldefs.go:22` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ontology_list` | `pkg/cortexdb/ontology_tooldefs.go:32` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `ontology_save` | `pkg/cortexdb/ontology_tooldefs.go:6` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `predict_graph_edges` | `pkg/cortexdb/graphrag_tool_defs.go:236` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `rank_graph_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:221` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `representative_records` | `pkg/cortexdb/graphrag_tool_defs.go:181` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `rules_apply` | `pkg/cortexdb/rules_tooldefs.go:54` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `rules_delete` | `pkg/cortexdb/rules_tooldefs.go:41` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `rules_list` | `pkg/cortexdb/rules_tooldefs.go:31` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `rules_save` | `pkg/cortexdb/rules_tooldefs.go:14` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `search_chunks_by_entities` | `pkg/cortexdb/graphrag_tool_defs.go:266` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : searches chunks of every collection) |
| `search_graphrag_lexical` | `pkg/cortexdb/graphrag_tool_defs.go:448` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `search_paths` | `pkg/cortexdb/graphrag_tool_defs.go:278` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : walks the shared graph and may return chunk text of any collection) |
| `search_text` | `pkg/cortexdb/graphrag_tool_defs.go:126` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `search_vector_range` | `pkg/cortexdb/graphrag_tool_defs.go:151` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (ligne de la liste blanche) |
| `uncited_facts` | `pkg/cortexdb/graphrag_tool_defs.go:361` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `upsert_entities` | `pkg/cortexdb/graphrag_tool_defs.go:44` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `upsert_relations` | `pkg/cortexdb/graphrag_tool_defs.go:97` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `vacuum_graph` | `pkg/cortexdb/temporal_tooldefs.go:92` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |
| `vector_dimension_repair` | `pkg/cortexdb/graphrag_tool_defs.go:482` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : rewrites every store) |
| `verify_claims` | `pkg/cortexdb/graphrag_tool_defs.go:371` |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé : reads or writes the shared graph, which no scope field confines) |

### ids : chaque propriété d'outil qui porte un identifiant (75)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `?.run_id` | `pkg/cortexdb/execution_tooldefs.go:92` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `apply_inference.document_id` | `pkg/cortexdb/inference_tooldefs.go:15` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `apply_inference.rule_id` | `pkg/cortexdb/inference_tooldefs.go:22` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `build_context.chunk_ids` | `pkg/cortexdb/graphrag_tool_defs.go:416` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `decision_record.supersedes` | `pkg/cortexdb/decision_tooldefs.go:89` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `delete_document_graph.document_id` | `pkg/cortexdb/graphrag_tool_defs.go:91` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `execution_run_finish.run_id` | `pkg/cortexdb/execution_tooldefs.go:165` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `execution_run_get.run_id` | `pkg/cortexdb/execution_tooldefs.go:177` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `execution_run_replay.run_id` | `pkg/cortexdb/execution_tooldefs.go:205` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `execution_step_lineage.step_id` | `pkg/cortexdb/execution_tooldefs.go:195` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `expand_graph.node_ids` | `pkg/cortexdb/graphrag_tool_defs.go:304` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `extract_conversation.session_id` | `pkg/cortexdb/graphrag_tool_defs.go:440` |  | TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover (argument refusé) |
| `fact_provenance.edge_id` | `pkg/cortexdb/graphrag_tool_defs.go:355` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `get_chunks.chunk_ids` | `pkg/cortexdb/graphrag_tool_defs.go:341` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `get_nodes.node_ids` | `pkg/cortexdb/graphrag_tool_defs.go:319` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `inference_explain.edge_id` | `pkg/cortexdb/rules_tooldefs.go:82` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `ingest_document.document_id` | `pkg/cortexdb/graphrag_tool_defs.go:31` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `knowledge_delete.knowledge_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:96` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `knowledge_get.knowledge_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:55` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `knowledge_graph_delete.triple_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:145` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_graph_infer_explain.triple_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:244` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_graph_infer_refresh.triple_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:226` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_build_context_pack.session_id` | `pkg/cortexdb/brain_tooldefs.go:72` |  | TestASessionIDCannotTakeAUserConfinedKeyIntoAnotherUsersSession (donnée : la façade tient le seau à la portée, CallerScope) |
| `knowledge_memory_build_context_pack.user_id` | `pkg/cortexdb/brain_tooldefs.go:71` |  | TestTheScopeRuleStillHoldsOnTheAllowlist (champ de portée, pas une ligne) |
| `knowledge_memory_consolidate.knowledge_id` | `pkg/cortexdb/brain_tooldefs.go:295` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_consolidate.memory_id` | `pkg/cortexdb/brain_tooldefs.go:281` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_consolidate.memory_ids` | `pkg/cortexdb/brain_tooldefs.go:294` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_consolidate.session_id` | `pkg/cortexdb/brain_tooldefs.go:247` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_consolidate.session_id` | `pkg/cortexdb/brain_tooldefs.go:283` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_consolidate.user_id` | `pkg/cortexdb/brain_tooldefs.go:246` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_consolidate.user_id` | `pkg/cortexdb/brain_tooldefs.go:282` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_expand_entity_context.entity_ids` | `pkg/cortexdb/brain_tooldefs.go:130` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_neighbors.node_id` | `pkg/cortexdb/brain_tooldefs.go:154` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_promote_to_knowledge.knowledge_id` | `pkg/cortexdb/brain_tooldefs.go:110` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_promote_to_knowledge.memory_ids` | `pkg/cortexdb/brain_tooldefs.go:109` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_recall.session_id` | `pkg/cortexdb/brain_tooldefs.go:33` |  | TestASessionIDCannotTakeAUserConfinedKeyIntoAnotherUsersSession (donnée : la façade tient le seau à la portée, CallerScope) |
| `knowledge_memory_recall.user_id` | `pkg/cortexdb/brain_tooldefs.go:32` |  | TestTheScopeRuleStillHoldsOnTheAllowlist (champ de portée, pas une ligne) |
| `knowledge_memory_reflect.session_id` | `pkg/cortexdb/brain_tooldefs.go:195` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_reflect.user_id` | `pkg/cortexdb/brain_tooldefs.go:194` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_remember.memory_id` | `pkg/cortexdb/brain_tooldefs.go:12` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `knowledge_memory_remember.session_id` | `pkg/cortexdb/brain_tooldefs.go:14` |  | TestASessionIDCannotTakeAUserConfinedKeyIntoAnotherUsersSession (donnée : la façade tient le seau à la portée, CallerScope) |
| `knowledge_memory_remember.user_id` | `pkg/cortexdb/brain_tooldefs.go:13` |  | TestTheScopeRuleStillHoldsOnTheAllowlist (champ de portée, pas une ligne) |
| `knowledge_memory_shortest_path.from_node_id` | `pkg/cortexdb/brain_tooldefs.go:170` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_memory_shortest_path.to_node_id` | `pkg/cortexdb/brain_tooldefs.go:171` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `knowledge_save.knowledge_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:14` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `knowledge_update.knowledge_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:35` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `memory_delete.chunk_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:377` |  | hors classe : propriété des schémas partagés `entities[]` et `relations[]` (toolEntityArraySchema, toolRelationArraySchema), attribuée ici au dernier outil du fichier ; `entities` et `relations` sont refusés en bloc à une clé confinée sur chaque outil qui les prend (TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey), `rule_id` est un nom de règle, pas une ligne |
| `memory_delete.chunk_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:394` |  | hors classe : propriété des schémas partagés `entities[]` et `relations[]` (toolEntityArraySchema, toolRelationArraySchema), attribuée ici au dernier outil du fichier ; `entities` et `relations` sont refusés en bloc à une clé confinée sur chaque outil qui les prend (TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey), `rule_id` est un nom de règle, pas une ligne |
| `memory_delete.memory_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:360` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `memory_delete.rule_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:398` |  | hors classe : propriété des schémas partagés `entities[]` et `relations[]` (toolEntityArraySchema, toolRelationArraySchema), attribuée ici au dernier outil du fichier ; `entities` et `relations` sont refusés en bloc à une clé confinée sur chaque outil qui les prend (TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey), `rule_id` est un nom de règle, pas une ligne |
| `memory_delete.support_edge_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:399` |  | hors classe : propriété des schémas partagés `entities[]` et `relations[]` (toolEntityArraySchema, toolRelationArraySchema), attribuée ici au dernier outil du fichier ; `entities` et `relations` sont refusés en bloc à une clé confinée sur chaque outil qui les prend (TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey), `rule_id` est un nom de règle, pas une ligne |
| `memory_get.memory_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:327` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `memory_get.memory_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:328` |  | TestMemoryGetByListHidesRowsOutsideTheCallerScopeAsMissing (donnée : la façade range l'étranger sous missing), TestMemoryGetTakesOneFormFromAConfinedKey (jamais avec memory_id) |
| `memory_save.memory_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:267` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `memory_save.session_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:269` |  | TestASessionIDCannotTakeAUserConfinedKeyIntoAnotherUsersSession (donnée : la façade tient le seau à la portée, CallerScope) |
| `memory_save.supersedes` | `pkg/cortexdb/knowledge_memory_tooldefs.go:277` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths, TestAMissingSupersedeTargetLeavesNoNewMemory |
| `memory_save.user_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:268` |  | TestTheScopeRuleStillHoldsOnTheAllowlist (champ de portée, pas une ligne) |
| `memory_search.session_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:340` |  | TestASessionIDCannotTakeAUserConfinedKeyIntoAnotherUsersSession (donnée : la façade tient le seau à la portée, CallerScope) |
| `memory_search.user_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:339` |  | TestTheScopeRuleStillHoldsOnTheAllowlist (champ de portée, pas une ligne) |
| `memory_update.memory_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:290` |  | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| `ontology_delete.schema_id` | `pkg/cortexdb/ontology_tooldefs.go:116` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `ontology_diff.schema_id` | `pkg/cortexdb/ontology_tooldefs.go:47` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `ontology_draft.schema_id` | `pkg/cortexdb/ontology_tooldefs.go:101` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `ontology_get.schema_id` | `pkg/cortexdb/ontology_tooldefs.go:27` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `predict_graph_edges.node_id` | `pkg/cortexdb/graphrag_tool_defs.go:241` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `rules_apply.document_id` | `pkg/cortexdb/rules_tooldefs.go:68` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `rules_apply.rule_ids` | `pkg/cortexdb/rules_tooldefs.go:62` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `rules_delete.rule_ids` | `pkg/cortexdb/rules_tooldefs.go:49` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `search_paths.node_ids` | `pkg/cortexdb/graphrag_tool_defs.go:284` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `upsert_entities.chunk_ids` | `pkg/cortexdb/graphrag_tool_defs.go:60` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `upsert_entities.document_id` | `pkg/cortexdb/graphrag_tool_defs.go:50` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `upsert_relations.chunk_ids` | `pkg/cortexdb/graphrag_tool_defs.go:113` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `upsert_relations.document_id` | `pkg/cortexdb/graphrag_tool_defs.go:103` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `upsert_relations.rule_id` | `pkg/cortexdb/graphrag_tool_defs.go:117` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |
| `upsert_relations.support_edge_ids` | `pkg/cortexdb/graphrag_tool_defs.go:118` |  | hors classe : outil refusé à une clé confinée (TestAConfinedKeyIsRefusedAToolOffTheAllowlist) |

### keys : chaque clé réservée (15)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `OntologyDecisionNoPrimaryKey = "no_primary_key_candidate"` | `pkg/cortexdb/ontology_draft.go:118` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `OntologyDecisionPrimaryKey = "primary_key_guess"` | `pkg/cortexdb/ontology_draft.go:117` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `entityCollectionsKey = "collections"` | `pkg/cortexdb/entity_collections.go:22` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `factRecordedAtKey = "recorded_at"` | `pkg/graphflow/temporal.go:35` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `factSupersededByKey = "superseded_by"` | `pkg/graphflow/temporal.go:231` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `factValidFromKey = "valid_from"` | `pkg/graphflow/temporal.go:33` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `factValidToKey = "valid_to"` | `pkg/graphflow/temporal.go:34` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `masteredKey = "mastered_at"` | `pkg/graphflow/learning.go:37` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `ontologyDraftPlaceholderKey = "name"` | `pkg/cortexdb/ontology_draft.go:199` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `owlRuleKey = "owl_has_key"` | `pkg/graph/owl_rl_rules.go:100` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `sharedKey = "legacy-token"` | `cmd/cortexdb-mcp-stdio/recall_remote.go:174` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `supersededAtKey = "superseded_at"` | `pkg/memoryflow/memory_edit.go:38` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `supersededByKey = "superseded_by"` | `pkg/memoryflow/memory_edit.go:37` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `supersedesKey = "supersedes"` | `pkg/memoryflow/memory_edit.go:39` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |
| `transcriptTurnsMetadataKey = "transcript_turns_json"` | `pkg/memoryflow/service.go:15` |  | hors classe : clé d'un autre domaine (ontologie, graphe temporel, mémoires, configuration) ; ni lue ni écrite par la liste blanche |

### rpcs : chaque RPC typé (43)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `Backup` | `proto/cortexdb/v1/admin.proto:39` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `CallTool` | `proto/cortexdb/v1/tools.proto:27` |  | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC, TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC |
| `ContractTally` | `proto/cortexdb/v1/contract.proto:124` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `DecisionChain` | `proto/cortexdb/v1/decision.proto:186` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `DeleteKnowledge` | `proto/cortexdb/v1/knowledge.proto:116` |  | TestTheTypedRPCsWithholdADocumentTheyCannotPlace (#73 : un document sans chunks, que la façade ne sait pas placer, est retenu à une clé confinée ; le reste, contrôle de `main` inchangé) |
| `DeleteKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:260` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `DeleteMemory` | `proto/cortexdb/v1/memory.proto:86` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `DeleteOntologySchema` | `proto/cortexdb/v1/graph.proto:272` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `ExplainInference` | `proto/cortexdb/v1/graph.proto:267` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `ExplainInferenceMatch` | `proto/cortexdb/v1/graph.proto:268` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `ExportKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:262` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `FindKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:259` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `GetKnowledge` | `proto/cortexdb/v1/knowledge.proto:114` |  | TestTheTypedRPCsWithholdADocumentTheyCannotPlace (#73 : un document sans chunks, que la façade ne sait pas placer, est retenu à une clé confinée ; le reste, contrôle de `main` inchangé) |
| `GetMemory` | `proto/cortexdb/v1/memory.proto:84` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `GetOntologySchema` | `proto/cortexdb/v1/graph.proto:270` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `Health` | `proto/cortexdb/v1/admin.proto:35` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `HybridSearchText` | `proto/cortexdb/v1/graphrag.proto:95` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `ImportKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:261` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `Info` | `proto/cortexdb/v1/admin.proto:36` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `InsertGraphDocument` | `proto/cortexdb/v1/graphrag.proto:90` |  | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes (#73 : arguments contrôlés contre la ligne de `ingest_document` avant la règle de ligne) |
| `InsertText` | `proto/cortexdb/v1/graphrag.proto:92` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `InsertTextBatch` | `proto/cortexdb/v1/graphrag.proto:93` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `ListNamespaces` | `proto/cortexdb/v1/graph.proto:257` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `ListOntologySchemas` | `proto/cortexdb/v1/graph.proto:271` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `ListTools` | `proto/cortexdb/v1/tools.proto:26` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `NeedsAttention` | `proto/cortexdb/v1/contract.proto:129` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `Precedents` | `proto/cortexdb/v1/decision.proto:189` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `QuerySparql` | `proto/cortexdb/v1/graph.proto:263` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `RecordDecision` | `proto/cortexdb/v1/decision.proto:184` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `RefreshInference` | `proto/cortexdb/v1/graph.proto:265` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `SaveKnowledge` | `proto/cortexdb/v1/knowledge.proto:112` |  | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes (#73 : arguments contrôlés contre la ligne de `knowledge_save` avant la règle de ligne) |
| `SaveMemory` | `proto/cortexdb/v1/memory.proto:82` |  | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes (#73 : arguments contrôlés contre la ligne de `memory_save` avant la règle de ligne) |
| `SaveOntologySchema` | `proto/cortexdb/v1/graph.proto:269` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `SearchGraphRag` | `proto/cortexdb/v1/graphrag.proto:91` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `SearchKnowledge` | `proto/cortexdb/v1/knowledge.proto:115` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `SearchMemory` | `proto/cortexdb/v1/memory.proto:85` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `SearchText` | `proto/cortexdb/v1/graphrag.proto:94` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `SummarizeInference` | `proto/cortexdb/v1/graph.proto:266` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `UpdateKnowledge` | `proto/cortexdb/v1/knowledge.proto:113` |  | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes (#73 : arguments contrôlés contre la ligne de `knowledge_update` avant la règle de ligne) |
| `UpdateMemory` | `proto/cortexdb/v1/memory.proto:83` |  | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes (#73 : arguments contrôlés contre la ligne de `memory_update` avant la règle de ligne) |
| `UpsertKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:258` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `UpsertNamespace` | `proto/cortexdb/v1/graph.proto:256` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |
| `ValidateShacl` | `proto/cortexdb/v1/graph.proto:264` |  | hors classe : RPC typé, contrôle de `main` (AuthorizeRows, règles de ligne) inchangé ; `scopePermitsMemory` et `scopePermitsKnowledge` appellent les règles exportées de #72 (scoped_keys_test.go) |

### Outils de la liste blanche absents de l'inventaire (définis hors des fichiers `*tooldefs*`)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `changes_since` | (table) |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé) |
| `graph_cypher_query` | (table) |  | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestAConfinedKeyIsRefusedAToolOffTheAllowlist (refusé) |

### Clés et arguments refusés par la liste blanche (hors motif `…Key = "…"`)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `entities`, `relations` (entiers, sur `knowledge_save`, `knowledge_update`, `memory_save`) | `pkg/authz/allowlist.go` | * | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey |
| `document_id`, `chunk_index`, `graph_kind` (métadonnées d'un document) | `pkg/authz/allowlist.go` | * | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey |
| `written_by`, `updated_by` (toute map de métadonnées) | `pkg/authz/allowlist.go` | * | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey |
| préfixe `_` | `pkg/authz/allowlist.go` | * | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey, TestAFilterMayNameAnEngineKey |
| `return_paths`, `max_paths`, `path_max_depth`, `relation_policies` | `pkg/authz/allowlist.go` | * | TestPathRetrievalIsRefusedToAConfinedKey |

### Routes REST typées d'écriture (#74)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `POST /v1/knowledge` | `pkg/httpapi/policy.go` | * | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes (ligne de `knowledge_save`), TestATypedWriteTakesTheSignatureTheFacadeReplaces (signature), TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers (clés `user_id`, `scope`, `namespace`, corps refusé par le décodeur) |
| `POST /v1/memory` | `pkg/httpapi/policy.go` | * | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes (ligne de `memory_save`), TestATypedWriteTakesTheSignatureTheFacadeReplaces (signature), TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers (clé de collection, corps refusé par le décodeur) |
