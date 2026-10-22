# Relecture adverse de #72, tour 5 (21 octobre 2026, tête `d007fee`)

Rapport de l'agent neuf (invite fixe, `docs/pipeline_pr.md`, section 1 bis). Verdict : aucun bloquant, 5 remarques probables, 3 nits. Suite donnée : voir la fin.

## Rapport (anglais, recopié)

**Verdict:** no blocker. 5 likely comments, 3 nits. All r4 blockers and comments are closed and each one has a test.

Verified by the agent: `go test -race ./pkg/authz/` green; `go test ./...` 30 packages ok (without `-race`); vet and gofmt clean; 895 lines added outside tests, 12 removed; 99 tools, 20 allowlisted, 79 refused; the generated table equals the body's; 11 row-id fields; `mutate_round2.py 72` 29/29 killed; 43 confined calls on a real brain (keys `notes` and `hermes`, modes graph, auto, lexical, ppr, hybrid, plus `extract_conversation`, `aggregate_metadata`, `representative_records`), nothing leaked; main against head through the facade under `CallerScope`: `GetMemory` by list and `supersedes` differ, both named.

- **L1.** `Scope.PermitsMemory` and `Scope.PermitsKnowledge` (`rows.go:56-71`): four clauses replaced by `true` one at a time (`Permits(FieldScope, memoryScope)`, `Permits(FieldNamespace, namespace)`, `Permits(FieldScope, "")` and `Permits(FieldNamespace, "")` in `PermitsKnowledge`), all four survive `pkg/authz`. No test uses a key confined on scope or namespace against a row.
- **L2.** `init()` (`allowlist.go:256`) refuses every unlisted tool with `graphOnlyReason`, so the branch "neither allowed nor refused" (`allowlist_test.go:60`) is dead and the body's "a new tool fails that test" is wrong (a new tool is refused silently). 66 of 79 refused tools carry the blanket reason, inaccurate for `ontology_*`. Deleting five explicit reasons leaves the suite green.
- **L3.** "The extraction … keeps the rest" has no test; the probe (`knowledge_save`, `knowledge_update`, `knowledge_delete` by the `notes` key) rewrote the `secrets` entity's vector, added `properties.type`, moved `valid_from`/`recorded_at` (a new bitemporal version), and left `"notes"` in `collections` after the delete. Description and edges unchanged.
- **L4.** The "document without chunks" rule is in `guardRow` (`rows.go:135`) only; the typed path (`rpcserver/ownership.go:150-155`, called at `:212`) still places such a document in `graphrag_chunks`. The commit message says the row rules "move into authz … so the typed RPCs and the tools share them"; on this head they are copied (`rows.go:56-71`, `ownership.go:130/150`, `caller.go:30-41`).
- **L5.** `CallerScope` in the facade covers the supersede target but not the row: under `CallerScope{hermes}`, `SaveMemory{memory_id:"zeus-1", user_id:"hermes"}` takes the row, `UpdateMemory(zeus-2)` overwrites it, `DeleteMemory(zeus-3)` deletes it. Unchanged from main; the body must say what `CallerScope` covers.
- **N1.** `plan.filters.document_ids` is on 6 allowlisted tools; the pin test covers 3; `failing_first.md:77` says four.
- **N2.** "Arrays do not spend depth" (`allowlist.go:566`) has no test: "arrays not scanned" and "depth-1" survive.
- **N3.** `isNotFound` also accepts `sql.ErrNoRows`; the commit message says "only the facade's not-found sentinel".
- Not findings: the mutant "gate check only when the tool has a scope" is equivalent; no test sleeps or reads the environment; two tests pass on main (stated).

## Suite donnée (21 octobre 2026, 21 h 15)

- L1 : `TestTheRowRulesHoldEveryConfinedField` (rows_test.go), quatre mutations au jeu 72, tuées.
- L2 : `graphOnlyReason` devient `defaultRefusal`, « reaches state that no scope field confines », vrai pour les 66 outils ; le test de complétude nomme les 13 outils à raison propre et perd sa branche morte ; deux mutations (raison effacée, outil nouveau admis), tuées ; corps corrigé (un outil nouveau est refusé par défaut, la liste se ferme d'elle-même).
- L3 : corps et message de commit nomment ce que l'extraction touche encore (collection ajoutée et gardée après effacement, type, embedding, nouvelle version), comportement de `main` pour toute clé, hors de cette PR.
- L4 : message de commit corrigé (les règles sont exportées pour les RPC typés). La règle des chunks sur le chemin typé va à #73.
- L5 : corps, point 5 : ce que `CallerScope` couvre et ne couvre pas ; la garde des lignes tient le reste.
- N1 : `failing_first.md` dit six outils, trois joués.
- N2 : `TestACopyOfTheScopeFieldInsideAListIsRead`, deux mutations, tuées.
- N3 : message de commit : `core.ErrNotFound` et `sql.ErrNoRows`.
