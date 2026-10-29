# Pipeline interne avant toute PR vers CortexDB

Procédure en vigueur depuis le 19 octobre 2026. Elle est née d'une soirée où six PR ont été amendées plusieurs fois sur rapports Codecov. Depuis qu'elle tourne, les six PR ouvertes sont à 100 % de couverture du patch, hors les instructions de `main`, et chaque poussée a été précédée de sa vérification.

## 1. Règles

1. Le code soumis est couvert à 100 % de ses lignes atteignables **avant le premier push**. La mesure est locale (`make patchcov`), pas le rapport Codecov.
2. **Le test s'écrit avant le code.** Il tombe sur l'arbre sans le correctif, puis le code le fait passer (« failing first »). Comme le relecteur ne voit pas cet ordre, la preuve livrée est une mutation à la main par règle, consignée dans `../thalamus_local/docs/pr/<lot>/failing_first.md` et résumée dans le corps de la PR.
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
| 3. Porte, une fois par lot | juste avant la poussée | `make prepr` (ou ses étapes à la main quand le bloc DDL arrête `precommit`) : `gremlins` sur le diff, rejeu `-race -count=10` des tests du patch qui lisent l'heure (`make clock`), suite complète `-race`, éval lexicale, sur la tête de chaque pile. Puis, une fois par lot, les deux bancs de bout en bout sur `proto` reconstruite (`make benches` ; `make level3` enchaîne `prepr` et les bancs). Elle se relance seulement si le code change, jamais pour un corps de PR, une réponse ou un document. | environ 25 min par PR, plus 3 min de bancs par lot |

Le 21 octobre 2026, le niveau 3 de l'amendement du second tour a trouvé quatre trous que les niveaux 1 et 2 ne voyaient pas, d'où les trois ajouts ci-dessus : les bancs ont trouvé un circuit entier refusé (`th_pending` sous la règle du préfixe `_`) et un banc resté sur l'ancienne API ; `gremlins` un survivant réel sur `knowledge_save` ; la suite `-race` un test qui lisait l'horloge deux fois (1 échec sur 12). Un quatrième ajout tient à la méthode : la mutation à la main passe par `make handmut`, qui restaure le fichier depuis une copie et refuse un fichier non commis, parce qu'un `git checkout -- <fichier>` après une mutation a retiré un correctif non commis.

### Niveau 2 : cinq contrôles par classe (décision de Lauro, 21 octobre 2026)

Le 21 octobre 2026, le second tour de revue des PR #72 à #77 a posé 14 points. Onze étaient trouvables avec ce que le dépôt contenait déjà. Le schéma : au tour 1, le relecteur nomme des cas ; nous corrigeons les cas nommés ; au tour 2, il teste la classe entière de chaque cas et compare avec `main`. Les contrôles ci-dessous font ce travail avant la poussée. Le cinquième est venu le soir même : une relecture adverse des têtes amendées a trouvé 28 points, dont 10 « blocker » ou « likely comment », que les quatre premiers ne voyaient pas (les classes « entrées exportées » et « RPC typés » manquaient à l'inventaire, et rien ne comparait l'inventaire à `failing_first.md`). Ils s'ajoutent au niveau 2, têtes gelées ensuite.

