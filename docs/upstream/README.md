# PR potentielles vers CortexDB (dossier en réserve)

Décision de Lauro, 20 octobre 2026 : les éléments mécaniques que Thalamus contourne aujourd'hui remontent vers CortexDB, par petites PR, quand Thalamus aura atteint une certaine maturité. **Rien ne part maintenant.** Ce dossier prépare chaque PR pour que, le jour venu, il reste à écrire le test, le code et le corps.

Principe de partage (voir `README.md` de Thalamus, section « Why ») : le moteur reçoit ce qui est mécanique et vaut pour tout client ; Thalamus garde ce qui est une politique d'organisation (`rights.yaml`, groupes, curateurs, fédération entre serveurs). Chaque PR remontée rend Thalamus plus mince. Mesure du 20 octobre 2026 : Thalamus 1 831 lignes hors tests, CortexDB 160 039.

## Conditions de départ

1. Les PR en cours sont fusionnées ou closes (#72 à #74, #78, #82 et #83 fusionnées ; #75 à #77 en relecture au 22 octobre 2026), et le commentaire de clôture de #66 est posté. Décision de Lauro du 22 octobre 2026 : les fiches 01 et 06 partent quand ces PR sont validées et que le dépôt Thalamus est public, pas avant.
2. Thalamus a tourné dans une vraie session Claude Code (`claude --plugin-dir ./plugin`), pas seulement sur le banc.
3. Le dépôt Thalamus est public, pour que chaque PR renvoie au code qui la motive.
4. Les règles de l'amont restent celles du 20 octobre 2026 : environ 400 lignes hors tests par PR, une règle par PR, tests « failing first », table outil par outil pour tout ce qu'une clé confinée atteint.

## Fiches

| Fiche | Titre | Ce que Thalamus retire après la fusion | Taille estimée | Ordre |
| :- | :- | :- | :- | :- |
| [`01-key-file-reload.md`](01-key-file-reload.md) | Rechargement du fichier de clés par `cortexdb-grpc` | Le redémarrage du serveur après `key issue`, `key revoke`, `key sync` (limite connue depuis la tranche 1) | 150 lignes | 1 |
| [`02-review-pending-tool.md`](02-review-pending-tool.md) | Outil `review_pending` | Les deux appels de `th_pending` (`aggregate_metadata` puis `knowledge_get` par document) | 120 lignes | 2 |
| [`03-tool-audit-ids.md`](03-tool-audit-ids.md) | Identifiants touchés dans le journal des requêtes | La moitié de `audit.go` (l'extraction des identifiants par outil, `idsOf`) | 200 lignes | 3 |
| [`04-session-text-permitted.md`](04-session-text-permitted.md) | `loadSessionText` passe par `memoryBucketPermitted` | Rien (sécurité du moteur, déjà couverte par la liste blanche côté clés) | 60 lignes | 4 |
| [`05-ingest-author-from-key.md`](05-ingest-author-from-key.md) | `author` et `written_by` depuis la clé appelante | La pose de `author` par `th_add` | 80 lignes | 5 |
| [`06-graph-tools-confined.md`](06-graph-tools-confined.md) | Les outils de graphe pour une clé confinée (deux PR) | Rien à retirer : la fiche débloque la fédération des entités pour les cerveaux avec `collection` | 250 + 150 lignes | 2, dès la fusion de #76 et #77 |
| [`07-european-lexical-search.md`](07-european-lexical-search.md) | La recherche lexicale pour les langues européennes (quatre PR : jeu d'évaluation suisse, accents égaux sur PostgreSQL, repli sur les trigrammes pour un terme non trouvé, mots vides) | Rien à retirer : la fiche rend les cerveaux en français, en allemand, en italien et en romanche utilisables sans embedder | 150 + 80 + 200 + 150 lignes | 6, le jeu d'évaluation d'abord |
| [`08-synthesis-check.md`](08-synthesis-check.md) | Outil `check_against_sources` : les faits d'une phrase contre ses sources, équivalents donnés par la requête | Rien à retirer : chaque producteur de synthèses (contaixt et les autres) garde un seul contrôle, celui du moteur | 250 lignes | 7 |
| [`09-ingest-embedder-and-repair.md`](09-ingest-embedder-and-repair.md) | 2 défauts de v2.136.0 : `ingest_document` ignorait l'encodeur, `vector_dimension_repair` effaçait `doc_id`, `metadata` et `acl` | **Close** : issue #84, corrigée en v2.137.0 (23 octobre 2026) |
| [`10-find-nodes-confined.md`](10-find-nodes-confined.md) | `find_nodes` ouvert à une clé confinée, derrière le filtre des nœuds cachés de v2.139.0 | Rien à retirer : `chercher` (kg-mcp) passe par `find_nodes` pour un cerveau confiné | 13 lignes | envoyée le 28 octobre 2026 (#111) |
| [`11-find-nodes-fold-once.md`](11-find-nodes-fold-once.md) | `find_nodes` normalise chaque nœud une fois par appel, pas une fois par nom | Rien à retirer : un lot de 8 noms passe de 800 à 160 ms sur ai-audio | 6 lignes et un test | fusionnée le 28 octobre 2026 (#110), v2.140.0 |

L'ordre suit l'utilité pour Thalamus et la taille. Une fiche par PR, un commit par PR, sur une branche `co-NN-<sujet>` du fork `lfoletti/cortexdb` (`co` pour CortexDB, décision de Lauro du 28 octobre 2026 ; les branches envoyées avant cette date gardent leur préfixe `gc-`), selon `docs/pipeline_pr.md`.

Décision de Lauro du 22 octobre 2026 : le produit se conçoit aussi pour des utilisateurs européens (français, allemand, italien, romanche). Chaque fiche dit ce qu'elle change pour ces langues.

## Ce qui ne remonte pas

- `rights.yaml`, les groupes, les lecteurs, rédacteurs et curateurs : une politique, pas un mécanisme.
- Le point fédéré `/mcp` et la fusion RRF entre serveurs : la fédération entre serveurs est le rôle du gateway.
- Le hook et le plugin Claude Code : un client, pas le moteur.
- `key list` et `key sync` : ils comparent le fichier de clés aux droits de `rights.yaml`, qui sont une politique de Thalamus.
