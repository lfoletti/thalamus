# ADR 0001 : liste blanche explicite par outil pour les clés confinées

- Statut : acceptée le 20 octobre 2026, en relecture amont (PR #72).
- Date : 20 octobre 2026.

## Contexte

CortexDB sert un cerveau partagé à plusieurs équipes. Une clé d'API confinée (`collection`, `user_id`, `scope` ou `namespace`) doit pouvoir appeler la boîte à outils (`CallTool` en gRPC, `POST /v1/tools/<name>` en REST), car le client MCP distant ne parle que par ces outils. Les arguments d'un outil sont du JSON opaque que la politique de droits ne lisait pas.

La première version (19 octobre 2026) dérivait la callabilité du schéma : un outil qui déclare le champ confiné au premier niveau passait. Le propriétaire a trouvé cinq sorties de la portée (revue du 20 octobre 2026) : `promotion.*` de `knowledge_memory_consolidate`, les outils de rappel qui lisent les deux magasins, le prefetch graphe de `cortex_query`, `session_id` d'`extract_conversation`, `supersedes` de `memory_save`. Toutes ont la même cause : la règle ne savait rien de ce que l'outil fait du reste de ses arguments.

## Décision

Une liste blanche explicite, une ligne par outil, fermée par défaut (`pkg/authz/allowlist.go`). Chaque ligne nomme : les champs de portée que l'outil honore, les magasins qu'il touche, les arguments de données admis, chaque champ porteur d'identifiant par son chemin complet (`plan.filters.document_ids`, `supersedes`), les portes à forcer, les arguments refusés avec leur raison. Un outil hors liste, ou un argument que la ligne ne nomme pas, est refusé à une clé confinée. Un test de complétude tient la liste au schéma des outils : un outil ou une propriété nouvelle fait tomber le test tant qu'il n'est pas classé. La table de la liste est générée dans le corps de la PR.

## Options écartées

- Dérivation du schéma (« l'outil déclare le champ ») : ne voit pas les identifiants imbriqués ni les magasins croisés. C'est la version refusée.
- Correction faille par faille : laisse ouvert tout outil futur. Le propriétaire a demandé une règle vérifiable outil par outil.
- Liste blanche sans test de complétude : diverge du schéma au premier outil ajouté.

## Conséquences

- Chaque outil nouveau demande une décision explicite, consignée dans la liste.
- La PR fait 727 lignes hors tests, dont environ 200 de liste. Le propriétaire a accepté de ne pas la couper.
- Toute PR qui ajoute un argument d'outil doit toucher la liste (voir ADR 0005).

## Références

- PR amont #72 (`gc-04a-authz-tool-rules`), discussion #66.
- `../thalamus_local/docs/pr/lot_B/GC-04a.body.md`, `../thalamus_local/docs/pr/lot_B/failing_first.md`.
