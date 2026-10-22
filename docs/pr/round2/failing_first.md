# Second tour : preuve « failing first »

Vérification du 21 octobre 2026 sur les six branches rebasées sur v2.135.0. Pour chaque règle corrigée après la revue du 21 octobre 2026 : le test est écrit avant le code et tombe sans lui (lignes `FAIL` observées au premier passage, consignées ici), puis une mutation à la main qui compile le fait retomber. Script : `mutate_round2.py <jeu> <copie>` (jeux `72`, `73`, `74`, `75`, `76`, `77`), sous le canari (`CORTEXDB_REMOTE=127.0.0.1:1`, `CORTEXDB_GRPC_TOKEN=canary`, `CORTEXDB_PATH=/dev/null/canary`). Les tables sont la sortie du script, sans retouche.

## #72 `gc-04a-authz-tool-rules`

Tests écrits d'abord, tombés sans le code (premier passage, `go test ./pkg/authz/ ./pkg/cortexdb/`) :

```
--- FAIL: TestTheAllowlistTableNamesEveryToolAndEveryIDField      (la table nomme encore plan.filters.document_ids)
--- FAIL: TestReservedKeysAreRefusedInEntityAndRelationMetadataForAConfinedKey   (cinq clés acceptées)
--- FAIL: TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths       (a pinned document id was guarded: knowledge "secret-doc" not found)
--- FAIL: TestAForeignIdAndAMissingIdAnswerTheSameOnEveryRowIdField   (huit champs : le refus dit plus que "not found")
--- FAIL: TestMemoryGetByListHidesRowsOutsideTheCallerScopeAsMissing  (missing [no-such], want [no-such zeus-1] ; served zeus-1 outside the scope)
```

`TestTheRowGuardSeesForeignRowsUnderACallerScope` est vert avant comme après : il tient une règle de conception (la façade ne cache rien par identifiant, R2 de `risques.md`), et tomberait si une correction future confinait `GetMemory` ou `GetKnowledge` par identifiant.

Troisième passage (21 octobre 2026, après la relecture adverse du lot B, tour 3 : B1, B2, B3, N6, L2, L5), tests écrits d'abord, tombés sans le code (`go test ./pkg/authz/`) :

```
--- FAIL: TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey   (entités et relations déclarées acceptées sur trois outils ; document_id, chunk_index, graph_kind acceptés dans les métadonnées d'un document)
--- FAIL: TestPathRetrievalIsRefusedToAConfinedKey                  (return_paths, max_paths, path_max_depth, relation_policies acceptés)
--- FAIL: TestMemoryGetTakesOneFormFromAConfinedKey                 (memory_id et memory_ids ensemble acceptés)
--- FAIL: TestTheRowGuardFailsClosedOnWhatItCannotDecide            (knowledge_update with authz.textNotFoundRows: a failed read was not reported as one: <nil>)
```

`TestAForeignIdAndAMissingIdAnswerTheSameOnEveryRowIdField` s'appelle maintenant `TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField` (L2 : il tient le garde, pas la réponse de bout en bout, que tiennent les tests des ports). `TestReservedKeysAreRefusedInEntityAndRelationMetadataForAConfinedKey` est remplacé par `TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey` : les entités et relations sont refusées en bloc, donc leurs métadonnées aussi. Tête `f1b24eb`, `precommit` OK (0 bloc ; le parcours des éléments d'un tableau, devenu sans chemin `[]` dans la liste, est tenu par un objet glissé dans `memory_ids`).

Quatrième passage (21 octobre 2026, après la relecture adverse de `f1b24eb` : deux bloquants, `docs/pr/round2/adversarial_review_72_r4.md`), tests écrits d'abord, tombés sans le code :

```
--- FAIL: TestAToolThatReadsBothStoresIsGatedForAConfinedKey      (a gate switched off under another spelling passed ; puis null : panic: assignment to entry in nil map)
--- FAIL: TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey (metadata.written_by et metadata.updated_by acceptés sur ingest_document, knowledge_update, memory_update)
--- FAIL: TestTheRowGuardWithholdsADocumentItCannotPlace           (a document with no chunks under a key confined to graphrag_chunks: <nil>)
--- FAIL: TestASupersedeTargetOutsideTheCallerScopeIsNotFound      (a foreign target: <nil>)
```

Sur #73 posé sur `f1b24eb`, le cas `null` de `TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC` fait tomber le paquet : `panic: assignment to entry in nil map` (le bloquant B1, reproduit). `TestAPinToAForeignDocumentFindsNothingOnARealBrain` tient un comportement de `main` que la PR ne change pas : vert avant comme après. Tête `d007fee`, `precommit` OK (0 bloc ; la branche d'erreur du seau de la cible couverte par un seau rendu illisible).

Cinquième passage (21 octobre 2026, après la relecture adverse de `d007fee` : aucun bloquant, `docs/pr/round2/adversarial_review_72_r5.md`). Règles déjà dans le code mais sans test, tenues maintenant : `TestTheRowRulesHoldEveryConfinedField` (scope et namespace sur les deux règles de ligne : quatre mutants survivaient), `TestACopyOfTheScopeFieldInsideAListIsRead` (deux mutants survivaient), les 13 raisons propres nommées dans `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys` (une raison effacée survivait). Code changé : `graphOnlyReason` devient `defaultRefusal`, texte vrai pour les 66 outils. Tête `d805527`, `precommit` OK.

