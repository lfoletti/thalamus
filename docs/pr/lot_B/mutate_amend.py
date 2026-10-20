#!/usr/bin/env python3 -I
"""Preuve « failing first » de l'amendement de #72 (liste blanche), 20 octobre 2026.

Pour chaque règle : une mutation à la main qui compile, `go test ./pkg/authz/`,
le ou les tests attendus doivent tomber. Le fichier est restauré après chaque
mutation. Usage : python3 -I mutate_amend.py <worktree de gc-04a>.
"""
import subprocess
import sys
from pathlib import Path

wt = Path(sys.argv[1])
pkg = wt / "pkg/authz"

MUTATIONS = [
    # (règle, fichier, avant, après, tests attendus en échec)
    ("R1 outil hors liste admis", "allowlist.go",
     "\trule, ok := confinedTools[tool]\n\tif ok {\n\t\treturn rule, nil",
     "\trule, ok := confinedTools[tool]\n\tif ok || tool != \"\" {\n\t\treturn rule, nil",
     ["TestAConfinedKeyIsRefusedAToolOffTheAllowlist"]),
    ("R1 argument inconnu admis", "allowlist.go",
     "\t\t\t\tif !allowed[child] {",
     "\t\t\t\tif !allowed[child] && false {",
     ["TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover"]),
    ("R1 argument refusé admis", "allowlist.go",
     "\t\t\t\tif why, no := denied[child]; no {",
     "\t\t\t\tif why, no := denied[child]; no && false {",
     ["TestAConfinedKeyIsRefusedAnArgumentTheAllowlistDoesNotCover"]),
    ("R2 champ non honoré", "allowlist.go",
     "\t\t\tif !contains(rule.scope, c.field) {",
     "\t\t\tif !contains(rule.scope, c.field) && false {",
     ["TestTheScopeRuleStillHoldsOnTheAllowlist"]),
    ("R2 argument absent", "allowlist.go",
     "\t\t\tif top == \"\" {",
     "\t\t\tif top == \"\" && false {",
     ["TestTheScopeRuleStillHoldsOnTheAllowlist"]),
    ("R2 mauvaise valeur", "allowlist.go",
     "\t\t\tif !c.allows(top) {",
     "\t\t\tif !c.allows(top) && false {",
     ["TestTheScopeRuleStillHoldsOnTheAllowlist"]),
    ("R2 copie imbriquée", "allowlist.go",
     "\t\t\tfor _, v := range nested {",
     "\t\t\tfor _, v := range nested[:0] {",
     ["TestTheScopeRuleStillHoldsOnTheAllowlist"]),
    ("R3 porte non exigée", "allowlist.go",
     "\t\t\tif on, _ := lookupKey(obj, gate).(bool); gate == \"\" || !on {",
     "\t\t\tif on, _ := lookupKey(obj, gate).(bool); gate == \"\" || !on && false {",
     ["TestAToolThatReadsBothStoresIsGatedForAConfinedKey"]),
    ("R3 porte non forcée", "allowlist.go",
     "\t\tobj[gate] = true\n\t}\n\t// Marshalling",
     "\t\tobj[gate] = false\n\t}\n\t// Marshalling",
     ["TestAToolThatReadsBothStoresIsGatedForAConfinedKey"]),
    ("R4 garde des lignes entière", "rows.go",
     "\tif s.IsZero() {\n\t\treturn nil\n\t}\n\trule, ok := confinedTools[tool]",
     "\tif s.IsZero() || true {\n\t\treturn nil\n\t}\n\trule, ok := confinedTools[tool]",
     ["TestAToolCallMayOnlyNameRowsWithinTheScope", "TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths"]),
    ("R4 supersedes oublié", "allowlist.go",
     "ids: []idField{{\"memory_id\", idMemory}, {\"supersedes\", idMemory}}",
     "ids: []idField{{\"memory_id\", idMemory}}",
     ["TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys", "TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths"]),
    ("R4 chemin imbriqué oublié", "allowlist.go",
     "\tplanIDs    = []idField{{\"plan.filters.document_ids\", idKnowledge}}",
     "\tplanIDs    = []idField{}",
     ["TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys", "TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths"]),
    ("R4 autre graphie d'un chemin", "allowlist.go",
     "\t\tif !strings.EqualFold(k, name) {",
     "\t\tif k != name {",
     ["TestTheIDsAToolNamesAreReadOffTheAllowlistedPaths"]),
    ("Complétude : argument de schéma oublié", "allowlist.go",
     "args: paths(queryArgs, walkArgs, planArgs, pprArgs, recallArgs, []string{\"detail\"})",
     "args: paths(queryArgs, walkArgs, planArgs, pprArgs, recallArgs)",
     ["TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys"]),
    ("Complétude : outil ni admis ni refusé", "allowlist.go",
     "\t\trefusedTools[name] = graphOnlyReason\n",
     "\t\tif name != \"vacuum_graph\" {\n\t\t\trefusedTools[name] = graphOnlyReason\n\t\t}\n",
     ["TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys"]),
]


def run_tests():
    p = subprocess.run(["go", "test", "./pkg/authz/"], cwd=wt, capture_output=True, text=True)
    failed = {ln.split()[2] for ln in p.stdout.splitlines() if ln.startswith("--- FAIL")}
    return p.returncode, failed, p.stdout + p.stderr


ok = True
print("| Règle désactivée | Mutation | Test qui échoue |")
print("| :- | :- | :- |")
for rule, fname, before, after, expected in MUTATIONS:
    f = pkg / fname
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
    if "build failed" in out or "cannot" in out and code != 0 and not failed:
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
