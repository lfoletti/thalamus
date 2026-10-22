# Note : la pyramide de synthèses et Thalamus

22 octobre 2026. Question de Lauro : la mécanique pyramidale de contaixt (`~/Agents/science-extractor`) doit-elle entrer dans CortexDB et dans ses fonctions `_recall`, ou rester en amont ou en aval ?

Réponse : ni en bloc dans le moteur, ni en dehors. La pyramide a 4 parties, et chaque partie a sa place.

## 1. La pyramide de contaixt

Source : `CHANTIER_COUT.md` (5 octobre 2026), `CHANTIER_CARTES.md`, `CHANTIER_CONTEXTE.md`, `scripts/synthese.py`.

| Niveau | Unité | Taille | Cite |
| :- | :- | :- | :- |
| 0 | le graphe (la carte) | environ 1 page | ses thèmes |
| 1 | le thème (une communauté du graphe) | 200 à 400 mots | ses notions |
| 2 | la notion | 150 à 300 mots | les lignes des fiches |
| 3 | le passage | quelques lignes | la fiche |
| 4 | la fiche | entière | l'adresse d'origine |

Les règles :

- Une question s'arrête au niveau le plus haut qui y répond. Les niveaux 0 à 2 orientent, les niveaux 3 et 4 prouvent.
- La construction se fait du bas vers le haut. Un modèle (`claude -p`) écrit les synthèses.
- Un contrôle mécanique, sans modèle, retire chaque phrase dont un nombre, une date ou un nom ne se lit pas dans les preuves citées.
- Une empreinte des preuves fait réécrire une synthèse quand ses preuves changent, en cascade (fiche, notion, thème).
- Depuis le 9 octobre 2026, des cartes par sujet déclaré ajoutent un étage entre le thème et la notion.

## 2. Ce que la mesure dit

Pilote du 22 octobre 2026 (`CHANTIER_PILOTE_CORTEXDB.md`, ai-audio, 41 questions avec des faits discriminants, sans lecteur, à budget égal) :

| Bras | Couverture des faits |
| :- | :- |
| contaixt, synthèses + passages | 76,4 % |
| contaixt, passages seuls | 50,5 % |
| CortexDB, mode graphe, sans pyramide | 74,8 % |
| CortexDB, mode graphe + pyramide servie d'abord | 78,5 % |
| **CortexDB, vecteurs, synthèses en tête puis morceaux** (remesure du soir) | **81,2 %** |
| CortexDB, vecteurs, synthèses et fiches classées par CortexDB lui-même | 75,5 % |
| CortexDB, sans vecteurs, synthèses en tête | 77,6 % |

La remesure du soir vient de l'autre session (contaixt, 22 octobre 2026), avec 1 837 synthèses d'ai-audio versées et l'encodeur de contaixt servi à CortexDB. Ce qu'elle établit :

- Les synthèses apportent l'essentiel : 19 points de plus que les fiches seules. Les vecteurs ajoutent 3 à 7 points.
- **L'ordre compte** : les synthèses en tête, placées par le client, battent le classement que CortexDB fait lui-même (81,2 contre 75,5). C'est donc au client de composer le contexte. Dans Thalamus, `th_recall` le fait.
- Le gain est surtout sur les questions récentes (16 points). Les questions de fond perdent 5,5 points.

Source : `cache/pilote_cortexdb/mesure/ai-audio_bilan_2026-10-22T21-01.json` de contaixt. La pyramide est stockée dans CortexDB comme des documents d'une collection `pyramide`, avec leurs preuves en métadonnée (`puits_cortexdb.synthese_vers_cortexdb`). Le client les place avant les passages.

Conclusion : la pyramide garde sa valeur quand CortexDB la stocke et la cherche. Le moteur n'a pas besoin de la construire.

### Mesure à travers le gateway (tranche 15, 22 octobre 2026, 21 h 30)

