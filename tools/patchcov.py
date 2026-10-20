#!/usr/bin/env python3
"""Couverture du patch : les lignes changées depuis une base que les tests
n'exécutent pas. C'est la mesure que Codecov applique à une PR.

    python3 -I tools/patchcov.py [--repo cortexdb] [--base origin/main] [--race] [--declare FICHIER[:L1-L2]]...

Compare l'arbre de travail (changements committés ou non) à la base, lance
`go test -coverprofile` sur les paquets touchés, et imprime chaque bloc de
code changé que rien n'exécute. Sortie 1 s'il en reste. Une ligne qui porte
le marqueur `patchcov:ignore` est exclue, avec son bloc ; réservé à ce qui ne
peut pas tourner sous `go test` (un `os.Exit`). Les blocs situés dans le
corps de `func main()` sont affichés à part et ne font pas échouer : `go test`
ne peut pas les exécuter, et chaque PR les déclare dans son corps. Un bloc hors
`main` que rien ne peut couvrir en local (la liste DDL PostgreSQL) se passe en
`--declare pkg/core/store_postgres.go` ou `--declare fichier:L1-L2` : il est
affiché à part, ne fait pas échouer, et la PR le déclare dans sa section
`## Codecov` comme les blocs de `main`.
"""
import argparse, os, re, subprocess, sys, tempfile

def run(args, cwd):
    r = subprocess.run(args, cwd=cwd, capture_output=True, text=True)
    if r.returncode != 0 and args[0] == "git":
        sys.exit(f"patchcov: {' '.join(args)}: {r.stderr.strip()}")
    return r

def hunks(repo, base):
    out = run(["git", "diff", "-U0", base, "--", "*.go"], repo).stdout
    added, cur = {}, None
    for line in out.split("\n"):
        if line.startswith("+++ b/"):
            cur = line[6:]
        m = re.match(r"@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@", line)
        if m and cur and not cur.endswith("_test.go"):
            start, n = int(m.group(1)), int(m.group(2) or 1)
            if n:
                added.setdefault(cur, []).append((start, start + n - 1))
    return added

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", default="cortexdb")
    ap.add_argument("--base", default="origin/main")
    ap.add_argument("--race", action="store_true")
    ap.add_argument("--declare", action="append", default=[], help="bloc incouvrable en local, déclaré dans la PR : FICHIER ou FICHIER:L1-L2")
    a = ap.parse_args()
    repo = os.path.abspath(a.repo)
    changed = hunks(repo, a.base)
    if not changed:
        print("patchcov: aucun fichier .go changé par rapport à", a.base)
        return 0
    pkgs = sorted({"./" + os.path.dirname(f) + "/" for f in changed})
    module = run(["go", "list", "-m"], repo).stdout.strip()
    prof = os.path.join(tempfile.mkdtemp(prefix="patchcov-"), "cover.out")
    cmd = ["go", "test", "-coverprofile=" + prof, "-covermode=atomic"] + (["-race"] if a.race else []) + pkgs
    print("patchcov:", " ".join(cmd))
    r = subprocess.run(cmd, cwd=repo, capture_output=True, text=True)
    print(r.stdout.strip())
    if r.returncode != 0:
        print(r.stderr.strip())
        print("patchcov: les tests échouent ; corriger avant de mesurer")
        return 1
    def in_main(src, l1):
        for k in range(l1 - 1, -1, -1):
            if src[k].startswith("func "):
                return src[k].startswith("func main()")
        return False
    declared = []
    for d in a.declare:
        f, _, rng = d.partition(":")
        l1, _, l2 = rng.partition("-")
        declared.append((f, int(l1) if l1 else 1, int(l2 or l1) if l1 else 10**9))
    def is_declared(f, l1, l2):
        return any(f == df and not (l2 < dl1 or l1 > dl2) for df, dl1, dl2 in declared)
    missing, in_main_blocks, declared_blocks, ignored = [], [], [], 0
    for line in open(prof, encoding="utf-8"):
        if line.startswith("mode:") or not line.strip().endswith(" 0"):
            continue
        loc = line.split()[0]
        f = loc.split(":")[0]
        f = f[len(module) + 1:] if f.startswith(module + "/") else f
        rng = loc.split(":")[1]
        l1 = int(rng.split(",")[0].split(".")[0]); l2 = int(rng.split(",")[1].split(".")[0])
        if not any(not (l2 < s or l1 > e) for s, e in changed.get(f, [])):
            continue
        src = open(os.path.join(repo, f), encoding="utf-8").read().split("\n")
        if any("patchcov:ignore" in src[k - 1] for k in range(l1, l2 + 1)):
            ignored += 1
            continue
        block = (f, l1, l2, src[l1 - 1].strip()[:90])
        if in_main(src, l1):
            in_main_blocks.append(block)
        elif is_declared(f, l1, l2):
            declared_blocks.append(block)
        else:
            missing.append(block)
    def show(blocks):
        seen = set()
        for f, l1, l2, code in sorted(blocks):
            if (f, l1) in seen:
                continue
            seen.add((f, l1))
            print(f"  {f}:{l1}-{l2}  {code}")
        return len(seen)
    n = show(missing)
    m = d = 0
    if in_main_blocks:
        print("patchcov: dans le corps de main(), à déclarer dans la PR :")
        m = show(in_main_blocks)
    if declared_blocks:
        print("patchcov: déclarés (--declare), à nommer dans la section Codecov de la PR :")
        d = show(declared_blocks)
    print(f"patchcov: {n} bloc(s) du patch sans test" + (f", {m} dans main()" if m else "") + (f", {d} déclaré(s)" if d else "") + (f", {ignored} ignoré(s)" if ignored else "") + f" ; base {a.base}, paquets {' '.join(pkgs)}")
    return 1 if n else 0

if __name__ == "__main__":
    sys.exit(main())
