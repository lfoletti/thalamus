# 24. Le repli d'une recherche filtrée ne garde que les meilleurs résultats

**Statut** : **envoyée en relecture le 15 novembre 2026 (liliang-cn/cortexdb#194, tête 7bf952f, branche `co-27-filtered-search-fallback` du fork lfoletti/cortexdb), base `main` 7e4b22b.** **Taille** : 17 lignes ajoutées et 11 retirées dans `pkg/core/store_search.go`, un fichier de tests. **Dépend de** : rien.

## Le défaut

Une recherche indexée demande à l'index les plus proches de tout le magasin. Quand trop peu passent le filtre de collection ou de métadonnées, elle se replie sur un balayage des lignes qui le passent (`searchLinear`). Ce balayage chargeait d'abord toutes ces lignes, vecteurs, contenu et métadonnées, puis triait : une recherche limitée à une grande collection tenait la collection en mémoire.

Second défaut, dans `searchLinearTopK` : la place du tas était réservée selon `TopK` avant la première ligne. Un `TopK` de cinq millions réservait 648 Mio pour soixante lignes. Le plus grand entier faisait paniquer (`makeslice: cap out of range`), sur `main` pour un magasin sans index.

## Qui l'atteint

Toute recherche filtrée sur une collection ou une clé de métadonnées peu représentée parmi les plus proches voisins.

## Preuves

Mémoire tenue pendant une recherche (`TopK` 5, 400 lignes de 128 Kio, collection de 50 Mio) : 51 Mio sur `main`, 2 à 4 Mio sur la branche.

## Le correctif

`searchLinear` délègue à `searchLinearTopK`, qui garde les meilleurs `TopK` au fil de la lecture. Le tas démarre à 64 lignes et grandit avec ce qu'il garde.

Tests : `pkg/core/filtered_fallback_test.go`, trois tests ; deux tombent sur `main`. 6 mutations à la main tuées ; aucun mutant `gremlins` dans le diff.

## Ce qui change

- Un résultat vide est une tranche vide, plus `nil`.
- L'ordre de lignes de même score est celui du tas, plus celui de la table.

## Laissé de côté

`SearchWithFilter` sur un magasin sans index, les recherches avancée, à facettes et en flux chargent encore les lignes qu'elles sélectionnent.

## Ce que Thalamus y gagne

Rien à retirer. Un rappel confiné à une collection ne charge plus cette collection quand l'index rend trop peu de ses lignes.

Pour européens : sans effet sur les langues.
