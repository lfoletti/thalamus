# Failing first : gc-17-query-graph-lane-scope

Tête `1b655c7`, base `origin/main` (`a591ae1`). Règle : avec une collection, la voie graphe de `db.Query` marche seulement ce que la collection peut voir (`neighborsInCollection`, `laneScope.mayEnter`) : graine connue de la collection ; chunks et documents de la collection ; entités connues (`entityKnownIn`) ; jamais une mémoire ; une arête qui nomme une provenance doit en nommer une de la collection, une arête sans provenance est suivie. Sans collection, `graph.Neighbors` comme avant.

## Tests sur la base

Fichiers de test posés sur `origin/main` (`query_graph_scope_walk_test.go` retiré : il appelle les fonctions nouvelles et ne compile pas sur la base).

Commande : `go test ./pkg/cortexdb/ ./pkg/httpapi/ -count=1 -run 'TestTheGraphLane|TestAConfinedQueryDoesNotWalk'`

```
--- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.06s)
    query_graph_scope_test.go:133: alpha, seed "Project Nightjar", 2 hops: "a1,a2", want ""
    query_graph_scope_test.go:133: alpha, seed "Kestrel Vale", 3 hops: "a1,a10,a11,a2,a3,a4,a5,a6,a9", want "a1,a10,a11,a4,a9"
FAIL
FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.495s
--- FAIL: TestAConfinedQueryDoesNotWalkAnotherCollectionsRelationOverHTTP (0.03s)
    query_graph_scope_test.go:40: alpha key, seed only beta knows, 2 hops: {"fusion":"rrf","results":[{"id":"chunk:a1:000","content":"kestrel vale runs the alpha mill.","doc_id":"a1","metadata":{"_chunk_overlap":"0","_ch
FAIL
FAIL	github.com/liliang-cn/cortexdb/v2/pkg/httpapi	0.683s
FAIL
```

`TestTheGraphLaneWithoutACollectionKeepsItsHops` passe sur la base : c'est le contrôle positif (chaque chemin refusé existe sans collection).

## Mutations à la main

Harnais `docs/pr/lot_E/mutate_gc17.py` (appelle `tools/handmut.py` pour chaque mutation, restauration depuis une copie). Commande des tests : `go test -count=1 -run 'TestTheGraphLane|TestTheScopedWalk|TestDocumentsInCollection|TestAScopedWalk|TestALaneStep' ./pkg/cortexdb/`. Rejouer : `python3 -I docs/pr/lot_E/mutate_gc17.py [M1-seed …]`.

| Mutation | Règle désactivée | Premier test qui tombe |
| :- | :- | :- |
| M1-seed | graine non contrôlée | `TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation` |
| M2-entity | `entityKnownIn` retiré | idem |
| M3-memory | mémoire entrée | idem |
| M4-chunkdoc | chunk ou document d'une autre collection entré | idem |
| M5-provrefused | arête d'une autre collection suivie | idem |
| M6-provdoc | provenance par `document_id` ignorée | idem |
| M7-provchunk | provenance par chunk ignorée | idem |
| M8-provchunkread | chunks de provenance non lus | idem |
| M9-noprov | arête sans provenance refusée | idem |
| M10-limit | limite retirée | `TestTheScopedWalkKeepsNodeTypesAndLimit` |
| M11-types | types de nœuds non appliqués | idem |
| M12-notypes | aucun type nommé = rien | idem |
| M13-depth | un saut de trop | `TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation` |
| M14-unscoped | marche confinée sans collection | `TestTheGraphLaneWithoutACollectionKeepsItsHops` |
| M15-direction | sens de l'arête perdu | `TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation` |
| M16-sqlcoll | test de collection du SQL inversé | idem (et `TestDocumentsInCollectionQuery`) |
| M17-seen | nœud visité deux fois | `TestTheScopedWalkKeepsNodeTypesAndLimit` |
| M18-seedskip | graine rendue | idem |
| M19-docsprov | documents de provenance non lus | `TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation` |
| M20-sourcechunk | `source_chunk_id` ignoré | `TestALaneStepIsJudgedByItsNodeAndItsProvenance` |
| M21-unread-node | nœud non lu accepté | idem |

Vingt et une mutations, vingt et une tuées. Écarts en route : M7, M8, M19 et M11 ne compilaient pas dans leur première forme (variable ou import inutilisé), réécrites pour compiler ; M8 et M19 ont d'abord survécu parce que la provenance nommait un chunk ou un document que le même tour lisait déjà : le banc nomme désormais le chunk `a5` et le document `a6`, que la marche ne rencontre pas. Une garde (`slices.DeleteFunc` des identifiants vides) a survécu, mutant équivalent : retirée.

Journal :

