# 07. La recherche lexicale pour les langues européennes

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : quatre PR, environ 150, 80, 200 et 150 lignes hors tests. **Dépend de** : rien dans le moteur. Les conditions de départ du `README.md` s'appliquent.

## Le besoin de Thalamus

Décision de Lauro du 22 octobre 2026 : Thalamus sert aussi une organisation suisse, avec des documents et des questions en français, en allemand, en italien et en romanche. Le produit se conçoit aussi pour des utilisateurs européens. Le moteur a un traitement dédié au chinois, au japonais et au coréen (`pkg/core/fts_cjk.go`, avec le jeu `pkg/eval/testdata/dataset_zh.json`). Les langues européennes n'ont pas encore cet équivalent.

## Constats du 22 octobre 2026 (v2.136.0)

Essai fait sur FTS5 (SQLite, `unicode61` par défaut) et sur PostgreSQL 'simple', avec les mêmes phrases :

| Requête | Texte stocké | SQLite (`chunks_fts`) | PostgreSQL ('simple') |
| :- | :- | :- | :- |
| `etat` | « Conseil d'État » | trouvé | **pas trouvé** |
| `universitat` | « Universität » | trouvé | **pas trouvé** |
| `Versicherung` | « Krankenversicherung » | **pas trouvé** | **pas trouvé** (index des mots) |
| `versicherung` | « Krankenversicherung » | trouvé par `chunks_fts_cjk` (trigrammes) | trouvé par `LIKE` (pg_trgm) |
| `décide*` | « deciso », « décision » | pas trouvé | pas trouvé |

Les faits du code :

- `unicode61` retire les accents par défaut (`remove_diacritics 1`). PostgreSQL 'simple' met les lettres en minuscules, mais garde les accents. Or le commentaire de `pkg/core/fts_postgres.go` dit que les deux moteurs doivent rendre les mêmes lignes pour la même requête (« a backend that quietly stemmed would return different rows »). Pour une langue avec des accents, ce n'est pas le cas aujourd'hui.
- `chunks_fts_cjk` et `messages_fts_cjk` contiennent **tous** les textes, pas seulement les textes CJK (déclencheurs `embeddings_cjk_*` dans `pkg/core/store_init.go`). Mais `CJKAwareIndex` les choisit seulement pour une requête qui contient une écriture CJK. Il y a 4 points d'appel : `pkg/core/advanced_search.go:112`, `pkg/core/chat.go:281`, `pkg/cortexdb/memory_api.go:1497` et `pkg/cortexdb/text_search.go:531`.
- `lexicalQueryStopwords` (`pkg/cortexdb/graphrag_tool_helpers.go:293`) et `entityStopwords` (`pkg/cortexdb/graphrag_helpers.go`) sont en anglais seulement.
- L'extraction heuristique des entités cherche les mots qui commencent par une majuscule. Le commentaire de `entityStopwords` parle de la majuscule en début de phrase en anglais. En allemand, chaque nom commence par une majuscule. [À mesurer : le nombre d'entités fausses sur un texte allemand]
- Le moteur ne fait pas de racinisation, et il n'en fait pas exprès : 'simple' plutôt que 'english', pour que les deux moteurs restent égaux.

## Faut-il un équivalent de `fts_cjk.go` ?

Oui, mais pas une copie. `fts_cjk.go` choisit l'index **selon l'écriture** de la requête, parce qu'une écriture CJK n'a pas d'espaces. Le français, l'allemand, l'italien et le romanche utilisent la même écriture latine que l'anglais. Le moteur ne peut donc pas choisir l'index par l'écriture, et une détection de la langue serait fragile sur une requête de 3 mots.

La règle équivalente se décide **selon le résultat**, terme par terme : un terme long que l'index des mots ne trouve pas est cherché une 2e fois dans l'index par trigrammes. Cette règle ne dépend d'aucune langue. Elle aide l'allemand et le romanche (mots composés), et aussi le néerlandais, les langues nordiques et le finnois. L'index existe déjà des deux côtés. Le fichier : `pkg/core/fts_compound.go`, à côté de `fts_cjk.go`.

### Constat du 22 octobre 2026, au soir : les variantes de la requête sont ignorées

La description des outils de recherche demande à l'agent des variantes dans d'autres langues (`alternate_queries`, `keywords`). Mais `searchTextCandidates` (`pkg/cortexdb/graphrag_tool_search.go`) s'arrête après la question : `if idx == 0 && len(merged) >= searchOpts.TopK { break }`. Une question en langue naturelle remplit presque toujours `top_k`, donc les variantes ne tournent pas. La partie vectorielle encode aussi la seule question.

Mesure sur le pilote ai-audio (50 questions en français, fiches surtout en anglais, `knowledge_search` en mode graphe avec encodeur) : avec ou sans une variante anglaise dans `alternate_queries`, les morceaux rendus sont identiques pour les 50 questions. La même variante, cherchée à part et fusionnée par RRF dans Thalamus (tranche 17), fait passer la couverture des faits de 81,3 % à 85,1 %.

PR possible (petite, avant les autres) : chaque variante tourne, et les listes sont fusionnées par rang, comme dans Thalamus. Le test : une variante qui seule trouve un document le fait entrer dans le résultat.

