#!/usr/bin/env python3 -I
"""Preuve « failing first » des câblages #73 (gRPC) et #74 (REST), 20 octobre 2026.

Les tests de câblage ne compilent pas sans le code (symboles nouveaux), donc
chaque règle est désactivée par une mutation qui compile, et le test attendu
doit tomber. Usage : python3 -I mutate_wiring.py grpc|rest <worktree>.
"""
import subprocess
import sys
from pathlib import Path

port, wt = sys.argv[1], Path(sys.argv[2])

GRPC = ("pkg/rpcserver", "pkg/rpcserver/ownership.go", [
    ("Porte non forcée avant l'appel",
     "\targs := key.ConfineToolArgs(call.GetName(), []byte(call.GetArgsJson()))",
     "\targs := []byte(call.GetArgsJson())",
     ["TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC"]),
    ("Liste blanche non consultée",
     "\tif err := key.AuthorizeToolArgs(call.GetName(), args); err != nil {",
     "\tif err := key.AuthorizeToolArgs(call.GetName(), args); err != nil && false {",
     ["TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC", "TestAConfinedKeyMayCallToolsWithinItsScope"]),
    ("Arguments forcés non transmis à l'outil",
     "\tcall.ArgsJson = string(args)\n",
     "\tcall.ArgsJson = call.GetArgsJson()\n",
     ["TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC"]),
])
REST = ("pkg/httpapi", "pkg/httpapi/policy.go", [
    ("Porte non forcée avant l'appel",
     "\t\t\t\tbody = key.ConfineToolArgs(tool, body)\n",
     "\t\t\t\tbody = append([]byte(nil), body...)\n",
     ["TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST"]),
    ("Liste blanche non consultée",
     "\t\t\t\tif err := key.AuthorizeToolArgs(tool, body); err != nil {",
     "\t\t\t\tif err := key.AuthorizeToolArgs(tool, body); err != nil && false {",
     ["TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST"]),
    ("Sans base, la garde est sautée",
     "\t\t\t\tif db == nil {\n\t\t\t\t\twriteError(w, http.StatusInternalServerError,",
     "\t\t\t\tif db == nil && false {\n\t\t\t\t\twriteError(w, http.StatusInternalServerError,",
     ["TestTheToolRouteRefusesAConfinedKeyWhenItHasNoDatabaseToGuardWith"]),
    ("Garde des lignes sautée",
     "\t\t\t\tif err := key.Scope.GuardToolRows(r.Context(), db, tool, body); err != nil {",
     "\t\t\t\tif err := key.Scope.GuardToolRows(r.Context(), db, tool, body); err != nil && false {",
     ["TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST"]),
    ("Corps forcé non transmis à l'outil",
     "\t\t\t\tr.Body = io.NopCloser(bytes.NewReader(body))\n\t\t\t\tr.ContentLength = int64(len(body))\n",
     "\t\t\t\t_ = io.NopCloser\n",
     ["TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST"]),
    ("Corps non objet admis",
     "\t\t\t\tif len(bytes.TrimSpace(body)) > 0 && !(isJSONObject(body) && json.Valid(body)) {",
     "\t\t\t\tif len(bytes.TrimSpace(body)) > 0 && !(isJSONObject(body) && json.Valid(body)) && false {",
     ["TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP"]),
    ("Corps vide traité comme une erreur de syntaxe",
     "\t\t\t\tif len(bytes.TrimSpace(body)) > 0 && !(isJSONObject(body) && json.Valid(body)) {",
     "\t\t\t\tif len(bytes.TrimSpace(body)) >= 0 && !(isJSONObject(body) && json.Valid(body)) {",
     ["TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP"]),
])
pkg, fname, MUTATIONS = GRPC if port == "grpc" else REST


def run_tests():
    p = subprocess.run(["go", "test", "./" + pkg + "/"], cwd=wt, capture_output=True, text=True)
    failed = {ln.split()[2] for ln in p.stdout.splitlines() if ln.startswith("--- FAIL")}
    return p.returncode, failed, p.stdout + p.stderr


ok = True
print("| Règle désactivée | Mutation | Test qui échoue |")
print("| :- | :- | :- |")
for rule, before, after, expected in MUTATIONS:
    f = wt / fname
    src = f.read_text()
    if src.count(before) != 1:
        print(f"| {rule} | **motif introuvable** | |")
        ok = False
        continue
    f.write_text(src.replace(before, after))
    try:
        code, failed, out = run_tests()
    finally:
        f.write_text(src)
    if "build failed" in out:
        print(f"| {rule} | `{fname}` | **ne compile pas** |")
        ok = False
        continue
    missing = [t for t in expected if t not in failed]
    mark = "" if not missing else " **(attendu non tombé : " + ", ".join(missing) + ")**"
    if missing:
        ok = False
    print(f"| {rule} | `{fname}` : `{after.strip().splitlines()[-1][:60]}` | {', '.join(sorted(failed)) or '**aucun**'}{mark} |")

code, failed, _ = run_tests()
print()
print("Arbre restauré, suite verte :", code == 0 and not failed)
sys.exit(0 if ok and code == 0 else 1)
