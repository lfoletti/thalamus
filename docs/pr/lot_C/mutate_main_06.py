#!/usr/bin/env python3 -I
"""Rejeu à la main des mutants des paquets main de #77 (lecture de CORTEXDB_REVIEW_GATE)
que gremlins laisse vivre. Usage : python3 -I mutate_main_06.py <worktree>."""
import subprocess, sys
from pathlib import Path
wt = Path(sys.argv[1])
G = "cmd/cortexdb-grpc/review_gate.go"; E = "cmd/cortexdb-mcp-stdio/embedder.go"
MUT = [
 ("grpc 17 variable vide lue comme mise", G, '\tif v == "" {\n\t\treturn false, nil', '\tif v != "" {\n\t\treturn false, nil'),
 ("grpc 21 valeur invalide admise", G, '\ton, err := strconv.ParseBool(v)\n\tif err != nil {', '\ton, err := strconv.ParseBool(v)\n\tif err == nil {'),
 ("grpc 31 erreur de lecture ignorée", G, '\ton, err := reviewGateFromEnv()\n\tif err != nil {', '\ton, err := reviewGateFromEnv()\n\tif err == nil {'),
 ("stdio 58 variable vide lue comme mise", E, 'if v := os.Getenv("CORTEXDB_REVIEW_GATE"); v != "" {', 'if v := os.Getenv("CORTEXDB_REVIEW_GATE"); v == "" {'),
 ("stdio 60 valeur invalide admise", E, '\t\ton, err := strconv.ParseBool(v)\n\t\tif err != nil {\n\t\t\treturn nil, fmt.Errorf("invalid CORTEXDB_REVIEW_GATE', '\t\ton, err := strconv.ParseBool(v)\n\t\tif err == nil {\n\t\t\treturn nil, fmt.Errorf("invalid CORTEXDB_REVIEW_GATE'),
]
ok = True
print("| Mutation | Test qui échoue |\n| :- | :- |")
for name, rel, old, new in MUT:
    f = wt / rel; src = f.read_text()
    assert src.count(old) == 1, (name, src.count(old))
    f.write_text(src.replace(old, new))
    try:
        p = subprocess.run(["go", "test", "./cmd/cortexdb-grpc/", "./cmd/cortexdb-mcp-stdio/", "-run", "Gate|Review|Env"], cwd=wt, capture_output=True, text=True)
    finally:
        f.write_text(src)
    fails = sorted({l.split()[2] for l in p.stdout.splitlines() if l.startswith("--- FAIL")})
    if not fails: ok = False
    print(f"| `{name}` | {', '.join(fails) or '**aucun**'} |")
sys.exit(0 if ok else 1)
