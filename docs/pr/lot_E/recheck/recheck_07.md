# Fiche 07 rejouée sur CortexDB v2.138.0 (23 octobre 2026)

Base : `cortexdb/` à `b2a1ab6` (v2.138.0). Copie de travail jetable `scratchpad/wt07` (supprimée à la fin). Tests jetables gardés dans `scratchpad/recheck07_tests/` (`zz_recheck07_test.go`, `zz_recheck07_ch_test.go`, paquet `cortexdb`, à poser dans `pkg/cortexdb/`). SQLite seul, sans `CORTEXDB_TEST_POSTGRES`, sans encodeur. Commande de rejeu :

```
env -u CORTEXDB_TEST_POSTGRES go test ./pkg/cortexdb/ -run 'TestRecheck07' -v -count=1
```

Les lignes ci-dessous sont celles de v2.138.0. Ce qui est **mesuré** porte la mention « mesuré ». Ce qui vient de la **lecture du code** porte la mention « code ». Les **suppositions** sont marquées « supposition ».

## 0. Ce que change 7e0aa86 (le fait nouveau)

Code, `git show 7e0aa86` (16 fichiers, +1121 −70) :

- `pkg/cortexdb/lexical_rarity.go` (nouveau, 462 lignes). Il calcule un score de couverture : `Σ idf(termes de la question que la ligne contient) / Σ idf(termes de la question)`. Les mots-clés, les variantes et les noms d'entités comptent comme synonymes, à la moitié de leur poids (`rarityExtraWeight = 0.5`, l. 49). Ils n'agrandissent pas le total. Le df vient d'un seul balayage `LIKE` (`ILIKE` sur PostgreSQL) de l'ensemble cherché (`scanCorpusStats`, l. 396). 100 lignes fictives sont ajoutées au dénominateur (`rarityPriorRows`).
- `containsWord` (l. 345) compare un terme mis en minuscules (`strings.ToLower`, l. 265) au mot entier, ou au début d'un mot pour un terme de 4 lettres ou plus, et jamais à l'intérieur d'un mot. **Il ne retire pas les accents.**
- `pkg/cortexdb/relevance_floors.go` (nouveau). Seuils : `DefaultKnowledgeKeywordFloor = 0.25` (l. 48), `DefaultKnowledgeVectorFloor = 0.40` (l. 42) et `DefaultMemoryKeywordFloor = 0.25` (l. 32). `applyKnowledgeFloors` (l. 91) retire un morceau qui n'atteint ni le seuil vectoriel, ni le seuil de couverture avec plus d'un mot de la question (`rarityScorer.answers`, l. 325), ni un lien vers une entité nommée (mode graphe). Ce filtre est appelé par `SearchKnowledge` (`knowledge_api.go:351`) hors du mode PPR.
- La mémoire (`memory_api.go:726-848`) note chaque candidat par la même couverture et applique `memoryKeywordFloor` (l. 838-843).
- Options : `WithKnowledgeFloors`, `WithMemoryKeywordFloor`, variables `CORTEXDB_*_FLOOR`.

**Conséquence (mesuré)** : sur SQLite, le seuil tient compte des accents, alors que FTS5 n'en tient pas compte. Une ligne que FTS5 trouve sans accent est ensuite retirée par le seuil :

| Requête (sans encodeur) | `SearchText` (outil) | `SearchKnowledge` (seuils par défaut) | `SearchKnowledge`, `WithKnowledgeFloors(0,0)` |
| :- | :- | :- | :- |
| `universitat` (texte « Universität ») | trouvé | **vide** | — |
| `universitat zurich recht` | trouvé | **vide** | trouvé |
| `universität zürich recht` | trouvé | trouvé | — |
| `conseil etat pouvoir` | trouvé | trouvé (2 autres mots couvrent) | trouvé |

`containsWord("conseil d'état", "etat")` rend `false` (mesuré).

**Sur le corpus du banc suisse** (mesuré, `TestRecheck07SwissMonolingual`). Corpus : les 232 articles × 4 langues, soit 928 documents. Requêtes : les 40 questions et leurs 80 variantes, chacune posée contre l'article dans sa propre langue. Mesure en lexical seul, `top_k` 10 :

| | n | @5 | @10 |
| :- | -: | -: | -: |
| avec accents, seuils par défaut | 120 | 0,97 | 0,97 |
| sans accents (NFD, marques retirées), seuils par défaut | 78 | **0,83** | **0,83** |
| avec accents, seuils coupés | 120 | 0,97 | 0,98 |
| sans accents, seuils coupés | 78 | 0,96 | 0,97 |

