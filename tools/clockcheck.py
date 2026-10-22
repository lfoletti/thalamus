#!/usr/bin/env python3
"""clockcheck : les tests du patch qui lisent l'heure, rejoués plusieurs fois.

Un test qui compare à un instant relatif à « maintenant » tombe quand une
seconde passe entre deux lectures de l'horloge (21 octobre 2026, #76 :
1 échec sur 12 sous `-race`). Cet outil trouve, dans les fichiers de test
touchés par le diff contre BASE, les fonctions `Test…` qui appellent
`time.Now` ou un auxiliaire de test qui l'appelle, puis les lance
`go test -race -count=N` sous le canari, un paquet à la fois.

Sortie 0 : tout vert (ou aucun test concerné). Sortie 1 : une ligne
`--- FAIL`. Sortie 2 : un paquet ne compile pas.

Usage :
    python3 -I tools/clockcheck.py --repo <copie> [--base origin/main] [--count 10]
"""

from __future__ import annotations

import argparse
import os
import re
import subprocess
import sys

CANARY = {
    "CORTEXDB_REMOTE": "127.0.0.1:1",
    "CORTEXDB_GRPC_TOKEN": "canary",
    "CORTEXDB_PATH": "/dev/null/canary",
}
FUNC = re.compile(r"^func (\w+)\(", re.M)
TEST = re.compile(r"^func (Test\w+)\(t \*testing\.T\)", re.M)


def functions(src: str) -> dict[str, str]:
    """Nom de chaque fonction de premier niveau → son corps (jusqu'à la suivante)."""
    out = {}
    marks = list(FUNC.finditer(src))
    for i, m in enumerate(marks):
        end = marks[i + 1].start() if i + 1 < len(marks) else len(src)
        out[m.group(1)] = src[m.start():end]
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--repo", required=True)
    ap.add_argument("--base", default="origin/main")
    ap.add_argument("--count", type=int, default=10)
    args = ap.parse_args()
    changed = subprocess.run(["git", "-C", args.repo, "diff", "--name-only", args.base, "--", "*_test.go"],
                             capture_output=True, text=True, check=True).stdout.split()
    by_pkg: dict[str, set[str]] = {}
    for rel in changed:
        path = os.path.join(args.repo, rel)
        if not os.path.exists(path):
            continue
        pkg = os.path.dirname(rel)
        # Les auxiliaires qui lisent l'heure, dans tout le paquet (un auxiliaire vit souvent dans un autre fichier).
        helpers = set()
        for name in os.listdir(os.path.join(args.repo, pkg)):
            if name.endswith("_test.go"):
                with open(os.path.join(args.repo, pkg, name), encoding="utf-8") as fh:
                    for fn, body in functions(fh.read()).items():
                        if not fn.startswith("Test") and "time.Now(" in body:
                            helpers.add(fn)
        with open(path, encoding="utf-8") as fh:
            src = fh.read()
        for fn, body in functions(src).items():
            if not TEST.match(body):
                continue
            if "time.Now(" in body or any(re.search(r"\b" + h + r"\(", body) for h in helpers):
                by_pkg.setdefault(pkg, set()).add(fn)
    if not by_pkg:
        print("clockcheck: aucun test du patch ne lit l'heure")
        return 0
    rc = 0
    env = dict(os.environ, **CANARY)
    for pkg, names in sorted(by_pkg.items()):
        pattern = "^(" + "|".join(sorted(names)) + ")$"
        print(f"clockcheck: ./{pkg}/ -count={args.count} : {', '.join(sorted(names))}")
        run = subprocess.run(["go", "test", "-race", f"-count={args.count}", "-run", pattern, f"./{pkg}/"],
                             cwd=args.repo, env=env, capture_output=True, text=True)
        out = run.stdout + run.stderr
        for l in out.splitlines():
            if l.startswith("--- FAIL") or l.startswith("FAIL") or l.startswith("ok ") or "_test.go:" in l:
                print("  " + l)
        if "[build failed]" in out or "setup failed" in out:
            rc = 2
        elif run.returncode != 0 and rc == 0:
            rc = 1
    print("clockcheck: " + ("OK" if rc == 0 else "ROUGE"))
    return rc


if __name__ == "__main__":
    sys.exit(main())
