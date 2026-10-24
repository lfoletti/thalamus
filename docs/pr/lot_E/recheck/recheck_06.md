# Fiche 06 relue contre CortexDB v2.138.0 (23 octobre 2026)

Clone : `/Users/mini/Documents/granular_cortex/cortexdb`, `b2a1ab6` (tag v2.138.0). Rien n'a été poussé, publié ni commenté, et aucune branche n'a changé. Les sondes ont tourné dans une copie jetable (`git worktree` à v2.138.0, retirée à la fin), sans `CORTEXDB_TEST_POSTGRES`. Les fichiers de sonde sont gardés dans `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/recheck06_probes/`.

Convention : **[fait]** = lu dans le code ou observé par une sonde, avec sa référence. **[hypothèse]** = déduction ou proposition non vérifiée.

## 1. Correction d'une prémisse

- **[fait]** La barrière des parcours par collection n'est pas arrivée en v2.138.0. Elle date de #70 (`f1c6623`, 19 octobre 2026, notre propre correctif, publié en v2.133.0) : `entity_collections.go` en entier et la lecture de `collections` dans `graph_runtime_sql.go:16-20,52-74` (`git blame`).
- **[fait]** Entre v2.136.0 et v2.138.0, les fichiers des outils de graphe n'ont changé qu'avec #75, #76 et #77 (`0f505b6`, `548c9c3`, `477a27d`, publiés en v2.137.0). v2.138.0 (`7e0aa86`, `d9eb16d`) touche au rappel et à l'index vectoriel, pas au graphe.
- **[fait]** La fiche dépendait de la fusion de #76 et #77. C'est fait (v2.137.0).

## 2. Ce que le moteur a aujourd'hui

- **[fait]** Barrière de #70 : `entityKnownIn` (`entity_collections.go:45-55`). Elle n'a **qu'un seul appelant**, `graph_runtime.go:167`, dans `expandGraphChunkNeighborhoods`, c'est-à-dire l'expansion de graphe de `knowledge_search`, du rappel et de `search_graphrag_lexical`. Elle lit `opts.Collection` (l'argument de la requête), pas le `CallerScope`. Les chunks voisins passent par `chunksInScope` (`graphrag_ppr.go:739-777`), qui lit la collection dans `embeddings` ⨝ `collections`.
- **[fait]** `CallerScope.Knowledge` n'est lu que par `knowledgeReadPermitted` (`memory_owner.go:117-120`), dont le seul appelant est `SearchKnowledge` (`knowledge_api.go:283`). `Recall` ne sert aucun fait de graphe à un appelant confiné (`brain.go:117`). Aucun outil de graphe ne lit le `CallerScope`.
- **[fait]** Filtre des documents cachés (expiration et relecture) :
  - `hiddenDocuments` (`expiry.go:127-152`), une requête sur `documents` ;
  - `hiddenAmong` (`expiry.go:174-192`), fermé sur panne sous la porte de relecture ;
  - `withoutExpiredDocuments` (`graphrag_tool_search.go:257-294`), qui ne retire que les nœuds `chunk`, `document` et `memory:` cachés et les arêtes qui les touchent. **Il ne regarde jamais un nœud d'entité.**
  - `withoutExpiredProvenance` (`graphrag_tool_search.go:325-402`) ;
  - `loadToolChunks` (`graphrag_tool_helpers.go:86-127`, filtre à la ligne 106).
- **[fait]** La table `documents` n'a pas de colonne de collection (`pkg/core/store_init.go:123-134`). Aucun code ne pose `metadata["collection"]` sur un document. Les nœuds `chunk` et `document` du graphe portent `document_id` mais pas la collection (`knowledge_tx.go:238-296`). La collection n'existe que sur `embeddings.collection_id` → `collections.name` (`store_init.go:138`). **La fiche se trompe** quand elle dit que « la requête lit déjà la table `documents`, qui porte la collection » (PR 1, point 1).
- **[fait]** Un nœud d'entité porte plusieurs propriétés lisibles par n'importe quel outil de graphe :
  - `collections`, la liste de toutes les collections qui le mentionnent ;
  - `source_document_ids` (`graphrag_tool_ingest.go:796-807`) ;
  - `description`, écrasée par le dernier écrivain (`knowledge_tx.go:615-617`, puis `mergePriorEntityProperties`, `graphrag_tool_ingest.go:756-789`) ;
  - les métadonnées libres.

  Une arête de relation porte `document_id` et `chunk_ids` (`knowledge_tx.go:676-681`).

