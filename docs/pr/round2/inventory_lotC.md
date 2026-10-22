## Inventaire de `b9777c5` contre `cc98c6c` (`*` = fichier touché par le diff)

### tools : chaque outil déclaré (97)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `aggregate_metadata` | `pkg/cortexdb/graphrag_tool_defs.go:170` | * |  |
| `apply_inference` | `pkg/cortexdb/inference_tooldefs.go:6` |  |  |
| `build_context` | `pkg/cortexdb/graphrag_tool_defs.go:416` | * |  |
| `contract_needs_attention` | `pkg/cortexdb/contract_tooldefs.go:52` |  |  |
| `contract_tally` | `pkg/cortexdb/contract_tooldefs.go:47` |  |  |
| `cortex_query` | `pkg/cortexdb/graphrag_tool_defs.go:151` | * |  |
| `decision_chain` | `pkg/cortexdb/decision_tooldefs.go:101` |  |  |
| `decision_precedents` | `pkg/cortexdb/decision_tooldefs.go:113` |  |  |
| `decision_record` | `pkg/cortexdb/decision_tooldefs.go:75` |  |  |
| `delete_document_graph` | `pkg/cortexdb/graphrag_tool_defs.go:90` | * |  |
| `delete_entities` | `pkg/cortexdb/graphrag_tool_defs.go:74` | * |  |
| `disambiguate_mentions` | `pkg/cortexdb/graphrag_tool_defs.go:257` | * |  |
| `execution_run_finish` | `pkg/cortexdb/execution_tooldefs.go:160` |  |  |
| `execution_run_get` | `pkg/cortexdb/execution_tooldefs.go:172` |  |  |
| `execution_run_replay` | `pkg/cortexdb/execution_tooldefs.go:201` |  |  |
| `execution_run_start` | `pkg/cortexdb/execution_tooldefs.go:126` |  |  |
| `execution_runs_list` | `pkg/cortexdb/execution_tooldefs.go:182` |  |  |
| `execution_step_begin` | `pkg/cortexdb/execution_tooldefs.go:139` |  |  |
| `execution_step_end` | `pkg/cortexdb/execution_tooldefs.go:147` |  |  |
| `execution_step_lineage` | `pkg/cortexdb/execution_tooldefs.go:191` |  |  |
| `execution_step_record` | `pkg/cortexdb/execution_tooldefs.go:154` |  |  |
| `expand_graph` | `pkg/cortexdb/graphrag_tool_defs.go:304` | * |  |
| `extract_conversation` | `pkg/cortexdb/graphrag_tool_defs.go:434` | * |  |
| `fact_provenance` | `pkg/cortexdb/graphrag_tool_defs.go:355` | * |  |
| `find_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:329` | * |  |
| `get_chunks` | `pkg/cortexdb/graphrag_tool_defs.go:341` | * |  |
| `get_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:319` | * |  |
| `graph_diff` | `pkg/cortexdb/temporal_tooldefs.go:71` |  |  |
| `graph_health` | `pkg/cortexdb/graphrag_tool_defs.go:399` | * |  |
| `graph_list_all` | `pkg/cortexdb/knowledge_memory_tooldefs.go:324` | * |  |
| `graph_property_values` | `pkg/cortexdb/graphrag_tool_defs.go:214` | * |  |
| `graph_schema` | `pkg/cortexdb/graphrag_tool_defs.go:199` | * |  |
| `graph_snapshot` | `pkg/cortexdb/temporal_tooldefs.go:56` |  |  |
| `graph_statistics` | `pkg/cortexdb/graphrag_tool_defs.go:252` | * |  |
| `inference_explain` | `pkg/cortexdb/rules_tooldefs.go:77` |  |  |
| `ingest_document` | `pkg/cortexdb/graphrag_tool_defs.go:25` | * |  |
| `knowledge_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:98` | * |  |
| `knowledge_get` | `pkg/cortexdb/knowledge_memory_tooldefs.go:58` | * |  |
| `knowledge_graph_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:147` | * |  |
| `knowledge_graph_export` | `pkg/cortexdb/knowledge_memory_tooldefs.go:172` | * |  |
| `knowledge_graph_find` | `pkg/cortexdb/knowledge_memory_tooldefs.go:137` | * |  |
| `knowledge_graph_import` | `pkg/cortexdb/knowledge_memory_tooldefs.go:160` | * |  |
| `knowledge_graph_infer_explain` | `pkg/cortexdb/knowledge_memory_tooldefs.go:247` | * |  |
| `knowledge_graph_infer_explain_match` | `pkg/cortexdb/knowledge_memory_tooldefs.go:258` | * |  |
| `knowledge_graph_infer_refresh` | `pkg/cortexdb/knowledge_memory_tooldefs.go:225` | * |  |
| `knowledge_graph_infer_summary` | `pkg/cortexdb/knowledge_memory_tooldefs.go:242` | * |  |
| `knowledge_graph_namespace_list` | `pkg/cortexdb/knowledge_memory_tooldefs.go:121` | * |  |
| `knowledge_graph_namespace_upsert` | `pkg/cortexdb/knowledge_memory_tooldefs.go:109` | * |  |
| `knowledge_graph_query` | `pkg/cortexdb/knowledge_memory_tooldefs.go:182` | * |  |
| `knowledge_graph_shacl_rules` | `pkg/cortexdb/knowledge_memory_tooldefs.go:209` | * |  |
| `knowledge_graph_shacl_validate` | `pkg/cortexdb/knowledge_memory_tooldefs.go:199` | * |  |
| `knowledge_graph_upsert` | `pkg/cortexdb/knowledge_memory_tooldefs.go:126` | * |  |
| `knowledge_memory_build_context_pack` | `pkg/cortexdb/brain_tooldefs.go:65` |  |  |
| `knowledge_memory_consolidate` | `pkg/cortexdb/brain_tooldefs.go:231` |  |  |
| `knowledge_memory_expand_entity_context` | `pkg/cortexdb/brain_tooldefs.go:125` |  |  |
| `knowledge_memory_neighbors` | `pkg/cortexdb/brain_tooldefs.go:149` |  |  |
| `knowledge_memory_promote_to_knowledge` | `pkg/cortexdb/brain_tooldefs.go:103` |  |  |
| `knowledge_memory_recall` | `pkg/cortexdb/brain_tooldefs.go:26` |  |  |
| `knowledge_memory_reflect` | `pkg/cortexdb/brain_tooldefs.go:178` |  |  |
| `knowledge_memory_remember` | `pkg/cortexdb/brain_tooldefs.go:6` |  |  |
| `knowledge_memory_shortest_path` | `pkg/cortexdb/brain_tooldefs.go:165` |  |  |
| `knowledge_save` | `pkg/cortexdb/knowledge_memory_tooldefs.go:8` | * |  |
| `knowledge_search` | `pkg/cortexdb/knowledge_memory_tooldefs.go:68` | * |  |
| `knowledge_update` | `pkg/cortexdb/knowledge_memory_tooldefs.go:33` | * |  |
| `memory_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:368` | * |  |
| `memory_get` | `pkg/cortexdb/knowledge_memory_tooldefs.go:336` | * |  |
| `memory_list_all` | `pkg/cortexdb/knowledge_memory_tooldefs.go:313` | * |  |
| `memory_save` | `pkg/cortexdb/knowledge_memory_tooldefs.go:269` | * |  |
| `memory_search` | `pkg/cortexdb/knowledge_memory_tooldefs.go:347` | * |  |
| `memory_update` | `pkg/cortexdb/knowledge_memory_tooldefs.go:295` | * |  |
| `object_set_resolve` | `pkg/cortexdb/ontology_tooldefs.go:56` |  |  |
| `ontology_action_apply` | `pkg/cortexdb/ontology_tooldefs.go:75` |  |  |
| `ontology_action_list` | `pkg/cortexdb/ontology_tooldefs.go:70` |  |  |
| `ontology_delete` | `pkg/cortexdb/ontology_tooldefs.go:110` |  |  |
| `ontology_diff` | `pkg/cortexdb/ontology_tooldefs.go:42` |  |  |
| `ontology_draft` | `pkg/cortexdb/ontology_tooldefs.go:93` |  |  |
| `ontology_get` | `pkg/cortexdb/ontology_tooldefs.go:22` |  |  |
| `ontology_list` | `pkg/cortexdb/ontology_tooldefs.go:32` |  |  |
| `ontology_save` | `pkg/cortexdb/ontology_tooldefs.go:6` |  |  |
| `predict_graph_edges` | `pkg/cortexdb/graphrag_tool_defs.go:241` | * |  |
| `rank_graph_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:226` | * |  |
| `representative_records` | `pkg/cortexdb/graphrag_tool_defs.go:186` | * |  |
| `rules_apply` | `pkg/cortexdb/rules_tooldefs.go:54` |  |  |
| `rules_delete` | `pkg/cortexdb/rules_tooldefs.go:41` |  |  |
| `rules_list` | `pkg/cortexdb/rules_tooldefs.go:31` |  |  |
| `rules_save` | `pkg/cortexdb/rules_tooldefs.go:14` |  |  |
| `search_chunks_by_entities` | `pkg/cortexdb/graphrag_tool_defs.go:271` | * |  |
| `search_graphrag_lexical` | `pkg/cortexdb/graphrag_tool_defs.go:453` | * |  |
| `search_paths` | `pkg/cortexdb/graphrag_tool_defs.go:283` | * |  |
| `search_text` | `pkg/cortexdb/graphrag_tool_defs.go:131` | * |  |
| `search_vector_range` | `pkg/cortexdb/graphrag_tool_defs.go:156` | * |  |
| `uncited_facts` | `pkg/cortexdb/graphrag_tool_defs.go:366` | * |  |
| `upsert_entities` | `pkg/cortexdb/graphrag_tool_defs.go:49` | * |  |
| `upsert_relations` | `pkg/cortexdb/graphrag_tool_defs.go:102` | * |  |
| `vacuum_graph` | `pkg/cortexdb/temporal_tooldefs.go:92` |  |  |
| `vector_dimension_repair` | `pkg/cortexdb/graphrag_tool_defs.go:487` | * |  |
| `verify_claims` | `pkg/cortexdb/graphrag_tool_defs.go:376` | * |  |

