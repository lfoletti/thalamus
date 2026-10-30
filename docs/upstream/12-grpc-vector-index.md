# 12. `cortexdb-grpc` laisse choisir l'index des vecteurs

**Statut** : fiche écrite le 30 octobre 2026, rien d'envoyé. **Taille estimée** : 40 lignes hors tests dans `cmd/cortexdb-grpc`, un test de lecture des variables. **Dépend de** : rien. Le moteur a déjà les trois index ; seul le serveur ne permet pas de les choisir.

## Le besoin de Thalamus

Les cerveaux de Contaixt sont servis par `cortexdb-grpc` sur un VPS dont la mémoire est la borne (`MemoryMax=6G` pour une dizaine de cerveaux). Le serveur ouvre chaque base avec `cortexdb.DefaultConfig` (`cmd/cortexdb-grpc/main.go`) : un index HNSW en float32, dont l'instantané recopie tous les vecteurs dans `index_snapshots`. Sur ai-quant, cet instantané pèse 126 Mo sur 549, et l'index tient 300 Mo de mémoire résidente pour 40 457 vecteurs de dimension 384.

Le moteur sait déjà faire plus léger :

- `QuantizationConfig{Type: "scalar", NBits: 8}` garde l'index HNSW en codes SQ8. `EmbeddedConfig` l'emploie déjà.
- `IndexTypeBinary` garde un code d'un bit par dimension, soit 1/32 des vecteurs. Il recalcule le score exact des candidats avec les vecteurs de la table `embeddings`, et n'écrit aucun instantané : il se reconstruit à l'ouverture en lisant la table (`pkg/core/store_binary.go`).

`cortexdb-grpc` n'expose aucun des deux.

## La proposition (une PR, une règle)

Règle : l'opérateur choisit l'index des vecteurs au démarrage du serveur. Le défaut ne change pas.

- `-vector-index` (variable `CORTEXDB_VECTOR_INDEX`) : `hnsw` (défaut, comme aujourd'hui), `hnsw-sq8`, ou `binary`.
- `-binary-oversample` (variable `CORTEXDB_BINARY_OVERSAMPLE`) : le multiplicateur de la liste de candidats de l'index binaire (`BinaryConfig.Oversample`). 0 garde le défaut du moteur, qui est 8.
- Une valeur inconnue arrête le serveur avec un message, comme `CORTEXDB_EMBED_DIM` invalide.
- Le journal de démarrage nomme l'index choisi, comme il nomme l'encodeur.

Points à vérifier dans l'essai :

1. Une base passée de `hnsw` à `binary` garde-t-elle ses lignes HNSW dans `index_snapshots` ? Si oui, la PR doit dire qui les retire. Sinon, le gain disque n'existe qu'après un nettoyage.
2. `knowledge_search` et les outils GraphRAG passent-ils tous par `Store.Search` ? Ils devraient donc profiter du choix sans autre changement.

## Mesure du 30 octobre 2026

La mesure porte sur une copie d'ai-quant (v2.141.0, Apple M2, une reconstruction de cerveau tournait en même temps). Les 201 questions sont chacune le milieu de deux vecteurs de morceaux. La référence est la recherche exacte (`IndexTypeFlat`), à 503 ms par question. Le programme de mesure appelle `cortexdb.Open` puis `Vector().Search` avec `TopK` 10.

| Index | Ouverture | Mémoire résidente ouverte | Tas | Par question | Rappel@10 | Même premier | Instantané |
| :- | -: | -: | -: | -: | -: | -: | -: |
| `hnsw` (aujourd'hui) | 0,80 s | 300 Mo | 136 Mo | 1,2 ms | 0,862 | 0,766 | 126 Mo |
| `hnsw-sq8` | 0,31 s (30,6 s la première fois, le temps de reconstruire l'index en codes) | 153 Mo | 88 Mo | 1,1 ms | 0,855 | 0,751 | 51 Mo |
| `binary`, suréchantillon 8 | 0,32 s | 42 Mo | 9 Mo | 1,4 ms | 0,777 | 0,955 | aucun |
| `binary`, suréchantillon 32 | 0,32 s | 41 Mo | 9 Mo | 5,0 ms | 0,904 | 0,945 | aucun |
| `binary`, suréchantillon 64 | 0,32 s | 42 Mo | 9 Mo | 9,6 ms | 0,941 | 0,955 | aucun |

Ce que la table montre :

- L'index HNSW d'aujourd'hui n'est pas exact : il manque 14 % des dix plus proches, et le plus proche une fois sur quatre.
- L'index binaire trouve le plus proche 19 fois sur 20, parce qu'il recalcule le score exact.
- Avec un suréchantillon de 32, il dépasse HNSW en rappel, pour 5 ms par question, sept fois moins de mémoire et aucun instantané.
- SQ8 divise la mémoire par deux et l'instantané par 2,5, pour un rappel égal à HNSW.

À dimension 384, le défaut de 8 donne un rappel@10 de 0,78. Le commentaire de `BinaryConfig` annonce 0,98 à 1,00 en dimension 768, d'où l'option `-binary-oversample`.

## Ce que Thalamus et Contaixt y gagnent

Contaixt passerait ses cerveaux à `binary` avec un suréchantillon de 32, après un contrôle de qualité sur le banc du pilote (`mesure_pilote_cortexdb.py`).

| Gain | ai-quant | ai-engineering (estimation) |
| :- | -: | -: |
| Instantané retiré du disque | 126 Mo | ≈ 430 Mo |
| Mémoire résidente | ≈ 250 Mo | ≈ 800 Mo |

Une ouverture ne lit plus l'instantané. Davantage de cerveaux tiendraient chauds sous la borne du VPS, et le rappel en sauterait moins faute de réponse dans son délai.

## Langues européennes

Rien de propre à une langue : l'index travaille sur les vecteurs de l'encodeur choisi.
