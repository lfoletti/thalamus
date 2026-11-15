# 21. Un index HNSW vidé puis réécrit garde ses nœuds

**Statut** : **envoyée en relecture le 15 novembre 2026 (liliang-cn/cortexdb#191, tête dbeb3c9, branche `co-28-hnsw-entry-point-tombstone` du fork lfoletti/cortexdb), base `main` 7e4b22b (v2.152.0 + 2).** **Taille** : 16 lignes dans `pkg/index/hnsw.go`, 206 lignes de tests. **Dépend de** : rien ; touche `liveNodeExistsLocked` comme la fiche 22, la seconde fusionnée se rebase (un conflit, deux fonctions ajoutées au même endroit, les deux gardées). Défaut trouvé par les scénarios différentiels de la fiche 22.

## Le défaut

Quand tous les nœuds d'un index HNSW sont supprimés, le point d'entrée reste sur la dernière pierre tombale. Réécrire cet identifiant faisait vider le point d'entrée par `liveNodeExistsLocked` : l'insertion prenait le nœud pour le premier d'un index vide, sans lien. Les nœuds écrits entre-temps restaient hors d'atteinte, sans erreur.

## Qui l'atteint

Tout magasin dont une collection est vidée puis réécrite sous les mêmes identifiants : une réingestion complète, un réencodage qui supprime puis réinsère.

## Preuves (15 novembre 2026, `main` 7e4b22b)

20 vecteurs écrits, supprimés, réécrits, 400 index par ordre ; nombre d'index où une recherche atteint les 20 nœuds :

| | même ordre | ordre mélangé |
| :- | -: | -: |
| `main` | 246 (154 n'atteignent qu'un nœud) | 176 (217 partiels, 7 à un nœud) |
| branche | 400 | 400 |

## Le correctif

Quand le nœud retiré est le point d'entrée, celui-ci passe au nœud vivant de la couche la plus haute (`highestLiveLocked`) ; sans nœud vivant, le nœud écrit est le premier de l'index, comme avant.

Tests : `pkg/index/hnsw_entry_test.go`, cinq tests, tombent sur `main`. 8 mutations à la main tuées ; `gremlins` 2 tués, 1 vivant équivalent (`>` en `>=`).

## Laissé de côté

- Un instantané déjà abîmé n'est pas réparé à l'ouverture ; il l'est une fois l'index reconstruit depuis la table.
- `Delete` laisse le point d'entrée sur la dernière pierre tombale ; `markDeletedLocked` prend le premier nœud vivant rencontré.

## Ce que Thalamus y gagne

Rien à retirer. Un cerveau dont une collection est vidée puis réécrite ne perd plus de passages à la recherche vectorielle.

Pour européens : sans effet sur les langues.