### reads : chaque fonction qui lit les documents ou les chunks (19)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `GetDocument` | `pkg/core/document.go:71` |  |  |
| `DeleteDocument` | `pkg/core/document.go:127` |  |  |
| `ListDocumentsWithFilter` | `pkg/core/document.go:184` |  |  |
| `GetDocument` | `pkg/core/store_postgres.go:580` | * |  |
| `DeleteDocument` | `pkg/core/store_postgres.go:645` | * |  |
| `ListDocumentsWithFilter` | `pkg/core/store_postgres.go:654` | * |  |
| `hiddenDocuments` | `pkg/cortexdb/expiry.go:109` | * |  |
| `searchGraphRAG` | `pkg/cortexdb/graphrag.go:522` | * |  |
| `searchTextCandidates` | `pkg/cortexdb/graphrag_tool_search.go:129` | * |  |
| `cleanupKnowledgeArtifactsIf` | `pkg/cortexdb/knowledge_helpers.go:55` | * |  |
| `contentHolderTx` | `pkg/cortexdb/knowledge_tx.go:865` | * |  |
| `Query` | `pkg/cortexdb/query_api.go:168` | * |  |
| `filterScoredEmbeddings` | `pkg/cortexdb/retrieval_plan.go:274` | * |  |
| `SearchVectorRange` | `pkg/cortexdb/vector_range.go:157` | * |  |
| `loadPersistedCommunities` | `pkg/graphflow/community.go:542` |  |  |
| `deleteHierarchyDocs` | `pkg/graphflow/community_hierarchy.go:342` |  |  |
| `loadHierarchyLevels` | `pkg/graphflow/community_hierarchy.go:368` |  |  |
| `OrganizeFromBrain` | `pkg/graphflow/organize.go:92` |  |  |
| `libraryDocs` | `pkg/liveview/library.go:470` |  |  |

