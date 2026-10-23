# Failing first: gc-16-allowlist-graph-tools

Head `5e87717`, base gc-15-graph-scoped-entities (3c4509d). Worktree: `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-16-allowlist-graph-tools`.

## Tests on the base

Run: the two end-to-end test files copied onto gc-15. allowlist_graph_test.go reads the new facade field and does not compile there; mutations P1 to P6 cover it. On gc-15 every graph walk is refused: 403 over REST, PermissionDenied over gRPC.

Command, from a worktree of the base with the files copied:

    CORTEXDB_PATH=/dev/null/canary go test -count=1 -run 'WalksTheGraphOfItsCollection' ./pkg/rpcserver/ ./pkg/httpapi/

FAIL lines observed:

```
--- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC (0.04s)
--- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverHTTP (0.04s)
```

## Hand mutations

One mutation per rule, through `tools/handmut.py` (copy and restore). Script: `docs/pr/lot_E/mut_gc-16-allowlist-graph-tools.py`. Replay from `/Users/mini/Documents/granular_cortex`:

    python3 -I docs/pr/lot_E/mut_gc-16-allowlist-graph-tools.py /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-16-allowlist-graph-tools

KILLED means handmut exited 0 on a `--- FAIL` line; the first FAIL lines are listed under each mutation.

```
P1 the facade still asks for a gate: KILLED
    --- FAIL: TestTheGraphWalksAreOpenToAConfinedKeyAndConfinedByTheFacade (0.00s)
P2 a walk off the allowlist: KILLED
    --- FAIL: TestTheGraphWalksAreOpenToAConfinedKeyAndConfinedByTheFacade (0.00s)
P3 the table hides the facade: KILLED
    --- FAIL: TestTheGraphWalksAreOpenToAConfinedKeyAndConfinedByTheFacade (0.00s)
P4 an argument left unclassified: KILLED
    --- FAIL: TestTheGraphWalksAreOpenToAConfinedKeyAndConfinedByTheFacade (0.00s)
    --- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
P5 find_nodes without its reason: KILLED
    --- FAIL: TestTheWholeGraphReadsStayRefusedToAConfinedKey (0.00s)
    --- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
P6 a whole-graph read without its reason: KILLED
    --- FAIL: TestTheWholeGraphReadsStayRefusedToAConfinedKey (0.00s)
    --- FAIL: TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys (0.00s)
P7 the facade ignores the scope (documents): KILLED
    --- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC (0.05s)
    --- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverHTTP (0.04s)
P8 the facade serves every entity: KILLED
    --- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC (0.04s)
    --- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverHTTP (0.04s)
P9 the facade serves the entity whole: KILLED
    --- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC (0.04s)
    --- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverHTTP (0.06s)
P10 the facade ignores the memory rule: KILLED
    --- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverGRPC (0.05s)
    --- FAIL: TestAConfinedKeyWalksTheGraphOfItsCollectionOnlyOverHTTP (0.04s)
```
