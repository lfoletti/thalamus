#!/usr/bin/env python3 -I
"""Les nombres du corps d'une PR, mesurés et comparés (règle 5 de
docs/pipeline_pr.md, section 3 bis, 22 octobre 2026).

Mesure, sur la tête de --repo contre --base :
  - les lignes ajoutées et retirées hors tests, avec et sans la documentation
    (*.md), et le nombre de fichiers ;
  - le nombre de mutations à la main des jeux de --mutations dont le nom
    commence par --sets (lecture du dictionnaire SETS, sans exécuter le script) ;
  - avec --pr, les lignes non couvertes du rapport Codecov de la tête.

Puis lit chaque --body et compare les phrases connues aux mesures :
  « <n> mutations for this/the amendment, <n> killed »,
  « <n> lines outside tests[ and docs][ over <n> files, <n> removed] »,
  « <n> of the added lines are in »,
  « The <n> lines left uncovered » (et le repère CODECOV_<pr>).
Avec --fill, remplace CODECOV_<pr> par le nombre en lettres.

Sortie 1 si une phrase ne correspond pas à la mesure, ou si un repère reste.

Usage :
  python3 -I tools/bodynums.py --repo <copie> --base <sha> --sets 77 \\
      --body ../thalamus_local/docs/pr/round2/GC-06.body.md --body ../thalamus_local/docs/pr/round2/review_replies.md#77 \\
      [--pr 77 [--fill]]
Un --body de la forme <fichier>#<pr> lit seulement la section « ## #<pr> ».
"""
import argparse
import ast
import json
import re
import subprocess
import sys
from pathlib import Path

UNITS = ["zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
         "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen",
         "seventeen", "eighteen", "nineteen"]