```
=== M1-seed exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.805s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'seen || !scope.mayEnter(step) {' -> 'seen || (depth > 0 && !scope.mayEnter(step)) {')
=== M2-entity exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.05s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.750s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if !entityKnownIn(node.Collections, s.collection) {' -> 'if false {')
=== M3-memory exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.06s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.773s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'case "memory":' -> 'case "memory":')
=== M4-chunkdoc exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.05s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.756s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if _, in := s.docs[node.DocumentID]; !in {' -> 'if false {')
=== M5-provrefused exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.04s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   --- FAIL: TestALaneStepIsJudgedByItsNodeAndItsProvenance (0.00s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.712s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'return true' -> 'return true')
=== M6-provdoc exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.08s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.816s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if _, in := s.docs[doc]; in {' -> 'if false {')
=== M7-provchunk exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.05s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   --- FAIL: TestALaneStepIsJudgedByItsNodeAndItsProvenance (0.00s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.740s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if _, in := s.docs[s.nodes[chunk].DocumentID]; in {' -> 'if _, in := s.docs[s.nodes[chunk].DocumentID]; in && false {')
=== M8-provchunkread exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.04s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.730s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'ids = append(append(ids, step.next), chunks...)' -> 'ids = append(append(ids, step.next), chunks[:0]...)')
=== M9-noprov exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.04s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   --- FAIL: TestAScopedWalkThatCannotReadTheStoreFails (0.22s)
   --- FAIL: TestTheGraphLaneOfAQueryStaysInTheCollection (0.02s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.715s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if doc == "" && len(chunks) == 0 {' -> 'if doc == "" && len(chunks) == 0 {')
=== M10-limit exit 0
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.720s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if opts.Limit > 0 && len(visited)-1 >= opts.Limit {' -> 'if false {')
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb [build failed]
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if len(opts.NodeTypes) == 0 || slices.Contains(opts.NodeType' -> 'if len(opts.NodeTypes) == 0 || true {')
=== M12-notypes exit 0
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   --- FAIL: TestAScopedWalkThatCannotReadTheStoreFails (0.20s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.688s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if len(opts.NodeTypes) == 0 || slices.Contains(' -> 'if slices.Contains(')
=== M13-depth exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.04s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.687s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if depth == opts.MaxDepth {' -> 'if depth == opts.MaxDepth+1 {')
=== M14-unscoped exit 0
   --- FAIL: TestTheGraphLaneWithoutACollectionKeepsItsHops (0.03s)
   --- FAIL: TestTheGraphLaneOfAQueryStaysInTheCollection (0.02s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.664s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if collection == "" {' -> 'if false {')
=== M15-direction exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.04s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.689s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if next == id {' -> '(retiré)')
=== M16-sqlcoll exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.07s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.07s)
   --- FAIL: TestDocumentsInCollectionQuery (0.00s)
   --- FAIL: TestTheGraphLaneOfAQueryStaysInTheCollection (0.02s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.827s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'WHERE c.name = ? AND' -> 'WHERE c.name <> ? AND')
=== M17-seen exit 0
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.738s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if _, seen := visited[step.next]; seen || !scope.mayEnter(st' -> 'if !scope.mayEnter(step) {')
=== M18-seedskip exit 0
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.04s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.725s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if depth == 0 {' -> '(retiré)')
=== M19-docsprov exit 0
   --- FAIL: TestTheGraphLaneOfAScopedQueryDoesNotCrossAnotherCollectionsRelation (0.05s)
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.05s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.801s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'docIDs = append(docIDs, doc)' -> 'docIDs = append(docIDs, doc[:0])')
=== M20-sourcechunk exit 0
   --- FAIL: TestALaneStepIsJudgedByItsNodeAndItsProvenance (0.00s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.685s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'chunks = append(chunks, src)' -> 'chunks = append(chunks, src[:0])')
=== M21-unread-node exit 0
   --- FAIL: TestALaneStepIsJudgedByItsNodeAndItsProvenance (0.00s)
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.802s
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if !ok {' -> 'if !ok {')
=== M11-types exit 0
   --- FAIL: TestTheScopedWalkKeepsNodeTypesAndLimit (0.05s)
   FAIL
   FAIL	github.com/liliang-cn/cortexdb/v2/pkg/cortexdb	0.811s
   FAIL
   handmut: tué (pkg/cortexdb/query_graph_scope.go : 'if len(opts.NodeTypes) == 0 || slices.Contains(opts.NodeType' -> 'if len(opts.NodeTypes) == 0 ||
```

## Classe

Entrées qui lancent la voie graphe de `db.Query` : l'outil `cortex_query` (MCP, `CallTool`, `POST /v1/tools/cortex_query`) et `POST /v1/query`, toutes par `runGraphQueryPrefetch`, le seul point de passage. Pas de RPC typé `Query` en gRPC. `search_chunks_by_entities` (sans collection) et l'expansion de `knowledge_search` (#70) ne changent pas.

## precommit

`make precommit REPO=<copie> BASE=origin/main` : OK sur `1b655c7` (gofmt, vet, `-race` de `pkg/cortexdb` 382 s et `pkg/httpapi`, `patchcov` 0 bloc sans test). Premier passage : 3 blocs sans test (`source_chunk_id`, nœud non lu, erreur de `eachRow`), couverts par `TestALaneStepIsJudgedByItsNodeAndItsProvenance` et par `errors.Join` dans `documentsInCollection` (plus de branche sur l'erreur de lecture).
