# 10. `find_nodes` pour une clé confinée

**Statut** : envoyée le 28 octobre 2026 sur demande de Lauro, branche `co-10-find-nodes-confined` du fork. **Taille** : 13 lignes hors tests sur 2 fichiers, 2 retirées ; tests : 93 lignes ajoutées sur 10 fichiers, 19 retirées (base `origin/main` 86131b2). **Dépend de** : rien. Le filtre que la liste blanche attend est dans v2.139.0 (ef87828).

La fiche 06 a proposé les autres outils de graphe pour une clé confinée (#95 à #97, v2.139.0). La liste blanche annonçait `find_nodes` pour un changement à part : cette fiche le propose.

## Le besoin de Thalamus

Le serveur de graphes de contaixt (kg-mcp) passe à Thalamus. Son outil `chercher` se relaie par `find_nodes` : une personne a un nom et veut entrer dans le graphe. C'est l'outil qui entre dans le graphe par un nom, sans tenir compte de la casse ni de la ponctuation, ou par inclusion, et qui dit comment il a trouvé. Les outils de parcours déjà ouverts à une clé confinée (`expand_graph`, `get_nodes`, `search_paths`) partent ensuite de l'identifiant qu'il rend.

Sans `find_nodes`, `chercher` passe aujourd'hui par `th_recall` (des passages) et par le document des sommets calculé à la construction. Avec `find_nodes`, `chercher` garde la recherche sans tenir compte de la casse, la recherche par inclusion et le filtre par type.

## Ce que le moteur fournit

1. **Le refus et sa condition.** La liste blanche (`pkg/authz/allowlist.go`, v2.139.1) refuse l'outil « matches the content of every node, chunks of every collection included; it opens in a change of its own, once its matches pass the hidden-node filter ».
2. **La condition est remplie depuis v2.139.0.** `ef87828` (« find_nodes leaves out the nodes of hidden documents and memories ») fait passer chaque fenêtre de résultats par `withoutExpiredDocuments` avant la limite (`graphrag_tool_find_nodes.go`). Ce filtre applique aussi le périmètre de l'appelant :
   - `hiddenAmong` → `outOfScopeDocuments` cache les nœuds `chunk` et `document` d'une autre collection, et cache tout si la lecture échoue ;
   - `hiddenMemoryNodes` cache les mémoires hors de la règle de mémoire ;
   - `scopedNode` cache une entité que seule une autre collection connaît et rend les autres réduites (nom et type).
3. **Les tests sont prêts.** `graph_caller_scope_test.go` et `graph_scoped_entities_test.go` portent une ligne `find_nodes` dans la table des outils de graphe, mise en attente par `const findNodesNeedsTheHiddenFilter = true` jusqu'au changement qui ouvre l'outil. Cette PR lève l'attente.

## Essai du 28 octobre 2026 (copie jetable de `origin/main` 86131b2)

| Étape | Résultat |
| :- | :- |
| La constante seule passe à `false` | les 3 lignes `find_nodes` passent : `TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection`, `TestWithoutACallerScopeTheGraphToolsAnswerAsBefore`, `TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows` |
| Mutation : `kept := ordered` à la place de `withoutExpiredDocuments` | `TestUnderACallerScope…AnotherCollection/find_nodes` tombe : il sert `BETASECRET money`, `chunk:b1:000`, `doc:b1`, `ZEUSPRIVATE`, `memory:zeus-mem` |
| Changement complet (ci-dessous) | `pkg/authz`, `pkg/rpcserver`, `pkg/httpapi`, `pkg/cortexdb` verts ; `go vet` propre |
| Tests seuls, sans le changement de `allowlist.go` | `pkg/authz` tombe (`TestTheGraphWalksAreOpenToAConfinedKeyAndConfinedByTheFacade`, `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys`, `TestRowIDsOfReadsEveryDeclaredPath`) ; `TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC` et `…OverHTTP` tombent sur le refus (`PermissionDenied`, 403) |
| Même mutation du filtre, changement complet en place | `TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC` et `…OverHTTP` tombent : `find_nodes` sert `secretly funds`, `chunk:b1:000`, `doc:b1`, `entity:project_nightjar`, `BETADESC`, `owes me money`, `memory:zeus-mem` ; les bancs « avant / après » (`TestAConfinedKeysGraphWalkDoesNotPassThroughAnotherCollectionsChunk`, `…DoesNotFollowAnotherCollectionsRelation`, gRPC et REST) tombent aussi |

Leçon de l'essai : les bancs « avant / après » seuls ne prouvent pas l'ouverture. Un refus ne fuit rien et ne change pas quand beta écrit, donc ils passent aussi sans le changement de la liste blanche. Le contrôle positif est la ligne de `graphWalks`, qui exige une réponse.

Pour le rejouer : la branche `co-10-find-nodes-confined` du fork, puis `go test ./pkg/authz ./pkg/rpcserver ./pkg/httpapi ./pkg/cortexdb -count=1`.

## La proposition (une PR, une règle)

Règle : `find_nodes` est ouvert à une clé confinée et tenu par la façade (`CallerScope`), comme `get_nodes`.

Code (`pkg/authz/allowlist.go`) :

- `find_nodes` entre dans `confinedTools` avec `stores: storeKnowledge | storeMemory, facade: true, args: names, node_types, limit`. Il n'a aucun champ d'identifiant : `RowIDsOf` ne journalise rien pour lui, comme pour `search_chunks_by_entities`.
- Sa ligne sort de `refusedTools`.

Code (`pkg/cortexdb/graphrag_tool_find_nodes.go`), ajouté après la relecture adverse : `node_types` se compare au nœud tel qu'il est servi. Une entité que le périmètre réduit n'a pas de type (`scopedNode`), donc elle ne répond à aucun filtre de type, comme dans les parcours (`ofTypes`). Sans ce point, une clé confinée lirait le type qu'une autre collection a donné à une entité commune. Sous un périmètre de collection, toute entité est réduite : pour une clé confinée, `node_types` qui nomme un type d'entité ne rend aucune entité ; les chunks et les documents gardent leur type. Un nœud écarté par ce filtre n'entame pas la limite. Test : `TestFindNodesMatchesNoTypeOnAnEntityTheScopeReduces`, qui couvre aussi la limite sur deux fenêtres.

Condition de fusion : le corps de la PR demande de fusionner seulement après le changement nommé dans une note privée envoyée au propriétaire le 28 octobre 2026.

Tests (écrits avant le code ; avec les tests seuls, `pkg/authz` et `graphWalks` des deux ports tombent sur le refus) :

- `pkg/cortexdb` : la constante `findNodesNeedsTheHiddenFilter` et ses deux sauts sont retirés. Les trois lignes `find_nodes` deviennent actives.
- `pkg/authz/allowlist_graph_test.go` : `find_nodes` entre dans `graphWalkCalls` avec tous ses arguments, et sort de la liste des refus de `TestTheWholeGraphReadsStayRefusedToAConfinedKey`.
- `pkg/authz/allowlist_test.go` : `find_nodes` sort des refus avec une raison propre.
- `pkg/authz/row_ids_test.go` : `"find_nodes": nil` dans la table des chemins journalisés.
- `pkg/rpcserver` et `pkg/httpapi`, `confined_graph_walk_test.go` : une ligne `find_nodes` dans `confinedWalks`, avec `{"names":["Kestrel Vale","Heron Mill","secret","deal"]}`. « secret » et « deal » ne se trouvent par inclusion que dans le chunk de beta. Le banc vérifie que le chunk n'est pas servi et que la réponse d'alpha ne change pas quand beta écrit. Aucun nom envoyé n'est un marqueur, pour qu'un écho de la requête dans `matches[].name` ne fasse pas tomber le test.
- `pkg/rpcserver` et `pkg/httpapi`, `confined_graph_tools_test.go` : `find_nodes` sort de la boucle qui attend `PermissionDenied`, et entre dans `graphWalks` avec `{"names":["kestrel vale","project nightjar","secretly","owes me"]}`. C'est le contrôle positif : alpha doit recevoir le nœud `entity:kestrel_vale` (son propre chunk porte aussi le nom, et l'écho en minuscules ne compte pas), et rien de beta ni de Zeus. « secretly » et « owes me » ne se trouvent par inclusion que dans le chunk de beta et dans la mémoire de Zeus.

