# 10. `find_nodes` pour une clé confinée

**Statut** : en réserve, ne pas envoyer sans la demande de Lauro. **Taille** : 2 lignes de code hors tests (essayée le 28 octobre 2026 sur `origin/main` 86131b2) ; tests : 10 lignes ajoutées ou changées, 16 retirées. **Dépend de** : rien. Le filtre que la liste blanche attend est déjà dans v2.139.0 (ef87828).

Chantier D4 du projet d'alignement. La fiche 06 a fait ouvrir les autres outils de graphe (#95 à #97, v2.139.0). `find_nodes` est resté seul de côté.

## Le besoin de Thalamus

kg-mcp disparaît (F1 du projet d'alignement). Son outil `chercher` passe à CortexDB par `find_nodes` : une personne a un nom et veut entrer dans le graphe. Pour un cerveau confiné à une collection, il n'y a pas d'autre porte d'entrée. `expand_graph`, `get_nodes` et `search_paths` sont déjà ouverts, mais ils partent d'un identifiant. `knowledge_memory_neighbors` et `search_paths` acceptent un nom, mais seulement le nom exact d'une entité, et ils ne disent pas comment il a été trouvé.

Sans `find_nodes`, `chercher` passe aujourd'hui par `th_recall` (des passages) et par le document des sommets calculé à la construction. Ce repli marche, mais il perd la recherche sans tenir compte de la casse, la recherche par inclusion et le filtre par type.

## Ce que le moteur a déjà

1. **Le refus et sa condition.** La liste blanche (`pkg/authz/allowlist.go`, v2.139.1) refuse l'outil « matches the content of every node, chunks of every collection included; it opens in a change of its own, once its matches pass the hidden-node filter ».
2. **La condition est remplie depuis v2.139.0.** `ef87828` (« find_nodes leaves out the nodes of hidden documents and memories ») fait passer chaque fenêtre de résultats par `withoutExpiredDocuments` avant la limite (`graphrag_tool_find_nodes.go`). Ce filtre applique aussi le périmètre de l'appelant :
   - `hiddenAmong` → `outOfScopeDocuments` cache les nœuds `chunk` et `document` d'une autre collection, et cache tout si la lecture échoue ;
   - `hiddenMemoryNodes` cache les mémoires hors de la règle de mémoire ;
   - `scopedNode` cache une entité que seule une autre collection connaît et rend les autres réduites (nom et type).
3. **Les tests existent, mais ils sont sautés.** `graph_caller_scope_test.go` et `graph_scoped_entities_test.go` portent une ligne `find_nodes` dans la table des outils de graphe. Ces lignes sont sautées par `const findNodesNeedsTheHiddenFilter = true`, qui annonce le changement de ef87828. Le changement est arrivé, mais la constante est restée.

## Essai du 28 octobre 2026 (copie jetable de `origin/main` 86131b2)

| Étape | Résultat |
| :- | :- |
| La constante seule passe à `false` | les 3 lignes `find_nodes` passent : `TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection`, `TestWithoutACallerScopeTheGraphToolsAnswerAsBefore`, `TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows` |
| Mutation : `kept := ordered` à la place de `withoutExpiredDocuments` | `TestUnderACallerScope…AnotherCollection/find_nodes` tombe : il sert `BETASECRET money`, `chunk:b1:000`, `doc:b1`, `ZEUSPRIVATE`, `memory:zeus-mem` |
| Changement complet (ci-dessous) | `pkg/authz`, `pkg/rpcserver`, `pkg/httpapi`, `pkg/cortexdb` verts ; `go vet` propre |
| Tests seuls, sans le changement de `allowlist.go` | `pkg/authz` tombe (`TestTheGraphWalksAreOpenToAConfinedKeyAndConfinedByTheFacade`, `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys`, `TestRowIDsOfReadsEveryDeclaredPath`) ; `TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC` et `…OverHTTP` tombent sur le refus (`PermissionDenied`, 403) |
| Même mutation du filtre, changement complet en place | `TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC` et `…OverHTTP` tombent : `find_nodes` sert `secretly funds`, `chunk:b1:000`, `doc:b1`, `entity:project_nightjar`, `BETADESC`, `owes me money`, `memory:zeus-mem` ; les bancs « avant / après » (`TestAConfinedKeysGraphWalkDoesNotPassThroughAnotherCollectionsChunk`, `…DoesNotFollowAnotherCollectionsRelation`, gRPC et REST) tombent aussi |

Leçon de l'essai : les bancs « avant / après » seuls ne prouvent pas l'ouverture. Un refus ne fuit rien et ne change pas quand beta écrit, donc ils passent aussi sans le changement de la liste blanche. Le contrôle positif est la ligne de `graphWalks`, qui exige une réponse.

Le patch de l'essai est gardé hors du dépôt, dans `thalamus_local/docs/pr/d4_find_nodes/d4-find-nodes-confined.patch`. Pour le rejouer : `git apply` sur `origin/main` 86131b2, puis `go test ./pkg/authz ./pkg/rpcserver ./pkg/httpapi ./pkg/cortexdb -count=1`.

## La proposition (une PR, une règle)

Règle : `find_nodes` est ouvert à une clé confinée et tenu par la façade (`CallerScope`), comme `get_nodes`.

Code (`pkg/authz/allowlist.go`) :

- `find_nodes` entre dans `confinedTools` avec `stores: storeKnowledge | storeMemory, facade: true, args: names, node_types, limit`. Il n'a aucun champ d'identifiant : `RowIDsOf` ne journalise rien pour lui, comme pour `search_chunks_by_entities`.
- Sa ligne sort de `refusedTools`.
- `CHANGELOG.md` : une entrée sous « Unreleased ».

Tests (écrits avant le code ; avec les tests seuls, `pkg/authz` et `graphWalks` des deux ports tombent sur le refus) :

- `pkg/cortexdb` : la constante `findNodesNeedsTheHiddenFilter` et ses deux sauts sont retirés. Les trois lignes `find_nodes` deviennent actives.
- `pkg/authz/allowlist_graph_test.go` : `find_nodes` entre dans `graphWalkCalls` avec tous ses arguments, et sort de la liste des refus de `TestTheWholeGraphReadsStayRefusedToAConfinedKey`.
- `pkg/authz/allowlist_test.go` : `find_nodes` sort des refus avec une raison propre.
- `pkg/authz/row_ids_test.go` : `"find_nodes": nil` dans la table des chemins journalisés.
- `pkg/rpcserver` et `pkg/httpapi`, `confined_graph_walk_test.go` : une ligne `find_nodes` dans `confinedWalks`, avec `{"names":["Kestrel Vale","Heron Mill","secret","deal"]}`. « secret » et « deal » ne se trouvent par inclusion que dans le chunk de beta. Le banc vérifie que le chunk n'est pas servi et que la réponse d'alpha ne change pas quand beta écrit. Aucun nom envoyé n'est un marqueur, pour qu'un écho de la requête dans `matches[].name` ne fasse pas tomber le test.
- `pkg/rpcserver` et `pkg/httpapi`, `confined_graph_tools_test.go` : `find_nodes` sort de la boucle qui attend `PermissionDenied`, et entre dans `graphWalks` avec `{"names":["Kestrel Vale","Project Nightjar","secretly","owes me"]}`. C'est le contrôle positif : alpha doit recevoir une réponse qui nomme Kestrel Vale, et rien de beta ni de Zeus. « secretly » et « owes me » ne se trouvent par inclusion que dans le chunk de beta et dans la mémoire de Zeus.

Table outil par outil (règle du propriétaire du 20 octobre 2026) :

| Outil | Arguments ouverts | Champ d'identifiant | Tenu par | Tests |
| :- | :- | :- | :- | :- |
| `find_nodes` | `names`, `node_types`, `limit` | aucun | la façade : `withoutExpiredDocuments` (documents par collection, mémoires par bucket, entités par les collections qui les listent) | les 3 lignes de `pkg/cortexdb` ; `graphWalkCalls` ; `graphWalks` et `confinedWalks`, gRPC et REST |

## Points à dire au propriétaire

1. **Coût.** `find_nodes` lit tous les nœuds qui ont un contenu, à chaque appel (« the graph has no index on content »). Une clé confinée paie ce parcours comme une clé ouverte, et le filtre s'ajoute ensuite. À mesurer avant l'envoi : temps d'un appel à 4 noms sur le cerveau ai-audio et sur un cerveau de 10 000 documents, comme pour #76. Si le propriétaire veut une borne, elle tient dans `limit` et dans le nombre de noms par appel, pas dans la liste blanche.
2. **L'ordre des résultats.** Les nœuds cachés sortent fenêtre par fenêtre, avant la limite, et `match` vient du premier nœud servi. La réponse d'alpha ne dépend donc pas de ce que beta écrit : le banc « before / after » le vérifie.
3. **`node_types`** passe par `expandOntologyTypeFilter`. L'ontologie est commune au serveur, mais l'argument ne rend que des nœuds, et ces nœuds sont filtrés.

## Ce que Thalamus fait après la fusion

- K3 du projet d'alignement : `chercher` relaie `find_nodes` avec la clé de la personne, sur la liste fermée des outils de graphe.
- Le repli du 15 novembre 2026 (binaire compilé depuis le fork) n'a plus d'objet. Ce repli est inscrit dans `AMONT.md` du projet d'alignement.
