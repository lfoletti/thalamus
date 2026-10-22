# Second tour de revue : plan d'amendement des PR #72 à #77

Rédigé le 21 octobre 2026 à partir des revues du propriétaire (21 octobre 2026, 00 h 25 UTC) et de l'analyse `docs/pr/round2/` (14 points, 11 trouvables à l'avance). Les six branches sont déjà rebasées sur `origin/main` à v2.135.0 (`cc98c6c`), sans changement de code, têtes ci-dessous. #78 est fusionnée (`ff6c5fc`). Le développement se fait **d'une seule traite, sans intervention de Lauro** : décisions prises ici, écarts consignés dans `docs/pr/round2/REPORT.md`. **Aucune poussée, aucun commentaire GitHub** : les deux restent sur commande explicite de Lauro.

## 0. État de départ

| PR | Branche | Tête rebasée (21 octobre 2026) | Base | Copie de travail |
| :- | :- | :- | :- | :- |
| #72 | `gc-04a-authz-tool-rules` | `15b1177` | `origin/main` | `…/scratchpad/wt/gc-04a-authz-tool-rules` |
| #73 | `gc-04b-grpc-confined-tools` | `b28407c` | 04a | `…/wt/gc-04b-grpc-confined-tools` |
| #74 | `gc-04c-rest-confined-tools` | `3709c92` | 04a | `…/wt/gc-04c-rest-confined-tools` |
| #75 | `gc-08-content-hash` | `910e1cf` | `origin/main` | `…/wt/gc-08-content-hash` |
| #76 | `gc-05-expires-at` | `8280fda` | 08 | `…/wt/gc-05-expires-at` |
| #77 | `gc-06-review-status` | `30f2380` | 05 | `…/wt/gc-06-review-status` |
| `proto/granular-cortex` | clone `cortexdb/` | `ccff645` (6 commits + classement lot C, changelog) | `origin/main` | `cortexdb/` |

`…` = `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/395d9eca-adec-4b02-abda-51fc00b17277/scratchpad`. Sauvegardes avant rebasage : tags `backup/<branche>-2026-10-21`. Conflit résolu au rebasage de 04c : `pkg/httpapi/scoped_keys_test.go`, l'amont et nous ajoutions chacun des tests en fin de fichier, les deux sont gardés (19 fonctions de test). Après le rebasage, sans changement de code : `go build ./...`, `go vet` et `go test -race` des paquets touchés sous le canari, verts sur 04b et 04c (`pkg/authz`, `pkg/rpcserver`, `pkg/httpapi`) ; lot C sur 06 : `pkg/cortexdb`, `pkg/core`, `pkg/rpcserver` verts, `cmd/cortexdb-mcp-stdio` rouge sur le seul `TestTheSweepCommandReportsAndRemovesExpiredRows`, que le canari fait tomber (point 3 de #76, à corriger ici). `proto` : `go build ./...` et `pkg/authz` (complétude de la liste blanche) verts. Le niveau 3 n'a pas été rejoué : il fait partie de l'amendement.

Ce que v2.135.0 change pour nous : `cortexdb.WithMemoryOwners` n'existe plus, `authz.Key.CallerContext` pose un `cortexdb.CallerScope` (règle mémoire, règle connaissance, id de clé). Le commentaire de `pkg/authz/allowlist.go` ligne ~171 nomme encore `WithMemoryOwners` : à réécrire. `/v1/recall` est corrigé en amont : rien à faire dans #74.

## 1. Décisions de conception (prises ici, à ne pas rouvrir)

**D1. Clés réservées sous le préfixe `_`.** Le moteur possède les clés qu'il écrit lui-même : `_expires_at`, `_review_status`, `_reviewed_by`, `_reviewed_at`, `_content_sha256`, `_chunk_size`, `_chunk_overlap`. Le préfixe suit `importflow`, qui écrit `_table`. Règles : (a) seules les valeurs des champs explicites des requêtes alimentent ces clés ; (b) une clé du client, quel que soit son nom (`expires_at`, `review_status`, `chunk_size`…), est enregistrée telle quelle, jamais validée, jamais lue comme une expiration ni un statut ; (c) une clé à préfixe `_` envoyée par le client dans `metadata` est retirée en silence (elle n'est pas à lui), sauf `_table` et toute clé que le moteur amont écrit déjà ; (d) sur `knowledge_update` et `memory_update`, les clés réservées sont **reportées** quoi que fasse le client avec sa carte (remplacement entier ou non), et changées seulement par les champs explicites. Cette règle remplace la règle de fusion du point 4 du premier tour de #77 (« merge when it names review_status »), devenue inutile. Alternative écartée : des colonnes dédiées (DDL dans les deux backends, hors taille).

