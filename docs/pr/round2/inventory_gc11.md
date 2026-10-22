# Inventaire de la PR à part `gc-11-rest-upsert-rows` (trou REST de `main`)

Généré par `tools/inventory.py --class rpcs` sur la tête `9e3ba10` contre `origin/main` (`cc98c6c`), colonne « Test » remplie à la main. La classe de la règle est « chaque route REST qui écrit une ligne nommée par identifiant » ; l'outil n'a pas de classe de routes, d'où la table à la main tirée de `routeAccess` (`pkg/httpapi/access.go`), statuts vérifiés par un essai le 21 octobre 2026 avec une clé confinée à une collection et une clé confinée à un utilisateur.

### Routes REST d'écriture (`routeAccess`, Access: Write)

| Route | Écrit par identifiant | Clé confinée | Test |
| :- | :- | :- | :- |
| `POST /v1/knowledge` | oui (upsert `knowledge_id`) | ligne gardée par ce correctif | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded |
| `POST /v1/memory` | oui (upsert `memory_id`) | ligne gardée par ce correctif | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP |
| `POST /v1/memory`, champ `supersedes` | oui (chaque cible marquée `superseded_by`) | cibles gardées par ce correctif ; absente et étrangère, même 404 | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| `POST /v1/knowledge`, `POST /v1/memory`, identifiant blanc | oui (la façade le garde tel quel) | gardé comme les autres, sur REST et gRPC | TestABlankUpsertIDIsGuardedLikeAnyOther, TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |
| `POST /v1/knowledge`, `POST /v1/memory`, champs `entities`, `relations` | par nom (nœuds partagés) | classe des écritures du graphe, fermée par #73 et #74 (écritures typées tenues à la ligne de leur outil) | hors patch, nommé dans le corps |
| `DELETE /v1/knowledge` | oui | refusée (403 : aucun champ de portée) | hors patch, comportement de `main` (commentaire de `routeAccess`) |
| `DELETE /v1/memory` | oui | refusée (403 : aucun champ de portée) | hors patch, comportement de `main` |
| `POST /v1/graph/entities`, `/relations`, `/triples` | par nom | inatteignable : la politique exige le champ de portée, le décodeur le refuse (400) | hors patch, comportement de `main` |
| `POST /v1/tools/<name>` | selon l'outil | refusée à toute clé confinée sur `main` | hors patch (#74 la câble) |

## Inventaire de `9e3ba10` contre `origin/main` (`*` = fichier touché par le diff)

### rpcs : chaque RPC typé (43)

| Élément | Fichier | Touché | Test |
| :- | :- | :- | :- |
| `Backup` | `proto/cortexdb/v1/admin.proto:39` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `CallTool` | `proto/cortexdb/v1/tools.proto:27` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ContractTally` | `proto/cortexdb/v1/contract.proto:124` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `DecisionChain` | `proto/cortexdb/v1/decision.proto:186` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `DeleteKnowledge` | `proto/cortexdb/v1/knowledge.proto:116` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `DeleteKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:260` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `DeleteMemory` | `proto/cortexdb/v1/memory.proto:86` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `DeleteOntologySchema` | `proto/cortexdb/v1/graph.proto:272` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ExplainInference` | `proto/cortexdb/v1/graph.proto:267` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ExplainInferenceMatch` | `proto/cortexdb/v1/graph.proto:268` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ExportKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:262` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `FindKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:259` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `GetKnowledge` | `proto/cortexdb/v1/knowledge.proto:114` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `GetMemory` | `proto/cortexdb/v1/memory.proto:84` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `GetOntologySchema` | `proto/cortexdb/v1/graph.proto:270` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `Health` | `proto/cortexdb/v1/admin.proto:35` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `HybridSearchText` | `proto/cortexdb/v1/graphrag.proto:95` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ImportKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:261` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `Info` | `proto/cortexdb/v1/admin.proto:36` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `InsertGraphDocument` | `proto/cortexdb/v1/graphrag.proto:90` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `InsertText` | `proto/cortexdb/v1/graphrag.proto:92` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `InsertTextBatch` | `proto/cortexdb/v1/graphrag.proto:93` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ListNamespaces` | `proto/cortexdb/v1/graph.proto:257` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ListOntologySchemas` | `proto/cortexdb/v1/graph.proto:271` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ListTools` | `proto/cortexdb/v1/tools.proto:26` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `NeedsAttention` | `proto/cortexdb/v1/contract.proto:129` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `Precedents` | `proto/cortexdb/v1/decision.proto:189` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `QuerySparql` | `proto/cortexdb/v1/graph.proto:263` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `RecordDecision` | `proto/cortexdb/v1/decision.proto:184` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `RefreshInference` | `proto/cortexdb/v1/graph.proto:265` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `SaveKnowledge` | `proto/cortexdb/v1/knowledge.proto:112` |  | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP (jumeau REST `POST /v1/knowledge`, gardé comme `guardKnowledgeUpsert`) |
| `SaveMemory` | `proto/cortexdb/v1/memory.proto:82` |  | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP (jumeau REST `POST /v1/memory`, gardé comme `guardMemoryUpsert`) |
| `SaveOntologySchema` | `proto/cortexdb/v1/graph.proto:269` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `SearchGraphRag` | `proto/cortexdb/v1/graphrag.proto:91` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `SearchKnowledge` | `proto/cortexdb/v1/knowledge.proto:115` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `SearchMemory` | `proto/cortexdb/v1/memory.proto:85` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `SearchText` | `proto/cortexdb/v1/graphrag.proto:94` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `SummarizeInference` | `proto/cortexdb/v1/graph.proto:266` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `UpdateKnowledge` | `proto/cortexdb/v1/knowledge.proto:113` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `UpdateMemory` | `proto/cortexdb/v1/memory.proto:83` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `UpsertKnowledgeGraph` | `proto/cortexdb/v1/graph.proto:258` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `UpsertNamespace` | `proto/cortexdb/v1/graph.proto:256` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
| `ValidateShacl` | `proto/cortexdb/v1/graph.proto:264` |  | hors classe : RPC typé, port gRPC inchangé ; aucun jumeau REST qui écrive une ligne par identifiant (routes REST ci-dessous) |
