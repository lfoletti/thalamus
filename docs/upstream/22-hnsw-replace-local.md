# 22. Réécrire un vecteur HNSW ne lit plus tout l'index

**Statut** : **envoyée en relecture le 15 novembre 2026 (liliang-cn/cortexdb#192, tête 5bd151e, branche `co-25-hnsw-replace-local` du fork lfoletti/cortexdb), base `main` 7e4b22b.** **Taille** : 79 lignes ajoutées et 27 retirées dans `pkg/index/hnsw.go`, un fichier de tests. **Dépend de** : rien ; même fonction que la fiche 21.

## Le défaut

Écrire un vecteur sous un identifiant que l'index a déjà tenu (`Upsert`, ou `Insert` après `Delete`) retirait l'ancien nœud, puis lisait la liste de voisins de chaque nœud, sur chaque couche, pour en ôter les liens. Le coût d'une réécriture croissait avec la taille de l'index.

## Qui l'atteint

Un réencodage du corpus, une réingestion de documents dont les morceaux gardent leurs identifiants : une lecture de tout l'index par morceau.

## Preuves (`BenchmarkHNSWUpsertExisting`, 64 dimensions, M=16)

| nœuds | `main` | branche |
| -: | -: | -: |
| 5 000 | 0,81 ms | 0,29 ms |
| 20 000 | 3,52 ms | 0,41 ms |

Rappel@10 après 8 000 réécritures vers une grappe tirée au sort (4 000 vecteurs, 40 grappes, 8 passages) : 0,60 en moyenne sur `main`, 0,89 sur la branche. Pour un vecteur réécrit dans sa grappe, ni gain ni perte (593 et 594 sur 600).

## Le correctif

- Les listes de l'ancien nœud nomment ses voisins : chacun perd son lien de retour, rien d'autre n'est lu.
- Un lien à sens unique reste et pointe vers le nœud à sa nouvelle place.
- Le nœud réécrit garde le niveau de celui qu'il remplace (décision de Lauro du 15 novembre 2026).
- La descente et le choix des voisins écartent le nœud lui-même.

Tests : `pkg/index/hnsw_replace_test.go` ; trois tombent sur `main`. 23 mutations à la main tuées ; `gremlins` 14 tués, 0 vivant.

## Ce qui change de comportement

Sur `main`, un nœud réécrit retire son niveau au sort, ce qui ramenait à la bonne forme un index bâti sous l'ancienne règle des niveaux. Ici il garde son niveau : un tel index garde sa forme jusqu'à sa reconstruction (simulation : 752 → 752 nœuds au-dessus de la couche 0, contre 779 → 87 sur `main`). Nommé dans le corps de la PR.

La branche laisse plus de liens vers une couche que la cible n'a pas : 27 à 63 sur environ 1 200 après dix tours, contre 4 à 11 sur `main` ; aucune perte de rappel mesurée.

## Ce que Thalamus y gagne

Rien à retirer. Le réencodage d'un cerveau et la réécriture d'un document coûtent le voisinage du vecteur, pas l'index.

Pour européens : sans effet sur les langues.
