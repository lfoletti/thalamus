# Lot C : preuve « failing first »

Vérification du 19 octobre 2026 sur la pile `gc-06-review-status` (contient les trois correctifs). Les tests utilisent des champs et fonctions nouveaux et ne compilent pas sur `main` ; la preuve est par règle : chaque règle est désactivée par une mutation qui compile, puis `go test ./pkg/cortexdb/ -run 'Reingest|Ingest|Dedupe|Hash|Expir|Sweep|Review|Gate|Reviewer'`. Vingt mutations, vingt tuées, puis sept de plus après les rapports Codecov sur #75 et #76, puis douze après celui de #77, puis quatre pour les chemins d'erreur (cerveau fermé) : quarante-deux tuées, deux survivantes classées (voir les notes). Passage automatique `gremlins unleash --diff origin/main` sur la tête 06 (19 octobre 2026, 5 min 34 s) : 91 mutants, 76 tués, 15 survivants, 5 non couverts. Les 15 survivants, lus un à un : 9 dans les paquets `main`, tous tués à la main (limite de l'outil sur ces paquets) ; 1 vrai manque, `reviewed_at` d'une mémoire non stocké (`review.go`, `normalizeReviewInAny`), test ajouté et mutation tuée ; 1 ligne amont voisine (`retrievalFiltersEmpty`, champ `Namespace`), hors patch ; 3 équivalents sans effet observable (durée du cache d'expiration, `len > 0` sur un JSON vide, capacité d'une map). Une mutation a survécu et la règle est classée défensive (voir la note). Script : `mutate_lotC.py` (archivé ici).

