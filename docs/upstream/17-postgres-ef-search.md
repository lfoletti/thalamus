# 17. Sur PostgreSQL, une recherche rend les `top_k` résultats demandés, pas 40 au plus

**Statut** : fiche écrite le 30 octobre 2026, rien d'envoyé. **Taille estimée** : 40 à 60 lignes hors tests, dans `pkg/core/store_postgres*.go` et `pkg/graph/pgvector.go`, plus une table de tests PostgreSQL. **Dépend de** : rien. Défaut trouvé en préparant la fiche 12 : il a fait retirer `binary` sur PostgreSQL de la PR #114.

## Le défaut

Un parcours d'index HNSW de pgvector rend au plus `hnsw.ef_search` lignes, 40 par défaut. CortexDB crée cet index (`idx_embeddings_vec`, `store_postgres.go`) dès que la dimension est fixée, et ne règle `ef_search` que dans la recherche binaire. Toute autre recherche par vecteur qui demande plus de 40 résultats en reçoit donc 40, **sans erreur**, dès que le planificateur choisit l'index ; le reste de la base n'est jamais vu.

La recherche binaire (`IndexTypeBinary`, `store_postgres_binary.go`) a le défaut inverse : elle fixe `SET LOCAL hnsw.ef_search = max(top_k × suréchantillon, 40)`, et pgvector refuse une valeur au-delà de 1000. Au suréchantillon par défaut (8), **toute recherche à partir de `top_k` 126 échoue** ; à 32, à partir de 32.

## Qui l'atteint aujourd'hui

Toute base PostgreSQL de dimension fixe (jusqu'à 2000) :

- `Search`, avec ou sans collection, `SearchWithACL`, `SearchWithAdvancedFilter` et le bras vectoriel de `HybridSearch` : même requête `ORDER BY vector <=> $1 LIMIT k` ;
- la recherche des nœuds du graphe (`pgVectorCandidates`, `pkg/graph/pgvector.go`), dont la limite par défaut est 50 : **même sans rien demander, elle rend 40 nœuds** ;
- la recherche des messages de chat n'est pas touchée (la table `messages` n'a pas d'index HNSW).

Un `top_k` de 50 à 100 est courant pour une étape qui reclasse ensuite (RRF, reranker, contexte GraphRAG) : c'est là que la perte passe inaperçue.

## Preuves (essai du 30 octobre 2026, `main` 21c7db8 = v2.141.0, PostgreSQL local, pgvector 0.8.7)

5 000 vecteurs aléatoires de dimension 16, `ANALYZE` fait, une requête ; nombre de lignes rendues :

| `top_k` | HNSW : `Search`, collection, ACL, filtre avancé | HNSW : filtre de métadonnées | binaire : `Search` |
| -: | :- | :- | :- |
| 10 | 10 | 10 | 10 |
| 40 | 40 | 40 | 40 |
| 41 | **40** | **40** | 41 |
| 100 | **40** | **40** | 100 |
| 500 | **40** | **40** | **erreur** (`4000 is outside the valid range for parameter "hnsw.ef_search" (1 .. 1000)`) |
| 1500 | **40** | 1500 (le planificateur passe au parcours séquentiel) | **erreur** |

Le plan de `ORDER BY vector <=> $1 LIMIT 100` est bien `Index Scan using idx_embeddings_vec`. Sur une base binaire, les recherches avec collection ou filtre passent par la requête exacte et rendent leur compte, parce que l'index binaire ne sert pas l'opérateur `<=>`.

Les deux remèdes de pgvector, mesurés sur la même table, 50 requêtes, contre la recherche exacte :

| Réglage dans la transaction | Lignes rendues pour `LIMIT 1500` | Rappel@100 |
| :- | -: | -: |
| aucun (`main`) | 40 | (40 lignes) |
| `hnsw.ef_search = 1000` | 1000 | |
| `hnsw.ef_search = 100` | | 0,999 |
| `hnsw.ef_search = 200` | | 1,000 |
| `hnsw.iterative_scan = strict_order` (pgvector ≥ 0.8) | 1500 | 0,998 |

## La proposition (une PR, une règle)

Règle : une recherche par vecteur sur PostgreSQL rend `min(top_k, lignes qui passent le filtre)` résultats, ou une erreur ; jamais moins sans le dire.

Forme proposée :

1. Une aide unique, appelée par chaque recherche par vecteur dans sa transaction : `SET LOCAL hnsw.ef_search = min(max(k, 40), 1000)`, où `k` est `top_k` (ou la liste courte du binaire). Les recherches qui lancent aujourd'hui `s.db.QueryContext` passent par `BeginTx` (lecture seule), comme la recherche binaire.
2. Au-delà de 1000, `SET LOCAL hnsw.iterative_scan = strict_order` si l'extension est en 0.8 ou plus (lue une fois à `Init`, comme le test de `binary_quantize`) ; sinon la requête exacte (`ORDER BY (vector <=> $1) + 0`, que l'index ne sert pas). Le préfixe `hnsw.` est réservé : un `SET` d'un paramètre inconnu d'une ancienne version **échoue** et ne peut pas être tenté à l'aveugle.
3. La recherche binaire garde sa liste courte de `top_k × suréchantillon` et suit la même règle, ce qui supprime l'erreur à partir de 126.

Tests à écrire d'abord, sur `main`, sous `CORTEXDB_TEST_POSTGRES` : une table de 2 000 vecteurs au moins (assez pour que le planificateur prenne l'index ; l'imposer par `SET LOCAL enable_seqscan = off` dans le test), `top_k` 41, 100 et 1500, chaque chemin du tableau, compte exact et égalité avec la recherche exacte au-dessus d'un rappel fixé ; la recherche des nœuds avec la limite par défaut ; la recherche binaire à `top_k` 126 et 500. Sur `main`, toutes les lignes en gras tombent.

## À vérifier avant l'envoi

- La CI amont n'a pas de PostgreSQL : les tests sautent, et Codecov comptera les lignes PostgreSQL comme non couvertes. Les déclarer dans une section `## Codecov`, ou extraire le calcul de `k` et le choix du réglage dans une fonction pure testée sans base.
- Le coût d'un `ef_search` plus grand : linéaire en gros ; à mesurer sur 100 000 vecteurs de dimension 384 avant de choisir entre « `ef_search` = `k` » et « `iterative_scan` toujours ».
- Une transaction de plus par recherche : vérifier qu'elle ne change pas la lecture sous un pool déjà en transaction (`UpsertBatchTx` existe : il y a peut-être un `SearchTx` à suivre).
- `hnsw.max_scan_tuples` (20 000 par défaut) borne le parcours itératif : au-delà, le compte redevient court. Le nommer, ou passer à la requête exacte pour `k` proche de la taille de la table.

## Ce que Thalamus y gagne

Rien de direct aujourd'hui : les cerveaux servis sont sur SQLite. Un cerveau PostgreSQL (gateway multi-utilisateur, base partagée) recevrait 40 morceaux pour un `top_k` de 50 ou 100, et la fiche 12 pourrait ensuite offrir `binary` sur PostgreSQL.

Pour européens : sans effet sur les langues.
