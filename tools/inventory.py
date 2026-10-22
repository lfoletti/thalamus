#!/usr/bin/env python3
"""inventory : la liste complète de chaque classe, générée depuis l'arbre.

Un point de revue nomme un cas. Le relecteur amont teste ensuite la classe
entière du cas : tous les outils, tous les chemins de lecture, tous les chemins
d'écriture, tous les champs d'identifiant, toutes les clés réservées. Cet outil
imprime chaque classe depuis le code, pour que `failing_first.md` tienne une
ligne par élément (docs/pipeline_pr.md, section 1, niveau 2).

Classes :
    tools     chaque outil déclaré (`Name: "…"` dans les fichiers de définition)
    reads     chaque fonction hors test qui lit les documents ou les chunks
              (SQL sur `documents` ou `chunks`, Cypher sur `:document` ou
              `:chunk`, filtrage des vecteurs)
    writes    chaque fonction hors test qui appelle une écriture de document,
              de mémoire, d'entité ou de relation
    ids       chaque propriété d'outil qui porte un identifiant de ligne
              (`*_id`, `*_ids`, `supersedes`) avec son outil
    keys      chaque constante de clé réservée (`…Key = "…"`)
    promises  chaque chaîne de description ajoutée par le diff contre la base
              (texte de schéma, promesse à tenir par un test nommé)
    exports   chaque fonction ou méthode exportée de `pkg/cortexdb` dont le nom
              dit une lecture (Search…, Query…, Recall…, Aggregate…, Get…) :
              l'entrée de bibliothèque qu'aucun outil ne nomme (revue adverse
              du 21 octobre 2026, point 1)
    rpcs      chaque RPC des fichiers `proto/` : le chemin gRPC typé, à côté de
              l'outil générique `CallTool`

Chaque ligne dit si le fichier est touché par le diff contre la base (`*`).
La colonne « Test » est vide : elle se remplit à la main dans
`failing_first.md`, par un nom de test ou « hors classe : <raison> ».

Avec `--check <failing_first.md>`, l'outil ne publie pas les tables : il
compare chaque élément des classes demandées aux lignes du fichier et
imprime chaque élément sans ligne, ou avec une colonne « Test » vide. Le
code de sortie est 1 s'il en manque un, 0 sinon. Un élément est trouvé par
son nom entre accents graves dans la première colonne d'une ligne de table.

Usage :
    python3 -I tools/inventory.py --repo <copie> [--base origin/main] \
        [--class tools|reads|writes|ids|keys|promises|exports|rpcs|all] \
        [--pattern <regex>]… [--check <failing_first.md>]
"""

from __future__ import annotations

import argparse
import os
import re
import subprocess
import sys

READ_PATTERNS = [
    r"\bFROM\s+documents\b",
    r"\bFROM\s+chunks\b",
    r"\bJOIN\s+documents\b",
    r"\bJOIN\s+chunks\b",
    r"MATCH\s*\(\w*:document\b",
    r"MATCH\s*\(\w*:chunk\b",
    r"\bfilterScoredEmbeddings\(",
    r"\bSearchEmbeddings\(",
    r"\bsearchChunks\w*\(",
]

WRITE_PATTERNS = [
    r"\.(SaveKnowledge|UpdateKnowledge|DeleteKnowledge)\(",
    r"\.(SaveMemory|UpdateMemory|DeleteMemory)\(",
    r"\.(IngestDocument|ingestDocument)\(",
    r"\.(UpsertEntities|UpsertRelations|upsertEntities|upsertRelations)\(",
    r"\.(DeleteDocumentGraph|SweepExpired)\(",
    r"\b(UpsertNode|UpsertEdge|UpsertNodes|UpsertEdges)\(",
]

DEF_FILE = re.compile(r"(tooldefs|tool_defs|_defs)\.go$")
FUNC_LINE = re.compile(r"^func\s+(\([^)]*\)\s*)?(\w+)")
TOOL_NAME = re.compile(r'\bName:\s*"([a-z_0-9]+)"')
ID_PROP = re.compile(r'"([a-z0-9_]*_ids?|supersedes)"\s*:')
KEY_CONST = re.compile(r'\b(\w+Key)\s*=\s*"([a-z0-9_.:-]+)"')
STRING_LIT = re.compile(r'"((?:[^"\\]|\\.)*)"')
EXPORT_READ = re.compile(
    r"^func\s+(?:\([^)]*\)\s*)?((?:Search|Query|Hybrid|Range|Recall|Aggregate|VectorAggregate|"
    r"Find|Get|List|Neighbors|Expand|Global|Cortex|Representative|Quick|Build)\w*)\(")
RPC_LINE = re.compile(r"^\s*rpc\s+(\w+)\s*\(")
ROW_CELLS = re.compile(r"^\|(.+)\|\s*$")


