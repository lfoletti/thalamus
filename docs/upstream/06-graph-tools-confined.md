# 06. Les outils de graphe pour une clé confinée

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : deux PR, environ 250 et 150 lignes hors tests. **Dépend de** : #76 (`expires_at`) et #77 (`review_status`) fusionnées, parce que la proposition étend leur filtre des parcours.

## Le besoin de Thalamus

Depuis v2.136.0 (#72 à #74), une clé confinée à une collection appelle la toolbox, mais la liste blanche (`pkg/authz/allowlist.go`) lui refuse tout le graphe :

| Outil | Raison donnée par la liste |
| :- | :- |
| `search_chunks_by_entities` | « searches chunks of every collection » |
| `search_paths` | « walks the shared graph and may return chunk text of any collection » |
| `knowledge_memory_expand_entity_context` | « returns chunk text of any collection » |
| `expand_graph`, `get_nodes`, `find_nodes`, `knowledge_memory_neighbors` | refus par défaut, « reaches state that no scope field confines » |

La raison est juste : le graphe est partagé entre les collections. Une entité extraite d'un document de la collection A est le même nœud qu'une entité du même nom extraite de la collection B. Un parcours depuis cette entité atteint les chunks des deux collections, et les arêtes portent la provenance des deux.

## Ce que Thalamus fait aujourd'hui

Rien. Un cerveau avec `collection` (clé confinée) n'a que les extraits de `knowledge_search` et `knowledge_memory_recall` : `th_recall` fusionne des passages, pas des entités. La fédération au niveau des entités attend cette fiche. Un cerveau sans `collection` (serveur entier, clé non confinée) a déjà les outils de graphe.

Contournement écarté : le gateway tient une clé non confinée par serveur, appelle les outils de graphe et filtre lui-même par collection. Il deviendrait alors le seul garde d'un accès complet au serveur, et il recopierait la règle de provenance du moteur. Les deux vont contre une couche mince.

## Ce que le moteur a déjà

- `CallerScope.Knowledge(collection) bool` (`pkg/cortexdb/memory_owner.go`, v2.135.0) : le périmètre de l'appelant dans le contexte, posé par les ports gRPC et REST pour une clé confinée.
- Le filtre des parcours de #76 et #77 (`pkg/cortexdb/graphrag_tool_search.go`) :
  - `withoutExpiredDocuments` retire d'un sous-graphe les nœuds de document et de chunk cachés, et les arêtes qui les touchent ;
  - `withoutExpiredProvenance` retire de la provenance d'une arête (`chunk_ids`, `source_chunk_id`, `document_id`) les chunks cachés. Une arête que seuls des documents cachés portent sort du résultat ;
  - `hiddenAmong` donne les documents cachés en une requête indexée, fermée sur panne sous la porte de relecture.
- Les outils qui passent par ce filtre depuis #76 et #77 : `expand_graph`, `search_paths`, `knowledge_memory_shortest_path`, `knowledge_memory_neighbors`, `get_nodes`, `knowledge_memory_expand_entity_context`.

La règle « un document caché n'apparaît dans aucun parcours » existe donc, avec sa classe d'outils et ses tests. La proposition y ajoute une seule cause : « un document hors du périmètre de l'appelant est caché ».

## La proposition

### PR 1 : le périmètre de l'appelant dans le filtre des parcours (environ 250 lignes)

Règle : sous un `CallerScope` dont `Knowledge` refuse une collection, un document de cette collection est caché comme un document expiré, sur tous les chemins qui passent par le filtre des parcours.

1. `hiddenDocuments` prend en plus le prédicat de collection de l'appelant. La requête lit déjà la table `documents`, qui porte la collection. Le calcul reste pur (`(sql, args)`, testé sans serveur) et une seule ligne fait la requête.
2. **Nœuds d'entité.** Une entité n'a pas de collection. Elle est servie à un appelant confiné seulement si un chunk visible la mentionne, ou si une arête avec une provenance visible la touche. Sinon le nom de l'entité dirait qu'un document d'une autre collection en parle.
3. **Propriétés des entités.** `upsert_entities` et les entités explicites d'`ingest_document` posent une `description` (`knowledge_tx.go`). Une description écrite pendant l'ingestion d'un document de la collection B est du texte de B. Sous un périmètre confiné, le nœud rendu garde `name` et `type`, et perd `description` et les autres propriétés libres. Point à soumettre au propriétaire : il connaît l'usage de ces propriétés.
4. **Arêtes.** Pas de règle nouvelle : la provenance hors périmètre est retirée comme une provenance expirée, et une arête que seuls des documents hors périmètre portent sort.
5. Sans `CallerScope`, ou avec un périmètre qui permet tout, rien ne change : bibliothèque embarquée, serveur sans clés, clé non confinée.

Pas de changement de la liste blanche dans cette PR : la règle se teste par l'API embarquée, avec `WithCallerScope`.

### PR 2 : la liste blanche ouvre les outils filtrés (environ 150 lignes)

Les outils qui passent par le filtre entrent dans `confinedTools`, avec `scope: collectionScope` et leurs arguments : `search_chunks_by_entities`, `search_paths`, `knowledge_memory_expand_entity_context`, `expand_graph`, `knowledge_memory_neighbors`, `get_nodes`, `find_nodes`. Chacun reçoit sa ligne et sa raison. Les outils qui lisent le graphe entier sans parcours restent refusés : `graph_list_all`, `graph_cypher_query`, `rank_graph_nodes`, `predict_graph_edges`, `disambiguate_mentions`, `graph_statistics`. Ils lisent le graphe stocké tel qu'il est, et #77 les nomme déjà comme tels.

Avant la PR, je vérifie un fait : `search_chunks_by_entities` et `find_nodes` passent-ils par le filtre ? Si non, la PR 1 les y fait passer d'abord, ou la PR 2 les laisse refusés.

## Tests, écrits avant le code

PR 1 (`pkg/cortexdb`, sans serveur) :

- `TestAWalkUnderACallerScopeServesOnlyItsCollections` : deux collections A et B. Un document de chaque mentionne « Kestrel Vale », et une relation vient de B seule. Sous un périmètre A, chaque outil de la classe rend le chunk de A et pas celui de B. La relation de B seule sort du résultat, et une arête que A et B portent garde la seule provenance de A. Un contrôle positif vérifie que chaque outil rend bien le contenu de A.
- `TestAnEntityOnlyAnotherCollectionMentionsIsNotServed` : une entité que seul B mentionne n'apparaît dans aucun outil sous un périmètre A, ni par `get_nodes` avec son identifiant.
- `TestAnEntityKeepsOnlyItsNameAndTypeUnderAScope` : la `description` posée depuis B ne sort pas sous un périmètre A.
- `TestWithoutAScopeEveryWalkServesEveryCollection` : sans `CallerScope`, les résultats sont les mêmes que sur `main` (scénarios `make diffmain`).
- Un test de la fonction pure qui rend le SQL du prédicat de collection.
- Mutations : prédicat de collection retiré ; entité servie sans mention visible ; `description` gardée ; périmètre ignoré quand le contexte n'a pas de `CallerScope`.

PR 2 (`pkg/authz`, `pkg/rpcserver`, `pkg/httpapi`) :

- `TestEveryToolAndEveryArgument…` de #72 tombe tant que chaque outil n'est pas classé.
- `TestAConfinedKeyWalksTheGraphOfItsCollectionOnly`, par gRPC et par REST : le scénario de la PR 1, de bout en bout, avec une clé confinée à A.
- Une table outil par outil (règle du propriétaire du 20 octobre 2026) : chaque outil ouvert, son argument d'identifiant, et le test qui le couvre.

## Questions pour le propriétaire

1. Une entité que seule une autre collection mentionne : la cacher (proposition) ou rendre son nom sans rien autour ?
2. Les propriétés libres d'une entité sous un périmètre confiné : les retirer (proposition), ou les rattacher à une provenance comme les arêtes ?
3. Le coût : une requête indexée de plus par parcours, déjà payée pour l'expiration. À mesurer sur 10 000 documents, comme pour #76.

## Après la fusion

Tranche Thalamus « fédération des entités » : `th_entities` sur `/mcp`, un appel par cerveau avec la clé de l'employé, fusion des entités par nom normalisé, provenance par cerveau, audit des identifiants lus. Les cerveaux avec et sans `collection` passent alors par le même chemin.
