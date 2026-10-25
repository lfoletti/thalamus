# Thalamus

Couche mince au-dessus de CortexDB pour des graphes de connaissances d'entreprise utilisés par des agents de code : fédération de cerveaux, droits, audit, relecture. Le thalamus relaie et filtre tout ce qui monte au cortex. Préfixe d'outils `th_`, binaires `th-gateway` et `th-hook`, module `github.com/lfoletti/thalamus/gateway`, plugin `thalamus`, clés `th/<login>/<cerveau>`.

La racine est le dépôt git `lfoletti/thalamus` (**public**, MIT, branche `main`, remote `origin`). Le dossier sur disque est `thalamus` (`granular_cortex` jusqu'au 25 octobre 2026). `cortexdb/` est un dépôt à part, ignoré par la racine.

## Rôle de ce fichier

Ce fichier est technique : règles permanentes, arborescence, commandes, accès à la mémoire. Il est public, comme le dépôt.

- **L'état, l'historique, les décisions et les leçons sont dans le cerveau `coding`**, pas ici. Un fait a une seule place.
- **Au début de chaque session**, lire l'état courant : `th_get` avec `brain: "coding"` et `knowledge_id: "state:thalamus:current"`.
- N'écrire ici aucun état (têtes, numéros de PR, comptes de contrôles, dates d'envoi) et aucun contenu privé.

## Arborescence

- `docs/` : documents du projet.
- `gateway/` : module Go (CortexDB en bibliothèque, version dans `gateway/go.mod` ; SDK MCP). Paquets `rights` (lecture de `rights.yaml`), `identity` (vérification d'un `id_token` OIDC, sans bibliothèque), `keys` (émission et révocation des clés confinées dans le fichier de clés du serveur), `relay` (serveur MCP HTTP par cerveau, relais gRPC avec la clé de l'appelant, `audit.jsonl`), `login` (côté employé de `/v1/login` : code + PKCE, boucle locale), `version` (la release du binaire, étiquette du build), `cmd/th-gateway`, `cmd/th-hook`. Lire `gateway/README.md`.
- `plugin/` : plugin Claude Code `thalamus` (hook de rappel, serveur MCP du gateway, compétence). Lire `plugin/README.md`.
- `extension/` : extension Chrome (Manifest V3, sans bibliothèque) qui propose la page ou la sélection par `POST /v1/add`. Règles dans `lib.js` (tests `node --test`, mutations dans `gateway/mutate.py` sous le paquet `extension`), pilote Playwright `e2e_driver.mjs`, `node_modules/` ignoré (`npm install` une fois, Playwright 1.58.2 = Chromium 1208 en cache). Lire `extension/README.md`.
- `.github/workflows/ci.yml` : la CI (gateway, extension, banc du gateway, scan des secrets, tests Python de `tools/`), sur chaque PR et sur `main`.
- `tools/hooks/` : les hooks du dépôt, `pre-commit` (aucun commit sur `main`) et `pre-push` (termes privés, tests des parties touchées), testés par `tools/test_hooks.py`. `CONTRIBUTING.md` et `SECURITY.md` : règles publiques de contribution et de signalement.
- `tests/` : bancs de bout en bout.
  - `tests/e2e_gateway.sh` : le gateway devant un serveur gRPC construit depuis `cortexdb/` (environ 1 min).
  - `tests/e2e_plugin.sh [modèle]` : une session Claude Code réelle avec le plugin. Ce banc coûte des appels au modèle et reste hors chaîne.
  - `tests/e2e_extension.sh` : l'extension dans Chromium devant un vrai gateway et le faux fournisseur OIDC `tests/fake_idp.mjs` (environ 30 s).
  - `tests/e2e_shared_brain.sh` : le clone seul (serveur gRPC, fichier de clés, client MCP distant, environ 2 min).
  - `tests/eval_ch/` : banc multilingue sur la Constitution fédérale (lire son `README.md`).
