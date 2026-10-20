# `cortex_query` et la liste blanche : analyse argument par argument (20 octobre 2026)

Base de la PR qui suivra GC-10 : ajouter `cortex_query` aux outils qu'une clé confinée à une collection peut appeler. Une clé confinée à un utilisateur ne l'atteint pas : l'outil lit le magasin de connaissance seulement.

## Ce que chaque voie fait de la collection (`pkg/cortexdb/query_api.go`, `query_source.go`)

| Voie (`prefetch[].kind`) | Appel | Collection honorée | Porte d'expiration / de relecture |
| :- | :- | :- | :- |
| `vector` | `store.Search(vector, {Collection})` | oui | **non** |
| `lexical` | `SearchTextOnly(query, {Collection})` | oui | **non** |
| `hybrid` (défaut) | `store.HybridSearch(vector, query, {Collection})` | oui | **non** |
| `graph` | `SearchChunksByEntities` puis `chunksInScope` | oui depuis GC-10 (avant : fuite, point 3 de la revue de #72) | **non** |
| `source` | `QuerySource.Search` externe puis `store.GetByID` par id ; une ligne d'une autre collection est écartée, **une ligne sans collection passe** | partielle | **non** |

Constat : aucune voie de `cortex_query` n'appelle `filterScoredEmbeddings` ni `allowDocument`. Avec #76 et #77 telles qu'amendées, un document expiré ou proposé est servi par `cortex_query` alors que le corps de #76 dit « every retrieval path that reads chunks ». À corriger dans #76 (la porte dans `Query()` après chaque voie, une ligne) et à couvrir dans #77 (cas de test sous la porte de relecture). Trouvé le 20 octobre 2026 au soir, pendant cette analyse.

## Arguments du schéma (`query_tool_schema.go`)

| Argument | Classe pour la liste blanche | Ce qu'il atteint |
| :- | :- | :- |
| `collection` | champ de portée (`scope`) | toutes les voies, sauf la voie `source` pour une ligne sans collection |
| `query` | donnée | texte de la voie lexicale, vectorielle (par l'embedder) ou hybride |
| `query_vector` | donnée | vecteur de la voie vectorielle ou hybride |
| `entity_names` | donnée | noms d'entités de la voie graphe ; nœuds partagés, chunks filtrés par `chunksInScope` depuis GC-10 |
| `prefetch[].name`, `kind`, `weight`, `limit`, `max_hops`, `keywords`, `alternate_queries`, `query`, `query_vector`, `entity_names` | données | la voie nommée, dans la collection |
| `prefetch[].source` | **à refuser** à une clé confinée | une source externe enregistrée sur le serveur ; ses résultats sont rechargés par id et une ligne sans collection passe. À admettre seulement si `runQuerySourcePrefetch` écarte toute ligne dont la collection est vide quand la requête en nomme une |
| `filter.must[]`, `should[]`, `must_not[]` (`field`, `op`, `value`, `values`, `number`, `has_number`) | données | filtre les résultats déjà obtenus dans la collection ; un `field` `id`, `doc_id` ou `collection` ne fait que restreindre, donc pas d'oracle |
| `formula.base_weight`, `field_boosts[]`, `numeric_boosts[]` | données | reclassement des résultats déjà obtenus |
| `fusion`, `limit`, `rrf_k`, `include_raw` | données | fusion et taille de la réponse |

Aucun champ porteur d'identifiant de ligne : rien pour `GuardToolRows`. Pas de porte : un seul magasin.

## Ligne proposée (`pkg/authz/allowlist.go`), après la fusion de GC-10 et la correction des portes

```go
"cortex_query": {scope: collectionScope, stores: storeKnowledge,
    args: paths([]string{"query", "query_vector", "entity_names", "fusion", "limit", "rrf_k", "include_raw",
        "prefetch", "prefetch.*", "filter", "filter.*", "formula", "formula.*"}),
    denied: map[string]string{"prefetch.source": "asks an external source, whose rows are loaded by id without a collection check"}},
```

À vérifier avant d'écrire la PR : la forme exacte des chemins imbriqués que `checkPaths` accepte (`prefetch[].source` ou `prefetch.source`), et la ligne de `search_text` comme modèle.

## Tests à écrire d'abord

1. Une clé `notes` appelle `cortex_query` avec `collection: notes` et une entité partagée : seuls les chunks de `notes` reviennent (gRPC et REST).
2. La même clé avec `prefetch: [{kind: source, source: x}]` est refusée avec la raison.
3. La même clé avec `collection: secrets`, ou sans collection, est refusée (règle de portée).
4. Sous `WithReviewGate` et avec un document expiré : `cortex_query` ne les sert pas (ces deux cas vont dans #76 et #77).