### writes : chaque fonction qui écrit une ligne (74)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `run` | `cmd/cortexdb-bench/main.go:251` |  |  |
| `runCaptureSession` | `cmd/cortexdb-mcp-stdio/capture_session.go:148` |  |  |
| `doctorSelfTest` | `cmd/cortexdb-mcp-stdio/doctor.go:192` |  |  |
| `SaveKnowledge` | `cmd/cortexdb-mcp-stdio/import_agent.go:91` |  |  |
| `SaveMemory` | `cmd/cortexdb-mcp-stdio/import_agent.go:96` |  |  |
| `importNotes` | `cmd/cortexdb-mcp-stdio/import_agent.go:261` |  |  |
| `importCodexMemory` | `cmd/cortexdb-mcp-stdio/import_agent.go:381` |  |  |
| `importSessions` | `cmd/cortexdb-mcp-stdio/import_agent.go:588` |  |  |
| `runImportCodeGraph` | `cmd/cortexdb-mcp-stdio/import_code.go:128` |  |  |
| `write` | `cmd/cortexdb-mcp-stdio/side_graph.go:165` |  |  |
| `sweepExpiredCommand` | `cmd/cortexdb-mcp-stdio/sweep_expired.go:52` | * |  |
| `applyDelete` | `pkg/connector/watcher.go:157` |  |  |
| `Remember` | `pkg/cortexdb/brain.go:35` | * |  |
| `PromoteToKnowledge` | `pkg/cortexdb/brain.go:178` | * |  |
| `RecordDecision` | `pkg/cortexdb/decision.go:318` |  |  |
| `StartRun` | `pkg/cortexdb/execution.go:261` |  |  |
| `FinishRun` | `pkg/cortexdb/execution.go:298` |  |  |
| `BeginStep` | `pkg/cortexdb/execution.go:415` |  |  |
| `EndStep` | `pkg/cortexdb/execution.go:482` |  |  |
| `SweepExpired` | `pkg/cortexdb/expiry.go:225` | * |  |
| `SweepExpiredTool` | `pkg/cortexdb/expiry_tool.go:41` | * |  |
| `ExtractConversation` | `pkg/cortexdb/extract_conversation.go:138` |  |  |
| `ReindexMemoryGraph` | `pkg/cortexdb/graph_maintenance.go:221` |  |  |
| `InsertGraphDocument` | `pkg/cortexdb/graphrag.go:217` | * |  |
| `Call` | `pkg/cortexdb/graphrag_tool_defs.go:563` | * |  |
| `ImportAgentMemory` | `pkg/cortexdb/import_agent_memory.go:143` |  |  |
| `UpdateKnowledge` | `pkg/cortexdb/knowledge_api.go:186` | * |  |
| `SaveKnowledge` | `pkg/cortexdb/knowledge_api.go:375` | * |  |
| `DeleteKnowledge` | `pkg/cortexdb/knowledge_api.go:395` | * |  |
| `NewMCPServer` | `pkg/cortexdb/mcp.go:63` | * |  |
| `SaveMemory` | `pkg/cortexdb/memory_api.go:588` | * |  |
| `UpdateMemory` | `pkg/cortexdb/memory_api.go:593` | * |  |
| `DeleteMemory` | `pkg/cortexdb/memory_api.go:608` | * |  |
| `saveMemoryGraph` | `pkg/cortexdb/memory_api.go:932` | * |  |
| `ApplyMemorySync` | `pkg/cortexdb/memory_sync.go:166` |  |  |
| `applyCreateObjectRule` | `pkg/cortexdb/ontology_action_apply.go:375` |  |  |
| `applyObjectTargetRule` | `pkg/cortexdb/ontology_action_apply.go:414` |  |  |
| `applyLinkRule` | `pkg/cortexdb/ontology_action_apply.go:435` |  |  |
| `UpsertNode` | `pkg/graph/graph.go:411` |  |  |
| `UpsertEdge` | `pkg/graph/graph.go:668` |  |  |
| `now` | `pkg/graph/temporal.go:179` |  |  |
| `Build` | `pkg/graphflow/build.go:69` |  |  |
| `BuildCommunitySummaries` | `pkg/graphflow/community.go:107` |  |  |
| `BuildCommunityHierarchy` | `pkg/graphflow/community_hierarchy.go:213` |  |  |
| `deleteHierarchyDocs` | `pkg/graphflow/community_hierarchy.go:357` |  |  |
| `ApplyGraphEdits` | `pkg/graphflow/graph_edit.go:243` |  |  |
| `ImportLearningGraph` | `pkg/graphflow/learning.go:158` |  |  |
| `MarkMastered` | `pkg/graphflow/learning.go:429` |  |  |
| `OrganizeFromBrain` | `pkg/graphflow/organize.go:185` |  |  |
| `organizeWithLLM` | `pkg/graphflow/organize_llm.go:122` |  |  |
| `recordAliases` | `pkg/graphflow/resolve.go:476` |  |  |
| `writePossiblySame` | `pkg/graphflow/resolve_linkfirst.go:500` |  |  |
| `SaveTemporalFact` | `pkg/graphflow/temporal.go:140` |  |  |
| `SupersedeFactWith` | `pkg/graphflow/temporal.go:306` |  |  |
| `Retain` | `pkg/hindsight/hindsight.go:159` |  |  |
| `CreateBank` | `pkg/hindsight/hindsight.go:776` |  |  |
| `flush` | `pkg/importflow/sink_rag.go:44` |  |  |
| `ApplyMemoryEdits` | `pkg/memoryflow/memory_edit.go:127` |  |  |
| `supersede` | `pkg/memoryflow/memory_edit.go:199` |  |  |
| `IngestTranscript` | `pkg/memoryflow/service.go:136` |  |  |
| `CloseSession` | `pkg/memoryflow/service.go:280` |  |  |
| `AppendDiaryEntry` | `pkg/memoryflow/service.go:382` |  |  |
| `_KnowledgeService_SaveKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:159` |  |  |
| `_KnowledgeService_UpdateKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:177` |  |  |
| `_KnowledgeService_DeleteKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:231` |  |  |
| `_MemoryService_SaveMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:159` |  |  |
| `_MemoryService_UpdateMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:177` |  |  |
| `_MemoryService_DeleteMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:231` |  |  |
| `SaveKnowledge` | `pkg/rpcserver/knowledge.go:44` |  |  |
| `UpdateKnowledge` | `pkg/rpcserver/knowledge.go:88` |  |  |
| `DeleteKnowledge` | `pkg/rpcserver/knowledge.go:168` |  |  |
| `SaveMemory` | `pkg/rpcserver/memory.go:40` |  |  |
| `UpdateMemory` | `pkg/rpcserver/memory.go:74` |  |  |
| `DeleteMemory` | `pkg/rpcserver/memory.go:137` |  |  |

