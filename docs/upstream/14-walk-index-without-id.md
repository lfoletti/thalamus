# 14. Les index de marche sans la colonne `id`

**Statut** : fiche écrite le 30 octobre 2026, rien d'envoyé. **Taille estimée** : 20 lignes hors tests (`pkg/graph/graph.go`, `pkg/graph/graph_ppr.go`, `pkg/cortexdb/graph_runtime_sql.go`), un test de plan de requête et un test d'ordre stable. **Dépend de** : rien.

## Le besoin

`graph_edges` porte deux index « couvrants » pour la marche (`pkg/graph/graph.go`) :

```sql
CREATE INDEX idx_edges_walk_from ON graph_edges(from_node_id, to_node_id, edge_type, weight, id);
CREATE INDEX idx_edges_walk_to   ON graph_edges(to_node_id, from_node_id, edge_type, weight, id);
```

`graph_edges` est une table à `rowid`, avec `id TEXT PRIMARY KEY`. SQLite range déjà le `rowid` dans chaque entrée d'index. La colonne `id` recopie donc dans chaque index la chaîne entière de l'identifiant de l'arête. Cet identifiant est long : 68 caractères en moyenne sur ai-quant, et 125 pour une relation entre notions (`edge:relation:entity:<nom>:entity:<nom>:<type>:doc:<empreinte>`).

La marche n'a pas besoin de l'identifiant lui-même. `pprIncidentEdges` (`graph_ppr.go`) le lit pour deux choses :

- écarter une arête vue deux fois, par ses deux extrémités ;
- trier les arêtes, pour que le sous-graphe ne dépende pas de l'ordre de lecture.

Les requêtes de `graph_runtime_sql.go` font `ORDER BY id` pour la même raison. Le `rowid` rend ces deux services : il est unique, il est stable entre deux lectures, et il est déjà dans l'index.

## La proposition (une PR, une règle)

Règle : les index de marche couvrent la marche sans recopier l'identifiant des arêtes.

1. **Schéma.** Les deux index sur `(…, edge_type, weight)` sans `id`, sous de nouveaux noms. L'ouverture retire les anciens, comme le moteur l'a fait pour `idx_edges_from` et `idx_edges_to` (`DROP INDEX IF EXISTS`).
2. **`pprIncidentEdges`.** La requête lit `rowid` à la place de `id`, puis écarte les doublons et trie par `rowid`. Rien d'autre dans `graph_ppr.go` ne lit `pprEdge.id` : il ne sert qu'aux doublons et au tri (v2.141.0, lignes 432, 444 et 609).
3. **`graph_runtime_sql.go`.** Les trois requêtes passent à `ORDER BY rowid`.
4. **Les autres requêtes ne changent pas.** Celles qui rendent l'identifiant à l'appelant lisent la table ou la clé primaire, comme aujourd'hui.

## Mesure du 30 octobre 2026 (copie d'ai-quant, v2.141.0)

| | Aujourd'hui | Sans `id` |
| :- | -: | -: |
| `idx_edges_walk_from` | 25,1 Mo | 12,9 Mo |
| `idx_edges_walk_to` | 25,1 Mo | 12,9 Mo |
| Fichier après `VACUUM` | 549,0 Mo | 524,8 Mo (−4,4 %) |

Le plan de la requête de marche reste couvrant :

```
EXPLAIN QUERY PLAN SELECT rowid, from_node_id, to_node_id, edge_type, weight FROM graph_edges WHERE from_node_id IN (…)
→ SEARCH graph_edges USING COVERING INDEX w_from (from_node_id=?)
```

Le gain suit le nombre d'arêtes : environ 100 Mo sur ai-engineering (estimation). Un index plus petit tient aussi mieux dans le cache de pages, que la borne mémoire du VPS compte.

## Ce qui change de comportement

- **L'ordre des arêtes de même poids** suit l'ordre d'insertion au lieu de l'ordre alphabétique des identifiants. La somme des scores de PPR peut différer au dernier chiffre, parce que l'ordre des additions change. Le test de PPR doit comparer à une tolérance, pas à l'égalité.
- **Un `VACUUM` peut renuméroter les `rowid`** d'une table sans `INTEGER PRIMARY KEY`. L'ordre reste stable entre deux lectures, mais peut changer après un `VACUUM`. Si l'amont tient à un ordre identique d'une base à sa copie compactée, il faut garder `ORDER BY id` dans `pprIncidentEdges` et payer une lecture de la table pour les seules arêtes retenues.

## Preuves à apporter

1. Test de plan : la requête de marche passe par un index couvrant, sans lecture de la table.
2. Test d'ordre : deux marches sur la même base rendent le même sous-graphe.
3. Test de PPR existant, vert à tolérance près.
4. Mesure de la taille et du temps d'une marche sur une copie d'ai-quant, avant et après.

## Langues européennes

Rien de propre à une langue.
