# 18. Sur PostgreSQL, un nœud écrit par lot est trouvé par son vecteur

**Statut** : fiche écrite le 30 octobre 2026, corrigée le même soir (chemin de l'ingestion, nœuds RDF) ; PR de code en préparation (branche `co-18-graph-batch-mirror`), rien d'envoyé. **Taille estimée** : 30 à 50 lignes hors tests dans `pkg/graph/graph_batch.go` (et une reprise des bases existantes), plus un test à deux backends. **Dépend de** : rien ; touche `pkg/graph/pgvector.go` comme la fiche 17, dans d'autres fonctions. Défaut trouvé en écrivant les tests de la fiche 17.

## Le défaut

Sur PostgreSQL, la recherche vectorielle du graphe (`vectorCandidates`, `pkg/graph/graph_hybrid.go`) lit la table `graph_node_vectors`, un miroir des vecteurs des nœuds indexé par HNSW. Seul `UpsertNode` tient ce miroir (`pgUpsertVector`, `graph.go`). Trois écritures ne le tiennent pas :

- `UpsertNodesBatch` (`graph_batch.go`) ;
- `upsertNodesBatchTx`, derrière `ExecuteBatch` et les transactions de connaissance (`knowledge_tx.go`) ;
- les nœuds RDF écrits avec les triplets (`upsertRDFTermNodeWithLabelTx`, `rdf.go`), qui portent un vecteur haché du terme.

Les nœuds qu'elles écrivent sont dans `graph_nodes`, avec leur vecteur, mais pas dans le miroir. Et `vectorCandidates` ne retombe sur la recherche en mémoire que si le miroir ne rend **rien** :

- miroir vide : la recherche passe au parcours complet (`GetAllNodes`) et trouve, lentement ;
- **un seul nœud dans le miroir** : la recherche ne voit que lui et ses semblables ; les nœuds écrits par lot ne sont **jamais trouvés**, sans erreur.

## Qui l'atteint aujourd'hui

L'ingestion : par l'outil d'ingestion (`graphrag_tool_ingest.go`), le document, ses morceaux et ses entités s'écrivent tous par lot, et aucun n'entre dans le miroir ; la recherche reste juste, par le parcours complet en Go. Le défaut se déclenche dès qu'**un** nœud entre dans le miroir par une écriture un par un : une décision, une étape d'exécution, une action d'ontologie, `hindsight`, ou le nœud du document de `graphrag.go` (écrit par `UpsertNode`). À partir de là, plus aucun nœud ingéré n'est trouvé par similarité (essai : un nœud écrit seul, puis un document ingéré ; le morceau ne revient pas pour son propre vecteur, seul le nœud écrit seul revient). Aussi : la mémoire (`memory_api.go`), l'import et la restauration du graphe (`graph_export.go`), les transactions de connaissance, les triplets RDF.

SQLite n'est pas touché : la recherche y passe par l'index en mémoire ou le parcours complet.

## Preuves (essai du 30 octobre 2026, `main` 21c7db8 = v2.141.0, PostgreSQL 17, pgvector 0.8.7)

200 nœuds de dimension 4 écrits par lot, puis 20 recherches `HybridSearch` (`TopK` 5), chacune avec le vecteur d'un de ces nœuds ; on compte les nœuds qui se trouvent eux-mêmes :

| Écriture | SQLite | PostgreSQL | lignes du miroir |
| :- | -: | -: | -: |
| `UpsertNodesBatch` seul | 20/20 | 20/20 (parcours complet) | 0 |
| `UpsertNodesBatch`, puis un nœud par `UpsertNode` | 20/20 | **0/20** | 1 |
| `ExecuteBatch`, puis un nœud par `UpsertNode` | 20/20 | **0/20** | 1 |

Le partage document par `UpsertNode`, morceaux et entités par lot, est lu dans le code de l'ingestion ; un test d'ingestion sur PostgreSQL le confirmera.

## La proposition (une PR, une règle)

Règle : tout nœud écrit dans `graph_nodes` sur PostgreSQL a son vecteur dans `graph_node_vectors`, ou n'y est pas s'il n'en a pas, comme `UpsertNode` le fait déjà.

Forme proposée :

1. `pgUpsertVector` prend l'exécuteur (base ou transaction) ; les deux écritures par lot l'appellent pour chaque nœud écrit, **dans leur transaction**, quand `vecCap.Enabled`. Une erreur du miroir compte comme une erreur du nœud (`FailedCount`), comme une erreur d'écriture.
2. Reprise des bases existantes : à `InitGraphSchema`, quand le miroir est actif, recopier les nœuds qui ont un vecteur et pas de ligne miroir (`LEFT JOIN graph_node_vectors ... WHERE m.node_id IS NULL`), en Go puisque le vecteur de `graph_nodes` est un blob encodé. Une fois par base ; les suivantes ne trouvent rien.

Tests à écrire d'abord, sur `main`, sous `CORTEXDB_TEST_POSTGRES` : le tableau ci-dessus (chaque nœud écrit par lot se trouve par son vecteur, un nœud écrit un par un présent) ; un nœud réécrit par lot sans vecteur sort du miroir ; une base écrite par l'ancien code se reprend à l'ouverture ; à l'échelle du moteur, un document ingéré sur PostgreSQL a autant de lignes miroir que de nœuds à vecteur.

## À vérifier avant l'envoi

- `UpsertNode` écrit son miroir hors de sa propre écriture : un échec laisse le nœud sans miroir. À aligner dans la même PR (une transaction), ou à nommer.
- Le coût de la reprise sur une grande base PostgreSQL (une lecture de tous les nœuds à vecteur) : la faire par pages, et la journaliser.
- `GraphFilter.NodeTypes`, appliqué après la limite (nommé dans la fiche 17) : une fois les morceaux et entités dans le miroir, une recherche typée sur les documents perdra des places au profit des morceaux. Mesurer avec un filtre, et rapprocher des deux fiches.
- CI amont sans PostgreSQL : lignes à déclarer dans `## Codecov`, ou partie testable sans base (le choix des nœuds à recopier).

## Ce que Thalamus y gagne

Rien de direct aujourd'hui : les cerveaux servis sont sur SQLite. Un cerveau PostgreSQL verrait ses morceaux et ses entités dans la recherche vectorielle du graphe, et pas seulement ses documents.

Pour européens : sans effet sur les langues.
