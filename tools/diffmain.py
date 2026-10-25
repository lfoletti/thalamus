#!/usr/bin/env python3
"""diffmain : joue les mêmes scénarios sur la base et sur la tête, puis compare.

Le relecteur amont rejoue chaque scénario sur `main` et sur la branche, puis
compare. Cet outil fait la même chose avant la poussée (docs/pipeline_pr.md,
section 1, niveau 2).

Un scénario est un fichier `*_test.go` posé sous `<scenarios>/<paquet>/`, par
exemple `../thalamus_local/docs/pr/lot_C/scenarios/pkg/cortexdb/reingest_scenario_test.go`. Il
n'affirme rien : il écrit une ligne `SCENARIO …` par observation avec
`t.Logf`, et n'utilise que l'API de la base, pour compiler sur les deux arbres.

Le script crée une copie de travail de la base (`git worktree add --detach`),
pose les scénarios sur les deux arbres, lance `go test -v -run <motif>` dans
chaque paquet, retire les scénarios et la copie, puis affiche le diff unifié
des lignes `SCENARIO` et des lignes `--- FAIL`. Sortie 0 si les deux arbres
disent la même chose, 1 s'ils diffèrent, 2 si un arbre ne compile pas.

Usage :
    python3 -I tools/diffmain.py --repo <copie> --base origin/main \
        --scenarios ../thalamus_local/docs/pr/lot_C/scenarios [--run Scenario]
"""

from __future__ import annotations

import argparse
import difflib
import os
import shutil
import subprocess
import sys
import tempfile


def sh(args: list[str], cwd: str, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
    return subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True)


def scenario_files(root: str) -> list[str]:
    """Chemins relatifs à `root` des fichiers `*_test.go` sous `root`."""
    found: list[str] = []
    for base, _dirs, files in os.walk(root):
        for name in files:
            if name.endswith("_test.go"):
                found.append(os.path.relpath(os.path.join(base, name), root))
    return sorted(found)


def place(tree: str, root: str, files: list[str]) -> list[str]:
    """Copie les scénarios dans l'arbre ; refuse d'écraser un fichier existant."""
    placed: list[str] = []
    for rel in files:
        dst = os.path.join(tree, rel)
        if os.path.exists(dst):
            raise SystemExit(f"diffmain: {dst} existe déjà dans l'arbre, le scénario ne l'écrase pas")
        os.makedirs(os.path.dirname(dst), exist_ok=True)
        shutil.copyfile(os.path.join(root, rel), dst)
        placed.append(dst)
    return placed


def remove(placed: list[str]) -> None:
    for path in placed:
        try:
            os.remove(path)
        except FileNotFoundError:
            pass


def observe(tree: str, files: list[str], run: str, label: str) -> list[str]:
    """Lance les scénarios de chaque paquet et rend les lignes à comparer."""
    env = dict(os.environ)
    # Les scénarios n'atteignent jamais un serveur ni le cerveau de la machine
    # (même canari que `make precommit`).
    env.update(CORTEXDB_REMOTE="127.0.0.1:1", CORTEXDB_GRPC_TOKEN="canary", CORTEXDB_PATH="/dev/null/canary")
    lines: list[str] = []
    for pkg in sorted({os.path.dirname(rel) for rel in files}):
        proc = sh(["go", "test", "-count=1", "-v", "-run", run, "./" + pkg + "/"], cwd=tree, env=env)
        out = proc.stdout + proc.stderr
        if "[build failed]" in out or "setup failed" in out or ("FAIL" in out and "--- FAIL" not in out and "SCENARIO" not in out):
            sys.stderr.write(f"diffmain: {label} : {pkg} ne compile pas ou ne démarre pas :\n{out}\n")
            raise SystemExit(2)
        for raw in out.splitlines():
            if "SCENARIO " in raw:
                lines.append(pkg + ": " + raw[raw.index("SCENARIO ") + len("SCENARIO "):].rstrip())
            elif raw.startswith("--- FAIL") or raw.startswith("panic:"):
                lines.append(pkg + ": " + raw.split(" (")[0].strip())
    return lines


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--repo", required=True, help="copie de travail qui porte la tête")
    ap.add_argument("--base", default="origin/main", help="révision de la base (défaut : origin/main)")
    ap.add_argument("--scenarios", required=True, help="dossier des scénarios, chemins relatifs à la racine du dépôt")
    ap.add_argument("--run", default="Scenario", help="motif -run des tests de scénario (défaut : Scenario)")
    args = ap.parse_args()

    repo = os.path.abspath(args.repo)
    root = os.path.abspath(args.scenarios)
    files = scenario_files(root)
    if not files:
        sys.stderr.write(f"diffmain: aucun scénario sous {root}\n")
        return 2
    head = sh(["git", "rev-parse", "--short", "HEAD"], cwd=repo).stdout.strip()
    if sh(["git", "rev-parse", "--verify", args.base + "^{commit}"], cwd=repo).returncode != 0:
        sys.stderr.write(f"diffmain: base inconnue : {args.base}\n")
        return 2

    work = tempfile.mkdtemp(prefix="diffmain-")
    base_tree = os.path.join(work, "base")
    added = sh(["git", "worktree", "add", "--detach", base_tree, args.base], cwd=repo)
    if added.returncode != 0:
        sys.stderr.write(added.stderr)
        return 2
    placed_head: list[str] = []
    placed_base: list[str] = []
    try:
        placed_base = place(base_tree, root, files)
        placed_head = place(repo, root, files)
        base_lines = observe(base_tree, files, args.run, args.base)
        head_lines = observe(repo, files, args.run, head)
    finally:
        remove(placed_head)
        remove(placed_base)
        sh(["git", "worktree", "remove", "--force", base_tree], cwd=repo)
        shutil.rmtree(work, ignore_errors=True)

    diff = list(difflib.unified_diff(base_lines, head_lines, fromfile=args.base, tofile=head, lineterm="", n=1))
    if not diff:
        print(f"diffmain: identique ({len(head_lines)} lignes de scénario, {args.base} contre {head})")
        return 0
    print("\n".join(diff))
    print(f"diffmain: {args.base} et {head} diffèrent. Chaque ligne est voulue et nommée dans le corps de la PR, ou c'est un défaut.")
    return 1


if __name__ == "__main__":
    sys.exit(main())
