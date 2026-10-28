#!/usr/bin/env python3 -I
"""Où les bancs et le Makefile trouvent le clone CortexDB : python3 -I tools/test_cortexdb_clone.py

CORTEXDB_CLONE désigne le clone ; sans elle, c'est cortexdb/ à la racine du dépôt, comme avant. Un clone absent
arrête un banc avant toute construction, avec un message qui nomme le chemin lu et la variable."""
import os
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
BENCHES = ["tests/e2e_gateway.sh", "tests/e2e_shared_brain.sh", "tests/e2e_plugin.sh", "tests/e2e_extension.sh",
           "tests/eval_ch/run.sh"]


def run_bench(bench: str, clone: str | None) -> tuple[subprocess.CompletedProcess, Path]:
    """Le banc copié dans une racine jetable, sans clone : il doit s'arrêter avant de construire quoi que ce soit."""
    tmp = Path(tempfile.mkdtemp()).resolve()          # macOS : /var est un lien vers /private/var
    for f in [bench, "tests/cortexdb_clone.sh"]:
        (tmp / f).parent.mkdir(parents=True, exist_ok=True)
        if (ROOT / f).exists():
            shutil.copy(ROOT / f, tmp / f)
    env = {k: v for k, v in os.environ.items() if k != "CORTEXDB_CLONE"}
    if clone is not None:
        env["CORTEXDB_CLONE"] = clone
    env["PATH"] = "/usr/bin:/bin"          # ni go ni npm : un banc qui irait plus loin tomberait autrement
    r = subprocess.run(["bash", str(tmp / bench), "lexical"], cwd=tmp, env=env, capture_output=True, text=True,
                       timeout=30)
    return r, tmp


class Benches(unittest.TestCase):
    def test_a_bench_reads_the_clone_from_cortexdb_clone(self):
        for bench in BENCHES:
            with self.subTest(bench=bench):
                r, tmp = run_bench(bench, "/nonexistent/cortexdb-clone")
                self.assertEqual(r.returncode, 1, r.stdout + r.stderr)
                self.assertIn("/nonexistent/cortexdb-clone", r.stderr)
                self.assertIn("CORTEXDB_CLONE", r.stderr)
                shutil.rmtree(tmp)

    def test_without_cortexdb_clone_a_bench_reads_cortexdb_at_the_root(self):
        for bench in BENCHES:
            with self.subTest(bench=bench):
                r, tmp = run_bench(bench, None)
                self.assertEqual(r.returncode, 1, r.stdout + r.stderr)
                self.assertIn(f"no CortexDB clone at {tmp}/cortexdb", r.stderr)
                shutil.rmtree(tmp)


class Makefile(unittest.TestCase):
    def repo(self, *args: str, env_clone: str | None = None) -> str:
        env = {k: v for k, v in os.environ.items() if k != "CORTEXDB_CLONE"}
        if env_clone is not None:
            env["CORTEXDB_CLONE"] = env_clone
        r = subprocess.run(["make", "-s", "--no-print-directory", "print-repo", *args], cwd=ROOT, env=env,
                           capture_output=True, text=True, timeout=30)
        self.assertEqual(r.returncode, 0, r.stderr)
        return r.stdout.strip()

    def test_repo_defaults_to_cortexdb_clone(self):
        self.assertEqual(self.repo(env_clone="/elsewhere/cortexdb"), "/elsewhere/cortexdb")

    def test_repo_defaults_to_cortexdb_at_the_root_without_it(self):
        self.assertEqual(self.repo(), "cortexdb")

    def test_an_explicit_repo_wins(self):
        self.assertEqual(self.repo("REPO=../wt/co-12", env_clone="/elsewhere/cortexdb"), "../wt/co-12")


if __name__ == "__main__":
    unittest.main()
