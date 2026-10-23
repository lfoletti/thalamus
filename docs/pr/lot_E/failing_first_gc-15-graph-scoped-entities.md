# Failing first: gc-15-graph-scoped-entities

Head `3c4509d`, base gc-14-graph-caller-scope (b416dbf). Worktree: `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-15-graph-scoped-entities`.

## Tests on the base

Run: graph_caller_scope_test.go and graph_scoped_entities_test.go of the head copied onto gc-14, without TestUnderACallerScopeSeedsThatCannotBeReadAreNotWalkedFrom, which calls servedSeeds and does not compile there (mutation N12 covers it). The search_chunks_by_entities row of the entity table passes on gc-14: that tool serves chunks, not entity nodes; its seed rule fails in TestAWalkDoesNotStartFromANodeTheScopeHides.

Command, from a worktree of the base with the files copied:

    CORTEXDB_PATH=/dev/null/canary go test -count=1 -run 'CallerScope|Entit.*Scope|Scope.*Entit|PathSeed|AWalkDoesNotStart|NoCollectionList' ./pkg/cortexdb/

FAIL lines observed:

```
--- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows (0.05s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows/search_paths (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows/knowledge_memory_expand_entity_context (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows/expand_graph (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows/get_nodes (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows/knowledge_memory_neighbors (0.00s)
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows/knowledge_memory_shortest_path (0.00s)
--- FAIL: TestUnderACallerScopeAnEntityKeepsOnlyItsNameAndType (0.03s)
--- FAIL: TestAPathSeedTheScopeHidesIsReportedUnresolved (0.03s)
--- FAIL: TestAWalkDoesNotStartFromANodeTheScopeHides (0.03s)
--- FAIL: TestUnderAKnowledgeRuleAnEntityWithNoCollectionListIsNotServed (0.05s)
```

## Hand mutations

One mutation per rule, through `tools/handmut.py` (copy and restore). Script: `docs/pr/lot_E/mut_gc-15-graph-scoped-entities.py`. Replay from `/Users/mini/Documents/granular_cortex`:

    python3 -I docs/pr/lot_E/mut_gc-15-graph-scoped-entities.py /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-15-graph-scoped-entities

KILLED means handmut exited 0 on a `--- FAIL` line; the first FAIL lines are listed under each mutation.

```
N1 entity rule removed: KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows (0.06s)
    --- FAIL: TestAPathSeedTheScopeHidesIsReportedUnresolved (0.05s)
    --- FAIL: TestAWalkDoesNotStartFromANodeTheScopeHides (0.03s)
    --- FAIL: TestUnderAKnowledgeRuleAnEntityWithNoCollectionListIsNotServed (0.03s)
N2 entity with no list served: KILLED
    --- FAIL: TestUnderAKnowledgeRuleAnEntityWithNoCollectionListIsNotServed (0.04s)
N3 properties kept: KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows (0.04s)
    --- FAIL: TestUnderACallerScopeAnEntityKeepsOnlyItsNameAndType (0.03s)
N4 vector kept: KILLED
    --- FAIL: TestUnderACallerScopeAnEntityKeepsOnlyItsNameAndType (0.05s)
N5 memory node under the entity rule: KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnotherCollection (0.07s)
N6 chunk node under the entity rule: KILLED
    --- FAIL: TestUnderACallerScopeAnExpiredDocumentOfTheCallersCollectionStaysHidden (0.02s)
    --- FAIL: TestUnderACallerScopeAnUnreadableExpiryStillHidesAnotherCollection (0.03s)
N7 reduced copy not served: KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows (0.03s)
    --- FAIL: TestUnderACallerScopeAnEntityKeepsOnlyItsNameAndType (0.02s)
N8 original nodes returned when nothing dropped: KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows (0.04s)
    --- FAIL: TestUnderACallerScopeAnEntityKeepsOnlyItsNameAndType (0.06s)
N9 search_chunks_by_entities seeds unfiltered: KILLED
    --- FAIL: TestAWalkDoesNotStartFromANodeTheScopeHides (0.05s)
N10 expand_graph seeds unfiltered: KILLED
    --- FAIL: TestAWalkDoesNotStartFromANodeTheScopeHides (0.03s)
N11 seeds unfiltered under a memory rule alone: KILLED
    --- FAIL: TestAWalkDoesNotStartFromANodeTheScopeHides (0.04s)
N12 unreadable seeds kept: KILLED
    --- FAIL: TestUnderACallerScopeSeedsThatCannotBeReadAreNotWalkedFrom (0.03s)
N13 neighbors seed unfiltered: KILLED
    --- FAIL: TestAWalkDoesNotStartFromANodeTheScopeHides (0.03s)
N14 shortest path keeps the unreduced nodes: KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows (0.03s)
N15 path seeds unfiltered: KILLED
    --- FAIL: TestUnderACallerScopeEveryGraphToolLeavesOutAnEntityOnlyAnotherCollectionKnows (0.03s)
    --- FAIL: TestAPathSeedTheScopeHidesIsReportedUnresolved (0.02s)
```