Une requête sans accent perd donc 13 points à cause du seuil de v2.138.0. Ce défaut n'existait pas en v2.136.0, qui n'avait pas de seuil. La mémoire passe par le même `containsWord`, donc le même défaut est attendu en rappel de mémoire (code, non mesuré).

Le rejeu du banc dans le README (« mêmes chiffres à une question près ») ne le montre pas : ses questions gardent leurs accents.

## 1. Les prémisses de la fiche, une par une

### 1.1 Les accents sur SQLite

- **Confirmé (code et mesuré).** `chunks_fts` et `messages_fts` utilisent le tokenizer par défaut `unicode61` (`pkg/core/store_init.go:195`, `:231`), qui retire les accents. Mesuré : `MATCH "etat"` trouve « État », et `"universitat"` trouve « Universität ».
- **Nouveau (mesuré)** : l'index par trigrammes `chunks_fts_cjk` (`store_init.go:233`, `tokenize='trigram'`) **garde** les accents. `MATCH "universitat"` y rend 0 ligne et `"etat"` 0 ligne. Un repli sur les trigrammes devra traiter les accents lui-même (`remove_diacritics 1` sur le tokenizer trigram, possible à partir de SQLite 3.45 ; modernc v1.38.2 devrait le permettre ; supposition, à vérifier).
- **Nouveau (code et mesuré)** : le seuil de 7e0aa86 tient compte des accents (section 0). SQLite n'est donc plus cohérent avec lui-même.
- `agentmem` : `pkg/agentmem/schema.go:126` (trigram) et `:136` (`unicode61 remove_diacritics 2`) (code ; la fiche est confirmée).

### 1.2 Les accents sur PostgreSQL

- **Confirmé (code, non exécuté).** Toutes les expressions utilisent `to_tsvector('simple', …)` et `plainto_tsquery('simple', …)` : index `pkg/core/fts_postgres.go:55`, conditions `:125` et `:134`, rang `:264`, `store_postgres_search.go:239`. La configuration 'simple' met en minuscules et garde les accents (comportement connu de PostgreSQL ; pas exécuté ici).
- Le commentaire `fts_postgres.go:23-25` demande toujours que les deux moteurs rendent les mêmes lignes.
- **Nouveau (code)** : `scanCorpusStats` compte le df avec `ILIKE` sur PostgreSQL (`lexical_rarity.go:396`). `ILIKE` tient compte des accents, comme le `LIKE` de SQLite. Le df d'un terme écrit sans accent vaut donc 0 sur les deux moteurs.

### 1.3 L'index par trigrammes et son usage

- **Confirmé (code)** : les déclencheurs `embeddings_cjk_*` et `messages_cjk_*` envoient **tous** les textes dans les index `_cjk` (`store_init.go:202-214`, `:233-…`).
- **Confirmé (code)** : `CJKAwareIndex` (`fts_cjk.go:41`) choisit `_cjk` seulement si la requête contient du CJK. Les quatre points d'appel ont changé de ligne : `pkg/core/advanced_search.go:112`, `pkg/core/chat.go:281`, `pkg/cortexdb/memory_api.go:1575` (1497 dans la fiche) et `pkg/cortexdb/text_search.go:540` (531).
- PostgreSQL (code) : l'index `gin_trgm_ops` est créé sur chaque colonne (`fts_postgres.go:59-61`). La condition `LIKE` passe seulement par le bras CJK (`fts_postgres.go:86-111`). Le bras non CJK utilise `tsquery`. La fiche dit « `versicherung` trouvé par `LIKE` (pg_trgm) ». Ce résultat vient d'un `LIKE` écrit à la main : le chemin du moteur ne le trouve pas (code ; supposition sur la manière dont la mesure du 22 octobre a été faite).
- **Confirmé (mesuré)** : `versicherung` donne 0 ligne dans `chunks_fts` et 1 dans `chunks_fts_cjk`. `SearchText("versicherung")` et `SearchKnowledge` rendent [].
- **Nouveau (code et mesuré)** : `containsWord("die krankenversicherung", "versicherung")` rend `false`. Même si le repli trouve la ligne, le seuil la retire : une requête d'un seul mot a une couverture de 0. Le repli sur les trigrammes demande donc aussi de changer la règle de couverture.

