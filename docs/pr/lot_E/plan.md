# Lot E : toutes les nouvelles PR vers CortexDB (plan)

23 octobre 2026. Commande de Lauro : « prépare toutes les nouvelles PR », fiche 06 puis fiche 01 en tête, envoi d'un coup avec un billet de discussion qui présente Thalamus (`discussion.md`). **Rien ne part et rien n'est poussé avant sa commande.**

Base : `origin/main` à `a591ae1` (v2.138.0 et un commit de commentaire). Chaque fiche a été vérifiée contre v2.138.0 par un agent neuf ; rapports et sondes dans `recheck/`. Les règles de l'amont restent celles du 20 octobre 2026 : environ 400 lignes hors tests par PR, une règle par PR, tests « failing first », table outil par outil pour ce qu'une clé confinée atteint.

## Les PR

| Branche | Fiche | Règle | Lignes hors tests (estimation) | Sur |
| :- | :- | :- | -: | :- |
| `gc-13-find-nodes-hidden` | 06-A | `find_nodes` laisse de côté les nœuds cachés (expirés, non relus), comme `get_nodes` | 20-40 | main |
| `gc-14-graph-caller-scope` | 06-B1 | Sous un `CallerScope`, les documents et les mémoires hors périmètre sont cachés dans toute réponse de graphe, par le même point que l'expiration | 150-200 | main |
| `gc-15-graph-scoped-entities` | 06-B2 | Sous un périmètre, une entité qu'aucune collection permise ne connaît n'est pas servie, et une entité garde seulement `name` et `type` | 150-200 | gc-14 |
| `gc-16-allowlist-graph-tools` | 06-C | La liste blanche ouvre les 8 outils de graphe, confinés par la façade | 120-180 | gc-15 |
| `gc-17-query-graph-lane-scope` | 06-E | La voie graphe d'une requête confinée ne suit plus les relations des autres collections (fuite à 2 sauts de `/v1/query`, présente aujourd'hui) | 60-120 | main |
| `gc-18-allowlist-cortex-query` | 06-D | La liste blanche ouvre `cortex_query`, `collection` forcée, `prefetch[].max_hops` refusé tant que gc-17 n'est pas fusionnée | 40-80 | main |
| `gc-19-key-file-reload` | 01 | Le serveur relit le fichier de clés quand il change, sans redémarrage ; un fichier cassé garde l'ancien jeu | 120-160 | main |
| `gc-20-session-text-scope` | 04 | `loadSessionText` respecte `memoryReadPermitted` | ~10 | main |
| `gc-21-ingest-signs-writer` | 05 | `ingest_document` signe `written_by` comme `knowledge_save`, et la liste blanche reprend `written_by` et `updated_by` | ~20 | main |
| `gc-22-trace-ids` | 03 | Le journal des requêtes nomme les identifiants donnés dans les arguments | ~60 | main |
| `gc-23-knowledge-get-list` | 02 | `knowledge_get` prend une liste d'identifiants (`missing` pour l'inconnu ou hors périmètre) | 40-50 | main |
| `gc-24-eval-swiss` | 07-1 | Jeu d'évaluation `dataset_ch.json` (Constitution fédérale, de, fr, it, rm) | ~10 + données | main |
| `gc-25-coverage-accents` | 07-2 | Le seuil de couverture de v2.138.0 compare les mots sans accents, comme FTS5 | 40-70 | main |
| `gc-26-variant-words` | 07-3 | Les mots des variantes entrent dans la requête OU | 10-20 | main |
| `gc-27-variants-run` | 07-4 | Les variantes tournent même quand la question remplit `top_k` (2 sorties anticipées) | ~5 | main |
| `gc-28-check-against-sources` | 08 | Outil `check_against_sources` : les faits d'une phrase absents de ses sources | 300-340 | main |

## Ce qui ne part pas en PR

Questions au propriétaire, dans le billet de discussion : `unaccent` sur PostgreSQL (07-5), mots composés et repli sur les trigrammes (07-6a et 6b, contraires à un choix explicite de 7e0aa86), mots vides (07, abandonnés : la rareté de v2.138.0 fait ce travail sans liste), le grade d'une synthèse (08).

## Méthode

Un agent de développement par groupe, chacun dans ses copies de travail (`scratchpad/wtE/<branche>`), sans poussée ni appel à GitHub, sans PostgreSQL. Pour chaque PR : tests écrits d'abord et vus en échec, code, une mutation à la main par règle tuée par un test nommé (`failing_first_<branche>.md`), `make precommit` vert (couverture 100 % du patch hors `main`), corps `GC-NN.body.md`. Puis relecture adverse par un agent neuf sur les têtes gelées (`docs/pipeline_pr.md`, section 1 bis), `make prepr` par tête, et une seule poussée sur commande de Lauro.