## La proposition

### PR 1 : un jeu d'évaluation suisse (environ 150 lignes, d'abord)

Comme `eval_zh_test.go` pour le chinois : `pkg/eval/testdata/dataset_ch.json` et `pkg/eval/eval_ch_test.go`. Environ 40 questions sur des documents dans les 4 langues, dont des questions croisées (question en français, document en allemand). Le test mesure le rappel lexical et donne la base de chaque PR suivante. Sans ce jeu, les PR 3 et 4 n'ont pas de mesure. Pour le romanche, sans personne qui parle cette langue (constat du 22 octobre 2026), le jeu utilise des textes officiels déjà traduits par des professionnels :

- le droit du canton des Grisons, publié en allemand, en italien et en romanche ;
- les textes fédéraux publiés aussi en romanche sur Fedlex (par exemple la Constitution fédérale).

Les documents en romanche sont donc ces textes officiels, alignés avec leur version allemande ou italienne. Les questions sont en français, en allemand ou en italien, et la réponse attendue est le passage en romanche. Les questions en romanche restent hors du jeu, jusqu'à la relecture par une personne qui parle cette langue, par exemple par la Lia Rumantscha.

### PR 2 : les accents égaux sur les deux moteurs (environ 80 lignes)

Règle : une requête sans accent trouve le texte avec accent sur PostgreSQL, comme sur SQLite.

- Voie A : l'extension `unaccent` et une configuration de recherche `cortexdb_simple` (`simple` + `unaccent`). C'est une 2e extension, et une instance gérée peut la refuser. La même règle « sans index, encore correct » que pour `pg_trgm` s'applique.
- Voie B : un `translate()` dans l'expression de l'index et de la requête, sans extension. Il est plus lourd à lire, et la table des lettres reste à tenir.
- Le test : les mêmes requêtes sur les deux moteurs rendent les mêmes lignes (`etat`, `universitat`, `perche` pour « perché »).
- Point à soumettre au propriétaire : il connaît la contrainte des instances gérées.

### PR 3 : le repli sur les trigrammes pour un terme non trouvé (environ 200 lignes)

Règle : si l'index des mots ne trouve aucune ligne pour un terme de 6 lettres ou plus, ce terme est cherché dans l'index par trigrammes (SQLite) ou par `LIKE` avec pg_trgm (PostgreSQL).

1. `pkg/core/fts_compound.go` : calcul pur. Il prend les termes de `ParseFTS5` et les termes que l'index des mots ne trouve pas. Il rend la requête de repli. Le test se fait sans serveur.
2. Les 4 points d'appel de `CJKAwareIndex` passent par une seule fonction. La requête CJK ne change pas.
3. Le rang : une ligne trouvée par repli vient après les lignes trouvées par l'index des mots. Le BM25 sur trigrammes ne se compare pas au BM25 sur mots (le commentaire de `fts_cjk.go` le dit déjà pour l'anglais).
4. Le seuil de 6 lettres se choisit avec le jeu de la PR 1. Un seuil trop bas trouve des parties de mots sans rapport (« Rat » dans « Ratgeber »).

### PR 4 : les mots vides en français, en allemand, en italien et en romanche (environ 150 lignes, après la mesure)

1. `lexicalQueryStopwords` reçoit les mots vides courts des 4 langues (« le », « la », « der », « die », « il », « di », « da »…). La liste reste courte et fermée, comme la liste anglaise.
2. L'extraction heuristique des entités : un nom allemand avec une majuscule n'est pas une entité seulement à cause de sa majuscule. Ce point se décide avec le propriétaire, parce qu'il touche la règle de `entityStopwords`.
3. La mesure de la PR 1 dit si ce point est nécessaire. Sinon, la PR ne part pas.

## Ce que le moteur ne reçoit pas

- **La racinisation.** Le propriétaire a choisi 'simple' pour que les deux moteurs restent égaux. Une racinisation Snowball égale sur FTS5 et PostgreSQL demande un tokenizer FTS5 en C. Je ne la propose pas.
- **La recherche d'une langue à une autre.** Elle vient de l'embedder (`CORTEXDB_EMBED_*`, par exemple `bge-m3`), sans changement du moteur.

## Ce qui reste dans Thalamus

- Le champ `language` d'un document : `th_add` le pose, et l'extension le lit dans l'attribut `lang` de la page.
- `th_recall` relaie `keywords` et `alternate_queries`. La compétence du plugin demande à l'agent la question dans les 4 langues.
- Une entité à 4 noms (« Conseil d'État », « Staatsrat », « Consiglio di Stato », « Cussegl guvernativ ») : un identifiant et un alias par langue, dans l'ontologie du cerveau. C'est un choix de l'organisation, pas un mécanisme.

## Avant l'envoi

- Rejouer le tableau des constats sur la version de l'amont du jour.
- Lire `pkg/core/advanced_search.go` et `pkg/agentmem/search.go` : `agentmem` a sa propre table avec `remove_diacritics 2` (`pkg/agentmem/schema.go:136`), donc il se comporte peut-être déjà autrement.
- Vérifier la licence des textes officiels repris dans le jeu (droit des Grisons, Fedlex).
