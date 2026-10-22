# Inventaire de #75 (gc-08-content-hash), rebasé sur v2.136.0

Généré par `make inventory` sur la tête de #75 contre `origin/main` (v2.136.0), classes `tools`, `reads`, `writes`, `keys`, `promises` ; colonne « Test » remplie par règle (`fill75.py` du 22 octobre 2026, chaque nom de test vérifié dans l'arbre). Les classes `ids`, `exports` et `rpcs` sont hors de #75 : il ne change ni un champ d'identifiant, ni une lecture exportée, ni un RPC typé ; le RPC `InsertGraphDocument` écrit par la fonction de même nom (classe `writes`).

## Inventaire de `28a777e` contre `origin/main` (`*` = fichier touché par le diff)

### tools : chaque outil déclaré (97)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `aggregate_metadata` | `pkg/cortexdb/graphrag_tool_defs.go:166` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `apply_inference` | `pkg/cortexdb/inference_tooldefs.go:6` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `build_context` | `pkg/cortexdb/graphrag_tool_defs.go:412` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `contract_needs_attention` | `pkg/cortexdb/contract_tooldefs.go:52` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `contract_tally` | `pkg/cortexdb/contract_tooldefs.go:47` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `cortex_query` | `pkg/cortexdb/graphrag_tool_defs.go:147` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `decision_chain` | `pkg/cortexdb/decision_tooldefs.go:101` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `decision_precedents` | `pkg/cortexdb/decision_tooldefs.go:113` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `decision_record` | `pkg/cortexdb/decision_tooldefs.go:75` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `delete_document_graph` | `pkg/cortexdb/graphrag_tool_defs.go:86` | * | TestUnchangedIsSaidOnlyWhenTheChunkNodesAreThere (après lui, le même texte reconstruit les nœuds) |
| `delete_entities` | `pkg/cortexdb/graphrag_tool_defs.go:70` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `disambiguate_mentions` | `pkg/cortexdb/graphrag_tool_defs.go:253` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_run_finish` | `pkg/cortexdb/execution_tooldefs.go:160` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_run_get` | `pkg/cortexdb/execution_tooldefs.go:172` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_run_replay` | `pkg/cortexdb/execution_tooldefs.go:201` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_run_start` | `pkg/cortexdb/execution_tooldefs.go:126` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_runs_list` | `pkg/cortexdb/execution_tooldefs.go:182` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_step_begin` | `pkg/cortexdb/execution_tooldefs.go:139` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_step_end` | `pkg/cortexdb/execution_tooldefs.go:147` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_step_lineage` | `pkg/cortexdb/execution_tooldefs.go:191` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `execution_step_record` | `pkg/cortexdb/execution_tooldefs.go:154` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `expand_graph` | `pkg/cortexdb/graphrag_tool_defs.go:300` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `extract_conversation` | `pkg/cortexdb/graphrag_tool_defs.go:430` | * | écrit par SaveKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `fact_provenance` | `pkg/cortexdb/graphrag_tool_defs.go:351` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `find_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:325` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `get_chunks` | `pkg/cortexdb/graphrag_tool_defs.go:337` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `get_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:315` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `graph_diff` | `pkg/cortexdb/temporal_tooldefs.go:71` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `graph_health` | `pkg/cortexdb/graphrag_tool_defs.go:395` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `graph_list_all` | `pkg/cortexdb/knowledge_memory_tooldefs.go:310` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `graph_property_values` | `pkg/cortexdb/graphrag_tool_defs.go:210` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `graph_schema` | `pkg/cortexdb/graphrag_tool_defs.go:195` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `graph_snapshot` | `pkg/cortexdb/temporal_tooldefs.go:56` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `graph_statistics` | `pkg/cortexdb/graphrag_tool_defs.go:248` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `inference_explain` | `pkg/cortexdb/rules_tooldefs.go:77` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ingest_document` | `pkg/cortexdb/graphrag_tool_defs.go:25` | * | TestIngestDocumentKeepsTheContentHash, TestAnIdenticalReingestWritesNothing, TestDedupeRefusesTheSameContentUnderAnotherID, TestATitleOnlyReingestKeepsTheGraphTheCallerAttached, TestAConfinedKeyDedupesWithinItsCollectionOnly |
| `knowledge_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:90` |  | hors classe : supprime l'enregistrement, donc ses marques |
| `knowledge_get` | `pkg/cortexdb/knowledge_memory_tooldefs.go:50` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:139` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_export` | `pkg/cortexdb/knowledge_memory_tooldefs.go:164` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_find` | `pkg/cortexdb/knowledge_memory_tooldefs.go:129` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_import` | `pkg/cortexdb/knowledge_memory_tooldefs.go:152` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_infer_explain` | `pkg/cortexdb/knowledge_memory_tooldefs.go:239` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_infer_explain_match` | `pkg/cortexdb/knowledge_memory_tooldefs.go:250` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_infer_refresh` | `pkg/cortexdb/knowledge_memory_tooldefs.go:217` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_infer_summary` | `pkg/cortexdb/knowledge_memory_tooldefs.go:234` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_namespace_list` | `pkg/cortexdb/knowledge_memory_tooldefs.go:113` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_namespace_upsert` | `pkg/cortexdb/knowledge_memory_tooldefs.go:101` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_query` | `pkg/cortexdb/knowledge_memory_tooldefs.go:174` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_shacl_rules` | `pkg/cortexdb/knowledge_memory_tooldefs.go:201` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_shacl_validate` | `pkg/cortexdb/knowledge_memory_tooldefs.go:191` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_graph_upsert` | `pkg/cortexdb/knowledge_memory_tooldefs.go:118` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_memory_build_context_pack` | `pkg/cortexdb/brain_tooldefs.go:65` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_memory_consolidate` | `pkg/cortexdb/brain_tooldefs.go:231` |  | écrit par SaveKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `knowledge_memory_expand_entity_context` | `pkg/cortexdb/brain_tooldefs.go:125` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_memory_neighbors` | `pkg/cortexdb/brain_tooldefs.go:149` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_memory_promote_to_knowledge` | `pkg/cortexdb/brain_tooldefs.go:103` |  | écrit par SaveKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `knowledge_memory_recall` | `pkg/cortexdb/brain_tooldefs.go:26` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_memory_reflect` | `pkg/cortexdb/brain_tooldefs.go:178` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_memory_remember` | `pkg/cortexdb/brain_tooldefs.go:6` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_memory_shortest_path` | `pkg/cortexdb/brain_tooldefs.go:165` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_save` | `pkg/cortexdb/knowledge_memory_tooldefs.go:8` |  | TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `knowledge_search` | `pkg/cortexdb/knowledge_memory_tooldefs.go:60` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `knowledge_update` | `pkg/cortexdb/knowledge_memory_tooldefs.go:29` |  | TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `memory_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:354` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `memory_get` | `pkg/cortexdb/knowledge_memory_tooldefs.go:322` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `memory_list_all` | `pkg/cortexdb/knowledge_memory_tooldefs.go:299` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `memory_save` | `pkg/cortexdb/knowledge_memory_tooldefs.go:261` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `memory_search` | `pkg/cortexdb/knowledge_memory_tooldefs.go:333` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `memory_update` | `pkg/cortexdb/knowledge_memory_tooldefs.go:284` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `object_set_resolve` | `pkg/cortexdb/ontology_tooldefs.go:56` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ontology_action_apply` | `pkg/cortexdb/ontology_tooldefs.go:75` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ontology_action_list` | `pkg/cortexdb/ontology_tooldefs.go:70` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ontology_delete` | `pkg/cortexdb/ontology_tooldefs.go:110` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ontology_diff` | `pkg/cortexdb/ontology_tooldefs.go:42` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ontology_draft` | `pkg/cortexdb/ontology_tooldefs.go:93` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ontology_get` | `pkg/cortexdb/ontology_tooldefs.go:22` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ontology_list` | `pkg/cortexdb/ontology_tooldefs.go:32` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `ontology_save` | `pkg/cortexdb/ontology_tooldefs.go:6` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `predict_graph_edges` | `pkg/cortexdb/graphrag_tool_defs.go:237` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `rank_graph_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:222` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `representative_records` | `pkg/cortexdb/graphrag_tool_defs.go:182` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `rules_apply` | `pkg/cortexdb/rules_tooldefs.go:54` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `rules_delete` | `pkg/cortexdb/rules_tooldefs.go:41` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `rules_list` | `pkg/cortexdb/rules_tooldefs.go:31` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `rules_save` | `pkg/cortexdb/rules_tooldefs.go:14` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `search_chunks_by_entities` | `pkg/cortexdb/graphrag_tool_defs.go:267` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `search_graphrag_lexical` | `pkg/cortexdb/graphrag_tool_defs.go:449` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `search_paths` | `pkg/cortexdb/graphrag_tool_defs.go:279` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `search_text` | `pkg/cortexdb/graphrag_tool_defs.go:127` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `search_vector_range` | `pkg/cortexdb/graphrag_tool_defs.go:152` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `uncited_facts` | `pkg/cortexdb/graphrag_tool_defs.go:362` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `upsert_entities` | `pkg/cortexdb/graphrag_tool_defs.go:45` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `upsert_relations` | `pkg/cortexdb/graphrag_tool_defs.go:98` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `vacuum_graph` | `pkg/cortexdb/temporal_tooldefs.go:92` |  | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `vector_dimension_repair` | `pkg/cortexdb/graphrag_tool_defs.go:483` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |
| `verify_claims` | `pkg/cortexdb/graphrag_tool_defs.go:372` | * | hors classe : n'écrit pas l'enregistrement d'un document ; une lecture rend les marques comme toute clé de métadonnées |

