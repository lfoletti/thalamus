# Note pour Claude Opus 5.5 : amendement des PR #72 à #77, second tour

Rédigée le 21 octobre 2026 par la session Fable 5.1 qui a analysé les revues. Tu développes **d'une seule traite, sans intervention de Lauro** : tu ne poses aucune question, tu décides selon `plan.md`, tu consignes chaque écart dans `REPORT.md`. Tu t'arrêtes quand tout est développé, testé, mesuré et documenté. La poussée et les commentaires GitHub restent à Lauro.

## 1. Mission

Amender en place les six PR CortexDB (#72, #73, #74, #75, #76, #77) selon `docs/pr/round2/plan.md`, section 2, avec les décisions D1 à D9 de sa section 1. Livrer : les six branches amendées, `proto/granular-cortex` reconstruite, le gateway Thalamus adapté, les corps de PR, les réponses au relecteur, la preuve « failing first », le rapport.

## 2. Règles non négociables

1. **Aucun `git push`, aucun `gh pr …` qui écrit, aucun commentaire GitHub.** Lecture (`gh pr view`, `gh api … GET`) permise.
2. **Jamais `--no-verify`.** Le hook `pre-push` n'est pas concerné puisque tu ne pousses pas, mais la règle vaut pour tout.
3. **Le test s'écrit avant le code et tombe sans lui.** Pour chaque règle : un test nommé, une mutation à la main qui le fait tomber, consignée dans `docs/pr/round2/failing_first.md` (modèle : `docs/pr/lot_C/failing_first.md`). « Les tests passent après » ne suffit jamais.
4. **Commits** : un commit par PR (`commit --amend` sur la branche), auteur `git -c user.name="Lauro Foletti" -c user.email="25608942+lfoletti@users.noreply.github.com"`, message conventional-commit, **sans** ligne `Co-Authored-By`, sans « Generated with ». Les corps de PR n'ont pas de pied de page non plus.
5. **Pas d'entrée `CHANGELOG.md` sur les branches de PR** ; les entrées `Unreleased` vivent sur `proto/granular-cortex` seulement.
6. **Taille** : viser 400 lignes hors tests par PR. Le propriétaire a toléré 727 (#72), 505 (#76), 451 (#77) avec le compte dit dans le corps. Dire le compte.
7. **Canari** : tous les tests tournent avec `CORTEXDB_REMOTE=127.0.0.1:1 CORTEXDB_GRPC_TOKEN=canary CORTEXDB_PATH=/dev/null/canary` (`make precommit`, `make prepr`, `make diffmain` le font). Un test qui lit l'environnement de la machine est un défaut, quel que soit son résultat.
8. **Niveau 2 avant le niveau 3** : inventaires, scénarios différentiels, promesses, corps, tout ce qui peut révéler un trou, **avant** `make prepr`. Lancer `prepr`, c'est geler les têtes. Un trou trouvé pendant la chaîne va dans une liste ; la chaîne repart une fois, à la fin. Jamais deux chaînes lourdes en parallèle sur la même copie, et `go clean -cache` avant une suite complète si le disque a moins de 10 Go libres.
9. **Pas de commentaire sur les outils du propriétaire** (Codecov, CI) dans les textes qui lui sont destinés.
10. Les textes destinés au propriétaire sont en anglais, factuels, point par point. Les documents du projet sont en français.

## 3. Où sont les choses

| Quoi | Où |
| :- | :- |
| Plan, décisions, revues intégrales | `docs/pr/round2/plan.md` |
| Réponses au relecteur (brouillons à compléter, crochets) | `docs/pr/round2/review_replies.md` |
| Pipeline (règles, outils, procédure, pièges) | `docs/pipeline_pr.md` ; résumé dans `CLAUDE.md` |
| Outils | `Makefile` à la racine : `precommit`, `patchcov`, `mutate`, `prepr`, `inventory`, `diffmain` ; `tools/patchcov.py`, `tools/inventory.py`, `tools/diffmain.py` |
| Scénarios différentiels existants | `docs/pr/lot_C/scenarios/pkg/cortexdb/reingest_scenario_test.go` |
| Dossiers des lots (README, corps, failing_first, harnais de mutation) | `docs/pr/lot_B/`, `docs/pr/lot_C/` |
| Clone CortexDB, branche `proto/granular-cortex` | `cortexdb/` (dépôt à part, ignoré par la racine) |
| Copies de travail des six branches | `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/395d9eca-adec-4b02-abda-51fc00b17277/scratchpad/wt/<branche>` |
| Gateway Thalamus (à adapter, plan section 3, étape 6) | `gateway/` (`README.md`, `relay/`, `mutate.py`, `docs/gateway/failing_first.md`), bancs `tests/e2e_gateway.sh`, `tests/e2e_shared_brain.sh` |
| Sauvegardes | tags `backup/<branche>-2026-10-21` dans `cortexdb/` |

Si les copies de travail n'existent plus (le dossier `scratchpad` d'une autre session peut être nettoyé) : `cd cortexdb && git worktree prune && git worktree add <ton scratchpad>/wt/<branche> <branche>` pour chacune des six. Ne jamais faire `checkout` d'une branche de PR dans `cortexdb/` même : ce clone porte `proto/granular-cortex`.

## 4. Démarrage

1. Lis `plan.md` en entier, puis `docs/pipeline_pr.md`, puis `docs/pr/lot_B/README.md` et `docs/pr/lot_C/README.md`, puis les deux `failing_first.md` et `review_replies.md` des lots (le style attendu).
2. Vérifie l'état : dans chaque copie, `git log --oneline -3`, `git status --short` (propre), et `git merge-base --is-ancestor origin/main HEAD`.
3. Lance une fois `make inventory REPO=<copie de gc-06-review-status> BASE=origin/main > docs/pr/round2/inventory_lotC.md` et la même chose pour `gc-04c-rest-confined-tools` (lot B). Lis les tables : ce sont tes listes de contrôle.
4. Lance `make diffmain REPO=<copie de gc-06-review-status> BASE=origin/main SCENARIOS=docs/pr/lot_C/scenarios` pour voir le diff de départ (quatre écarts, tous des défauts à corriger). À la fin, le même diff ne doit contenir que des lignes nommées dans les corps.
5. Commence par #72 (point 1, puis 2), puis #73, #74, puis #75, #76, #77, dans l'ordre du plan, section 3.

## 5. Vérification, par PR puis par lot

Par point : test écrit, il tombe ; code ; `make precommit REPO=<copie> BASE=<base>` ; mutation à la main (restaure l'arbre après) ; ligne dans `failing_first.md`.

Par PR, niveau 2 : table « élément de la classe → test ou raison » complète ; promesses du diff (`make inventory --class promises`) chacune avec son test ; corps rédigé.

Par lot, niveau 3, têtes gelées : `make prepr REPO=<copie> BASE=<base>` (lot C : `DECLARED=pkg/core/store_postgres.go make prepr …` sur 05 et 06 ; si le bloc DDL arrête `precommit` malgré tout, les étapes à la main sont dans `docs/pipeline_pr.md`, section 6). Survivants `gremlins` lus un à un et classés (section 5 du pipeline). `make diffmain` sur les têtes de pile. Puis `proto` reconstruite, `go build ./...`, `go test -race -count=1 -timeout 25m ./...` sous le canari, `sh tests/e2e_shared_brain.sh` depuis la racine (30 contrôles), `gateway/` : `go test -race ./...`, `python3 -I mutate.py`, `sh tests/e2e_gateway.sh` (55 contrôles). Tout résultat, vert ou non, va dans `REPORT.md` avec les lignes `FAIL` quand il y en a. Ne jamais lire le code de sortie d'un `go test … | tail` : lire les lignes `FAIL`.

## 6. Livrables (tous locaux)

- Six branches amendées, un commit chacune, rebasées en pile (04a → 04b, 04c ; 08 → 05 → 06), `git status` propre dans chaque copie.
- `proto/granular-cortex` = `origin/main` + les six commits + le commit de classement lot C + les entrées `Unreleased`.
- `gateway/` adapté, ses tests, ses mutations et ses bancs verts, `gateway/README.md` et `CLAUDE.md` à jour.
- `docs/pr/round2/` : `GC-04a.body.md`, `GC-04b.body.md`, `GC-04c.body.md`, `GC-08.body.md`, `GC-05.body.md`, `GC-06.body.md` (corps complets, prêts pour `gh pr edit --body-file`), `review_replies.md` complété, `failing_first.md`, `inventory_*.md`, `REPORT.md`.
- `CLAUDE.md` (section « État de la branche »), `docs/pr/lot_B/README.md`, `docs/pr/lot_C/README.md` : têtes, tailles, état « amendé en local, non poussé ».
- Tags `backup/<branche>-2026-10-21-amended` sur les têtes finales.

## 7. Quand le plan ne suffit pas

Décide, puis consigne dans `REPORT.md`, section « Écarts au plan » : le point, les options, ce que tu as choisi, pourquoi. Critères, dans l'ordre : le comportement de `main` gagne sauf ligne nommée dans le corps ; la règle la plus simple qui ferme la classe entière ; la taille. Si un point ne peut pas être fait sans casser une règle de la section 2, ne le fais pas, fais tout le reste, et dis-le en tête de `REPORT.md`. Ne t'arrête pas sur une erreur : corrige, relance, consigne.

## 8. Ce qui reste à Lauro après toi

La lecture de `REPORT.md`, puis la commande de poussée (procédure dans `docs/pr/lot_B/README.md` et `docs/pr/lot_C/README.md` : `gh pr ready --undo`, `git push --force-with-lease=<branche>:<ancienne tête> fork <branche>` par branche, `gh pr edit --body-file`, un commentaire par PR, commentaire de lot sous #66, `gh pr ready` quand Codecov montre zéro ligne hors celles déclarées).
