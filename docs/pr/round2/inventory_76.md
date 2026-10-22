# Inventaire de #76 (gc-05-expires-at), sur #75 rebasée sur v2.136.0

Généré par `make inventory` sur la tête de #76 contre la tête de #75, classes `tools`, `reads`, `writes`, `keys`, `exports`, `rpcs`, `promises` ; colonne « Test » remplie par règle (`fill76.py` du 22 octobre 2026). La classe `tools` suit le tableau `readToolsAndTheExpiry` de `pkg/cortexdb/expiry_class_test.go` : un outil de lecture porte « porte » (appelé avec les identifiants du document expiré) ou « nommé » (dans le corps de la PR). La classe `ids` est hors de #76 : il n'ajoute aucun champ d'identifiant.

## Inventaire de `c80797b` contre `28a777e`, lignes revues sur `e4c6cf0` contre `53121d1` (tours 3 et 4) (`*` = fichier touché par le diff)

### tools : chaque outil déclaré (97)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `aggregate_metadata` | `pkg/cortexdb/graphrag_tool_defs.go:167` | * | TestEveryReadToolIsGatedOrNamed, TestAggregateMetadataCountsVisibleRowsOnly (porte) |
| `apply_inference` | `pkg/cortexdb/inference_tooldefs.go:6` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `build_context` | `pkg/cortexdb/graphrag_tool_defs.go:413` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré), TestTheChunkToolsLeaveAnExpiredDocumentOut |
| `contract_needs_attention` | `pkg/cortexdb/contract_tooldefs.go:52` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: contracts) |
| `contract_tally` | `pkg/cortexdb/contract_tooldefs.go:47` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: contracts) |
| `cortex_query` | `pkg/cortexdb/graphrag_tool_defs.go:148` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré), TestAnExpiredDocumentIsLeftOutOfCortexQuery |
| `decision_chain` | `pkg/cortexdb/decision_tooldefs.go:101` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: decisions) |
| `decision_precedents` | `pkg/cortexdb/decision_tooldefs.go:113` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: decisions) |
| `decision_record` | `pkg/cortexdb/decision_tooldefs.go:75` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `delete_document_graph` | `pkg/cortexdb/graphrag_tool_defs.go:87` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `delete_entities` | `pkg/cortexdb/graphrag_tool_defs.go:71` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `disambiguate_mentions` | `pkg/cortexdb/graphrag_tool_defs.go:254` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : entities and relations) |
| `execution_run_finish` | `pkg/cortexdb/execution_tooldefs.go:160` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `execution_run_get` | `pkg/cortexdb/execution_tooldefs.go:172` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: runs) |
| `execution_run_replay` | `pkg/cortexdb/execution_tooldefs.go:201` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: runs) |
| `execution_run_start` | `pkg/cortexdb/execution_tooldefs.go:126` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `execution_runs_list` | `pkg/cortexdb/execution_tooldefs.go:182` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: runs) |
| `execution_step_begin` | `pkg/cortexdb/execution_tooldefs.go:139` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `execution_step_end` | `pkg/cortexdb/execution_tooldefs.go:147` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `execution_step_lineage` | `pkg/cortexdb/execution_tooldefs.go:191` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: runs) |
| `execution_step_record` | `pkg/cortexdb/execution_tooldefs.go:154` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `expand_graph` | `pkg/cortexdb/graphrag_tool_defs.go:301` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré), TestTheChunkToolsLeaveAnExpiredDocumentOut |
| `extract_conversation` | `pkg/cortexdb/graphrag_tool_defs.go:431` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `fact_provenance` | `pkg/cortexdb/graphrag_tool_defs.go:352` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : entities and relations: names the chunk ids, reports an expired one missing) |
| `find_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:326` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : entities and relations) |
| `get_chunks` | `pkg/cortexdb/graphrag_tool_defs.go:338` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré), TestTheChunkToolsLeaveAnExpiredDocumentOut |
| `get_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:316` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré) |
| `graph_diff` | `pkg/cortexdb/temporal_tooldefs.go:71` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `graph_health` | `pkg/cortexdb/graphrag_tool_defs.go:396` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `graph_list_all` | `pkg/cortexdb/knowledge_memory_tooldefs.go:312` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `graph_property_values` | `pkg/cortexdb/graphrag_tool_defs.go:211` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `graph_schema` | `pkg/cortexdb/graphrag_tool_defs.go:196` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `graph_snapshot` | `pkg/cortexdb/temporal_tooldefs.go:56` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `graph_statistics` | `pkg/cortexdb/graphrag_tool_defs.go:249` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `inference_explain` | `pkg/cortexdb/rules_tooldefs.go:77` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : entities and relations) |
| `ingest_document` | `pkg/cortexdb/graphrag_tool_defs.go:25` | * | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly, TestAClientKeyNamedExpiresAtIsStoredAsIsAndNeverRead, TestAConfinedKeySetsAnExpiryButDoesNotSweep |
| `knowledge_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:92` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_get` | `pkg/cortexdb/knowledge_memory_tooldefs.go:52` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : by id: the record, until the sweep) |
| `knowledge_graph_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:141` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_graph_export` | `pkg/cortexdb/knowledge_memory_tooldefs.go:166` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `knowledge_graph_find` | `pkg/cortexdb/knowledge_memory_tooldefs.go:131` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `knowledge_graph_import` | `pkg/cortexdb/knowledge_memory_tooldefs.go:154` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_graph_infer_explain` | `pkg/cortexdb/knowledge_memory_tooldefs.go:241` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `knowledge_graph_infer_explain_match` | `pkg/cortexdb/knowledge_memory_tooldefs.go:252` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `knowledge_graph_infer_refresh` | `pkg/cortexdb/knowledge_memory_tooldefs.go:219` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_graph_infer_summary` | `pkg/cortexdb/knowledge_memory_tooldefs.go:236` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `knowledge_graph_namespace_list` | `pkg/cortexdb/knowledge_memory_tooldefs.go:115` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `knowledge_graph_namespace_upsert` | `pkg/cortexdb/knowledge_memory_tooldefs.go:103` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_graph_query` | `pkg/cortexdb/knowledge_memory_tooldefs.go:176` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_graph_shacl_rules` | `pkg/cortexdb/knowledge_memory_tooldefs.go:203` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_graph_shacl_validate` | `pkg/cortexdb/knowledge_memory_tooldefs.go:193` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `knowledge_graph_upsert` | `pkg/cortexdb/knowledge_memory_tooldefs.go:120` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_memory_build_context_pack` | `pkg/cortexdb/brain_tooldefs.go:65` |  | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré) |
| `knowledge_memory_consolidate` | `pkg/cortexdb/brain_tooldefs.go:231` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_memory_expand_entity_context` | `pkg/cortexdb/brain_tooldefs.go:125` |  | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré) |
| `knowledge_memory_neighbors` | `pkg/cortexdb/brain_tooldefs.go:149` |  | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré) |
| `knowledge_memory_promote_to_knowledge` | `pkg/cortexdb/brain_tooldefs.go:103` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_memory_recall` | `pkg/cortexdb/brain_tooldefs.go:26` |  | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré) |
| `knowledge_memory_reflect` | `pkg/cortexdb/brain_tooldefs.go:178` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_memory_remember` | `pkg/cortexdb/brain_tooldefs.go:6` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `knowledge_memory_shortest_path` | `pkg/cortexdb/brain_tooldefs.go:165` |  | TestEveryReadToolIsGatedOrNamed (porte : depuis le chunk expiré, « no path » ; depuis le chunk vivant, un chemin), TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut (arête seulement expirée : « no path » ; arête partagée : provenance vivante seule) |
| `knowledge_save` | `pkg/cortexdb/knowledge_memory_tooldefs.go:8` | * | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly, TestAClientKeyNamedExpiresAtIsStoredAsIsAndNeverRead, TestAConfinedKeySetsAnExpiryButDoesNotSweep |
| `knowledge_search` | `pkg/cortexdb/knowledge_memory_tooldefs.go:62` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré), TestAnEmptyExpiresAtClearsTheExpiry, TestAnExpiredDocumentIsLeftOutOfSearch, TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo |
| `knowledge_update` | `pkg/cortexdb/knowledge_memory_tooldefs.go:30` | * | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly, TestAClientKeyNamedExpiresAtIsStoredAsIsAndNeverRead, TestAConfinedKeySetsAnExpiryButDoesNotSweep |
| `memory_delete` | `pkg/cortexdb/knowledge_memory_tooldefs.go:356` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `memory_get` | `pkg/cortexdb/knowledge_memory_tooldefs.go:324` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : memories: their own expires_at, as on main) |
| `memory_list_all` | `pkg/cortexdb/knowledge_memory_tooldefs.go:301` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : memories: their own expires_at, as on main) |
| `memory_save` | `pkg/cortexdb/knowledge_memory_tooldefs.go:263` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `memory_search` | `pkg/cortexdb/knowledge_memory_tooldefs.go:335` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : memories: their own expires_at, as on main) |
| `memory_update` | `pkg/cortexdb/knowledge_memory_tooldefs.go:286` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `object_set_resolve` | `pkg/cortexdb/ontology_tooldefs.go:56` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: ontologies) |
| `ontology_action_apply` | `pkg/cortexdb/ontology_tooldefs.go:75` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `ontology_action_list` | `pkg/cortexdb/ontology_tooldefs.go:70` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: ontologies) |
| `ontology_delete` | `pkg/cortexdb/ontology_tooldefs.go:110` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `ontology_diff` | `pkg/cortexdb/ontology_tooldefs.go:42` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: ontologies) |
| `ontology_draft` | `pkg/cortexdb/ontology_tooldefs.go:93` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: ontologies) |
| `ontology_get` | `pkg/cortexdb/ontology_tooldefs.go:22` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: ontologies) |
| `ontology_list` | `pkg/cortexdb/ontology_tooldefs.go:32` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: ontologies) |
| `ontology_save` | `pkg/cortexdb/ontology_tooldefs.go:6` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `predict_graph_edges` | `pkg/cortexdb/graphrag_tool_defs.go:238` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `rank_graph_nodes` | `pkg/cortexdb/graphrag_tool_defs.go:223` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : the whole graph) |
| `representative_records` | `pkg/cortexdb/graphrag_tool_defs.go:183` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré), TestAnUnreadableBrainIsNotAnEmptyBrain, TestRepresentativeRecordsNeverPickAHiddenChunk |
| `rules_apply` | `pkg/cortexdb/rules_tooldefs.go:54` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `rules_delete` | `pkg/cortexdb/rules_tooldefs.go:41` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `rules_list` | `pkg/cortexdb/rules_tooldefs.go:31` |  | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : no document: rules) |
| `rules_save` | `pkg/cortexdb/rules_tooldefs.go:14` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `search_chunks_by_entities` | `pkg/cortexdb/graphrag_tool_defs.go:268` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré), TestTheChunkToolsLeaveAnExpiredDocumentOut |
| `search_graphrag_lexical` | `pkg/cortexdb/graphrag_tool_defs.go:450` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré) |
| `search_paths` | `pkg/cortexdb/graphrag_tool_defs.go:280` | * | TestEveryReadToolIsGatedOrNamed (porte : relation portée par le chunk expiré, avec include_bookkeeping), TestAnEdgeDoesNotNameAnExpiredChunkOrDocument, TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut |
| `search_text` | `pkg/cortexdb/graphrag_tool_defs.go:128` | * | TestEveryReadToolIsGatedOrNamed (porte : appel avec les identifiants du document expiré), TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo, TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `search_vector_range` | `pkg/cortexdb/graphrag_tool_defs.go:153` | * | TestEveryReadToolIsGatedOrNamed, TestSearchVectorRangeLeavesExpiredDocumentsOut (porte) |
| `uncited_facts` | `pkg/cortexdb/graphrag_tool_defs.go:363` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : entities and relations) |
| `upsert_entities` | `pkg/cortexdb/graphrag_tool_defs.go:46` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `upsert_relations` | `pkg/cortexdb/graphrag_tool_defs.go:99` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `vacuum_graph` | `pkg/cortexdb/temporal_tooldefs.go:92` |  | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `vector_dimension_repair` | `pkg/cortexdb/graphrag_tool_defs.go:484` | * | hors classe : outil d'écriture (Mutates, TestEveryToolDeclaresWhetherItWrites) ; ne pose ni ne lit l'expiration d'un document |
| `verify_claims` | `pkg/cortexdb/graphrag_tool_defs.go:373` | * | TestEveryReadToolIsGatedOrNamed (nommé dans le corps : entities and relations) |

### reads : chaque fonction qui lit les documents ou les chunks (20)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `excludeDocuments` | `pkg/core/aggregations.go:519` | * | TestBothStoresLeaveExcludedDocumentsOutOfAggregates, TestTheAggregatesHoldPastTheBoundParameterLimit |
| `GetDocument` | `pkg/core/document.go:71` |  | TestABadExpiresAtIsRefused, TestADocumentWrittenWithoutTheMarksIsNeitherUnchangedNorAHolder, TestAKnowledgeWriteWhoseTransactionFailsIsReported, TestARetireThatCannotCompleteIsReportedAndWritesNothing |
| `DeleteDocument` | `pkg/core/document.go:127` |  | hors classe : suppression ; le balayage la précède d'une seconde lecture (TestTheSweepDeletesOnlyWhatIsStillExpiredWhenItDeletes) |
| `ListDocumentsWithFilter` | `pkg/core/document.go:184` |  | hors classe : liste des enregistrements du magasin, sous la porte ; aucun outil de lecture ne la rend sans la porte (TestEveryReadToolIsGatedOrNamed) |
| `GetDocument` | `pkg/core/store_postgres.go:583` | * | TestABadExpiresAtIsRefused, TestADocumentWrittenWithoutTheMarksIsNeitherUnchangedNorAHolder, TestAKnowledgeWriteWhoseTransactionFailsIsReported, TestARetireThatCannotCompleteIsReportedAndWritesNothing |
| `DeleteDocument` | `pkg/core/store_postgres.go:648` | * | hors classe : suppression ; le balayage la précède d'une seconde lecture (TestTheSweepDeletesOnlyWhatIsStillExpiredWhenItDeletes) |
| `ListDocumentsWithFilter` | `pkg/core/store_postgres.go:657` | * | hors classe : liste des enregistrements du magasin, sous la porte ; aucun outil de lecture ne la rend sans la porte (TestEveryReadToolIsGatedOrNamed) |
| `pgAggregateWhere` | `pkg/core/store_postgres_aggregate.go:75` | * | TestBothStoresLeaveExcludedDocumentsOutOfAggregates (PostgreSQL, CORTEXDB_TEST_POSTGRES) |
| `hiddenDocuments` | `pkg/cortexdb/expiry.go:102` | * | TestASecondProcessSeesAnExpiryAtOnce, TestAnUnreadableBrainIsNotAnEmptyBrain, TestTheDocumentExpiryIsIndexed |
| `searchTextCandidates` | `pkg/cortexdb/graphrag_tool_search.go:132` | * | TestEveryReadToolIsGatedOrNamed (search_text), TestAnExpiredDocumentIsLeftOutOfSearch |
| `cleanupKnowledgeArtifactsIf` | `pkg/cortexdb/knowledge_helpers.go:55` | * | TestTheSweepDeletesOnlyWhatIsStillExpiredWhenItDeletes, TestARowTheSweepCannotReadIsReported |
| `contentHolderQuery` | `pkg/cortexdb/knowledge_tx.go:843` | * | TestDedupeLooksUpThroughTheSearchGate |
| `Query` | `pkg/cortexdb/query_api.go:169` | * | TestAnExpiredDocumentIsLeftOutOfCortexQuery, TestASourceLaneDoesNotServeAnExpiredRow (voie source) |
| `filterScoredEmbeddings` | `pkg/cortexdb/retrieval_plan.go:269` | * | TestAnExpiredDocumentIsLeftOutOfSearch, TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo |
| `gatedSearch` | `pkg/cortexdb/retrieval_plan.go:296` | * | TestGatedSearchTrimsToTheBudgetAndReportsAFailure, TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `loadPersistedCommunities` | `pkg/graphflow/community.go:542` |  | hors classe : résumés de communautés (documents du moteur, sans expiration) |
| `deleteHierarchyDocs` | `pkg/graphflow/community_hierarchy.go:342` |  | hors classe : documents de hiérarchie du moteur, sans expiration |
| `loadHierarchyLevels` | `pkg/graphflow/community_hierarchy.go:368` |  | hors classe : documents de hiérarchie du moteur, sans expiration |
| `OrganizeFromBrain` | `pkg/graphflow/organize.go:92` |  | nommé dans le corps : construit les entités de chaque document, expirés compris jusqu'au balayage |
| `libraryDocs` | `pkg/liveview/library.go:470` |  | hors classe : vue locale du cerveau entier (liveview), comme graph_list_all, nommée avec les vues du graphe entier |

### writes : chaque fonction qui écrit une ligne (74)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `run` | `cmd/cortexdb-bench/main.go:251` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `runCaptureSession` | `cmd/cortexdb-mcp-stdio/capture_session.go:148` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `doctorSelfTest` | `cmd/cortexdb-mcp-stdio/doctor.go:192` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `SaveKnowledge` | `cmd/cortexdb-mcp-stdio/import_agent.go:91` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `SaveMemory` | `cmd/cortexdb-mcp-stdio/import_agent.go:96` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `importNotes` | `cmd/cortexdb-mcp-stdio/import_agent.go:261` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `importCodexMemory` | `cmd/cortexdb-mcp-stdio/import_agent.go:381` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `importSessions` | `cmd/cortexdb-mcp-stdio/import_agent.go:588` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `runImportCodeGraph` | `cmd/cortexdb-mcp-stdio/import_code.go:128` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `write` | `cmd/cortexdb-mcp-stdio/side_graph.go:165` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `sweepExpiredCommand` | `cmd/cortexdb-mcp-stdio/sweep_expired.go:52` | * | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `applyDelete` | `pkg/connector/watcher.go:157` |  | hors classe : supprime un document, expiration comprise |
| `Remember` | `pkg/cortexdb/brain.go:35` | * | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `PromoteToKnowledge` | `pkg/cortexdb/brain.go:178` | * | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `RecordDecision` | `pkg/cortexdb/decision.go:318` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `StartRun` | `pkg/cortexdb/execution.go:261` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `FinishRun` | `pkg/cortexdb/execution.go:298` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `BeginStep` | `pkg/cortexdb/execution.go:415` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `EndStep` | `pkg/cortexdb/execution.go:482` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `sweepMemory` | `pkg/cortexdb/expiry.go:233` | * | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `SweepExpiredTool` | `pkg/cortexdb/expiry_tool.go:41` | * | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `ExtractConversation` | `pkg/cortexdb/extract_conversation.go:138` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `ReindexMemoryGraph` | `pkg/cortexdb/graph_maintenance.go:221` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `InsertGraphDocument` | `pkg/cortexdb/graphrag.go:217` | * | TestInsertGraphDocumentKeepsTheStoredExpiry (pas de champ expires_at ; garde l'expiration stockée, tour 3) |
| `Call` | `pkg/cortexdb/graphrag_tool_defs.go:560` | * | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly (ingest_document), TestTheSweepIsAToolOfTheSharedBrain |
| `ImportAgentMemory` | `pkg/cortexdb/import_agent_memory.go:143` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `UpdateKnowledge` | `pkg/cortexdb/knowledge_api.go:184` | * | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly, TestAnExpiryCanBeSetOnARecordWithoutMetadata, TestTheEngineKeyInClientMetadataIsDropped |
| `SaveKnowledge` | `pkg/cortexdb/knowledge_api.go:373` | * | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly, TestAnExpiryCanBeSetOnARecordWithoutMetadata, TestTheEngineKeyInClientMetadataIsDropped |
| `DeleteKnowledge` | `pkg/cortexdb/knowledge_api.go:393` | * | hors classe : supprime un document, expiration comprise |
| `NewMCPServer` | `pkg/cortexdb/mcp.go:63` | * | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `SaveMemory` | `pkg/cortexdb/memory_api.go:593` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `UpdateMemory` | `pkg/cortexdb/memory_api.go:598` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `DeleteMemory` | `pkg/cortexdb/memory_api.go:613` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `saveMemoryGraph` | `pkg/cortexdb/memory_api.go:935` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `ApplyMemorySync` | `pkg/cortexdb/memory_sync.go:166` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `applyCreateObjectRule` | `pkg/cortexdb/ontology_action_apply.go:375` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `applyObjectTargetRule` | `pkg/cortexdb/ontology_action_apply.go:414` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `applyLinkRule` | `pkg/cortexdb/ontology_action_apply.go:435` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `UpsertNode` | `pkg/graph/graph.go:411` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `UpsertEdge` | `pkg/graph/graph.go:668` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `now` | `pkg/graph/temporal.go:179` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `Build` | `pkg/graphflow/build.go:69` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `BuildCommunitySummaries` | `pkg/graphflow/community.go:107` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `BuildCommunityHierarchy` | `pkg/graphflow/community_hierarchy.go:213` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `deleteHierarchyDocs` | `pkg/graphflow/community_hierarchy.go:357` |  | hors classe : supprime un document, expiration comprise |
| `ApplyGraphEdits` | `pkg/graphflow/graph_edit.go:243` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `ImportLearningGraph` | `pkg/graphflow/learning.go:158` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `MarkMastered` | `pkg/graphflow/learning.go:429` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `OrganizeFromBrain` | `pkg/graphflow/organize.go:185` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `organizeWithLLM` | `pkg/graphflow/organize_llm.go:122` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `recordAliases` | `pkg/graphflow/resolve.go:476` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `writePossiblySame` | `pkg/graphflow/resolve_linkfirst.go:500` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `SaveTemporalFact` | `pkg/graphflow/temporal.go:140` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `SupersedeFactWith` | `pkg/graphflow/temporal.go:306` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `Retain` | `pkg/hindsight/hindsight.go:159` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `CreateBank` | `pkg/hindsight/hindsight.go:776` |  | hors classe : n'écrit aucun enregistrement de document ; l'expiration n'est ni posée ni lue |
| `flush` | `pkg/importflow/sink_rag.go:44` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `ApplyMemoryEdits` | `pkg/memoryflow/memory_edit.go:127` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `supersede` | `pkg/memoryflow/memory_edit.go:199` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `IngestTranscript` | `pkg/memoryflow/service.go:136` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `CloseSession` | `pkg/memoryflow/service.go:280` |  | hors classe : écrit par SaveKnowledge ou UpdateKnowledge sans ExpiresAt ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `AppendDiaryEntry` | `pkg/memoryflow/service.go:382` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `_KnowledgeService_SaveKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:159` |  | hors classe : gestionnaire généré, classé avec le serveur typé |
| `_KnowledgeService_UpdateKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:177` |  | hors classe : gestionnaire généré, classé avec le serveur typé |
| `_KnowledgeService_DeleteKnowledge_Handler` | `pkg/rpc/v1/knowledge_grpc.pb.go:231` |  | hors classe : gestionnaire généré, classé avec le serveur typé |
| `_MemoryService_SaveMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:159` |  | hors classe : gestionnaire généré, classé avec le serveur typé |
| `_MemoryService_UpdateMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:177` |  | hors classe : gestionnaire généré, classé avec le serveur typé |
| `_MemoryService_DeleteMemory_Handler` | `pkg/rpc/v1/memory_grpc.pb.go:231` |  | hors classe : gestionnaire généré, classé avec le serveur typé |
| `SaveKnowledge` | `pkg/rpcserver/knowledge.go:44` |  | nommé dans le corps : message typé sans expires_at ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `UpdateKnowledge` | `pkg/rpcserver/knowledge.go:88` |  | nommé dans le corps : message typé sans expires_at ; champ absent, l'expiration stockée reste (TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly) |
| `DeleteKnowledge` | `pkg/rpcserver/knowledge.go:168` |  | hors classe : supprime un document, expiration comprise |
| `SaveMemory` | `pkg/rpcserver/memory.go:40` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `UpdateMemory` | `pkg/rpcserver/memory.go:74` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |
| `DeleteMemory` | `pkg/rpcserver/memory.go:137` |  | hors classe : mémoire, son expires_at de main (ttl_seconds), nommé dans le corps |

### keys : chaque clé réservée (19)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `OntologyDecisionNoPrimaryKey = "no_primary_key_candidate"` | `pkg/cortexdb/ontology_draft.go:118` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `OntologyDecisionPrimaryKey = "primary_key_guess"` | `pkg/cortexdb/ontology_draft.go:117` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `chunkOverlapKey = "_chunk_overlap"` | `pkg/cortexdb/graphrag_tool_ingest.go:249` | * | TestTheIngestMarksNeverOverwriteACallersKey (#75) |
| `chunkSizeKey = "_chunk_size"` | `pkg/cortexdb/graphrag_tool_ingest.go:248` | * | TestTheIngestMarksNeverOverwriteACallersKey (#75) |
| `contentHashKey = "_content_sha256"` | `pkg/cortexdb/graphrag_tool_ingest.go:247` | * | TestTheIngestMarksNeverOverwriteACallersKey (#75) |
| `entityCollectionsKey = "collections"` | `pkg/cortexdb/entity_collections.go:22` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `expiresAtKey = "_expires_at"` | `pkg/cortexdb/expiry.go:20` | * | TestAClientKeyNamedExpiresAtIsStoredAsIsAndNeverRead, TestTheEngineKeyInClientMetadataIsDropped, TestTheExpiryIsKeptOnTheRecordOnly |
| `factRecordedAtKey = "recorded_at"` | `pkg/graphflow/temporal.go:35` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `factSupersededByKey = "superseded_by"` | `pkg/graphflow/temporal.go:231` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `factValidFromKey = "valid_from"` | `pkg/graphflow/temporal.go:33` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `factValidToKey = "valid_to"` | `pkg/graphflow/temporal.go:34` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `masteredKey = "mastered_at"` | `pkg/graphflow/learning.go:37` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `ontologyDraftPlaceholderKey = "name"` | `pkg/cortexdb/ontology_draft.go:199` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `owlRuleKey = "owl_has_key"` | `pkg/graph/owl_rl_rules.go:100` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `sharedKey = "legacy-token"` | `cmd/cortexdb-mcp-stdio/recall_remote.go:174` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `supersededAtKey = "superseded_at"` | `pkg/memoryflow/memory_edit.go:38` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `supersededByKey = "superseded_by"` | `pkg/memoryflow/memory_edit.go:37` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `supersedesKey = "supersedes"` | `pkg/memoryflow/memory_edit.go:39` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |
| `transcriptTurnsMetadataKey = "transcript_turns_json"` | `pkg/memoryflow/service.go:15` |  | hors classe : clé d'un autre domaine ; ni lue ni écrite par #76 |

### promises : chaque description ajoutée par le diff (3)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `` Optional RFC 3339 instant (2031-10-18T00:00:00Z) after which the document is left out of every search and removed, with its chunks and graph nodes, by `cortexdb-mcp-stdio --sweep-expired`. This field alone sets it; a metadata key of the same name is yours and is not read. `` | `pkg/cortexdb/graphrag_tool_defs.go:41` | * | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly, TestAClientKeyNamedExpiresAtIsStoredAsIsAndNeverRead, TestSweepExpiredRemovesExpiredDocumentsAndMemories, TestEveryReadToolIsGatedOrNamed |
| `` Optional RFC 3339 instant after which the item is left out of every search and removed by `cortexdb-mcp-stdio --sweep-expired`. This field alone sets it; a metadata key of the same name is yours and is not read. `` | `pkg/cortexdb/knowledge_memory_tooldefs.go:22` | * | TestAClientKeyNamedExpiresAtIsStoredAsIsAndNeverRead, TestSweepExpiredRemovesExpiredDocumentsAndMemories |
| `Optional RFC 3339 instant after which the item is left out of every search and removed by the sweep. An empty string removes the expiry; when absent, the stored expiry is kept, whatever metadata does.` | `pkg/cortexdb/knowledge_memory_tooldefs.go:44` | * | TestAnEmptyExpiresAtClearsTheExpiry, TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly |

### exports : chaque entrée exportée de pkg/cortexdb qui lit (66)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `DB.Aggregate` | `pkg/cortexdb/aggregate.go:32` | * | TestTheAggregatesHoldPastTheBoundParameterLimit |
| `DB.FindKnowledgeGraph` | `pkg/cortexdb/knowledge_graph_api.go:63` |  | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_graph_find, nommé : the whole graph) |
| `DB.GetKnowledge` | `pkg/cortexdb/knowledge_api.go:236` | * | TestTheSweepIsAToolOfTheSharedBrain |
| `DB.GetMemory` | `pkg/cortexdb/memory_api.go:236` |  | TestTheSweepDeletesOnlyWhatIsStillExpiredWhenItDeletes |
| `DB.GetOntologySchema` | `pkg/cortexdb/ontology_api.go:62` |  | TestEveryReadToolIsGatedOrNamed (par l'outil ontology_get, nommé : no document: ontologies) |
| `DB.GetRun` | `pkg/cortexdb/execution.go:305` |  | TestEveryReadToolIsGatedOrNamed (par l'outil execution_run_get, nommé : no document: runs) |
| `DB.GetStep` | `pkg/cortexdb/execution.go:499` |  | TestEveryReadToolIsGatedOrNamed (par l'outil execution_step_lineage, nommé : no document: runs) |
| `DB.HybridSearchText` | `pkg/cortexdb/text_search.go:196` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `DB.HybridSearchTextWithOptions` | `pkg/cortexdb/text_search.go:223` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `DB.ListActionTypes` | `pkg/cortexdb/ontology_action_apply.go:57` |  | TestEveryReadToolIsGatedOrNamed (par l'outil ontology_action_list, nommé : no document: ontologies) |
| `DB.ListAllMemories` | `pkg/cortexdb/memory_api.go:506` |  | TestEveryReadToolIsGatedOrNamed (par l'outil memory_list_all, nommé : memories: their own expires_at, as on main) |
| `DB.ListAllMemoriesPaged` | `pkg/cortexdb/memory_list_all.go:90` |  | TestEveryReadToolIsGatedOrNamed (par l'outil memory_list_all, nommé : memories: their own expires_at, as on main) |
| `DB.ListGraphAll` | `pkg/cortexdb/graph_list_all.go:125` |  | TestEveryReadToolIsGatedOrNamed (par l'outil graph_list_all, nommé : the whole graph) |
| `DB.ListKnowledgeNamespaces` | `pkg/cortexdb/knowledge_graph_api.go:25` |  | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_graph_namespace_list, nommé : the whole graph) |
| `DB.ListOntologySchemas` | `pkg/cortexdb/ontology_api.go:70` |  | TestEveryReadToolIsGatedOrNamed (par l'outil ontology_list, nommé : no document: ontologies) |
| `DB.ListRules` | `pkg/cortexdb/rules_api.go:199` |  | TestEveryReadToolIsGatedOrNamed (par l'outil rules_list, nommé : no document: rules) |
| `DB.ListRuns` | `pkg/cortexdb/execution.go:322` |  | TestEveryReadToolIsGatedOrNamed (par l'outil execution_runs_list, nommé : no document: runs) |
| `DB.Query` | `pkg/cortexdb/query_api.go:144` | * | TestAnExpiredDocumentIsLeftOutOfCortexQuery |
| `DB.QueryCypher` | `pkg/cortexdb/graph_cypher.go:59` |  | TestEveryReadToolIsGatedOrNamed (par l'outil graph_cypher_query, nommé : the whole graph) |
| `DB.QueryKnowledgeGraph` | `pkg/cortexdb/knowledge_graph_api.go:145` |  | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_graph_find, nommé : the whole graph) |
| `DB.QuerySources` | `pkg/cortexdb/query_source.go:96` |  | hors classe : sources de requête nommées (configuration), aucun document |
| `DB.Quick` | `pkg/cortexdb/quick.go:17` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `DB.RangeSearchText` | `pkg/cortexdb/vector_range.go:93` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `DB.RangeSearchVector` | `pkg/cortexdb/vector_range.go:65` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `DB.SearchGraphRAG` | `pkg/cortexdb/graphrag.go:456` | * | TestSearchGraphRAGLeavesAnExpiredPassageOut, TestAPinWithNothingExpiredFiltersTheSeedsAsOnMain |
| `DB.SearchKnowledge` | `pkg/cortexdb/knowledge_api.go:245` | * | TestAnEmptyExpiresAtClearsTheExpiry, TestAnExpiredDocumentIsLeftOutOfSearch, TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo, TestTheGraphWalkLeavesAnExpiredPassageOut |
| `DB.SearchMemory` | `pkg/cortexdb/memory_api.go:287` |  | TestEveryReadToolIsGatedOrNamed (par l'outil memory_search, nommé : memories: their own expires_at, as on main) |
| `DB.SearchPaths` | `pkg/cortexdb/graphrag_paths.go:109` |  | TestEveryReadToolIsGatedOrNamed (par l'outil search_paths, porte) |
| `DB.SearchText` | `pkg/cortexdb/text_search.go:168` | * | TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo, TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `DB.SearchTextInCollection` | `pkg/cortexdb/text_search.go:173` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `DB.SearchTextOnly` | `pkg/cortexdb/text_search.go:287` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `DB.SearchWithFacets` | `pkg/cortexdb/aggregate.go:57` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget, TestGatedSearchTrimsToTheBudgetAndReportsAFailure |
| `DB.VectorAggregate` | `pkg/cortexdb/aggregate.go:80` | * | TestTheAggregatesHoldPastTheBoundParameterLimit |
| `GraphRAGToolbox.AggregateMetadata` | `pkg/cortexdb/aggregate.go:111` | * | TestAggregateMetadataCountsVisibleRowsOnly, TestAnUnreadableBrainIsNotAnEmptyBrain |
| `GraphRAGToolbox.BuildContext` | `pkg/cortexdb/graphrag_tool_search.go:309` | * | TestTheChunkToolsLeaveAnExpiredDocumentOut |
| `GraphRAGToolbox.ExpandGraph` | `pkg/cortexdb/graphrag_tool_search.go:183` | * | TestTheChunkToolsLeaveAnExpiredDocumentOut |
| `GraphRAGToolbox.FindKnowledgeGraph` | `pkg/cortexdb/knowledge_graph_tools.go:21` |  | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_graph_find, nommé : the whole graph) |
| `GraphRAGToolbox.FindNodes` | `pkg/cortexdb/graphrag_tool_find_nodes.go:32` |  | TestEveryReadToolIsGatedOrNamed (par l'outil find_nodes, nommé : entities and relations) |
| `GraphRAGToolbox.GetChunks` | `pkg/cortexdb/graphrag_tool_search.go:300` | * | TestTheChunkToolsLeaveAnExpiredDocumentOut |
| `GraphRAGToolbox.GetKnowledge` | `pkg/cortexdb/knowledge_api.go:382` | * | TestTheSweepIsAToolOfTheSharedBrain |
| `GraphRAGToolbox.GetMemory` | `pkg/cortexdb/memory_api.go:602` |  | TestTheSweepDeletesOnlyWhatIsStillExpiredWhenItDeletes |
| `GraphRAGToolbox.GetNodes` | `pkg/cortexdb/graphrag_tool_search.go:290` | * | TestEveryReadToolIsGatedOrNamed (par l'outil get_nodes, porte) |
| `GraphRAGToolbox.GetOntologySchema` | `pkg/cortexdb/ontology_api.go:197` |  | TestEveryReadToolIsGatedOrNamed (par l'outil ontology_get, nommé : no document: ontologies) |
| `GraphRAGToolbox.ListActionTypes` | `pkg/cortexdb/ontology_action_apply.go:639` |  | TestEveryReadToolIsGatedOrNamed (par l'outil ontology_action_list, nommé : no document: ontologies) |
| `GraphRAGToolbox.ListKnowledgeNamespaces` | `pkg/cortexdb/knowledge_graph_tools.go:11` |  | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_graph_namespace_list, nommé : the whole graph) |
| `GraphRAGToolbox.ListOntologySchemas` | `pkg/cortexdb/ontology_api.go:201` |  | TestEveryReadToolIsGatedOrNamed (par l'outil ontology_list, nommé : no document: ontologies) |
| `GraphRAGToolbox.ListRules` | `pkg/cortexdb/rules_api.go:356` |  | TestEveryReadToolIsGatedOrNamed (par l'outil rules_list, nommé : no document: rules) |
| `GraphRAGToolbox.Query` | `pkg/cortexdb/graphrag_tool_types.go:310` | * | TestAnExpiredDocumentIsLeftOutOfCortexQuery |
| `GraphRAGToolbox.QueryKnowledgeGraph` | `pkg/cortexdb/knowledge_graph_tools.go:41` |  | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_graph_find, nommé : the whole graph) |
| `GraphRAGToolbox.RepresentativeRecords` | `pkg/cortexdb/aggregate.go:171` | * | TestAnUnreadableBrainIsNotAnEmptyBrain, TestRepresentativeRecordsNeverPickAHiddenChunk |
| `GraphRAGToolbox.SearchChunksByEntities` | `pkg/cortexdb/graphrag_tool_search.go:136` | * | TestTheChunkToolsLeaveAnExpiredDocumentOut |
| `GraphRAGToolbox.SearchGraphRAGLexical` | `pkg/cortexdb/graphrag_tool_search.go:344` | * | TestEveryReadToolIsGatedOrNamed (par l'outil search_graphrag_lexical, porte) |
| `GraphRAGToolbox.SearchKnowledge` | `pkg/cortexdb/knowledge_api.go:387` | * | TestAnEmptyExpiresAtClearsTheExpiry, TestAnExpiredDocumentIsLeftOutOfSearch, TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo, TestTheGraphWalkLeavesAnExpiredPassageOut |
| `GraphRAGToolbox.SearchMemory` | `pkg/cortexdb/memory_api.go:607` |  | TestEveryReadToolIsGatedOrNamed (par l'outil memory_search, nommé : memories: their own expires_at, as on main) |
| `GraphRAGToolbox.SearchPaths` | `pkg/cortexdb/graphrag_paths.go:114` |  | TestEveryReadToolIsGatedOrNamed (par l'outil search_paths, porte) |
| `GraphRAGToolbox.SearchText` | `pkg/cortexdb/graphrag_tool_search.go:15` | * | TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo, TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `GraphRAGToolbox.SearchVectorRange` | `pkg/cortexdb/vector_range.go:139` | * | TestSearchVectorRangeCountsOnlyVisibleRowsForTruncated, TestSearchVectorRangeLeavesExpiredDocumentsOut |
| `KnowledgeMemory.BuildContextPack` | `pkg/cortexdb/brain.go:133` | * | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_memory_build_context_pack, porte) |
| `KnowledgeMemory.ExpandEntityContext` | `pkg/cortexdb/brain.go:205` | * | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_memory_expand_entity_context, porte) |
| `KnowledgeMemory.Neighbors` | `pkg/cortexdb/brain.go:294` | * | TestNeighborsLeaveExpiredNodesOut |
| `KnowledgeMemory.Recall` | `pkg/cortexdb/brain.go:39` | * | TestEveryReadToolIsGatedOrNamed (par l'outil knowledge_memory_recall, porte) |
| `Quick.Search` | `pkg/cortexdb/quick.go:46` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `Quick.SearchInCollection` | `pkg/cortexdb/quick.go:51` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget (Quick.Search l'appelle avec la collection vide) |
| `Quick.SearchText` | `pkg/cortexdb/quick.go:87` | * | TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo, TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `Quick.SearchTextInCollection` | `pkg/cortexdb/quick.go:92` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |
| `Quick.SearchTextOnly` | `pkg/cortexdb/quick.go:97` | * | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget |

### rpcs : chaque RPC typé (43)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `Backup` | `proto/cortexdb/v1/admin.proto:39` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `CallTool` | `proto/cortexdb/v1/tools.proto:27` |  | TestAConfinedKeySetsAnExpiryButDoesNotSweep (par REST, même liste blanche) ; outils : voir la classe tools |
| `ContractTally` | `proto/cortexdb/v1/contract.proto:124` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `DecisionChain` | `proto/cortexdb/v1/decision.proto:186` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `DeleteKnowledge` | `proto/cortexdb/v1/knowledge.proto:116` |  | hors classe : supprime un document |
| `DeleteKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:260` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `DeleteMemory` | `proto/cortexdb/v1/memory.proto:86` |  | hors classe : mémoire, expires_at de main |
| `DeleteOntologySchema` | `proto/cortexdb/v1/graph.proto:272` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `ExplainInference` | `proto/cortexdb/v1/graph.proto:267` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `ExplainInferenceMatch` | `proto/cortexdb/v1/graph.proto:268` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `ExportKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:262` |  | nommé dans le corps : le graphe entier, comme l'outil équivalent (tour 3, L4) |
| `FindKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:259` |  | nommé dans le corps : le graphe entier, comme l'outil équivalent (tour 3, L4) |
| `GetKnowledge` | `proto/cortexdb/v1/knowledge.proto:114` |  | nommé dans le corps : lecture par identifiant jusqu'au balayage, comme knowledge_get (tour 3, L4) |
| `GetMemory` | `proto/cortexdb/v1/memory.proto:84` |  | hors classe : mémoire, expires_at de main |
| `GetOntologySchema` | `proto/cortexdb/v1/graph.proto:270` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `Health` | `proto/cortexdb/v1/admin.proto:35` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `HybridSearchText` | `proto/cortexdb/v1/graphrag.proto:95` |  | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget (DB.HybridSearchText, options) |
| `ImportKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:261` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `Info` | `proto/cortexdb/v1/admin.proto:36` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `InsertGraphDocument` | `proto/cortexdb/v1/graphrag.proto:90` |  | TestInsertGraphDocumentKeepsTheStoredExpiry (DB.InsertGraphDocument ; pas de champ expires_at) |
| `InsertText` | `proto/cortexdb/v1/graphrag.proto:92` |  | nommé dans le corps : pas de champ expires_at |
| `InsertTextBatch` | `proto/cortexdb/v1/graphrag.proto:93` |  | nommé dans le corps : pas de champ expires_at |
| `ListNamespaces` | `proto/cortexdb/v1/graph.proto:257` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `ListOntologySchemas` | `proto/cortexdb/v1/graph.proto:271` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `ListTools` | `proto/cortexdb/v1/tools.proto:26` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `NeedsAttention` | `proto/cortexdb/v1/contract.proto:129` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `Precedents` | `proto/cortexdb/v1/decision.proto:189` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `QuerySparql` | `proto/cortexdb/v1/graph.proto:263` |  | nommé dans le corps : le graphe entier, comme l'outil équivalent (tour 3, L4) |
| `RecordDecision` | `proto/cortexdb/v1/decision.proto:184` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `RefreshInference` | `proto/cortexdb/v1/graph.proto:265` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `SaveKnowledge` | `proto/cortexdb/v1/knowledge.proto:112` |  | nommé dans le corps : message typé sans expires_at |
| `SaveMemory` | `proto/cortexdb/v1/memory.proto:82` |  | hors classe : mémoire, expires_at de main |
| `SaveOntologySchema` | `proto/cortexdb/v1/graph.proto:269` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `SearchGraphRag` | `proto/cortexdb/v1/graphrag.proto:91` |  | TestSearchGraphRAGLeavesAnExpiredPassageOut (DB.SearchGraphRAG, quatre modes) |
| `SearchKnowledge` | `proto/cortexdb/v1/knowledge.proto:115` |  | TestAnExpiredDocumentIsLeftOutOfSearch (DB.SearchKnowledge) |
| `SearchMemory` | `proto/cortexdb/v1/memory.proto:85` |  | hors classe : mémoire, expires_at de main |
| `SearchText` | `proto/cortexdb/v1/graphrag.proto:94` |  | TestEveryExportedSearchLeavesExpiredDocumentsOutAndFillsTheBudget (DB.SearchText, DB.SearchTextInCollection) |
| `SummarizeInference` | `proto/cortexdb/v1/graph.proto:266` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `UpdateKnowledge` | `proto/cortexdb/v1/knowledge.proto:113` |  | nommé dans le corps : message typé sans expires_at ; l'expiration stockée reste |
| `UpdateMemory` | `proto/cortexdb/v1/memory.proto:83` |  | hors classe : mémoire, expires_at de main |
| `UpsertKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:258` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `UpsertNamespace` | `proto/cortexdb/v1/graph.proto:256` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
| `ValidateShacl` | `proto/cortexdb/v1/graph.proto:264` |  | hors classe : RPC sans document (graphe RDF, ontologie, décision, contrat, administration) |
