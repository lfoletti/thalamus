# Pipeline interne avant toute PR vers CortexDB

Procédure en vigueur depuis le 19 octobre 2026. Elle est née d'une soirée où six PR ont été amendées plusieurs fois sur rapports Codecov. Depuis qu'elle tourne, les six PR ouvertes sont à 100 % de couverture du patch, hors les instructions de `main`, et chaque poussée a été précédée de sa vérification.

## 1. Règles

1. Le code soumis est couvert à 100 % de ses lignes atteignables **avant le premier push**. La mesure est locale (`make patchcov`), pas le rapport Codecov.
2. **Le test s'écrit avant le code.** Il tombe sur l'arbre sans le correctif, puis le code le fait passer (« failing first »). Comme le relecteur ne voit pas cet ordre, la preuve livrée est une mutation à la main par règle, consignée dans `docs/pr/<lot>/failing_first.md` et résumée dans le corps de la PR.
3. Rien ne part sans le hook `pre-push`, qui exécute `make precommit` sur la copie poussée. Pas de `--no-verify`.
4. Une PR s'ouvre **en brouillon** (`gh pr create --draft`). Elle passe « ready for review » seulement quand le rapport Codecov de sa tête montre zéro ligne non couverte, ou exactement les lignes de `main` déclarées dans son corps quand zéro n'est pas possible.
5. Une PR déjà ouverte qu'il faut amender repasse en brouillon (`gh pr ready --undo`) le temps de la vérification, puis suit la règle 4.
6. Toute poussée se fait en une fois par lot, après la vérification complète (section 4), jamais au fil des rapports Codecov.

### Trois niveaux de vérification (décision de Lauro, 20 octobre 2026)

Le 20 octobre 2026, chaque trou trouvé après le lancement de la chaîne lourde (porte manquante sur `cortex_query`, fermeture MCP sans test) a rendu caduc un passage de 25 minutes par PR. La règle qui en sort : la vérification lourde se lance une fois, à la fin, sur des têtes gelées.

| Niveau | Quand | Quoi | Durée |
| :- | :- | :- | :- |
| 1. Boucle rapide | à chaque changement de code ou de test | `make precommit REPO=<copie> BASE=<base>` (gofmt, vet, `-race` et couverture du patch sur les paquets touchés), puis la mutation à la main de la règle changée, consignée dans `failing_first.md`. Un changement qui ne touche que des tests relance `patchcov` seul : les mutants visent le code. | 2 à 7 min |
| 2. Travail de tête, têtes gelées | avant de lancer le niveau 3 | Analyse, relecture des corps et des réponses, ADR, lecture des survivants déjà connus : tout ce qui peut révéler un trou se fait **avant** la chaîne, jamais pendant. Lancer la chaîne, c'est déclarer les têtes gelées. Un trou trouvé après va dans une liste ; la chaîne repart une fois, à la fin, pas à chaque constat. Pendant la chaîne, personne ne touche aux copies de travail qu'elle utilise. | sans machine |
| 3. Porte, une fois par lot | juste avant la poussée | `make prepr` (ou ses étapes à la main quand le bloc DDL arrête `precommit`) : `gremlins` sur le diff, suite complète `-race`, éval lexicale, sur la tête de chaque pile. Elle se relance seulement si le code change, jamais pour un corps de PR, une réponse ou un document. | environ 25 min par PR |

Entre deux amendements, la suite complète n'apporte rien qu'un changement confiné à un paquet puisse casser : les tests `-race` des paquets touchés (niveau 1) suffisent. Le disque conditionne le niveau 3 : avec moins de 10 Go libres, un `go clean -cache` entre deux suites coûte environ 2 min de recompilation chacune et interdit deux chaînes en parallèle.

## 2. Outils (racine de `granular_cortex`, `Makefile`)

