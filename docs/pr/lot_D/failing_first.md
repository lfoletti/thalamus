# Preuve « failing first », lot D (GC-10, 20 octobre 2026)

Branche `gc-10-cortex-query-prefetch`, base `origin/main` à v2.134.1 (`60a8831`).

## Le test tombe sur `main`

`pkg/cortexdb/query_prefetch_scope_test.go`, `TestTheGraphLaneOfAQueryStaysInTheCollection`, posé sur l'arbre sans le correctif :

```
--- FAIL: TestTheGraphLaneOfAQueryStaysInTheCollection (0.39s)
    query_prefetch_scope_test.go:57: with collection notes the lane served [n s] (<nil>)
```

C'est la reproduction du point 3 de la revue de #72 : une requête sur `notes` reçoit un chunk de `secrets` par la voie graphe.

## Mutations à la main (`mutate_10.py`, `go test ./pkg/cortexdb/ -run Query`)

Quatre mutations, quatre tuées :

| Règle désactivée | Mutation | Test qui échoue |
| :- | :- | :- |
| Collection ignorée | `if collection != "" && false {` | TestTheGraphLaneOfAQueryStaysInTheCollection |
| Chunks hors collection gardés | `if _, ok := inScope[chunk.ID]; !ok {` | TestTheGraphLaneOfAQueryStaysInTheCollection |
| Collection illisible ignorée | `if err != nil && false {` après `chunksInScope` | TestTheGraphLaneOfAQueryStaysInTheCollection |
| Collection non transmise à la voie | `db.runGraphQueryPrefetch(ctx, "", entityNames, …)` | TestTheGraphLaneOfAQueryStaysInTheCollection |

Arbre restauré, tests `Query` verts.

## Couverture et passage automatique (`make prepr REPO=<worktree> BASE=origin/main`, 20 octobre 2026, 18 h)

- `precommit` : gofmt, vet, `-race` de `pkg/cortexdb` verts ; `patchcov` : 0 bloc du patch sans test.
- `gremlins unleash --diff origin/main` : 2 mutants, 2 tués, 0 survivant, 0 non couvert.
- Suite complète `go test -race -count=1 -timeout 25m ./...` verte, éval lexicale `TestLexicalRetrievalQuality` verte.
