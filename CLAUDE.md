# Thalamus

Couche mince au-dessus de CortexDB pour des graphes de connaissances d'entreprise utilisés par des agents de code. Contenu : les documents du projet, un clone de CortexDB avec sept PR ouvertes en amont, et depuis le 20 octobre 2026 la première tranche du gateway (`gateway/`). La racine est un dépôt git (initialisé le 20 octobre 2026, branche `main`), poussé le même jour sur GitHub (**privé**, MIT, remote `origin`). Nom du projet depuis le 20 octobre 2026 au soir : **Thalamus** (décision de Lauro ; le thalamus relaie et filtre tout ce qui monte au cortex), préfixe d'outils `th_`, binaires `th-gateway` et `th-hook`, module `github.com/lfoletti/thalamus/gateway`, plugin `thalamus`, clés `th/<login>/<cerveau>`. Le dépôt GitHub s'appelle encore `lfoletti/granular-cortex` : `gh repo rename thalamus` sur commande de Lauro (GitHub redirige l'ancien nom), puis `git remote set-url`. Le dossier sur disque reste `granular_cortex` ; `cortexdb/` est un dépôt à part, ignoré par la racine. Passage en public prévu à la fusion des sept PR amont, avec le commentaire de clôture de #66. Aucun push sans demande explicite de Lauro, même vers ce dépôt.

## Arborescence