Table outil par outil (règle du propriétaire du 20 octobre 2026) :

| Outil | Arguments ouverts | Champ d'identifiant | Tenu par | Tests |
| :- | :- | :- | :- | :- |
| `find_nodes` | `names`, `node_types`, `limit` | aucun | la façade : `withoutExpiredDocuments` (documents par collection, mémoires par bucket, entités par les collections qui les listent) | les 3 lignes de `pkg/cortexdb` ; `TestFindNodesMatchesNoTypeOnAnEntityTheScopeReduces` ; `graphWalkCalls` ; `graphWalks` et `confinedWalks`, gRPC et REST |

## Points à dire au propriétaire

1. **Coût, mesuré le 28 octobre 2026** (copie du cerveau ai-audio : 30 161 nœuds avec un contenu, dont 15 160 chunks et 5 244 documents ; bibliothèque v2.139.1 ; Apple M2 ; médiane de 10 appels après un appel de chauffe) :

   | Noms par appel | Sans clé | Clé confinée (`graphrag_chunks`) |
   | -: | -: | -: |
   | 1 | | 121 ms |
   | 4 | 388 ms | 390 ms |
   | 8 | | 800 ms |

   Le filtre ajoute environ 2 ms. Le temps se partage entre la lecture des nœuds (environ 25 ms) et le rapprochement de chaque nom (environ 95 ms par nom, `foldName` sur le contenu de chaque nœud). Une clé confinée ne coûte donc pas plus qu'une clé ouverte. La fiche 11 propose, dans une PR à part, de normaliser chaque contenu une fois par appel : 8 noms en 156 ms au lieu de 798 ms.
2. **L'ordre des résultats.** Les nœuds cachés sortent fenêtre par fenêtre, avant la limite, et `match` vient du premier nœud servi. La réponse d'alpha ne dépend donc pas de ce que beta écrit : le banc « before / after » le vérifie.
3. **`node_types`** passe par `expandOntologyTypeFilter`. L'ontologie est commune au serveur, mais l'argument ne rend que des nœuds, et ces nœuds sont filtrés, type compris.
4. **Avec la fiche 11.** Les deux PR ajoutent un test en fin de `graphrag_tool_find_nodes_test.go` : la seconde fusionnée demande un rebasage qui garde les deux tests.

## Ce que Thalamus fait après la fusion

- `chercher` relaie `find_nodes` avec la clé de la personne, sur la liste fermée des outils de graphe que Thalamus relaie.