### 1.4 Les mots vides

- **Confirmé (code)** : `lexicalQueryStopwords` est en anglais seulement (`graphrag_tool_helpers.go:300` ; 293 dans la fiche). La même liste sert à `extract_conversation.go:81`. `entityStopwords` est aussi en anglais seulement (`graphrag_helpers.go:49`, remplie dans `init`). La fiche oublie une 2e liste anglaise, `pkg/graphflow/organize_stopwords.go:8`.
- **Nouveau (code et mesuré)** : depuis 7e0aa86, un mot fréquent pèse peu par l'idf, sans liste. Mesuré (`TestRecheck07FloorOnVariants`) avec 20 documents français qui contiennent « qui », « la », « des » et « règle » : couverture 0,365 pour un document allemand trouvé par la variante. Sans ces documents, le corpus de 2 lignes et les 100 lignes fictives donnent le poids maximal à chaque mot, et la même ligne passe sous le seuil (mesuré, `TestRecheck07AlternateIsAnd` : `SearchText` la trouve, `SearchKnowledge` rend []). Les mots vides de la PR 4 n'ont donc plus de rôle dans la couverture une fois le corpus de taille réelle. Ils en gardent un dans la requête OU des mots-clés (« la OR des »). Le BM25 y donne déjà peu de poids à un mot fréquent (code ; l'effet sur le rang n'est pas mesuré).

### 1.5 La sortie anticipée sur les variantes

- **Confirmé (code et mesuré), à deux endroits** : `pkg/cortexdb/graphrag_tool_search.go:94` (celui de la fiche) et **`pkg/cortexdb/text_search.go:465`** (`searchTextOnlyExpanded`, atteint par `SearchTextOnly` quand on lui passe `Keywords` ou `AlternateQueries`, donc par le préchargement lexical de `cortex_query`, `query_api.go:285`).
- Mesuré (`TestRecheck07EarlyExit`) : 6 documents français contiennent tous les mots de « conseil procédure cantons », et un document allemand ne répond qu'à la variante « Bundesrat Verfahren Kantone ». Avec `top_k` 3, la variante est ignorée : `SearchText` et `SearchKnowledge` rendent [fr0 fr1 fr2]. Avec `top_k` 10, le document allemand arrive en 5e position seulement : chaque liste garde son BM25 brut fois 0,95, et la fusion prend le maximum, sans fusion par rang.
- **Nuance importante (code et mesuré)** : la question passe par `sanitizeFTSQuery` (l. 180), qui met chaque mot entre guillemets. FTS5 lit donc un **ET** de tous les mots, mots vides compris. Une question en langue naturelle remplit rarement `top_k` à elle seule, et la sortie anticipée se déclenche surtout pour les questions courtes faites de mots-clés.
- **Défaut plus probable que la sortie anticipée (code et mesuré)** : chaque variante tourne aussi comme un ET de tous ses mots (`graphrag_tool_helpers.go:226`). Ses mots n'entrent jamais dans la requête OU des mots-clés : `autoKeywords` vient de la seule question (l. 230), et `plannedKeywords` des seuls `keywords`. Mesuré (`TestRecheck07AlternateIsAnd`) : une variante « Wer regelt das Verfahren der Kantone? » rend [] même dans `SearchText`, et une variante courte « Bundesrat Verfahren Kantone » trouve le document. Supposition : c'est la cause du constat ai-audio (« morceaux identiques avec ou sans variante »), plutôt que la sortie anticipée.
- La mémoire n'a pas de sortie anticipée, ni en v2.136.0 ni en v2.138.0 (`memory_api.go:732-800`). Toutes ses requêtes donnent des candidats, notés ensuite par la couverture. C'est le modèle à suivre pour le savoir. Le défaut du ET sur les variantes la touche aussi (même `lexicalSearchQueries`).

### 1.6 Les jeux d'évaluation existants

Mesuré : `TestLexicalRetrievalQuality` et `TestChineseLexicalRetrievalQuality` passent sur v2.138.0 (MRR 1,000 chacun).

Format des jeux (code, `pkg/eval/dataset.go`) : `{name, description, documents[{id,title,content,group?}], queries[{id,text,relevant[],category?}], negatives[{id,text}]}`, embarqué par `//go:embed`. Taille : `dataset.json` fait 11 Ko (44 documents de 112 caractères en moyenne, 29 requêtes) ; `dataset_zh.json` fait 10 Ko (29 documents de 47 caractères, 22 requêtes, 6 négatives). Seuils : `eval_cortexdb_test.go:72,75` (recall@10 0,85, MRR 0,80) ; `eval_zh_test.go`, `floorZhRecall5 = 0.90` et `floorZhMRR = 0.85`. Le format n'a pas de champ pour les variantes.

