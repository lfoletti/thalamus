# Failing first : gc-18-allowlist-cortex-query

Tête `4746ff2`, base `origin/main` (`a591ae1`). Règle : `cortex_query` est sur la liste blanche, `collection` en portée forcée, chaque autre argument du schéma classé en données, `prefetch[].max_hops` refusé avec sa raison, `prefetch[].source` refusé d'office (absent du schéma).

## Tests sur la base

Commande : `go test -count=1 ./pkg/authz/ ./pkg/rpcserver/ ./pkg/httpapi/ -run 'CortexQuery|QueriesItsCollection|MaxHops|QuerySource'`

```
--- FAIL: TestAConfinedKeyQueriesItsCollectionOnly (0.00s)
    allowlist_cortex_query_test.go:26: cortex_query with every argument on the key's collection was refused: denied: key "notes" is confined and cortex_query is not a tool a confined key may call: it 
--- FAIL: TestAConfinedKeyMayNotSendMaxHopsToTheGraphLane (0.00s)
    allowlist_cortex_query_test.go:56: cortex_query {"collection":"notes","prefetch":[{"kind":"graph","entity_names":["Orpheus"],"max_hops":2}]}: denied: key "notes" is confined and cortex_query is no
    allowlist_cortex_query_test.go:56: cortex_query {"collection":"notes","prefetch":[{"kind":"graph","entity_names":["Orpheus"]},{"kind":"graph","entity_names":["Orpheus"],"MAX_HOPS":1}]}: denied: ke
    allowlist_cortex_query_test.go:60: the graph lane at its default depth was refused: denied: key "notes" is confined and cortex_query is not a tool a confined key may call: it its graph prefetch se
--- FAIL: TestAConfinedKeyMayNotNameAQuerySource (0.00s)
    allowlist_cortex_query_test.go:71: a query source: denied: key "notes" is confined and cortex_query is not a tool a confined key may call: it its graph prefetch searches chunks of every collection
FAIL
FAIL	github.com/liliang-cn/cortexdb/v2/pkg/authz	0.411s
--- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverGRPC (0.03s)
    confined_cortex_query_test.go:51: notes key, "query":"Orpheus ledger","entity_names":["Orpheus"]: rpc error: code = PermissionDenied desc = denied: key "notes" is confined and cortex_query is not 
FAIL
FAIL	github.com/liliang-cn/cortexdb/v2/pkg/rpcserver	0.758s
--- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverHTTP (0.03s)
    confined_cortex_query_test.go:41: notes-secret, {"collection":"notes","limit":10,"query":"Orpheus ledger","entity_names":["Orpheus"]}: 403 {"error":"denied: key \"notes\" is confined and cortex_qu
    confined_cortex_query_test.go:41: notes-secret, {"collection":"notes","limit":10,"prefetch":[{"kind":"lexical","query":"Orpheus ledger"}]}: 403 {"error":"denied: key \"notes\" is confined and cort
    confined_cortex_query_test.go:41: notes-secret, {"collection":"notes","limit":10,"prefetch":[{"kind":"graph","entity_names":["Orpheus"]}]}: 403 {"error":"denied: key \"notes\" is confined and cort
    confined_cortex_query_test.go:52: notes key, {"query":"Orpheus","collection":"secrets"}: "denied: key \"notes\" is confined and cortex_query is not a tool a confined key may call: it its graph pre
    confined_cortex_query_test.go:52: notes key, {"query":"Orpheus"}: "denied: key \"notes\" is confined and cortex_query is not a tool a confined key may call: it its graph prefetch searches chunks o
    confined_cortex_query_test.go:52: notes key, {"collection":"notes","prefetch":[{"kind":"graph","entity_names":["Orpheus"],"max_hops":2}]}: "denied: key \"notes\" is confined and cortex_query is no
    confined_cortex_query_test.go:52: notes key, {"collection":"notes","prefetch":[{"kind":"source","source":"external","query":"Orpheus"}]}: "denied: key \"notes\" is confined and cortex_query is not
FAIL
FAIL	github.com/liliang-cn/cortexdb/v2/pkg/httpapi	0.430s
FAIL
```

Les deux tests de relecture amendés et le test de complétude tombent aussi sur la base (`allowlist.go` de `origin/main` posé sur la tête) :

```
--- FAIL: TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC (0.03s)
    confined_tools_review_test.go:77: notes queries: rpc error: code = PermissionDenied desc = denied: key "notes" is confined and cortex_query is not a tool a confined key may call: it its graph pref
--- FAIL: TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST (0.04s)
    confined_tools_review_test.go:64: cortex_query as "notes-secret" = 403, want 200: {"error":"denied: key \"notes\" is confined and cortex_query is not a tool a confined key may call: it its graph p
--- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
    allowlist_test.go:68: cortex_query has a reason of its own the test does not name: "its graph prefetch searches chunks of every collection"
```