| Commande | Ce qu'elle fait | Durée |
| :- | :- | :- |
| `make precommit REPO=<copie> BASE=<base>` | `gofmt -l`, `go vet` et `go test -race` des paquets touchés, puis couverture du patch. Sortie 1 au premier écart. | 1 à 7 min selon les paquets |
| `make patchcov REPO=… BASE=… [DECLARED=…]` | `python3 -I tools/patchcov.py` : chaque bloc changé que les tests n'exécutent pas. Les blocs du corps de `func main()` sont affichés à part et ne bloquent pas. `DECLARED=pkg/core/store_postgres.go` (ou `fichier:L1-L2`) déclare un bloc hors `main` que rien ne peut couvrir en local : affiché à part, sans bloquer, à nommer dans la section `## Codecov` de la PR. Le hook `pre-push` lit la variable dans l'environnement : `DECLARED=… git push …`. | environ 1 min |
| `make mutate REPO=… BASE=…` | `gremlins unleash --diff <base>` : mutation automatique des lignes changées. Rapport à lire, voir section 5. | 5 à 6 min pour un lot comme le lot C |
| `make prepr REPO=… BASE=…` | `precommit`, `mutate`, puis la suite complète `go test -race ./...` et l'éval lexicale, comme en CI. | environ 25 min |
| `make hooks` | Installe le hook `pre-push` dans le clone. Il vaut pour toutes les copies de travail du clone. | immédiat |

`REPO` est le clone ou la copie de travail (`git worktree`) qui porte la branche. `BASE` est la base du diff : `origin/main` pour une PR posée sur `main`, la branche du dessous pour une PR empilée. Le hook utilise toujours `origin/main`, ce qui mesure aussi les PR du dessous : elles doivent déjà être propres.

## 3. Écrire le code pour qu'il soit couvrable

- `main` ne contient que la ligne de répartition et l'appel à une fonction qui rend le code de sortie : `os.Exit(xCommand(os.Args[2:], os.Stdout, os.Stderr))`. La lecture de l'environnement est dans une fonction qui rend une erreur, jamais dans `main`.
- Chaque chemin d'erreur a son test, écrit avant lui, qui provoque la panne. Pannes déterministes disponibles en SQLite, sans faux pilote :
  - lecture d'une ligne qui échoue : une ligne à `id` NULL, insérée après `dropChangeLogTriggers` (voir `pkg/cortexdb/expiry_test.go`) ;
  - suppression qui échoue : `CREATE TRIGGER … BEFORE DELETE ON <table> BEGIN SELECT RAISE(ABORT, 'blocked'); END` ;
  - requête qui échoue alors que les autres marchent : `ALTER TABLE <table> RENAME COLUMN <col> TO <col>_hidden` ;
  - cerveau fermé : `db.Close()` puis l'appel ;
  - cerveau par défaut : `t.Setenv("HOME", t.TempDir())` et `CORTEXDB_PATH` vide.
- Un `rows.Err()` après la dernière ligne n'a pas de panne déterministe : passer par `eachRow(rows, each)` (`pkg/cortexdb/expiry.go`), qui le rend en dernière instruction.
- Un garde-fou qu'aucune entrée ne peut atteindre se retire, il ne se déclare pas.
- Un bras d'une `switch` sur les erreurs qui n'est pas atteignable depuis la route se sort dans une fonction pure testée sur chaque cas (`toolGuardStatus`, `pkg/httpapi/policy.go`).

## 4. Procédure par PR

1. Crée la branche du correctif depuis `origin/main`, ou depuis la branche du dessous si la PR est empilée. Un commit par correctif, sans entrée `CHANGELOG` (elle reste sur `proto/granular-cortex`).
2. Écris d'abord le test de la règle et lance-le : il doit tomber. Écris ensuite le code qui le fait passer. Même ordre pour chaque chemin d'erreur (section 3) : d'abord la panne provoquée et l'assertion, puis le code qui la traite.
3. Niveau 1 : lance `make precommit REPO=<copie> BASE=<base>`. Si un bloc hors `main` reste, ajoute le test ou retire le code. Note les blocs de `main` pour le corps.
4. Niveau 1 : pour chaque règle, rejoue une mutation à la main (harnais : `docs/pr/lot_C/mutate_lotC.py`), restaure l'arbre, consigne « mutation → test qui tombe » dans `failing_first.md`.
5. Niveau 2 : rédige le corps (étape 6), relis-le, fais l'analyse qui reste. Puis gèle les têtes et lance le niveau 3 : `make prepr REPO=<copie> BASE=<base>`, une fois par lot, tête de chaque pile. Lis les survivants de `gremlins` (section 5). Un trou trouvé pendant la chaîne va dans une liste ; la chaîne repart une fois, après la correction de toute la liste.
6. Rédige le corps dans `docs/pr/<lot>/<id>.body.md` : quoi, pourquoi, tests, section « Failing first », et, si des lignes de `main` restent, une section `## Codecov` après `## Breaking Changes`, de même rang, qui dit « The N lines left uncovered are inevitable. … », les nomme et dit pourquoi. Pas de pied de page (décision de Lauro du 19 octobre 2026 : ni `Co-Authored-By` dans les commits, ni ligne « Generated with » dans les corps).
7. Pousse vers le fork. Le hook tourne. Pour un amendement, `git push --force-with-lease=<branche>:<ancienne tête>`.
8. Ouvre la PR **en brouillon** : `gh pr create --draft --body-file …`.
9. Attends la CI et le rapport Codecov de cette tête. Si le rapport montre zéro ligne, ou exactement les lignes déclarées : `gh pr ready <n>`. Sinon, retour à l'étape 2, sans poussée intermédiaire.
10. Après la fusion amont : rebase les branches empilées dessus, pousse-les à travers le hook, reconstruis `proto/granular-cortex` avec les entrées `CHANGELOG`.

