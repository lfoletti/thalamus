# 23. Les recherches continuent pendant qu'un document entre dans l'index

**Statut** : **envoyée en relecture le 15 novembre 2026 (liliang-cn/cortexdb#193, tête 0a19eee, branche `co-26-ingest-sync-shared-lock` du fork lfoletti/cortexdb), base `main` 7e4b22b.** **Taille** : 10 lignes ajoutées et 4 retirées dans `pkg/core/store_tx.go` et `pkg/core/index_sync.go`, un fichier de tests. **Dépend de** : rien.

## Le défaut

`SyncUpsertedEmbeddings` met dans les index en mémoire (HNSW, IVF, binaire) les vecteurs d'une transaction commise. Elle prenait le verrou du magasin en exclusif pour toute la boucle : chaque recherche attendait que le dernier vecteur du document soit indexé. `SyncDeletedEmbeddingIDs`, appelée avant elle, attendait ce verrou en exclusif. Or un écrivain en attente sur un `RWMutex` de Go arrête tout nouveau lecteur.

## Qui l'atteint

Toute recherche lancée pendant une écriture de connaissance : plus le document est grand, plus l'arrêt est long.

## Preuves (5 000 vecteurs de 64 dimensions, une synchronisation de 500)

| | recherches finies pendant la synchronisation | recherche la plus longue |
| :- | -: | -: |
| `main` | 1 | 133 à 137 ms |
| branche | 500 | 1 ms |

## Le correctif

Les deux synchronisations prennent le verrou en partagé, comme `UpsertBatch` et `Delete` le font pour les mêmes mises à jour ; chaque index a son propre verrou.

Tests : `pkg/core/store_tx_sync_test.go`, cinq tests ; trois tombent sur `main`. 6 mutations à la main tuées ; aucun mutant `gremlins` dans le diff.

## Laissé de côté

- Les suppressions du magasin (`DeleteBatch`, `DeleteByDocID`, `DeleteDocument`, écritures multi-vecteurs) gardent le verrou exclusif : l'une d'elles en attente arrête encore les recherches. Une autre PR.
- Deux synchronisations peuvent s'entrelacer ; une recherche peut voir un document à moitié indexé.
- Avec l'index binaire, le recentrage tient le verrou de l'index pendant deux lectures de la table.

## Ce que Thalamus y gagne

Rien à retirer. Un rappel lancé pendant qu'un curateur accepte un document ne s'arrête plus le temps de l'indexation.

Pour européens : sans effet sur les langues.