Copie du cerveau `ai-audio-vecteurs.db` du pilote, avec son encodeur (384 dimensions) et son binaire `cortexdb-grpc` v2.136.0. Un gateway devant la copie, un cerveau `lab/ai-audio` avec `pyramid: {collection: pyramide, sources: preuves}`, une clé confinée aux 2 collections. Mêmes 41 questions, même règle de couverture (`mesure_pilote_cortexdb.couverture`), même budget par question (4 044 caractères en moyenne). Le pilote n'a pas changé.

| Bras | Couverture |
| :- | :- |
| `th_recall`, 3 résultats, extraits (ce que le hook injecte) | 51,7 % |
| `th_recall`, 10 résultats, extraits | 54,4 % |
| `th_recall`, 10 résultats, documents entiers par `th_get` | 67,2 % |
| idem, `syntheses: 0` | 37,8 % |
| relais `/mcp/lab/ai-audio`, `search_graphrag_lexical` en mode graphe, synthèses en tête (le bras du pilote) | **81,2 %** |
| idem, passages seuls | 62,0 % |

Ce que la mesure établit :

- Le chiffre du pilote se retrouve à travers le gateway, sous la clé confinée d'un employé : 81,2 %. La clé à 2 collections et la liste blanche ne coûtent rien.
- Dans `th_recall`, la pyramide ajoute 29 points (37,8 % → 67,2 %).
- Mais `th_recall` perd 14 points sur le bras du pilote. La cause est l'outil de recherche : `th_recall` appelle `knowledge_search`, qui classe des documents et rend un extrait coupé à environ 180 caractères. Un document entier remplit le budget avec son début, pas avec le passage utile. `search_graphrag_lexical` rend les morceaux utiles. Passages seuls : 37,8 % contre 62,0 %.
- Le hook est le cas le plus faible (51,7 %), parce qu'il injecte 3 extraits.
- Temps moyen de `th_recall` : 110 ms par question.

Suite proposée : `th_recall` demande les morceaux par `search_graphrag_lexical` (mode graphe, `max_context_chars`), garde la fusion RRF et la provenance par document, et le hook injecte le texte des morceaux dans son budget.

Défaut vu pendant la mesure : `th-gateway serve` s'arrête proprement sur SIGINT seulement (`cmd/th-gateway/main.go`). Sur SIGTERM, l'audit n'est pas fermé, et les lectures comptées du jour sont perdues (0 ligne après environ 200 lectures).

Scripts et résultats : `scratchpad/m15/` de la session (`measure.py`, `probe2.py`, `score.json`), hors dépôt.

## 3. Les 4 parties et leur place

