# 11. `find_nodes` normalise chaque nœud une fois par appel

**Statut** : en réserve, envoyée avec la fiche 10 sur demande de Lauro (28 octobre 2026). **Taille** : 4 lignes de code hors tests, un test de 30 lignes (essayée le 28 octobre 2026 sur `origin/main` 86131b2). **Dépend de** : rien. Elle est indépendante de la fiche 10 : les deux PR partent de `main` et ne touchent pas les mêmes lignes.

## Le constat

La mesure de la fiche 10 montre que le coût de `find_nodes` croît avec le nombre de noms demandés. La cause est dans la boucle de rapprochement (`graphrag_tool_find_nodes.go`). Pour chaque nom et pour chaque nœud, `foldName(c.content)` est appelé, et même deux fois quand le nœud n'est pas égal au nom : une fois pour la comparaison normalisée, une fois pour l'inclusion. `foldName` construit une chaîne neuve à chaque appel. Un appel à k noms sur n nœuds normalise donc jusqu'à 2·k·n fois le même contenu. Or la description de l'outil demande de grouper les noms (« Batch them: one scan serves all »).

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

Le patch est gardé hors du dépôt, dans `thalamus_local/docs/pr/d4_find_nodes/d5-find-nodes-fold-once.patch`. Pour le rejouer : `git apply` sur `origin/main` 86131b2, puis `go test ./pkg/cortexdb -run TestFindNodes -count=1`. Retirer le changement de `graphrag_tool_find_nodes.go` fait tomber le test.

## Ce que Thalamus y gagne

Le relais de `chercher` (K3) envoie les noms par lots. Avec ce changement, un lot de 8 noms coûte environ 160 ms au lieu de 800 ms sur un cerveau de la taille d'ai-audio.