### ids : chaque propriété d'outil qui porte un identifiant (75)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `?.run_id` | `pkg/cortexdb/execution_tooldefs.go:92` |  |  |
| `apply_inference.document_id` | `pkg/cortexdb/inference_tooldefs.go:15` |  |  |
| `apply_inference.rule_id` | `pkg/cortexdb/inference_tooldefs.go:22` |  |  |
| `build_context.chunk_ids` | `pkg/cortexdb/graphrag_tool_defs.go:421` | * |  |
| `decision_record.supersedes` | `pkg/cortexdb/decision_tooldefs.go:89` |  |  |
| `delete_document_graph.document_id` | `pkg/cortexdb/graphrag_tool_defs.go:96` | * |  |
| `execution_run_finish.run_id` | `pkg/cortexdb/execution_tooldefs.go:165` |  |  |
| `execution_run_get.run_id` | `pkg/cortexdb/execution_tooldefs.go:177` |  |  |
| `execution_run_replay.run_id` | `pkg/cortexdb/execution_tooldefs.go:205` |  |  |
| `execution_step_lineage.step_id` | `pkg/cortexdb/execution_tooldefs.go:195` |  |  |
| `expand_graph.node_ids` | `pkg/cortexdb/graphrag_tool_defs.go:309` | * |  |
| `extract_conversation.session_id` | `pkg/cortexdb/graphrag_tool_defs.go:445` | * |  |
| `fact_provenance.edge_id` | `pkg/cortexdb/graphrag_tool_defs.go:360` | * |  |
| `get_chunks.chunk_ids` | `pkg/cortexdb/graphrag_tool_defs.go:346` | * |  |
| `get_nodes.node_ids` | `pkg/cortexdb/graphrag_tool_defs.go:324` | * |  |
| `inference_explain.edge_id` | `pkg/cortexdb/rules_tooldefs.go:82` |  |  |
| `ingest_document.document_id` | `pkg/cortexdb/graphrag_tool_defs.go:31` | * |  |
| `knowledge_delete.knowledge_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:104` | * |  |
| `knowledge_get.knowledge_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:63` | * |  |
| `knowledge_graph_delete.triple_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:153` | * |  |
| `knowledge_graph_infer_explain.triple_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:252` | * |  |
| `knowledge_graph_infer_refresh.triple_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:234` | * |  |
| `knowledge_memory_build_context_pack.session_id` | `pkg/cortexdb/brain_tooldefs.go:72` |  |  |
| `knowledge_memory_build_context_pack.user_id` | `pkg/cortexdb/brain_tooldefs.go:71` |  |  |
| `knowledge_memory_consolidate.knowledge_id` | `pkg/cortexdb/brain_tooldefs.go:295` |  |  |
| `knowledge_memory_consolidate.memory_id` | `pkg/cortexdb/brain_tooldefs.go:281` |  |  |
| `knowledge_memory_consolidate.memory_ids` | `pkg/cortexdb/brain_tooldefs.go:294` |  |  |
| `knowledge_memory_consolidate.session_id` | `pkg/cortexdb/brain_tooldefs.go:247` |  |  |
| `knowledge_memory_consolidate.session_id` | `pkg/cortexdb/brain_tooldefs.go:283` |  |  |
| `knowledge_memory_consolidate.user_id` | `pkg/cortexdb/brain_tooldefs.go:246` |  |  |
| `knowledge_memory_consolidate.user_id` | `pkg/cortexdb/brain_tooldefs.go:282` |  |  |
| `knowledge_memory_expand_entity_context.entity_ids` | `pkg/cortexdb/brain_tooldefs.go:130` |  |  |
| `knowledge_memory_neighbors.node_id` | `pkg/cortexdb/brain_tooldefs.go:154` |  |  |
| `knowledge_memory_promote_to_knowledge.knowledge_id` | `pkg/cortexdb/brain_tooldefs.go:110` |  |  |
| `knowledge_memory_promote_to_knowledge.memory_ids` | `pkg/cortexdb/brain_tooldefs.go:109` |  |  |
| `knowledge_memory_recall.session_id` | `pkg/cortexdb/brain_tooldefs.go:33` |  |  |
| `knowledge_memory_recall.user_id` | `pkg/cortexdb/brain_tooldefs.go:32` |  |  |
| `knowledge_memory_reflect.session_id` | `pkg/cortexdb/brain_tooldefs.go:195` |  |  |
| `knowledge_memory_reflect.user_id` | `pkg/cortexdb/brain_tooldefs.go:194` |  |  |
| `knowledge_memory_remember.memory_id` | `pkg/cortexdb/brain_tooldefs.go:12` |  |  |
| `knowledge_memory_remember.session_id` | `pkg/cortexdb/brain_tooldefs.go:14` |  |  |
| `knowledge_memory_remember.user_id` | `pkg/cortexdb/brain_tooldefs.go:13` |  |  |
| `knowledge_memory_shortest_path.from_node_id` | `pkg/cortexdb/brain_tooldefs.go:170` |  |  |
| `knowledge_memory_shortest_path.to_node_id` | `pkg/cortexdb/brain_tooldefs.go:171` |  |  |
| `knowledge_save.knowledge_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:14` | * |  |
| `knowledge_update.knowledge_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:39` | * |  |
| `memory_delete.chunk_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:391` | * |  |
| `memory_delete.chunk_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:408` | * |  |
| `memory_delete.memory_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:374` | * |  |
| `memory_delete.rule_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:412` | * |  |
| `memory_delete.support_edge_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:413` | * |  |
| `memory_get.memory_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:341` | * |  |
| `memory_get.memory_ids` | `pkg/cortexdb/knowledge_memory_tooldefs.go:342` | * |  |
| `memory_save.memory_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:275` | * |  |
| `memory_save.session_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:277` | * |  |
| `memory_save.supersedes` | `pkg/cortexdb/knowledge_memory_tooldefs.go:288` | * |  |
| `memory_save.user_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:276` | * |  |
| `memory_search.session_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:354` | * |  |
| `memory_search.user_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:353` | * |  |
| `memory_update.memory_id` | `pkg/cortexdb/knowledge_memory_tooldefs.go:301` | * |  |
| `ontology_delete.schema_id` | `pkg/cortexdb/ontology_tooldefs.go:116` |  |  |
| `ontology_diff.schema_id` | `pkg/cortexdb/ontology_tooldefs.go:47` |  |  |
| `ontology_draft.schema_id` | `pkg/cortexdb/ontology_tooldefs.go:101` |  |  |
| `ontology_get.schema_id` | `pkg/cortexdb/ontology_tooldefs.go:27` |  |  |
| `predict_graph_edges.node_id` | `pkg/cortexdb/graphrag_tool_defs.go:246` | * |  |
| `rules_apply.document_id` | `pkg/cortexdb/rules_tooldefs.go:68` |  |  |
| `rules_apply.rule_ids` | `pkg/cortexdb/rules_tooldefs.go:62` |  |  |
| `rules_delete.rule_ids` | `pkg/cortexdb/rules_tooldefs.go:49` |  |  |
| `search_paths.node_ids` | `pkg/cortexdb/graphrag_tool_defs.go:289` | * |  |
| `upsert_entities.chunk_ids` | `pkg/cortexdb/graphrag_tool_defs.go:65` | * |  |
| `upsert_entities.document_id` | `pkg/cortexdb/graphrag_tool_defs.go:55` | * |  |
| `upsert_relations.chunk_ids` | `pkg/cortexdb/graphrag_tool_defs.go:118` | * |  |
| `upsert_relations.document_id` | `pkg/cortexdb/graphrag_tool_defs.go:108` | * |  |
| `upsert_relations.rule_id` | `pkg/cortexdb/graphrag_tool_defs.go:122` | * |  |
| `upsert_relations.support_edge_ids` | `pkg/cortexdb/graphrag_tool_defs.go:123` | * |  |