## 2. Ordre des PR proposé

Cette proposition remplace l'ordre de la fiche. Les deux défauts des variantes et le défaut des accents dans le seuil sont petits, se mesurent sur SQLite sans serveur et ne dépendent d'aucune langue. Ils passent avant PostgreSQL et les trigrammes. Une règle par PR, tests écrits d'abord. Les nombres de lignes sont estimés hors tests (supposition).

### PR 1 : le jeu suisse `dataset_ch.json` (environ 10 lignes de code, plus les données et le test)

- **Règle** : aucune règle de code. C'est la base qui mesure les PR 2 à 6.
- **Fichiers** : `pkg/eval/testdata/dataset_ch.json`, `pkg/eval/dataset.go` (`//go:embed` + `BuiltinSwiss()`, environ 8 lignes), `pkg/eval/eval_ch_test.go` (copie de `eval_zh_test.go`, environ 100 lignes de test).
- **Contenu, au format existant** :
  - Documents : 24 articles courts (moins de 400 caractères ; la médiane du corpus est de 350 à 380) × 4 langues, soit 96 documents. `id` = `art_<n>__<langue>`, `group` = `art_<n>` (le champ `Group` existe déjà). Taille environ 35 à 40 Ko, plus que les 10 Ko des deux jeux actuels.
  - Requêtes : 36 requêtes dans la langue du document (12 en de, 12 en fr, 12 en it), tirées de `questions.json` et de ses variantes `v`. La moitié porte au moins un mot accentué écrit sans accent (`category: "unaccented"`), quelques-unes un mot simple contenu dans un mot composé (`category: "compound"`). `relevant` = l'article dans la langue de la requête.
  - Négatives : 4, sans réponse dans le corpus.
  - Pas de question en romanche. Les documents romanches servent de distracteurs.
  - Les questions d'une langue à l'autre restent hors de ce jeu : sans encodeur, elles demandent des variantes, et `eval.Query` n'a pas de champ pour elles. Ce champ vient avec la PR 3.
- **Seuils** : un peu sous la mesure du jour, et par catégorie. Mesure du jour sur le corpus entier (section 0) : 0,97 avec accents, 0,83 sans accent. Le test enregistre aussi `unaccented` à part, pour que la PR 2 le relève.
- **Source à noter dans `description`** : Constitution fédérale de la Confédération suisse, RS 101, état du 3 mars 2024, Fedlex (fedlex.admin.ch), versions officielles de, fr, it et rm. Extraction par `tests/eval_ch/extract.py` de Thalamus. Licence non vérifiée (supposition : les textes officiels sont exclus du droit d'auteur par l'art. 5 LDA).
- **Questions** : écrites par le modèle (Claude), pas par des locuteurs. À dire dans `description`.
- **Failing first** : sans objet, aucune règle. Les seuils se vérifient en coupant le seuil de couverture : `WithKnowledgeFloors(0,0)` ne doit pas faire tomber le test, et une mutation qui rend `containsWord` toujours faux doit le faire tomber.

### PR 2 : le seuil de couverture ne tient pas compte des accents (environ 40 à 70 lignes)