- `cortexdb/` : clone complet de `github.com/liliang-cn/cortexdb` (amont `origin`, fork `fork` = `lfoletti/cortexdb`). Une branche par PR (`gc-NN-<sujet>`). Branche d'intégration `proto/granular-cortex`. Avant un rebasage ou une reconstruction, poser un tag `backup/<branche>-<date>`.

## Règles de publication

- **Aucun push, aucune pull request, aucune issue** vers l'amont, le fork ou ce dépôt sans demande explicite de Lauro. La soumission se fait par petits lots.
- **Rien de privé dans un dépôt public.** Une faille signalée en privé à un tiers ne paraît dans aucun commit, corps de PR ou fichier suivi, avant que le propriétaire le permette. Avant un push de ce dépôt, relire les commits non poussés pour cette règle.
- **Ce dépôt** : `main` est protégée (PR obligatoire, les 5 jobs CI verts, 1 approbation, pas de force push). Une PR de Lauro se fusionne avec `gh pr merge --admin` après une CI verte. Hooks : `git config core.hooksPath tools/hooks`, chemin **relatif**, jamais absolu (il se résout dans chaque copie de travail et survit à un déplacement du dossier) ; termes privés dans `.git/info/private-terms` (jamais suivi). Bonne pratique pour tout dépôt public : `practice:github:public-repo-basics` dans `coding`.
- **Une copie de travail par session.** La copie principale reste sur `main` et ne reçoit aucun commit : le hook `pre-commit` le refuse. Chaque session travaille dans son propre `git worktree`, sur sa propre branche créée depuis `origin/main` : `git worktree add -b <branche> ../thalamus_wt/<branche> origin/main`. Jamais dans la copie d'une autre session. Un commit n'entre dans `main` que par sa propre PR. Après la fusion : `git worktree remove ../thalamus_wt/<branche>`, puis `gh pr merge --delete-branch`.
- Avant de lire `origin/main`, faire `git fetch`.
- Tous les correctifs se développent **localement**, **un commit par correctif**. Une PR ne touche pas `CHANGELOG.md` : le propriétaire amont l'écrit seul à la release (`docs/plan_publication.md`, section 0 bis).
- Le produit final se teste en entier avant toute soumission.
- **Tests « failing first »** (exigence du propriétaire amont) : chaque test d'une PR échoue sans son code. Vérifier avec les fichiers de test posés sur l'arbre sans le correctif, ou par une mutation de chaque règle si les tests ne compilent pas sans elle. Preuve dans le corps de la PR, section « Failing first ». Méthode et exemple : `../thalamus_local/docs/pr/lot_B/failing_first.md`.
- **PR en brouillon d'abord** : une PR s'ouvre en brouillon et passe à la relecture seulement après un rapport Codecov à zéro ligne non couverte sur sa tête (ou les seules lignes de `main` déclarées). Voir le pipeline.
- Auteur des commits : `git -c user.name="Lauro Foletti" -c user.email="25608942+lfoletti@users.noreply.github.com" commit` (identité GitHub, jamais l'adresse personnelle, ni dans un commit ni dans une PR), message en style conventional-commit, **sans** ligne `Co-Authored-By` (tous les commits de ce dépôt et des PR amont).
- Dans les textes publics, ne jamais juger CortexDB ni écrire ce qui lui manque. Dire ce que Thalamus ajoute, et décrire une PR amont comme une fonction proposée.

## Pipeline avant toute PR (procédure complète : `docs/pipeline_pr.md`)

**Trois niveaux** (`docs/pipeline_pr.md`, section 1) :
- Niveau 1, à chaque changement : `make precommit` et la mutation à la main de la règle changée.
- Niveau 2, **avant** la chaîne lourde : analyse, relecture, tout ce qui peut révéler un trou. Cinq contrôles par classe : canari dans `precommit` (aucun test n'atteint un serveur, un embedder ni le cerveau de la machine) ; `make inventory` et `make invcheck FF=<inventaire>` ; `make diffmain SCENARIOS=../thalamus_local/docs/pr/<lot>/scenarios` ; chaque promesse d'un schéma ou d'un corps tenue par un test nommé ; relecture adverse par un agent neuf sur les têtes gelées (invite fixe, `docs/pipeline_pr.md`, section 1 bis). Ensuite, les têtes sont gelées et aucune copie de travail de la chaîne n'est touchée.
- Niveau 3, juste avant la poussée : `make prepr` par tête (avec `make clock`), puis `make benches` une fois par lot sur `proto` (`make level3` enchaîne les deux). Relancer seulement si le code change. Un trou trouvé pendant la chaîne va dans une liste, et la chaîne repart une fois, à la fin.
- Mutation à la main par `make handmut` (restauration depuis une copie, refus d'un fichier non commis), jamais par `git checkout`.

Règles :
1. Le patch est couvert à 100 % de ses lignes atteignables **avant le premier push**, mesuré en local : `make precommit REPO=<copie> BASE=<base>` (gofmt, vet, `-race` des paquets touchés, `patchcov`). Les blocs de `func main()` sont affichés à part : les déclarer dans une section `## Codecov` du corps, après `## Breaking Changes` et de même rang. Une PR empilée vise `main` : mesurer contre `main`, sans `CORTEXDB_TEST_POSTGRES`.
2. **Le test s'écrit avant le code** et tombe sans lui. Preuve : une mutation à la main par règle qui fait tomber un test nommé, dans `../thalamus_local/docs/pr/<lot>/failing_first.md`. `make mutate` en plus, survivants lus un à un ; ceux des paquets `main` se rejouent à la main.
3. Avant la PR : `make prepr REPO=… BASE=…` (precommit, mutate, suite complète `-race`, éval lexicale), environ 25 min.
4. Le hook `pre-push` du clone (`make hooks`) exécute `precommit` à chaque push. Jamais de `--no-verify`.
5. **Toute PR s'ouvre en brouillon** (`gh pr create --draft`) et passe « ready for review » (`gh pr ready`) seulement quand Codecov montre zéro ligne non couverte, ou exactement les lignes de `main` déclarées. Une PR à amender repasse en brouillon pendant la vérification.
6. Une poussée par lot, après la vérification complète, jamais au fil des rapports Codecov. Procédure : `gh pr ready --undo`, `git push --force-with-lease fork <branche>`, `gh pr edit --body-file`, commentaire de réponse, puis `gh pr ready`. Les nombres du corps viennent d'un outil (`make bodynums`), jamais d'une copie à la main.
7. Forme de codage d'une tranche (`docs/pipeline_pr.md`, section 3 bis) : noyau pur puis une seule ligne d'E/S, test de classe d'abord, un seul point de contrôle par règle, corps écrit depuis les noms des tests, environ 400 lignes hors tests.
8. Code couvrable : `main` réduit à la répartition et à `os.Exit(xCommand(...))`, environnement lu dans des fonctions qui rendent une erreur, chemins d'erreur provoqués par des pannes SQLite déterministes (ligne à `id` NULL, déclencheur `RAISE(ABORT)`, colonne renommée), `rows.Err()` via `eachRow`.
9. Ne jamais lancer deux suites avec `CORTEXDB_TEST_POSTGRES` en même temps (`max_connections` 100, faux échecs `53300`). Après un banc, retirer les copies `tmp.*` de `$TMPDIR`.

Les mêmes règles de vérification s'appliquent au gateway (`gateway/mutate.py`, couverture 100 % hors les `main`).

## Commandes

Dans `cortexdb/` :
- Compiler : `go build ./...` (Go 1.26, environ 25 s).
- Tests des paquets touchés : `go test -race ./pkg/authz/ ./pkg/httpapi/ ./pkg/rpcserver/ ./pkg/cortexdb/ ./cmd/cortexdb-mcp-stdio/` (environ 5 min, `pkg/cortexdb` seul prend 4 min).
- Tests rapides : `go test ./pkg/cortexdb/ -run '<Nom>'`.
- Avant un commit : `gofmt -l pkg cmd` et `go vet` sur les paquets touchés.
- Suivre l'amont : `git fetch origin && git rebase origin/main` sur la branche de travail.
- Comme en CI : `go test -race -timeout 25m ./...` (environ 12 min) et `go test ./pkg/eval -run TestLexicalRetrievalQuality -v`.

Dans `gateway/` : `go test -race ./...` et `python3 -I mutate.py`. Depuis la racine : `sh tests/<banc>.sh`.

## Documents

- `docs/pipeline_pr.md` : pipeline interne avant toute PR. À lire avant d'ouvrir une PR.
- `docs/plan_publication.md` : plan de publication vers l'amont.
- `docs/upstream/` : fiches des PR amont potentielles. Rien n'est envoyé sans commande explicite.
- Notes de travail d'un lot (plans, corps, preuves « failing first », scénarios, relectures adverses) : dans l'archive locale hors du dépôt, `../thalamus_local/docs/pr/<lot>/`, jamais publiée (règle de Lauro, 25 octobre 2026). `../thalamus_local/docs/pr/` garde les lots plus anciens, déjà suivis.
- `docs/revision_graphes_entreprise.md` : architecture (CortexDB comme moteur, couche mince Thalamus). Section 3.9 : granularité par collection et liens logiques entre cerveaux.

## Mémoire du projet : le cerveau partagé `coding`

- Le cerveau du projet est le cerveau partagé **`coding`** du gateway Thalamus (`https://th.contaixt.ai`). Il n'a **pas** de copie locale : un cerveau partagé ne reste pas local.
- Accès par le plugin `thalamus` (marketplace locale `.claude-plugin/marketplace.json`, portée `local`). Le plugin `cortexdb` est désactivé dans ce projet, et le cerveau global `~/.cortexdb/cortexdb.db` n'est pas utilisé ici.
- Réglages dans `.claude/settings.local.json` : `TH_GATEWAY`, `TH_RECALL_FILE` (`~/.thalamus/recall_thalamus`, `only=coding`), `TH_RECALL_TIMEOUT`. La clé `TH_KEY` est dans `~/.thalamus/env` (0600) : la session se lance avec `set -a; . ~/.thalamus/env; set +a; claude`, sinon le serveur MCP du plugin n'a pas de clé. Ne jamais afficher la clé.
- Lecture : l'état courant par `th_get` sur `state:thalamus:current`. Pour une décision passée ou un contexte : `th_recall` avec `brains: ["coding"]`, puis `th_get` pour lire un document entier. Le hook `UserPromptSubmit` sert aussi des extraits, mais il peut manquer un fait : il cherche par les mots.
- Écriture : `th_add` avec `brain: "coding"`, un identifiant stable (`document_id`, par exemple `decision:thalamus:<sujet>`), un titre et un contenu concis. Pour corriger un contenu, réécrire le même `document_id`. La clé de Lauro a le droit `curate` : ses ajouts sont acceptés tout de suite.
- `coding` n'a pas d'entités ni de relations : une clé confinée ne peut pas en créer. Mettre les liens entre notes dans le texte (identifiants, `[[nom]]`).
- Lauro est le seul lecteur de `coding`, qui contient des contenus privés. Avant d'ouvrir `coding` à d'autres personnes, déplacer ces contenus dans un cerveau privé.
- Gateway arrêté : pas de rappel, et une écriture échoue. Le dire, et réécrire le contenu quand le gateway répond.
- **Mise à jour automatique** (règle de Lauro, 24 octobre 2026) :
  1. Avant de donner un état ou des étapes du projet, lire `state:thalamus:current`, puis `th_recall` si nécessaire.
  2. Quand une vérification (GitHub, branches, fichiers, bancs) montre un état différent de `coding`, mettre à jour `coding` dans la même réponse, sans attendre une demande : réécrire `state:thalamus:current`, et le document concerné s'il existe. Puis dire ce qui a changé.
  3. Modifier ce fichier seulement quand une règle, une commande ou l'arborescence change. Le commit attend la commande de Lauro.