### keys : chaque clé réservée (22)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `OntologyDecisionNoPrimaryKey = "no_primary_key_candidate"` | `pkg/cortexdb/ontology_draft.go:118` |  |  |
| `OntologyDecisionPrimaryKey = "primary_key_guess"` | `pkg/cortexdb/ontology_draft.go:117` |  |  |
| `chunkOverlapKey = "_chunk_overlap"` | `pkg/cortexdb/graphrag_tool_ingest.go:247` | * |  |
| `chunkSizeKey = "_chunk_size"` | `pkg/cortexdb/graphrag_tool_ingest.go:246` | * |  |
| `contentHashKey = "_content_sha256"` | `pkg/cortexdb/graphrag_tool_ingest.go:245` | * |  |
| `entityCollectionsKey = "collections"` | `pkg/cortexdb/entity_collections.go:22` |  |  |
| `expiresAtKey = "_expires_at"` | `pkg/cortexdb/expiry.go:18` | * |  |
| `factRecordedAtKey = "recorded_at"` | `pkg/graphflow/temporal.go:35` |  |  |
| `factSupersededByKey = "superseded_by"` | `pkg/graphflow/temporal.go:231` |  |  |
| `factValidFromKey = "valid_from"` | `pkg/graphflow/temporal.go:33` |  |  |
| `factValidToKey = "valid_to"` | `pkg/graphflow/temporal.go:34` |  |  |
| `masteredKey = "mastered_at"` | `pkg/graphflow/learning.go:37` |  |  |
| `ontologyDraftPlaceholderKey = "name"` | `pkg/cortexdb/ontology_draft.go:199` |  |  |
| `owlRuleKey = "owl_has_key"` | `pkg/graph/owl_rl_rules.go:100` |  |  |
| `reviewStatusKey = "_review_status"` | `pkg/cortexdb/review.go:19` | * |  |
| `reviewedAtKey = "_reviewed_at"` | `pkg/cortexdb/review.go:21` | * |  |
| `reviewedByKey = "_reviewed_by"` | `pkg/cortexdb/review.go:20` | * |  |
| `sharedKey = "legacy-token"` | `cmd/cortexdb-mcp-stdio/recall_remote.go:174` |  |  |
| `supersededAtKey = "superseded_at"` | `pkg/memoryflow/memory_edit.go:38` |  |  |
| `supersededByKey = "superseded_by"` | `pkg/memoryflow/memory_edit.go:37` |  |  |
| `supersedesKey = "supersedes"` | `pkg/memoryflow/memory_edit.go:39` |  |  |
| `transcriptTurnsMetadataKey = "transcript_turns_json"` | `pkg/memoryflow/service.go:15` |  |  |