**D2. Champs explicites.** `ingest_document`, `knowledge_save`, `knowledge_update` : `expires_at` (déjà sur `ingest_document` dans #76, à étendre), `review_status`, `reviewed_by`, `reviewed_at`. `memory_save`, `memory_update` : `review_status`, `reviewed_by`, `reviewed_at` (l'expiration d'une mémoire reste `ttl_seconds`). Validation (RFC 3339, valeurs du statut) sur les champs explicites seulement. Les quatre champs sont des pointeurs (`*string`) dans les requêtes Go : absent = inchangé, `""` = retiré (plus d'expiration ; pour le statut, vider une ligne relue reste refusé, règle du tour 1). Test : poser puis retirer (R9). `reviewed_by` = clé appelante quand il y en a une (règle du premier tour, inchangée). Le filtre `plan.filters.review_status` reste.

**D3. Le filtre de relecture est une curation, pas un contrôle d'accès.** Une phrase, la même, dans le corps de #77, dans les descriptions d'outil et dans `docs/` : « a hidden row stays reachable by id, any writer can save `accepted`, and `plan.filters.review_status` lists proposed rows to any reader ». Aucune habilitation de relecteur dans le moteur : la couche au-dessus (Thalamus) tient la liste des curateurs.

**D4. Entités et relations : pas de statut de relecture.** Les schémas de #77 retirent la promesse. Les clés de relecture dans `entities[].metadata` et `relations[].metadata` sont des clés du client (D1 b), sans effet. Alternative écartée : valider et filtrer les nœuds (toutes les requêtes de graphe, hors taille).

**D5. Porte sans cache de 15 s.** La porte (`expires_at`, `review_status`) interroge l'index d'expression à chaque recherche, une requête par appel, mémoïsée dans la requête seulement. Les deux index existent déjà (#76, #77). Le point 4 de #76 (second processus qui sert une ligne cachée pendant 15 s) disparaît. Mesurer avant et après sur 10 000 documents dont 100 expirés (`go test -bench`, une ligne dans le corps). Si la mesure montre plus de 2 ms par recherche, garder un cache invalidé à chaque écriture du processus et de 2 s au plus, et le dire dans le corps.

**D6. Un seul libellé pour un identifiant hors portée** (précisée le 21 octobre 2026 après `risques.md`, R1 à R4). Trois couches répondent « not found » chacune à sa façon (façade v2.135.0, garde `authz`, magasin). Règle : la garde rend `ErrWithheld` nu (texte `not found`, sans genre ni identifiant) ; les ports (#73, #74), **pour une clé confinée seulement**, rendent tout « not found » issu d'un outil sous ce même texte, code NOT_FOUND / 404. Les listes : `plan.filters.document_ids` redevient une **donnée** (le filtre de collection, forcé, suffit ; le tour 1 l'avait constaté ; un test tient qu'un document étranger nommé dans le filtre ne rend aucun chunk) ; `memory_get.memory_ids` : la garde retire les identifiants étrangers et les rend au port, qui reconstruit `missing` **dans l'ordre de la requête** ; `memory_save.supersedes` garde le refus (`not found`), comme un identifiant absent y est une erreur du magasin. La façade ne cache rien par identifiant : la garde lit les lignes à travers elle (R2), test `TestTheRowGuardSeesForeignRowsUnderACallerScope`.

**D7. Ré-ingestion.** Même contenu et même découpage : l'enregistrement, le nœud document et les métadonnées des chunks sont mis à jour en place ; entités, arêtes de mention et relations restent. Contenu ou découpage changé : les chunks sont remplacés (nœuds de chunk retirés avec leurs arêtes de mention), les entités et les relations attachées par le client restent, comme sur `main`. « Unchanged » n'est dit que si l'enregistrement, les chunks et les nœuds de chunk sont tous présents ; sinon tout est reconstruit. Le différentiel contre `main` (`make diffmain`) tranche tout cas non prévu ici : le comportement de `main` gagne, sauf ligne nommée dans le corps.

**D8. Dédup.** La recherche du contenu se fait dans la transaction d'écriture, sans plafond de lignes, dans la collection, et passe par la même porte que la recherche : ce qui est caché (expiré, rejeté) ne détient pas le contenu. Dans #75 seul, la porte est vide.

**D9. Métadonnées des entités et des relations pour une clé confinée (#72).** Les clés `chunk_ids`, `support_edge_ids`, `document_id` et toute clé à préfixe `_` sont refusées dans `entities[].metadata` et `relations[].metadata` (liste blanche, même mécanisme que les champs de premier niveau). Le moteur lui-même continue d'appliquer `metadata` aux propriétés : l'écrasement par un client non confiné est hors sujet de #72.

## 2. Points par PR

Chaque point : classe à inventorier, test écrit d'abord (nom provisoire), changement, mutation à la main. Les noms de tests sont des propositions ; le nom final va dans `failing_first.md`.

### #72 `gc-04a-authz-tool-rules` (base `origin/main`)

| # | Point | Classe | Tests (d'abord) | Changement |
| :- | :- | :- | :- | :- |
| 1 | Identifiant étranger et absent distingués (D6) | `ids` de `make inventory` : les 75 propriétés `*_id`, `*_ids`, `supersedes` ; une ligne par propriété dans `failing_first.md` | `TestAForeignIdAndAMissingIdAnswerTheSameOnEveryRowIdField` (table : chaque champ de la liste blanche, étranger puis absent, même erreur ou même `missing`) ; `TestAForeignIdInAListIsDroppedLikeAMissingOne` (`memory_get.memory_ids`, `plan.filters.document_ids`, `supersedes`) | `pkg/authz/rows.go` : un seul `ErrWithheld` renvoyé pour les deux cas, texte unique ; les listes : retirer les identifiants étrangers avant l'appel et laisser l'outil les déclarer `missing`. Garder la sémantique de `rpcserver/ownership.go` (déjà « same answer »). |
| 2 | `metadata.*` écrit les propriétés réservées (D9) | `keys` de l'inventaire + `ids` imbriqués | `TestReservedKeysAreRefusedInEntityAndRelationMetadataForAConfinedKey` | `pkg/authz/allowlist.go` : règle `metadata` « carte sans clé réservée » sur `ingest_document.entities[]`, `relations[]`, `upsert_entities`, `upsert_relations`, `knowledge_save.entities[]`/`relations[]`. |
| 3 | Rebasage v2.135.0 | — | `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys` reste vert | Commentaire `allowlist.go` : `CallerScope`. |
| 4 | Arguments du lot C | — | le test de complétude | Garder la classification (`dedupe`, `expires_at`, `review_status`, `reviewed_by`, `reviewed_at`, `sweep_expired` refusé) **sur `proto` seulement** (commit `ccff645`), pas dans la branche de PR : la branche part de `main`, qui ne connaît pas ces arguments. Le corps le dit (section « Integration with #75–#77 »). |

### #73 `gc-04b-grpc-confined-tools` (sur 04a)

Rien de nouveau. Tests des points 1 et 2 de #72 tenus par gRPC : `TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC` (un identifiant étranger dans `memory_get.memory_ids` → `missing` ; `relations[].metadata.chunk_ids` → refus). Une mutation (court-circuit du passage par `AuthorizeToolArgs`).

### #74 `gc-04c-rest-confined-tools` (sur 04a)

Même chose par REST : `TestTheRoundTwoRulesOfTheAllowlistHoldOverREST`. Rien pour `/v1/recall`.

### #75 `gc-08-content-hash` (base `origin/main`)

| # | Point | Classe | Tests (d'abord) | Changement |
| :- | :- | :- | :- | :- |
| 1 | La ré-ingestion détruit le graphe attaché (D7) | `writes` : chaque chemin qui écrit un document (`IngestDocument`, `SaveKnowledge`, `UpdateKnowledge`, plan de mutation `knowledge_tx.go`) | `TestATitleOnlyReingestKeepsTheGraphTheCallerAttached` (compte des nœuds et arêtes par type avant et après, `search_chunks_by_entities` 1 → 1) ; `TestAContentChangeReplacesTheChunksAndKeepsTheEntities` | `graphrag_tool_ingest.go` : la voie « même contenu, même découpage » met à jour en place (enregistrement, nœud document, métadonnées des chunks) ; la voie « contenu changé » retire les nœuds de chunk seuls. Plus de `replace=true` sur le plan. |
| 2 | « Unchanged » ne répare pas le graphe (D7) | idem | `TestUnchangedIsSaidOnlyWhenTheChunkNodesAreThere` (après `delete_document_graph`, même texte → nœuds reconstruits, réponse non `unchanged`) | `sameIngest` : vérifier aussi les nœuds de chunk. |
| 3a | Dédup compte les expirés et rejetés (D8) | — dans #75 seul | `TestDedupeLooksUpThroughTheSearchGate` (porte vide ici ; #76 et #77 branchent la leur et ajoutent leur cas) | `documentWithContentHash` prend la porte (`allowDocument` ou un prédicat). |
| 3b | Deux ingestions concurrentes stockent deux fois (D8) | — | `TestTwoConcurrentDedupeIngestsStoreOneCopy` (deux goroutines, même contenu, une `duplicate_of`) | La recherche entre dans la transaction d'écriture (`knowledge_tx.go`), verrou d'écriture SQLite ; PostgreSQL : `SELECT … FOR UPDATE` ou contrainte, à dire dans le corps comme non exécuté ici. |
| 3c | `MaxRows: 16` | — | `TestDedupeFindsTheCopyPastSixteenHolders` (17 documents de même hash dans une autre collection, le 18e dans la bonne) | Plus de plafond, ou requête filtrée sur la collection avant le plafond. |
| 4 | Marques d'ingestion écrasent les clés du client (D1) | `keys` | `TestTheIngestMarksNeverOverwriteACallersKey` (`metadata: {chunk_size: "mine", content_sha256: "mine"}` conservées, marques sous `_…`) | Renommage des trois constantes (`_content_sha256`, `_chunk_size`, `_chunk_overlap`), retrait des clés `_…` du client. |

Scénarios `make diffmain` à écrire en plus de `reingest_scenario_test.go` : contenu changé (comptes), dédup avec 17 détenteurs, deux ingestions concurrentes (sérialisées dans le journal par leur résultat, pas par leur ordre).

### #76 `gc-05-expires-at` (sur 08)

| # | Point | Classe | Tests (d'abord) | Changement |
| :- | :- | :- | :- | :- |
| 1 | `expires_at` en métadonnée ordinaire (D1, D2) | `keys`, `writes` | `TestAMetadataKeyNamedExpiresAtIsTheCallersAndNeverAnExpiry` (valeur `2030-01-01 00:00:00` acceptée, stockée telle quelle, document servi, balayage ne le touche pas) ; `TestARowWrittenByMainWithThatKeyIsStillUpdatable` (ligne posée par SQL direct, `knowledge_update` du titre passe) ; `TestImportflowColumnsAreNotValidated` (un `_table` et une colonne `expires_at` DATETIME passent) | `expiry.go` : `expiresAtKey = "_expires_at"`, lu et écrit par le moteur seul ; champ explicite `ExpiresAt` sur `knowledge_save` et `knowledge_update` en plus de `ingest_document` ; report sur `update` (D1 d). |
| 2 | Quatre chemins non filtrés | `tools` de l'inventaire : les 97 outils, une ligne chacun (« filtré par <test> », « écriture », « hors documents »). Un filtre explicite sur le statut (`filters.review_status`, `plan.filters.review_status`) l'emporte sur la porte, sur `aggregate_metadata` comme ailleurs : `th_pending` en dépend (R8). | `TestSearchVectorRangeLeavesExpiredDocumentsOut`, `TestNeighborsLeaveExpiredNodesOut`, `TestAggregateMetadataCountsVisibleRowsOnly`, `TestRepresentativeRecordsNeverPickAHiddenChunk` | `vector_range.go`, `brain.go` (neighbors), `aggregate.go` : passage par `allowDocument` (ou le filtre de `RetrievalFilters`). |
| 3 | Test qui atteint un serveur réel | — | Le canari du `Makefile` fait tomber la version actuelle : garder cette preuve dans `failing_first.md` (« sans `t.Setenv("CORTEXDB_REMOTE", "")`, le test tombe sous `make precommit` : connection refused ») | `sweep_expired_test.go` : `t.Setenv` de `CORTEXDB_REMOTE`, `CORTEXDB_GRPC_TOKEN`, `CORTEXDB_PATH`, `HOME` dans chaque test, sur le modèle de `doctor_test.go`. |
| 4a | Suppression sans revérification | `writes` | `TestTheSweepSkipsADocumentRenewedAfterTheListing` (hook entre la liste et la suppression, ou suppression conditionnelle vérifiée par SQL) | `SweepExpired` : `DELETE … WHERE id = ? AND <prédicat d'expiration>` ou relecture dans la transaction. |
| 4b | Cache par processus (D5) | — | `TestASecondProcessSeesAnExpiryAtOnce` (deux `DB` ouvertes sur le même fichier) ; `BenchmarkDocumentGate` | `expiry.go` : plus de `documentExpiryCacheTTL`, requête indexée par appel. |

Classification lot C dans #72 : la branche #76 n'ajoute rien à la liste blanche (elle part de `main`), voir #72 point 4.

### #77 `gc-06-review-status` (sur 05)

| # | Point | Classe | Tests (d'abord) | Changement |
| :- | :- | :- | :- | :- |
| 1 | `review_status` en métadonnée ordinaire (D1, D2) | `keys`, `writes` | `TestAMetadataKeyNamedReviewStatusIsTheCallers` (`approved` accepté tel quel, sans effet sur la porte) ; `TestARowWrittenByMainWithThatKeyIsStillUpdatable` | `review.go` : clés `_review_*` ; champs explicites `ReviewStatus`, `ReviewedBy`, `ReviewedAt` sur les six requêtes ; report sur `update` (D1 d) ; la règle de fusion de premier tour est retirée avec son test (remplacée par D1 d, test `TestAReviewDecisionKeepsEveryOtherKey`). Schémas : les descriptions nomment les champs explicites, plus les clés de métadonnées. |
| 2 | `memory_save` sur un identifiant existant efface le statut | `writes` : chaque chemin d'écriture d'une mémoire (`SaveMemory`, `UpdateMemory`, `memory_save` par le toolbox, `knowledge_memory_remember`, import, `supersedes`) | `TestMemorySaveOverAnExistingIdCarriesTheReviewOver` | `memory_api.go` ligne ~126 : charger la ligne existante, passer `before`. |
| 3 | Quatre chemins non filtrés | les mêmes tests que #76 point 2, avec la porte de relecture (`TestTheReviewGateHoldsOnEveryReadPath`, table des quatre) | même changement, porte commune. |
| 4 | Entités et relations (D4) | `promises` de l'inventaire | `TestTheSchemasPromiseNothingAboutEntitiesAndRelations` (les descriptions des outils d'entités et de relations ne contiennent pas « review ») ; `TestAReviewKeyOnAnEntityIsTheCallersData` | `knowledge_memory_tooldefs.go`, `graphrag_tool_defs.go` : textes. |
| 5 | Curation, pas contrôle d'accès (D3) | — | `TestTheToolDescriptionsSayTheGateIsACurationFilter` (la phrase dans chaque description qui mentionne la porte) | Corps, descriptions, `docs/`. |

## 3. Ordre de travail

Les tests fixent l'horloge (champ `now` sur `DB` ou équivalent), jamais `time.Sleep` (R12). `REPORT.md` est tenu après chaque point (R13).

1. Lire : ce plan, `docs/pipeline_pr.md` (sections 1 à 4, dont « Niveau 2 : quatre contrôles par classe »), `docs/pr/lot_B/README.md`, `docs/pr/lot_C/README.md`, les deux `failing_first.md`, les revues (section 4 ci-dessous), le code de chaque point.
2. Niveau 2 d'abord, par PR : `make inventory REPO=<copie> BASE=<base> > docs/pr/round2/inventory_<pr>.md`, puis la table « élément → test ou raison » dans `docs/pr/round2/failing_first.md`. Écrire les scénarios différentiels manquants dans `docs/pr/lot_C/scenarios/` (lot C) et `docs/pr/round2/scenarios_lot_B/` (lot B : un identifiant étranger sur chaque champ, par gRPC et REST, journal des réponses).
3. Développer dans l'ordre de la pile, un `commit --amend` par branche (un commit par PR, auteur Lauro Foletti, sans `Co-Authored-By`), puis rebaser la branche du dessus : 04a → 04b, 04c ; 08 → 05 → 06. Pour chaque point : test d'abord (il tombe), code, `make precommit`, mutation à la main consignée.
4. `make diffmain` sur `gc-06-review-status` contre `origin/main` et sur `gc-08-content-hash` contre `origin/main` : chaque ligne du diff est nommée dans le corps de la PR concernée, ou corrigée.
5. Niveau 3, têtes gelées : `make prepr` sur 04b, 04c (04a comprise), 08 ; sur 05 et 06, `DECLARED=pkg/core/store_postgres.go make prepr …`. Survivants de `gremlins` lus et classés.
6. Reconstruire `proto/granular-cortex` (puis `go test -race -count=1 -timeout 25m ./...` sous le canari sur `proto`, R6, et `git grep` des sept clés réservées hors `pkg/cortexdb`, chaque occurrence classée dans `REPORT.md`, R14) : `origin/main` + les six commits amendés + le commit de classement lot C (réécrit si les arguments changent : `reviewed_by`, `reviewed_at`, `expires_at` sur `knowledge_save` et `knowledge_update`) + entrées `Unreleased` du changelog. Puis `sh tests/e2e_shared_brain.sh` (30 contrôles) depuis la racine, et `sh tests/e2e_gateway.sh` (55) : le gateway Thalamus lit les clés de métadonnées (`th_pending`, `th_review`, `th_add` avec `review_status: proposed`) ; **adapter `gateway/` aux champs explicites et aux clés `_review_*`** (paquet `relay`, tests `go test -race ./...` dans `gateway/`, mutations dans `gateway/mutate.py`, `docs/gateway/failing_first.md`). Mettre à jour la tranche concernée dans `CLAUDE.md` et `gateway/README.md`.
7. Corps des PR : `docs/pr/round2/GC-*.body.md`, copie du corps actuel (`gh pr view <n> --json body -q .body`) avec une section « Changes since review 2 » en tête, la table des outils filtrés (#76), les tailles, les sections `## Codecov` si des lignes `main` restent. Réponses : `docs/pr/round2/review_replies.md`, crochets remplis. Commentaire de lot sous #66 : déjà rédigé dans le même fichier.
8. `docs/pr/round2/REPORT.md` : état final (têtes, tailles hors tests, mutations par PR, résultats `prepr`, `diffmain`, bancs, écarts au plan et pourquoi, ce qui reste à Lauro : la commande de poussée, dans l'ordre du `README.md` des lots). Mettre à jour `CLAUDE.md` (section « État de la branche »), `docs/pr/lot_B/README.md` et `docs/pr/lot_C/README.md`.

## 4. Revues du propriétaire (texte intégral, 21 octobre 2026, 00 h 25 UTC)

Commande pour les relire : `gh api repos/liliang-cn/cortexdb/pulls/<n>/reviews --jq '.[] | select(.submitted_at > "2026-10-20T12:00:00Z") | .body'`.

### #72

1. **Foreign and missing ids are still told apart.** `rows.go` answers a foreign id with `memory "X" not found`, while a missing one comes back from the store as `embedding not found`. `memory_get` with `memory_ids`, and `knowledge_search` with `plan.filters.document_ids`, are sharper still: a foreign id is an error, a missing id is a normal reply. That holds for `memory_get`, `memory_delete`, `memory_save.supersedes`, `knowledge_get` and `knowledge_delete`, and it is the existence oracle `rpcserver/ownership.go` exists to prevent. Please answer both the same way: one wording, and a foreign id in a list goes to `missing` or an empty result rather than an error.
2. **`relations[].metadata.*` writes the provenance fields the allowlist refuses.** Its keys overwrite the edge's properties (`graphrag_tool_ingest.go`, `knowledge_tx.go`), so a confined key can set `chunk_ids`, `support_edge_ids` and `document_id` on a shared edge by putting them in `metadata`. `entities[].metadata.*` can likewise rewrite properties of entities another tenant created. Strip the reserved keys from metadata, or refuse them.

Rebase v2.135.0 : `cortexdb.WithMemoryOwners` is gone. `authz.Key.CallerContext` now hands the facade a `cortexdb.CallerScope`. Several of your `ConfineToolArgs` gates are now also enforced underneath, which is fine; the comment in `allowlist.go` that names `WithMemoryOwners` needs updating. `/v1/recall` over REST crossed stores on main as well; that is fixed there now, so #74 doesn't need to handle it. Integration with #75–#77: whichever stack lands second has to classify their new arguments. The completeness test catches it, as it should.

### #73, #74

The wiring held on every probe. Waits for the two fixes on #72 and a rebase on v2.135.0.

### #75

1. **Re-ingest destroys the graph the caller attached.** The single-transaction path uses `replace=true` (`graphrag_tool_ingest.go`), which retires the chunk nodes with their mention edges, relation edges that carry `document_id`, and entities left orphaned. After a title-only re-ingest, entities went from 4 to 2, relations from 1 to 0, and `search_chunks_by_entities` from 1 hit to 0; `main` keeps them all. When the content and chunking are unchanged, update the record, the node and the chunk metadata in place.
2. **"Unchanged" never repairs the graph.** After `delete_document_graph`, re-ingesting the same text leaves 0 of 2 nodes; `main` rebuilds them.
3. **Dedupe:** it counts expired and rejected documents as holders of the content (a dedupe against an expired document stores nothing, and after the sweep the content is gone). Two concurrent dedupe ingests both store, because the check runs outside the transaction. The lookup stops at `MaxRows: 16`, so it can miss the copy in this collection.
4. Minor: caller metadata keys `chunk_size`, `chunk_overlap` and `content_sha256` are overwritten.

### #76

1. **`expires_at` in ordinary metadata is taken as the expiry.** importflow copies table columns into metadata by name. An `expires_at` DATETIME column (`2030-01-01 00:00:00`) fails the whole import. A past RFC 3339 value silently hides the imported row, and the sweep then deletes it. A database written by `main` that already has such a key can't be updated at all, not even with a title-only `knowledge_update`. Please honour only the explicit `ExpiresAt` field, or a namespaced key, and don't validate keys the caller owns.
2. **Not gated yet:** `search_vector_range` serves expired documents when an embedder is set, `knowledge_memory_neighbors` returns their nodes, and `aggregate_metadata` and `representative_records` count hidden rows (the representative can be a hidden chunk).
3. **The test reached a live server.** `sweep_expired_test.go` clears `CORTEXDB_PATH` but not `CORTEXDB_REMOTE`. On a machine with a shared brain configured, the test sent a dry-run `sweep_expired` to it with the configured token; it was refused only because that server predates the tool. Once the tool is deployed, the same test would sweep the shared brain for real. Please `t.Setenv("CORTEXDB_REMOTE", "")` and the token, and keep tests from reading the caller's environment for a server.
4. Smaller: the sweep lists the expired ids, then deletes each without checking again, so a document renewed in between is still deleted. The gate cache is per process, so a second process on the same file or PostgreSQL serves a newly hidden row for up to 15 s.

### #77

1. **`review_status` in ordinary metadata.** Same as `expires_at` on #76: an imported `review_status=approved` column fails the import, and an existing row with such a key can no longer be updated.
2. **`memory_save` over an existing `memory_id` clears a proposed status.** It passes `before=nil` (`memory_api.go`), so a re-save makes a proposed memory visible to recall.
3. **Read paths not gated yet:** the same as #76 (`search_vector_range`, `knowledge_memory_neighbors`, `aggregate_metadata`, `representative_records`).
4. **Entities and relations.** The schemas (`knowledge_memory_tooldefs.go`) say the review keys hide entities and relations. They don't: recall served a proposed relation, and `review_status=bogus` is accepted on them unvalidated.

For the record, since it's by design: hidden rows stay reachable by id (`knowledge_get`, `memory_get`, `get_nodes`, Cypher, `changes_since`), any read-write key can save a row as `accepted`, and `plan.filters.review_status` lets any reader list proposed rows. That makes the gate a curation filter rather than access control, which is fine as long as the docs say so. If "a person must look first" is meant to be enforced, it needs a reviewer clearance; that would be its own PR.