| PR | Règle désactivée | Test qui échoue |
| :- | :- | :- |
| GC-08 | hash non stocké dans les métadonnées | `TestIngestDocumentKeepsTheContentHash`, `TestDedupeRefusesTheSameContentUnderAnotherID` |
| GC-08 | ré-ingestion identique jamais détectée (`sameIngest && false`) | `TestAnIdenticalReingestWritesNothing` |
| GC-08 | `dedupe` ignoré | `TestDedupeRefusesTheSameContentUnderAnotherID` |
| GC-05 | `expires_at` invalide accepté | `TestABadExpiresAtIsRefused` |
| GC-05 | fuseau non ramené en UTC | `TestABadExpiresAtIsRefused` |
| GC-05 | document expiré toujours servi | `TestAnExpiredDocumentIsLeftOutOfSearch` |
| GC-05 | balayage qui ne supprime rien | `TestSweepExpiredRemovesExpiredDocumentsAndMemories` |
| GC-05 | essai à blanc qui supprime | `TestSweepExpiredRemovesExpiredDocumentsAndMemories` |
| GC-05 | mémoires non balayées | `TestSweepExpiredRemovesExpiredDocumentsAndMemories` |
| GC-06 | relecteur non exigé pour une décision | `TestADecisionNamesItsReviewer` |
| GC-06 | statut inventé accepté | `TestADecisionNamesItsReviewer` |
| GC-06 | `reviewed_at` invalide accepté | `TestADecisionNamesItsReviewer` |
| GC-06 | porte ignorée, documents (`return true`) | `TestTheReviewGateHidesAProposedDocumentUntilAccepted` |
| GC-06 | porte toujours active, documents (opt-in cassé) | `TestWithoutTheGateEveryReviewedRowIsServed` |
| GC-06 | porte ignorée, mémoires | `TestTheReviewGateLeavesAProposedMemoryOutOfRecall` |
| GC-06 | porte toujours active, mémoires | `TestWithoutTheGateEveryReviewedRowIsServed` |
| GC-06 | filtre `review_status=<statut>` ignoré | `TestWithoutTheGateEveryReviewedRowIsServed`, `TestTheReviewGateHidesAProposedDocumentUntilAccepted` |
| GC-06 | filtre `review_status=accepted` ignoré | les deux mêmes |
| GC-06 | `WithReviewGate` inerte | `TestTheReviewGateHidesAProposedDocumentUntilAccepted`, `TestTheReviewGateLeavesAProposedMemoryOutOfRecall` |
| GC-06 | statut absent de `MemoryRecord` | `TestTheReviewGateLeavesAProposedMemoryOutOfRecall` |
| GC-08 (Codecov) | contrôle du nombre de clés de métadonnées retiré | `TestAnIdenticalReingestWritesNothing`, cas « clé retirée » |
| GC-05 (Codecov) | expiration ignorée au premier étage PPR (`graphrag_ppr.go`, premier site) | `TestTheGraphWalkLeavesAnExpiredPassageOut` |
| GC-05 (Codecov) | expiration ignorée sur les chunks parcourus (second site) | `TestTheGraphWalkLeavesAnExpiredPassageOut` |
| GC-05 (Codecov) | `expires_at` vide gardé sur l'enregistrement | `TestAnEmptyExpiresAtClearsTheExpiry` |
| GC-05 (Codecov) | commande : mauvais drapeau sans code 2 | `TestTheSweepCommandReportsAndRemovesExpiredRows` |
| GC-05 (Codecov) | commande : échec d'ouverture sans code 1 | idem |
| GC-05 (Codecov) | commande : `--dry-run` ignoré | idem |
| GC-06 (Codecov) | porte retirée sur le chemin sémantique (`memory_api.go`) | `TestTheReviewGateHoldsOnEveryMemoryPath`, cas `semantic` |
| GC-06 (Codecov) | porte retirée sur le chemin graphe (`memory_api.go`) | idem, cas `graph` |
| GC-06 (Codecov) | porte retirée sur le chemin PPR (`graphrag_ppr.go`) | idem, cas `ppr` |
| GC-06 (Codecov) | porte retirée sur le chemin lexical CJK (`lexical_cjk.go`) | idem, cas `cjk` |
| GC-06 (Codecov) | `memory_update` sans contrôle de relecture | `TestADecisionNamesItsReviewer` |
| GC-06 (Codecov) | `reviewed_at` fourni remplacé par maintenant | `TestADecisionNamesItsReviewer` |
| GC-06 (Codecov) | `source_url` absent du nœud de document | `TestADecisionNamesItsReviewer` |
| GC-06 (Codecov) | `review_status` de la requête n'écrase pas celui du plan | `TestReviewFiltersMergeWithTheRestOfThePlan` |
| GC-06 (Codecov) | stdio : `CORTEXDB_REVIEW_GATE=1` sans effet | `TestTheReviewGateIsReadFromTheEnvironment` (stdio) |
| GC-06 (Codecov) | stdio : valeur invalide silencieuse | idem |
| GC-06 (Codecov) | grpc : valeur invalide silencieuse | `TestTheReviewGateIsReadFromTheEnvironment` (grpc) |
| GC-06 (Codecov) | grpc : valeur ignorée | idem |
| GC-05 (erreurs) | porte illisible qui cache tout | `TestAnUnreadableBrainIsNotAnEmptyBrain` |
| GC-05 (erreurs) | balayage qui dit « fait » sans pouvoir lister | idem |
| GC-08 (erreurs) | recherche de hash en échec comptée comme « pas de doublon » | **survivante** : sur un cerveau fermé l'écriture suivante échoue aussi ; le test tient le contrat de bout en bout, la ligne est une propagation d'erreur simple |
| GC-06 (erreurs) | `checkReview` sans statut renvoie une erreur | `TestADecisionNamesItsReviewer` |