### promises : chaque description ajoutée par le diff (30)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `When true, content already stored under another document_id is not ingested again; the response names that document in duplicate_of. A re-ingest identical to what is stored writes nothing either way and reports unchanged.` | `pkg/cortexdb/graphrag_tool_defs.go:40` | * |  |
| `Optional RFC 3339 instant (2031-10-18T00:00:00Z) after which the document is left out of every search and removed, with its chunks and graph nodes, by `cortexdb-mcp-stdio --sweep-expired`. This field alone sets it; a metadata key of the same name is yours and is not read.` | `pkg/cortexdb/graphrag_tool_defs.go:41` | * |  |
| `Optional review decision: proposed, accepted or rejected. Left out, the stored decision is kept, whatever this call does with metadata; a decision is never cleared. ` | `pkg/cortexdb/graphrag_tool_defs.go:42` | * |  |
| `Who accepted or rejected. The calling key when the server runs with keys; required otherwise.` | `pkg/cortexdb/graphrag_tool_defs.go:43` | * |  |
| `Optional RFC 3339 instant of the decision. Left out: now, or the stored instant when the same reviewer repeats the same decision.` | `pkg/cortexdb/graphrag_tool_defs.go:44` | * |  |
| `Optional metadata equality filter, applied before aggregating. Expired documents are never counted. A filter on _review_status counts rows of that status only, which is how a curator lists what waits for review (_review_status: proposed) under the review gate. ` | `pkg/cortexdb/graphrag_tool_defs.go:178` | * |  |
| `Optional metadata equality filter. Expired documents are never chosen. A filter on _review_status chooses among rows of that status only, which lifts the review gate for that status. ` | `pkg/cortexdb/graphrag_tool_defs.go:192` | * |  |
| `Optional RFC 3339 instant after which the item is left out of every search and removed by `cortexdb-mcp-stdio --sweep-expired`. This field alone sets it; a metadata key of the same name is yours and is not read.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:22` | * |  |
| `Optional review decision: proposed, accepted or rejected. Left out, the stored decision is kept, whatever this call does with metadata; a decision is never cleared. ` | `pkg/cortexdb/knowledge_memory_tooldefs.go:24` | * |  |
| `Who accepted or rejected. The calling key when the server runs with keys; required otherwise.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:25` | * |  |
| `Optional RFC 3339 instant of the decision. Left out: now, or the stored instant when the same reviewer repeats the same decision.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:26` | * |  |
| `Optional RFC 3339 instant after which the item is left out of every search and removed by the sweep. An empty string removes the expiry; when absent, the stored expiry is kept, whatever metadata does.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:47` | * |  |
| `Optional replacement metadata. The engine's own keys (_expires_at, _review_status, _reviewed_by, _reviewed_at, the content marks) are carried whatever this map says.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:48` | * |  |
| `Optional review decision: proposed, accepted or rejected. Left out, the stored decision is kept, whatever this call does with metadata; a decision is never cleared. ` | `pkg/cortexdb/knowledge_memory_tooldefs.go:49` | * |  |
| `Who accepted or rejected. The calling key when the server runs with keys; required otherwise.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:50` | * |  |
| `Optional RFC 3339 instant of the decision. Left out: now, or the stored instant when the same reviewer repeats the same decision.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:51` | * |  |
| `Optional user ID for user-scoped memory.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:276` | * |  |
| `Optional session ID for session-scoped memory.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:277` | * |  |
| `Optional message role. Defaults to memory.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:280` | * |  |
| `Optional TTL in seconds.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:284` | * |  |
| `Optional review decision: proposed, accepted or rejected. Left out, the stored decision is kept, whatever this call does with metadata; a decision is never cleared. ` | `pkg/cortexdb/knowledge_memory_tooldefs.go:285` | * |  |
| `Who accepted or rejected. The calling key when the server runs with keys; required otherwise.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:286` | * |  |
| `Optional RFC 3339 instant of the decision. Left out: now, or the stored instant when the same reviewer repeats the same decision.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:287` | * |  |
| `Memory IDs this memory replaces. They stay stored and exported, but recall stops returning them as current.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:288` | * |  |
| `Optional metadata fields to merge.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:303` | * |  |
| `Optional updated importance score.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:304` | * |  |
| `Optional updated TTL in seconds.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:305` | * |  |
| `Optional review decision: proposed, accepted or rejected. Left out, the stored decision is kept, whatever this call does with metadata; a decision is never cleared. ` | `pkg/cortexdb/knowledge_memory_tooldefs.go:306` | * |  |
| `Who accepted or rejected. The calling key when the server runs with keys; required otherwise.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:307` | * |  |
| `Optional RFC 3339 instant of the decision. Left out: now, or the stored instant when the same reviewer repeats the same decision.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:308` | * |  |
