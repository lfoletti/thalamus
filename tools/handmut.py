#!/usr/bin/env python3
"""handmut : une mutation à la main, sûre.

Remplace OLD par NEW (une seule occurrence) dans FILE de la copie REPO, lance
`go test` sous le canari, puis restaure FILE depuis la copie prise avant la
mutation, jamais depuis HEAD. Le 21 octobre 2026, un `git checkout -- FILE`
après une mutation à la main a retiré un correctif non commis avec la
mutation (#72) : d'où cet outil, et son refus de muter un fichier non commis.

Sortie 0 : le mutant est tué (une ligne `--- FAIL` ou un paquet `FAIL`).
Sortie 1 : le mutant survit. Sortie 2 : refus (fichier non commis, OLD absent
ou présent plusieurs fois, restauration non identique).

Usage :
    python3 -I tools/handmut.py --repo <copie> --file <chemin> \
        --old '<texte>' --new '<texte>' [--allow-dirty] -- <args de go test>
    Exemple : -- -count=1 -run 'Ingest|KnowledgeSave' ./pkg/cortexdb/
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys

CANARY = {
    "CORTEXDB_REMOTE": "127.0.0.1:1",
    "CORTEXDB_GRPC_TOKEN": "canary",
    "CORTEXDB_PATH": "/dev/null/canary",
}


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--repo", required=True)
    ap.add_argument("--file", required=True, help="chemin relatif à REPO")
    ap.add_argument("--old", required=True)
    ap.add_argument("--new", required=True)
    ap.add_argument("--allow-dirty", action="store_true", help="mute un fichier non commis (à éviter)")
    ap.add_argument("gotest", nargs="*", help="arguments de go test, après --")
    args = ap.parse_args()
    if args.old == args.new:
        print("handmut: OLD et NEW sont identiques", file=sys.stderr)
        return 2
    path = os.path.join(args.repo, args.file)
    dirty = subprocess.run(["git", "-C", args.repo, "diff", "--quiet", "--", args.file]).returncode != 0
    if dirty and not args.allow_dirty:
        print(f"handmut: {args.file} a des changements non commis : commets-les ou passe --allow-dirty", file=sys.stderr)
        return 2
    with open(path, encoding="utf-8") as fh:
        original = fh.read()
    if original.count(args.old) != 1:
        print(f"handmut: OLD trouvé {original.count(args.old)} fois dans {args.file}, il en faut exactement une", file=sys.stderr)
        return 2
    mutated = original.replace(args.old, args.new, 1)
    with open(path, "w", encoding="utf-8") as fh:
        fh.write(mutated)
    try:
        env = dict(os.environ, **CANARY)
        run = subprocess.run(["go", "test", *(args.gotest or ["./..."])], cwd=args.repo, env=env, capture_output=True, text=True)
    finally:
        with open(path, "w", encoding="utf-8") as fh:
            fh.write(original)
    with open(path, encoding="utf-8") as fh:
        if fh.read() != original:
            print(f"handmut: {args.file} n'est pas restauré à l'identique", file=sys.stderr)
            return 2
    out = run.stdout + run.stderr
    fails = [l for l in out.splitlines() if l.startswith("--- FAIL") or l.startswith("FAIL") or l.startswith("ok ")]
    for l in fails:
        print(l)
    killed = any(l.startswith("--- FAIL") or l.startswith("FAIL") for l in fails)
    def head(text: str) -> str:
        lines = text.strip().splitlines()
        return lines[0][:60] if lines else "(retiré)"
    print(f"handmut: {'tué' if killed else 'SURVIVANT'} ({args.file} : {head(args.old)!r} -> {head(args.new)!r})")
    return 0 if killed else 1


if __name__ == "__main__":
    sys.exit(main())