## 5. Lire un rapport `gremlins`

- `KILLED` : la règle a un test.
- `NOT COVERED` : à comparer avec `patchcov` ; attendu seulement dans `main` et sur les constantes.
- `LIVED` dans un paquet `main` (`cmd/…`) : `gremlins` ne tue rien dans ces paquets, même quand les tests tuent la mutation. Rejoue chaque survivant à la main dans une copie de travail séparée.
- `LIVED` ailleurs : lis la ligne. Trois issues : un vrai manque (ajoute le test), un équivalent sans effet observable (durée d'un cache, capacité d'une map, `len > 0` sur un vide : classe-le dans `failing_first.md`), une ligne amont voisine du patch (hors périmètre).

## 6. Pièges rencontrés

- Codecov mesure le diff empilé : une PR posée sur une autre affiche aussi les lignes de celle du dessous.
- `zsh` ne découpe pas `set -- $var` : utiliser `while read -r a b c`.
- Un `| tail` cache le code de sortie de `go test` : toujours `|| { echo FAILED; exit 1; }`.
- Un `#` dans `$(shell …)` d'un Makefile est un commentaire : utiliser `|` comme séparateur de `sed`.
- `go test -run` est sensible à la casse : vérifier que le filtre retient bien le test visé.
- Le disque : `go clean -cache` avant la suite complète, et une seule suite à la fois.
- Un script de mutation réécrit les fichiers de l'arbre : jamais en arrière-plan pendant un `git add` ou un `commit --amend` du même worktree. Le 20 octobre 2026, une mutation (`|| true`) a été commise ainsi et détectée seulement par les tests de la branche empilée dessus.
- `git stash -u` puis `rebase` puis `stash pop` : vérifier que les fichiers non suivis sont revenus (un fichier de test est resté dans le stash le 20 octobre 2026).
- Un bloc hors `main` que rien ne peut couvrir en local (liste DDL PostgreSQL de `pkg/core/store_postgres.go`) arrêtait `precommit`, donc `prepr` et le hook. Depuis le 20 octobre 2026 au soir : `DECLARED=pkg/core/store_postgres.go make prepr …` et `DECLARED=… git push …` le déclarent, et la PR le nomme dans `## Codecov`. Avant cette option, il fallait lancer les étapes suivantes à la main : `make mutate`, puis `go clean -cache && go test -race -count=1 -timeout 25m ./...` et `go test -count=1 ./pkg/eval -run TestLexicalRetrievalQuality`, et déclarer le bloc dans `## Codecov`.
- `go test … | tail` cache le code de sortie : lire les lignes `FAIL` du fichier de sortie, pas le code du pipeline. Le 20 octobre 2026 au soir, une suite `-race` passée en arrière-plan a rendu 0 avec un paquet en échec (un nouvel outil absent de `toolWrites` dans `tool_mutates_test.go`).
- `git cherry-pick` n'a pas d'option `-q` : la commande échoue en affichant l'usage. Dans une chaîne `&&` suivie d'un `;`, la suite s'exécute quand même et peut amender le mauvais commit. Reconstruire `proto` en étapes vérifiées, un `echo` par étape, et garder un tag de sauvegarde avant le `reset --hard`.
- Un nouvel outil dans `ToolDefinitions()` doit être ajouté à `toolWrites` (`pkg/cortexdb/tool_mutates_test.go`), à `toolCount` (même fichier), et, sur `proto`, à la liste blanche ou à `refusedTools` de `pkg/authz/allowlist.go` (test de complétude).
- Lancer la chaîne lourde avant la fin de l'analyse (20 octobre 2026 au soir : deux relances de 50 min pour deux trous trouvés après le lancement). Les trois niveaux de la section 1 règlent l'ordre : tête d'abord, chaîne ensuite, une fois.