- `docs/` : documents du projet.
- `gateway/` : module Go `github.com/lfoletti/thalamus/gateway` (CortexDB v2.134.1 en bibliothèque, SDK MCP v1.6.1). Paquets `rights` (lecture de `rights.yaml`), `keys` (émission et révocation des clés confinées dans le fichier de clés du serveur), `relay` (serveur MCP HTTP par cerveau, relais gRPC avec la clé de l'appelant, `audit.jsonl`), `cmd/th-gateway`. Lire `gateway/README.md`.
- `plugin/` : plugin Claude Code `thalamus` (hook de rappel, serveur MCP du gateway, compétence). Lire `plugin/README.md`.
- `tests/e2e_gateway.sh` : banc de bout en bout du gateway (serveur gRPC construit depuis `cortexdb/`, 40 contrôles, environ 1 min). `tests/e2e_shared_brain.sh` : le banc du clone seul (30 contrôles).
- `cortexdb/` : clone complet de `github.com/liliang-cn/cortexdb` (amont, `origin` ; fork de l'utilisateur, `fork` = `lfoletti/cortexdb`). Branche de travail : `proto/granular-cortex`, créée le 18 octobre 2026, reconstruite sur `main` à v2.134.0 le 19 octobre 2026.

## État de la branche `proto/granular-cortex` (20 octobre 2026)

Reconstruite le 20 octobre 2026 sur `origin/main` à v2.134.1 (`60a8831`), après la revue du propriétaire (« changes requested » sur les six PR, 20 octobre 2026, 01 h 21 UTC). Les correctifs 1, 2, 3, 7, 9 et 12 sont fusionnés en amont (PR #67 à #71, publiés en v2.133.0). Il reste six commits, un par correctif, chacun avec son entrée `Unreleased` dans `cortexdb/CHANGELOG.md` (texte réécrit le 20 octobre 2026), identiques aux têtes des six branches de PR plus le changelog :

| Commit `proto` | Branche de PR (tête) | PR | Correctif |
| :- | :- | :- | :- |
| `e491f29` | `gc-04a-authz-tool-rules` (`3c93d00`) | #72 | 4a : liste blanche explicite par outil (`pkg/authz/allowlist.go`), fermée par défaut, portes forcées (`ConfineToolArgs`), identifiants par chemin complet. Non câblée. |
| `5c82926` | `gc-04b-grpc-confined-tools` (`49949eb`, sur 04a) | #73 | 4b : câblage gRPC (`CallTool`), avec les correctifs 10 et 11. |
| `81b3a62` | `gc-04c-rest-confined-tools` (`4b3ff4c`, sur 04a) | #74 | 4c : câblage REST (`POST /v1/tools/<name>`), refus (500) sans base. |
| `8160817` | `gc-08-content-hash` (`10e4b76`) | #75 | 8 : hash SHA-256 et chunking stockés, « unchanged » exact, dédup par collection, écriture en une transaction, chunks en ordre. |
| `99fe25c` | `gc-05-expires-at` (`3c7a9fd`, sur 08) | #76 | 5 : `expires_at` sur tous les chemins (vectoriel, outils par chunks, `expand_graph`, voies de `cortex_query`), index d'expression, outil `sweep_expired`, `--sweep-expired` balaie le cerveau partagé par gRPC en mode distant. |
| `ce6bb8e` | `gc-06-review-status` (`036d1eb`, sur 05) | #77 | 6 : `review_status` opt-in (`CORTEXDB_REVIEW_GATE=1`), relecteur = clé appelante, report de la décision, pas d'effacement, fusion des métadonnées sur une décision (`knowledge_update`), `reviewed_at` stable, fermeture sur panne, index. |
| `cd5ad43` | `gc-10-cortex-query-prefetch` (`49c5366`, sur `main`) | #78 | 10 : la voie graphe de `cortex_query` reste dans la collection (point 3 de la revue de #72). Dossier `docs/pr/lot_D/`. |

Un commit sur `proto` seulement (`d5131f7`, entre 06 et 10) classe dans la liste blanche les arguments et l'outil du lot C (`dedupe`, `expires_at`, `plan.filters.review_status`, `sweep_expired` refusé) : chaque PR du lot C l'absorbe pour ses propres arguments au rebasage sur un `main` qui contient #72 (voir `docs/pr/lot_C/README.md`, « Dépendance croisée »). **Lot B poussé le 20 octobre 2026 à 19 h 50** sur commande de Lauro : #72, #73, #74 repassées en brouillon, têtes amendées poussées (hook vert, 0 bloc sans test), corps mis à jour, un commentaire de réponse chacune ; GC-10 ouverte en brouillon, **PR #78**. Codecov sur les trois têtes du lot B : « 100.00% of diff hit », donc #72, #73, #74 passées « ready for review » à 19 h 50 ; #78 : CI verte, Codecov « 100.00% of diff hit », passée « ready for review » à 19 h 55. **Lot C poussé le 20 octobre 2026 à 20 h 05** sur commande de Lauro : #75, #76, #77 repassées en brouillon, têtes poussées (`10e4b76`, `3c7a9fd`, `036d1eb` ; hook vert avec `DECLARED=pkg/core/store_postgres.go`, option `--declare` de `patchcov` ajoutée ce soir), corps mis à jour, un commentaire par PR, commentaire de lot sous la discussion #66. Codecov (API `compare`) : #75 0 ligne, #76 3 lignes (2 `main` de `cortexdb-mcp-stdio`, 1 DDL), #77 7 lignes (3 `main` de `cortexdb-grpc`, 2 `main` de `cortexdb-mcp-stdio`, 2 DDL), exactement les lignes déclarées ; les trois passées « ready for review » à 20 h 20. **Les sept PR (#72 à #78) sont en relecture.** Commentaire de clôture de #66 en réserve (`docs/pr/closing_comment_66.md`), à poster après les fusions. Pipeline complet passé sur les six têtes le 20 octobre 2026 (`prepr` sur 04a, 04b, 04c, 08 ; `mutate` + suite `-race` + éval à la main sur 05 et 06, où le bloc DDL PostgreSQL arrête `precommit`) : suites vertes, mutants survivants lus et classés dans les `failing_first.md`. Dossiers : `docs/pr/lot_B/` (#72 à #74) et `docs/pr/lot_C/` (#75 à #77), chacun avec `README.md`, les corps `GC-*.body.md`, `failing_first.md` (mutations de l'amendement), `review_replies.md` (brouillons de réponses au relecteur). Procédure de poussée, sur commande explicite de Lauro : `gh pr ready --undo` sur la PR, `git push --force-with-lease fork <branche>` (hook `pre-push` = `precommit`), corps mis à jour (`gh pr edit --body-file`), commentaire de réponse, puis `gh pr ready` quand Codecov montre zéro ligne non couverte hors celles déclarées. Après chaque fusion, rebaser les branches empilées dessus et pousser de force. Sauvegardes : tags `backup/*-2026-10-19`, `backup/*-2026-10-20`, `backup/*-2026-10-20-evening`, `backup/proto-2026-10-20-night`. Niveau 3 passé le 20 octobre 2026 au soir sur les têtes finales de #76 et #77 (journal dans `docs/pr/lot_C/failing_first.md`).

Résumé des demandes du propriétaire (20 octobre 2026) : mémoire CortexDB `feedback:granular-cortex:upstream-review-2026-10-20`. Une PR à part viendra pour `cortex_query` (prefetch sans collection, `pkg/cortexdb/query_api.go`) ; l'outil entre dans la liste blanche après.

Essai de bout en bout (serveur gRPC + fichier de clés + client MCP distant) : `sh tests/e2e_shared_brain.sh` depuis la racine, 30 contrôles, environ 2 minutes. Vert sur `proto` reconstruite le 20 octobre 2026 au soir, dernière fois à 19 h 30 sur `cd5ad43` (30/30, huit commits). Plan de publication : `docs/plan_publication.md`, à valider par l'utilisateur avant toute action publique.

## Gateway (tranches 1 à 5, 20 octobre 2026)

Tranche 1 : droits (`rights.yaml`), clés confinées émises dans le fichier du serveur, relais MCP par cerveau (`/mcp/<cerveau>`), `audit.jsonl`. Tranche 2 : point fédéré `/mcp` (`th_brains`, `th_recall` : fan-out avec la clé de l'employé par cerveau, fusion RRF pondérée par `weight`, provenance, `skipped`, audit des identifiants lus) et cerveau sans `collection` = serveur entier, clé non confinée, outils de graphe ouverts (réponse à la question du 20 octobre 2026 sur l'usage graphe de connaissances hors codage). Tranche 3 : le client, `gateway/cmd/th-hook` (hook `UserPromptSubmit` : `th_recall` avec `TH_GATEWAY`/`TH_KEY`, bloc `additionalContext` avec provenance, budget 9 000, jamais bloquant) et le plugin Claude Code `plugin/` (manifeste, `.mcp.json` HTTP vers `/mcp` avec `${TH_GATEWAY}`/`${TH_KEY}`, `hooks/hooks.json`, lanceur `bin/th-recall`, compétence). Tranche 4 : administration des clés, `th-gateway key list` (clés par cerveau contre les droits, statut `ok` ou `stale: …`) et `key sync [--dry-run]` (émet, réémet, révoque, porteurs hors `rights.yaml` compris, plan puis actions). Tranche 5 : circuit d'écriture sur `/mcp` (`th_add` → `ingest_document` avec `review_status: proposed`, auteur = l'employé, accepté d'office pour un curateur ; `th_pending` par `aggregate_metadata` puis `knowledge_get` sans `collection`, `th_review` par `knowledge_update` ; les deux réservés aux `curators` de `rights.yaml`, nouveau niveau `Curate` au-dessus de `Write`) et commande `/th-add` du plugin. Vert : `go test -race ./...` dans `gateway/` (couverture 100 % hors les deux `main`), `python3 -I mutate.py` (91 mutations à la main, 91 tuées, `docs/gateway/failing_first.md`), `sh tests/e2e_gateway.sh` (40/40, dont un document proposé caché par la porte de relecture, listé par le curateur, accepté, puis servi). Limites connues : `cortexdb-grpc` lit son fichier de clés au démarrage seulement (redémarrage après `key issue` ou `key revoke` ; rechargement à proposer en amont) ; rappel fédéré sur les extraits, pas sur les nœuds ; pas d'OIDC, pas d'extraction, pas de contradictions. Les mêmes règles de vérification que pour les PR s'appliquent (trois niveaux, tests avant le code, mutation à la main par règle). Prochaines tranches : rechargement des clés (PR amont après les fusions) et de `rights.yaml` par le gateway, `/v1/ingest` REST pour l'extension Chrome, fédération au niveau des entités, OIDC. Le plugin n'a pas encore été essayé dans une session Claude Code réelle (`claude --plugin-dir ./plugin`) : seul le hook l'est, par le banc.

## Règles de publication

- Tous les correctifs se développent **localement** sur `proto/granular-cortex`, **un commit par correctif**, avec une entrée `Unreleased` dans `CHANGELOG.md`. Cette entrée est à usage interne : le propriétaire amont écrit seul le changelog à la release, donc chaque PR retire cette modification (`docs/plan_publication.md`, section 0 bis).
- Le produit final se teste en entier avant toute soumission.
- **Tests « failing first »** (exigence du propriétaire amont, règle générale depuis le 19 octobre 2026) : chaque test d'une PR doit échouer sans son code. Vérifier avant toute PR : fichiers de test posés sur l'arbre sans le correctif, ou mutation de chaque règle si les tests ne compilent pas sans elle. Preuve dans le corps de la PR, section « Failing first ». Méthode et exemple : `docs/pr/lot_B/failing_first.md`.
- **Aucun push, aucune pull request** vers l'amont ou le fork sans demande explicite de l'utilisateur. La soumission se fera par petits lots, après validation d'un plan de publication.
- **PR en brouillon d'abord** : une PR s'ouvre en brouillon et ne passe à la relecture qu'après un rapport Codecov à zéro ligne non couverte sur sa tête (ou les seules lignes de `main` déclarées). Voir le pipeline ci-dessous.
- Auteur des commits : `git -c user.name="Lauro Foletti" -c user.email="lauro.foletti@gmail.com" commit`, message en style conventional-commit, **sans** ligne `Co-Authored-By` (décision de Lauro du 19 octobre 2026 : rien de tel sur les PR amont ; la règle vaut pour tous les commits de ce dépôt).

## Pipeline avant toute PR (procédure complète : `docs/pipeline_pr.md`)

En vigueur depuis le 19 octobre 2026, validé le soir même sur six PR (quatre à 100 % de couverture du patch, deux à 100 % hors `main`).

**Trois niveaux** (décision du 20 octobre 2026, `docs/pipeline_pr.md`, section 1) : niveau 1, `make precommit` et la mutation à la main de la règle changée, à chaque changement ; niveau 2, analyse, relecture, ADR, tout ce qui peut révéler un trou, **avant** de lancer la chaîne lourde, têtes gelées ensuite, aucune copie de travail touchée pendant la chaîne ; niveau 3, `make prepr` une fois par lot, juste avant la poussée, relancé seulement si le code change. Un trou trouvé pendant la chaîne va dans une liste ; la chaîne repart une fois, à la fin.

1. Le patch est couvert à 100 % de ses lignes atteignables **avant le premier push**, mesuré en local : `make precommit REPO=<copie> BASE=<base>` (gofmt, vet, `-race` des paquets touchés, `patchcov`). Les blocs du corps de `func main()` sont affichés à part, sans bloquer : les déclarer dans une section `## Codecov` du corps de la PR, après `## Breaking Changes` et de même rang (« The N lines left uncovered are inevitable. … »).
2. **Le test s'écrit avant le code** et tombe sans lui. La preuve livrée au relecteur : une mutation à la main par règle qui fait tomber un test nommé, consignée dans `docs/pr/<lot>/failing_first.md`. `make mutate` en plus, survivants lus un à un ; ceux des paquets `main` se rejouent à la main (`gremlins` n'y tue rien).
3. Avant la PR : `make prepr REPO=… BASE=…` (precommit, mutate, suite complète `-race`, éval lexicale), environ 25 min.
4. Le hook `pre-push` du clone (`make hooks`) exécute `precommit` à chaque push. Jamais de `--no-verify`.
5. **Toute PR s'ouvre en brouillon** (`gh pr create --draft`) et passe « ready for review » (`gh pr ready`) seulement quand le rapport Codecov de sa tête montre zéro ligne non couverte, ou exactement les lignes de `main` déclarées quand zéro n'est pas possible. Une PR à amender repasse en brouillon pendant la vérification.
6. Une poussée par lot, après la vérification complète, jamais au fil des rapports Codecov.
7. Code couvrable : `main` réduit à la répartition et à `os.Exit(xCommand(...))`, environnement lu dans des fonctions qui rendent une erreur, chemins d'erreur provoqués par des pannes SQLite déterministes (ligne à `id` NULL, déclencheur `RAISE(ABORT)`, colonne renommée), `rows.Err()` via `eachRow`, garde-fou injouable retiré.

## Commandes (dans `cortexdb/`)

- Compiler : `go build ./...` (Go 1.26, environ 25 s).
- Tests des paquets touchés : `go test -race ./pkg/authz/ ./pkg/httpapi/ ./pkg/rpcserver/ ./pkg/cortexdb/ ./cmd/cortexdb-mcp-stdio/` (environ 5 min, `pkg/cortexdb` seul prend 4 min ; vert le 18 octobre 2026).
- Tests rapides pendant le développement : `go test ./pkg/cortexdb/ -run '<Nom>'`.
- Avant un commit : `gofmt -l pkg cmd` et `go vet` sur les paquets touchés.
- Suivre l'amont : `git fetch origin && git rebase origin/main` sur la branche de travail.
- Comme en CI, avant toute PR : `go test -race -timeout 25m ./...` (environ 12 min, vert le 18 octobre 2026) et `go test ./pkg/eval -run TestLexicalRetrievalQuality -v`.

## Documents

- `docs/pipeline_pr.md` : pipeline interne avant toute PR (règles, outils, code couvrable, procédure, lecture de `gremlins`, pièges). À lire avant d'ouvrir une PR.
- `docs/revision_graphes_entreprise.md` : révision d'architecture (18 octobre 2026). Décision : CortexDB comme moteur, couche mince Thalamus (fédération, droits, audit, relecture). Évolution par forks successives et petites PR vers l'amont. Section 3.9 : granularité par collection et liens logiques entre cerveaux.

## Mémoire CortexDB

- Ce projet a son **propre cerveau CortexDB**, isolé du cerveau global : `.cortexdb/cortexdb.db`.
- Configuré dans `.claude/settings.local.json` via `CORTEXDB_PATH` (lu par le serveur MCP et les hooks du plugin au démarrage de la session).
- Le cerveau global `~/.cortexdb/cortexdb.db` n'est **pas** utilisé ici.
- Avant de répondre sur des décisions passées ou le contexte du projet : `knowledge_memory_recall`.
- Enregistrer les décisions, préférences et faits durables avec `memory_save` / `knowledge_save` (ids stables, contenu concis).
- `.cortexdb/` est un état local : ne pas le versionner.
