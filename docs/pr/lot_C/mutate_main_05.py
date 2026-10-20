#!/usr/bin/env python3 -I
"""Rejeu à la main des mutants de `cmd/cortexdb-mcp-stdio/sweep_expired.go`
que gremlins laisse vivre (paquet main). Usage : python3 -I mutate_main_05.py <worktree>."""
import subprocess, sys
from pathlib import Path
wt = Path(sys.argv[1]); f = wt / "cmd/cortexdb-mcp-stdio/sweep_expired.go"
MUT = [
 ("distant ignoré", 'remoteConfigured(); ok {', 'remoteConfigured(); !ok {'),
 ("connexion en échec ignorée", '\tif err != nil {\n\t\tfmt.Fprintf(stderr, "cortexdb: %v\\n", err)', '\tif err == nil {\n\t\tfmt.Fprintf(stderr, "cortexdb: %v\\n", err)'),
 ("appel distant en échec ignoré", '\tif err != nil {\n\t\tfmt.Fprintf(stderr, "cortexdb: sweep expired on %s: %v\\n", addr, err)', '\tif err == nil {\n\t\tfmt.Fprintf(stderr, "cortexdb: sweep expired on %s: %v\\n", addr, err)'),
 ("rapport distant illisible ignoré", 'read report: %v\\n", addr, err)\n\t\treturn 1', 'read report: %v\\n", addr, err)\n\t\treturn 0'),
 ("rapport distant non écrit", '\twriteSweepReport(stdout, &report, asJSON)\n\treturn 0', '\treturn 0'),
 ("chemin par défaut", 'if dbPath == "" {', 'if dbPath != "" {'),
 ("ouverture en échec ignorée", '\tif err != nil {\n\t\tfmt.Fprintf(stderr, "cortexdb: open', '\tif err == nil {\n\t\tfmt.Fprintf(stderr, "cortexdb: open'),
 ("JSON non écrit", '\tif asJSON {\n\t\t_ = json.NewEncoder(stdout).Encode(report)\n\t\treturn\n\t}', '\tif asJSON {\n\t\treturn\n\t}'),
 ("rapport texte non écrit", '\tfmt.Fprintf(stdout, "sweep expired: documents=%d memories=%d dryRun=%v\\n",', '\tfmt.Fprintf(io.Discard, "sweep expired: documents=%d memories=%d dryRun=%v\\n",'),
 ("rapport local non écrit", '\tif report != nil {\n\t\twriteSweepReport(stdout, report, asJSON)\n\t}', '\tif report == nil {\n\t\twriteSweepReport(stdout, report, asJSON)\n\t}'),
 ("échec du balayage ignoré", '\tif err != nil {\n\t\tfmt.Fprintf(stderr, "cortexdb: sweep expired: %v', '\tif err == nil {\n\t\tfmt.Fprintf(stderr, "cortexdb: sweep expired: %v'),
]
src = f.read_text(); ok = True
print("| Mutation | Test qui échoue |\n| :- | :- |")
for name, old, new in MUT:
    assert src.count(old) == 1, (name, src.count(old))
    f.write_text(src.replace(old, new))
    try:
        p = subprocess.run(["go", "test", "./cmd/cortexdb-mcp-stdio/", "-run", "Sweep|Remote|Unreadable"], cwd=wt, capture_output=True, text=True)
    finally:
        f.write_text(src)
    fails = sorted({l.split()[2] for l in p.stdout.splitlines() if l.startswith("--- FAIL")})
    if not fails: ok = False
    print(f"| `{name}` : `{new.strip().splitlines()[0][:50]}` | {', '.join(fails) or '**aucun**'} |")
sys.exit(0 if ok else 1)