Mutations (`python3 -I mutate_round2.py 72 <copie>`, 21 octobre 2026, cinquième passage, 37 sur 37 tuées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Refus qui nomme la ligne (mémoire) | `pkg/authz/rows.go` : `			return fmt.Errorf("memory %q %w", id, ErrWithheld)` | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField |
| Refus qui nomme la ligne (document) | `pkg/authz/rows.go` : `		return fmt.Errorf("knowledge %q %w", id, ErrWithheld)` | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField |
| document_ids redevenu identifiant | `pkg/authz/allowlist.go` : `		args: paths(queryArgs, walkArgs, planArgs, pprArgs, []string{"top_k"` | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| memory_ids redevenu identifiant | `pkg/authz/allowlist.go` : `"memory_get":    {stores: storeMemory, ids: []idField{{"memory_id", id` | TestAToolCallMayOnlyNameRowsWithinTheScope, TestTheAllowlistTableNamesEveryToolAndEveryIDField, TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField |
| Entités déclarées permises à knowledge_save (graphe partagé réécrit ; relecture adverse du lot B, B2) | `pkg/authz/allowlist.go` : `		args: paths(docArgs, []string{"entities", "entities[].name", "entiti` | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey, TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover, TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys |
| Entités déclarées permises à memory_save (B2) | `pkg/authz/allowlist.go` : `args: paths(memoryArgs, []string{"entities", "entities[].name", "entit` | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey, TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover, TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestTheAllowlistTableNamesEveryToolAndEveryIDField |
| Recherche de chemins permise (return_paths sans collection ; B1) | `pkg/authz/allowlist.go` : `args: paths(queryArgs, walkArgs, planArgs, []string{"top_k", "diversit` | TestPathRetrievalIsRefusedToAConfinedKey |
| Les deux formes de memory_get ensemble (oracle d'existence ; B3) | `pkg/authz/allowlist.go` : `if len(sent) > 2 {` | TestMemoryGetTakesOneFormFromAConfinedKey |
| Clés de chunk du moteur permises dans les métadonnées d'un document (N6) | `pkg/authz/allowlist.go` : `if why, no := engineChunkKeys[lower]; meta && no && false {` | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey |
| Clés de chunk refusées aussi dans les métadonnées d'une mémoire | `pkg/authz/allowlist.go` : `if why, no := engineChunkKeys[lower]; meta && no {` | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey |
| Texte « not found » pris pour une ligne absente (L5) | `pkg/authz/rows.go` : `return errors.Is(err, core.ErrNotFound) || errors.Is(err, sql.ErrNoRow` | TestTheRowGuardFailsClosedOnWhatItCannotDecide |
| Ligne absente prise pour une erreur de lecture | `pkg/authz/rows.go` : `return errors.Is(err, core.ErrNotFound) && errors.Is(err, sql.ErrNoRow` | TestAToolCallMayOnlyNameRowsWithinTheScope, TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths, TestTheRowGuardSeesForeignRowsUnderACallerScope |
| Élément d'un tableau non parcouru (objet glissé dans une liste d'identifiants) | `pkg/authz/allowlist.go` : `if err := walk(path+"[]", elem); err != nil && false {` | TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover |
| Arguments null pris pour un objet (panique de ConfineToolArgs ; relecture adverse tour 4, B1) | `pkg/authz/allowlist.go` : `		return nil, fmt.Errorf("tool arguments must be a JSON object, not nu` | TestAToolThatReadsBothStoresIsGatedForAConfinedKey |
| written_by et updated_by permis dans les métadonnées (tour 4, B2) | `pkg/authz/allowlist.go` : `if why, no := writerKeys[lower]; meta && no && false {` | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey |
| Porte lue sous une seule orthographe (tour 4, L2) | `pkg/authz/allowlist.go` : `		}` | TestAToolThatReadsBothStoresIsGatedForAConfinedKey |
| Porte absente prise pour ouverte | `pkg/authz/allowlist.go` : `	return seen || true` | TestAToolThatReadsBothStoresIsGatedForAConfinedKey |
| Document sans chunks placé dans la collection par défaut (tour 4, L3) | `pkg/authz/rows.go` : `if !s.PermitsKnowledge(resp.Knowledge.Collection) {` | TestTheRowGuardWithholdsADocumentItCannotPlace |
| Cible supersedes hors de la portée de l'appelant acceptée par la façade (tour 4, L1) | `pkg/cortexdb/memory_api.go` : `			return fmt.Errorf("supersede %q: %w", target, core.ErrNotFound)` | TestASupersedeTargetOutsideTheCallerScopeIsNotFound |
| Seau illisible de la cible pris pour un « not found » | `pkg/cortexdb/memory_api.go` : `		if err != nil && false {` | TestASupersedeTargetOutsideTheCallerScopeIsNotFound |
| Cible étrangère refusée avec un autre libellé que la cible absente | `pkg/cortexdb/memory_api.go` : `return fmt.Errorf("supersede %q: %w", target, errMemoryBucketWithheld)` | TestASupersedeTargetOutsideTheCallerScopeIsNotFound |
| Champ scope d'une ligne de mémoire non lu (relecture adverse tour 5, L1) | `pkg/authz/rows.go` : `true &&` | TestTheRowRulesHoldEveryConfinedField |
| Champ namespace d'une ligne de mémoire non lu (tour 5, L1) | `pkg/authz/rows.go` : `true &&` | TestTheRowRulesHoldEveryConfinedField |
| Clé confinée sur scope admise sur une ligne de connaissance (tour 5, L1) | `pkg/authz/rows.go` : `true &&` | TestTheRowRulesHoldEveryConfinedField |
| Clé confinée sur namespace admise sur une ligne de connaissance (tour 5, L1) | `pkg/authz/rows.go` : `}` | TestTheRowRulesHoldEveryConfinedField |
| Raison propre d'un outil refusé effacée (tour 5, L2) | `pkg/authz/allowlist.go` : `(ligne retirée)` | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys |
| Outil nouveau admis au lieu d'être refusé par défaut (tour 5, L2) | `pkg/authz/allowlist.go` : `confinedTools[name] = toolRule{}` | TestAConfinedKeyIsRefusedAToolOffTheAllowlist, TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys |
| Copie du champ de portée dans une liste non lue (tour 5, N2) | `pkg/authz/allowlist.go` : `			out = append(out, jsonValuesIn(elem, field, depth)...)` | TestACopyOfTheScopeFieldInsideAListIsRead |
| Une liste dépense la profondeur (tour 5, N2) | `pkg/authz/allowlist.go` : `out = append(out, jsonValuesIn(elem, field, depth-1)...)` | TestACopyOfTheScopeFieldInsideAListIsRead |
| Forme unique de memory_get absente du tableau | `pkg/authz/allowlist.go` : `if len(r.oneOf) > 9 {` | TestTheAllowlistTableNamesEveryToolAndEveryIDField |
| Préfixe _ accepté | `pkg/authz/allowlist.go` : `if meta && strings.HasPrefix(lower, "_") && false {` | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey, TestAFilterMayNameAnEngineKey |
| Cible vide ou mémoire elle-même prise pour une cible manquante | `pkg/cortexdb/memory_api.go` : `		// A target outside` | TestAMissingSupersedeTargetLeavesNoNewMemory |
| Cible de supersedes absente vérifiée après l'écriture (oracle par écriture partielle ; relecture adverse) | `pkg/cortexdb/memory_api.go` : `(ligne retirée)` | TestAMissingSupersedeTargetLeavesNoNewMemory, TestASupersedeTargetOutsideTheCallerScopeIsNotFound |
| Un champ d'identifiant perdu par la liste blanche (le plancher < 8 le laissait passer ; relecture adverse) | `pkg/authz/allowlist.go` : `"memory_update": {stores: storeMemory, args: []string{"content", "impo` | TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField |
| Préfixe _ refusé aussi dans un filtre (th_pending lit _review_status ; trouvé par tests/e2e_gateway.sh au niveau 3) | `pkg/authz/allowlist.go` : `meta := true` | TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey, TestAFilterMayNameAnEngineKey |
| Liste de memory_get sans portée | `pkg/cortexdb/memory_api.go` : `			}` | TestMemoryGetByListHidesRowsOutsideTheCallerScopeAsMissing |
| Seau illisible pris pour lisible | `pkg/cortexdb/memory_api.go` : `				ok = true; _ = err` | TestMemoryGetByListHidesRowsOutsideTheCallerScopeAsMissing |

Note : la huitième mutation a d'abord survécu. Le test rendait la table `sessions` illisible par un renommage, mais la lecture de la ligne elle-même (jointure sur `sessions`) tombait avant la porte, et l'erreur venait d'ailleurs. Le test rend maintenant `created_at` illisible (`UPDATE sessions SET created_at = 'not a time'`) : la ligne se charge, le seau non, et la mutation tombe. Un test qui tombe pour une autre raison que la règle visée ne prouve rien : la mutation l'a révélé.

Classe `ids` (D6) : les champs d'identifiant de la liste blanche sont lus par `TestTheGuardWithholdsAForeignIdAndPassesAMissingOneOnEveryRowIdField` depuis la table elle-même (`confinedTools[*].ids`), donc un champ ajouté plus tard est tenu sans test ajouté ; le test exige exactement 11 champs. Les listes : `memory_get.memory_ids` (donnée, `missing` par la façade), `plan.filters.document_ids` sur six outils (donnée, filtre de collection forcé ; `TestAPinToAForeignDocumentFindsNothingOnARealBrain` en joue trois, `knowledge_search`, `search_text` et `search_graphrag_lexical` ; le rappel et le paquet de contexte ne sont pas joués par ce test), `memory_save.supersedes` (identifiant, refus `not found`).

## #75 `gc-08-content-hash`

Tests écrits d'abord, tombés sans le code (21 octobre 2026) : `TestATitleOnlyReingestKeepsTheGraphTheCallerAttached` (le graphe attaché retiré par une ré-ingestion du même texte), `TestAContentChangeReplacesTheChunksAndKeepsTheEntities` (première forme : entités retirées ; seconde forme, après `make diffmain` : les mentions du chunk 0 perdues, `edges has_chunk=1,loves=1`), `TestUnchangedIsSaidOnlyWhenTheChunkNodesAreThere`, `TestTwoConcurrentDedupeIngestsStoreOneCopy`, `TestDedupeFindsTheCopyPastSixteenHolders`, `TestTheIngestMarksNeverOverwriteACallersKey`. Le cas « chunk rows kept » du test des pannes couvre un chemin d'erreur (déclencheur sur `embeddings`), sans règle nouvelle.

Mutations (`python3 -I mutate_round2.py 75 <copie>`, 21 octobre 2026, tête `1e5ead6`, sous `CORTEXDB_TEST_POSTGRES`) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Ré-ingestion du même texte traitée comme un autre texte | `pkg/cortexdb/graphrag_tool_ingest.go` : `if sameText(existing, metadata) && storedIn == ingestOpts.Collection &` | TestAContentChangeReplacesTheChunksAndKeepsTheEntities, TestAnIdenticalReingestWritesNothing, TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking, TestTheChunkIDsOfAnUnchangedReingestComeInChunkOrder, TestTheIngestMarksNeverOverwriteACallersKey, TestTheIngestPathsHoldOnPostgres, TestUnchangedIsSaidOnlyWhenTheChunkNodesAreThere **(attendu non tombé : TestATitleOnlyReingestKeepsTheGraphTheCallerAttached)** |
| Autre texte écrit en place, anciens chunks gardés | `pkg/cortexdb/graphrag_tool_ingest.go` : `		}` | TestAContentChangeReplacesTheChunksAndKeepsTheEntities, TestAnIngestWritesEverythingOrNothing, TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking |
| Autre texte retire tout le graphe (comportement du tour 1) | `pkg/cortexdb/graphrag_tool_ingest.go` : `		}` | TestAContentChangeReplacesTheChunksAndKeepsTheEntities |
| Autre texte qui retire aussi les chunks survivants, avec leur graphe (chunk i reste chunk i sur main) | `pkg/cortexdb/knowledge_tx.go` : `(ligne retirée)` | TestAContentChangeReplacesTheChunksAndKeepsTheEntities |
| knowledge_save sur un identifiant existant qui ne retire rien (survivant gremlins du niveau 3) | `pkg/cortexdb/knowledge_api.go` : `		retire = retireGraph` | TestAKnowledgeSaveOverAnExistingIdRetiresWhatTheOldTextLeft |
| « Unchanged » sans vérifier les nœuds | `pkg/cortexdb/graphrag_tool_ingest.go` : `if present || true {` | TestUnchangedIsSaidOnlyWhenTheChunkNodesAreThere |
| Nœuds comptés sans comparer au nombre attendu | `pkg/cortexdb/graphrag_tool_ingest.go` : `return len(nodes) >= 0, nil` | TestUnchangedIsSaidOnlyWhenTheChunkNodesAreThere |
| Dédup sans verrou sur PostgreSQL (CORTEXDB_TEST_POSTGRES requis) | `pkg/cortexdb/knowledge_tx.go` : `_ = sqldialect.Postgres` | TestADedupeIngestThatCannotLockTheHashIsReported, TestTheIngestPathsHoldOnPostgres |
| Dédup plafonnée au premier détenteur, toutes collections | `pkg/cortexdb/knowledge_tx.go` : `AND ? <> ''` | TestDedupeFindsTheCopyPastSixteenHolders, TestDedupeIsDecidedWithinTheCollection |
| Dédup ignorée | `pkg/cortexdb/knowledge_tx.go` : `			return duplicateContent{holder: holder}` | TestDedupeFindsTheCopyPastSixteenHolders, TestDedupeIsDecidedWithinTheCollection, TestDedupeRefusesTheSameContentUnderAnotherID, TestTheContentHashFollowsTheContentThroughKnowledgeWrites, TestTheIngestPathsHoldOnPostgres, TestTwoConcurrentDedupeIngestsStoreOneCopy |
| Marques sous les noms du client | `pkg/cortexdb/graphrag_tool_ingest.go` : `contentHashKey  = "content_sha256"` | TestTheIngestMarksNeverOverwriteACallersKey |
| Copie forgée d'une marque gardée | `pkg/cortexdb/graphrag_tool_ingest.go` : `	}` | TestTheIngestMarksNeverOverwriteACallersKey |

Niveau 3, `gremlins` (21 octobre 2026) : #72 et #74 aucun survivant ; #73 « NOT COVERED » `pkg/rpcserver/ownership.go:194` (`case err == nil`), rejoué à la main (`err != nil`) : `TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC`, `TestAConfinedKeyCannotReachAnotherRowByIDThroughTheTools`, `TestTheCallToolChecksFailClosedOnTheirOwn` tombent. Trou trouvé par `tests/e2e_gateway.sh` : le préfixe `_` refusé aussi dans les filtres (`th_pending`) ; test `TestAFilterMayNameAnEngineKey` tombé avant le code, mutation `meta := true` (jeu `72`, 9 mutations au second passage).

### Troisième passage (rebasage sur v2.136.0, 22 octobre 2026, tête `ebe5a42`)

Rebasage sur `origin/main` à v2.136.0 (`0328e0a`, qui contient #72). Le test de complétude de la liste blanche tombait : `ingest_document: the allowlist classifies` (`dedupe` absent). `dedupe` classé comme donnée sur `ingest_document`. Inventaire de la tête : `inventory_75.md` (classes `tools`, `reads`, `writes`, `keys`, `promises`, 203 sur 203 avec un test ou une raison). La classe `writes` a montré un chemin sans test : `InsertGraphDocument` écrit l'enregistrement sans métadonnées, donc sans marques. Test ajouté : `TestADocumentWrittenWithoutTheMarksIsNeitherUnchangedNorAHolder`. `diffmain` : une seule différence, celle que le corps nomme (point 1, chunks au-delà de la fin).

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `75port` | `dedupe` hors de la ligne d'`ingest_document` | `args: docArgs, …` | TestAConfinedKeyDedupesWithinItsCollectionOnly (403 « takes no dedupe argument ») |
| `75port` | Recherche du doublon dans toutes les collections | `AND (c.name = ? OR 1=1)` | TestAConfinedKeyDedupesWithinItsCollectionOnly |
| `75marks` | `InsertGraphDocument` qui garde les métadonnées d'une ingestion antérieure | ligne `existing.Metadata = doc.Metadata` retirée | TestADocumentWrittenWithoutTheMarksIsNeitherUnchangedNorAHolder |

Mutant retiré, équivalent : `sameText` qui accepte un document sans hash. `sameIngest` compare aussi les métadonnées, marques comprises, donc le résultat ne change pas.

Jeu `75` rejoué sur la tête rebasée sous `CORTEXDB_TEST_POSTGRES` : 12 sur 12. Un premier passage a donné 11 sur 12 et un arbre rouge : la commande prenait le DSN du `Makefile` avec le texte `$(USER)` non remplacé, donc chaque test PostgreSQL tombait. Commandes :

```
CORTEXDB_TEST_POSTGRES="postgres://$USER@127.0.0.1:5432/cortexdb_test?sslmode=disable" python3 -I docs/pr/round2/mutate_round2.py 75 <copie>
python3 -I docs/pr/round2/mutate_round2.py 75port <copie>
python3 -I docs/pr/round2/mutate_round2.py 75marks <copie>
```

### Quatrième passage (relecture adverse r1, 22 octobre 2026, tête `dcdccbf`)

Relecture adverse de `ebe5a42` : `adversarial_review_75_r1.md`. Tests écrits d'abord, tombés sans le code :

```
--- FAIL: TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces (gRPC et REST : 403 « takes no metadata._content_sha256 from a confined key », main répond 200)
--- FAIL: TestTwoConcurrentDedupeIngestsStoreOneCopy             (réécrit : 25 tours de quatre écrivains ; sous M1, recherche dans sa propre transaction, deux copies stockées, 6 fois sur 6)
```

Survivants de la relecture, tenus maintenant (tests verts sur le code, tombés sous la mutation) : M4 `TestUnchangedNeedsEveryChunkRow`, M8 `TestDedupeNamesTheSmallestHolder`, M9 `TestARetiredChunkIsRecordedAsAReingest`, M10, M11 et M12 `TestTheMarksDescribeTheStoredChunks`. M3 : `sameText` retiré, code mort (`sameIngest` compare les marques). Les contrôles de comptes du graphe comparent des valeurs exactes (`hasCount`). Le mutant de l'index manquait à la liste : il est au jeu `75`.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `75race` | Recherche du doublon hors de la transaction d'écriture (M1) | recherche dans sa propre transaction | TestTwoConcurrentDedupeIngestsStoreOneCopy |
| `75r1` | gRPC : marques rendues à la façade | `marked(m)` retiré de `SaveKnowledge` | TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r1` | REST : marques rendues à la façade | boucle sur `EngineDocumentKeys` retirée | TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r1` | « Unchanged » sans compter les lignes de chunk (M4) | `len(chunkIDs) == len(chunks)` retiré | TestUnchangedNeedsEveryChunkRow |
| `75r1` | Détenteur le plus grand (M8) | `MIN(d.id)` → `MAX(d.id)` | TestDedupeNamesTheSmallestHolder |
| `75r1` | Producteur de l'historique (M9) | `producerDocumentReingest` → `producerDocumentCleanup` | TestARetiredChunkIsRecordedAsAReingest |
| `75r1` | Marques de `knowledge_update` au découpage par défaut (M10) | taille 0 | TestTheMarksDescribeTheStoredChunks |
| `75r1` | Marques de `knowledge_save` au découpage par défaut (M11) | taille 0 | TestTheMarksDescribeTheStoredChunks |
| `75r1` | Hash absent du nœud de document (M12) | ligne retirée | TestTheMarksDescribeTheStoredChunks |
| `75r1` | Une marque oubliée dans les clés rendues à la façade | `chunkOverlapKey` retiré d'`EngineDocumentKeys` | TestEngineDocumentKeysNameEveryKeyTheEngineWrites, TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |

Total sur `dcdccbf` sous `CORTEXDB_TEST_POSTGRES` : jeux `75` (12), `75port` (2), `75marks` (1), `75race` (1), `75r1` (9), 25 sur 25 tombées. Les deux motifs de retrait du jeu `75` suivent le nouvel ordre du bloc « unchanged » (`if sameIngest(` en premier). Commande : `for s in 75 75port 75marks 75race 75r1; do CORTEXDB_TEST_POSTGRES=… python3 -I docs/pr/round2/mutate_round2.py $s <copie>; done`.

### Cinquième passage (relecture adverse r2, partielle, 22 octobre 2026, tête `53121d1`)

Relecture partielle de `dcdccbf` (différence `ebe5a42..dcdccbf`) : `adversarial_review_75_r2.md`. Point bloquant A (route des outils : un document écrit par `ingest_document`, relu et renvoyé, refusé là où `main` répond 200) corrigé par la règle `rewritten` de la liste blanche sur les trois outils de document ; `marked` retiré des ports, `pkg/httpapi/policy.go` revenu à `main`. Survivants E, G, D, H tués ; point 6 du tour 1 (`chunk_node_ids: null`) corrigé ; remarques D et F corrigées. Tests tombés sans le code : `TestAToolWriteTakesTheMarksTheFacadeReplaces` (REST et gRPC, 403 sur `ingest_document` et `knowledge_update`), `TestAnotherChunkingOfTheSameCountIsNotUnchanged`, `TestTheMarksDescribeTheStoredChunks` (recouvrement 1 et 2), `TestDedupeNamesTheSmallestHolder` (`"chunk_node_ids":[]`). `879d864` : `patchcov` montrait la règle `rewritten` sans test dans `pkg/authz` (les tests des ports ne comptent pas pour ce paquet) : `TestTheRewrittenMarksPassOnTheDocumentWritesOnly` ajouté, tête `53121d1`.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `75r2` | Marques refusées sur `ingest_document` (bloquant A) | `rewritten` retiré | TestAToolWriteTakesTheMarksTheFacadeReplaces, TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r2` | Marques refusées sur `knowledge_update` | `rewritten` retiré | TestAToolWriteTakesTheMarksTheFacadeReplaces, TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r2` | Autre graphie d'une marque acceptée | clé comparée en minuscules | TestAToolWriteTakesTheMarksTheFacadeReplaces, TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r2` | `InsertGraphDocument` garde une marque envoyée | refus retiré | TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r2` | Taille de chunk de l'appelant gardée (E) | marque de l'appelant gardée | TestAToolWriteTakesTheMarksTheFacadeReplaces, TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r2` | Recouvrement de l'appelant gardé (G) | marque de l'appelant gardée | TestAToolWriteTakesTheMarksTheFacadeReplaces, TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r2` | `sameIngest` sans le découpage (D) | marques de découpage ignorées | TestAnotherChunkingOfTheSameCountIsNotUnchanged |
| `75r2` | `knowledge_save` marque un recouvrement nul (H) | recouvrement 0 | TestTheMarksDescribeTheStoredChunks |
| `75r2` | Doublon qui rend `chunk_node_ids: null` | liste nulle | TestDedupeNamesTheSmallestHolder |
| `75r3` | Marques jamais reprises par la liste blanche | règle désactivée | TestTheRewrittenMarksPassOnTheDocumentWritesOnly |
| `75r3` | Marques reprises sur une écriture de mémoire | `r.rewritten` ignoré | TestTheRewrittenMarksPassOnTheDocumentWritesOnly |

Total sur `53121d1` sous `CORTEXDB_TEST_POSTGRES` : `75` (12), `75port` (2), `75marks` (1), `75race` (1), `75r1` (7 : les deux entrées de `marked`, retiré des ports, sont parties), `75r2` (9), `75r3` (2), soit 34 sur 34. `precommit` vert (0 bloc), `invcheck` 203 sur 203, `diffmain` : seule la différence nommée du point 1. Rejouer : `python3 -I docs/pr/round2/mutate_round2.py 75r3 <copie>`.

### Sixième passage (relecture adverse r3, partielle, 22 octobre 2026, têtes `025c490` puis `de70c73`)

Relecture partielle de `53121d1` (différence `dcdccbf..53121d1`) : `adversarial_review_75_r3.md` (aucun point bloquant ; A à F et le point 6 du tour 1 fermés). Correctifs : G1, `InsertGraphDocument` refusait sur le port gRPC la métadonnée d'un document relu, que `main` accepte : la façade retire maintenant la copie des marques de l'appelant des métadonnées des chunks, et le refus du port est retiré (`ownership.go` revient à `main`, plus un commentaire) ; G2, un cas par clé de découpage, même nombre de chunks ; G3, les tests de `pkg/authz` et du port gRPC bouclent sur `EngineDocumentKeys()`. Tests tombés sans le code :

```
--- FAIL: TestInsertGraphDocumentDropsACallersCopyOfTheMarks          (the chunk kept the caller's _content_sha256, _chunk_size, _chunk_overlap)
--- FAIL: TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces       (gRPC : InsertGraphDocument refused the metadata it read, trois clés)
```

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `75r4` | `InsertGraphDocument` garde la copie des marques (G1) | `if false && …` | TestInsertGraphDocumentDropsACallersCopyOfTheMarks |
| `75r4` | `InsertGraphDocument` retire aussi les clés de l'appelant | `|| k == "team"` | TestInsertGraphDocumentDropsACallersCopyOfTheMarks |
| `75r4` | Port gRPC : marque refusée sur `InsertGraphDocument` | refus remis | TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces |
| `75r4` | `sameIngest` sans le recouvrement (G2, N5) | clé ignorée | TestAnotherChunkingOfTheSameCountIsNotUnchanged |
| `75r4` | `sameIngest` sans la taille (G2, N9) | clé ignorée | TestAnotherChunkingOfTheSameCountIsNotUnchanged |
| `75r4` | Liste blanche : première marque seule reprise (G3, N4) | `EngineDocumentKeys()[:1]` | TestTheRewrittenMarksPassOnTheDocumentWritesOnly |

L'entrée « `InsertGraphDocument` qui garde une marque envoyée » du jeu `75r2` est retirée : le refus qu'elle désactivait n'existe plus. Relecture partielle r4 de `025c490` (`adversarial_review_75_r4.md`, aucun point bloquant) : G1 à G3 fermés ; G4, le test de la façade ne tenait que `team` parmi les clés gardées (sa mutation M2, retirer toute clé en `_`, survivait) : il tient maintenant `_table` et une autre graphie d'une marque ; G5, le message du commit disait encore que `InsertGraphDocument` refuse les marques : réécrit ; « Breaking Changes » nomme la clé d'un appelant remplacée ou retirée. La relecture r4 de #76 (F2) a vu que rien ne lisait l'enregistrement après une clé forgée envoyée à `InsertGraphDocument` (sa mutation M4, l'enregistrement qui prend la métadonnée de l'appelant, survivait) : le test lit maintenant l'enregistrement. Tête `f4c5aa7`, jeu `75r4` porté à 8. Total sur `f4c5aa7` sous `CORTEXDB_TEST_POSTGRES` : `75` (12), `75port` (2), `75marks` (1), `75race` (1), `75r1` (7), `75r2` (8), `75r3` (2), `75r4` (8), soit 41 sur 41. Rejouer : `python3 -I docs/pr/round2/mutate_round2.py 75r4 <copie>`.

## #76 `gc-05-expires-at`

Tête `c7a620d` (sur `f5bc6ff`). Tests écrits avant le code : au premier passage, `TestAClientKeyNamedExpiresAtIsStoredAsIsAndNeverRead` et `TestTheEngineKeyInClientMetadataIsDropped` tombaient sur la clé `expires_at` lue comme expiration (tour 1), les quatre tests des chemins (`search_vector_range`, `knowledge_memory_neighbors`, `aggregate_metadata`, `representative_records`) sur les lignes servies, `TestASecondProcessSeesAnExpiryAtOnce` sur le cache de 15 s, `TestTheSweepDeletesOnlyWhatIsStillExpiredWhenItDeletes` sur la suppression sans revérification. Mutations (`CORTEXDB_TEST_POSTGRES=… python3 -I mutate_round2.py 76 <copie>`, 21 octobre 2026, 15 sur 15 tombées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Clé du client lue comme une expiration | `pkg/cortexdb/expiry.go` : `const expiresAtKey = "expires_at"` | TestAClientKeyNamedExpiresAtIsStoredAsIsAndNeverRead, TestARowTheSweepCannotReadIsReported, TestTheDocumentExpiryIsIndexed |
| Copie du client de la clé du moteur gardée | `pkg/cortexdb/expiry.go` : `normalized, err := normalizeExpiresAt(explicit)` | TestTheEngineKeyInClientMetadataIsDropped, TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly |
| Champ explicite ignoré par knowledge_save | `pkg/cortexdb/knowledge_api.go` : `stampWriter(ctx, metadata, plainString)` | TestABadExpiresAtIsRefused, TestAggregateMetadataCountsVisibleRowsOnly, TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo, TestNeighborsLeaveExpiredNodesOut, TestRepresentativeRecordsNeverPickAHiddenChunk, TestSearchVectorRangeLeavesExpiredDocumentsOut, TestTheChunkToolsLeaveAnExpiredDocumentOut, TestTheEngineKeyInClientMetadataIsDropped, TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly, TestTheGraphWalkLeavesAnExpiredPassageOut |
| Expiration non reportée par un knowledge_update qui ne dit rien | `pkg/cortexdb/expiry.go` : `if value, ok := stored[expiresAtKey].(string); ok && value != "" && fa` | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly |
| Chaîne vide qui ne retire pas l'expiration | `pkg/cortexdb/expiry.go` : `		return setExpiry(metadata, *explicit)` | TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly |
| search_vector_range sans la porte | `pkg/cortexdb/vector_range.go` : `(ligne retirée)` | TestSearchVectorRangeLeavesExpiredDocumentsOut |
| knowledge_memory_neighbors sans la porte | `pkg/cortexdb/brain.go` : `(ligne retirée)` | TestNeighborsLeaveExpiredNodesOut |
| aggregate_metadata sans l'exclusion | `pkg/cortexdb/aggregate.go` : `	return db.store.Aggregate(ctx, req)` | TestAggregateMetadataCountsVisibleRowsOnly |
| representative_records sans l'exclusion | `pkg/cortexdb/aggregate.go` : `	return db.store.VectorAggregate(ctx, req)` | TestRepresentativeRecordsNeverPickAHiddenChunk |
| Ligne sans document exclue avec les cachées (NULL NOT IN) | `pkg/core/aggregations.go` : `query += " AND doc_id NOT IN (" + strings.Join(placeholders, ", ") + "` | TestAggregateMetadataCountsVisibleRowsOnly, TestBothStoresLeaveExcludedDocumentsOutOfAggregates |
| Suppression sans revérification de l'expiration | `pkg/cortexdb/knowledge_helpers.go` : ``DELETE FROM documents WHERE id = ?`, knowledgeID)` | TestTheSweepDeletesOnlyWhatIsStillExpiredWhenItDeletes |
| Porte gardée en mémoire d'un appel à l'autre (le cache du tour 1) | `pkg/cortexdb/expiry.go` : `}` | TestASecondProcessSeesAnExpiryAtOnce, TestAnEmptyExpiresAtClearsTheExpiry, TestTheExpiryIsSetAndRemovedByTheExplicitFieldOnly |
| Dédup qui compte un détenteur expiré | `pkg/cortexdb/knowledge_tx.go` : `(ligne retirée)` | TestDedupeLooksUpThroughTheSearchGate |
| Test du balayage qui lit le serveur de la machine | `cmd/cortexdb-mcp-stdio/sweep_expired_test.go` : `(ligne retirée)` | TestTheSweepCommandReportsAndRemovesExpiredRows |
| Prédicat sur une autre expression que celle de l'index (un parcours complet) | `pkg/cortexdb/expiry.go` : `	return `(` + expr + ` > '' AND ` + expr + ` <= ?)`` | TestTheDocumentExpiryIsIndexed, TestTheGraphWalkLeavesAnExpiredPassageOut |

La quinzième mutation du premier jet (`IS NOT NULL AND <= ?`) ne tombait pas : `EXPLAIN QUERY PLAN` montre que SQLite sert encore cette forme depuis l'index d'expression (`SEARCH … USING INDEX idx_documents_expires_at (<expr><?)`) ; ce n'était pas une mutation. Remplacée par un prédicat sur `json_extract(metadata, '$._expires_at')` sans la garde `json_valid`, une expression autre que celle de l'index : parcours complet, `TestTheDocumentExpiryIsIndexed` tombe. Le commentaire du code a été corrigé dans le même sens (c'est l'expression qui compte, pas la forme de la comparaison).

Niveau 3, `gremlins` sur #75 : survivant `pkg/cortexdb/knowledge_api.go:63` (`existing != nil` → `retireGraph`), réel : test `TestAKnowledgeSaveOverAnExistingIdRetiresWhatTheOldTextLeft` ajouté (sous la mutation : 3 lignes de chunk, nœuds `chunk=3`), mutation au jeu `75` (12). Sur #76 : survivant `knowledge_api.go:120:39`, la clause `writerOf(ctx) != ""` de `main`, étendue sans changement de sens, laissée sans test.

Niveau 3, `gremlins` sur #76, paquet `main` (`cmd/cortexdb-mcp-stdio/sweep_expired.go`), sept « LIVED » (`gremlins` n'y tue rien) rejoués à la main : lignes 42 (`dbPath == ""`), 46 (`openBrainDB` en erreur), 53 (`report != nil`), 56 (balayage en erreur) → `TestTheSweepCommandReportsAndRemovesExpiredRows` ; lignes 67 (`dialCortexDB`), 76 (`CallTool`), 81 (`json.Unmarshal`) → `TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet`, `TestAnUnreadableReportFromTheSharedBrainIsAFailure`. « NOT COVERED » `main.go:204` et `sweep_expired.go:91` : lignes de `main()` déclarées et constante de délai.

### Rebasage sur v2.136.0 (22 octobre 2026, tête `9dc1016`, sur #75 `dcdccbf`)

Test de complétude de la liste blanche tombé au rebasage (`expires_at` et `sweep_expired` non classés) : `expires_at` classé comme donnée, `sweep_expired` refusé. Tests nouveaux, tombés sans le code : `TestAConfinedKeySetsAnExpiryButDoesNotSweep`, `TestEveryReadToolIsGatedOrNamed` (`get_nodes` rendait le nœud du document expiré), `TestATypedKnowledgeWriteTakesTheExpiryKeyTheFacadeReplaces` (gRPC et REST : 403 sur `metadata._expires_at` lu puis renvoyé). `TestSearchGraphRAGLeavesAnExpiredPassageOut` tient une entrée exportée sans test nommé, vert sur le code, tombé sous la mutation des graines.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `76port` | `expires_at` hors de la ligne des documents | retiré de `docArgs` | TestAConfinedKeySetsAnExpiryButDoesNotSweep |
| `76port` | `sweep_expired` sans raison propre | entrée de `refusedTools` retirée | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys |
| `76port` | `sweep_expired` ouvert à une clé confinée | outil ajouté à `confinedTools` | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys |
| `76class` | `get_nodes` sans la porte | ligne retirée | TestEveryReadToolIsGatedOrNamed |
| `76class` | `expand_graph` sans la porte | sous-graphe servi tel quel | TestEveryReadToolIsGatedOrNamed |
| `76class` | Graines de `SearchGraphRAG` sans la porte | `gatedSearch` contourné | TestSearchGraphRAGLeavesAnExpiredPassageOut |
| `76rt` | `_expires_at` hors des clés rendues à la façade | retiré de `EngineDocumentKeys` | TestATypedKnowledgeWriteTakesTheExpiryKeyTheFacadeReplaces |

Jeu `76` rejoué sur `c80797b` (code identique sur `9dc1016`) sous `CORTEXDB_TEST_POSTGRES` : 25 sur 25. Avec `76port` (3), `76class` (3), `76rt` (1) : 32 sur 32. Inventaire : `inventory_76.md`, 322 sur 322 (`inventory.py` écrit maintenant entre accents graves doubles un nom qui contient un accent grave).

### Relecture adverse r1 (22 octobre 2026, tête `8cfbc13`, sur #75 `879d864`)

Relecture complète de `9dc1016` : `adversarial_review_76_r1.md`. Points corrigés, chacun avec un test écrit avant le code et tombé sans lui (fichiers posés sur `17983c4` sans le correctif) : B1 (`InsertGraphDocument`), B2 (`knowledge_memory_shortest_path`, provenance des arêtes de `search_paths` et `expand_graph` ; le test de classe vérifie maintenant que chaque outil « porte » rend le document vivant, et tombait sur les deux), L1 (filtres du plan après la porte, test sur `dcdccbf` vert, sur la tête rouge avant le correctif), L2, L3, L5 et deux remarques (`knowledge_delete` d'un enregistrement disparu, expiration vide écrite à la main). L4 et les six outils d'entités : nommés dans le corps. Code mort `knowledge_api.go:124` retiré. Les tests des marques de #75 sautent `_expires_at` par son nom (un document sans expiration n'en porte pas ; aller-retour couvert par `TestATypedKnowledgeWriteTakesTheExpiryKeyTheFacadeReplaces`) : ils étaient rouges sur `17983c4`, amendé sans la suite complète.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `76r3` | `InsertGraphDocument` garde l'expiration stockée | `recordOnlyKey` ignoré | TestInsertGraphDocumentKeepsTheStoredExpiry |
| `76r3` | `knowledge_memory_shortest_path` sans la porte | contrôle des nœuds désactivé | TestEveryReadToolIsGatedOrNamed |
| `76r3` | provenance des chemins de `search_paths` | ligne retirée | TestEveryReadToolIsGatedOrNamed |
| `76r3` | provenance des arêtes de `expand_graph` | arêtes rendues telles quelles | TestEveryReadToolIsGatedOrNamed |
| `76r3` | `source_chunk_id` non lu | branche désactivée | TestAnEdgeDoesNotNameAnExpiredChunkOrDocument |
| `76r3` | `source_chunk_id` d'un chunk expiré gardé | suppression désactivée | TestAnEdgeDoesNotNameAnExpiredChunkOrDocument |
| `76r3` | nœuds illisibles : arêtes perdues | `return nil` | TestAnEdgeDoesNotNameAnExpiredChunkOrDocument |
| `76r3` | filtres du plan dans la relance (L1) | `resolution.Plan.Filters` passé à `gatedSearch` | TestAPinWithNothingExpiredFiltersTheSeedsAsOnMain |
| `76r3` | année hors de 0001 à 9999 (L3) | contrôle désactivé | TestAnExpiryOutsideFourDigitYearsIsRefused |
| `76r3` | porte après chaque préchargement (L2) | ligne retirée | TestASourceLaneDoesNotServeAnExpiredRow |
| `76r3` | score d'un chunk expiré (L2) | `delete(scores, …)` retiré | TestAScoredExpansionDoesNotScoreAnExpiredChunk |
| `76r3` | expiration vide = aucune | branche `<= ''` retirée | TestAnEmptyExpiryWrittenByHandIsNoExpiry |
| `76r3` | `knowledge_delete` d'un enregistrement disparu | `errKept` sans condition | TestAnUnconditionalCleanupOfAGoneRecordIsNoError |
| `76r3` | prédicat d'expiration, sur PostgreSQL (L5) | `AND 1=0` | TestTheExpiryHoldsOnPostgres (et trois tests SQLite) |

Le premier passage de `76r3` a laissé deux survivants, vus seulement par la marque « attendu non tombé » du journal (mon premier comptage des lignes ne la lisait pas) : `source_chunk_id` non lu (l'arête `chunks` du test nommait le même chunk : appel ajouté avec une seule arête) et, dans `76class`, les graines de `SearchGraphRAG` sans la porte (le filtre du plan après la porte relisait l'expiration : remplacé par `allowDocumentID` seul, comme sur `main`). Second passage : `76r3` 14 sur 14, `76class` 3 sur 3, `76` 25 sur 25, `76port` 3 sur 3, `76rt` 1 sur 1, soit 46 sur 46, sous `CORTEXDB_TEST_POSTGRES`. `precommit` (PG, `DECLARED=pkg/core/store_postgres.go`) vert sur `8cfbc13` : 0 bloc, 1 ligne de `main`. `invcheck` 322 sur 322, `diffmain` identique contre `879d864` (21 lignes). Rejouer : `python3 -I docs/pr/round2/mutate_round2.py 76r3 <copie>`.

### Relecture adverse r2, partielle (22 octobre 2026, tête `e4c6cf0`, sur #75 `53121d1`)

Relecture partielle de `8cfbc13` : `adversarial_review_76_r2.md` (B1, L1 à L5 et les remarques fermés ; N1 bloquant, N2 à N4 probables, N5 et N6 remarques). Correctifs : une arête que seuls des documents expirés soutiennent est retirée, et un chemin par une arête retirée n'est plus un chemin, dans `search_paths` (une passe pour tous les chemins de la réponse, N6) et `knowledge_memory_shortest_path` (N1, N2) ; arêtes nettoyées de `shortest_path` tenues par un test (N3) ; recherche et balayage à blanc sur l'expiration vide (N4) ; paramètre `filters` de `gatedSearch` retiré (N5) ; contrôle positif du test de classe sur le chunk ou le texte vivant, que la requête ne contient pas, et `include_bookkeeping` pour `search_paths`. Test nouveau tombé sans le code : `TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut` (sept échecs sur `8cfbc13`). Le contrôle des nœuds dans `search_paths`, écrit d'abord, était redondant (une arête qui touche un nœud retiré part avec lui) : retiré avec sa mutation survivante. La mutation de L1 du jeu `76r3` est retirée : sans paramètre `filters`, le plan ne peut plus atteindre la relance, et un filtre dans la fonction de recherche fait croire le magasin épuisé (mutant équivalent) ; le test de L1 est tombé sur `17983c4` avant le correctif.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `76r4` | arête que seul un document expiré soutient | gardée | TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut, TestAnEdgeDoesNotNameAnExpiredChunkOrDocument |
| `76r4` | arête sans provenance gardée | retirée avec les expirées | TestAnEdgeDoesNotNameAnExpiredChunkOrDocument |
| `76r4` | chemin de `search_paths` par une arête retirée | arête d'origine gardée | TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut |
| `76r4` | `shortest_path` par une arête retirée | contrôle des arêtes retiré | TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut |
| `76r4` | arêtes nettoyées de `shortest_path` rendues (N3) | `_ = edges` | TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut |
| `76r4` | expiration vide = aucune, pour la porte (N4) | `> ''` → `>= ''` | TestAnEmptyExpiryWrittenByHandIsNoExpiry |
| `76r4` | contrôle positif du test de classe | `search_text` rend `chunks[:0]` | TestEveryReadToolIsGatedOrNamed |

Rejeu sous `CORTEXDB_TEST_POSTGRES` : `76r4` 7 sur 7, `76r3` 13 sur 13, `76class` 3 sur 3 (motif de `gatedSearch` mis à jour), `76` 25 sur 25 (idem), `76port` 3 sur 3, `76rt` 1 sur 1, soit 52 sur 52. `precommit` (PG, `DECLARED=pkg/core/store_postgres.go`) vert sur `e4c6cf0` : 0 bloc, 1 ligne de `main` ; `pkg/rpcserver` et `pkg/httpapi` verts sous `-race`. `invcheck` 322 sur 322, `diffmain` identique contre `53121d1` (21 lignes). Rejouer : `python3 -I docs/pr/round2/mutate_round2.py 76r4 <copie>`.

### Relecture adverse r3, partielle (22 octobre 2026, tête `e100a34`, sur #75 `025c490`)

Relecture partielle de `e4c6cf0` : `adversarial_review_76_r3.md` (aucun point bloquant ; N1 à N6, le contrôle positif et la ligne morte fermés). L1 et L2 (le parcours tourne sur le graphe stocké : `shortest_path` peut répondre « no path » quand le plus court chemin passe par une arête retirée, et un chemin retiré prend une place sous `max_paths`) et L3 (`knowledge_memory_neighbors` et `expand_graph` listent encore l'entité atteinte par l'arête retirée) sont nommés dans le corps, avec la raison : filtrer le parcours touche `pkg/graph`, à proposer à part ; « out of the walk » réécrit en « left out of the edges and paths returned » dans le corps, la réponse, le commentaire de `brain.go` et le nom du test (`TestAnEdgeKnownOnlyFromAnExpiredDocumentIsLeftOut`). L4 : deux cas ajoutés à `TestAnEdgeDoesNotNameAnExpiredChunkOrDocument` (document vivant aux chunks expirés, chunk source vivant seul). Remarque : format corrigé dans `expiry_r3_test.go`. Rebasage sur #75 `025c490` : conflit de commentaire d'`EngineDocumentKeys` seul.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `76r5` | arête d'un document vivant aux chunks expirés gardée (L4, M12) | `|| true` | TestAnEdgeDoesNotNameAnExpiredChunkOrDocument |
| `76r5` | arête d'un chunk source vivant gardée (L4, M13) | condition à `true` | TestAnEdgeDoesNotNameAnExpiredChunkOrDocument |

Total sur `e100a34` sous `CORTEXDB_TEST_POSTGRES`, jeux joués l'un après l'autre : `76` (25), `76port` (3), `76class` (3), `76rt` (1), `76r3` (13), `76r4` (7), `76r5` (2), soit 54 sur 54. `precommit` (PG, `DECLARED=pkg/core/store_postgres.go`, base `025c490`) vert : 0 bloc, 1 ligne de `main`. Une première passe a donné de faux échecs : trois suites PostgreSQL en même temps dépassaient `max_connections` (SQLSTATE 53300) ; rejouées seules, vertes.

### Relecture adverse r4, partielle (22 octobre 2026, tête `2384236`, sur #75 `f4c5aa7`)

Relecture partielle de `e100a34` : `adversarial_review_76_r4.md` (aucun point bloquant ; L1 à L4 et la remarque fermés). F1 : le corps nomme maintenant les effets du parcours aussi pour un chemin par un nœud de chunk d'un document expiré. F2 : assertion ajoutée dans #75 (l'enregistrement après une clé forgée), qui boucle sur `EngineDocumentKeys` et tient donc `_expires_at` ici. F3 (les tests des marques par les ports sautent `_expires_at`) laissé : l'autorisation et le retrait par la façade sont tenus à part. Aucun code de cette PR ne change ; rebasage sur `f4c5aa7` sans conflit.

## #77 `gc-06-review-status`

Tests écrits d'abord, tombés sans le code (21 octobre 2026) : `TestAMetadataKeyNamedReviewStatusIsTheCallers`, `TestARowWrittenByMainWithTheReviewKeyIsStillUpdatable`, `TestTheEngineReviewKeysInClientMetadataAreDropped`, `TestMemorySaveOverAnExistingIdCarriesTheReviewOver`, `TestTheReviewGateHoldsOnEveryReadPath`, `TestAStatusFilterOnTheAggregatesListsWhatWaits`, `TestTheSchemasPromiseNothingAboutEntitiesAndRelations`, `TestAReviewKeyOnAnEntityIsTheCallersData`, `TestTheToolDescriptionsSayTheGateIsACurationFilter`, `TestDecideReviewKeepsTheInstantOfARepeatedDecisionOnly`, `TestTheSweepLeavesProposedDocumentsAlone` (premier passage), puis `TestDedupeSeesAProposedOrRejectedHolderUnderTheGate` (tombé : les deux détenteurs cachés, la même page stockée deux fois) et `TestAKnowledgeUpdateCanDecideOnARecordWithoutMetadata` (tenu par la mutation « décision perdue », rejouée à la main : `a decision without a reviewer on a bare record: <nil>`). `TestAMemorySaveReportsARowItCannotRead` couvre un chemin d'erreur (colonne renommée), sans règle nouvelle.

Deux faux résultats du premier passage du jeu `77`, corrigés : « filtre de statut sur les agrégats » déclarée survivante parce que le filtre `-run` du jeu ne lançait pas son test (garde ajoutée au script) ; « clé du client lue comme une décision » et « memory_save repart de rien » ne compilaient pas (carte littérale du test d'entité avec les deux clés, variable `row` inutilisée) : test scindé en deux instructions, `_ = row` dans la mutation. Les trois tombent au second passage.

Classe `promises` (descriptions ajoutées par le diff, `inventory_lotC.md`) : les trois champs de relecture sur cinq outils et la phrase de curation sur chaque description qui nomme la porte, tenus par `TestTheToolDescriptionsSayTheGateIsACurationFilter` ; les descriptions d'entité et de relation, tenues par `TestTheSchemasPromiseNothingAboutEntitiesAndRelations` ; `review_status` du plan de recherche, tenu par `TestAStatusFilterOnTheAggregatesListsWhatWaits` et `TestTheReviewGateHoldsOnEveryReadPath`.

Niveau 3, `gremlins` « NOT COVERED » `review.go:68:11`, `68:29`, `68:50`, `68:72`, `70:11`, `144:12`, `146:12`, et « LIVED » dans les paquets `main` (`cmd/cortexdb-grpc/review_gate.go:17`, `21`, `31` : `TestTheReviewGateIsReadFromTheEnvironment`, `TestTheReviewGateOptionFollowsTheEnvironment` ; `cmd/cortexdb-mcp-stdio/embedder.go:58`, `60` : `TestTheReviewGateIsReadFromTheEnvironment`, `TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet`) : couverture de `gremlins` fausse ou paquet `main` ; rejoués à la main, chacun tombe (`TestDecideReviewKeepsTheInstantOfARepeatedDecisionOnly` et `TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate` pour les trois premiers ; `TestTheReviewGateHidesAProposedDocumentUntilAccepted` et `TestWithoutTheGateEveryReviewedRowIsServed` ; `TestTheReviewStatusIsIndexed`, `TestAMetadataKeyNamedReviewStatusIsTheCallers`).

Mutations (`python3 -I mutate_round2.py 77 <copie>`, 21 octobre 2026, tête `b9777c5`, identique sur `3e167a8` au code près du rebasage) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Clé du client lue comme une décision | `pkg/cortexdb/review.go` : `reviewStatusKey = "review_status"` | TestAMetadataKeyNamedReviewStatusIsTheCallers, TestAReviewKeyOnAnEntityIsTheCallersData, TestARowWrittenByMainWithTheReviewKeyIsStillUpdatable, TestTheReviewStatusIsIndexed |
| Copie du client des clés du moteur gardée | `pkg/cortexdb/review.go` : `r, err := decideReview` | TestTheEngineReviewKeysInClientMetadataAreDropped |
| Décision sans relecteur acceptée | `pkg/cortexdb/review.go` : `if by == "" && false {` | TestADecisionNamesItsReviewer, TestAKnowledgeUpdateCanDecideOnARecordWithoutMetadata |
| Relecteur = le champ, pas la clé appelante | `pkg/cortexdb/review.go` : `if writer := writerOf(ctx); writer != "" && false {` | TestTheReviewerIsTheKeyThatDecided |
| Instant redaté quand le même relecteur répète la même décision | `pkg/cortexdb/review.go` : `case false:` | TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate, TestDecideReviewKeepsTheInstantOfARepeatedDecisionOnly |
| Décision perdue par un knowledge_update sans metadata | `pkg/cortexdb/knowledge_api.go` : `) {` | TestAKnowledgeUpdateCanDecideOnARecordWithoutMetadata |
| memory_save sur un identifiant existant repart de rien (before=nil, point 2) | `pkg/cortexdb/memory_api.go` : `		_ = row` | TestMemorySaveOverAnExistingIdCarriesTheReviewOver |
| Porte de relecture jamais posée | `pkg/cortexdb/review.go` : `case db.reviewGate && false:` | TestTheReviewGateHidesAProposedDocumentUntilAccepted, TestTheReviewGateHoldsOnEveryReadPath, TestTheReviewGateHoldsOnTheVectorPath, TestTheReviewStatusIsIndexed |
| accepted demandé cache les documents sans statut | `pkg/cortexdb/review.go` : `return `(` + expr + ` IS NULL OR ` + expr + ` <> '` + ReviewAccepted +` | TestTheReviewGateHidesAProposedDocumentUntilAccepted, TestWithoutTheGateEveryReviewedRowIsServed |
| search_vector_range sans la porte de relecture | `pkg/cortexdb/vector_range.go` : `(ligne retirée)` | TestTheReviewGateHoldsOnEveryReadPath |
| knowledge_memory_neighbors sans la porte de relecture | `pkg/cortexdb/brain.go` : `(ligne retirée)` | TestTheReviewGateHoldsOnEveryReadPath |
| aggregate_metadata sans la porte de relecture | `pkg/cortexdb/aggregate.go` : `	return db.store.Aggregate(ctx, req)` | TestTheReviewGateHoldsOnEveryReadPath |
| representative_records sans la porte de relecture | `pkg/cortexdb/aggregate.go` : `	return db.store.VectorAggregate(ctx, req)` | TestTheReviewGateHoldsOnEveryReadPath |
| Filtre de statut sur les agrégats qui ne lève pas la porte (R8, th_pending) | `pkg/cortexdb/aggregate.go` : `if true {` | TestAStatusFilterOnTheAggregatesListsWhatWaits |
| Dédup qui ne voit pas un détenteur proposé ou rejeté sous la porte (la porte de relecture dans le prédicat de dédup) | `pkg/cortexdb/expiry.go` : `}` | TestDedupeSeesAProposedOrRejectedHolderUnderTheGate |
| Porte ouverte sur une panne de lecture | `pkg/cortexdb/expiry.go` : `return nil` | TestTheGateFailsClosedWhenTheBrainCannotBeRead |
| Balayage qui liste les documents proposés | `pkg/cortexdb/expiry.go` : `hidden, err := db.hiddenDocuments(ctx, now, db.reviewHiddenPredicate("` | TestTheSweepLeavesProposedDocumentsAlone |
| Rappel des mémoires sans la porte | `pkg/cortexdb/review.go` : `return false` | TestMemorySaveOverAnExistingIdCarriesTheReviewOver, TestTheReviewGateHoldsOnEveryMemoryPath, TestTheReviewGateLeavesAProposedMemoryOutOfRecall |
| Champs de memory_update débranchés | `pkg/cortexdb/knowledge_memory_types.go` : `	return reviewFields{}` | TestADecisionNamesItsReviewer, TestTheReviewGateLeavesAProposedMemoryOutOfRecall, TestTheReviewerIsTheKeyThatDecided |
| Champs d'ingest_document débranchés | `pkg/cortexdb/graphrag_tool_types.go` : `	return reviewFields{}` | TestADecisionNamesItsReviewer, TestTheReviewerIsTheKeyThatDecided |
| Promesse de relecture remise sur les entités (D4) | `pkg/cortexdb/knowledge_memory_tooldefs.go` : `func toolRelationArraySchema` | TestTheSchemasPromiseNothingAboutEntitiesAndRelations |
| Phrase de curation retirée (D3) | `pkg/cortexdb/review.go` : `const ReviewGateNote = "The review gate hides what was not looked at:` | TestTheToolDescriptionsSayTheGateIsACurationFilter |
| Index du statut absent | `pkg/core/store_init.go` : `(ligne retirée)` | TestTheReviewStatusIsIndexed |

### Rebasage sur #75 `025c490` et #76 `e100a34` (22 octobre 2026, tête `f4cd85f`)

Rebasé deux fois (sur #76 `8cfbc13` puis sur les têtes amendées). Conflits : commentaires d'`EngineDocumentKeys` et d'`EngineMemoryKeys`, condition de `knowledge_api.go`, `policy.go` et `ownership.go` (côté #75), appel de `gatedSearch` dans `text_search.go` : le paramètre `filters` de #76 retiré, le statut demandé par le plan passe par `gatedSearchFor`, partie de la porte, les autres filtres du plan après elle. Règle `rewritten` de la liste blanche en fonction (`EngineDocumentKeys` sur les trois outils de document, `EngineMemoryKeys` sur `memory_save` et `memory_update`). Tests nouveaux : `TestEveryGatedReadToolHidesAProposedDocument` (test de classe de #76 sous la porte de relecture, fonction commune `everyGatedToolHides`), `TestUnderTheReviewGateAnEdgeDoesNotNameAProposedChunk` (adapté : sous la porte, nœuds illisibles, l'arête est retirée), `TestInsertGraphDocumentKeepsTheStoredReviewDecision`, `TestAToolWriteTakesTheReviewKeysTheFacadeReplaces` (REST), test mémoire gRPC étendu à `CallTool`. Trois tests de `main` changent d'exemple de clé refusée (`_review_status` → `_table`, `_scope`) : nommés dans le corps.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `77r3` | porte partagée sans la relecture quand rien n'est demandé | prédicat vide si `want == ""` | TestEveryGatedReadToolHidesAProposedDocument |
| `77r3` | porte de relecture ouverte sur nœuds illisibles | `return gone` | TestUnderTheReviewGateAnEdgeDoesNotNameAProposedChunk |
| `77r3` | statut du plan hors de la porte (`gatedSearchFor`) | `&& false` | TestTheReviewGateHidesAProposedDocumentUntilAccepted |
| `77r3` | `InsertGraphDocument` garde la décision envoyée sur les chunks | `if false && …` | TestInsertGraphDocumentKeepsTheStoredReviewDecision |
| `77r3` | `InsertGraphDocument` perd la décision enregistrée | `reviewStatusKey` retiré de `recordOnlyKey` | TestInsertGraphDocumentKeepsTheStoredReviewDecision |
| `77r3` | clés de relecture refusées sur `memory_save` | `rewritten` retiré | TestAToolWriteTakesTheReviewKeysTheFacadeReplaces |
| `77r3` | clés de relecture refusées sur `knowledge_memory_remember` (trouvé par l'inventaire) | `rewritten` retiré | TestAToolWriteTakesTheReviewKeysTheFacadeReplaces, TestTheRewrittenMarksPassOnTheDocumentWritesOnly |
| `77r3` | clés de relecture refusées sur un document relu | retirées d'`EngineDocumentKeys` | TestAToolWriteTakesTheReviewKeysTheFacadeReplaces |

Le jeu `77` avait six motifs périmés après le rebasage (refaits) et un septième après le retrait du paramètre `filters` (`search_vector_range`, refait). L'inventaire (`inventory_77.md`) a trouvé `knowledge_memory_remember` sans la règle `rewritten` : une mémoire relue et renvoyée par cet outil recevait 403 (tests tombés sans le code : `TestTheRewrittenMarksPassOnTheDocumentWritesOnly`, `TestAToolWriteTakesTheReviewKeysTheFacadeReplaces`). Le test de #75 sur `memory_save` boucle sur toutes les marques : sous #77, les clés de relecture sont reprises sur une mémoire, il distingue donc les deux listes. Total sur `f4cd85f` sous `CORTEXDB_TEST_POSTGRES`, jeux joués l'un après l'autre : `77` (23), `77r3` (8), soit 31 sur 31. Rejouer : `python3 -I docs/pr/round2/mutate_round2.py 77r3 <copie>`.

### Relecture adverse r1 (22 octobre 2026, tête `cc8716c`, puis `e437a25`)

Relecture complète de `cc8716c` : `adversarial_review_77_r1.md` (aucun point bloquant, 13 points). Décisions de Lauro pour les points 1 et 6 : filtrer les nœuds de mémoire dans les parcours, refuser la promotion d'une mémoire cachée. Les tests de `review_r4_test.go` sont écrits avant le code. Posés sur `cc8716c` sans le code (`git archive cc8716c`, fichier de test seul ajouté), six tombent : TestAProposedOrRejectedMemoryIsOutOfTheGraphWalks, TestAWalkHidesEveryMemoryWhenTheMemoriesCannotBeRead, TestPromotionRefusesAMemoryTheGateHides, TestAnAcceptedFilterOnTheAggregatesCountsTheDocumentsWithNoStatus, TestAStatusFilterIsClosedOnAnOutageWithoutTheGate, TestEveryKnowledgeSearchLaneHonoursThePlanStatus. Les deux autres sont tenus par mutation : TestWithoutTheGateTheWalksServeEveryMemory (garde du comportement de `main`) et TestABlankReviewerIsNoReviewer (M25 de la relecture).

Le test des voies a trouvé un défaut que la relecture n'avait pas vu en entier (elle signalait M7 sur la voie PPR) : la voie vectorielle de `QueryGraphRAG` (`graphrag.go`), commune aux modes hybride, graphe et PPR, appelait `gatedSearch` sans le statut du plan. Seul le mode lexical obéissait à `plan.filters.review_status`, et sans la porte le filtre `accepted` servait tout.

Autres suites : point 2, le corps dit « sur l'enregistrement seulement » ; point 3, `accepted` prend les documents sans statut sur les agrégats comme dans un plan ; point 7, lignes PostgreSQL déclarées dans la section Codecov ; point 8, `knowledge_memory_remember` nommé dans le corps et le GUIDE ; point 10, branche `errMemoryBucketWithheld` morte (`loadMemoryRow` ne rend jamais cette erreur), retirée ; point 11, un filtre de statut est fermé sur panne même sans la porte ; point 12, commentaire corrigé ; point 13, le corps nomme les trois tests de garde.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `77r4` | voie vectorielle sans le statut du plan | `""` au lieu de `reviewStatusOf(…)` | TestEveryKnowledgeSearchLaneHonoursThePlanStatus |
| `77r4` | parcours qui sert les mémoires cachées | `dropped` vide | TestAProposedOrRejectedMemoryIsOutOfTheGraphWalks |
| `77r4` | mémoire cachée laissée dans les scores | `delete(scores, …)` retiré | TestAProposedOrRejectedMemoryIsOutOfTheGraphWalks (appel `expand_graph` avec `relation_policies`, ajouté après un premier survivant) |
| `77r4` | mémoire illisible servie | `err != nil \|\| …` | TestAWalkHidesEveryMemoryWhenTheMemoriesCannotBeRead |
| `77r4` | mémoires cachées sans la porte | `if false` | TestWithoutTheGateTheWalksServeEveryMemory |
| `77r4` | promotion d'une mémoire cachée | `if false` | TestPromotionRefusesAMemoryTheGateHides |
| `77r4` | promotion refusée sans la porte | `reviewHides(…)` seul | TestPromotionRefusesAMemoryTheGateHides |
| `77r4` | `accepted` des agrégats sans les documents sans statut | `if false` | TestAnAcceptedFilterOnTheAggregatesCountsTheDocumentsWithNoStatus |
| `77r4` | filtre de statut ouvert sur panne sans la porte | `&& want == ""` retiré | TestAStatusFilterIsClosedOnAnOutageWithoutTheGate |
| `77r4` | mémoire rejetée servie (M21) | `ReviewRejected` retiré de `reviewHides` | TestAProposedOrRejectedMemoryIsOutOfTheGraphWalks |
| `77r4` | relecteur fait de blancs (M25) | `TrimSpace` retiré | TestABlankReviewerIsNoReviewer |

Le motif du jeu `77` « Porte ouverte sur une panne de lecture » est refait sur la nouvelle condition. Total : `77` (23), `77r3` (9), `77r4` (11), soit 43 sur 43, le cas `withoutKey` trouvé par `patchcov` compris dans `77r3`.

### Relecture adverse r2, partielle (22 octobre 2026, tête `e437a25`, puis `be0c89d`)

`adversarial_review_77_r2.md` : aucun point bloquant. A (outils qui lisent le graphe stocké tel quel) : décision de Lauro, la classe est nommée dans le corps et le GUIDE, sans code, comme #76 le fait pour une ligne expirée. B : `loadSessionText` laisse de côté, sous la porte, les mémoires cachées ; TestConversationExtractionLeavesOutAMemoryTheGateHides écrit d'abord et tombé sur `e437a25` (« hidden memories in the summary = true »), vert après. Point 4 : TestThePlanStatusHoldsThroughTheEntityLanes (sonde de la relecture, deux documents liés par leurs entités). N1, N2 : un statut vide compté comme `accepted`, une proposition avec un relecteur fait de blancs. N3, N4 : nommés dans le corps. Contrôles positifs ajoutés au test des parcours (`memory_search` sert `m-ok`). Les tests des points 4, N1 et N2 passent sur `e437a25` : la règle y était, la mutation les prouve.

| Jeu | Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- | :- |
| `77r5` | résumé fait des mémoires cachées | `if false` | TestConversationExtractionLeavesOutAMemoryTheGateHides |
| `77r5` | résumé filtré sans la porte | `if true` | TestConversationExtractionLeavesOutAMemoryTheGateHides |
| `77r5` | `accepted` des agrégats sans le statut vide | bras `= ''` retiré | TestAnAcceptedFilterOnTheAggregatesCountsTheDocumentsWithNoStatus |
| `77r5` | proposition avec un relecteur fait de blancs refusée | `TrimSpace` retiré | TestAProposalWithABlankReviewerIsAProposal |
| `77r5` | expansion PPR sans le statut du plan (M7) | `nil` au lieu de `filters` | TestThePlanStatusHoldsThroughTheEntityLanes |

Total : `77` (23), `77r3` (9), `77r4` (11), `77r5` (5), soit 48 sur 48.

## #73 `gc-04b-grpc-confined-tools`

Test écrit avec le code du port (`oneNotFound`) ; la preuve qu'il tombe sans le code est la troisième mutation, qui retire l'appel. Mutations (`python3 -I mutate_round2.py 73 <copie>`, 21 octobre 2026) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/rpcserver/ownership.go` : `confined && status.Code(st) == codes.NotFound && false` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé unique imposé à tout appelant | `pkg/rpcserver/ownership.go` : `(confined \|\| true) && …` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Normalisation court-circuitée dans `CallTool` | `pkg/rpcserver/tools.go` : `return nil, toStatus(err)` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |

Le test tient aussi, par gRPC, les points 1 et 2 de #72 : `missing` dans l'ordre demandé sur deux ordres, le document épinglé d'une autre collection sans résultat et sans erreur, les trois clés réservées refusées, une clé du client acceptée. `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC` (tour 1) attendait NOT_FOUND sur le document épinglé : attente changée en « OK, sans le secret ».

Quatrième passage (21 octobre 2026, tête `e67be29` sur #72 `d007fee`). Tests écrits d'abord : `TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes` tombait sur les cinq écritures typées (entités, `written_by`, clé `_` acceptées) ; le cas `null` de `TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC` faisait paniquer le paquet sur #72 `f1b24eb` (`panic: assignment to entry in nil map`). `precommit` OK contre `d007fee`.

Cinquième passage (relecture adverse de #72, tour 5, L4) : `TestTheTypedRPCsWithholdADocumentTheyCannotPlace` tombait sur la tête rebasée (`GetKnowledge` et `DeleteKnowledge` d'un document sans chunks servis à une clé confinée à la collection par défaut, et le document effacé). Têtes rebasées sur #72 `d805527` : #73 `a04378d` puis message amendé, #74 `7a2ba97` ; `precommit` OK contre `d805527`.

Mutations (`python3 -I mutate_round2.py 73 <copie>`, 14 sur 14 tuées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Document sans chunks placé par les RPC typés (relecture adverse tour 5, L4) | `pkg/rpcserver/ownership.go` : `return scope.PermitsKnowledge(record.Collection)` | TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Cible supersedes manquante vérifiée après l'écriture (oracle par la nouvelle ligne) | `pkg/cortexdb/memory_api.go` : `if err := db.checkSupersedeTargets(ctx, req.MemoryID, nil); err != nil` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/rpcserver/ownership.go` : `if _, confined := confinedCaller(ctx); confined && status.Code(st) == ` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé unique imposé à tout appelant | `pkg/rpcserver/ownership.go` : `if _, confined := confinedCaller(ctx); (confined || true) && status.Co` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Tout refus d'une clé confinée réécrit en NOT_FOUND (L6, relecture adverse) | `pkg/rpcserver/ownership.go` : `if _, confined := confinedCaller(ctx); confined {` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| RPC typés hors de la liste blanche (L3, relecture adverse) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Entités typées non transmises à la liste blanche | `pkg/rpcserver/ownership.go` : `if len(entities) > 0 && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Relations typées non transmises à la liste blanche | `pkg/rpcserver/ownership.go` : `if len(relations) > 0 && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées typées d'un document non transmises | `pkg/rpcserver/ownership.go` : `_ = meta` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées de SaveMemory non transmises | `pkg/rpcserver/ownership.go` : `		if m := r.GetMetadata().AsMap(); len(m) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Champs de portée de SaveMemory non transmis | `pkg/rpcserver/ownership.go` : `if v != "" && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées d'UpdateMemory non transmises | `pkg/rpcserver/ownership.go` : `		if m := r.GetMetadata().AsMap(); len(m) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Collection d'UpdateKnowledge non transmise | `pkg/rpcserver/ownership.go` : `tool = "knowledge_update"` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Normalisation court-circuitée dans CallTool | `pkg/rpcserver/tools.go` : `return nil, toStatus(err)` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |

Sixième passage (relecture adverse de #73, tour 1, `adversarial_review_73_r1.md`, 21 octobre 2026). Tests écrits d'abord, sur `ff24c70` sans le code : `TestATypedWriteTakesTheSignatureTheFacadeReplaces` tombait (`SaveKnowledge` avec `updated_by` refusé, L3), `TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne` tombait (`get knowledge: vectorstore: get_document: embedding not found`, L1), `TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks` tombait (`UpdateMemory` refusé « switch off with =true », `SaveMemory` refusé par la liste blanche, L4). Les cas ajoutés à des tests existants (relations d'`UpdateKnowledge`, L2 ; namespace et scope de `SaveMemory`, N1 ; refus propre de l'outil comparé à celui de l'opérateur, L5) tiennent du code déjà là : la preuve est la mutation qui survivait à la relecture et tombe maintenant. Décisions de Lauro : L3, la signature passe sur les quatre écritures signées par la façade ; L4, corrigé dans #73 seule.

Niveau 3 de `ff24c70` : `gremlins` « LIVED » `ownership.go:178:43` et `183:43` (`len(m) > 0` → `>= 0`) : équivalents, une carte `metadata` vide n'a aucune clé à contrôler et `metadata` est permis à `memory_save` et `memory_update`. Sur `1c149f7` (`prepr` vert) : les mêmes à `202:43` et `210:43`, et « NOT COVERED » `294:11` (`case err == nil`), déjà rejoué à la main (niveau 3, plus haut).

Mutations (`python3 -I mutate_round2.py 73 <copie>`, 31 sur 31 tuées ; les trois motifs de `oneNotFound` portent maintenant leur ligne suivante, `withheldIfMissing` ayant la même condition) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Document sans chunks placé par les RPC typés (relecture adverse tour 5, L4) | `pkg/rpcserver/ownership.go` : `return scope.PermitsKnowledge(record.Collection)` | TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Cible supersedes manquante vérifiée après l'écriture (oracle par la nouvelle ligne) | `pkg/cortexdb/memory_api.go` : `if err := db.checkSupersedeTargets(ctx, req.MemoryID, nil); err != nil` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé unique imposé à tout appelant | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Tout refus d'une clé confinée réécrit en NOT_FOUND (L6, relecture adverse) | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| RPC typés hors de la liste blanche (L3, relecture adverse) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Entités typées non transmises à la liste blanche | `pkg/rpcserver/ownership.go` : `if len(entities) > 0 && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Relations typées non transmises à la liste blanche | `pkg/rpcserver/ownership.go` : `if len(relations) > 0 && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées typées d'un document non transmises | `pkg/rpcserver/ownership.go` : `_ = meta` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées de SaveMemory non transmises | `pkg/rpcserver/ownership.go` : `		if m := r.GetMetadata().AsMap(); len(m) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Champs de portée de SaveMemory non transmis | `pkg/rpcserver/ownership.go` : `if v != "" && false {` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks, TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées d'UpdateMemory non transmises | `pkg/rpcserver/ownership.go` : `		if m := r.GetMetadata().AsMap(); len(m) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Collection d'UpdateKnowledge non transmise | `pkg/rpcserver/ownership.go` : `tool = "knowledge_update"` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne, TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Normalisation court-circuitée dans CallTool | `pkg/rpcserver/tools.go` : `return nil, toStatus(err)` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Relations d'UpdateKnowledge non transmises (L2) | `pkg/rpcserver/ownership.go` : `		graph(r.GetEntities(), nil)` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Namespace de SaveMemory non transmis (N1) | `pkg/rpcserver/ownership.go` : `authz.FieldNamespace: ""` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Scope de SaveMemory non transmis (N1) | `pkg/rpcserver/ownership.go` : `authz.FieldScope: ""` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Autre refus d'une clé confinée réécrit (L5) | `pkg/rpcserver/ownership.go` : `// withheldIfMissing` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| written_by signé par la façade refusé (L3) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| updated_by signé par la façade refusé (L3) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Autre graphie retirée aussi, que la façade ne remplace pas (L3) | `pkg/rpcserver/ownership.go` : `		}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur SaveKnowledge seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.UpdateKnowledgeRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur UpdateKnowledge seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.InsertGraphDocumentRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée aussi sur InsertGraphDocument, non signé (L3) | `pkg/rpcserver/ownership.go` : `		}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur SaveMemory seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.UpdateMemoryRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur UpdateMemory seule, aller-retour (L3) | `pkg/rpcserver/ownership.go` : `			args["metadata"] = m` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| SaveMemory d'une clé de collection tenue à la liste blanche (L4) | `pkg/rpcserver/ownership.go` : `tool = "memory_save"` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks |
| UpdateMemory d'une clé de collection tenue à la liste blanche (L4) | `pkg/rpcserver/ownership.go` : `tool = "memory_update"` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks |
| Toute clé dispensée de la liste blanche sur SaveMemory (L4) | `pkg/rpcserver/ownership.go` : `		tool = "memory_save"` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Identifiant absent d'UpdateKnowledge rendu tel quel (L1) | `pkg/rpcserver/ownership.go` : `		return withheld(kind, id)` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne |
| Libellé retenu imposé à l'opérateur sur UpdateKnowledge (L1) | `pkg/rpcserver/ownership.go` : `		return withheld(kind, id)` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne |


Septième passage (relecture adverse de #73, tour 2, `adversarial_review_73_r2.md`, un bloquant). Tests écrits d'abord, sur `1c149f7` sans le code : `TestATypedIngestCannotTakeADocumentOfAnotherCollection` tombait (`InsertGraphDocument` de la clé `notes` sur `sec` : OK, document passé dans `notes` et réécrit, B1) ; le cas `InsertGraphDocument` ajouté à `TestTheTypedRPCsWithholdADocumentTheyCannotPlace` tombait (le document sans chunks atteint l'ingestion). Les cas `SaveKnowledge` et `UpdateKnowledge` du même test (L1) et le seau d'un autre utilisateur dans `TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC` (L3) tiennent du code déjà là : preuve par les mutations M1 et M3 du relecteur, qui survivaient et tombent maintenant. N3 : `InsertGraphDocument` du test des écritures typées tourne avec l'embedder factice et doit réussir. N5 : mutant équivalent (champs de portée vides transmis), un champ confiné vide est refusé par `AuthorizeRows`.

Mutations (`python3 -I mutate_round2.py 73 <copie>`, 34 sur 34 tuées sur `4f71850`). Ensuite, sur décision de Lauro, B1 est sorti de #73 vers la PR à part `gc-12-grpc-ingest-guard` (section en fin de fichier), avec `TestATypedIngestCannotTakeADocumentOfAnotherCollection`, le cas `InsertGraphDocument` de `TestTheTypedRPCsWithholdADocumentTheyCannotPlace` et la mutation « InsertGraphDocument typé sans garde de ligne (B1) » ; le jeu `73` compte 33 mutations :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Document sans chunks placé par les RPC typés (relecture adverse tour 5, L4) | `pkg/rpcserver/ownership.go` : `return scope.PermitsKnowledge(record.Collection)` | TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Cible supersedes manquante vérifiée après l'écriture (oracle par la nouvelle ligne) | `pkg/cortexdb/memory_api.go` : `if err := db.checkSupersedeTargets(ctx, req.MemoryID, nil); err != nil` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé unique imposé à tout appelant | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Tout refus d'une clé confinée réécrit en NOT_FOUND (L6, relecture adverse) | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| RPC typés hors de la liste blanche (L3, relecture adverse) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Entités typées non transmises à la liste blanche | `pkg/rpcserver/ownership.go` : `if len(entities) > 0 && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Relations typées non transmises à la liste blanche | `pkg/rpcserver/ownership.go` : `if len(relations) > 0 && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées typées d'un document non transmises | `pkg/rpcserver/ownership.go` : `_ = meta` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées de SaveMemory non transmises | `pkg/rpcserver/ownership.go` : `		if m := r.GetMetadata().AsMap(); len(m) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Champs de portée de SaveMemory non transmis | `pkg/rpcserver/ownership.go` : `if v != "" && false {` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks, TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées d'UpdateMemory non transmises | `pkg/rpcserver/ownership.go` : `		if m := r.GetMetadata().AsMap(); len(m) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Collection d'UpdateKnowledge non transmise | `pkg/rpcserver/ownership.go` : `tool = "knowledge_update"` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne, TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes, TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Normalisation court-circuitée dans CallTool | `pkg/rpcserver/tools.go` : `return nil, toStatus(err)` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Relations d'UpdateKnowledge non transmises (L2) | `pkg/rpcserver/ownership.go` : `		graph(r.GetEntities(), nil)` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Namespace de SaveMemory non transmis (N1) | `pkg/rpcserver/ownership.go` : `authz.FieldNamespace: ""` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Scope de SaveMemory non transmis (N1) | `pkg/rpcserver/ownership.go` : `authz.FieldScope: ""` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Autre refus d'une clé confinée réécrit (L5) | `pkg/rpcserver/ownership.go` : `// withheldIfMissing` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| written_by signé par la façade refusé (L3) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| updated_by signé par la façade refusé (L3) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Autre graphie retirée aussi, que la façade ne remplace pas (L3) | `pkg/rpcserver/ownership.go` : `		}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur SaveKnowledge seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.UpdateKnowledgeRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur UpdateKnowledge seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.InsertGraphDocumentRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée aussi sur InsertGraphDocument, non signé (L3) | `pkg/rpcserver/ownership.go` : `		}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur SaveMemory seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.UpdateMemoryRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur UpdateMemory seule, aller-retour (L3) | `pkg/rpcserver/ownership.go` : `			args["metadata"] = m` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| SaveMemory d'une clé de collection tenue à la liste blanche (L4) | `pkg/rpcserver/ownership.go` : `tool = "memory_save"` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks |
| UpdateMemory d'une clé de collection tenue à la liste blanche (L4) | `pkg/rpcserver/ownership.go` : `tool = "memory_update"` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks |
| Toute clé dispensée de la liste blanche sur SaveMemory (L4) | `pkg/rpcserver/ownership.go` : `		tool = "memory_save"` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Identifiant absent d'UpdateKnowledge rendu tel quel (L1) | `pkg/rpcserver/ownership.go` : `		return withheld(kind, id)` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne |
| Libellé retenu imposé à l'opérateur sur UpdateKnowledge (L1) | `pkg/rpcserver/ownership.go` : `		return withheld(kind, id)` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne |
| InsertGraphDocument typé sans garde de ligne (B1) | `pkg/rpcserver/graphrag.go` : `(ligne retirée)` | TestATypedIngestCannotTakeADocumentOfAnotherCollection, TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Écriture typée d'un document sans chunks laissée passer (L1) | `pkg/rpcserver/ownership.go` : `	if err == nil {` | TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Libellé « memory bucket not found » rendu à une clé confinée (L3) | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |

Huitième passage (relecture adverse de #73, tour 3, `adversarial_review_73_r3.md`, aucun bloquant propre à la PR). L3 et L4 : `TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne` compare maintenant un refus qui n'est pas NOT_FOUND à celui de l'opérateur ; `TestTheTypedMemoryRowChecksHoldEveryMemoryField` (clés `lab` et `mine` contre des lignes stockées). Tous deux tiennent du code déjà là : preuve par les mutations V, X et Y du relecteur, qui survivaient et tombent maintenant. L1 (identifiant blanc) : corrigé par #82, nommé dans le corps. L2 : artefact de la session (matériel modifié pendant la relecture). Corps : N2, N4, N5 ; message de commit : N3.

Mutations (`python3 -I mutate_round2.py 73 <copie>`, 36 sur 36 tuées, sans B1) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Document sans chunks placé par les RPC typés (relecture adverse tour 5, L4) | `pkg/rpcserver/ownership.go` : `return scope.PermitsKnowledge(record.Collection)` | TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Cible supersedes manquante vérifiée après l'écriture (oracle par la nouvelle ligne) | `pkg/cortexdb/memory_api.go` : `if err := db.checkSupersedeTargets(ctx, req.MemoryID, nil); err != nil` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Libellé unique imposé à tout appelant | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Tout refus d'une clé confinée réécrit en NOT_FOUND (L6, relecture adverse) | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| RPC typés hors de la liste blanche (L3, relecture adverse) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Entités typées non transmises à la liste blanche | `pkg/rpcserver/ownership.go` : `if len(entities) > 0 && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Relations typées non transmises à la liste blanche | `pkg/rpcserver/ownership.go` : `if len(relations) > 0 && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées typées d'un document non transmises | `pkg/rpcserver/ownership.go` : `_ = meta` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées de SaveMemory non transmises | `pkg/rpcserver/ownership.go` : `		if m := r.GetMetadata().AsMap(); len(m) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Champs de portée de SaveMemory non transmis | `pkg/rpcserver/ownership.go` : `if v != "" && false {` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks, TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées d'UpdateMemory non transmises | `pkg/rpcserver/ownership.go` : `		if m := r.GetMetadata().AsMap(); len(m) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Collection d'UpdateKnowledge non transmise | `pkg/rpcserver/ownership.go` : `tool = "knowledge_update"` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne, TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes, TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Normalisation court-circuitée dans CallTool | `pkg/rpcserver/tools.go` : `return nil, toStatus(err)` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Relations d'UpdateKnowledge non transmises (L2) | `pkg/rpcserver/ownership.go` : `		graph(r.GetEntities(), nil)` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Namespace de SaveMemory non transmis (N1) | `pkg/rpcserver/ownership.go` : `authz.FieldNamespace: ""` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Scope de SaveMemory non transmis (N1) | `pkg/rpcserver/ownership.go` : `authz.FieldScope: ""` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Autre refus d'une clé confinée réécrit (L5) | `pkg/rpcserver/ownership.go` : `// withheldIfMissing` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| written_by signé par la façade refusé (L3) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| updated_by signé par la façade refusé (L3) | `pkg/rpcserver/ownership.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Autre graphie retirée aussi, que la façade ne remplace pas (L3) | `pkg/rpcserver/ownership.go` : `		}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur SaveKnowledge seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.UpdateKnowledgeRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur UpdateKnowledge seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.InsertGraphDocumentRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée aussi sur InsertGraphDocument, non signé (L3) | `pkg/rpcserver/ownership.go` : `		}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur SaveMemory seule (L3) | `pkg/rpcserver/ownership.go` : `	case *rpcv1.UpdateMemoryRequest:` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Signature retirée sur UpdateMemory seule, aller-retour (L3) | `pkg/rpcserver/ownership.go` : `			args["metadata"] = m` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| SaveMemory d'une clé de collection tenue à la liste blanche (L4) | `pkg/rpcserver/ownership.go` : `tool = "memory_save"` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks |
| UpdateMemory d'une clé de collection tenue à la liste blanche (L4) | `pkg/rpcserver/ownership.go` : `tool = "memory_update"` | TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks |
| Toute clé dispensée de la liste blanche sur SaveMemory (L4) | `pkg/rpcserver/ownership.go` : `		tool = "memory_save"` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Identifiant absent d'UpdateKnowledge rendu tel quel (L1) | `pkg/rpcserver/ownership.go` : `		return withheld(kind, id)` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne |
| Libellé retenu imposé à l'opérateur sur UpdateKnowledge (L1) | `pkg/rpcserver/ownership.go` : `		return withheld(kind, id)` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne |
| Écriture typée d'un document sans chunks laissée passer (L1) | `pkg/rpcserver/ownership.go` : `	if err == nil {` | TestTheTypedRPCsWithholdADocumentTheyCannotPlace |
| Libellé « memory bucket not found » rendu à une clé confinée (L3) | `pkg/rpcserver/ownership.go` : `		return status.Error(codes.NotFound, authz.ErrWithheld.Error())` | TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC |
| Tout refus de la façade réécrit pour une clé confinée sur UpdateKnowledge (L3, V) | `pkg/rpcserver/ownership.go` : `		return withheld(kind, id)` | TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne |
| Namespace de la ligne non transmis à la règle des mémoires (L4, X) | `pkg/rpcserver/ownership.go` : `return scope.PermitsMemory(record.UserID, record.Scope, "")` | TestTheTypedMemoryRowChecksHoldEveryMemoryField |
| Scope de la ligne non transmis à la règle des mémoires (L4, Y) | `pkg/rpcserver/ownership.go` : `return scope.PermitsMemory(record.UserID, "", record.Namespace)` | TestTheTypedMemoryRowChecksHoldEveryMemoryField |

## #74 `gc-04c-rest-confined-tools`

Même structure par REST. Mutations (`python3 -I mutate_round2.py 74 <copie>`) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/httpapi/policy.go` : `… && status == http.StatusNotFound && false` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé unique imposé à toute clé | `pkg/httpapi/policy.go` : `ok && (key.Scope.IsZero() \|\| true) && …` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Normalisation court-circuitée dans la route | `pkg/httpapi/tools.go` : `code, message := statusFor(err), err.Error()` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |

Première écriture de la deuxième mutation : `ok && status == …` ne compilait pas (`key` déclarée, non utilisée) ; réécrite.

Quatrième passage (21 octobre 2026, tête `953a68b` sur #72 `d007fee`). `TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes` tombait sur les deux routes typées. Le corps `null` était déjà refusé en 400 par la route (pas d'appel à `ConfineToolArgs`) : le cas tient ce comportement. `precommit` OK contre `d007fee`. Mutations (`python3 -I mutate_round2.py 74 <copie>`, 10 sur 10 tuées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Cible supersedes manquante vérifiée après l'écriture (oracle par la nouvelle ligne) | `pkg/cortexdb/memory_api.go` : `if err := db.checkSupersedeTargets(ctx, req.MemoryID, nil); err != nil` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && !key.Scope.IsZero() && status == htt` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé unique imposé à toute clé | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && (key.Scope.IsZero() || true) && stat` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Tout refus d'une clé confinée réécrit en 404 (L6, relecture adverse) | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && !key.Scope.IsZero() {` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| JSON cassé pris pour un objet (403 au lieu de 400 ; L6, relecture adverse) | `pkg/httpapi/policy.go` : `!(isJSONObject(body))` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Routes typées hors de la liste blanche (L3, relecture adverse) | `pkg/httpapi/policy.go` : `if tool, typed := typedWriteTools[rr.String()]; typed && false {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Entités et relations typées non transmises | `pkg/httpapi/policy.go` : `if list, _ := v.([]any); len(list) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées typées non transmises | `pkg/httpapi/policy.go` : `authz.FieldNamespace, "metadata_off":` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Champs de portée typés non transmis | `pkg/httpapi/policy.go` : `case "metadata":` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Normalisation court-circuitée dans la route | `pkg/httpapi/tools.go` : `code, message := statusFor(err), err.Error()` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |

Cinquième passage (22 octobre 2026, avant la relecture adverse de #74 ; tête `7a2ba97` sur #72 `d805527`). Les décisions L3 et L4 de #73 appliquées au port REST, après sondes sur `main` et sur la tête :
- L3 : `written_by` et `updated_by` sur `POST /v1/knowledge` et `POST /v1/memory`, 200 sur `main` (la façade les remplace), 403 sur la tête. Ils ne sont plus transmis à la liste blanche ; une autre orthographe (`Written_By`), que la façade ne remplace pas, et la route des outils gardent le refus.
- L4 : clé qui n'atteint aucune ligne du magasin de la route (collection sur `/v1/memory`, `user_id`, `scope` ou `namespace` sur `/v1/knowledge`) : 400 `json: unknown field` sur `main`, 403 de la liste blanche sur la tête. Elle est laissée au contrôle des lignes et au décodeur.

Sans le code :

```
--- FAIL: TestATypedWriteTakesTheSignatureTheFacadeReplaces
    /v1/knowledge with the facade's signature keys: 403 {"error":"denied: key \"notes\" is confined and knowledge_save takes no metadata.written_by …
    /v1/memory with the facade's signature keys: 403 {"error":"denied: key \"hermes\" is confined and memory_save takes no metadata.updated_by …
--- FAIL: TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers
    /v1/memory {"memory_id":"m1","collection":"notes","content":"x"}: 403 … memory_save takes no collection argument …
    /v1/knowledge {"knowledge_id":"l1","namespace":"lab","content":"x"}: 403 … knowledge_save takes no namespace argument …
    /v1/knowledge {"knowledge_id":"h1","user_id":"hermes","content":"x"}: 403 … knowledge_save takes no user_id argument …
```

Les mêmes sondes sur `main` donnent les réponses que le second test attend. Trois motifs anciens mis à jour (le code de `typedToolArgs` et la condition des routes typées ont changé). `prepr` sur `41f68aa` : inventaire 230/230, `patchcov` 0, `gremlins` Killed 11 / Lived 1 / Not covered 0, suite `-race` verte, exit 0. Le survivant, `policy.go:239` (`len(list) > 0` → `>= 0`), est réel : une liste vide d'entités ou de relations passe sur la route typée (200, comme sur `main`) et est refusée par la route des outils (403) ; aucun test ne l'envoyait. Cas ajouté à `TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes`, mutation à la main ajoutée.

Mutations (`python3 -I mutate_round2.py 74 <copie>`, 21 sur 21 tuées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| written_by gardé pour la liste blanche sur une route typée | `pkg/httpapi/policy.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| updated_by gardé pour la liste blanche sur une route typée | `pkg/httpapi/policy.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Toute orthographe de written_by retirée | `pkg/httpapi/policy.go` : `				}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Clé d'un autre magasin tenue à la liste blanche | `pkg/httpapi/policy.go` : `typed {` | TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers |
| Clé confinée à user_id tenue à la liste blanche des connaissances | `pkg/httpapi/policy.go` : `return` | TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers |
| Clé confinée à scope tenue à la liste blanche des connaissances | `pkg/httpapi/policy.go` : `s.Permits(authz.FieldNamespace, "")` | TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers |
| Clé confinée à namespace tenue à la liste blanche des connaissances | `pkg/httpapi/policy.go` : `(ligne retirée)` | TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers |
| Clé de collection tenue à la liste blanche des mémoires | `pkg/httpapi/policy.go` : `return true` | TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers |
| Liste blanche sautée sur la route des connaissances | `pkg/httpapi/policy.go` : `return false && s.Permits(authz.FieldUserID, "") &&` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Liste blanche sautée sur la route des mémoires | `pkg/httpapi/policy.go` : `return false` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Cible supersedes manquante vérifiée après l'écriture (oracle par la nouvelle ligne) | `pkg/cortexdb/memory_api.go` : `if err := db.checkSupersedeTargets(ctx, req.MemoryID, nil); err != nil` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && !key.Scope.IsZero() && status == htt` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé unique imposé à toute clé | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && (key.Scope.IsZero() || true) && stat` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Tout refus d'une clé confinée réécrit en 404 (L6, relecture adverse) | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && !key.Scope.IsZero() {` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| JSON cassé pris pour un objet (403 au lieu de 400 ; L6, relecture adverse) | `pkg/httpapi/policy.go` : `!(isJSONObject(body))` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Routes typées hors de la liste blanche (L3, relecture adverse) | `pkg/httpapi/policy.go` : `if tw, typed := typedWriteTools[rr.String()]; typed && tw.reaches(key.` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Listes vides d'entités et de relations transmises (survivant de gremlins, policy.go:239) | `pkg/httpapi/policy.go` : `if list, _ := v.([]any); len(list) >= 0 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Entités et relations typées non transmises | `pkg/httpapi/policy.go` : `if list, _ := v.([]any); len(list) > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées typées non transmises | `pkg/httpapi/policy.go` : `case "metadata_off":` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Champs de portée typés non transmis | `pkg/httpapi/policy.go` : `case "scope_off":` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Normalisation court-circuitée dans la route | `pkg/httpapi/tools.go` : `code, message := statusFor(err), err.Error()` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |

Sixième passage (22 octobre 2026, relecture adverse de #74, tour 1, `adversarial_review_74_r1.md`, sur `bc2175d`) : 1 bloquant, 2 remarques probables, 4 nits.
- B1 corrigé : la route typée lisait les champs par `strings.ToLower`, le décodeur du gestionnaire les plie par Unicode (`entitieſ`, s long U+017F, est son `Entities`) ; une clé confinée écrivait des entités et des relations du graphe partagé. Le contrôle décode maintenant le corps dans le type de requête de la façade avec le même décodeur (`decodeTyped`) ; un corps que le décodeur refuse va au gestionnaire (400, comme sur `main`). Le prédicat `reaches` du cinquième passage devient inutile (le champ de portée de l'autre magasin n'est pas un champ de la route, donc 400) : retiré, ses sept mutations aussi (cinq devenaient équivalentes).
- L1 corrigé : la branche « unknown tool » de `tools.go` passe par `oneNotFound`.
- L2 corrigé : corps (section Fix, liste des tests, compte des mutations du premier tour).
- N1 (`memory_update` d'une clé de collection : « … with =true ») est un libellé de #72 ; noté pour #72. N2, N3 (assertions faibles de tests du premier tour, libellé ancien), N4 (mutants équivalents `ContentLength`, `!key.Scope.IsZero()`) : laissés.

Sans le code (sur `bc2175d`) :

```
--- FAIL: TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes
    /v1/knowledge {… "entitieſ":[…]}: 200 …
    /v1/knowledge {… "relationſ":[…]}: 500 …
    /v1/memory {… "entitieſ":[…]}: 200 …
    /v1/memory {… "Relationſ":[…]}: 500 …
--- FAIL: TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther
    a missing supersedes id that reads "unknown tool": 404 {"error":"supersede \"unknown tool q\": embedding not found"}, want 404 "not found"
```

Mutations (`python3 -I mutate_round2.py 74 <copie>`, 23 sur 23 tuées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Cible supersedes manquante vérifiée après l'écriture (oracle par la nouvelle ligne) | `pkg/cortexdb/memory_api.go` : `if err := db.checkSupersedeTargets(ctx, req.MemoryID, nil); err != nil` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && !key.Scope.IsZero() && status == htt` | TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther, TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé unique imposé à toute clé | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && (key.Scope.IsZero() || true) && stat` | TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther, TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Tout refus d'une clé confinée réécrit en 404 (L6, relecture adverse) | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && !key.Scope.IsZero() {` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| JSON cassé pris pour un objet (403 au lieu de 400 ; L6, relecture adverse) | `pkg/httpapi/policy.go` : `!(isJSONObject(body))` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Routes typées hors de la liste blanche (L3, relecture adverse) | `pkg/httpapi/policy.go` : `if tw, typed := typedWriteTools[rr.String()]; typed && false {` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| written_by gardé pour la liste blanche sur une route typée | `pkg/httpapi/policy.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| updated_by gardé pour la liste blanche sur une route typée | `pkg/httpapi/policy.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Toute orthographe de written_by retirée | `pkg/httpapi/policy.go` : `	}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Corps lu par nom et non comme le décodeur (relecture adverse, B1) | `pkg/httpapi/policy.go` : `	args := map[string]any{authz.FieldCollection: req.Collection}` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Entités et relations d'une mémoire non transmises (B1) | `pkg/httpapi/policy.go` : `	req.Entities, req.Relations = nil, nil` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Corps refusé par le décodeur tenu à la liste blanche (clé de l'autre magasin comprise) | `pkg/httpapi/policy.go` : `if args, decoded := tw.args(body); decoded || true {` | TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers |
| Entités typées non transmises | `pkg/httpapi/policy.go` : `if entities > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Relations typées non transmises | `pkg/httpapi/policy.go` : `if relations > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Listes vides d'entités transmises (survivant de gremlins) | `pkg/httpapi/policy.go` : `if entities >= 0 {` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Listes vides de relations transmises | `pkg/httpapi/policy.go` : `if relations >= 0 {` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées typées non transmises | `pkg/httpapi/policy.go` : `if len(meta) > 99 {` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées d'un document non transmises | `pkg/httpapi/policy.go` : `	meta := make(map[string]any, len(req.Metadata))` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées d'une mémoire non transmises | `pkg/httpapi/policy.go` : `	meta := make(` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Collection d'un document non transmise | `pkg/httpapi/policy.go` : `args := map[string]any{}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Champs de portée d'une mémoire non transmis | `pkg/httpapi/policy.go` : `			args[field] = v` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Normalisation court-circuitée dans la route | `pkg/httpapi/tools.go` : `message := err.Error()` | TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther, TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Réponse « unknown tool » rendue telle quelle (relecture adverse, L1) | `pkg/httpapi/tools.go` : `				return` | TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther |

Septième passage (22 octobre 2026, relecture adverse de #74, tour 2 partiel, `adversarial_review_74_r2.md`, sur `ac16f92`) : aucun bloquant ; B1, L1, L2 fermés. Suite donnée :
- M4 (survivant du relecteur) : `namespace` ou `scope` seul retiré de `memoryToolArgs` ne faisait rien tomber. Écritures légitimes des clés `lab` (namespace) et `mine` (scope) ajoutées, une mutation par champ.
- M6 (survivant du relecteur) : `{"collection":"notes","Collection":""}`, 200 sur `main` et sur `bc2175d` (le document va dans `graphrag_chunks`), 403 sur la tête sans test. Cas ajouté, nommé dans le corps comme trou de `main` fermé. Sur `bc2175d` le cas tombe : `200 {"knowledge":{"id":"n1",…,"collection":"graphrag_chunks"…`.
- Nits : la réponse et le message nomment les deux réponses (400 du gestionnaire, 403 du contrôle des lignes) ; la liste des mutations du corps suit le tableau.
- M5 (équivalent : `user_id` codé en dur, seul le doublon vide diffère, refusé par la façade) : laissé.

Mutations (`python3 -I mutate_round2.py 74 <copie>`, 26 sur 26 tuées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Cible supersedes manquante vérifiée après l'écriture (oracle par la nouvelle ligne) | `pkg/cortexdb/memory_api.go` : `if err := db.checkSupersedeTargets(ctx, req.MemoryID, nil); err != nil` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé de l'outil rendu tel quel à une clé confinée | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && !key.Scope.IsZero() && status == htt` | TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther, TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Libellé unique imposé à toute clé | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && (key.Scope.IsZero() || true) && stat` | TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther, TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Tout refus d'une clé confinée réécrit en 404 (L6, relecture adverse) | `pkg/httpapi/policy.go` : `if key, ok := keyFrom(ctx); ok && !key.Scope.IsZero() {` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| JSON cassé pris pour un objet (403 au lieu de 400 ; L6, relecture adverse) | `pkg/httpapi/policy.go` : `!(isJSONObject(body))` | TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Routes typées hors de la liste blanche (L3, relecture adverse) | `pkg/httpapi/policy.go` : `if tw, typed := typedWriteTools[rr.String()]; typed && false {` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| written_by gardé pour la liste blanche sur une route typée | `pkg/httpapi/policy.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| updated_by gardé pour la liste blanche sur une route typée | `pkg/httpapi/policy.go` : `(ligne retirée)` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Toute orthographe de written_by retirée | `pkg/httpapi/policy.go` : `	}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces |
| Corps lu par nom et non comme le décodeur (relecture adverse, B1) | `pkg/httpapi/policy.go` : `	args := map[string]any{authz.FieldCollection: req.Collection}` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Entités et relations d'une mémoire non transmises (B1) | `pkg/httpapi/policy.go` : `	req.Entities, req.Relations = nil, nil` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Corps refusé par le décodeur tenu à la liste blanche (clé de l'autre magasin comprise) | `pkg/httpapi/policy.go` : `if args, decoded := tw.args(body); decoded || true {` | TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers |
| Entités typées non transmises | `pkg/httpapi/policy.go` : `if entities > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Relations typées non transmises | `pkg/httpapi/policy.go` : `if relations > 99 {` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Listes vides d'entités transmises (survivant de gremlins) | `pkg/httpapi/policy.go` : `if entities >= 0 {` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Listes vides de relations transmises | `pkg/httpapi/policy.go` : `if relations >= 0 {` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées typées non transmises | `pkg/httpapi/policy.go` : `if len(meta) > 99 {` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées d'un document non transmises | `pkg/httpapi/policy.go` : `	meta := make(map[string]any, len(req.Metadata))` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Métadonnées d'une mémoire non transmises | `pkg/httpapi/policy.go` : `	meta := make(` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Collection d'un document non transmise | `pkg/httpapi/policy.go` : `args := map[string]any{}` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| namespace d'une mémoire non transmis (relecture adverse, tour 2, M4) | `pkg/httpapi/policy.go` : `authz.FieldScope: req.Scope}` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| scope d'une mémoire non transmis (relecture adverse, tour 2, M4) | `pkg/httpapi/policy.go` : `authz.FieldNamespace: req.Namespace}` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Collection d'un document lue par nom et non décodée (relecture adverse, tour 2, M6) | `pkg/httpapi/policy.go` : `args := map[string]any{authz.FieldCollection: "notes"}` | TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Champs de portée d'une mémoire non transmis | `pkg/httpapi/policy.go` : `			args[field] = v` | TestATypedWriteTakesTheSignatureTheFacadeReplaces, TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes |
| Normalisation court-circuitée dans la route | `pkg/httpapi/tools.go` : `message := err.Error()` | TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther, TestTheRoundTwoRulesOfTheAllowlistHoldOverREST |
| Réponse « unknown tool » rendue telle quelle (relecture adverse, L1) | `pkg/httpapi/tools.go` : `				return` | TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther |

Niveau 3 sur `7b27a64` (22 octobre 2026). Premier `prepr` rouge à l'étape `mutate` : `TestDoctorBrainInReadOnlyDirectory` (`cmd/cortexdb-mcp-stdio`, hors du patch, identique à la base) a reçu `database is locked (5)` pendant le passage de couverture de `gremlins`, tous paquets en même temps ; seul, il passe 5 fois sur 5. Test instable de l'amont, noté. Second `prepr` : inventaire 230/230, `patchcov` 0, horloge, suite `-race` et éval vertes, exit 0 ; mais `gremlins` y a lu la couverture du cache (1,8 s au lieu de 71 s) et tous les mutants du patch ont dépassé le délai (`TIMED OUT`, résumé 0/0/0, sans valeur). `go clean -testcache && make mutate` seul : couverture 55 s, Killed 17 / Lived 1 / Not covered 0. Le survivant, `policy.go:264` (`len(meta) > 0` → `>= 0`), est équivalent : une map de métadonnées vide n'a aucune clé à refuser. Leçon : après un `prepr` rouge, relancer avec le cache des tests vidé, sinon `gremlins` calcule ses délais sur un temps de cache.

## PR à part `gc-11-rest-upsert-rows` (trou REST de `main`)

Trouvé le 21 octobre 2026 pendant la correction de la relecture adverse de #72 (tour 5), sur décision de Lauro sortie de #74 en PR à part, basée sur `main` (`cc98c6c`), tête `a833b4c`. Sur `main`, avant le code :

```
--- FAIL: TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP
    POST /v1/knowledge as "notes-secret" = 200, want 404
--- FAIL: TestAConfinedUpsertIsRefusedWithoutABrainToGuardWith
    an upsert with no brain to guard it answered 200 (served: true)
```

`precommit` OK contre `origin/main` (0 bloc). Mutations (`python3 -I mutate_round2.py gc11 <copie>`, 15 sur 15 tuées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Garde des écritures par identifiant non appelée | `pkg/httpapi/policy.go` : `if u, upserts := upsertRoutes[rr.String()]; upserts && false && !guard` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestAConfinedUpsertIsRefusedWithoutABrainToGuardWith, TestEverySpellingOfAnUpsertIDIsGuarded |
| Sans base, la garde laisse passer | `pkg/httpapi/upsert.go` : `		writeError(w, http.StatusInternalServerError,` | TestAConfinedUpsertIsRefusedWithoutABrainToGuardWith |
| Autres orthographes de l'identifiant non lues | `pkg/httpapi/upsert.go` : `range []string{top}` | TestEverySpellingOfAnUpsertIDIsGuarded |
| Identifiant vide envoyé à la lecture | `pkg/httpapi/upsert.go` : `			continue` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded |
| Lecture en panne prise pour une création | `pkg/httpapi/upsert.go` : `			writeError(w, http.StatusInternalServerError, err.Error())` | TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Erreur de lecture non rendue | `pkg/httpapi/upsert.go` : `	return nil` | TestAnUpsertWhoseRowCannotBeReadIsRefused, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Refus qui ne nomme pas le type de ligne | `pkg/httpapi/upsert.go` : `fmt.Sprintf("row %q not found", id)` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP |
| Clause collection d'un document non lue | `pkg/httpapi/upsert.go` : `resp.Knowledge.Collection == resp.Knowledge.Collection &&` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause user_id sur un document non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause scope sur un document non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause namespace sur un document non lue | `pkg/httpapi/upsert.go` : `true, nil` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause user_id d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause scope d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause namespace d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause collection sur une mémoire non lue | `pkg/httpapi/upsert.go` : `true, nil` | TestTheUpsertRowRuleHoldsEveryConfinedField |

### Second passage (relecture adverse, tour 1, tête `afca137`)

Rapport : `adversarial_review_gc11_r1.md` (1 bloquant, 4 remarques probables, 3 nits). Suite donnée :
- B1 corrigé : les cibles `supersedes` de `POST /v1/memory` (champ REST seul, absent du message gRPC) sont lues sous chaque orthographe ; cible étrangère et cible absente, même 404 avant toute écriture.
- L3 corrigé sur les deux ports : seul l'identifiant vide est sauté (`upsert.go`, `ownership.go` `guardKnowledgeUpsert`, `guardMemoryUpsert`).
- L5 corrigé : l'absence se lit sur les sentinelles `core.ErrNotFound` et `sql.ErrNoRows`, plus sur le texte.
- L2 nommé dans le corps : les entités et relations des écritures typées sont la classe des écritures du graphe, fermée par #73 et #74.
- L4 nommé dans le corps (déplacer la règle dans la façade, plus tard, si le propriétaire le veut).
- N6 : une lecture en panne reste 500 fixe (aucun test ne distingue 499 ou 504). N7 : six tests tombent sur `main` (deux ne compilent pas), test des clés non confinées ajouté, orthographe `Memory_ID` testée, fenêtre lecture puis écriture nommée.

Sur `main` (`cc98c6c`, copie exportée, tests qui compilent posés sans le code) : `TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP`, `TestAConfinedUpsertIsRefusedWithoutABrainToGuardWith`, `TestEverySpellingOfAnUpsertIDIsGuarded`, `TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP`, `TestABlankUpsertIDIsGuardedLikeAnyOther`, `TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC` : FAIL. `TestAnUnconfinedKeyUpsertsAnyRowOverHTTP` : PASS (comportement de `main` tenu).

Mutations (`python3 -I mutate_round2.py gc11 <copie>`, 24 sur 24 tuées ; `gc11g`, 2 sur 2) :

| Règle | Mutation | Test qui tombe |
| :- | :- | :- |
| Garde des écritures par identifiant non appelée | `pkg/httpapi/policy.go` : `if u, upserts := upsertRoutes[rr.String()]; upserts && false && !guard` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestAConfinedUpsertIsRefusedWithoutABrainToGuardWith, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded |
| Sans base, la garde laisse passer | `pkg/httpapi/upsert.go` : `		writeError(w, http.StatusInternalServerError,` | TestAConfinedUpsertIsRefusedWithoutABrainToGuardWith |
| Autres orthographes de l'identifiant non lues | `pkg/httpapi/upsert.go` : `return []string{top}` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestEverySpellingOfAnUpsertIDIsGuarded |
| Identifiant blanc pris pour une absence (L3, relecture adverse) | `pkg/httpapi/upsert.go` : `			continue` | TestABlankUpsertIDIsGuardedLikeAnyOther |
| Lecture en panne prise pour une création | `pkg/httpapi/upsert.go` : `		if err != nil && false {` | TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Erreur de lecture non rendue | `pkg/httpapi/upsert.go` : `}` | TestAnUpsertWhoseRowCannotBeReadIsRefused, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Absence non reconnue par sa sentinelle (L5, relecture adverse) | `pkg/httpapi/upsert.go` : `if false && errors.Is(err, core.ErrNotFound) || errors.Is(err, sql.Err` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUnconfinedKeyUpsertsAnyRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Refus qui ne nomme pas le type de ligne | `pkg/httpapi/upsert.go` : `fmt.Sprintf("row %q not found", id)` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP |
| Clause collection d'un document non lue | `pkg/httpapi/upsert.go` : `resp.Knowledge.Collection == resp.Knowledge.Collection &&` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause user_id sur un document non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause scope sur un document non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause namespace sur un document non lue | `pkg/httpapi/upsert.go` : `true, nil` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause user_id d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause scope d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause namespace d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause collection sur une mémoire non lue | `pkg/httpapi/upsert.go` : `true, nil` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Cibles supersedes non lues (B1, relecture adverse) | `pkg/httpapi/upsert.go` : `		return true` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Route des mémoires sans cibles supersedes | `pkg/httpapi/upsert.go` : `field: "memory_id", supersedes: false}` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Cible absente laissée passer (oracle) | `pkg/httpapi/upsert.go` : `if found && !permitted {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible étrangère laissée passer | `pkg/httpapi/upsert.go` : `if !found || !permitted && false {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Autres orthographes de supersedes non lues | `pkg/httpapi/upsert.go` : `if k != "supersedes" {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible blanche lue au lieu d'être sautée | `pkg/httpapi/upsert.go` : `(ligne retirée)` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible égale à la nouvelle mémoire lue | `pkg/httpapi/upsert.go` : `if target == "" {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible illisible prise pour un passage | `pkg/httpapi/upsert.go` : `		if err != nil && false {` | TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Mémoire à identifiant blanc non gardée en gRPC (L3) | `pkg/rpcserver/ownership.go` : `	_, _, err := guardMemory(ctx, db, id)` | TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |
| Document à identifiant blanc non gardé en gRPC (L3) | `pkg/rpcserver/ownership.go` : `	_, _, err := guardKnowledge(ctx, db, id)` | TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |

`precommit` OK contre `origin/main` (paquets `pkg/httpapi`, `pkg/rpcserver`), 0 bloc sans test.

### Troisième passage (relecture adverse, tour 2, tête `91ea926`)

Rapport : `adversarial_review_gc11_r2.md` (aucun bloquant, 2 remarques probables, 5 nits). Suite donnée :
- L1 : `TestAnUnconfinedKeyUpsertsAnyRowOverHTTP` tient le libellé de la façade et la mémoire écrite ; mutation « garde appelée pour une clé non confinée » tuée.
- L2 : `UpdateKnowledge` nommé et testé (identifiant blanc, gRPC).
- N1 : la garde lit la nouvelle mémoire et ses cibles dans `cortexdb.MemorySaveRequest`, comme le gestionnaire ; `supersedeTargets` et `containsString` retirés.
- N2, N3 : cible propre entourée d'espaces et libellé `supersede "<id>": not found` tenus par le test.
- N4 : `sql.ErrNoRows` retiré.
- N5 : différence gRPC (seconde lecture en panne, comportement de `main`) nommée dans le corps.
- Base absente : le test passe par `NewWithPolicy(nil, …)`, chemin public.

Mutations (`gc11`, 26 sur 26 tuées ; `gc11g`, 3 sur 3) :

| Règle | Mutation | Test qui tombe |
| :- | :- | :- |
| Garde des écritures par identifiant non appelée | `pkg/httpapi/policy.go` : `if u, upserts := upsertRoutes[rr.String()]; upserts && false && !guard` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestAConfinedUpsertIsRefusedWithoutABrainToGuardWith, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded |
| Sans base, la garde laisse passer | `pkg/httpapi/upsert.go` : `		writeError(w, http.StatusInternalServerError,` | TestAConfinedUpsertIsRefusedWithoutABrainToGuardWith |
| Autres orthographes de l'identifiant non lues | `pkg/httpapi/upsert.go` : `return []string{top}` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestEverySpellingOfAnUpsertIDIsGuarded |
| Identifiant blanc pris pour une absence (L3, relecture adverse) | `pkg/httpapi/upsert.go` : `			continue` | TestABlankUpsertIDIsGuardedLikeAnyOther |
| Lecture en panne prise pour une création | `pkg/httpapi/upsert.go` : `		if err != nil && false {` | TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Erreur de lecture non rendue | `pkg/httpapi/upsert.go` : `}` | TestAnUpsertWhoseRowCannotBeReadIsRefused, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Absence non reconnue par sa sentinelle (L5, relecture adverse) | `pkg/httpapi/upsert.go` : `if false && errors.Is(err, core.ErrNotFound) {` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUnconfinedKeyUpsertsAnyRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Refus qui ne nomme pas le type de ligne | `pkg/httpapi/upsert.go` : `fmt.Sprintf("row %q not found", id)` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP |
| Clause collection d'un document non lue | `pkg/httpapi/upsert.go` : `resp.Knowledge.Collection == resp.Knowledge.Collection &&` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause user_id sur un document non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause scope sur un document non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause namespace sur un document non lue | `pkg/httpapi/upsert.go` : `true, nil` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause user_id d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause scope d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause namespace d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause collection sur une mémoire non lue | `pkg/httpapi/upsert.go` : `true, nil` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Cibles supersedes non lues (B1, relecture adverse) | `pkg/httpapi/upsert.go` : `		return true` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Route des mémoires sans cibles supersedes | `pkg/httpapi/upsert.go` : `field: "memory_id", supersedes: false}` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Cible absente laissée passer (oracle) | `pkg/httpapi/upsert.go` : `if found && !permitted {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible étrangère laissée passer | `pkg/httpapi/upsert.go` : `if !found || !permitted && false {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Autre orthographe de l'identifiant prise pour la nouvelle mémoire (N1, tour 2) | `pkg/httpapi/upsert.go` : `if target == "" || target == save.MemoryID || target == fields["Memory` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Garde appelée pour une clé non confinée (L1, tour 2) | `pkg/httpapi/policy.go` : `				next(w, r)` | TestAnUnconfinedKeyUpsertsAnyRowOverHTTP |
| Libellé du refus d'une cible non épinglé (N3, tour 2) | `pkg/httpapi/upsert.go` : `fmt.Sprintf("supersede %q: withheld", target)` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible non rognée : blanche lue, propre entourée d'espaces refusée (N2, tour 2) | `pkg/httpapi/upsert.go` : `target = strings.Clone(target)` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible égale à la nouvelle mémoire lue | `pkg/httpapi/upsert.go` : `if target == "" {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible illisible prise pour un passage | `pkg/httpapi/upsert.go` : `		if err != nil && false {` | TestAnUpsertWhoseRowCannotBeReadIsRefused |
| UpdateKnowledge sans garde de la ligne (L2, tour 2) | `pkg/rpcserver/knowledge.go` : `	if err := guardKnowledgeUpsert(ctx, s.db, ""); err != nil {` | TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |
| Mémoire à identifiant blanc non gardée en gRPC (L3) | `pkg/rpcserver/ownership.go` : `	_, _, err := guardMemory(ctx, db, id)` | TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |
| Document à identifiant blanc non gardé en gRPC (L3) | `pkg/rpcserver/ownership.go` : `	_, _, err := guardKnowledge(ctx, db, id)` | TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |

`precommit` OK contre `origin/main`, 0 bloc sans test.

### Quatrième passage (relecture adverse, tour 3, tête `9e3ba10`)

Rapport : `adversarial_review_gc11_r3.md` (méthode révisée ; aucun bloquant, 3 remarques probables, 4 nits). Suite donnée :
- L1 : mémoire du seau global et document de la collection par défaut ajoutés à `TestTheUpsertRowRuleHoldsEveryConfinedField` ; mutations P et R tuées.
- L2 : `TestOnlyTheNotFoundSentinelIsAMissingRow` ; mutation H tuée.
- L3 : libellés et conflit avec #74 nommés dans le corps (section Scope) ; déplacement dans la façade toujours proposé.
- N1 : ordre des refus nommé dans le corps et dans le commentaire de `guardUpsert`.
- N2 : branche `db == nil` retirée (état injouable : `NewWithPolicy` reçoit toujours le cerveau), avec son test.
- N3 : le test des clés non confinées ne fige plus le libellé de la façade ; la mémoire écrite suffit à tuer la garde appelée pour tous.
- N4 : dépendance aux lectures par identifiant non filtrées par `CallerScope` écrite dans le commentaire.

Sur `main` : cinq tests tombent (`TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP`, `TestEverySpellingOfAnUpsertIDIsGuarded`, `TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP`, `TestABlankUpsertIDIsGuardedLikeAnyOther`, `TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC`), trois ne compilent pas.

Mutations (`gc11`, 28 sur 28 tuées ; `gc11g`, 3 sur 3) :

| Règle | Mutation | Test qui tombe |
| :- | :- | :- |
| Mémoire du seau global prise par une clé confinée à un utilisateur (L1, tour 3) | `pkg/httpapi/upsert.go` : `(resp.Memory.UserID == "" || scope.Permits(authz.FieldUserID, resp.Mem` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Document de la collection par défaut pris par une clé confinée (L1, tour 3) | `pkg/httpapi/upsert.go` : `(resp.Knowledge.Collection == "graphrag_chunks" || scope.Permits(authz` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Absence lue sur le texte de l'erreur (L2, tour 3) | `pkg/httpapi/upsert.go` : `if errors.Is(err, core.ErrNotFound) || strings.Contains(err.Error(), "` | TestOnlyTheNotFoundSentinelIsAMissingRow |
| Garde des écritures par identifiant non appelée | `pkg/httpapi/policy.go` : `if u, upserts := upsertRoutes[rr.String()]; upserts && false && !guard` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded |
| Autres orthographes de l'identifiant non lues | `pkg/httpapi/upsert.go` : `return []string{top}` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestEverySpellingOfAnUpsertIDIsGuarded |
| Identifiant blanc pris pour une absence (L3, relecture adverse) | `pkg/httpapi/upsert.go` : `			continue` | TestABlankUpsertIDIsGuardedLikeAnyOther |
| Lecture en panne prise pour une création | `pkg/httpapi/upsert.go` : `		if err != nil && false {` | TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Erreur de lecture non rendue | `pkg/httpapi/upsert.go` : `}` | TestAnUpsertWhoseRowCannotBeReadIsRefused, TestOnlyTheNotFoundSentinelIsAMissingRow, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Absence non reconnue par sa sentinelle (L5, relecture adverse) | `pkg/httpapi/upsert.go` : `if false && errors.Is(err, core.ErrNotFound) {` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUnconfinedKeyUpsertsAnyRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded, TestOnlyTheNotFoundSentinelIsAMissingRow, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Refus qui ne nomme pas le type de ligne | `pkg/httpapi/upsert.go` : `fmt.Sprintf("row %q not found", id)` | TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP |
| Clause collection d'un document non lue | `pkg/httpapi/upsert.go` : `resp.Knowledge.Collection == resp.Knowledge.Collection &&` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestEverySpellingOfAnUpsertIDIsGuarded, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause user_id sur un document non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause scope sur un document non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause namespace sur un document non lue | `pkg/httpapi/upsert.go` : `true, nil` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause user_id d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestABlankUpsertIDIsGuardedLikeAnyOther, TestAConfinedKeyCannotUpsertAnotherScopesRowOverHTTP, TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause scope d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause namespace d'une mémoire non lue | `pkg/httpapi/upsert.go` : `true &&` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Clause collection sur une mémoire non lue | `pkg/httpapi/upsert.go` : `true, nil` | TestTheUpsertRowRuleHoldsEveryConfinedField |
| Cibles supersedes non lues (B1, relecture adverse) | `pkg/httpapi/upsert.go` : `		return true` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Route des mémoires sans cibles supersedes | `pkg/httpapi/upsert.go` : `field: "memory_id", supersedes: false}` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP, TestAnUpsertWhoseRowCannotBeReadIsRefused |
| Cible absente laissée passer (oracle) | `pkg/httpapi/upsert.go` : `if found && !permitted {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible étrangère laissée passer | `pkg/httpapi/upsert.go` : `if !found || !permitted && false {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Autre orthographe de l'identifiant prise pour la nouvelle mémoire (N1, tour 2) | `pkg/httpapi/upsert.go` : `if target == "" || target == save.MemoryID || target == fields["Memory` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Garde appelée pour une clé non confinée (L1, tour 2) | `pkg/httpapi/policy.go` : `				next(w, r)` | TestAnUnconfinedKeyUpsertsAnyRowOverHTTP |
| Libellé du refus d'une cible non épinglé (N3, tour 2) | `pkg/httpapi/upsert.go` : `fmt.Sprintf("supersede %q: withheld", target)` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible non rognée : blanche lue, propre entourée d'espaces refusée (N2, tour 2) | `pkg/httpapi/upsert.go` : `target = strings.Clone(target)` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible égale à la nouvelle mémoire lue | `pkg/httpapi/upsert.go` : `if target == "" {` | TestASupersedeTargetOutsideTheKeysScopeIsNotFoundOverHTTP |
| Cible illisible prise pour un passage | `pkg/httpapi/upsert.go` : `		if err != nil && false {` | TestAnUpsertWhoseRowCannotBeReadIsRefused |
| UpdateKnowledge sans garde de la ligne (L2, tour 2) | `pkg/rpcserver/knowledge.go` : `	if err := guardKnowledgeUpsert(ctx, s.db, ""); err != nil {` | TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |
| Mémoire à identifiant blanc non gardée en gRPC (L3) | `pkg/rpcserver/ownership.go` : `	_, _, err := guardMemory(ctx, db, id)` | TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |
| Document à identifiant blanc non gardé en gRPC (L3) | `pkg/rpcserver/ownership.go` : `	_, _, err := guardKnowledge(ctx, db, id)` | TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC |

`precommit` OK contre `origin/main`, 0 bloc sans test.

## PR à part `gc-12-grpc-ingest-guard` (trou gRPC de `main`)

Trouvé par la relecture adverse de #73, tour 2 (B1, `adversarial_review_73_r2.md`). Lauro a choisi une PR à part, sur `main` (`cc98c6c`), parce que le défaut vient de `main` (21 octobre 2026, 23 h 40). Test écrit d'abord, sur `cc98c6c` sans le code : `TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection` tombait (`InsertGraphDocument over another collection's document: <nil>, want NOT_FOUND "knowledge \"sec\" not found"`, `the document is now "Overwritten by notes." in "notes", want it untouched`). Inventaire : `inventory_gc12.md` (43/43).

Relecture adverse, tour 1 (`adversarial_review_gc12_r1.md`, aucun bloquant) : L1 (identifiant blanc) nommé dans le corps, fermé par #82 ; L2 et N1 corrigés dans le corps ; N2 et N3 : le test relit le document créé dans `notes` et envoie un contenu vide sur un identifiant étranger, preuve par deux mutations (création hors de la collection demandée, garde sautée pour un contenu vide, cette dernière M10 du relecteur, qui survivait).

Mutations (`python3 -I mutate_round2.py gc12 <copie>`, 6 sur 6 tuées) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| InsertGraphDocument sans garde de ligne | `pkg/rpcserver/graphrag.go` : `(ligne retirée)` | TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection |
| Garde appelée sans l'identifiant de la requête | `pkg/rpcserver/graphrag.go` : `guardKnowledgeUpsert(ctx, s.db, "")` | TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection |
| Refus de la garde ignoré | `pkg/rpcserver/graphrag.go` : `_ = guardKnowledgeUpsert(ctx, s.db, req.GetId())` | TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection |
| Garde sautée quand le contenu est vide (relecture, N3) | `pkg/rpcserver/graphrag.go` : `if err := guardKnowledgeUpsert(ctx, s.db, req.GetId()); err != nil && ` | TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection |
| Document créé hors de la collection demandée (relecture, N2) | `pkg/rpcserver/graphrag.go` : `Collection:   "elsewhere",` | TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection |
| Toute écriture nommée refusée, la sienne comprise | `pkg/rpcserver/graphrag.go` : `	}` | TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection |

### Niveau 3 du lot C, 22 octobre 2026 (têtes `f4c5aa7`, `2384236`, `be0c89d`)

`make diffmain` sur #77 contre `2384236` : identique (21 lignes de scénario). `prepr` en série, un seul passage PostgreSQL à la fois, exit 0 sur les trois têtes, aucune ligne `FAIL`. #75 : `patchcov` 0 bloc, `gremlins` Killed 40 / Lived 0 / Not covered 0. #76 : `patchcov` 0 bloc, Killed 112 / Lived 7 / Not covered 4, les sept « LIVED » dans `cmd/cortexdb-mcp-stdio/sweep_expired.go` (paquet `main`, déjà rejoués à la main, plus haut). #77 : `patchcov` 0 bloc et 2 dans `main()`, Killed 107 / Lived 5 / Not covered 6 ; les cinq « LIVED » dans `cmd/cortexdb-grpc/review_gate.go` et `cmd/cortexdb-mcp-stdio/embedder.go` (paquets `main`) et « NOT COVERED » `review.go:76` et `78` (les lignes 68 et 70 d'avant, décalées) : même classe que plus haut, rejoués à la main par le jeu `77` (motif `case at == "" && f.Status == prev.status …`).

### Lignes Codecov évitables, #76 et #77 (22 octobre 2026, après la poussée)

Question de Lauro sur les lignes non couvertes. Codecov comptait 15 lignes sur #76 et 23 sur #77. Dans `store_postgres_aggregate.go`, 12 lignes étaient déclarées « PostgreSQL » à tort : `pgAggregateWhere` et `pgBindEach` rendent du texte SQL et des arguments, sans connexion. Correctif (décision de Lauro : « corrige #76 et #77, passe les 2 en brouillon ») : #76 et #77 repassées en brouillon, un test sans serveur par PR dans `pkg/core/store_postgres_aggregate_where_test.go`. #76 `73ed740` : `TestThePostgresAggregateWhereExcludesDocumentsByOneSubquery`, jeu `76cc`, 3/3 (`if false`, `if true` sur `ExcludeDocuments`, `?` non lié). #77 `4fefb91`, rebasée sans conflit : `TestThePostgresAggregateWhereKeepsOnlyTheSelectedDocuments`, jeu `77cc`, 3/3 (`if false`, `if true` sur `OnlyDocuments`, `pgBindEach` sans lien). Les lignes qui restent ont besoin d'un serveur (verrou consultatif, `CREATE INDEX`) ou sont dans `main` : les 3 lignes de `cortexdb-grpc` suivent le modèle du plancher sémantique, construit en ligne dans `main` ; les sortir demande de refaire `main`, hors sujet. Corps : 57 mutations sur #76, 51 sur #77.