## 3. Outil par outil, sous un `CallerScope` confiné à `alpha`

Scénario de la sonde (`zz_recheck06_test.go`, API embarquée, `WithCallerScope`) :

- `a1` dans `alpha` : « Kestrel Vale runs the alpha mill » ;
- `b1` dans `beta` : « Kestrel Vale secretly funds Project Nightjar with BETASECRET money » ;
- une entité « Kestrel Vale » partagée, avec une description posée depuis `beta` (`BETADESC`) ;
- une entité « Project Nightjar » de `beta` seul ;
- une relation `funds` de `beta` seul.

Ce que les outils ont rendu :

| Outil | Collection en argument | Barrière #70 | Filtre des cachés | Fuite observée (sonde) | Preuve |
| :- | :- | :- | :- | :- | :- |
| `search_chunks_by_entities` | non (`graphrag_tool_types.go:172-176`) | non | oui, par `loadToolChunks` (`graphrag_tool_search.go:178`) | texte du chunk `b1` (`BETASECRET`), `Nightjar`, `b1` | `graphrag_tool_search.go:139-183`, `graph.Neighbors` ligne 154 sans collection |
| `search_paths` | non (`graphrag_paths.go:26-56`) | non | oui, sur les arêtes (`graphrag_paths.go:154`) et sur le texte par `GetChunks` (`:200-201`) | relation `funds`, `Nightjar`, `b1`, texte `BETASECRET` (`with_text`) | `graphrag_paths.go:114-208` ; `Seeds` révèle l'existence d'une entité |
| `knowledge_memory_expand_entity_context` | non (`brain_types.go:134-150`) | non | oui, par `ExpandGraph` et `SearchChunksByEntities` | tout : texte de `b1`, `BETADESC`, `Nightjar`, `beta` | `brain.go:210-296` (lignes 225, 236, 262) |
| `expand_graph` | non (`graphrag_tool_types.go:184-196`) | non | oui (`graphrag_tool_search.go:248`) | tout, plus le **texte d'une mémoire de zeus** (`zz_recheck06c_test.go`) | `graphrag_tool_search.go:186-250` |
| `get_nodes` | non | non | oui (`:411`) | `BETADESC`, `Nightjar`, `collections:["alpha","beta"]` | `graphrag_tool_search.go:406-413` |
| `find_nodes` | non | non | **non** | tout ; c'est une recherche plein texte sur `graph_nodes.content`, donc sur le texte des chunks de toutes les collections | `graphrag_tool_find_nodes.go:65-66` (requête sans filtre), `:139` |
| `knowledge_memory_neighbors` | non (`brain_types.go:162-170`) | non | oui (`brain.go:327`) | tout | `brain.go:299-333` |
| `knowledge_memory_shortest_path` | non (`brain_types.go:179-184`) | non | oui (`brain.go:357`) | chemin `funds` vers `Nightjar`, `b1`, `BETADESC` | `brain.go:336-371` |
| `cortex_query` (collection `alpha` forcée) | oui | non | oui (`query_api.go:169`) | voies vecteur, lexicale et hybride : rien ; voie graphe : voir section 5 | `query_api.go:144-371` |
| `knowledge_search` en mode graphe (`alpha`), contrôle positif | oui | oui | oui | rien | `graph_runtime.go:167` |

**[fait] Bogue hors périmètre.** Sans aucun `CallerScope`, `find_nodes` sert le texte d'un document expiré (`zz_recheck06b_test.go` : `servesExpiredText=true` ; `get_nodes` sur le même nœud : `false`). Le même trou vaut sous la porte de relecture **[hypothèse : même code, non sondé]**. C'est un trou de la classe de #76 et #77 : la fiche listait les outils qui passent par le filtre, et `find_nodes` n'y est pas.

**[fait] Mémoires.** Les nœuds `memory:` sont dans le graphe partagé (`memory_api.go:1004-1011`). Sous un périmètre `alpha` dont `Memory` refuse tout, `expand_graph` rend le texte d'une mémoire de l'utilisateur zeus. `hiddenMemoryNodes` (`graphrag_tool_search.go:299-315`) ne regarde que la porte de relecture. La fiche ne parle pas des mémoires.

