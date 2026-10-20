# Failing first, gateway (tranches 1 et 2, 20 octobre 2026)

Preuve : `gateway/mutate.py`. Le script applique une mutation à la main par règle, lance le test nommé, remet le fichier. Sortie attendue : `KILLED` sur chaque ligne, `alive: 0`.

Commande : `cd gateway && python3 -I mutate.py`.

Résultat du 20 octobre 2026 : tranche 1, 29 mutations, 29 tuées ; tranches 1 et 2 ensemble, 48 mutations, 48 tuées.

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
| relay (2) | `gc_brains` dit que chaque cerveau a une clé | `TestGcBrainsListsTheCallersBrains` |
| cmd | secret pas seul sur stdout | `TestKeyIssueAndRevoke` |
| cmd | erreur d'usage en code 1 | `TestTheCommandLineIsChecked` |
| cmd | arrêt propre rendu comme erreur | `TestServeListensAndStops` |

Couverture mesurée (`go test -race -coverprofile`) : `keys`, `relay`, `rights` à 100 %, `cmd/gc-gateway` à 96,7 %, les quatre lignes non couvertes sont `main`.
