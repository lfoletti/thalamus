# 15. `find_nodes` : chercher parmi les noms, pas dans le texte des morceaux (deux PR)

**Statut** : fiche écrite le 30 octobre 2026, rien d'envoyé. **Taille estimée** :
- PR A : 25 lignes hors tests et deux tests ;
- PR B : 150 lignes hors tests, une table et ses écritures, et des tests de cohérence.

**Dépend de** : rien. La fiche 11, fusionnée dans v2.140.0, a déjà fait normaliser chaque nœud une fois par appel au lieu d'une fois par nom.

## Le besoin de Thalamus

`chercher` de kg-mcp se relaie par `th_find_nodes`. L'expertise du 29 octobre 2026 l'a mesuré à 658 ms sur ai-engineering, contre quelques millisecondes pour kg-mcp. Le moteur le dit lui-même : « The graph has no index on content » (`graphrag_tool_find_nodes.go`, v2.141.0). Chaque appel fait trois choses :

1. il relit le contenu de **tous** les nœuds qui en ont un, les morceaux compris ;
2. il normalise chacun de ces contenus (`foldName`) ;
3. il compare chaque nom demandé à chacun d'eux, par égalité, par égalité normalisée, puis par inclusion.

Les morceaux sont des passages de texte, pas des noms. Sur ai-security (v2.141.0, 1,48 Go), ils font presque tout le travail :

| Nœuds avec un contenu | Nombre | Texte |
| :- | -: | -: |
| Morceaux (`chunk`) | 111 929 | 43,6 Mo |
| Tous les autres (entités, documents, etc.) | 64 501 | 2,1 Mo |

Comme l'inclusion porte aussi sur le texte des morceaux, un nom très cité remplit la réponse de passages qui le mentionnent. Pour les 8 entités les plus citées d'ai-security, demandées avec `limit` 20, 48 des 160 places rendues sont des morceaux. Ils ne prennent pas la place d'une entité : une inclusion se classe par longueur, et un morceau, plus long qu'un nom, vient presque toujours après. Ils remplissent les places qu'aucun nœud nommé n'occupe, mais chaque appel paie pour les trouver. Le banc qui compare Thalamus à kg-mcp les écarte. Un agent qui cherche « noyb » attend l'entité, pas les passages : pour les passages, il y a la recherche (`knowledge_search`, `search_chunks_by_entities`).

## Mesure du 30 octobre 2026

La mesure porte sur ai-security, en reproduisant la boucle de `FindNodes` de v2.141.0 dans un programme Go, avec le pilote `modernc.org/sqlite` du moteur, sur Apple M2. Les noms sont les 8 entités les plus citées : « Schneier on Security », « Bruce Schneier », « Trail of Bits — blog », « Trail of Bits », « noyb — European Center for Digital Rights », « noyb », « Access Now », « Clive Robinson ».

| Lecture | Nœuds lus | Lecture et normalisation | Comparaison des 8 noms | Total | Morceaux dans les 20 premiers |
| :- | -: | -: | -: | -: | -: |
| Aujourd'hui (tous les nœuds) | 176 430 | 543 ms | 276 ms | 819 ms | 48 |
| Sans les morceaux (PR A) | 64 501 | 127 ms | 15 ms | 142 ms | 0 |
| Table des noms normalisés et son index (PR B) | — | — | 56 ms (égalité par l'index, inclusion par `instr`) | 56 ms | 0 |

La table des noms d'ai-security pèse 9,8 Mo avec son index, soit 0,7 % du cerveau.

## PR A : les morceaux restent hors de la recherche par nom, sauf sur demande

Règle : `find_nodes` cherche parmi les nœuds nommés. Le texte d'un morceau n'est comparé que si l'appelant le demande.

- Un argument `include_chunks`, booléen. La PR propose `false` par défaut, ce qui change le comportement par défaut. L'amont peut préférer `true` pour ne rien changer : Thalamus passerait alors `false` explicitement. C'est à demander dans la PR.
- Sans les morceaux, la requête ajoute `AND node_type != 'chunk'`. L'index `idx_nodes_type` existe déjà.
- `include_chunks` entre dans la liste blanche des arguments (`pkg/authz/allowlist.go`, `find_nodes` : `names`, `node_types`, `limit`), comme les autres. Le filtre des nœuds cachés ne change pas.

Pourquoi ce n'est pas `node_types` : sous une clé confinée, une entité est servie sans son type (`scopedNode`). Un filtre `node_types` écarte alors les entités après coup, et le banc tombait à 0 avec ce filtre sous une clé ordinaire. Exclure un type de la lecture ne dépend pas du type servi.

Preuves à apporter :

1. Un test où un nom cité dans un morceau rend l'entité sans le morceau, puis le morceau avec `include_chunks: true`.
2. Une mutation qui retire le filtre fait tomber ce test.
3. Les tests de portée existants (`find_nodes` dans la table des outils de graphe) restent verts, avec et sans l'argument.

## PR B : une table des noms normalisés, tenue à l'écriture

Règle : un nom se cherche par un index, sans normaliser les nœuds à la lecture.

- **La table.** `graph_node_names(node_id TEXT PRIMARY KEY, folded TEXT NOT NULL) WITHOUT ROWID`, avec un index sur `folded`. Elle couvre les nœuds hors morceaux qui ont un contenu.
- **Tenue à jour.** Chaque écriture de nœud la tient, dans la même transaction : création, mise à jour du contenu, suppression, fusion d'entités, retrait. Une base ouverte sans cette table la remplit une fois à l'ouverture, comme le moteur l'a fait pour d'autres migrations.
- **Égalité et égalité normalisée.** Une recherche par l'index sur `folded = foldName(nom)` donne les deux rangs. L'égalité exacte se déduit du contenu lu ensuite : un contenu égal au nom a la même forme normalisée.
- **Inclusion.** `instr(folded, ?) > 0` sur la table, sans normaliser à la lecture. Il n'y a pas d'index pour une inclusion, mais la table est courte.

Les morceaux, demandés par `include_chunks`, gardent la lecture d'aujourd'hui.

Points à vérifier :

- Chaque chemin qui écrit `graph_nodes.content` passe par un même point. Sinon, la PR doit le créer avant d'ajouter la table.
- La recherche CJK donne le même résultat : `foldName` garde les lettres CJK, et `instr` les compare octet par octet.

Preuves à apporter :

1. Un test de cohérence : après chaque sorte d'écriture, la table égale `foldName` de chaque contenu.
2. Les mêmes réponses que la PR A, avec la même empreinte du JSON, sur une copie d'ai-audio : les 306 noms de la fiche 11, par lots de 8.
3. Le temps mesuré sur une copie d'ai-security et d'ai-engineering.

## Ordre conseillé

Envoyer la PR A d'abord. Elle donne l'essentiel du gain, 819 → 142 ms (× 5,8), sans schéma ni écriture nouvelle, et elle rend des réponses plus propres. La PR B gagne encore × 2,5 (142 → 56 ms). En échange, une table doit rester juste dans tous les chemins d'écriture du moteur. Elle ne vaut d'être envoyée que si la PR A ne suffit pas au parcours d'un agent par Thalamus.

## Langues européennes

`foldName` retire la ponctuation et les espaces et met en minuscules, mais il garde les accents : « Données » et « Donnees » ne se rejoignent pas. C'est la règle actuelle, que ces deux PR ne changent pas. La fiche 07 traite des accents.
