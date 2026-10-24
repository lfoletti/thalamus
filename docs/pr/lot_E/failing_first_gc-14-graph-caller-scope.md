# Failing first: gc-14-graph-caller-scope

Head `b416dbf`, base a591ae1 (origin/main). Worktree: `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-14-graph-caller-scope`.

## Tests on the base

Run: graph_caller_scope_test.go copied onto a591ae1. expiry_scope_test.go calls the new functions and does not compile there; mutations M1, M2 and M4 cover it. TestWithoutACallerScopeTheGraphToolsAnswerAsBefore and TestUnderACallerScopeAnExpiredDocumentOfTheCallersCollectionStaysHidden pass there by design: they hold what must not change (M7, M12).

Command, from a worktree of the base with the files copied:

    CORTEXDB_PATH=/dev/null/canary go test -count=1 -run 'CallerScope' ./pkg/cortexdb/

FAIL lines observed:

```
--- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection (0.03s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection/search_chunks_by_entities (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection/search_paths (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection/knowledge_memory_expand_entity_context (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection/expand_graph (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection/get_nodes (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection/knowledge_memory_neighbors (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection/knowledge_memory_shortest_path (0.00s)
--- FAIL: TestUnderACallerScopeAGraphWalkLeavesOutAnotherUsersMemory (0.03s)
--- FAIL: TestUnderACallerScopeAnUnreadableMemoryIsHidden (0.04s)
--- FAIL: TestUnderACallerScopeAnEdgeOnlyAnotherCollectionBacksIsLeftOut (0.04s)
--- FAIL: TestUnderACallerScopeAnUnreadableCollectionHidesTheDocument (0.05s)
--- FAIL: TestUnderACallerScopeAnUnreadableExpiryStillHidesAnotherCollection (0.04s)
```

## Hand mutations

One mutation per rule, through `tools/handmut.py` (copy and restore). Script: `docs/pr/lot_E/mut_gc-14-graph-caller-scope.py`. Replay from `/Users/mini/Documents/granular_cortex`:

    python3 -I docs/pr/lot_E/mut_gc-14-graph-caller-scope.py /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-14-graph-caller-scope

KILLED means handmut exited 0 on a `--- FAIL` line; the first FAIL lines are listed under each mutation.

```
M1 collection rule removed: exit 0 KILLED
    --- FAIL: TestADocumentIsOutOfScopeUnlessEveryChunkIsInAPermittedCollection (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection (0.04s)
    --- FAIL: TestUnderACallerScopeAnEdgeOnlyAnotherCollectionBacksIsLeftOut (0.06s)
    --- FAIL: TestUnderACallerScopeAnUnreadableExpiryStillHidesAnotherCollection (0.05s)
M2 chunkless document shown: exit 0 KILLED
    --- FAIL: TestADocumentIsOutOfScopeUnlessEveryChunkIsInAPermittedCollection (0.00s)
M3 scope ignored in hiddenAmong: exit 0 KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection (0.03s)
    --- FAIL: TestUnderACallerScopeAnEdgeOnlyAnotherCollectionBacksIsLeftOut (0.06s)
    --- FAIL: TestUnderACallerScopeAnUnreadableCollectionHidesTheDocument (0.05s)
    --- FAIL: TestUnderACallerScopeAnUnreadableExpiryStillHidesAnotherCollection (0.05s)
M4 collections query error dropped: exit 0 KILLED
    --- FAIL: TestAScopeWithNoKnowledgeRuleHidesNoDocument (0.04s)
    --- FAIL: TestUnderACallerScopeAnUnreadableCollectionHidesTheDocument (0.04s)
M5 hiddenAmong open on a scope error: exit 0 KILLED
    --- FAIL: TestUnderACallerScopeAnUnreadableCollectionHidesTheDocument (0.03s)
M6 advisory expiry error skips the scope: exit 0 KILLED
    --- FAIL: TestUnderACallerScopeAnUnreadableExpiryStillHidesAnotherCollection (0.03s)
M7 scope replaces the expiry: exit 0 KILLED
    --- FAIL: TestUnderACallerScopeAnExpiredDocumentOfTheCallersCollectionStaysHidden (0.04s)
M8 memory rule not a cause: exit 0 KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection (0.03s)
    --- FAIL: TestUnderACallerScopeAGraphWalkLeavesOutAnotherUsersMemory (0.03s)
    --- FAIL: TestUnderACallerScopeAnUnreadableMemoryIsHidden (0.04s)
M9 memory rule ignored: exit 0 KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection (0.03s)
    --- FAIL: TestUnderACallerScopeAGraphWalkLeavesOutAnotherUsersMemory (0.03s)
M10 unreadable memory shown: exit 0 KILLED
    --- FAIL: TestUnderACallerScopeAnUnreadableMemoryIsHidden (0.03s)
    --- FAIL: TestAWalkHidesEveryMemoryWhenTheMemoriesCannotBeRead (0.03s)
M11 review gate dropped from memory nodes: exit 0 KILLED
    --- FAIL: TestAProposedOrRejectedMemoryIsOutOfTheGraphWalks (0.04s)
M12 review status hides without the gate: exit 0 KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection (0.03s)
    --- FAIL: TestWithoutACallerScopeTheGraphToolsAnswerAsBefore (0.04s)
    --- FAIL: TestUnderACallerScopeAGraphWalkLeavesOutAnotherUsersMemory (0.02s)
```
