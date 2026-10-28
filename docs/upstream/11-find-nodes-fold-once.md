# 11. `find_nodes` normalise chaque nœud une fois par appel

**Statut** : envoyée avec la fiche 10 le 28 octobre 2026 sur demande de Lauro, branche `co-11-find-nodes-fold-once` du fork. **Taille** : 6 lignes de code (dont un commentaire de deux lignes), 3 retirées, et une ligne de CI ; un test de 33 lignes (base `origin/main` 86131b2). **Dépend de** : rien. Elle est indépendante de la fiche 10 : les deux PR partent de `main` et ne touchent pas les mêmes lignes.

## Le besoin de Thalamus

Le relais de `chercher` envoie les noms par lots, comme la description de l'outil le demande (« Batch them: one scan serves all »). La mesure de la fiche 10 montre que le temps d'un appel suit le nombre de noms : la boucle de rapprochement (`graphrag_tool_find_nodes.go`) appelle `foldName(c.content)` pour chaque nom et chaque nœud, une fois pour la comparaison normalisée et une fois pour l'inclusion. La lecture des nœuds, elle, se fait une fois par appel. Cette fiche propose de faire de même pour la normalisation.

## La proposition (une PR, une règle)

Règle : le contenu de chaque nœud est normalisé une fois par appel, pendant la lecture, et pas une fois par nom.

Le code ajoute un champ `folded` à `candidate`, rempli par `foldName(c.content)` à la lecture. Les deux comparaisons lisent ce champ. Le comportement ne change pas : mêmes correspondances, même ordre, même force.

## Preuves (essai du 28 octobre 2026)

| Contrôle | Résultat |
| :- | :- |
| Test `TestFindNodesFoldsEachNodeOnceACall` (500 entités, 1 puis 9 noms sans correspondance, `testing.AllocsPerRun`, au plus 50 allocations par nom ajouté) | sur `main` : **tombe**, 2 026 allocations par nom (6 090 pour un nom, 22 299 pour neuf) ; avec le changement : passe, 4 allocations par nom (5 079, 5 112) |
| Mutation : l'inclusion normalise encore à chaque nom (`strings.Contains(foldName(c.content), folded)`) | tombe, 1 015 allocations par nom |
| Suite de `pkg/cortexdb` et `go vet` | verts |
| Réponses identiques : 306 noms sur une copie d'ai-audio (150 entités au hasard, leurs variantes en majuscules, sans espaces ni tirets, des fragments, du CJK, de la ponctuation seule), par lots de 8, `limit` 20 | même empreinte SHA-256 du JSON sur `main` et avec le changement |

Le test suit un précédent de l'amont : `TestALongRunOfGroupsIsJoinedOnce` (`check_sources_test.go`) tient déjà un coût par `testing.AllocsPerRun`.

Temps d'un appel, sur une copie d'ai-audio (30 161 nœuds avec un contenu, Apple M2, médiane de 10 appels après une chauffe, sans clé et clé confinée à 2 ms près) :

| Noms par appel | `main` | avec le changement | gain |
| -: | -: | -: | -: |
| 1 | 119 ms | 78 ms | × 1,5 |
| 4 | 387 ms | 88 ms | × 4,4 |
| 8 | 798 ms | 156 ms | × 5,1 |

La mémoire ajoutée est une chaîne normalisée par nœud, le temps d'un appel. Elle est de la taille du contenu déjà gardé dans `all`.

Sous `-race`, que la CI lance, le détecteur ajoute ses propres allocations (420 à 664 par nom) : le test se saute alors (`raceEnabled`) et tourne dans l'étape « race-exempt » de la CI, comme `TestALargeUpsertDoesNotExhaustTheSQLPlaceholders`.

Pour le rejouer : la branche `co-11-find-nodes-fold-once` du fork, puis `go test ./pkg/cortexdb -run TestFindNodes -count=1`. Retirer le changement de `graphrag_tool_find_nodes.go` fait tomber le test.

## Ce que Thalamus y gagne

Avec ce changement, un lot de 8 noms coûte environ 160 ms au lieu de 800 ms sur un cerveau de la taille d'ai-audio.
