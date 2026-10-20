#!/usr/bin/env python3 -I
"""Preuve « failing first » de GC-10 (la voie graphe de cortex_query reste dans
la collection), 20 octobre 2026. Une mutation par règle, qui compile ;
`go test ./pkg/cortexdb/ -run Query`, le test nommé doit tomber. Le fichier est
restauré après chaque mutation. Usage : python3 -I mutate_10.py <worktree>."""
import subprocess
import sys
from pathlib import Path

wt = Path(sys.argv[1])
f = wt / "pkg/cortexdb/query_api.go"
T = "TestTheGraphLaneOfAQueryStaysInTheCollection"
MUT = [
    ("Collection ignorée", '\tif collection != "" {\n\t\tids := make', '\tif collection != "" && false {\n\t\tids := make', [T]),
    ("Chunks hors collection gardés", "\t\t\tif _, ok := inScope[chunk.ID]; ok {", "\t\t\tif _, ok := inScope[chunk.ID]; !ok {", [T]),
    ("Collection illisible ignorée", "\t\tinScope, err := db.chunksInScope(ctx, ids, collection, nil)\n\t\tif err != nil {", "\t\tinScope, err := db.chunksInScope(ctx, ids, collection, nil)\n\t\tif err != nil && false {", [T]),
    ("Collection non transmise à la voie", "db.runGraphQueryPrefetch(ctx, req.Collection, entityNames,", 'db.runGraphQueryPrefetch(ctx, "", entityNames,', [T]),
]
src = f.read_text()
ok = True
print("| Règle désactivée | Mutation | Test qui échoue |\n| :- | :- | :- |")
for name, old, new, expected in MUT:
    if src.count(old) != 1:
        print(f"| {name} | **motif trouvé {src.count(old)} fois** | |"); ok = False; continue
    f.write_text(src.replace(old, new))
    try:
        p = subprocess.run(["go", "test", "./pkg/cortexdb/", "-run", "Query"], cwd=wt, capture_output=True, text=True)
    finally:
        f.write_text(src)
    failed = sorted({l.split()[2] for l in p.stdout.splitlines() if l.startswith("--- FAIL")})
    missing = [t for t in expected if t not in failed]
    if missing or ("build failed" in p.stdout + p.stderr):
        ok = False
    print(f"| {name} | `{new.strip().splitlines()[0][:60]}` | {', '.join(failed) or '**aucun**'}{' **(build)**' if 'build failed' in p.stdout + p.stderr else ''} |")
p = subprocess.run(["go", "test", "./pkg/cortexdb/", "-run", "Query"], cwd=wt, capture_output=True, text=True)
print("\nArbre restauré, tests Query verts :", p.returncode == 0)
sys.exit(0 if ok and p.returncode == 0 else 1)