**[fait] Ports.** Les routes typées sont fermées.

- REST : `POST /v1/graph/expand` répond 403 sans `collection`, et 400 avec (`json: unknown field "collection"`, `DisallowUnknownFields`, `pkg/httpapi/json.go:33`).
- `/v1/tools/expand_graph` répond 403 par la liste blanche (sonde `httpapi_zz_recheck06_test.go`).
- gRPC : aucun RPC typé pour ces outils (`pkg/rpcserver`, seul `QuerySparql` touche au graphe) **[hypothèse : vérifié par grep, pas par un test]**.

## 4. Ce que le propriétaire a fait, ce qui manque

Fait en amont :

- #70 : la barrière par entité, sur la seule expansion de `knowledge_search` et du rappel ;
- #78 : la voie graphe de `cortex_query` reste dans la collection ;
- #76 et #77 : le filtre des cachés sur six outils de graphe ;
- v2.135.0 : `CallerScope`.

Manque :

1. `find_nodes` ne passe par aucun filtre (bogue d'expiration et de relecture, indépendant des clés).
2. Aucun outil de graphe ne lit le `CallerScope`, pour les documents comme pour les mémoires.
3. Les nœuds d'entité ne sont jamais filtrés, ni par la barrière de #70, ni par le périmètre.
4. Les propriétés d'entité (`description`, `collections`, `source_document_ids`) sortent telles quelles.
5. La voie graphe de `cortex_query` à plus d'un saut sert de révélateur structurel (section 5).

## 5. `cortex_query` : la raison du refus tient-elle encore ?

- **[fait]** La raison écrite (`allowlist.go:251`, « its graph prefetch searches chunks of every collection ») **n'est plus exacte pour le contenu**. Depuis #78, `runGraphQueryPrefetch` (`query_api.go:325-371`) garde seulement les chunks de `req.Collection` (`chunksInScope`, ligne 346). Les voies vecteur, lexicale et hybride passent `Collection` au magasin (`query_api.go:279, 286, 304`). Chaque voie passe ensuite le filtre des cachés (`query_api.go:169`).
- **[fait] Fuite résiduelle.** La voie graphe n'applique pas `entityKnownIn` et suit les relations de toutes les collections. Avec la graine « Project Nightjar » (entité de `beta` seul) et la collection `alpha` :
  - `max_hops: 1` ne rend rien ;
  - `max_hops: 2` rend le chunk `a1`.

  L'appelant `alpha` apprend ainsi que « Project Nightjar » existe et qu'une relation de `beta` le relie à une entité d'`alpha`. Aucun texte de `beta` ne sort (sonde : `leaks=[]`).
- **[fait] La porte est déjà ouverte en REST.** `POST /v1/query` (`routes.go:46`, `access.go:65`, Read) appelle `db.Query` pour une clé confinée qui envoie `"collection":"alpha"`. La sonde REST le confirme : 200, chunk `a1` seul ; et avec `max_hops: 2` sur la graine de `beta`, 200 avec `a1`. Le refus de l'outil ne protège donc rien que la route typée ne serve déjà. Le révélateur à deux sauts existe aujourd'hui en v2.138.0 par REST, sous une clé confinée.
- **[fait]** `db.Query` n'appelle pas `knowledgeReadPermitted`. Seule la porte du port (collection forcée) le confine.
- **[fait] Voie `source`.** Elle n'est pas dans le schéma : l'énumération de `kind` ne liste pas `source`, et `prefetch[].source` est absent (`query_tool_schema.go:27-41`). La liste blanche refuse donc d'office tout argument non classé, `prefetch[].source` compris (`allowlist.go:27-29`). Avec `kind:"source"` sans nom, la voie échoue (`query_source.go:108-111`). Aucun binaire n'enregistre de source (`WithQuerySource` n'a pas d'appelant hors tests). La voie relit les lignes par identifiant et écarte l'autre collection, mais laisse passer une ligne de collection vide (`query_source.go:142`).

**Conclusion.** `cortex_query` peut s'ouvrir, `prefetch[].max_hops` refusé tant que le moteur ne borne pas la voie graphe.

Arguments à classer :

- portée : `collection` ;
- données : `query`, `query_vector`, `entity_names`, `prefetch`, `prefetch[].name`, `prefetch[].kind`, `prefetch[].query`, `prefetch[].query_vector`, `prefetch[].entity_names`, `prefetch[].weight`, `prefetch[].limit`, `prefetch[].keywords`, `prefetch[].alternate_queries`, `fusion`, `filter` et ses sous-chemins (`must`, `should`, `must_not` et leurs `field`, `op`, `value`, `values`, `number`, `has_number`), `formula` et ses sous-chemins, `limit`, `rrf_k`, `include_raw` ;
- refusé : `prefetch[].max_hops`, « walks past the mention edge, over relations another collection asserts » ;
- `prefetch[].source` : refusé d'office tant qu'il n'est pas dans le schéma. S'il y entre, il va dans `denied` : « asks an external lane whose rows the guard does not see ».

Le défaut de saut de la voie graphe est 1 (`graphrag_tool_search.go:146-148`), donc `entity_names` reste sûr. Les chemins exacts des sous-objets se lisent avec `schemaPaths` du test de complétude **[hypothèse sur l'orthographe exacte, par exemple `filter.must[].field`]**.

## 6. Ce que la fiche disait de faux ou de trop court

- « La table `documents` porte la collection » : faux (section 2).
- `find_nodes` « passe-t-il par le filtre ? » : non, et il cache aussi un bogue d'expiration.
- Les mémoires du graphe sont oubliées.
- `knowledge_memory_shortest_path` manque à la liste des outils à ouvrir en PR 2.
- Les entités ont déjà leur liste de collections (#70). La règle « entité servie seulement si un chunk visible la mentionne » peut se lire sur cette propriété au lieu d'un calcul de mentions.
- Les ports ne peuvent pas forcer `collection` sur ces outils, puisque leurs schémas n'en ont pas. La PR de la liste blanche doit les classer comme `knowledge_get` (sans `scope`, confinés par la façade), pas avec `scope: collectionScope`.

## 7. Conception révisée

Règle du propriétaire : environ 400 lignes hors tests par PR, une règle par PR, tests qui tombent sans le code, table outil par outil. Estimations en lignes hors tests, **[hypothèse]** toutes.

### PR A : `find_nodes` laisse de côté les nœuds cachés (environ 20 à 40 lignes)

- Règle : un nœud que le filtre des cachés retire (chunk ou document d'un document expiré ou caché par la relecture, mémoire cachée) n'est pas une correspondance de `find_nodes`.
- Fichier : `pkg/cortexdb/graphrag_tool_find_nodes.go`. On passe les `match.Nodes` par `withoutExpiredDocuments`, comme `get_nodes`, avant de choisir `Match`. Un nom dont toutes les correspondances sont cachées rend `nodes: []` sans `match`.
- Tests d'abord :
  - `TestFindNodesLeavesOutTheNodesOfAnExpiredDocument` ;
  - `TestFindNodesLeavesOutAProposedDocumentUnderTheReviewGate` ;
  - `TestFindNodesStillFindsAnEntityThatAHiddenChunkMentions` (contrôle positif).
- Mutation : retirer l'appel.
- Sans dépendance, à envoyer d'abord : c'est un correctif de la classe de #76 et #77, sans rapport avec les clés.

### PR B1 : le périmètre de l'appelant cache documents et mémoires dans tout parcours (environ 150 à 200 lignes)

- Règle : sous un `CallerScope`, un document dont la collection n'est pas permise est caché comme un document expiré, et une mémoire dont le seau n'est pas permis est cachée comme une mémoire non relue.
- Point d'étranglement unique : `hiddenAmong` (`expiry.go:174`). Sous un `CallerScope` avec `Knowledge`, il ajoute les documents dont `Knowledge(collection)` est faux. La collection se lit par une fonction pure qui rend `(sql, args)` (`SELECT DISTINCT e.doc_id, COALESCE(c.name,'') FROM embeddings e LEFT JOIN collections c ON c.id = e.collection_id WHERE e.doc_id IN (…)`), puis une ligne d'I/O. Comportement fermé sur panne sous un périmètre. Un document sans chunk est caché.
- `hiddenMemoryNodes` gagne la cause `memoryReadPermitted`.
- Tous les chemins qui passent par `withoutExpiredDocuments`, `withoutExpiredProvenance` et `loadToolChunks` en héritent : `expand_graph`, `get_nodes`, `knowledge_memory_neighbors`, `knowledge_memory_shortest_path`, `search_paths`, `search_chunks_by_entities`, `knowledge_memory_expand_entity_context`, et `find_nodes` après la PR A.
- Coût : une requête indexée de plus sous un périmètre seulement (index sur `embeddings.doc_id` à vérifier).
- Fichiers : `expiry.go`, `graphrag_tool_search.go` (mémoires), `memory_owner.go` (prédicat).
- Tests d'abord :
  - `TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection`, table des huit outils sur le scénario de la sonde. Contrôle négatif : ni `BETASECRET` ni `b1`. Contrôle positif : `alpha mill` rendu par chaque outil qui rend du texte ;
  - `TestUnderACallerScopeAGraphWalkLeavesOutAnotherUsersMemory` ;
  - `TestUnderACallerScopeAnEdgeOnlyAnotherCollectionBacksIsLeftOut`, et une arête portée par A et B garde la provenance de A seule ;
  - `TestUnderACallerScopeAnUnreadableCollectionHidesTheDocument` (colonne `collections.name` renommée) ;
  - `TestWithoutACallerScopeTheGraphToolsAnswerAsBefore` ;
  - test pur de la fonction SQL.
- Mutations : prédicat de collection retiré ; périmètre ignoré ; ouverture sur panne ; cause mémoire retirée.

### PR B2 : une entité qu'aucune collection permise ne connaît n'est pas servie, et ses propriétés sont réduites (environ 150 à 200 lignes)

Deux sous-règles liées. Une division en B2a et B2b est possible si le propriétaire tient à une règle stricte par PR.

- B2a. Sous un `CallerScope`, `withoutExpiredDocuments` retire un nœud d'entité dont la liste `collections` est non vide et ne contient aucune collection permise, avec ses arêtes. `search_paths` rend une graine retirée dans `unresolved`. `knowledge_memory_expand_entity_context` n'écho pas l'identifiant **[hypothèse : `resolveEntityInputs` non lu en détail]**.
- B2b. Sous un `CallerScope`, un nœud d'entité rendu garde `name` et `type`. On filtre `collections` aux collections permises et `source_document_ids` aux documents visibles. On retire `description` et les métadonnées libres, parce que la description est celle du dernier écrivain et qu'on ne peut pas l'attribuer.
- Fichiers : `entity_collections.go` (prédicat sous périmètre), `graphrag_tool_search.go`, `graphrag_paths.go`.
- Tests d'abord :
  - `TestAnEntityOnlyAnotherCollectionKnowsIsNotServed` (y compris par `get_nodes` sur son identifiant et par `find_nodes`) ;
  - `TestUnderACallerScopeAnEntityKeepsOnlyItsNameAndType` ;
  - `TestAPathSeedTheScopeHidesIsReportedUnresolved` ;
  - `TestAnEntityWithNoCollectionListIsServedUnderAScope`, ou son contraire selon la réponse du propriétaire.
- Mutations : barrière d'entité retirée ; `description` gardée ; `collections` non filtrée.
- Question au propriétaire : une entité sans liste (écrite par `upsert_entities` ou avant #70) est « connue partout » depuis #70. Sous un périmètre, faut-il la garder ou la cacher ?

### PR C : la liste blanche ouvre les huit outils de graphe (environ 120 à 180 lignes, surtout la table)

- Règle : `search_chunks_by_entities`, `search_paths`, `knowledge_memory_expand_entity_context`, `expand_graph`, `get_nodes`, `find_nodes`, `knowledge_memory_neighbors` et `knowledge_memory_shortest_path` entrent dans `confinedTools` sans `scope`, `stores: storeKnowledge | storeMemory`, confinés par la façade comme `knowledge_get`.
- Les identifiants de nœud sont des données : un nœud caché répond comme un nœud absent, même raisonnement que `plan.filters.document_ids`, `allowlist.go:120-125`.
- Arguments, lus dans les schémas :
  - `search_chunks_by_entities` : `entity_names`, `top_k`, `max_hops` ;
  - `search_paths` : `entity_names`, `node_ids`, `max_depth`, `direction`, `edge_types`, `relation_policies.*`, `decay`, `use_edge_weights`, `max_paths`, `max_expansions`, `include_bookkeeping`, `with_text` ;
  - `expand_graph` : `node_ids`, `max_hops`, `edge_types`, `node_types`, `limit`, `relation_policies.*` ;
  - `get_nodes` : `node_ids` ;
  - `find_nodes` : `names`, `node_types`, `limit` ;
  - `knowledge_memory_expand_entity_context` : `entity_ids`, `entity_names`, `max_hops`, `edge_types`, `node_types`, `limit`, `top_k_chunks`, `max_context_chunks`, `max_context_chars`, `per_document_limit`, `chunk_window`, `retrieval_mode`, `disable_graph`, `graph_light`, `max_entities_per_chunk` ;
  - `knowledge_memory_neighbors` : `node_id`, `entity_name`, `max_depth`, `edge_types`, `node_types`, `direction`, `limit` ;
  - `knowledge_memory_shortest_path` : `from_node_id`, `to_node_id`, `from_entity_name`, `to_entity_name`.
- Restent refusés, en lisant le graphe stocké sans parcours : `graph_list_all`, `graph_cypher_query`, `rank_graph_nodes`, `predict_graph_edges`, `disambiguate_mentions`, `graph_statistics`, `get_chunks` et `build_context`. Ces deux derniers pourraient s'ouvrir après B1, puisque `loadToolChunks` portera le périmètre **[hypothèse]**.
- Fichiers : `pkg/authz/allowlist.go`, `allowlist_test.go`, tests `pkg/rpcserver` et `pkg/httpapi`.
- Tests d'abord :
  - `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys` (existe, tombe) ;
  - `TestAConfinedKeyWalksTheGraphOfItsCollectionOnly`, par gRPC (`CallTool`) et par REST (`/v1/tools/<nom>`), sur le scénario de B1 et B2 ;
  - table outil par outil (outil, arguments d'identifiant, test qui le couvre).
- Dépend de A, B1 et B2.

### PR D : la liste blanche ouvre `cortex_query` (environ 40 à 80 lignes)

- Règle : `cortex_query` entre dans `confinedTools` avec `scope: collectionScope` et les arguments de la section 5, `prefetch[].max_hops` refusé. `prefetch[].source` reste refusé d'office.
- Tests d'abord :
  - `TestAConfinedKeyQueriesItsCollectionOnly` (gRPC et REST, voies vecteur, lexicale, hybride et graphe, contrôle positif) ;
  - `TestAConfinedKeyMayNotSendMaxHopsToTheGraphLane` ;
  - `TestAConfinedKeyMayNotNameAQuerySource`.
- Ne dépend que de lui-même : peut partir avant B1.

### PR E, à proposer : la voie graphe d'une requête par collection suit seulement ce que la collection connaît (environ 60 à 120 lignes)

- Fuite actuelle : `/v1/query` par REST sert le révélateur à deux sauts à une clé confinée.
- Règle possible : avec une collection, la voie graphe garde les graines connues de la collection (`entityKnownIn`) et ne suit que les relations dont une provenance est visible. Plus simple : sous un `CallerScope`, la voie prend un saut.
- Tests d'abord :
  - `TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation` (le cas `max_hops: 2` de la sonde) ;
  - `TestTheGraphLaneWithoutAScopeKeepsItsHops`.
- Après E, `prefetch[].max_hops` peut passer en données dans la liste blanche.

Ordre proposé : A, D, B1, B2, C, puis E (ou E avant D, et D ouvre `max_hops`).

## 8. Questions au propriétaire

1. Une entité sans liste `collections`, sous un périmètre : la garder (comme #70) ou la cacher ?
2. Les propriétés libres d'une entité sous un périmètre : les retirer (proposition B2b) ?
3. `POST /v1/query` sert aujourd'hui la voie graphe à deux sauts à une clé confinée (section 5) : correctif moteur (E) ou saut unique sous un périmètre ?
4. Coût de B1 : une requête de plus par appel confiné. À mesurer sur 10 000 documents comme pour #76.