| Partie | Place | Raison |
| :- | :- | :- |
| **Construction** : écriture par un modèle, choix des notions, thèmes, cartes, cascade | **Hors du moteur**, dans la couche de veille (contaixt, ou un autre producteur de l'organisation). Elle lit ses preuves dans CortexDB et écrit ses synthèses dans CortexDB. L'autre session dit « en aval de CortexDB » pour ce sens. | Elle appelle un modèle et coûte de l'argent. Le choix des notions et des sujets est éditorial. Elle ne garde pas un second graphe. |
| **Stockage** : une synthèse = un document, son niveau, ses preuves, son empreinte | **CortexDB**, sans changement : `ingest_document` dans une collection de synthèses | Le pilote le fait déjà. Le hachage du contenu (#75) dit si une synthèse a changé. |
| **Service** : les synthèses d'abord, puis les passages, puis la descente vers les preuves | **En aval** : Thalamus (`th_recall`, `th_descend`) | C'est un ordre de service. Le propriétaire de CortexDB laisse ce choix à l'appelant (voir la section 4). |
| **Contrôle** : chaque nombre, date ou nom d'une phrase se lit dans les preuves citées | **CortexDB**, par la fiche amont 08 | C'est mécanique, sans modèle, et utile à chaque producteur. |

## 4. Pourquoi pas dans `_recall`

Le propriétaire de CortexDB écrit dans `pkg/graphflow/global_search_tool.go` : « nothing inspects a query's wording to send it here, because deciding that a question is "global" is understanding it ». Le choix du niveau appartient donc à l'appelant. Une fonction `_recall` qui choisit le niveau de la pyramide va contre ce principe.

Le moteur a déjà des parties proches :

- `build_community_hierarchy` et `global_search` : des rapports par communauté, proches des niveaux 0 et 1. Ces outils sont dans le serveur MCP local, pas dans le serveur gRPC de v2.136.0.
- `verify_claims` : le contrôle de triplets (sujet, relation, objet) contre le graphe. Le contrôle d'une phrase contre ses sources n'existe pas.
- Le contrat `_grade` (`pkg/cortexdb/contract.go`). Le pilote donne `self_consistent` aux synthèses. Mais la définition de ce grade dit « derived deterministically », et une synthèse écrite par un modèle n'est pas déterministe. Point à soumettre au propriétaire avec la fiche 08.

## 5. Le service dans Thalamus (tranche 15)

1. **`rights.yaml`** : un cerveau déclare sa collection de synthèses, par exemple `pyramid: pyramide`.
2. **La clé** : une clé confinée peut porter plusieurs collections (`Scope.Collections`). Une requête nomme toujours une seule collection. `key issue` et `key sync` donnent les 2 collections à la clé.
3. **`th_recall`** : pour un cerveau avec `pyramid`, le gateway cherche aussi dans la collection des synthèses. Les synthèses viennent en tête, dans une part du budget, avec leur niveau dans la provenance.
4. **`th_descend`** : l'agent donne l'identifiant d'une synthèse. Le gateway lit ses preuves en métadonnée et rend les documents cités de la collection du cerveau.
5. **Le hook** : rien de nouveau. Il reçoit l'ordre de `th_recall`.
6. **Les droits** ne changent pas. Un lecteur du cerveau lit ses synthèses.
7. **La mesure** : le banc ai-audio du pilote, passé par le gateway (`th_recall`), doit donner la même couverture que la mesure directe (81,2 %).

## 6. Les preuves : identifiants de documents ou de morceaux

L'autre session propose que les preuves d'une synthèse deviennent des identifiants de morceaux (chunks). CortexDB peut alors suivre la provenance et contrôler les phrases (fiche 08).

Une contrainte vient de la liste blanche de v2.136.0 : `get_chunks` est refusé à une clé confinée (« names chunks, whose document the guard does not load »). `th_descend` lit donc les preuves par `knowledge_get`, avec des identifiants de **documents**. Si les preuves deviennent des morceaux, la synthèse doit garder aussi les documents de ses preuves, sous une autre clé de métadonnée. Sinon, `th_descend` ne peut rien ouvrir pour une clé confinée.

## 7. Deux défauts de CortexDB v2.136.0 (constat de l'autre session)

- `ingest_document` écrit des vecteurs factices (mots hachés), même avec un encodeur branché. Il faut recalculer les vecteurs ensuite avec `vector_dimension_repair`.
- `vector_dimension_repair` vide le document d'origine, les métadonnées et les droits de chaque ligne recalculée (12 218 lignes dans le pilote). La cause donnée : `pkg/core/dimension_report.go` relit la ligne sans ces colonnes, puis la réécrit entière. Le défaut est encore sur `main` (`0328e0a`).

Ces deux défauts touchent tout cerveau avec un encodeur. Les 2 sont confirmés par un test sur `main` le 22 octobre 2026 (fiche amont 09).

## 8. Ce qui reste ouvert

- La langue des synthèses dans une organisation en plusieurs langues : une langue pivot, ou une synthèse par langue (voir `docs/notes/multilinguisme.md`).
- La table des équivalents du contrôle (« RGPD », « GDPR », « DSGVO ») : dans contaixt, elle est écrite dans le code (`synthese.EQUIVALENTS`). Dans le moteur, elle devient une donnée que le producteur donne avec la phrase.
- Les cartes par sujet : un niveau de plus, à stocker de la même façon quand contaixt les construit.
