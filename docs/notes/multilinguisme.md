# Note : le multilinguisme dans Thalamus

22 octobre 2026. Décision de Lauro : le produit se conçoit aussi pour des utilisateurs européens. Le premier cas est une organisation suisse, avec des documents et des questions en français, en allemand, en italien et en romanche.

## 1. Le problème

Un employé pose une question dans sa langue. Le document qui répond est souvent dans une autre langue. Par exemple, une question en français porte sur un règlement publié en allemand. Le rappel doit donc trouver ce document, à travers la langue.

Il y a 3 niveaux de difficulté :

1. **Une même langue, des formes différentes** : les accents (« État », « etat »), les mots composés allemands (« Krankenversicherung »), les flexions (« décider », « décision »).
2. **Des langues différentes** : la question est en français, le document en allemand.
3. **Une même chose avec 4 noms** : « Conseil d'État », « Staatsrat », « Consiglio di Stato », « Cussegl guvernativ ».

## 2. L'état du moteur (CortexDB v2.136.0, mesuré le 22 octobre 2026)

| Point | SQLite | PostgreSQL |
| :- | :- | :- |
| Accents (`etat` trouve « État ») | oui | **non** |
| Mot composé (`Versicherung` trouve « Krankenversicherung ») | non (index des mots) | non (index des mots) |
| Index par trigrammes | existe pour tous les textes, utilisé seulement pour une requête CJK | idem (pg_trgm) |
| Racinisation | non, par choix | non, par choix ('simple') |
| Mots vides | anglais seulement | anglais seulement |
| Recherche d'une langue à une autre | seulement avec un embedder (`CORTEXDB_EMBED_*`) | idem |

Les outils de recherche acceptent déjà `keywords` et `alternate_queries`. Leur description demande à l'agent des variantes dans d'autres langues.

Le détail et les propositions au moteur sont dans la fiche amont `docs/upstream/07-european-lexical-search.md`.

## 3. Le pipeline proposé, en 3 couches

| Couche | Ce qu'elle règle | Où | Coût |
| :- | :- | :- | :- |
| 1. Embedder multilingue (par exemple `bge-m3`) sur chaque serveur | Niveau 2 : la question trouve le document dans une autre langue | Réglage du serveur | 1 à 2 Go de mémoire, aucun code |
| 2. Variantes de la requête | Niveaux 1 et 2 sans embedder : l'agent donne la question dans les 4 langues | Thalamus : `th_recall` relaie `keywords` et `alternate_queries`, la compétence du plugin les demande | Petit changement |
| 3. Langue et alias | Niveau 3 : un identifiant d'entité et 1 alias par langue ; chaque document porte sa langue | Thalamus (`th_add`, extension) et l'ontologie du cerveau | Choix de l'organisation |

Le moteur règle le reste (fiche 07) : les accents égaux sur les deux serveurs, le repli sur les trigrammes pour les mots composés, les mots vides des 4 langues.

## 4. Le cas du hook

Le hook de rappel automatique envoie le prompt tel quel. Il n'a pas de modèle, donc il ne peut pas traduire la question. Sans embedder multilingue sur le serveur, le rappel automatique reste donc dans la langue du prompt. Le rappel à la demande (`th_recall` par l'agent) peut, lui, ajouter les variantes de la couche 2.

Conséquence : pour une organisation en plusieurs langues, l'embedder multilingue est la condition du rappel automatique.

## 5. Le romanche

Le romanche est le cas difficile. Les modèles ouverts connaissent peu cette langue, pour les embeddings et pour l'extraction. [Supposé, à mesurer] Nous n'avons pas de personne qui parle romanche (constat du 22 octobre 2026).

La méthode retenue :

- Le jeu d'évaluation utilise des textes officiels déjà traduits par des professionnels : le droit des Grisons (allemand, italien, romanche) et les textes fédéraux en romanche sur Fedlex.
- Les questions sont en français, en allemand ou en italien. La réponse attendue est le passage en romanche.
- Les questions en romanche restent hors du jeu jusqu'à une relecture, par exemple par la Lia Rumantscha.
- Pour la recherche, la version allemande ou italienne d'un texte officiel peut porter le rappel, et la version romanche reste la source citée.

## 6. La mesure d'abord

Aucun choix ne se fait sans mesure. Le jeu d'évaluation suisse (`dataset_ch.json`, environ 40 questions, dont des questions croisées) se mesure sur 3 configurations :

1. l'index lexical seul ;
2. l'index lexical avec les variantes de la requête ;
3. l'embedder `bge-m3`.

Les chiffres disent quelles PR de la fiche 07 sont nécessaires, et ce que vaut le romanche.

## 7. Ce qui reste ouvert

- L'interface de l'extension et les messages du gateway sont en anglais. La traduction de l'interface est un autre sujet.
- La langue des synthèses, si une pyramide de synthèses entre dans le produit : une langue pivot, ou une synthèse par langue.
- La licence des textes officiels repris dans le jeu d'évaluation.
