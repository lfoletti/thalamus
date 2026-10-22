# Inventaire de la PR à part `gc-12-grpc-ingest-guard` (trou gRPC de `main`)

Généré par `tools/inventory.py --class rpcs` sur la tête `e68061d` contre `cc98c6c`, colonne « Test » remplie à la main. La classe de la règle est « chaque RPC typé qu'une clé confinée atteint et qui écrit ou lit une ligne nommée par identifiant ». Trouvé par la relecture adverse de #73, tour 2 (B1, `adversarial_review_73_r2.md`), qui a tiré la classe des protos : seuls les écrits ci-dessous portent un champ de portée ou sont dispensés vers leur gestionnaire.

## Inventaire de `e68061d` contre `cc98c6c` (`*` = fichier touché par le diff)

### rpcs : chaque RPC typé (43)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `Backup` | `proto/cortexdb/v1/admin.proto:39` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `CallTool` | `proto/cortexdb/v1/tools.proto:27` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ContractTally` | `proto/cortexdb/v1/contract.proto:124` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `DecisionChain` | `proto/cortexdb/v1/decision.proto:186` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `DeleteKnowledge` | `proto/cortexdb/v1/knowledge.proto:116` |  | efface le document nommé : gardé sur main par le gestionnaire (guardKnowledge), TestEveryIdAddressedMethodIsGuardedByAHandler |
| `DeleteKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:260` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `DeleteMemory` | `proto/cortexdb/v1/memory.proto:86` |  | efface la mémoire nommée : gardé sur main par le gestionnaire (guardMemory), TestEveryIdAddressedMethodIsGuardedByAHandler |
| `DeleteOntologySchema` | `proto/cortexdb/v1/graph.proto:272` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ExplainInference` | `proto/cortexdb/v1/graph.proto:267` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ExplainInferenceMatch` | `proto/cortexdb/v1/graph.proto:268` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ExportKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:262` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `FindKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:259` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `GetKnowledge` | `proto/cortexdb/v1/knowledge.proto:114` |  | lit par identifiant : gardé sur main (guardKnowledge), TestEveryIdAddressedMethodIsGuardedByAHandler |
| `GetMemory` | `proto/cortexdb/v1/memory.proto:84` |  | lit par identifiant : gardé sur main (guardMemory), TestEveryIdAddressedMethodIsGuardedByAHandler |
| `GetOntologySchema` | `proto/cortexdb/v1/graph.proto:270` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `Health` | `proto/cortexdb/v1/admin.proto:35` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `HybridSearchText` | `proto/cortexdb/v1/graphrag.proto:95` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ImportKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:261` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `Info` | `proto/cortexdb/v1/admin.proto:36` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `InsertGraphDocument` | `proto/cortexdb/v1/graphrag.proto:90` |  | écrit le document nommé : gardé par ce correctif, TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection |
| `InsertText` | `proto/cortexdb/v1/graphrag.proto:92` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `InsertTextBatch` | `proto/cortexdb/v1/graphrag.proto:93` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ListNamespaces` | `proto/cortexdb/v1/graph.proto:257` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ListOntologySchemas` | `proto/cortexdb/v1/graph.proto:271` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ListTools` | `proto/cortexdb/v1/tools.proto:26` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `NeedsAttention` | `proto/cortexdb/v1/contract.proto:129` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `Precedents` | `proto/cortexdb/v1/decision.proto:189` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `QuerySparql` | `proto/cortexdb/v1/graph.proto:263` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `RecordDecision` | `proto/cortexdb/v1/decision.proto:184` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `RefreshInference` | `proto/cortexdb/v1/graph.proto:265` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `SaveKnowledge` | `proto/cortexdb/v1/knowledge.proto:112` |  | écrit le document nommé : gardé sur main (guardKnowledgeUpsert), TestAConfinedKeyCannotStealADocumentByUpdatingItsCollection |
| `SaveMemory` | `proto/cortexdb/v1/memory.proto:82` |  | écrit la mémoire nommée : gardé sur main (guardMemoryUpsert), TestAConfinedKeyCannotOverwriteAnothersMemoryByGuessingItsID |
| `SaveOntologySchema` | `proto/cortexdb/v1/graph.proto:269` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `SearchGraphRag` | `proto/cortexdb/v1/graphrag.proto:91` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `SearchKnowledge` | `proto/cortexdb/v1/knowledge.proto:115` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `SearchMemory` | `proto/cortexdb/v1/memory.proto:85` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `SearchText` | `proto/cortexdb/v1/graphrag.proto:94` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `SummarizeInference` | `proto/cortexdb/v1/graph.proto:266` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `UpdateKnowledge` | `proto/cortexdb/v1/knowledge.proto:113` |  | écrit le document nommé : gardé sur main (guardKnowledgeUpsert), TestAConfinedKeyCannotStealADocumentByUpdatingItsCollection |
| `UpdateMemory` | `proto/cortexdb/v1/memory.proto:83` |  | écrit la mémoire nommée : gardé sur main par le gestionnaire (guardMemory), TestEveryIdAddressedMethodIsGuardedByAHandler |
| `UpsertKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:258` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `UpsertNamespace` | `proto/cortexdb/v1/graph.proto:256` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
| `ValidateShacl` | `proto/cortexdb/v1/graph.proto:264` |  | hors classe : n'écrit aucune ligne nommée par identifiant avec un champ de portée ; sans champ de portée, refusé à toute clé confinée par AuthorizeRows (main) |