TENS = ["", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"]


def words(n):
    """Le nombre en lettres anglaises, de 0 à 999 (« forty-one »)."""
    if not 0 <= n <= 999:
        raise ValueError(f"{n} hors de 0 à 999")
    if n < 20:
        return UNITS[n]
    if n < 100:
        return TENS[n // 10] + ("" if n % 10 == 0 else "-" + UNITS[n % 10])
    rest = n % 100
    return UNITS[n // 100] + " hundred" + ("" if rest == 0 else " and " + words(rest))


def number(text):
    """Un nombre écrit en chiffres ou en lettres, ou None."""
    text = text.strip().lower()
    if text.isdigit():
        return int(text)
    for n in range(1000):
        if words(n) == text:
            return n
    return None


def is_test(path):
    return path.endswith("_test.go") or "/testdata/" in path or path.startswith("testdata/")


def line_counts(numstat):
    """Compte la sortie de `git diff --numstat` : lignes ajoutées, retirées et
    fichiers hors tests, avec et sans les fichiers *.md."""
    out = {"added": 0, "removed": 0, "files": 0, "docs_added": 0, "docs_removed": 0, "docs_files": 0}
    for row in numstat.splitlines():
        parts = row.split("\t")
        if len(parts) != 3 or parts[0] == "-":
            continue
        added, removed, path = int(parts[0]), int(parts[1]), parts[2]
        if is_test(path):
            continue
        out["added"] += added
        out["removed"] += removed
        out["files"] += 1
        if path.endswith(".md"):
            out["docs_added"] += added
            out["docs_removed"] += removed
            out["docs_files"] += 1
    return out


def mutation_counts(source, prefix):
    """Le nombre de mutations de chaque jeu de SETS dont le nom commence par
    prefix, lu dans le texte du script sans l'exécuter."""
    for node in ast.walk(ast.parse(source)):
        if isinstance(node, ast.Assign) and any(isinstance(t, ast.Name) and t.id == "SETS" for t in node.targets):
            sets = {}
            for key, value in zip(node.value.keys, node.value.values):
                name = ast.literal_eval(key)
                if name.startswith(prefix):
                    sets[name] = len(value.elts[2].elts)
            return sets
    raise ValueError("pas de SETS dans le script des mutations")


def codecov_lines(comment):
    """Les lignes non couvertes d'un commentaire Codecov, ou None."""
    m = re.search(r"with `(\d+) lines?` in your changes missing coverage", comment)
    if m:
        return int(m.group(1))
    if re.search(r"All modified and coverable lines are covered", comment):
        return 0
    return None


N = r"([A-Za-z]+(?:-[A-Za-z]+)?|\d+)"


def check(body, nums, pr=None):
    """Compare les phrases connues du corps aux mesures. Rend la liste des
    écarts, chacun en une ligne."""
    problems = []
    lines, total = nums["lines"], nums.get("mutations")
    for m in re.finditer(N + r" mutations for (?:this|the) amendment, " + N + r" killed", body):
        said, killed = number(m.group(1)), number(m.group(2))
        if total is not None and (said != total or killed != total):
            problems.append(f"« {m.group(0)} » : {total} mutations mesurées")
    for m in re.finditer(r"(\d+) lines outside tests( and docs)?(?: over (\d+) files, (\d+) removed)?", body):
        docs = bool(m.group(2))
        added = lines["added"] - (lines["docs_added"] if docs else 0)
        files = lines["files"] - (lines["docs_files"] if docs else 0)
        removed = lines["removed"] - (lines["docs_removed"] if docs else 0)
        want = [(int(m.group(1)), added, "lignes")]
        if m.group(3):
            want += [(int(m.group(3)), files, "fichiers"), (int(m.group(4)), removed, "retirées")]
        for said, real, what in want:
            if said != real:
                problems.append(f"« {m.group(0)} » : {real} {what} mesurées")
    for m in re.finditer(r"(\d+) of the added lines are in", body):
        if int(m.group(1)) != lines["docs_added"]:
            problems.append(f"« {m.group(0)} » : {lines['docs_added']} lignes de documentation mesurées")
    uncovered = nums.get("codecov")
    for m in re.finditer(r"The " + N + r" lines? left uncovered", body):
        if uncovered is not None and number(m.group(1)) != uncovered:
            problems.append(f"« {m.group(0)} » : Codecov en compte {uncovered}")
    for m in re.finditer(r"CODECOV_\d+", body):
        problems.append(f"repère {m.group(0)} non rempli")
    return problems


def section(text, pr):
    """La section « ## #<pr> » d'un fichier de réponses, jusqu'au titre suivant."""
    m = re.search(rf"^## #{pr}\b.*?$(.*?)(?=^## |\Z)", text, re.S | re.M)
    if not m:
        raise ValueError(f"pas de section ## #{pr}")
    return m.group(1)


def fill(body, pr, uncovered):
    """Remplace le repère CODECOV_<pr> par le nombre en lettres."""
    return body.replace(f"CODECOV_{pr}", words(uncovered))


def git(repo, *args):
    return subprocess.run(["git", "-C", repo, *args], check=True, capture_output=True, text=True).stdout


def gh(*args):
    return json.loads(subprocess.run(["gh", *args], check=True, capture_output=True, text=True).stdout)


def fetch_codecov(repo_slug, pr, head):
    """Les lignes non couvertes du rapport Codecov de la tête, ou une erreur
    si le rapport n'est pas encore celui de cette tête."""
    runs = gh("api", f"repos/{repo_slug}/commits/{head}/check-runs")["check_runs"]
    patch = [r for r in runs if r["name"] == "codecov/patch" and r["status"] == "completed"]
    if not patch:
        raise RuntimeError(f"codecov/patch pas encore terminé sur {head[:7]}")
    comments = gh("api", f"repos/{repo_slug}/issues/{pr}/comments", "--paginate")
    bot = [c for c in comments if "codecov" in c["user"]["login"]]
    if not bot or bot[-1]["updated_at"] < patch[0]["started_at"]:
        raise RuntimeError(f"le commentaire Codecov de #{pr} précède le rapport de {head[:7]}")
    uncovered = codecov_lines(bot[-1]["body"])
    if uncovered is None:
        raise RuntimeError(f"commentaire Codecov de #{pr} illisible")
    return uncovered


# Les notes de lot vivent dans l'archive locale, à côté du dépôt.
DEFAULT_MUTATIONS = str(Path(__file__).resolve().parent.parent.parent / "thalamus_local/docs/pr/round2/mutate_round2.py")


def main(argv):
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--repo", required=True)
    ap.add_argument("--base", required=True)
    ap.add_argument("--sets", help="préfixe des jeux de mutations de la PR")
    ap.add_argument("--mutations", default=DEFAULT_MUTATIONS)
    ap.add_argument("--body", action="append", default=[])
    ap.add_argument("--pr", type=int)
    ap.add_argument("--slug", default="liliang-cn/cortexdb")
    ap.add_argument("--fill", action="store_true")
    a = ap.parse_args(argv)

    nums = {"lines": line_counts(git(a.repo, "diff", "--numstat", f"{a.base}..HEAD"))}
    l = nums["lines"]
    print(f"lignes hors tests : {l['added']} ajoutées, {l['removed']} retirées, {l['files']} fichiers ; "
          f"dont *.md : {l['docs_added']} ajoutées, {l['docs_removed']} retirées, {l['docs_files']} fichiers")
    if a.sets:
        sets = mutation_counts(Path(a.mutations).read_text(), a.sets)
        nums["mutations"] = sum(sets.values())
        print(f"mutations : {nums['mutations']} ({', '.join(f'{k} {v}' for k, v in sorted(sets.items()))})")
    if a.pr:
        head = git(a.repo, "rev-parse", "HEAD").strip()
        nums["codecov"] = fetch_codecov(a.slug, a.pr, head)
        print(f"Codecov #{a.pr} sur {head[:7]} : {nums['codecov']} lignes non couvertes")

    bad = 0
    for path in a.body:
        path, _, part = path.partition("#")
        text = Path(path).read_text()
        if part:
            problems = check(section(text, int(part)), nums, a.pr)
            for p in problems:
                print(f"{path}#{part} : {p}")
            bad += len(problems)
            continue
        if a.fill and a.pr:
            filled = fill(text, a.pr, nums["codecov"])
            if filled != text:
                Path(path).write_text(filled)
                print(f"{path} : CODECOV_{a.pr} remplacé par « {words(nums['codecov'])} »")
                text = filled
        problems = check(text, nums, a.pr)
        for p in problems:
            print(f"{path} : {p}")
        bad += len(problems)
    print(f"bodynums : {bad} écart(s)")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