### reads : chaque fonction qui lit les documents ou les chunks (16)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `GetDocument` | `pkg/core/document.go:71` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `DeleteDocument` | `pkg/core/document.go:127` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `ListDocumentsWithFilter` | `pkg/core/document.go:184` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `GetDocument` | `pkg/core/store_postgres.go:580` | * | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `DeleteDocument` | `pkg/core/store_postgres.go:645` | * | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `ListDocumentsWithFilter` | `pkg/core/store_postgres.go:654` | * | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `searchGraphRAG` | `pkg/cortexdb/graphrag.go:522` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `searchTextCandidates` | `pkg/cortexdb/graphrag_tool_search.go:129` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `cleanupKnowledgeArtifacts` | `pkg/cortexdb/knowledge_helpers.go:39` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `contentHolderQuery` | `pkg/cortexdb/knowledge_tx.go:840` | * | TestDedupeIsDecidedWithinTheCollection, TestDedupeFindsTheCopyPastSixteenHolders, TestTheDedupeLookupUsesTheContentHashIndex, TestAConfinedKeyDedupesWithinItsCollectionOnly |
| `filterScoredEmbeddings` | `pkg/cortexdb/retrieval_plan.go:263` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `loadPersistedCommunities` | `pkg/graphflow/community.go:542` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `deleteHierarchyDocs` | `pkg/graphflow/community_hierarchy.go:342` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `loadHierarchyLevels` | `pkg/graphflow/community_hierarchy.go:368` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `OrganizeFromBrain` | `pkg/graphflow/organize.go:92` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |
| `libraryDocs` | `pkg/liveview/library.go:470` |  | hors classe : lecture inchangée par #75 ; les marques sont des clés de métadonnées comme les autres, rendues telles quelles |

