"""Kit de la relecture adverse (docs/pipeline_pr.md, section 1 bis).

Usage : python3 -I tools/advkit.py --repo <copie git> --head <sha> --base <sha> --out <dossier>

Exporte la tête et la base dans <out>/head et <out>/main (git archive, sans
.git), écrit <out>/pr.diff et <out>/stat.txt, et deux scripts :
  <out>/both.sh '<arguments de go test>'  : même commande sur main puis sur la
      tête, sous le canari, sortie réduite aux lignes ok/FAIL/panic et aux
      messages des tests qui tombent ;
  <out>/mutone.py <fichier> <ancien> <nouveau> <paquet> <motif -run> : une
      mutation sur une copie de la tête (<out>/mut), le test, puis la
      restauration depuis <out>/head ; rend killed, survived ou build-failed.
"""
import argparse, os, shutil, subprocess, sys

CANARY = ("CORTEXDB_REMOTE=127.0.0.1:1 CORTEXDB_GRPC_TOKEN=canary CORTEXDB_PATH=/dev/null/canary "
          "CORTEXDB_EMBED_BASE_URL=http://127.0.0.1:1 OPENAI_BASE_URL=http://127.0.0.1:1 OPENAI_API_KEY= "
          "CORTEXDB_LLM_BASE_URL=http://127.0.0.1:1 CORTEXDB_LLM_API_KEY=canary CORTEXDB_MEMORY_SEMANTIC_FLOOR=canary")

BOTH = """#!/bin/sh
# Même commande go test sur main puis sur la tête, sous le canari, sortie filtrée.
K=$(cd "$(dirname "$0")" && pwd)
for t in main head; do
  echo "== $t"
  (cd "$K/$t" && env @CANARY@ go test -count=1 "$@" 2>&1) | grep -E '^(ok|FAIL|--- FAIL|panic)|_test\\.go:[0-9]+:' | cut -c1-300
done
""".replace("@CANARY@", CANARY)

MUTONE = '''"""Une mutation sur une copie de la tête : killed, survived ou build-failed."""
import os, shutil, subprocess, sys
k = os.path.dirname(os.path.abspath(__file__))
f, old, new, pkg, run = sys.argv[1:6]
head, mut = os.path.join(k, "head"), os.path.join(k, "mut")
if not os.path.isdir(mut):
    shutil.copytree(head, mut)
src = os.path.join(mut, f)
text = open(src).read()
if text.count(old) != 1:
    sys.exit("anchor found %d times" % text.count(old))
open(src, "w").write(text.replace(old, new))
env = dict(os.environ)
for kv in """@CANARY@""".split():
    a, _, b = kv.partition("=")
    env[a] = b
try:
    r = subprocess.run(["go", "test", "-count=1", "-run", run, pkg], cwd=mut, env=env, capture_output=True, text=True)
finally:
    shutil.copy2(os.path.join(head, f), src)
out = r.stdout + r.stderr
if "[build failed]" in out or "[setup failed]" in out:
    print("build-failed"); print(out[-1500:])
elif r.returncode != 0:
    print("killed:", " ".join(l.split()[2] for l in out.splitlines() if l.startswith("--- FAIL")))
else:
    print("survived")
'''.replace("@CANARY@", CANARY)


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--repo", required=True)
    p.add_argument("--head", required=True)
    p.add_argument("--base", required=True)
    p.add_argument("--out", required=True)
    a = p.parse_args()
    if os.path.exists(a.out):
        shutil.rmtree(a.out)
    for name, rev in (("head", a.head), ("main", a.base)):
        d = os.path.join(a.out, name)
        os.makedirs(d)
        arc = subprocess.run(["git", "-C", a.repo, "archive", rev], check=True, capture_output=True).stdout
        subprocess.run(["tar", "-x", "-C", d], input=arc, check=True)
    diff = subprocess.run(["git", "-C", a.repo, "diff", a.base, a.head], check=True, capture_output=True, text=True).stdout
    open(os.path.join(a.out, "pr.diff"), "w").write(diff)
    stat = subprocess.run(["git", "-C", a.repo, "diff", "--stat", a.base, a.head], check=True, capture_output=True, text=True).stdout
    open(os.path.join(a.out, "stat.txt"), "w").write(stat)
    both = os.path.join(a.out, "both.sh")
    open(both, "w").write(BOTH)
    os.chmod(both, 0o755)
    open(os.path.join(a.out, "mutone.py"), "w").write(MUTONE)
    print("kit:", a.out)
    print(stat.strip().splitlines()[-1] if stat.strip() else "diff vide")


if __name__ == "__main__":
    main()
