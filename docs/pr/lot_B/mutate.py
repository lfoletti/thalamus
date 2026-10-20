import subprocess, sys
M = [
 ("R1 declared field",      "pkg/authz/tools.go", "if !declaredFields[c.field] {", "if false {"),
 ("R1 unknown tool",        "pkg/authz/tools.go", "if !known {\n\t\treturn fmt.Errorf(\"%w: %s is not a tool this server defines\"", "if false {\n\t\treturn fmt.Errorf(\"%w: %s is not a tool this server defines\""),
 ("R2 nil lookup",          "pkg/authz/tools.go", "if lookup == nil {", "if false {"),
 ("R2 argument unset",      "pkg/authz/tools.go", "if top == \"\" {", "if false {"),
 ("R2 wrong value",         "pkg/authz/tools.go", "if !c.allows(top) {", "if false && !c.allows(top) {"),
 ("R2 nested copy",         "pkg/authz/tools.go", "for _, v := range nested {", "for _, v := range []string(nil) {"),
 ("R2 other spelling top",  "pkg/authz/tools.go", "if k != field && strings.EqualFold(k, field) {", "if false {"),
 ("R2 other spelling deep", "pkg/authz/tools.go", "if strings.EqualFold(k, field) {\n\t\t\t\tif s, _ := inner.(string)", "if k == field {\n\t\t\t\tif s, _ := inner.(string)"),
 ("R3 whole row guard",     "pkg/authz/rows.go",  "func (s Scope) GuardToolRows(ctx context.Context, db RowLoader, tool string, args []byte) error {", "func (s Scope) GuardToolRows(ctx context.Context, db RowLoader, tool string, args []byte) error {\n\treturn nil"),
 ("R3 knowledge row",       "pkg/authz/rows.go",  "func (s Scope) PermitsKnowledge(collection string) bool {", "func (s Scope) PermitsKnowledge(collection string) bool {\n\treturn true"),
 ("R3 memory row",          "pkg/authz/rows.go",  "func (s Scope) PermitsMemory(userID, memoryScope, namespace string) bool {", "func (s Scope) PermitsMemory(userID, memoryScope, namespace string) bool {\n\treturn true"),
]
for name, path, old, new in M:
    src = open(path, encoding="utf-8").read()
    if src.count(old) != 1:
        print(f"{name:26s} SKIP (pattern found {src.count(old)} times)"); continue
    open(path, "w", encoding="utf-8").write(src.replace(old, new))
    r = subprocess.run(["go", "test", "./pkg/authz/"], capture_output=True, text=True)
    fails = [l for l in r.stdout.splitlines() if l.startswith("--- FAIL")]
    verdict = "KILLED by " + ", ".join(f.split()[2] for f in fails) if r.returncode != 0 else "SURVIVED (no test fails)"
    if r.returncode != 0 and not fails:
        verdict = "KILLED (build or vet error): " + (r.stderr or r.stdout).strip().splitlines()[0][:100]
    print(f"{name:26s} {verdict}")
    subprocess.run(["git", "checkout", "984cda7", "--", path], check=True)
print(subprocess.run(["git", "status", "--short"], capture_output=True, text=True).stdout or "tree clean")
