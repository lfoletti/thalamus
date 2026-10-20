# ADR 0004 : expiration, filtre consultatif, balayage comme garantie, balayage exposé comme outil

- Statut : acceptée le 20 octobre 2026, en relecture amont (PR #76).
- Date : 20 octobre 2026 (première version le 19 octobre 2026).

## Contexte

Une règle de rétention (« cette source est gardée cinq ans ») devait s'appliquer à la main. Une mémoire avait un `ttl_seconds` que les lectures sautaient, mais rien ne supprimait les lignes. Un document n'avait pas d'expiration. La revue du 20 octobre 2026 a montré que le filtre manquait au chemin vectoriel, à `search_text`, aux outils par chunks et à `expand_graph`, que `--sweep-expired` balayait le fichier local d'un client distant, et que la requête de porte n'avait pas d'index.

## Décision

- `expires_at` (RFC 3339) sur `ingest_document`, `knowledge_save` et `knowledge_update`, en argument ou en métadonnée ; une valeur vide efface l'expiration ; une valeur sans `expires_at` dans une requête reporte l'expiration stockée.
- Le filtre est consultatif : un document expiré est laissé hors de chaque chemin qui lit des chunks (lexical, vectoriel, marche de graphe, `search_text`, outils par chunks, `expand_graph`, chaque voie de `cortex_query`), par une requête indexée mise en cache 15 s. Les voies de `cortex_query` lisent le magasin directement : la porte s'applique dans `Query`, une fois par voie. Le balayage est la garantie : `DB.SweepExpired` supprime les documents et mémoires expirés avec leurs nœuds.
- Le balayage est un outil de la boîte à outils (`sweep_expired`, `dry_run`), déclaré comme écriture, refusé aux clés confinées. La commande `cortexdb-mcp-stdio --sweep-expired` l'appelle par gRPC quand `CORTEXDB_REMOTE` est défini, et balaie le fichier local sinon. Elle ne touche jamais le fichier local d'un client distant.
- Index d'expression sur l'extraction gardée de `expires_at`, en SQLite et en PostgreSQL ; la requête de porte utilise le prédicat que le planificateur sert depuis l'index (`> ''`, pas `IS NOT NULL`).

## Options écartées

- Refuser le balayage en mode distant avec un message (version du 20 octobre au matin) : ferme la faute mais laisse le cerveau partagé sans balayage depuis un client. Remplacé le soir même, sur demande de Lauro.
- Un RPC `SweepExpired` dans `AdminService` : demande la chaîne protobuf (absente de la machine) et un client par langage. L'outil existant `CallTool` suffit.
- Un filtre bloquant sur panne pour l'expiration : le propriétaire accepte l'ouverture pour l'expiration seule ; la fermeture vaut pour la porte de relecture (ADR 0003).

## Conséquences

- Tout client MCP peut lancer le balayage avec une clé non confinée, local ou distant.
- Une PR qui ajoute un outil doit le classer dans `toolWrites`, `toolCount` et la liste blanche (ADR 0005).

## Références

- PR amont #76 (`gc-05-expires-at`), `docs/pr/lot_C/GC-05.body.md`, `docs/pr/lot_C/failing_first.md`.
