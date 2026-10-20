# Lot B : preuve « failing first »

Le propriétaire demande (réponse sous #66, 18 octobre 2026) trois PR, « each with its failing-first tests » : chaque test de la PR échoue sans le code de la PR. Vérification faite le 19 octobre 2026, sur les branches à v2.134.0. Les tests sont **dans** les PR ; cette page montre qu'ils échouent avant.

## GC-04a : règles `authz` seules (`9827f2a`)

Les règles sont des fonctions nouvelles : sans elles, les tests ne compilent pas. La preuve utile est par règle : on désactive la règle, on garde le code compilable, et un test doit tomber. Onze mutations, onze tuées (`go test ./pkg/authz/` après chaque mutation, fichier restauré ensuite) :

| Règle désactivée | Mutation (`pkg/authz`) | Test qui échoue |
| :- | :- | :- |
| R1 : champ non déclaré par l'outil | `if !declaredFields[c.field] && false` | `TestAConfinedKeyIsRefusedAToolCallOutsideItsScope` |
| R1 : outil inconnu | `if !known && false` | `TestAConfinedKeyIsRefusedAToolCallOutsideItsScope` |
| R2 : arguments illisibles (`lookup == nil`) | `if false` | `TestAConfinedKeyIsRefusedAToolCallOutsideItsScope` |
| R2 : argument absent | `if false` à la place de `if top == ""` | `TestAConfinedKeyIsRefusedAToolCallOutsideItsScope` |
| R2 : mauvaise valeur | `if false && !c.allows(top)` | `TestAConfinedKeyIsRefusedAToolCallOutsideItsScope` |
| R2 : copie imbriquée | `range nested[:0]` | `TestAConfinedKeyIsRefusedAToolCallOutsideItsScope`, `TestEverySpellingOfAScopeFieldIsHeldToTheConfinement` |
| R2 : autre graphie au premier niveau | `&& false` sur le `EqualFold` de `JSONFieldLookup` | `TestEverySpellingOfAScopeFieldIsHeldToTheConfinement` |
| R2 : autre graphie dans une copie imbriquée | `k == field` à la place de `EqualFold` dans `jsonValuesIn` | **aucun test avant le 19 octobre** ; cas ajouté à `TestEverySpellingOfAScopeFieldIsHeldToTheConfinement` (`plan.filters.USER_ID`), qui tue maintenant la mutation |
| R3 : garde des lignes entière | `return nil` en tête de `GuardToolRows` | `TestAToolCallMayOnlyNameRowsWithinTheScope` |
| R3 : ligne de connaissance | `return true` en tête de `PermitsKnowledge` | `TestAToolCallMayOnlyNameRowsWithinTheScope` |
| R3 : ligne de mémoire | `return true` en tête de `PermitsMemory` | `TestAToolCallMayOnlyNameRowsWithinTheScope` |

Complément du 19 octobre 2026, après le rapport Codecov sur #72 (17 lignes non couvertes) : deux branches fonctionnelles sans test, la copie d'un champ de portée **dans un tableau** (`plans: [{filters: {...}}]`) et la borne de profondeur de la marche JSON ; cinq chemins d'erreur (arguments non objet, outil inconnu, lecture en échec, dans la garde des lignes). Tests ajoutés : `TestNestedCopiesInsideArraysAreHeldToTheConfinement`, `TestTheRowGuardFailsClosedOnWhatItCannotDecide`. La garde refuse désormais aussi un outil inconnu (avant : passait, parce qu'`AuthorizeToolRows` le refuse en amont). Plus aucune ligne non couverte dans `rows.go` et `tools.go`. 04a amendée en `abe694f`, 04b `77cac1d`, 04c `ae3775a`.

Codecov sur #73 et #74 (mêmes 17 lignes `authz`, déjà couvertes, plus `ownership.go` 6 lignes et `policy.go` 8 lignes) : tests ajoutés `TestTheCallToolChecksFailClosedOnTheirOwn` (rpcserver : message d'un autre type, outil inconnu, arguments non objet, lecture en échec par contexte annulé) et `TestTheToolRouteFailsClosedForAConfinedKey` (httpapi : corps trop gros → 413, base fermée → 500). Patch d'`ownership.go` entièrement couvert ; dans `policy.go` reste le bras 403 de la garde, défensif et inatteignable. 04b amendée en `dd69ed9`, 04c en `6b1430a`.

Reproduire : script `mutate.py` (archivé dans ce dossier), ou à la main : appliquer une ligne du tableau, `go test ./pkg/authz/`, `git checkout -- pkg/authz`.

## GC-04b : câblage gRPC (`19c9028`)

Les deux fichiers de test de la PR, posés sur le code de GC-04a seule (`git checkout 19c9028 -- pkg/rpcserver/scoped_keys_test.go pkg/rpcserver/tools_auth_test.go` depuis `9827f2a`), puis `go test ./pkg/rpcserver/` :

```
--- FAIL: TestAConfinedKeyMayCallToolsWithinItsScope (0.01s)
--- FAIL: TestAKeyConfinedToACollectionMayIngestAndSearchItThroughTheTools (0.01s)
--- FAIL: TestAKeyConfinedToSeveralCollectionsReachesEachOfThem (0.01s)
--- FAIL: TestAConfinedKeyCannotReachAnotherRowByIDThroughTheTools (0.02s)
--- FAIL: TestAConfinedKeyIsRefusedToolCallsItsScopeCannotCover (0.01s)
FAIL	github.com/liliang-cn/cortexdb/v2/pkg/rpcserver	1.425s
```

Avec le code de la PR : `ok`.

## GC-04c : câblage REST (`bc214f6`)

Le fichier de test de la PR, posé sur le code de GC-04a + GC-04b (`policy.go` inchangé), puis `go test ./pkg/httpapi/` :

```
--- FAIL: TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP (0.01s)
--- FAIL: TestAConfinedKeyCannotReachAnotherRowByIDOverHTTP (0.01s)
FAIL	github.com/liliang-cn/cortexdb/v2/pkg/httpapi	6.575s
```

Avec le code de la PR : `ok`.

## Ce qui va dans chaque PR

Les tests eux-mêmes, et dans le corps de la PR une section **Failing first** qui donne la commande et les lignes `FAIL` ci-dessus (04b, 04c) ou le tableau règle → test (04a). Le relecteur peut rejouer la commande. Rien d'autre n'est nécessaire : le dépôt fusionne par écrasement, un commit « test rouge » séparé disparaîtrait à la fusion.

## Amendement du 20 octobre 2026 : liste blanche par outil (#72)

Après la revue du propriétaire (cinq failles, voir `README.md`), la règle dérivée du schéma est remplacée par une liste blanche explicite par outil (`pkg/authz/allowlist.go`). Chaque règle a sa mutation à la main, rejouée par `mutate_amend.py` (`python3 -I docs/pr/lot_B/mutate_amend.py <worktree>`), `go test ./pkg/authz/` après chaque mutation, fichier restauré ensuite. Quinze mutations, quinze tuées :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| R1 outil hors liste admis | `allowlist.go` : `		return rule, nil` | TestAConfinedKeyIsRefusedAToolOffTheAllowlist |
| R1 argument inconnu admis | `allowlist.go` : `if !allowed[child] && false {` | TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover |
| R1 argument refusé admis | `allowlist.go` : `if why, no := denied[child]; no && false {` | TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover |
| R2 champ non honoré | `allowlist.go` : `if !contains(rule.scope, c.field) && false {` | TestTheScopeRuleStillHoldsOnTheAllowlist |
| R2 argument absent | `allowlist.go` : `if top == "" && false {` | TestTheScopeRuleStillHoldsOnTheAllowlist |
| R2 mauvaise valeur | `allowlist.go` : `if !c.allows(top) && false {` | TestTheScopeRuleStillHoldsOnTheAllowlist |
| R2 copie imbriquée | `allowlist.go` : `for _, v := range nested[:0] {` | TestTheScopeRuleStillHoldsOnTheAllowlist |
| R3 porte non exigée | `allowlist.go` : `if on, _ := lookupKey(obj, gate).(bool); gate == "" || !on &` | TestAToolThatReadsBothStoresIsGatedForAConfinedKey |
| R3 porte non forcée | `allowlist.go` : `	// Marshalling` | TestAToolThatReadsBothStoresIsGatedForAConfinedKey |
| R4 garde des lignes entière | `rows.go` : `	rule, ok := confinedTools[tool]` | TestAToolCallMayOnlyNameRowsWithinTheScope, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths, TestTheRowGuardFailsClosedOnWhatItCannotDecide |
| R4 supersedes oublié | `allowlist.go` : `ids: []idField{{"memory_id", idMemory}}` | TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover, TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| R4 chemin imbriqué oublié | `allowlist.go` : `planIDs    = []idField{}` | TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover, TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys, TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| R4 autre graphie d'un chemin | `allowlist.go` : `if k != name {` | TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths |
| Complétude : argument de schéma oublié | `allowlist.go` : `args: paths(queryArgs, walkArgs, planArgs, pprArgs, recallAr` | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys |
| Complétude : outil ni admis ni refusé | `allowlist.go` : `		}` | TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys |

Arbre restauré, suite verte : True

Couverture du patch (`make patchcov`, base `origin/main` à v2.134.1) : zéro bloc non couvert dans `pkg/authz`.

### Câblage gRPC (#73)

Les tests ne compilent pas sans le câblage (symboles nouveaux). Mutations à la main, `python3 -I docs/pr/lot_B/mutate_wiring.py grpc <worktree>`, `go test ./pkg/rpcserver/` après chacune :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Porte non forcée avant l'appel | `pkg/rpcserver/ownership.go` : `args := []byte(call.GetArgsJson())` | TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC |
| Liste blanche non consultée | `pkg/rpcserver/ownership.go` : `if err := key.AuthorizeToolArgs(call.GetName(), args); err !` | TestAConfinedKeyIsRefusedToolCallsItsScopeCannotCover, TestAConfinedKeyMayCallToolsWithinItsScope, TestAKeyConfinedToACollectionMayIngestAndSearchItThroughTheTools, TestAKeyConfinedToSeveralCollectionsReachesEachOfThem, TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC |
| Arguments forcés non transmis à l'outil | `pkg/rpcserver/ownership.go` : `call.ArgsJson = call.GetArgsJson()` | TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC |

Arbre restauré, suite verte : True

Couverture du patch (base `gc-04a-authz-tool-rules`) : zéro bloc non couvert dans `pkg/rpcserver`.

### Câblage REST (#74)

Même méthode, `python3 -I docs/pr/lot_B/mutate_wiring.py rest <worktree>`, `go test ./pkg/httpapi/` après chacune :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Porte non forcée avant l'appel | `pkg/httpapi/policy.go` : `body = append([]byte(nil), body...)` | TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST |
| Liste blanche non consultée | `pkg/httpapi/policy.go` : `if err := key.AuthorizeToolArgs(tool, body); err != nil && f` | TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP, TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST |
| Sans base, la garde est sautée | `pkg/httpapi/policy.go` : `					writeError(w, http.StatusInternalServerError,` | TestTheToolRouteRefusesAConfinedKeyWhenItHasNoDatabaseToGuardWith |
| Garde des lignes sautée | `pkg/httpapi/policy.go` : `if err := key.Scope.GuardToolRows(r.Context(), db, tool, bod` | TestAConfinedKeyCannotReachAnotherRowByIDOverHTTP, TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST |
| Corps forcé non transmis à l'outil | `pkg/httpapi/policy.go` : `_ = io.NopCloser` | TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST |
| Corps non objet admis | `pkg/httpapi/policy.go` : `if len(bytes.TrimSpace(body)) > 0 && !(isJSONObject(body) &&` | TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP |

Arbre restauré, suite verte : True

Couverture du patch (base `gc-04a-authz-tool-rules`) : zéro bloc non couvert dans `pkg/httpapi`.

### Passage automatique sur #72 (`gremlins unleash --diff origin/main`, 20 octobre 2026)

Première passe : 37 tués, 5 survivants, tous dans le code de présentation de la table (`store.String`, `idKind.String`, `ConfinedToolsTable` : portes et « by id only »). Trois lignes attendues ajoutées au test `TestTheAllowlistTableNamesEveryToolAndEveryIDField` (commit amendé en `3c93d00`). Seconde passe : 42 tués, 0 survivant, 0 non couvert.

### Passage automatique sur #73 (`gremlins unleash --diff gc-04a-authz-tool-rules`, 20 octobre 2026)

Mutants propres au câblage gRPC : 3 tués (`tools.go:42`, `ownership.go:94`, `ownership.go:118`), 0 survivant, 1 « non couvert » selon l'outil : `ownership.go:194:11` (`case err == nil` de `guardToolRows`). Rejoué à la main (`err != nil`) : `TestAConfinedKeyCannotReachAnotherRowByIDThroughTheTools` et `TestTheCallToolChecksFailClosedOnTheirOwn` tombent. La ligne est couverte (`tools_guard_test.go`) ; le « non couvert » est un artefact du profil de l'outil. Les cinq survivants de `allowlist.go` listés dans ce passage sont ceux de #72 avant son amendement `3c93d00` (la branche 04b n'était pas encore rebasée dessus).

### Passage automatique sur #74 (`gremlins unleash --diff gc-04a-authz-tool-rules`, 20 octobre 2026)

Mutants propres au câblage REST : 1 survivant, `policy.go:136:35` (`len(bytes.TrimSpace(body)) > 0` → `>= 0` : un corps vide répondu 400 au lieu de passer par la politique). Cas ajouté à `TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP` (corps vide → 403 « collection »), mutation rejouée à la main et par `mutate_wiring.py rest` : tuée. Commit amendé en `4b3ff4c`, rebasé sur `3c93d00` ; `gc-04b` rebasé de même en `49949eb` (changement de test seul sur `pkg/authz`, suites `authz`, `rpcserver`, `httpapi` vertes).
