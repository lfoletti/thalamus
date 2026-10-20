# Failing first, gateway (tranches 1 à 8, 20 octobre 2026)

Preuve : `gateway/mutate.py`. Le script applique une mutation à la main par règle, lance le test nommé, remet le fichier. Sortie attendue : `KILLED` sur chaque ligne, `alive: 0`.

Commande : `cd gateway && python3 -I mutate.py`.

Résultat du 20 octobre 2026 : tranche 1, 29 mutations, 29 tuées ; tranches 1 et 2 ensemble, 48 mutations, 48 tuées ; tranches 1 à 3, 58 mutations, 58 tuées (une mutation de la tranche 3 a survécu au premier passage, « prompt court demandé », jusqu'à ce que le test compte les appels au gateway) ; tranches 1 à 4, 71 mutations, 71 tuées ; tranches 1 à 5, 91 mutations, 91 tuées ; tranches 1 à 6, 98 mutations, 98 tuées ; tranches 1 à 7, 108 mutations, 108 tuées ; tranches 1 à 8, 113 mutations, 113 tuées. Un défaut trouvé par l'essai réel du plugin (tranche 8, `tests/e2e_plugin.sh`) et non par le banc : les extraits du hook sont coupés par le moteur (`compactSnippet`), et sans outil de lecture entière la session rappelait `th_recall` avec la même question jusqu'à `error_max_turns`. Correctif : `th_get`, nommé dans le pied du bloc du hook et dans la compétence ; la session suivante a répondu en un `th_get`. Un fait appris par le banc, tranche 5 : `knowledge_search` avec un filtre exige un texte de requête (« empty text provided »), donc `th_pending` liste par `aggregate_metadata` puis `knowledge_get`, et `knowledge_get` refuse l'argument `collection` d'une clé confinée (la garde des lignes suffit). Un vrai défaut attrapé par un test écrit d'abord, tranche 4 : la première version de `key sync` ne parcourait que les utilisateurs nommés par `rights.yaml`, donc ne révoquait jamais la clé d'un porteur retiré du fichier ; `TestKeySyncBringsTheKeysToTheRights` est tombé sur « would revoke th/ghost/team/acme » absent.

| Paquet | Règle mutée | Test qui tombe |
| :- | :- | :- |
| rights | un rédacteur lit seulement ; un lecteur écrit | `TestAWriterMayAlsoRead` |
| rights | cerveau inconnu = pas d'accès (au lieu d'une erreur) | `TestAnUnknownBrainIsAnError` |
| rights | groupe non déclaré admis | `TestAMemberMustBeAGroupOrAUser` |
| rights | champ mal orthographié ignoré | `TestAMisspelledFieldIsRefused` |
| rights | cerveau sans `addr` admis | `TestABrainNeedsItsThreeFields` |
| rights | `BrainsOf` liste tous les cerveaux | `TestBrainsOfListsTheUsersBrainsInOrder` |
| keys | clé émise sans accès ; clé émise deux fois | `TestNoAccessAndNoDuplicateKey` |
| keys | clé sans portée ; lecteur en `read-write` | `TestAnIssuedKeyIsConfinedToTheBrainsCollection` |
| keys | révocation d'une clé absente admise ; révocation qui garde la mauvaise clé | `TestRevokeRemovesOneKey` |
| keys | clé d'opérateur avec un porteur | `TestHolderNamesTheUserBehindASecret` |
| keys | fichier que le serveur refuse écrit quand même | `TestAFileTheServerWouldRefuseIsNotWritten` |
| keys | écriture en place (pas de fichier temporaire) | `TestAFailedWriteLeavesTheFileAsItWas` |
| relay | clé d'un autre cerveau admise ; droit retiré admis ; sans jeton admis | `TestRefusalsHappenBeforeTheBrainIsReached` |
| relay | le cerveau voit une autre clé ; identifiants absents de l'audit ; succès non audité ; arguments perdus | `TestACallIsRelayedWithTheCallersKeyAndAudited` |
| relay | échec non audité ; échec rendu comme succès | `TestAToolFailureIsAudited` |
| relay | fichier d'audit tronqué à l'ouverture | `TestTheAuditFileAppendsAndNamesItselfOnFailure` |
| rights (2) | poids ignoré ; poids nul admis | `TestABrainHasAPositiveWeight` |
| rights (2) | cerveau sans collection pas reconnu comme serveur entier | `TestABrainNeedsItsAddressAndKeyFile` |
| keys (2) | clé de serveur entier confinée à une collection vide | `TestAKeyForAWholeServerIsNotConfined` |
| keys (2) | `SecretOf` rend la clé de n'importe qui | `TestSecretOfFindsTheUsersKey` |
| relay (2) | fusion sans poids ; même texte en deux résultats ; `also_in` vide ; résultats non triés ; `top_k` ignoré | `TestFuseByReciprocalRank` |
| relay (2) | clé inconnue admise sur `/mcp` ; porteur sans droit admis | `TestTheFederatedEndpointAdmitsOnlyGatewayKeys` |
| relay (2) | cerveau non lisible interrogé ; audit sans identifiants ; cerveau sauté non signalé ; question vide envoyée | `TestGcRecallFusesTheBrains` |
| relay (2) | restriction `brains` ignorée ; collection envoyée à un serveur entier | `TestGcRecallCanBeRestrictedToSomeBrains` |
| relay (2) | `th_brains` dit que chaque cerveau a une clé | `TestGcBrainsListsTheCallersBrains` |
| th-hook (3) | un problème bloque le prompt (exit 1) ; prompt court ou commande demandé ; `TH_RECALL_TOPK` invalide admis ; erreur d'outil lue comme résultat | `TestTheHookIsSilentOnEveryProblem` |
| th-hook (3) | `TH_RECALL_TOPK` ignoré ; `also_in` absent ; cerveaux sautés absents | `TestTheHookInjectsTheExtractsWithTheirProvenance` |
| th-hook (3) | la coupe vide tout ; budget ignoré ; en-tête sur rien envoyé | `TestTheContextFitsTheBudget` |
| rights (4) | `Users` oublie les membres directs | `TestUsersListsEveryLoginOnce` |
| keys (4) | autres clés non comptées ; liste non triée par utilisateur | `TestListReadsTheIssuedKeys` |
| cmd (4) | clé sans droit dite `ok` ; mauvaise clearance dite `ok` ; mauvaise collection dite `ok` ; clés d'un autre cerveau listées sous celui-ci | `TestKeyListShowsTheKeysAgainstTheRights` |
| cmd (4) | porteur non nommé par les droits gardé ; clé périmée gardée ; clé manquante non émise ; `--dry-run` qui écrit ; réémission qui garde l'ancienne clé ; nouveau secret non imprimé | `TestKeySyncBringsTheKeysToTheRights` |
| rights (5) | curateur réduit à rédacteur ; curateur sans écriture ; curateurs ni vérifiés ni comptés | `TestACuratorWritesAndReviews` |
| relay (5) | document d'un rédacteur accepté d'office ; document d'un curateur en attente ; auteur absent ; écriture non auditée ; collection non envoyée | `TestGcAddProposesADocumentForReview` |
| relay (5) | lecteur qui ajoute ; document vide ajouté ; échec d'écriture non audité | `TestGcAddIsRefusedWithoutAWriteRight` |
| relay (5) | rédacteur qui relit ; `th_pending` sans filtre ; `top_k` ignoré ; identifiants listés non audités ; extrait coupé au milieu d'un mot ; liste en échec non auditée ; `knowledge_get` avec collection ; décision quelconque admise ; relecteur non nommé | `TestGcPendingAndGcReviewAreForCurators` |
| rights (6) | fichier changé jamais relu ; nouveaux droits parsés mais pas servis ; panne jamais signalée ; panne signalée à chaque lecture ; même texte reparsé | `TestASourceReloadsTheFileWhenItChanges` |
| relay (6) | droits figés au démarrage | `TestTheRelayFollowsTheRightsFile` |
| cmd (6) | `serve` ne signale pas une panne de rechargement | `TestServeListensAndStops` |
| relay (7) | routes `/v1/` non servies ; `POST /v1/brains` liste ; appelant refusé en 200 ; liste REST non auditée | `TestRestListsTheBrains` |
| relay (7) | cerveau inconnu en 502 ; droit absent en 502 ; mauvais argument en 502 ; ajout refusé en 200 ; réponse du moteur non rendue ; `GET /v1/add` ajoute | `TestRestAddsADocument` |
| relay (8) | lecture sans droit ; id vide envoyé ; `th_get` avec collection ; id lu non audité | `TestThGetReadsADocumentWhole` |
| hook (8) | le pied du bloc ne nomme pas `th_get` | `TestTheHookInjectsTheExtractsWithTheirProvenance` |
| cmd | secret pas seul sur stdout | `TestKeyIssueAndRevoke` |
| cmd | erreur d'usage en code 1 | `TestTheCommandLineIsChecked` |
| cmd | arrêt propre rendu comme erreur | `TestServeListensAndStops` |

Couverture mesurée (`go test -race -coverprofile`) : `keys`, `relay`, `rights` à 100 %, `cmd/th-gateway` et `cmd/th-hook` à 100 % hors `main`.