- **Règle** : un terme de la question couvre une ligne si leurs formes **sans accents** se correspondent, comme pour FTS5 `unicode61`. C'est une correction de 7e0aa86, sans liste de langues.
- **Fichiers** : `pkg/cortexdb/lexical_rarity.go`. Une fonction `foldDiacritics` (NFD puis retrait des marques `Mn`) s'applique à `rarityWords` (l. 103), à `newRarityQuery` (l. 75) et au contenu dans `coverage` (l. 265). `containsWord` reste inchangé et reçoit les formes sans accents.
- **Dépendance** : `golang.org/x/text` est déjà dans `go.mod` comme dépendance indirecte (`v0.36.0`). La PR la rend directe : point à soumettre au propriétaire. L'autre voie est une table de lettres latines écrite à la main, plus longue mais sans dépendance.
- **Le df** : `scanCorpusStats` compte avec `LIKE`, qui tient compte des accents. Le df d'un terme écrit sans accent reste 0, donc ce terme a l'idf maximal. Le plafond `min(1, …)` borne la couverture, et un idf trop fort rend la règle plus dure, pas plus lâche. Je propose de le laisser tel quel et de le dire dans le commentaire. Un df sans accents demanderait de calculer en Go ou un `unaccent`, donc la PR 5.
- **Tests à écrire d'abord** :
  - `TestKnowledgeKeywordFloorMatchesUnaccentedQuery` : « universitat » trouve « Universität » par `SearchKnowledge` avec les seuils par défaut (aujourd'hui [], mesuré).
  - `TestKnowledgeKeywordFloorMatchesAccentedQueryOnUnaccentedText` : le sens inverse, « État » contre « Etat ».
  - `TestMemoryKeywordFloorMatchesUnaccentedQuery` : même chose en rappel de mémoire.
  - `TestRarityCoverageFoldsDiacritics` : test unitaire de `coverage`.
  - Le seuil `unaccented` de `eval_ch_test.go` relevé (de 0,8 environ à 0,95 environ). Il tombe sans la PR.
- **Mutations à la main** :
  - `foldDiacritics` rend son argument : les 4 tests tombent.
  - Pli sur la question seulement, pas sur le contenu : le test du sens inverse tombe.
  - Pli sur le contenu seulement : le test « universitat » tombe… à vérifier, car une requête sans accent contre un contenu sans accent passe déjà.
  - Pas de pli dans `newRarityQuery` pour les synonymes : il faut un test avec un mot-clé sans accent (`TestKnowledgeKeywordFloorFoldsSynonyms`).

### PR 3 : les mots des variantes entrent dans la requête OU (environ 10 à 20 lignes)

- **Règle** : les mots d'une variante (`alternate_queries`) entrent dans la requête OU des mots-clés, comme ceux des `keywords`. Aujourd'hui, une variante en langue naturelle n'est cherchée que comme un ET de tous ses mots.
- **Fichier** : `pkg/cortexdb/graphrag_tool_helpers.go:200-243`, `lexicalSearchQueries` : `plannedKeywords` prend `keywords` **et** `alternateQueries`. Le nombre de termes est borné, comme `rarityMaxWords`. Le changement touche aussi la mémoire (même fonction).
- **Optionnel** : le champ `alternates` dans `eval.Query` (`dataset.go`, 3 lignes) et les questions d'une langue à l'autre dans `dataset_ch.json`, accompagnées de leurs variantes.
- **Tests à écrire d'abord** :
  - `TestKnowledgeSearchFindsDocumentThroughAlternateSentence` : la question française, la variante « Wie regelt der Bund das Verfahren der Kantone? », le document allemand dans le résultat (aujourd'hui [], mesuré).
  - `TestMemorySearchFindsMemoryThroughAlternateSentence`.
  - `TestLexicalSearchQueriesTakeAlternateWords` : test pur de la liste des requêtes.
- **Mutation à la main** : retirer `alternateQueries` de la fusion des mots-clés. Les trois tests tombent.
- **Interaction avec le seuil** : les mots de la variante comptent déjà à demi-poids (`relevance_floors.go:107`). Mesuré avec un corpus réaliste : 0,365, au-dessus de 0,25.

### PR 4 : chaque variante tourne même si la question remplit `top_k` (environ 5 lignes)

- **Règle** : la fiche décrivait cette règle. Les deux `break` sortent : `graphrag_tool_search.go:94` et `text_search.go:465`.
- **Tests à écrire d'abord** :
  - `TestSearchTextRunsAlternatesWhenQuestionFillsTopK` : le scénario `TestRecheck07EarlyExit` avec `top_k` 3. Le document allemand doit être dans le résultat ; aujourd'hui il ne l'est pas (mesuré).
  - `TestSearchTextOnlyRunsAlternatesWhenQuestionFillsTopK` : le même scénario par `SearchTextOnly` avec `AlternateQueries`, pour le second `break`.
- **Mutations à la main** : remettre chaque `break`. Chacune fait tomber son test.
- **Attention** : sans les `break`, la fusion garde le maximum des BM25 bruts, multipliés par 0,95 à 0,8. Une variante qui ne trouve qu'elle-même peut rester derrière les `top_k` lignes de la question. Le test doit donc choisir un `top_k` où la ligne entre ; avec le scénario mesuré, il vaut mieux vérifier sa présence que son rang. La fusion par rang (RRF, comme dans Thalamus) est **une autre règle**, donc une PR 4 bis éventuelle, à mesurer d'abord sur `dataset_ch` (environ 40 lignes ; supposition).

### PR 5 : les accents sur PostgreSQL (environ 80 lignes, après accord du propriétaire)

- **Règle** : identique à celle de la fiche. Voie A (`unaccent` et une configuration `cortexdb_simple`) ou voie B (`translate()`).
- **Fichiers** : `fts_postgres.go` (l. 55, 125, 134, 264), `store_postgres_search.go:239`, et `ILIKE` dans `lexical_rarity.go:396` pour le df (`unaccent(col) ILIKE unaccent(?)`).
- **Tests** :
  - `TestPostgresLexicalConditionFoldsAccents` : forme du SQL, sans serveur.
  - `TestPostgresFindsUnaccentedQuery` : un test réel sous `CORTEXDB_TEST_POSTGRES`, en CI seulement.
- **Mutation** : la condition sans `unaccent`.
- **Ordre** : cette PR dépend de la décision du propriétaire sur les instances gérées. Elle n'est pas mesurable ici sans PostgreSQL.

### PR 6 : le repli sur les trigrammes pour un mot composé (environ 200 lignes, plus la règle de couverture)

- **Règle** : identique à celle de la fiche. Il y a deux nouveaux faits, mesurés :
  - l'index trigram garde les accents ;
  - le seuil de couverture refuse un terme à l'intérieur d'un mot (`containsWord`, « never inside one »).
- **Conséquence** : il faut d'abord une PR 6a pour la couverture : un terme de 6 lettres ou plus couvre une ligne qui le contient à l'intérieur d'un mot. Cette règle contredit un choix explicite de 7e0aa86, donc il faut l'accord du propriétaire. Environ 10 lignes, test `TestRarityCoverageCountsLongTermInsideWord`, mutation : remettre la règle du début de mot.
- **Ensuite** : PR 6b pour le repli lui-même (`pkg/core/fts_compound.go`, les 4 points d'appel, `remove_diacritics 1` sur `_cjk` ou un index trigram à part ; changer le tokenizer d'une table existante demande une migration). Le seuil de 6 lettres se mesure sur la catégorie `compound` de `dataset_ch`.

### Abandonnée : les mots vides de, fr, it et rm

Code et mesuré : depuis 7e0aa86, la rareté donne déjà peu de poids aux mots fréquents, et ce mécanisme ne dépend d'aucune langue. C'est la préférence du propriétaire. Je propose de ne pas envoyer la liste. Le point qui reste est celui des noms allemands pris pour des entités à cause de leur majuscule (`graphrag_helpers.go:49`, et la liste de `graphflow/organize_stopwords.go:8`). Il faut d'abord le mesurer sur les documents allemands de `dataset_ch`. Si le défaut se confirme, une règle sans liste reste à trouver, par exemple par la fréquence du mot en minuscules dans le corpus (supposition).

## 3. Ordre résumé

1. PR 1 : le jeu `dataset_ch` (données ; environ 10 lignes de code).
2. PR 2 : les accents dans le seuil de couverture (40 à 70 lignes). C'est un défaut de v2.138.0 : 0,83 contre 0,96 sur le corpus suisse.
3. PR 3 : les mots des variantes dans la requête OU (10 à 20 lignes).
4. PR 4 : les deux `break` sortent (environ 5 lignes). Une PR 4 bis pour la fusion par rang est possible après mesure.
5. PR 5 : `unaccent` sur PostgreSQL (environ 80 lignes, décision du propriétaire).
6. PR 6a puis 6b : la couverture à l'intérieur d'un mot, puis le repli sur les trigrammes (10 lignes puis environ 200, décision du propriétaire pour la 6a).
7. Mots vides : non envoyés. Les entités allemandes : à mesurer d'abord.

Les PR 2, 3 et 4 se rejouent sur SQLite seul. La PR 1 sert de mesure à toutes.

## 4. Ce qui reste à vérifier

- Si le tokenizer trigram de modernc v1.38.2 accepte `remove_diacritics 1` (supposition).
- Le comportement réel de PostgreSQL : non exécuté ici, conformément à la consigne.
- Si la PR 3 explique à elle seule le constat ai-audio (supposition) : à rejouer sur le pilote après la PR 3, avec un moteur local.
- La licence de Fedlex (art. 5 LDA, supposition) : hors du périmètre demandé.
