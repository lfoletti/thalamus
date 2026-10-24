# Failing first : gc-13-find-nodes-hidden

Tête `f435282`, base `origin/main` (`a591ae1`). Règle : `find_nodes` laisse de côté les nœuds que `withoutExpiredDocuments` retire (chunks et documents d'un document expiré ou caché par la relecture, mémoires cachées), avant la limite ; `match` est la force du premier nœud servi.

## Tests sur la base (fichiers de test posés sur `origin/main`, code de la base)

Commande : `go test ./pkg/cortexdb/ -count=1 -run 'TestFindNodes|TestEveryReadToolIsGatedOrNamed|TestEveryGatedReadToolHidesAProposedDocument'`

```
--- FAIL: TestEveryReadToolIsGatedOrNamed (0.07s)
    expiry_class_test.go:89: find_nodes returned the expired document (harbour): {"matches":[{"name":"Arvid Lund","nodes":[{"id":"entity:arvid_lund","vector":[0,0,0,0,0,0
--- FAIL: TestFindNodesLeavesOutTheNodesOfAnExpiredDocument (0.03s)
    graphrag_tool_find_nodes_hidden_test.go:47: find_nodes served the expired document (EXPIREDTEXT): {"matches":[{"name":"Zephyr Gate","nodes":[{"id":"entity:zephyr_gate
    graphrag_tool_find_nodes_hidden_test.go:47: find_nodes served the expired document (chunk:old:000): {"matches":[{"name":"Zephyr Gate","nodes":[{"id":"entity:zephyr_ga
    graphrag_tool_find_nodes_hidden_test.go:47: find_nodes served the expired document (doc:old): {"matches":[{"name":"Zephyr Gate","nodes":[{"id":"entity:zephyr_gate","v
--- FAIL: TestFindNodesLeavesOutAProposedDocumentUnderTheReviewGate (0.03s)
    graphrag_tool_find_nodes_hidden_test.go:81: find_nodes served what the review gate hides (PROPOSEDTEXT): {"matches":[{"name":"Zephyr Gate","nodes":[{"id":"entity:zeph
    graphrag_tool_find_nodes_hidden_test.go:81: find_nodes served what the review gate hides (PROPOSEDMEMORY): {"matches":[{"name":"Zephyr Gate","nodes":[{"id":"entity:ze
    graphrag_tool_find_nodes_hidden_test.go:81: find_nodes served what the review gate hides (chunk:draft:000): {"matches":[{"name":"Zephyr Gate","nodes":[{"id":"entity:z
    graphrag_tool_find_nodes_hidden_test.go:81: find_nodes served what the review gate hides (memory:m-draft): {"matches":[{"name":"Zephyr Gate","nodes":[{"id":"entity:ze
--- FAIL: TestFindNodesStillFindsAnEntityThatAHiddenChunkMentions (0.02s)
    graphrag_tool_find_nodes_hidden_test.go:104: find_nodes = [entity:zephyr_gate chunk:old:000 doc:old] (exact), want the entity alone, exact
--- FAIL: TestFindNodesFillsItsLimitPastHiddenMatches (0.03s)
    graphrag_tool_find_nodes_hidden_test.go:143: find_nodes limit 1 = [chunk:old1:000] (exact), want chunk:live1:000, contains
    graphrag_tool_find_nodes_hidden_test.go:143: find_nodes limit 2 = [chunk:old1:000 chunk:live1:000] (exact), want chunk:live1:000,chunk:live2:000, contains
    graphrag_tool_find_nodes_hidden_test.go:143: find_nodes limit 3 = [chunk:old1:000 chunk:live1:000 chunk:live2:000] (exact), want chunk:live1:000,chunk:live2:000,chunk
    graphrag_tool_find_nodes_hidden_test.go:155: find_nodes on hidden matches only = [chunk:old2:000 doc:old2] ("contains"), want none
--- FAIL: TestEveryGatedReadToolHidesAProposedDocument (0.05s)
    review_r3_test.go:73: find_nodes returned the expired document (harbour): {"matches":[{"name":"Arvid Lund","nodes":[{"id":"entity:arvid_lund","vector":[0,0,0,0,0,0,0,
```

## Mutations à la main (`tools/handmut.py`, restauration depuis une copie)

Commande : `python3 -I tools/handmut.py --repo <copie> --file pkg/cortexdb/graphrag_tool_find_nodes.go --old '<OLD>' --new '<NEW>' -- -count=1 -run 'TestFindNodes|TestEveryReadToolIsGatedOrNamed|TestEveryGatedReadToolHidesAProposedDocument' ./pkg/cortexdb/`

| Mutation | OLD → NEW | Test qui tombe |
| :- | :- | :- |
| M1 filtre retiré | `kept, _ := t.db.withoutExpiredDocuments(ctx, ordered, nil, nil)` → `kept := ordered` | `TestFindNodesLeavesOutTheNodesOfAnExpiredDocument`, `TestFindNodesLeavesOutAProposedDocumentUnderTheReviewGate`, `TestFindNodesStillFindsAnEntityThatAHiddenChunkMentions`, `TestFindNodesFillsItsLimitPastHiddenMatches`, `TestEveryReadToolIsGatedOrNamed`, `TestEveryGatedReadToolHidesAProposedDocument` |
| M2 première fenêtre seule (limite avant le filtre) | `start < len(hits)` → `start < min(len(hits), limit)` | `TestFindNodesFillsItsLimitPastHiddenMatches` |
| M3 force du premier classé | `[rank[node.ID]]` → `[hits[0].rank]` | `TestFindNodesFillsItsLimitPastHiddenMatches` |
| M4 limite non appliquée à la fenêtre | `kept[:min(len(kept), limit-len(match.Nodes))]` → `kept` | `TestFindNodesFillsItsLimitPastHiddenMatches` |
| M5 force du dernier servi | `if len(match.Nodes) == 0 {` → `if true {` | `TestFindNodesFillsItsLimitPastHiddenMatches`, `TestFindNodesResolvesANameToAnID` |
| M6 ordre du classement perdu | `ordered = append(ordered, node)` → `ordered = append([]*graph.GraphNode{node}, ordered...)` | `TestFindNodesFillsItsLimitPastHiddenMatches`, `TestFindNodesResolvesANameToAnID`, `TestFindNodesPrefersTheExactNameOverOneContainingIt` |

Six mutations, six tuées. M5 a d'abord survécu (toutes les correspondances servies étaient `contains`) : le cas « Zephyr Gate shuts » (exact puis contains) a été ajouté au test, puis M5 tombe.

Journal :

```
=== M1-filter
--- FAIL: TestEveryReadToolIsGatedOrNamed (0.07s)
--- FAIL: TestFindNodesLeavesOutTheNodesOfAnExpiredDocument (0.03s)
--- FAIL: TestFindNodesLeavesOutAProposedDocumentUnderTheReviewGate (0.03s)
--- FAIL: TestFindNodesStillFindsAnEntityThatAHiddenChunkMentions (0.02s)
--- FAIL: TestFindNodesFillsItsLimitPastHiddenMatches (0.04s)
--- FAIL: TestEveryGatedReadToolHidesAProposedDocument (0.08s)
handmut: tué (pkg/cortexdb/graphrag_tool_find_nodes.go : 'kept, _ := t.db.withoutExpiredDocuments(ctx, ordered, nil, n' -> 'kept := ordered')
=== M2-firstwindow
--- FAIL: TestFindNodesFillsItsLimitPastHiddenMatches (0.04s)
handmut: tué (pkg/cortexdb/graphrag_tool_find_nodes.go : 'for start := 0; start < len(hits) && len(match.Nodes) < limi' -> 'for start := 0; start < min(len(hits), limit) && len(match.N')
=== M3-rankedstrength
--- FAIL: TestFindNodesFillsItsLimitPastHiddenMatches (0.04s)
handmut: tué (pkg/cortexdb/graphrag_tool_find_nodes.go : '[...]string{"exact", "fold", "contains"}[rank[node.ID]]' -> '[...]string{"exact", "fold", "contains"}[hits[0].rank]')
=== M4-nolimit
--- FAIL: TestFindNodesFillsItsLimitPastHiddenMatches (0.03s)
handmut: tué (pkg/cortexdb/graphrag_tool_find_nodes.go : 'kept[:min(len(kept), limit-len(match.Nodes))]' -> 'kept')
=== M5-laststrength
--- FAIL: TestFindNodesFillsItsLimitPastHiddenMatches (0.03s)
--- FAIL: TestFindNodesResolvesANameToAnID (0.02s)
handmut: tué (pkg/cortexdb/graphrag_tool_find_nodes.go : 'if len(match.Nodes) == 0 {' -> 'if true {')
=== M6-order
--- FAIL: TestFindNodesFillsItsLimitPastHiddenMatches (0.03s)
--- FAIL: TestFindNodesResolvesANameToAnID (0.02s)
--- FAIL: TestFindNodesPrefersTheExactNameOverOneContainingIt (0.02s)
handmut: tué (pkg/cortexdb/graphrag_tool_find_nodes.go : 'ordered = append(ordered, node)' -> 'ordered = append([]*graph.GraphNode{node}, ordered...)')
```

## Classe

Classe « outils qui lisent des nœuds du graphe » : `readToolsAndTheExpiry` (`pkg/cortexdb/expiry_class_test.go`). `find_nodes` y était classé « entities and relations » à tort (il lit le contenu des chunks) ; il passe en « gated », donc les deux tests de classe de #76 et #77 l'appellent (expiration et porte de relecture). Aucun autre outil de la table ne change.

## precommit

`make precommit REPO=<copie> BASE=origin/main` : OK (gofmt, vet, `-race` de `pkg/cortexdb` 509 s, `patchcov` 0 bloc sans test).