### writes : chaque fonction qui écrit une ligne (71)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `run` | `cmd/cortexdb-bench/main.go:251` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `runCaptureSession` | `cmd/cortexdb-mcp-stdio/capture_session.go:148` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `doctorSelfTest` | `cmd/cortexdb-mcp-stdio/doctor.go:192` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `SaveKnowledge` | `cmd/cortexdb-mcp-stdio/import_agent.go:91` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `SaveMemory` | `cmd/cortexdb-mcp-stdio/import_agent.go:96` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `importNotes` | `cmd/cortexdb-mcp-stdio/import_agent.go:261` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `importCodexMemory` | `cmd/cortexdb-mcp-stdio/import_agent.go:381` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `importSessions` | `cmd/cortexdb-mcp-stdio/import_agent.go:588` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `runImportCodeGraph` | `cmd/cortexdb-mcp-stdio/import_code.go:128` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `write` | `cmd/cortexdb-mcp-stdio/side_graph.go:165` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `applyDelete` | `pkg/connector/watcher.go:157` |  | hors classe : supprime un document, donc ses marques |
| `Remember` | `pkg/cortexdb/brain.go:35` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `PromoteToKnowledge` | `pkg/cortexdb/brain.go:178` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `RecordDecision` | `pkg/cortexdb/decision.go:318` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `StartRun` | `pkg/cortexdb/execution.go:261` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `FinishRun` | `pkg/cortexdb/execution.go:298` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `BeginStep` | `pkg/cortexdb/execution.go:415` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `EndStep` | `pkg/cortexdb/execution.go:482` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `ExtractConversation` | `pkg/cortexdb/extract_conversation.go:138` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `ReindexMemoryGraph` | `pkg/cortexdb/graph_maintenance.go:221` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `InsertGraphDocument` | `pkg/cortexdb/graphrag.go:217` |  | TestADocumentWrittenWithoutTheMarksIsNeitherUnchangedNorAHolder (enregistrement sans métadonnées, donc sans marques : ni « unchanged » ni détenteur) |
| `Call` | `pkg/cortexdb/graphrag_tool_defs.go:555` | * | TestIngestDocumentKeepsTheContentHash, TestAnIdenticalReingestWritesNothing, TestDedupeRefusesTheSameContentUnderAnotherID, TestATitleOnlyReingestKeepsTheGraphTheCallerAttached (répartition vers IngestDocument) |
| `ImportAgentMemory` | `pkg/cortexdb/import_agent_memory.go:143` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `UpdateKnowledge` | `pkg/cortexdb/knowledge_api.go:172` | * | TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques), TestAKnowledgeSaveOverAnExistingIdRetiresWhatTheOldTextLeft |
| `SaveKnowledge` | `pkg/cortexdb/knowledge_api.go:361` | * | TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques), TestAKnowledgeSaveOverAnExistingIdRetiresWhatTheOldTextLeft |
| `DeleteKnowledge` | `pkg/cortexdb/knowledge_api.go:381` | * | hors classe : supprime l'enregistrement, donc ses marques ; un document supprimé ne détient plus rien (requête du détenteur sur `documents`) |
| `NewMCPServer` | `pkg/cortexdb/mcp.go:62` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `SaveMemory` | `pkg/cortexdb/memory_api.go:593` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `UpdateMemory` | `pkg/cortexdb/memory_api.go:598` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `DeleteMemory` | `pkg/cortexdb/memory_api.go:613` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `saveMemoryGraph` | `pkg/cortexdb/memory_api.go:935` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `ApplyMemorySync` | `pkg/cortexdb/memory_sync.go:166` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `applyCreateObjectRule` | `pkg/cortexdb/ontology_action_apply.go:375` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `applyObjectTargetRule` | `pkg/cortexdb/ontology_action_apply.go:414` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `applyLinkRule` | `pkg/cortexdb/ontology_action_apply.go:435` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `UpsertNode` | `pkg/graph/graph.go:411` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `UpsertEdge` | `pkg/graph/graph.go:668` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `now` | `pkg/graph/temporal.go:179` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `Build` | `pkg/graphflow/build.go:69` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `BuildCommunitySummaries` | `pkg/graphflow/community.go:107` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `BuildCommunityHierarchy` | `pkg/graphflow/community_hierarchy.go:213` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `deleteHierarchyDocs` | `pkg/graphflow/community_hierarchy.go:357` |  | hors classe : supprime un document, donc ses marques |
| `ApplyGraphEdits` | `pkg/graphflow/graph_edit.go:243` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `ImportLearningGraph` | `pkg/graphflow/learning.go:158` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `MarkMastered` | `pkg/graphflow/learning.go:429` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `OrganizeFromBrain` | `pkg/graphflow/organize.go:185` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `organizeWithLLM` | `pkg/graphflow/organize_llm.go:122` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `recordAliases` | `pkg/graphflow/resolve.go:476` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `writePossiblySame` | `pkg/graphflow/resolve_linkfirst.go:500` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `SaveTemporalFact` | `pkg/graphflow/temporal.go:140` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `SupersedeFactWith` | `pkg/graphflow/temporal.go:306` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `Retain` | `pkg/hindsight/hindsight.go:159` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `CreateBank` | `pkg/hindsight/hindsight.go:776` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `flush` | `pkg/importflow/sink_rag.go:44` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `ApplyMemoryEdits` | `pkg/memoryflow/memory_edit.go:127` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `supersede` | `pkg/memoryflow/memory_edit.go:199` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `IngestTranscript` | `pkg/memoryflow/service.go:136` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `CloseSession` | `pkg/memoryflow/service.go:280` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `AppendDiaryEntry` | `pkg/memoryflow/service.go:382` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `_KnowledgeService_SaveKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:159` |  | hors classe : gestionnaire généré ; appelle le serveur typé (pkg/rpcserver), classé à part |
| `_KnowledgeService_UpdateKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:177` |  | hors classe : gestionnaire généré ; appelle le serveur typé (pkg/rpcserver), classé à part |
| `_KnowledgeService_DeleteKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:231` |  | hors classe : gestionnaire généré ; appelle le serveur typé (pkg/rpcserver), classé à part |
| `_MemoryService_SaveMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:159` |  | hors classe : gestionnaire généré ; appelle le serveur typé (pkg/rpcserver), classé à part |
| `_MemoryService_UpdateMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:177` |  | hors classe : gestionnaire généré ; appelle le serveur typé (pkg/rpcserver), classé à part |
| `_MemoryService_DeleteMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:231` |  | hors classe : gestionnaire généré ; appelle le serveur typé (pkg/rpcserver), classé à part |
| `SaveKnowledge` | `pkg/rpcserver/knowledge.go:44` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `UpdateKnowledge` | `pkg/rpcserver/knowledge.go:88` |  | par SaveKnowledge ou UpdateKnowledge : TestTheContentHashFollowsTheContentThroughKnowledgeWrites (SaveKnowledge et UpdateKnowledge posent les marques) |
| `DeleteKnowledge` | `pkg/rpcserver/knowledge.go:168` |  | hors classe : supprime un document, donc ses marques |
| `SaveMemory` | `pkg/rpcserver/memory.go:40` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `UpdateMemory` | `pkg/rpcserver/memory.go:74` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |
| `DeleteMemory` | `pkg/rpcserver/memory.go:137` |  | hors classe : n'écrit aucun enregistrement de document (mémoire, nœud ou arête de graphe, exécution) ; les marques ne sont ni lues ni écrites |

