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
| **CortexDB, mode graphe + pyramide servie d'abord** | **78,5 %** |

Source : `cache/pilote_cortexdb/mesure/ai-audio_bilan_2026-10-22T21-01.json` de contaixt. La pyramide est stockée dans CortexDB comme des documents d'une collection `pyramide`, avec leurs preuves en métadonnée (`puits_cortexdb.synthese_vers_cortexdb`). Le client les place avant les passages.

Conclusion : la pyramide garde sa valeur quand CortexDB la stocke et la cherche. Le moteur n'a pas besoin de la construire.

## 3. Les 4 parties et leur place

| Partie | Place | Raison |
| :- | :- | :- |
| **Construction** : écriture par un modèle, choix des notions, thèmes, cartes, cascade | **En amont** : contaixt, ou un autre producteur de l'organisation | Elle appelle un modèle et coûte de l'argent. Le choix des notions et des sujets est éditorial. |
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
7. **La mesure** : le banc ai-audio du pilote, passé par le gateway (`th_recall`), doit donner la même couverture que la mesure directe (78,5 %).

## 6. Ce qui reste ouvert

- La langue des synthèses dans une organisation en plusieurs langues : une langue pivot, ou une synthèse par langue (voir `docs/notes/multilinguisme.md`).
- La table des équivalents du contrôle (« RGPD », « GDPR », « DSGVO ») : dans contaixt, elle est écrite dans le code (`synthese.EQUIVALENTS`). Dans le moteur, elle devient une donnée que le producteur donne avec la phrase.
- Les cartes par sujet : un niveau de plus, à stocker de la même façon quand contaixt les construit.
