# Banc suisse : le rappel d'une langue à une autre

22 octobre 2026. Premier jeu d'évaluation multilingue de Thalamus (note `docs/notes/multilinguisme.md`, fiche amont 07).

## Le jeu

- **Corpus** : les articles de la Constitution fédérale (RS 101, version du 3 mars 2024, Fedlex), chacun dans **une seule langue**, de, fr, it et rm à tour de rôle. 219 articles (les articles abrogés sont exclus). Sources HTML dans `sources/`, extraction par `extract.py` vers `articles.json`.
- **Questions** : 40 questions dans `questions.json`, en français, en allemand ou en italien. Aucune question n'est dans la langue de son article. 11 articles attendus sont en romanche. Aucune question ni variante n'est en romanche.
- **Variantes** : chaque question a sa traduction dans les 2 autres langues (`v`). Pour un article en de, fr ou it, une des variantes est donc dans la langue de l'article. Pour un article en romanche, aucune.
- **Mesure** : le rang de l'article attendu dans les 10 résultats de `th_recall` (`syntheses: 0`), sans variante et avec les 2 variantes dans `alternate_queries`. Écriture par `knowledge_save` à travers le gateway.

## Commande

```
sh tests/eval_ch/run.sh lexical
sh tests/eval_ch/run.sh sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2 384
sh tests/eval_ch/run.sh intfloat/multilingual-e5-large 1024
```

Encodeur local par `fastembed` (`encoder.py`), servi à `cortexdb-grpc` par `OPENAI_BASE_URL`. Résultats par question dans `results/`.

## Résultats (CortexDB v2.136.0, Thalamus tranche 17)

Article attendu au 1er rang, dans les 5 premiers, dans les 10 premiers, et rang réciproque moyen (MRR) :

| Configuration | @1 | @5 | @10 | MRR | rm @5 | rm @10 |
| :- | -: | -: | -: | -: | -: | -: |
| lexical, question seule | 0,00 | 0,05 | 0,07 | 0,02 | 0,09 | 0,09 |
| lexical, avec variantes | 0,10 | 0,50 | 0,62 | 0,28 | 0,09 | 0,18 |
| MiniLM-L12 multilingue (384), question seule | 0,25 | 0,70 | 0,78 | 0,44 | 0,36 | 0,45 |
| MiniLM-L12 multilingue, avec variantes | 0,65 | 0,80 | 0,88 | 0,72 | 0,45 | 0,55 |
| multilingual-e5-large (1024), question seule | 0,03 | 0,55 | 0,82 | 0,21 | 0,45 | 0,82 |
| multilingual-e5-large, avec variantes | **0,75** | **0,97** | **0,97** | **0,85** | **0,91** | **0,91** |

Rejeu sur CortexDB v2.138.0 (23 octobre 2026, rappel par rareté des mots) : mêmes chiffres à une question près. Lexical avec variantes 0,65 à @10 (0,62 avant) ; MiniLM avec variantes 0,90 (0,88), romanche 0,64 (0,55) ; e5-large sans variante @1 0,00 (0,03), avec variantes @1 0,72 (0,75), @5 et @10 inchangés.

## Constats

1. **Sans encodeur, le rappel ne traverse pas la langue** : 7 % des articles dans les 10 premiers. Les variantes aident seulement quand une variante est dans la langue du document (62 % en tout, 18 % pour le romanche).
2. **L'encodeur multilingue est la condition du rappel automatique** (le hook n'envoie pas de variante). e5-large trouve 82 % des articles dans les 10 premiers sans variante, romanche compris (82 %).
3. **Les variantes et l'encodeur s'ajoutent** : e5-large avec variantes donne 97 % dans les 5 premiers, 91 % pour le romanche. Le romanche profite des variantes en de et en it, qui sont proches de lui dans l'espace de l'encodeur.
4. **Le 1er rang souffre de la fusion.** Une liste dans la mauvaise langue a aussi un 1er rang, sans rapport. RRF le met à égalité avec le bon 1er rang d'une variante, et l'ordre des listes décide. D'où un @1 bas en lexical avec variantes (10 %) alors que @5 est à 50 %.
5. **e5-large sans variante a un @1 de 3 %** pour un @10 de 82 %. Causes possibles, non vérifiées : e5 attend les préfixes `query:` et `passage:`, que CortexDB n'ajoute pas ; la partie lexicale de la recherche hybride met en tête des articles de la langue de la question.
6. **Le coût** : écriture de 219 articles en 1 s (lexical), 14 s (MiniLM), 140 s (e5-large) sur la machine de mesure (Mac mini, CPU).

## Limites

- 40 questions : un écart de quelques points ne dit rien.
- Les questions et les variantes sont écrites par le modèle (Claude), pas par des locuteurs. Le texte romanche est la traduction officielle.
- Un article par document, court : le corpus d'une organisation a des documents plus longs et plus variés.
- `bge-m3` n'est pas dans `fastembed` ; e5-large est le modèle de la même classe disponible.

## Jeu d'évaluation pour CortexDB

`make_dataset_ch.py` construit le jeu `pkg/eval/testdata/dataset_ch.json` proposé à CortexDB : les 24 articles les plus courts des questions, en de, fr, it et rm (96 documents), et 230 requêtes dans la langue de l'article (question et variantes, titres, titres sans accents, mots composés allemands), plus 6 requêtes négatives. Commande : `python3 -I tests/eval_ch/make_dataset_ch.py . dataset_ch.json`. Le fichier sort identique octet par octet à chaque exécution.