def sh(args: list[str], cwd: str) -> str:
    return subprocess.run(args, cwd=cwd, capture_output=True, text=True, check=False).stdout


def go_files(repo: str, tests: bool = False) -> list[str]:
    out = sh(["git", "ls-files", "--", "pkg/*.go", "cmd/*.go"], repo).splitlines()
    files = [f for f in out if f.endswith(".go") and f.endswith("_test.go") == tests]
    return sorted(files)


def touched(repo: str, base: str) -> set[str]:
    return set(sh(["git", "diff", "--name-only", base, "--"], repo).splitlines())


def enclosing_func(lines: list[str], index: int) -> str:
    for i in range(index, -1, -1):
        m = FUNC_LINE.match(lines[i])
        if m:
            return m.group(2)
    return "(hors fonction)"


def by_function(repo: str, files: list[str], patterns: list[str]) -> list[tuple[str, int, str]]:
    """(fichier, ligne du func, nom) pour chaque fonction qui contient un motif."""
    regs = [re.compile(p) for p in patterns]
    seen: dict[tuple[str, str], int] = {}
    for rel in files:
        with open(os.path.join(repo, rel), encoding="utf-8") as fh:
            lines = fh.read().splitlines()
        for i, line in enumerate(lines):
            if any(r.search(line) for r in regs):
                name = enclosing_func(lines, i)
                key = (rel, name)
                if key not in seen:
                    seen[key] = i + 1
    return sorted((rel, line, name) for (rel, name), line in seen.items())


def tools_of(repo: str, files: list[str]) -> list[tuple[str, int, str]]:
    found: list[tuple[str, int, str]] = []
    for rel in files:
        if not DEF_FILE.search(rel):
            continue
        with open(os.path.join(repo, rel), encoding="utf-8") as fh:
            for i, line in enumerate(fh):
                m = TOOL_NAME.search(line)
                if m:
                    found.append((rel, i + 1, m.group(1)))
    return sorted(found, key=lambda t: t[2])


def ids_of(repo: str, files: list[str]) -> list[tuple[str, int, str]]:
    found: list[tuple[str, int, str]] = []
    for rel in files:
        if not DEF_FILE.search(rel):
            continue
        tool = "?"
        with open(os.path.join(repo, rel), encoding="utf-8") as fh:
            for i, line in enumerate(fh):
                m = TOOL_NAME.search(line)
                if m:
                    tool = m.group(1)
                for p in ID_PROP.findall(line):
                    found.append((rel, i + 1, f"{tool}.{p}"))
    return sorted(found, key=lambda t: t[2])


def keys_of(repo: str, files: list[str]) -> list[tuple[str, int, str]]:
    found: list[tuple[str, int, str]] = []
    for rel in files:
        with open(os.path.join(repo, rel), encoding="utf-8") as fh:
            for i, line in enumerate(fh):
                for const, value in KEY_CONST.findall(line):
                    found.append((rel, i + 1, f'{const} = "{value}"'))
    return sorted(found, key=lambda t: t[2])


def exports_of(repo: str, files: list[str]) -> list[tuple[str, int, str]]:
    found: list[tuple[str, int, str]] = []
    for rel in files:
        if not rel.startswith("pkg/cortexdb/"):
            continue
        with open(os.path.join(repo, rel), encoding="utf-8") as fh:
            for i, line in enumerate(fh):
                m = EXPORT_READ.match(line)
                if m:
                    recv = re.match(r"^func\s+\(\w*\s*\*?(\w+)\)", line)
                    name = f"{recv.group(1)}.{m.group(1)}" if recv else m.group(1)
                    found.append((rel, i + 1, name))
    return sorted(found, key=lambda t: t[2])


def rpcs_of(repo: str) -> list[tuple[str, int, str]]:
    files = sh(["git", "ls-files", "--", "proto/*.proto"], repo).splitlines()
    found: list[tuple[str, int, str]] = []
    for rel in files:
        with open(os.path.join(repo, rel), encoding="utf-8") as fh:
            for i, line in enumerate(fh):
                m = RPC_LINE.match(line)
                if m:
                    found.append((rel, i + 1, m.group(1)))
    return sorted(found, key=lambda t: t[2])


def documented(path: str) -> dict[str, str]:
    """Nom entre accents graves de la première colonne -> dernière colonne."""
    rows: dict[str, str] = {}
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            m = ROW_CELLS.match(line.rstrip("\n"))
            if not m:
                continue
            cells = [c.strip() for c in m.group(1).split("|")]
            # A name with a backtick is written between double backticks.
            for _, name in re.findall(r"(?<!`)(``?)(?!`) ?(.+?) ?(?<!`)\1(?!`)", cells[0]):
                if rows.get(name.strip(), "") == "":
                    rows[name] = cells[-1] if len(cells) > 1 else ""
    return rows


