# 16. Un instantané quantifié se relit avec le quantifieur qui l'a écrit, sinon l'index se reconstruit

**Statut** : fiche écrite le 30 octobre 2026, rien d'envoyé. **Taille estimée** : environ 30 lignes hors tests dans `pkg/core/store_index.go`, plus une table de tests de réouverture. **Dépend de** : rien. Défaut trouvé en préparant la fiche 12 : il a fait retirer `hnsw-sq8` de la PR #114.

## Le défaut

Une base SQLite dont l'index HNSW est quantifié (`Config.Quantization`) enregistre à la fermeture deux choses dans `index_snapshots` : la ligne `QUANTIZER` et un graphe dont les nœuds gardent leurs codes (`HNSWNode.Quantized`) et pas leurs vecteurs. À la réouverture, `loadIndexSnapshot` lit la ligne `QUANTIZER` sans regarder si la configuration du moment demande la même quantification :

- Quantification éteinte : la branche `else` construit `quantization.NewScalarQuantizer(dim, 0)`, qui rend `nil` et une erreur ignorée, puis appelle `Load` sur `nil`. **Panique à l'ouverture.**
- Quantification d'une autre sorte (SQ8 contre binaire), ou absente quand le graphe a été écrit avec : le graphe de codes est chargé tel quel. La réconciliation l'accepte (ses nœuds n'ont pas de dimension flottante), et la recherche mesure les distances avec un quantifieur qui ne lit pas ces codes, ou sans quantifieur. **Aucune erreur, et des résultats faux.**

Le cas inverse est déjà traité : un graphe écrit sans quantification et relu avec est reconstruit (« HNSW snapshot is not quantized, rebuilding »).

## Qui l'atteint aujourd'hui

`cortexdb.EmbeddedConfig`, configuration publique pour les petits appareils et citée dans le README, active SQ8. Tous les binaires du dépôt ouvrent avec `DefaultConfig`, sans quantification : `cortexdb-mcp-stdio` (`embedder.go`, `doctor.go`, `side_graph.go`), `cortexdb-connector-mcp`, `cortexdb-grpc` et `cortexdb-bench`. Une base créée par une application sous `EmbeddedConfig` fait donc paniquer l'outil MCP ou le diagnostic qu'on ouvre dessus. Et une application qui passe de l'une à l'autre (ou à la quantification binaire de HNSW) obtient des recherches fausses sans le savoir.

## Preuves (essai du 30 octobre 2026, `main` 21c7db8 = v2.141.0)

400 vecteurs aléatoires de dimension 32, fermés sous une configuration, rouverts sous une autre. Chaque recherche demande un vecteur stocké (top 1) : 40 requêtes. « Récentes » : rouvert aussitôt, toutes les lignes sont plus récentes que l'instantané et la réconciliation les réécrit, ce qui cache le défaut. « Anciennes » : `created_at` reculé de deux jours avant la réouverture, comme une base réelle.

| Fermée sous → rouverte sous | `main`, lignes récentes | `main`, lignes anciennes | avec la règle |
| :- | :- | :- | :- |
| SQ8 (`EmbeddedConfig`) → `DefaultConfig` | panique | panique | 40/40 |
| SQ8 → HNSW quantifié binaire | 40/40 | **1/40** | 40/40 |
| HNSW quantifié binaire → `DefaultConfig` | 40/40 | **1/40** | 40/40 |
| HNSW quantifié binaire → SQ8 | 40/40 | **1/40** | 40/40 |
| SQ8 → SQ8 | 37/40 | 36/40 | 37/40 et 36/40 (perte de SQ8) |
| les 11 autres paires (`DefaultConfig`, `IndexTypeBinary`) | 40/40 | 40/40 | 40/40 |

Une garde d'une ligne (ne lire la ligne `QUANTIZER` que si la quantification est active) supprime la panique mais donne 1/40 sur les lignes anciennes : elle remplace un arrêt par une réponse fausse. Ce n'est pas le correctif.

## La proposition (une PR, une règle)

Règle : un graphe quantifié se relit seulement si la configuration demande la même quantification et que son quantifieur se recharge ; sinon `loadIndexSnapshot` rend « pas chargé » et l'index se reconstruit depuis `embeddings`, comme dans le cas inverse.

Forme essayée : une méthode `loadQuantizer(data) bool` (fausse si la quantification est éteinte, si `NewScalarQuantizer` échoue ou si `Load` échoue), appelée quand la ligne `QUANTIZER` existe ; si elle est fausse, un message « index snapshot was quantized differently, rebuilding » et `return false, nil`. Les 16 paires du tableau passent.

Tests à écrire d'abord, sur `main` : la table des paires de réouverture, avec lignes anciennes, chaque paire devant trouver chaque vecteur stocké ; elle panique sur la première ligne et tombe à 1/40 sur trois autres.

## À vérifier avant l'envoi

- L'appel IVF passe aussi par `loadIndexSnapshot` : une base qui a gardé une ligne `QUANTIZER` d'un ancien HNSW quantifié reconstruirait son IVF à chaque ouverture. Soit la règle ne s'applique qu'à `HNSW`, soit la ligne est retirée quand la configuration change. À trancher par un test.
- Le coût : la reconstruction se paie une fois, à la première ouverture sous la nouvelle configuration ; l'instantané réécrit à la fermeture porte ensuite la bonne quantification.
- PostgreSQL n'est pas touché (pas d'instantané).

## Ce que Thalamus y gagne

Rien de direct : Thalamus sert des cerveaux sans quantification. La fiche corrige un défaut que la fiche 12 aurait rendu atteignable par une variable d'environnement, et rend possible plus tard un `hnsw-sq8` dans `cortexdb-grpc` (la moitié de la mémoire de HNSW).

Pour européens : sans effet sur les langues.
