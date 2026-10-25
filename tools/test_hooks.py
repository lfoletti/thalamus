#!/usr/bin/env python3
"""Tests of tools/hooks: python3 -I tools/test_hooks.py

Each test builds a throwaway repository whose tools/hooks is a copy of ours,
installed the way CONTRIBUTING.md says (git config core.hooksPath tools/hooks,
a relative path), then tries commits where the rule says yes or no.
"""
import os
import pathlib
import shutil
import subprocess
import tempfile
import unittest

HOOKS = pathlib.Path(__file__).resolve().parent / "hooks"
AUTHOR = ["-c", "user.name=Test", "-c", "user.email=test@example.com"]


def clean_env():
    """A hook runs with GIT_DIR and friends exported; a git command would then act on the real repository,
    whatever its -C. The throwaway repository gets an environment without any GIT_* variable."""
    return {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}


def git(repo, *args, check=True):
    return subprocess.run(["git", *AUTHOR, "-C", str(repo), *args], capture_output=True, text=True, check=check,
                          env=clean_env())


def repo_with_hooks(root):
    repo = pathlib.Path(root) / "repo"
    repo.mkdir()
    git(repo, "init", "-q", "-b", "main")
    shutil.copytree(HOOKS, repo / "tools" / "hooks")
    git(repo, "config", "core.hooksPath", "tools/hooks")
    (repo / "README.md").write_text("first\n")
    # The first commit is made without the hook: a repository has to be born on main.
    git(repo, "-c", "core.hooksPath=/dev/null", "add", "-A")
    git(repo, "-c", "core.hooksPath=/dev/null", "commit", "-q", "-m", "first")
    return repo


def commit(repo, name):
    (repo / name).write_text(name + "\n")
    git(repo, "add", name)
    return git(repo, "commit", "-q", "-m", name, check=False)


class ThrowawayRepository(unittest.TestCase):
    def test_the_git_variables_of_a_hook_do_not_reach_the_throwaway_repository(self):
        with tempfile.TemporaryDirectory() as root:
            saved = dict(os.environ)
            os.environ["GIT_DIR"] = str(pathlib.Path(root) / "elsewhere.git")
            os.environ["GIT_WORK_TREE"] = root
            try:
                repo = repo_with_hooks(root)
                top = git(repo, "rev-parse", "--show-toplevel").stdout.strip()
            finally:
                os.environ.clear()
                os.environ.update(saved)
            self.assertEqual(pathlib.Path(top).resolve(), repo.resolve())
            self.assertFalse((pathlib.Path(root) / "elsewhere.git").exists())


class PreCommit(unittest.TestCase):
    def test_a_commit_on_main_is_refused(self):
        with tempfile.TemporaryDirectory() as root:
            repo = repo_with_hooks(root)
            r = commit(repo, "on_main.txt")
            self.assertNotEqual(r.returncode, 0)
            self.assertIn("no commit on main", r.stderr)
            self.assertIn("git worktree add", r.stderr)
            self.assertEqual(git(repo, "rev-list", "--count", "HEAD").stdout.strip(), "1")

    def test_a_commit_on_a_branch_passes(self):
        with tempfile.TemporaryDirectory() as root:
            repo = repo_with_hooks(root)
            git(repo, "checkout", "-q", "-b", "topic")
            r = commit(repo, "on_topic.txt")
            self.assertEqual(r.returncode, 0, r.stderr)
            self.assertEqual(git(repo, "rev-list", "--count", "HEAD").stdout.strip(), "2")

    def test_the_relative_hooks_path_runs_in_a_worktree(self):
        with tempfile.TemporaryDirectory() as root:
            repo = repo_with_hooks(root)
            git(repo, "checkout", "-q", "-b", "topic")
            wt = pathlib.Path(root) / "wt_main"
            git(repo, "worktree", "add", "-q", str(wt), "main")
            r = commit(wt, "in_worktree.txt")
            self.assertNotEqual(r.returncode, 0)
            self.assertIn("no commit on main", r.stderr)

    def test_a_detached_head_passes(self):
        with tempfile.TemporaryDirectory() as root:
            repo = repo_with_hooks(root)
            git(repo, "checkout", "-q", "--detach")
            r = commit(repo, "detached.txt")
            self.assertEqual(r.returncode, 0, r.stderr)


if __name__ == "__main__":
    unittest.main()