def promises_of(repo: str, base: str) -> list[tuple[str, int, str]]:
    """Chaînes longues ajoutées par le diff dans les fichiers de définition."""
    diff = sh(["git", "diff", "-U0", base, "--", "pkg/*defs*.go", "pkg/*_defs.go"], repo).splitlines()
    found: list[tuple[str, int, str]] = []
    rel, line_no = "", 0
    for raw in diff:
        if raw.startswith("+++ b/"):
            rel = raw[6:]
        elif raw.startswith("@@"):
            m = re.search(r"\+(\d+)", raw)
            line_no = int(m.group(1)) - 1 if m else 0
        elif raw.startswith("+") and not raw.startswith("+++"):
            line_no += 1
            for s in STRING_LIT.findall(raw):
                if len(s) >= 24 and len(s.split()) >= 4:
                    found.append((rel, line_no, s))
    return found


def print_table(title: str, rows: list[tuple[str, int, str]], hot: set[str]) -> None:
    print(f"\n### {title} ({len(rows)})\n")
    print("| Élément | Fichier | Touché | Test |")
    print("| :- | :- | :- | :- |")
    for rel, line, name in rows:
        mark = "*" if rel in hot else ""
        cell = f"`` {name} ``" if "`" in name else f"`{name}`"
        print(f"| {cell} | `{rel}:{line}` | {mark} |  |")


def check(path: str, classes: dict[str, list[tuple[str, int, str]]]) -> int:
    rows = documented(path)
    missing = 0
    for cls, items in classes.items():
        for rel, line, name in items:
            test = rows.get(name.strip())
            if test is None:
                print(f"{cls}: `{name}` ({rel}:{line}) sans ligne dans {path}")
                missing += 1
            elif test == "":
                print(f"{cls}: `{name}` ({rel}:{line}) a une colonne Test vide")
                missing += 1
    total = sum(len(v) for v in classes.values())
    print(f"inventory --check : {total - missing}/{total} éléments documentés")
    return 1 if missing else 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--repo", required=True)
    ap.add_argument("--base", default="origin/main")
    ap.add_argument("--class", dest="classes", action="append", default=[],
                    choices=["tools", "reads", "writes", "ids", "keys", "promises", "exports", "rpcs", "all"])
    ap.add_argument("--pattern", action="append", default=[], help="motif supplémentaire pour la classe reads")
    ap.add_argument("--check", help="failing_first.md à comparer aux classes demandées")
    args = ap.parse_args()
    repo = os.path.abspath(args.repo)
    wanted = set(args.classes) or {"all"}
    if "all" in wanted:
        wanted = {"tools", "reads", "writes", "ids", "keys", "promises", "exports", "rpcs"}

    files = go_files(repo)
    hot = touched(repo, args.base)
    if args.check:
        builders = {
            "tools": lambda: tools_of(repo, files),
            "reads": lambda: by_function(repo, files, READ_PATTERNS + args.pattern),
            "writes": lambda: by_function(repo, files, WRITE_PATTERNS),
            "ids": lambda: ids_of(repo, files),
            "keys": lambda: keys_of(repo, files),
            "promises": lambda: promises_of(repo, args.base),
            "exports": lambda: exports_of(repo, files),
            "rpcs": lambda: rpcs_of(repo),
        }
        return check(args.check, {c: builders[c]() for c in sorted(wanted)})
    head = sh(["git", "rev-parse", "--short", "HEAD"], repo).strip()
    print(f"## Inventaire de `{head}` contre `{args.base}` (`*` = fichier touché par le diff)")
    if "tools" in wanted:
        print_table("tools : chaque outil déclaré", tools_of(repo, files), hot)
    if "reads" in wanted:
        print_table("reads : chaque fonction qui lit les documents ou les chunks",
                    by_function(repo, files, READ_PATTERNS + args.pattern), hot)
    if "writes" in wanted:
        print_table("writes : chaque fonction qui écrit une ligne", by_function(repo, files, WRITE_PATTERNS), hot)
    if "ids" in wanted:
        print_table("ids : chaque propriété d'outil qui porte un identifiant", ids_of(repo, files), hot)
    if "keys" in wanted:
        print_table("keys : chaque clé réservée", keys_of(repo, files), hot)
    if "promises" in wanted:
        print_table("promises : chaque description ajoutée par le diff", promises_of(repo, args.base), hot)
    if "exports" in wanted:
        print_table("exports : chaque entrée exportée de pkg/cortexdb qui lit", exports_of(repo, files), hot)
    if "rpcs" in wanted:
        print_table("rpcs : chaque RPC typé", rpcs_of(repo), hot)
    return 0


if __name__ == "__main__":
    sys.exit(main())
