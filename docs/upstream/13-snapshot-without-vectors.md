# 13. L'instantané HNSW sans copie des vecteurs, voisins en entiers

**Statut** : fiche écrite le 30 octobre 2026, rien d'envoyé. **Taille estimée** : 150 lignes hors tests dans `pkg/index/hnsw.go` et `pkg/core/store_index.go`, plus un test d'aller-retour et un test de lecture de l'ancien format. **Dépend de** : rien. **Ordre** : après la fiche 12. Si Contaixt passe à l'index binaire (fiche 12), cette fiche ne sert plus qu'aux bases qui gardent HNSW.

## Le besoin

`HNSW.Save` (`pkg/index/hnsw.go`) écrit en gob chaque `HNSWNode` entier : son identifiant, son vecteur float32, ses codes, son niveau, et ses voisins sous forme de chaînes (`Neighbors [][]string`). Décodage de l'instantané d'ai-quant (v2.141.0, 31 lignes de `index_snapshots`, 126,1 Mo) :

| Contenu | Octets |
| :- | -: |
| Vecteurs float32 de 40 457 nœuds | 62,1 Mo, la copie exacte de `embeddings.vector` (62,3 Mo) |
| 1 494 267 voisins, chacun l'identifiant du nœud en chaîne | 32,9 Mo |
| Identifiants des nœuds | 0,9 Mo |
| Encadrement gob (longueurs, champs, structures) | ≈ 30 Mo |

Deux redondances apparaissent :

- **Les vecteurs sont recopiés.** Ils sont déjà dans la table `embeddings`.
- **Chaque voisin répète en clair l'identifiant d'un nœud.** Un nœud en a 37 en moyenne.

Au chargement, gob alloue une chaîne par voisin : la mémoire porte la même répétition que le disque.

## La proposition (une PR, une règle)

Règle : l'instantané garde la structure du graphe, pas les données que la base a déjà.

- Un format v2, marqué par un en-tête de version :
  - la liste des identifiants une fois ;
  - pour chaque nœud, son niveau et ses voisins en positions `uint32` dans cette liste ;
  - les codes SQ8 ou binaires quand l'index est quantifié, puisqu'ils sont l'index lui-même ;
  - aucun vecteur float32.
- Au chargement, les vecteurs float32 se relisent de la table `embeddings` par identifiant, en flux, comme le fait déjà `reloadBinaryIndex`. Les voisins reprennent la chaîne du nœud et ne sont pas alloués une seconde fois.
- L'ancien format se lit toujours. Le prochain enregistrement l'écrit en v2.
- Un nœud de l'instantané absent de la table refuse l'instantané, et l'index se reconstruit. C'est la règle actuelle pour un instantané incohérent (« links to missing node »).

## Gain estimé (non mesuré, à prouver dans l'essai)

**Sur le disque, pour ai-quant :**

| Contenu | Taille |
| :- | -: |
| Voisins (1,49 M × 4 octets) | ≈ 6 Mo |
| Identifiants | ≈ 1 Mo |
| Niveaux et encadrement | quelques Mo |
| **Total, au lieu de 126 Mo** | **≈ 10 Mo** |

Pour ai-engineering, on passerait d'environ 430 Mo à environ 35 Mo.

**En mémoire :** une chaîne de moins par voisin, soit 1,49 M allocations de 16 octets d'en-tête plus le texte, environ 50 Mo sur ai-quant. Les vecteurs float32 restent en mémoire, comme aujourd'hui.

**Coût :** une lecture de la table `embeddings` à l'ouverture. Elle est mesurée à 0,3 s pour ai-quant par l'index binaire, qui fait la même lecture (fiche 12).

## Preuves à apporter

1. Test d'aller-retour : un index enregistré en v2 puis relu rend les mêmes résultats que l'index d'origine sur un jeu fixe.
2. Test de l'ancien format : un instantané v1 se lit encore.
3. Test d'incohérence : un identifiant absent de la table entraîne une reconstruction, pas un index à moitié chargé.
4. Mesure de la taille et de la mémoire sur une copie d'ai-quant, avant et après.

## Langues européennes

Rien de propre à une langue.
