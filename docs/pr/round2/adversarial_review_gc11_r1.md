# Relecture adverse de la PR à part `gc-11-rest-upsert-rows`, tour 1 (21 octobre 2026, tête `a833b4c`)

Rapport de l'agent neuf (invite fixe, `docs/pipeline_pr.md`, section 1 bis). Verdict : 1 bloquant, 4 remarques probables, 3 nits. Résumé fidèle :

- **B1.** `supersedes` sur `POST /v1/memory` (`MemorySaveRequest.Supersedes`, `markMemoriesSuperseded`, `memory_api.go:153`, `:1265`) écrit les lignes d'une autre portée sans contrôle : la clé `hermes` marque `zeus-1` `superseded_by:"hermes-2"` (200, même chose sur `main`). Oracle : une cible absente donne 404 `supersede "no-such-id": embedding not found`, une cible étrangère 200. Écriture partielle : la nouvelle ligne reste malgré le 404. Le `SaveMemoryRequest` protobuf n'a pas `supersedes` : la parité gRPC ne le couvre pas. La liste blanche de #72 classe déjà `supersedes` en identifiant.
- **L2.** Des `entities` dans un upsert confiné réécrivent des nœuds partagés d'autres collections (description, propriétés, vecteur, `collections`). Antérieur à la PR, présent aussi en gRPC ; #72 les refuse aux clés confinées sur les outils. Fermer ou nommer hors périmètre.
- **L3.** Un identifiant blanc (`" "`, `"\t"`) contourne la garde (`strings.TrimSpace(id) == ""`) alors que la façade l'accepte : la clé `notes` déplace `" "` de `secrets` à `notes`, la clé `hermes` prend `"\t"` de Zeus. Même saut en gRPC (`ownership.go:231`, `:255`). Ne sauter que `""`.
- **L4.** Troisième copie de la règle de ligne (après `rpcserver/ownership.go` et `authz/caller.go`) ; un contrôle dans la façade (`CallerScope` depuis v2.135.0) fermerait les trois chemins à la fois.
- **L5.** « Absent » décidé sur le texte de l'erreur (`statusFor(err) == 404`) ; #72 est passé aux sentinelles.
- **N6.** La garde répond 500 pour toute panne de lecture, là où `statusFor` donnerait 499 ou 504.
- **N7.** Trois tests tombent sur `main`, pas deux ; pas d'orthographe de `memory_id` testée ; rien ne tient « les clés non confinées ne sont pas contrôlées » ; la fenêtre lecture puis écriture n'est pas nommée.
- **N8.** « depuis v2.91.0 » est juste.

Vérifié par l'agent : `go test -race ./pkg/httpapi/` vert ; couverture 100 % des fonctions nouvelles ; pas de `sleep`, rien ne lit l'environnement. Les 15 mutations n'ont pas été rejouées par l'agent.