### keys : chaque clé réservée (18)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `OntologyDecisionNoPrimaryKey = "no_primary_key_candidate"` | `pkg/cortexdb/ontology_draft.go:118` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `OntologyDecisionPrimaryKey = "primary_key_guess"` | `pkg/cortexdb/ontology_draft.go:117` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `chunkOverlapKey = "_chunk_overlap"` | `pkg/cortexdb/graphrag_tool_ingest.go:238` | * | TestTheIngestMarksNeverOverwriteACallersKey, TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking |
| `chunkSizeKey = "_chunk_size"` | `pkg/cortexdb/graphrag_tool_ingest.go:237` | * | TestTheIngestMarksNeverOverwriteACallersKey, TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking |
| `contentHashKey = "_content_sha256"` | `pkg/cortexdb/graphrag_tool_ingest.go:236` | * | TestTheIngestMarksNeverOverwriteACallersKey, TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking |
| `entityCollectionsKey = "collections"` | `pkg/cortexdb/entity_collections.go:22` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `factRecordedAtKey = "recorded_at"` | `pkg/graphflow/temporal.go:35` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `factSupersededByKey = "superseded_by"` | `pkg/graphflow/temporal.go:231` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `factValidFromKey = "valid_from"` | `pkg/graphflow/temporal.go:33` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `factValidToKey = "valid_to"` | `pkg/graphflow/temporal.go:34` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `masteredKey = "mastered_at"` | `pkg/graphflow/learning.go:37` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `ontologyDraftPlaceholderKey = "name"` | `pkg/cortexdb/ontology_draft.go:199` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `owlRuleKey = "owl_has_key"` | `pkg/graph/owl_rl_rules.go:100` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `sharedKey = "legacy-token"` | `cmd/cortexdb-mcp-stdio/recall_remote.go:174` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `supersededAtKey = "superseded_at"` | `pkg/memoryflow/memory_edit.go:38` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `supersededByKey = "superseded_by"` | `pkg/memoryflow/memory_edit.go:37` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `supersedesKey = "supersedes"` | `pkg/memoryflow/memory_edit.go:39` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |
| `transcriptTurnsMetadataKey = "transcript_turns_json"` | `pkg/memoryflow/service.go:15` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #75 |

### promises : chaque description ajoutée par le diff (1)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `When true, content already stored under another document_id of the same collection is not ingested again; the response names that document in duplicate_of. A re-ingest identical to what is stored writes nothing either way and reports unchanged.` | `pkg/cortexdb/graphrag_tool_defs.go:40` | * | TestDedupeRefusesTheSameContentUnderAnotherID (duplicate_of, rien d'écrit), TestAnIdenticalReingestWritesNothing (unchanged sans dedupe) |