## Mutations à la main

Harnais `docs/pr/lot_E/mutate_gc18.py` (`python3 -I docs/pr/lot_E/mutate_gc18.py [G1 …]`) (copie du fichier, remplacements, `go test` sous le canari, restauration depuis la copie, contrôle de l'identité). Une mutation à deux remplacements (G3) ne passe pas par `handmut.py`, qui n'en fait qu'un. Commande des tests : `go test -count=1 -run 'CortexQuery|QueriesItsCollection|MaxHops|QuerySource|Classified|OffTheAllowlist|Review' ./pkg/authz/ ./pkg/rpcserver/ ./pkg/httpapi/`.

| Mutation | Changement | Tests qui tombent |
| :- | :- | :- |
| G1 outil hors liste | `"cortex_query": {` → `"cortex_query_off": {` | `TestAConfinedKeyQueriesItsCollectionOnly`, `…OverGRPC`, `…OverHTTP`, `TestAConfinedKeyMayNotSendMaxHopsToTheGraphLane`, `TestAConfinedKeyMayNotNameAQuerySource`, les deux tests de relecture |
| G2 pas de portée | `scope: collectionScope` → `scope: nil` | les mêmes, et `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys` |
| G3 `max_hops` en données | `max_hops` ajouté aux données de `prefetch[]`, refus retiré | `TestAConfinedKeyMayNotSendMaxHopsToTheGraphLane`, `…OverGRPC`, `…OverHTTP`, complétude, les deux tests de relecture |
| G4 argument du filtre non classé | `has_number` retiré de `queryConditionArgs` | `TestAConfinedKeyQueriesItsCollectionOnly`, complétude |
| G5 encore refusé | `cortex_query` remis dans `refusedTools` | complétude |
| G6 source ouverte | `source` ajouté aux données de `prefetch[]` | `TestAConfinedKeyMayNotNameAQuerySource`, `…OverGRPC`, `…OverHTTP`, complétude |

Six mutations, six tuées. Journal :

```
=== G1-offlist KILLED
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnly (0.00s)
   --- FAIL: TestAConfinedKeyMayNotSendMaxHopsToTheGraphLane (0.00s)
   --- FAIL: TestAConfinedKeyMayNotNameAQuerySource (0.00s)
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverGRPC (0.03s)
   --- FAIL: TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC (0.02s)
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverHTTP (0.03s)
   --- FAIL: TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST (0.02s)
=== G2-noscope KILLED
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnly (0.00s)
   --- FAIL: TestAConfinedKeyMayNotSendMaxHopsToTheGraphLane (0.00s)
   --- FAIL: TestAConfinedKeyMayNotNameAQuerySource (0.00s)
   --- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverGRPC (0.03s)
   --- FAIL: TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC (0.02s)
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverHTTP (0.03s)
   --- FAIL: TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST (0.02s)
=== G3-maxhops-open KILLED
   --- FAIL: TestAConfinedKeyMayNotSendMaxHopsToTheGraphLane (0.00s)
   --- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverGRPC (0.04s)
   --- FAIL: TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC (0.03s)
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverHTTP (0.03s)
   --- FAIL: TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST (0.03s)
=== G4-condition-arg KILLED
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnly (0.00s)
   --- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
=== G5-still-refused KILLED
   --- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
=== G6-source-open KILLED
   --- FAIL: TestAConfinedKeyMayNotNameAQuerySource (0.00s)
   --- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverGRPC (0.03s)
   --- FAIL: TestAConfinedKeyQueriesItsCollectionOnlyOverHTTP (0.03s)
```

## Classe

Classe « outils » de la liste blanche : `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys` tient la ligne de `cortex_query` au schéma, chemin par chemin (`schemaPaths`), dans les deux sens. Classe « entrées qui lancent `db.Query` » : l'outil `cortex_query` (MCP, `CallTool`, `POST /v1/tools/cortex_query`) et `POST /v1/query` (route typée, déjà ouverte à une clé confinée qui nomme sa collection). Pas de RPC typé `Query` en gRPC.

## precommit

`make precommit REPO=<copie> BASE=origin/main` : OK (gofmt, vet, `-race` de `pkg/authz`, `pkg/httpapi`, `pkg/rpcserver`, `patchcov` 0 bloc sans test).
