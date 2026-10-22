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

L'ordre suit l'utilité pour Thalamus et la taille. Une fiche par PR, un commit par PR, sur une branche `th-NN-<sujet>` du fork `lfoletti/cortexdb`, selon `docs/pipeline_pr.md`.

Décision de Lauro du 22 octobre 2026 : le produit se conçoit aussi pour des utilisateurs européens (français, allemand, italien, romanche). Chaque fiche dit ce qu'elle change pour ces langues.

## Ce qui ne remonte pas

- `rights.yaml`, les groupes, les lecteurs, rédacteurs et curateurs : une politique, pas un mécanisme.
- Le point fédéré `/mcp` et la fusion RRF entre serveurs : la fédération entre serveurs est le rôle du gateway.
- Le hook et le plugin Claude Code : un client, pas le moteur.
- `key list` et `key sync` : ils comparent le fichier de clés aux droits de `rights.yaml`, qui sont une politique de Thalamus.