| Contrôle | Règle | Outil | Ce qu'il aurait trouvé le 21 octobre 2026 |
| :- | :- | :- | :- |
| 1. Canari | Aucun test n'atteint un serveur ni le cerveau de la machine. `make precommit` et `make prepr` lancent les tests avec `CORTEXDB_REMOTE=127.0.0.1:1`, un jeton factice et `CORTEXDB_PATH=/dev/null/canary`. Un test qui lit l'environnement au lieu de le poser tombe sur « connection refused ». Un test d'un paquet `cmd/` pose chaque variable qu'il lit (`t.Setenv`), sur le modèle de `doctor_test.go`. Depuis le soir du 21 octobre 2026, le canari pose aussi l'embedder (`CORTEXDB_EMBED_BASE_URL`, `OPENAI_BASE_URL` sur un port fermé, `OPENAI_API_KEY` vide : `TestMappingFromDDLWithLLM_Live` de `main` est un test réel ouvert par cette clé, et une clé factice le fait tomber, trouvé par `prepr` le 21 octobre 2026), le modèle de réécriture des requêtes (`CORTEXDB_LLM_BASE_URL` sur un port fermé) et un seuil de rappel invalide (`CORTEXDB_MEMORY_SEMANTIC_FLOOR=canary`). `main` (`cc98c6c`) reste vert sous ce canari, suite complète comprise. | `CANARY` dans le `Makefile`, automatique | #76, point 3 : `sweep_expired_test.go` vidait `CORTEXDB_PATH` mais pas `CORTEXDB_REMOTE`, et le test a envoyé un balayage à vide au cerveau partagé du relecteur. Rejoué le 21 octobre 2026 : le test tombe sous le canari. |
| 2. Inventaire par classe | Chaque point de revue, et chaque règle nouvelle, nomme sa classe : outils, chemins de lecture, chemins d'écriture, champs d'identifiant, clés réservées, entrées exportées de la bibliothèque, RPC typés. `make inventory` imprime la liste complète de chaque classe depuis l'arbre. Dans l'inventaire du lot (ou `failing_first.md`), chaque élément a une ligne : le nom du test qui le tient, ou « hors classe : <raison> ». `make invcheck` compare les deux et sort 1 par élément sans ligne ou sans test ; `make prepr` le lance en premier, donc un élément sans ligne arrête le niveau 3. | `make inventory REPO=… BASE=…`, `make invcheck REPO=… BASE=… FF=<inventaire>` (`tools/inventory.py --check`) | #76 et #77 : `search_vector_range`, `aggregate_metadata`, `representative_records`, `knowledge_memory_neighbors` sans porte (classe « outils ») ; #77, point 2 : `memory_save` sur un identifiant existant (classe « écritures ») ; #72, point 1 : `memory_get`, `knowledge_get`, les listes (classe « identifiants ») ; #72, point 2, et #75, point 4 : les clés réservées écrasées par `metadata` et l'inverse (classe « clés »). |
| 3. Scénarios différentiels | Le relecteur rejoue chaque scénario sur `main` et sur la branche, puis compare. `make diffmain` fait la même chose : des tests de scénario, posés sur les deux arbres, qui n'affirment rien et écrivent une ligne `SCENARIO …` par observation ; le diff des deux journaux est la sortie. Chaque ligne du diff est voulue et nommée dans le corps de la PR, ou c'est un défaut. Les scénarios n'utilisent que l'API de `main` et vivent dans `../thalamus_local/docs/pr/<lot>/scenarios/<paquet>/`. Scénarios obligatoires pour un changement d'écriture : la même entrée deux fois, l'entrée changée d'un seul champ, la suppression puis la même entrée, une base écrite par `main` puis modifiée par la branche. | `make diffmain REPO=… BASE=… SCENARIOS=../thalamus_local/docs/pr/<lot>/scenarios` (`tools/diffmain.py`, environ 3 min) | #75, points 1 et 2 (la ré-ingestion du titre seul retire les entités, « unchanged » ne reconstruit pas le graphe) ; #76 et #77, point 1 (une clé de métadonnées du client fait échouer l'écriture). Rejoué le 21 octobre 2026 sur `036d1eb` contre `origin/main` : les quatre écarts sortent dans le diff (`../thalamus_local/docs/pr/lot_C/scenarios/`). |
| 4. Promesses tenues | Toute phrase nouvelle dans un schéma d'outil, un corps de PR ou le changelog qui affirme un comportement reçoit un nom de test dans `failing_first.md`. `make inventory` imprime les descriptions ajoutées par le diff (classe `promises`). | `make inventory` | #77, point 4 : les schémas de `memory_save` et des entités promettent un filtrage des entités et des relations qui n'existe pas ; #72 : « a foreign id and a missing id both come back "not found" » tenu pour `supersedes` seul. |
| 5. Relecture adverse | Sur les têtes gelées, après les contrôles 1 à 4 et avant la chaîne lourde : un agent neuf (`upstream-reviewer`, outils de lecture seulement, sans cerveau du projet, aucune note de développement en entrée, kit de `tools/advkit.py`) joue le propriétaire amont. Il lit chaque diff en entier, les corps, les réponses, `failing_first.md` et l'inventaire, lance les tests nommés sous le canari, et rend des points classés « blocker », « likely comment », « nit », chacun avec fichier, ligne et preuve. Chaque point se vérifie dans l'arbre, puis se corrige ou se nomme dans le corps. Le texte de l'agent se garde tel quel dans `../thalamus_local/docs/pr/<lot>/adversarial_review.md`. La relecture se rejoue sur les têtes corrigées ; un nouveau « blocker » relance le niveau 2. Invite : section 1 bis. | Outil `Agent`, environ 20 min | Le 21 octobre 2026 au soir, sur les têtes du second amendement : recherches typées et de bibliothèque sans porte (#76), tableau du corps de #72 contraire au code, messages de commit de l'ancienne conception, index promis absent (#75), `supersedes` encore oracle (#72), agrégats liés à une liste d'identifiants (#76), test qui lit l'embedder de la machine (#77). |

### 1 bis. Invite de la relecture adverse (contrôle 5)

Révisée le 21 octobre 2026 au soir, sur demande de Lauro (rationaliser les requêtes sans perdre de qualité). Mesure du tour 2 de la PR à part, avant la révision : 44 tours, contexte de 45 891 à 174 043 tokens, 5,26 millions de tokens relus dans le cache, environ 743 000 tokens d'entrée en équivalent. La relecture du contexte à chaque tour domine le coût ; le contexte de départ (45 900 tokens) venait surtout des définitions de plus de 100 outils CortexDB que l'agent n'utilise pas.

Procédure :
1. Préparer le kit : `python3 -I tools/advkit.py --repo <copie> --head <sha> --base <sha> --out <scratchpad>/adv-<pr>-r<n>`. Il exporte les deux arbres, écrit `pr.diff` et `stat.txt`, et deux scripts : `both.sh` (même `go test` sur `main` et sur la tête, sous le canari, sortie filtrée) et `mutone.py` (une mutation sur une copie de la tête, verdict `killed`, `survived` ou `build-failed`).
2. Lancer l'agent `upstream-reviewer` (`.claude/agents/upstream-reviewer.md`) : outils `Bash`, `Read`, `Grep`, `Glob`, `Write` seulement, donc ni outils CortexDB ni cerveau du projet. Ses consignes fixes : juger sur ce que voit le propriétaire, lire par plages, grouper les lectures.
3. Lui donner l'invite ci-dessous, avec les têtes et les chemins remplis, sans autre consigne ni résumé du développement.
4. Garder son rapport tel quel dans `../thalamus_local/docs/pr/<lot>/adversarial_review*.md`, puis mesurer : `python3 -I tools/agentcost.py <journal de l'agent>`, et noter la ligne dans le rapport.

Qualité : l'agent ne rejoue plus les mutations de l'auteur, déjà prouvées dans `failing_first.md`. Il écrit les siennes, sur les promesses du corps : au tour 2 de la PR à part, ses propres mutations ont trouvé 5 des 7 points (L1, N1 à N4), le rejeu des 26 mutations listées aucun.

```text
You are the upstream maintainer of CortexDB (liliang-cn/cortexdb). You already
asked for changes twice on these PRs. Your habit: in round 1 you name cases; in
round 2 you test the whole class of each case (every tool, every read path, every
write path, every exported search method, every typed RPC, every reserved key)
and you replay scenarios on main and on the branch, then compare.

Read-only review. Do not push, do not comment on GitHub, do not change a file
outside the kit directory. Run Go only through the kit's scripts, which set
this environment, or with it set yourself:
  CORTEXDB_REMOTE=127.0.0.1:1 CORTEXDB_GRPC_TOKEN=canary CORTEXDB_PATH=/dev/null/canary
  CORTEXDB_EMBED_BASE_URL=http://127.0.0.1:1 OPENAI_BASE_URL=http://127.0.0.1:1
  OPENAI_API_KEY= CORTEXDB_LLM_BASE_URL=http://127.0.0.1:1
  CORTEXDB_LLM_API_KEY=canary CORTEXDB_MEMORY_SEMANTIC_FLOOR=canary

Heads (worktrees under <WT>, read-only for you): <branch> <sha>, one line each,
with the base of each.
Kit (one per head): <KIT>/head and <KIT>/main are exported copies; <KIT>/pr.diff
and <KIT>/stat.txt are the diff; <KIT>/both.sh '<go test args>' runs the same
tests on main then on the head, filtered; python3 -I <KIT>/mutone.py <file>
<old> <new> <pkg> <-run pattern> applies one mutation to a copy of the head,
runs the tests and restores it. Put probe tests in a zz_*_test.go file in both
copies, run them with both.sh, and remove them at the end.
Material: the PR bodies (<dir>/GC-*.body.md), the draft replies
(<dir>/review_replies.md, if any), the proof of failing tests
(<dir>/failing_first.md), the class inventory (<dir>/inventory_*.md), the
previous review comments or reports (<files>), the commit messages
(git log -1 on each head).

For each PR:
1. Read the full diff against its base. Read every function the diff calls
   whose behavior the body describes.
2. For each rule the PR adds, list the whole class it applies to, from the
   code, not from the inventory, and find the members it misses.
3. For each claim in the body, the replies, a schema text or a commit message,
   find the code and the test that hold it. A claim without both is a finding.
   Do not replay the author's listed mutations; write at least five of your
   own with mutone.py, aimed at the claims, and report each survivor.
4. Compare with main: same input on both, what differs, and is it named.
   Give the scenarios you ran as a table (scenario, main, head).
5. Look for tests that pass without the code (equivalent mutants, weak asserts),
   tests that read the machine's environment, and tests that sleep.
Rank each finding "blocker", "likely comment" or "nit", with file:line and the
evidence you checked yourself. Report counts you computed, not counts you read.
Write in English, as you would write the review.
```

Ce que ces contrôles ne couvrent pas : un `main` qui bouge après la poussée (v2.135.0 publiée 5 minutes après les revues du 21 octobre 2026), et un désaccord de conception (clé de métadonnées contre champ explicite). Pour le second, un ADR court dans le corps de la PR, avec l'alternative écartée, laisse le relecteur trancher au tour 1.

Entre deux amendements, la suite complète n'apporte rien qu'un changement confiné à un paquet puisse casser : les tests `-race` des paquets touchés (niveau 1) suffisent. Le disque conditionne le niveau 3 : avec moins de 10 Go libres, un `go clean -cache` entre deux suites coûte environ 2 min de recompilation chacune et interdit deux chaînes en parallèle.

## 2. Outils (racine du dépôt, `Makefile`)

| Commande | Ce qu'elle fait | Durée |
| :- | :- | :- |
| `make precommit REPO=<copie> BASE=<base>` | `gofmt -l`, `go vet` et `go test -race` des paquets touchés sous le canari (`CORTEXDB_REMOTE=127.0.0.1:1`, `CORTEXDB_PATH=/dev/null/canary`), puis couverture du patch. Sortie 1 au premier écart. | 1 à 7 min selon les paquets |
| `make patchcov REPO=… BASE=… [DECLARED=…]` | `python3 -I tools/patchcov.py` : chaque bloc changé que les tests n'exécutent pas. Les blocs du corps de `func main()` sont affichés à part et ne bloquent pas. `DECLARED=pkg/core/store_postgres.go` (ou `fichier:L1-L2`) déclare un bloc hors `main` que rien ne peut couvrir en local : affiché à part, sans bloquer, à nommer dans la section `## Codecov` de la PR. Le hook `pre-push` lit la variable dans l'environnement : `DECLARED=… git push …`. | environ 1 min |
| `make mutate REPO=… BASE=…` | `gremlins unleash --diff <base>` : mutation automatique des lignes changées. Rapport à lire, voir section 5. | 5 à 6 min pour un lot comme le lot C |
| `make clock REPO=… BASE=… [COUNT=10]` | `python3 -I tools/clockcheck.py` : dans les fichiers de test du diff, chaque `Test…` qui appelle `time.Now` ou un auxiliaire de test qui l'appelle (`pastInstant`, `futureInstant`), rejoué `go test -race -count=COUNT` sous le canari, un paquet à la fois. Sortie 1 au premier `--- FAIL`. | 2 à 7 min selon le nombre de tests |
| `make prepr REPO=… BASE=… FF=…` | `invcheck`, `precommit`, `mutate`, `clock`, puis la suite complète `go test -race ./...` et l'éval lexicale, comme en CI. | environ 25 min |
| `make benches` | `tests/e2e_shared_brain.sh` (30 contrôles) puis `tests/e2e_gateway.sh` (55 contrôles), sur le clone CortexDB tel qu'il est (`CORTEXDB_CLONE`, sinon `cortexdb/` ; donc `proto` reconstruite) et le gateway. Lignes `FAIL` et `pass=` seules, sortie 1 au premier banc rouge. | environ 3 min |
| `make level3 REPO=… BASE=…` | `prepr` puis `benches`. | environ 30 min |
| `make handmut REPO=… FILE=… OLD='…' NEW='…' GOTEST='…'` | `python3 -I tools/handmut.py` : remplace `OLD` par `NEW` (une occurrence exacte) dans `FILE`, lance `go test GOTEST` sous le canari, restaure `FILE` depuis la copie prise avant, jamais depuis `HEAD`. Refuse un fichier non commis (`--allow-dirty` pour passer outre). Sortie 0 mutant tué, 1 survivant, 2 refus. | le temps des tests |
| `make inventory REPO=… BASE=…` | `python3 -I tools/inventory.py` : la liste complète de chaque classe (outils, lectures, écritures, identifiants, clés réservées, descriptions ajoutées par le diff, entrées exportées de `pkg/cortexdb` qui lisent, RPC typés de `proto/`), en Markdown, une colonne « Test » vide à remplir dans `failing_first.md`. `--class <nom>` restreint, `--pattern <regex>` étend la classe des lectures. | immédiat |
| `make invcheck REPO=… BASE=… FF=<fichier> [INVCLASSES=…]` | `tools/inventory.py --check` : chaque élément des classes (par défaut toutes sauf `promises`) a une ligne de table dans `FF`, nom entre accents graves en première colonne, dernière colonne remplie. Imprime chaque manque, sortie 1 s'il y en a un. Premier pas de `make prepr`, qui demande donc `FF`. | immédiat |
| `make diffmain REPO=… BASE=… SCENARIOS=<dossier>` | `python3 -I tools/diffmain.py` : copie de travail de `BASE`, scénarios posés sur les deux arbres, `go test -v -run Scenario` par paquet sous le canari, scénarios et copie retirés, diff unifié des lignes `SCENARIO` et `--- FAIL`. Sortie 0 identique, 1 différent, 2 un arbre ne compile pas. Jamais pendant une chaîne de niveau 3 sur la même copie. | environ 3 min |
| `make bodynums REPO=… BASE=… SETS=<pr> BODY='<corps> ../thalamus_local/docs/pr/round2/review_replies.md#<pr>' [PR=<n>] [FILL=1]` | `python3 -I tools/bodynums.py` : mesure les lignes hors tests (avec et sans `*.md`, fichiers, retirées), compte les mutations des jeux de `mutate_round2.py` dont le nom commence par `SETS` (lecture de `SETS` sans exécuter le script) et, avec `PR`, lit les lignes non couvertes du rapport Codecov de la tête (refus si le rapport précède la tête). Compare les phrases du corps (« N mutations for this amendment, N killed », « N lines outside tests[ and docs] over N files, N removed », « N of the added lines are in », « The N lines left uncovered ») et signale chaque repère `CODECOV_<pr>`. `FILL=1` remplit le repère en lettres. `<fichier>#<pr>` lit la seule section « ## #<pr> ». Sortie 1 au premier écart. Tests : `python3 -I tools/test_bodynums.py`. | immédiat, quelques secondes avec `PR` |
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

### 3 bis. Forme de codage d'une tranche (décision de Lauro, 22 octobre 2026)

Leçons du lot C, second tour. Sur #76 et #77, 12 lignes déclarées « PostgreSQL » ne demandaient aucun serveur (`pgAggregateWhere`, `pgBindEach`). La plupart des constats des relectures étaient des membres oubliés d'une classe : la voie vectorielle sans le statut du plan, `get_nodes` hors de la porte, les parcours qui servaient les mémoires cachées. Les corps avaient des affirmations trop larges et des nombres faux, écrits à la main.

Règles :

1. **Calcul pur, puis une ligne d'entrée-sortie.** Chaque requête SQL, commande ou décision sort d'une fonction pure qui rend `(sql, args)` ou une valeur, testée sans base ni serveur. Seule la ligne finale fait l'entrée-sortie (`Exec`, `Query`, `os.Exit`).
2. **Une ligne n'est inévitable qu'après un essai de test écrit.** La déclaration nomme le besoin (« needs a server », « in `main` », « reads the clock »), jamais le fichier. Dans le `main` de l'amont, suivre son modèle et déclarer.
3. **Le test de classe d'abord.** Inventaire de la classe depuis l'arbre, puis un test en table sur tous les membres, rouge pour tous, avant le code. Un membre ajouté sans la règle fait tomber le test.
4. **Un seul point de passage, sans défaut silencieux.** Une règle passe par une seule fonction. Une valeur vide ou zéro ne vaut pas « pas de règle » : un site d'appel qui oublie la règle ne compile pas, ou fait tomber un test.
5. **Le corps s'écrit depuis les tests.** Chaque phrase de comportement nomme son test. Une phrase sans nom de test se relit. Les nombres (mutations, lignes hors tests, lignes Codecov) sortent d'un outil, jamais d'une copie à la main. Outil : `make bodynums`, avant chaque `gh pr edit` et avant `gh pr ready`.
6. **Environ 400 lignes hors tests par PR** (demande du propriétaire, 20 octobre 2026). Une règle et sa classe complète tiennent dans une PR. Le coût de la relecture adverse suit la taille.

Ordre d'une tranche :

1. Inventaire de la classe, `invcheck` vide.
2. Test de classe en table, rouge sur tous les membres.
3. Cœur pur et ses tests unitaires.
4. Câblage membre par membre, avec la mutation à la main de chaque règle au moment de la règle.
5. `patchcov` en local. Chaque bloc restant reçoit un essai de test, puis une déclaration avec son besoin.
6. Corps de la PR écrit depuis les noms des tests, nombres contrôlés par `make bodynums`.
7. Relecture adverse, puis niveau 3.

## 4. Procédure par PR

1. Crée la branche du correctif depuis `origin/main`, ou depuis la branche du dessous si la PR est empilée. Un commit par correctif, sans entrée `CHANGELOG` (elle reste sur `proto/granular-cortex`).
2. Écris d'abord le test de la règle et lance-le : il doit tomber. Écris ensuite le code qui le fait passer. Même ordre pour chaque chemin d'erreur (section 3) : d'abord la panne provoquée et l'assertion, puis le code qui la traite.
3. Niveau 1 : lance `make precommit REPO=<copie> BASE=<base>`. Si un bloc hors `main` reste, ajoute le test ou retire le code. Note les blocs de `main` pour le corps.
4. Niveau 1 : pour chaque règle, rejoue une mutation à la main, par `make handmut` (une mutation) ou par le harnais du lot (`../thalamus_local/docs/pr/round2/mutate_round2.py`, un jeu par PR, chaque test attendu vérifié contre le filtre `-run`). Jamais à la main avec `git checkout -- <fichier>` pour restaurer. Consigne « mutation → test qui tombe » dans `failing_first.md`.
5. Niveau 2 : pour chaque règle et chaque point de revue, nomme la classe et lance `make inventory` ; une ligne par élément dans `failing_first.md`. Écris les scénarios différentiels du lot et lance `make diffmain` ; nomme chaque ligne du diff dans le corps ou corrige. Relève chaque promesse nouvelle (classe `promises`) et nomme son test. Rédige le corps (étape 6), relis-le, fais l'analyse qui reste. Puis gèle les têtes et lance le niveau 3 : `make prepr REPO=<copie> BASE=<base>` sur la tête de chaque pile, puis `make benches` une fois sur `proto` reconstruite (`make level3` enchaîne les deux sur une tête). Lis les survivants de `gremlins` (section 5) et rejoue ceux des paquets `main` par `make handmut`. Un trou trouvé pendant la chaîne va dans une liste ; la chaîne repart une fois, après la correction de toute la liste.
6. Rédige le corps dans `../thalamus_local/docs/pr/<lot>/<id>.body.md` : quoi, pourquoi, tests, section « Failing first », et, si des lignes de `main` restent, une section `## Codecov` après `## Breaking Changes`, de même rang, qui dit « The N lines left uncovered are inevitable. … », les nomme et dit pourquoi. Pas de pied de page (décision de Lauro du 19 octobre 2026 : ni `Co-Authored-By` dans les commits, ni ligne « Generated with » dans les corps).
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
- Corriger les cas nommés par le relecteur, pas leur classe (21 octobre 2026 : 11 points sur 14 au second tour étaient dans la classe d'un point du premier tour). Les cinq contrôles du niveau 2 (section 1) font l'énumération avant la poussée.
- Un test d'un paquet `cmd/` qui pose `CORTEXDB_PATH` mais pas `CORTEXDB_REMOTE` parle au serveur de la machine qui le lance (21 octobre 2026, `sweep_expired_test.go`, balayage à vide envoyé au cerveau partagé du relecteur). Le canari du `Makefile` le fait tomber ; le test pose chaque variable qu'il lit.
- Une seule chaîne à la fois dès que `CORTEXDB_TEST_POSTGRES` est posé, quelle que soit la copie : deux `precommit` en parallèle sur deux copies ont saturé la base de test (« FATAL: sorry, too many clients already », `max_connections` = 100) et fait tomber 11 tests PostgreSQL de `pkg/cortexdb` sans défaut du code (21 octobre 2026).
- Une mutation qui ne fait rien tomber n'est pas forcément un trou de test : vérifier d'abord qu'elle change le comportement. La forme `IS NOT NULL AND <= ?` du prédicat d'expiration est encore servie par l'index d'expression de SQLite (`EXPLAIN QUERY PLAN` le montre) ; la mutation fidèle est un prédicat sur une autre expression que celle de l'index (21 octobre 2026, #76).
- `patchcov` compte tout le corps d'une fonction déplacée ou renommée comme nouveau : ses chemins d'erreur préexistants entrent dans le patch. Préférer une délégation (`f` garde son corps et appelle `fIf(…, condition)`) à un déplacement, pour que le diff ne porte que la règle nouvelle (21 octobre 2026, `cleanupKnowledgeArtifactsIf`).
- Un fichier de scénario de `make diffmain` n'utilise que l'API de `main` : un champ nouveau (`Dedupe`, `Unchanged`) le fait échouer à la compilation sur la base, et l'outil s'arrête. Les scénarios d'une fonction nouvelle sont des tests unitaires de la branche, pas des scénarios différentiels (21 octobre 2026).
- Un script de mutation qui lance la suite avec `-run <regex>` doit vérifier que chaque test attendu correspond au filtre : le 21 octobre 2026, une mutation de #77 a été déclarée survivante parce que le test visé (`TestAStatusFilterOnTheAggregatesListsWhatWaits`) ne contenait aucun mot du filtre ; rejouée à la main, elle tombait. Une survivante se rejoue à la main avec `-run '^<Nom>$'` avant d'être classée.
- Une mutation rejouée à la main sur un arbre non commis, suivie de `git checkout -- <fichier>`, retire le correctif avec la mutation (21 octobre 2026, #72 : l'amendement n'a emporté que le test ; détecté par le test rouge sur `proto`). Commettre ou `git stash` avant de muter à la main, ou restaurer depuis une copie du fichier, jamais depuis `HEAD`.
- Les bancs de bout en bout font partie du niveau 3 pour de bon : le 21 octobre 2026 ils ont trouvé un circuit entier refusé (`th_pending` sous la règle du préfixe `_`) et un banc resté sur l'ancienne API, que ni `precommit` ni les mutations ne pouvaient voir.
- Un test qui compare à un instant relatif à « maintenant » (`pastInstant()`, `time.Now().Add(-h).Format(RFC3339)`) appelé deux fois tombe quand une seconde passe entre les deux appels : 1 échec sur 12 sous `-race` le 21 octobre 2026. Lire l'instant une fois dans une variable et comparer à elle.
- Un conflit de rebasage résolu « ours » sur un fichier de test garde le libellé attendu de la branche du dessus, pas celui du moteur rebasé : relire chaque hunk de test résolu contre le message du code (21 octobre 2026, « scan document gate » contre « scan document expiry »).
- Un rebasage « ours » peut retirer des lignes que la branche du dessous venait d'ajouter dans le même hunk (21 octobre 2026 : `if metadata == nil { metadata = map… }` de #76 perdu au rebasage de #77, `nil map` en test) : après un rebasage avec conflits, `git diff <base> -- <fichier>` doit se lire hunk par hunk.