Note (Codecov, #77) : la mutation « filtre `document_ids` ignoré par la porte » (`expiry.go`, `allowDocument`) a survécu sur le chemin lexical et sur le chemin PPR : chaque appelant applique déjà `allowDocumentID` avant la porte. La ligne est gardée pour que `allowDocument` soit complète seule, et classée défensive. Le cas PPR ajouté pour l'éprouver a été retiré du test, car il ne prouvait rien.

Note de méthode (Codecov, #75) : un cas « clé de métadonnées ajoutée » ne prouvait pas la règle `len`, car la boucle de comparaison refuse déjà une clé en plus. Seule une clé en moins passe par la règle. Choisir le cas qui n'est refusé que par la règle visée, puis le vérifier par mutation.

Note de méthode : au premier passage, la mutation « ré-ingestion identique » a paru survivre. Le filtre `-run` était `Ingest|...`, sensible à la casse, et ne retenait pas `TestAnIdenticalReingestWritesNothing`. Rejouée avec `Reingest` dans le filtre : tuée. Un filtre `-run` se vérifie avec `-v` avant de conclure qu'une mutation survit.

Les PR du lot C portent chacune une section « Failing first » qui reprend les lignes de ce tableau.

## Troisième passe (19 octobre 2026, soir) : couverture complète du patch

Déclencheur : rapports Codecov finaux à 91,4 % (#75), 83,0 % (#76) et 87,4 % (#77), et décision de Lauro de ne plus rien pousser sans vérification complète en local. Chaque ligne « injouable » a été relue. Presque toutes se provoquent de façon déterministe en SQLite depuis le test, sans faux pilote :

| Chemin | Panne provoquée | Test |
| :- | :- | :- |
| `rows.Scan` d'un document ou d'une mémoire | ligne à `id` NULL, après retrait des déclencheurs `change_log` de la table | `TestARowTheSweepCannotReadIsReported` |
| métadonnées de mémoire qui ne sont pas du JSON | ligne insérée avec `'not json'` | idem |
| requête de la liste des mémoires qui échoue | `ALTER TABLE messages RENAME COLUMN metadata` | `TestASweepStopsAtTheFirstFailure` |
| suppression qui échoue en cours de balayage (documents, puis mémoires) | `CREATE TRIGGER … BEFORE DELETE … RAISE(ABORT)` | idem, et `TestTheSweepCommandReportsAndRemovesExpiredRows` (code 1, rapport imprimé) |
| cerveau par défaut sous `$HOME` | `HOME` temporaire, `CORTEXDB_PATH` vide | `TestTheSweepCommandReportsAndRemovesExpiredRows` |
| table des blocs illisible à la ré-ingestion identique (#75) | `ALTER TABLE embeddings RENAME COLUMN doc_id` | `TestAnUnreadableChunkTableFailsTheUnchangedPath` |
| filtre `document_ids` et bloc sans document dans `allowDocument` (#77) | appel direct | `TestAnUnreadableBrainIsNotAnEmptyBrain` |
| `CORTEXDB_REVIEW_GATE` côté gRPC (#77) | `appendReviewGate` sorti de `main` | `TestTheReviewGateOptionFollowsTheEnvironment` |

Deux lignes restaient hors d'atteinte sans faux pilote : `rows.Err()` après la dernière ligne, dans les deux boucles. Les boucles passent maintenant par `eachRow(rows, each)`, qui ferme le curseur et rend la première erreur, celle de `each` ou celle du curseur en dernière instruction : plus de branche séparée, et l'erreur de `each` est tuée par les tests à `id` NULL. Un garde-fou injouable de #75 (`len(row) == 0` sur un `RETURN` d'une colonne) est retiré.

Reste non couvert, déclaré dans chaque corps : les instructions de `main` (2 dans `cortexdb-mcp-stdio`, le `if` de répartition et l'`os.Exit`, comme chaque autre sous-commande du fichier ; 3 dans `cortexdb-grpc`, l'appel à `appendReviewGate` et la sortie sur erreur).

Mutations à la main de cette passe, toutes tuées (`go test -run`, harnais Python, arbre restauré après chaque essai) :

| Mutation | Test qui tombe |
| :- | :- |
| 08 : erreur de la liste des blocs avalée | `TestAnUnreadableChunkTableFailsTheUnchangedPath` |
| 05 : `eachRow` avale l'erreur de `each` | `TestARowTheSweepCannotReadIsReported` |
| 05 : erreur de `Scan` ignorée (documents) | idem |
| 05 : erreur de `Scan` ignorée (mémoires) | idem |
| 05 : métadonnées non JSON traitées comme une erreur | idem |
| 05 : liste des mémoires en échec rend quand même un rapport | `TestASweepStopsAtTheFirstFailure` |
| 05 : suppression de document en échec laissée dans le rapport | idem |
| 05 : suppression de mémoire en échec laissée dans le rapport | idem |
| 05 : bloc sans document traité comme expiré | `TestAnUnreadableBrainIsNotAnEmptyBrain` |
| 05 : commande, code 0 après un balayage en échec | `TestTheSweepCommandReportsAndRemovesExpiredRows` |
| 05 : commande, chemin par défaut faux | idem |
| 06 : `appendReviewGate` ignore une valeur fausse | `TestTheReviewGateOptionFollowsTheEnvironment` |
| 06 : gate demandé mais option non ajoutée | idem |
| 06 : filtre `document_ids` ignoré par `allowDocument` | `TestAnUnreadableBrainIsNotAnEmptyBrain` |
| 06 : bloc sans document exclu | idem |
| 06 : erreur de `Scan` du gate ignorée | `TestARowTheSweepCannotReadIsReported` |

Passage automatique `gremlins unleash --diff origin/main` sur la tête 06 (`6bb43e7`), 19 octobre 2026, 5 min 40 s : 95 mutants, 76 tués, 14 survivants, 5 non couverts. Les 5 non couverts sont les instructions de `main` et la constante `documentExpiryCacheTTL`. Les 14 survivants, lus un à un : 10 dans les paquets `main` (`review_gate.go` 17, 21, 31 ; `embedder.go` 58, 60 ; `sweep_expired.go` 35, 39, 46, 48, 58), tous tués en rejouant la mutation à la main dans une copie de travail séparée (limite connue de `gremlins` sur ces paquets) ; 3 équivalents sans effet observable (`expiry.go` 83, durée du cache, deux mutants ; `knowledge_tx.go` 211, capacité d'une map) ; 1 ligne amont voisine du patch (`retrieval_plan.go` 235, champ `Namespace` de `retrievalFiltersEmpty`). Aucun vrai manque.

Vérification complète du 19 octobre 2026, 19 h 27 à 19 h 55 UTC, séquentielle (`scratchpad/verify_lotC.sh`), toute verte :

| Contrôle | Résultat |
| :- | :- |
| `patchcov` tête 06 contre `origin/main` (diff empilé) | 3 blocs, tous dans `main` (`cmd/cortexdb-mcp-stdio/main.go:204`, `cmd/cortexdb-grpc/main.go:118-121`) |
| `gremlins unleash --diff origin/main` tête 06 | 95 mutants, 76 tués, 14 survivants lus (ci-dessus), 5 non couverts (`main`, constante) |
| `go test -race` tête 06 | `pkg/cortexdb` 238 s, `cmd/cortexdb-mcp-stdio` 25 s, `cmd/cortexdb-grpc` 8 s, verts |
| `go test -race` tête 05 (`17f9208`) | `pkg/cortexdb` 248 s, `cmd/cortexdb-mcp-stdio` 9 s, verts |
| `go test -race` tête 08 (`bd96d99`) | `pkg/cortexdb` 257 s, vert |
| `go test -race -timeout 25m ./...` tête 06 | 30 paquets verts, 0 échec |
| `go test ./pkg/eval -run TestLexicalRetrievalQuality` | vert |


## Amendement du 20 octobre 2026 (revue du propriétaire)

Chaque point de la revue a son test, écrit avant le code, et sa mutation à la main (`python3 -I docs/pr/lot_C/mutate_amend.py 08|05|06 <worktree>`), fichier restauré après chacune.

### #75, hash de contenu (base `origin/main` à v2.134.1)

Huit mutations, huit tuées, `go test ./pkg/cortexdb/ -run "Ingest|Dedupe|Hash|Unchanged|Reingest|ChunkIDs|Unreadable"` :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Dédup hors collection | `pkg/cortexdb/graphrag_tool_ingest.go` : `			return id, nil` | TestDedupeIsDecidedWithinTheCollection |
| Découpage non comparé | `pkg/cortexdb/graphrag_tool_ingest.go` : `metadata[chunkSizeKey] = "0"` | TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking |
| Collection non comparée | `pkg/cortexdb/graphrag_tool_ingest.go` : `if storedIn == ingestOpts.Collection || true {` | TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking |
| Mise à jour sans re-hash | `pkg/cortexdb/knowledge_api.go` : `stampIngestMarks(metadata, existing.Content, resolvedIngestO` | TestTheContentHashFollowsTheContentThroughKnowledgeWrites |
| knowledge_save sans hash | `pkg/cortexdb/knowledge_api.go` : `stampIngestMarks(metadata, "", resolvedIngestOptions(req.Col` | TestTheContentHashFollowsTheContentThroughKnowledgeWrites |
| Hash absent du nœud à la sauvegarde | `pkg/cortexdb/knowledge_tx.go` : `_ = hash` | TestTheContentHashFollowsTheContentThroughKnowledgeWrites |
| Ré-ingestion sans retrait des anciens chunks | `pkg/cortexdb/graphrag_tool_ingest.go` : `if err := t.db.applyKnowledgeMutation(ctx, req.DocumentID, f` | TestAnIngestWritesEverythingOrNothing |
| Chunks triés par texte | `pkg/cortexdb/graphrag_tool_ingest.go` : `		return chunks[i].ID < chunks[j].ID` | TestTheChunkIDsOfAnUnchangedReingestComeInChunkOrder |

Arbre restauré, suite verte : True

Couverture du patch (`make patchcov`, base `origin/main`) : zéro bloc non couvert.

### #76, expiration (base `gc-08-content-hash`)

Treize mutations, treize tuées, `go test ./pkg/cortexdb/ ./cmd/cortexdb-mcp-stdio/ -run "Expir|Sweep|Chunk|Vector|Indexed|ToolCount"` (cinq ajoutées le 20 octobre 2026 au soir pour le balayage du cerveau partagé, qui remplace le refus en mode distant, et une pour les voies de `cortex_query`, trouvée pendant l'analyse de la liste blanche, et une pour la fermeture MCP de `sweep_expired`, que `precommit` a montrée sans test le 20 octobre au soir : le test appelle maintenant l'outil par une session MCP en mémoire) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Chemin vectoriel non filtré | `pkg/cortexdb/retrieval_plan.go` : `if db.allowDocument(ctx, filters, result.DocID) || true {` | TestAnEmptyExpiresAtClearsTheExpiry, TestAnExpiredDocumentIsLeftOutOfSearch, TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo, TestTheGraphWalkLeavesAnExpiredPassageOut |
| cortex_query non filtré | `pkg/cortexdb/query_api.go` : ligne `db.filterScoredEmbeddings` retirée | TestAnExpiredDocumentIsLeftOutOfCortexQuery |
| Fermeture MCP sans balayage | `pkg/cortexdb/expiry_tool.go` : `return &ExpiredSweepReport{}, nil` | TestTheSweepIsAToolOfTheSharedBrain |
| Outils par chunks non filtrés | `pkg/cortexdb/graphrag_tool_helpers.go` : `if !ok {` | TestTheChunkToolsLeaveAnExpiredDocumentOut |
| expand_graph : nœuds gardés | `pkg/cortexdb/graphrag_tool_search.go` : `&& docID == "\x00" {` | TestTheChunkToolsLeaveAnExpiredDocumentOut |
| expand_graph : arêtes gardées | `pkg/cortexdb/graphrag_tool_search.go` : `		}` | TestTheChunkToolsLeaveAnExpiredDocumentOut |
| Balayage local malgré CORTEXDB_REMOTE | `cmd/cortexdb-mcp-stdio/sweep_expired.go` : `if addr, token, ok := remoteConfigured(); ok && false {` | TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| dry_run non transmis au cerveau partagé | `cmd/cortexdb-mcp-stdio/sweep_expired.go` : `cortexdb.SweepExpiredToolRequest{DryRun: false}` | TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| Outil sweep_expired non déclaré | `pkg/cortexdb/graphrag_tool_defs.go` : ligne `expiryToolDefinitions()` retirée | TestTheSweepIsAToolOfTheSharedBrain, TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet, TestToolCountIsUnchangedByTheCursorWork |
| Outil sweep_expired non dispatché | `pkg/cortexdb/expiry_tool.go` : `if name != "\x00" {` | TestTheSweepIsAToolOfTheSharedBrain, TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| dry_run de l'outil ignoré | `pkg/cortexdb/expiry_tool.go` : `return db.SweepExpired(ctx, time.Now(), false)` | TestTheSweepIsAToolOfTheSharedBrain, TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| Index sur une autre expression | `pkg/core/store_init.go` : `ON documents (author);` | TestTheDocumentExpiryIsIndexed |
| Prédicat que l'index ne sert pas | `pkg/cortexdb/expiry.go` : `` FROM documents WHERE ` + expr + ` IS NOT NULL`` | TestTheDocumentExpiryIsIndexed |

Arbre restauré, suite verte : True

Couverture du patch (`make patchcov`, base `gc-08-content-hash`) : zéro bloc non couvert hors `main` et hors la liste DDL PostgreSQL (`pkg/core/store_postgres.go`), que rien ne lance ici. Les deux sont déclarés dans la section `## Codecov` du corps.

Incident : le script de mutation tournait en arrière-plan pendant un `git commit --amend` du même arbre. La mutation « chemin vectoriel non filtré » (`|| true`) a été commise, puis détectée par les tests de #77 rebasée dessus, et retirée. Règle retenue : jamais de mutation et d'amendement en parallèle sur le même worktree.

### #77, statut de relecture (base `gc-05-expires-at`)

Quinze mutations, quinze tuées, `go test ./pkg/cortexdb/ -run "Review|Gate|Reviewer|Cleared|Acceptance|Indexed|Decision|Expir"` (les deux « relecteur forcé sur une proposition » ajoutées après le passage automatique, voir plus bas ; les deux « acceptation » ajoutées le 20 octobre 2026 au soir, quand la fusion des métadonnées a remplacé le simple report des clés système) :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Relecteur du document : le client, pas la clé | `pkg/cortexdb/review.go` : `	at, err := checkReview(status, metadata[reviewedByKey]` | TestTheReviewerIsTheKeyThatDecided |
| Relecteur forcé sur une proposition (document) | `pkg/cortexdb/review.go` : `	at, err` | TestTheReviewerIsTheKeyThatDecided |
| Relecteur forcé sur une proposition (mémoire) | `pkg/cortexdb/review.go` : `	by, _ :=` | TestTheReviewerIsTheKeyThatDecided |
| Relecteur de la mémoire : le client, pas la clé | `pkg/cortexdb/review.go` : `	by, _ := stringFromAny` | TestTheReviewerIsTheKeyThatDecided |
| Effacement admis sur un document | `pkg/cortexdb/review.go` : `		if prev.status != "" && false {` | TestAReviewDecisionIsNotClearedSilently |
| Effacement admis sur une mémoire | `pkg/cortexdb/review.go` : `		if prev.status != "" && false {` | TestAReviewDecisionIsNotClearedSilently |
| Décision non reportée quand la requête n'en parle pas | `pkg/cortexdb/review.go` : `		if prev.status != "" && false {` | TestAReviewDecisionIsNotClearedSilently, TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate |
| Statut non chaîne lu comme absent | `pkg/cortexdb/review.go` : `if present && raw != nil && !isString && false {` | TestAReviewDecisionIsNotClearedSilently |
| reviewed_at redaté à chaque sauvegarde | `pkg/cortexdb/review.go` : `reviewedBy == prev.by && prev.at != "" && false {` | TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate |
| expires_at perdu à l'acceptation | `pkg/cortexdb/review.go` : `if at, _ := stringFromAny(before[expiresAtKey]); at != "" &&` | TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate (cas « mise à jour sans relecture » ajouté : depuis la fusion, l'acceptation ne passe plus par `carryExpiry`, et la mutation survivait) |
| Acceptation qui remplace au lieu de fusionner | `pkg/cortexdb/knowledge_api.go` : `case reviewNamed(req.Metadata) && false:` | TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate |
| Décision stockée gardée sous la nouvelle | `pkg/cortexdb/review.go` : `delete(stored, reviewedAtKey)` retiré | TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate |
| Porte ouverte sur panne | `pkg/cortexdb/expiry.go` : `return true` | TestTheGateFailsClosedWhenTheBrainCannotBeRead |
| Index review_status sur une autre expression | `pkg/core/store_init.go` : `idx_documents_review_status ON documents (author);` | TestTheDocumentExpiryIsIndexed, TestTheReviewStatusIsIndexed |
| Prédicat review_status que l'index ne sert pas | `pkg/cortexdb/expiry.go` : `` > '' OR ` + review + ` IS NOT NULL`` | TestTheDocumentExpiryIsIndexed, TestTheReviewStatusIsIndexed |

Couverture du patch (`make patchcov`, base `gc-05-expires-at`) : zéro bloc non couvert hors `main` de `cortexdb-grpc` (trois lignes, appel de `appendReviewGate` et sortie sur erreur) et hors la liste DDL PostgreSQL. Déclarés dans la section `## Codecov` du corps.

### Rejeu des mutations de la première version (tête `gc-06-review-status` à `0b8dc71`)

`mutate_lotC.py` rejoué sur la tête de la pile après l'amendement : vingt mutations, vingt tuées (premier test tombé par ligne) :

| Mutation | Test qui échoue |
| :- | :- |
| `08 hash not stored` | TestIngestDocumentKeepsTheContentHash |
| `08 unchanged never detected` | TestAnUnreadableCollectionTableFailsTheUnchangedPathAndDedupe |
| `08 dedupe ignored` | TestDedupeRefusesTheSameContentUnderAnotherID |
| `05 bad expires_at accepted` | TestABadExpiresAtIsRefused |
| `05 zone not normalized` | TestABadExpiresAtIsRefused |
| `05 expired still served` | TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo |
| `05 sweep removes nothing` | TestSweepExpiredRemovesExpiredDocumentsAndMemories |
| `05 dry run deletes` | TestSweepExpiredRemovesExpiredDocumentsAndMemories |
| `05 memories not swept` | TestSweepExpiredRemovesExpiredDocumentsAndMemories |
| `06 reviewer not required` | TestADecisionNamesItsReviewer |
| `06 made-up status accepted` | TestADecisionNamesItsReviewer |
| `06 bad reviewed_at accepted` | TestADecisionNamesItsReviewer |
| `06 gate ignored (documents)` | TestTheReviewGateHoldsOnTheVectorPath |
| `06 gate always on (documents)` | TestWithoutTheGateEveryReviewedRowIsServed |
| `06 gate ignored (memories)` | TestTheReviewGateLeavesAProposedMemoryOutOfRecall |
| `06 gate always on (memories)` | TestWithoutTheGateEveryReviewedRowIsServed |
| `06 filter by status ignored` | TestWithoutTheGateEveryReviewedRowIsServed |
| `06 filter accepted ignored` | TestWithoutTheGateEveryReviewedRowIsServed |
| `06 WithReviewGate no-op` | TestTheReviewGateHoldsOnTheVectorPath |
| `06 status not on MemoryRecord` | TestAReviewDecisionIsNotClearedSilently |

### Passage automatique sur #75 (`gremlins unleash --diff origin/main`, 20 octobre 2026)

29 tués, 1 survivant, 0 non couvert. Survivant : `graphrag_tool_ingest.go:271:12`, comparateur de tri `a < b` → `a <= b` sur `chunk_index`. Équivalent : les indices sont distincts par document, l'ordre rendu est le même. Classé sans effet observable. Couverture du patch : 0 bloc sans test.

### Passage automatique sur #76 (`gremlins unleash --diff gc-08-content-hash`, 20 octobre 2026)

`prepr` s'arrête à `precommit` sur le bloc DDL PostgreSQL (déclaré) ; étapes suivantes à la main. Dernier passage le 20 octobre 2026 au soir, sur la tête `963c41c` (balayage distant, porte sur `cortex_query`) : `precommit` a montré la fermeture MCP de `sweep_expired` (`expiry_tool.go:61-63`) sans test, d'où le test par session MCP et la treizième mutation ci-dessus (tête `3c7a9fd`, test seul). Gremlins : 48 mutants, 34 tués, 9 survivants, 5 non couverts. Suite complète `-race` et éval lexicale : voir le journal de la chaîne.

- Non couverts : `main.go:204` (trois mutants, répartition de `--sweep-expired` dans `main`, déclarée) ; `expiry.go:33` (constante `documentExpiryCacheTTL`, durée du cache, sans effet observable, déjà classée le 19 octobre) ; `sweep_expired.go:91` (constante `remoteSweepTimeout`, délai de l'appel distant, même classe).
- Survivants dans `expiry.go:79` (deux) : test de fraîcheur du cache (`c.until != nil`, `<` → `<=`) ; effet sur le nombre de rechargements seulement, pas sur les lignes servies. Classés équivalents.
- Survivants dans `cmd/cortexdb-mcp-stdio/sweep_expired.go` (sept : lignes 42, 46, 53, 56, 67, 76, 81, soit le chemin par défaut, l'ouverture en échec, le rapport local, l'échec du balayage, la connexion, l'appel distant, le rapport illisible) : paquet `main`, où l'outil ne tue rien. Rejoués à la main (`mutate_main_05.py`, jeu réécrit le 20 octobre au soir pour le balayage distant), onze tués :

| Mutation | Test qui échoue |
| :- | :- |
| `distant ignoré` : `remoteConfigured(); !ok {` | TestTheSweepCommandReportsAndRemovesExpiredRows, TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| `connexion en échec ignorée` : `if err == nil {` | TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| `appel distant en échec ignoré` : `if err == nil {` | TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| `rapport distant illisible ignoré` : `return 0` | TestAnUnreadableReportFromTheSharedBrainIsAFailure (test ajouté : la mutation survivait, le serveur d'essai répond toujours un rapport lisible ; un faux `ToolsService` répond du texte) |
| `rapport distant non écrit` : `return 0` | TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| `chemin par défaut` : `if dbPath != "" {` | TestTheSweepCommandReportsAndRemovesExpiredRows |
| `ouverture en échec ignorée` : `if err == nil {` | TestTheSweepCommandReportsAndRemovesExpiredRows |
| `JSON non écrit` : `if asJSON { return }` | TestTheSweepCommandReportsAndRemovesExpiredRows, TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| `rapport texte non écrit` : `fmt.Fprintf(io.Discard, …)` | TestTheSweepCommandReportsAndRemovesExpiredRows, TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet |
| `rapport local non écrit` : `if report == nil {` | TestTheSweepCommandReportsAndRemovesExpiredRows |
| `échec du balayage ignoré` : `if err == nil {` | TestTheSweepCommandReportsAndRemovesExpiredRows |

### Passage automatique sur #77 (`gremlins unleash --diff gc-05-expires-at`, 20 octobre 2026)

`prepr` s'arrête à `precommit` sur le bloc DDL PostgreSQL (déclaré) ; étapes suivantes à la main. Premier passage (fusion des métadonnées) : 55 tués, 10 survivants, 1 non couvert (`cmd/cortexdb-grpc/main.go:119`, déclaré). Dernier passage le 20 octobre 2026 au soir, tête `b6a8899` (porte de relecture sur les voies de `cortex_query`) : `precommit` sans bloc hors `main` et DDL ; gremlins 56 tués, 9 survivants (les mêmes neuf : deux du cache, la capacité de map, la ligne amont voisine, cinq des paquets `main` rejoués ci-dessous), 2 non couverts : `cmd/cortexdb-grpc/main.go:119` (déclaré) et `knowledge_api.go:122` (`case req.Metadata == nil:`, cas vide que le profil de couverture ne suit pas) ; ce dernier rejoué à la main (`==` → `!=`) : `TestTheContentHashFollowsTheContentThroughKnowledgeWrites` tombe. Suite complète `-race` : 27 paquets verts, 7 paquets sans build (disque plein pendant l'édition de liens : `examples/05, 10, 12, 14, 19`, `pkg/connector`, `pkg/graph`), rejoués seuls après nettoyage du disque : `go test -race -count=1` sur ces sept paquets, exit 0 (`examples/19` 18 s, `pkg/connector` 6 s, `pkg/graph` 174 s, les quatre autres exemples sans fichier de test, construits). Éval lexicale verte.

- Un vrai manque : `review.go:125` (`status == ReviewRejected` → `!=`), le relecteur forcé sur une proposition. Cas ajouté à `TestTheReviewerIsTheKeyThatDecided` (document et mémoire), deux mutations ajoutées au jeu « 06 », tuées. Commit amendé en `97e7365`.
- Équivalents : `expiry.go:83` (deux, fraîcheur du cache), `knowledge_tx.go:208` (capacité d'une map), `retrieval_plan.go:236` (ligne amont voisine, champ `Namespace`, hors patch) ; tous déjà classés le 19 octobre.
- Paquets `main` (cinq) : rejoués à la main (`mutate_main_06.py`), cinq tués :

| Mutation | Test qui échoue |
| :- | :- |
| `grpc 17 variable vide lue comme mise` | TestTheReviewGateIsReadFromTheEnvironment, TestTheReviewGateOptionFollowsTheEnvironment |
| `grpc 21 valeur invalide admise` | TestTheReviewGateIsReadFromTheEnvironment, TestTheReviewGateOptionFollowsTheEnvironment |
| `grpc 31 erreur de lecture ignorée` | TestTheReviewGateOptionFollowsTheEnvironment |
| `stdio 58 variable vide lue comme mise` | TestTheReviewGateIsReadFromTheEnvironment |
| `stdio 60 valeur invalide admise` | TestTheReviewGateIsReadFromTheEnvironment |
