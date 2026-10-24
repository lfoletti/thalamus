#!/usr/bin/env python3 -I
"""Tests du cœur pur de topic_watch.py, sans GitHub : python3 -I tools/test_topic_watch.py"""
import importlib.util
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location("topic_watch", Path(__file__).resolve().parent / "topic_watch.py")
tw = importlib.util.module_from_spec(spec)
spec.loader.exec_module(tw)


class FakeGitHub:
    """Remplace tw.gh : rend une réponse par chemin de recherche, et garde les chemins lus."""

    def __init__(self, answers):
        self.answers, self.paths = answers, []

    def __call__(self, path):
        self.paths.append(path)
        return self.answers[path]


def m(total, new):
    return {"total": total, "new": new, "momentum": round(new / total, 3) if total else 0.0}


class Count(unittest.TestCase):
    def setUp(self):
        self.real = tw.gh
        self.addCleanup(setattr, tw, "gh", self.real)

    def test_total_new_and_momentum(self):
        tw.gh = FakeGitHub({
            "search/repositories?q=topic:mcp&per_page=1": {"total_count": 400},
            "search/repositories?q=topic:mcp+created:>2026-09-24&per_page=1": {"total_count": 30},
        })
        self.assertEqual(tw.count("mcp", "2026-09-24"), {"total": 400, "new": 30, "momentum": 0.075})

    def test_a_topic_without_repository_has_momentum_zero(self):
        tw.gh = lambda path: {"total_count": 0}
        self.assertEqual(tw.count("nothing", "2026-09-24"), {"total": 0, "new": 0, "momentum": 0.0})


class Candidates(unittest.TestCase):
    def setUp(self):
        self.real = tw.gh
        self.addCleanup(setattr, tw, "gh", self.real)

    def page(self, *repos):
        return {"items": [{"topics": list(t)} for t in repos]}

    def fake(self, pages):
        return FakeGitHub({
            f"search/repositories?q=topic:{seed}+created:>2026-09-24&sort=stars&order=desc&per_page=50": page
            for seed, page in pages.items()
        })

    def test_a_candidate_occurs_on_enough_repositories_and_is_not_own(self):
        tw.gh = self.fake({
            "a": self.page(["x", "y", "own"], ["x", "y", "own"], ["x", "own"]),
            "b": self.page(["y", "z"], ["z"]),
        })
        # x: 3, y: 3, own: 3 (exclu), z: 2 (sous MIN_HITS).
        self.assertEqual(sorted(tw.candidates({"own"}, "2026-09-24", ["a", "b"])), ["x", "y"])

    def test_a_topic_counts_once_per_repository(self):
        tw.gh = self.fake({"a": self.page(["x", "x", "x"], ["x"])})
        self.assertEqual(tw.candidates(set(), "2026-09-24", ["a"]), [])

    def test_the_most_frequent_come_first_and_the_list_is_capped(self):
        topics = [f"t{i:02d}" for i in range(tw.CANDIDATES + 5)]
        repos = []
        for i, t in enumerate(topics):
            repos += [[t]] * (tw.MIN_HITS + len(topics) - i)
        tw.gh = self.fake({"a": self.page(*repos)})
        self.assertEqual(tw.candidates(set(), "2026-09-24", ["a"]), topics[:tw.CANDIDATES])


class Report(unittest.TestCase):
    OWN = {"rag": m(1000, 50), "mcp": m(400, 40), "gdpr": m(200, 2)}

    def rising(self, text):
        part = text.split("## Rising candidates")[1].split("## Repository topics")[0]
        return [line.split("`")[1] for line in part.splitlines() if line.startswith("| `")]

    def test_a_rising_candidate_passes_the_floor_and_the_median(self):
        # floor = 2 (gdpr), median momentum = 0.05 (rag).
        cand = {"up": m(100, 10), "low-new": m(10, 1), "slow": m(1000, 20)}
        text = tw.report("o/r", "2026-10-24", self.OWN, cand, {})
        self.assertEqual(self.rising(text), ["up"])
        self.assertIn("`low-new`", text.split("## Other candidates")[1])
        self.assertIn("`slow`", text.split("## Other candidates")[1])

    def test_no_candidate_passes(self):
        text = tw.report("o/r", "2026-10-24", self.OWN, {"slow": m(1000, 20)}, {})
        self.assertIn("No candidate passes the two limits.", text)

    def test_the_weakest_topics_come_first_to_replace(self):
        text = tw.report("o/r", "2026-10-24", self.OWN, {"up": m(100, 10)}, {})
        self.assertIn("first to replace: `gdpr`, `rag`, `mcp`.", text)

    def test_the_change_of_new_compares_with_the_previous_snapshot(self):
        text = tw.report("o/r", "2026-10-24", self.OWN, {}, {"mcp": m(400, 35)})
        self.assertIn("| `mcp` | 400 | 40 | 0.100 | +5 |", text)
        self.assertIn("| `rag` | 1000 | 50 | 0.050 | n/a |", text)

    def test_repository_topics_are_sorted_by_momentum(self):
        part = tw.report("o/r", "2026-10-24", self.OWN, {}, {}).split("## Repository topics")[1].split("## Other")[0]
        self.assertEqual([l.split("`")[1] for l in part.splitlines() if l.startswith("| `")], ["mcp", "rag", "gdpr"])

    def test_the_header_gives_the_topic_count_and_the_limit(self):
        self.assertIn("The repository has 3 topics of 20.", tw.report("o/r", "2026-10-24", self.OWN, {}, {}))


class ParseArgs(unittest.TestCase):
    HOME = Path("/home/u")

    def test_defaults(self):
        self.assertEqual(tw.parse_args([], self.HOME),
                         ("lfoletti/thalamus", self.HOME / ".thalamus" / "topic_watch", tw.SEEDS))

    def test_repo_out_and_seeds_in_any_order(self):
        want = ("o/r", Path("/tmp/w"), ["a", "b"])
        self.assertEqual(tw.parse_args(["o/r", "--out", "/tmp/w", "--seeds", "a,b"], self.HOME), want)
        self.assertEqual(tw.parse_args(["--seeds", "a,b", "--out", "/tmp/w", "o/r"], self.HOME), want)

    def test_empty_seeds_are_dropped(self):
        self.assertEqual(tw.parse_args(["--seeds", "a,,b,"], self.HOME)[2], ["a", "b"])


if __name__ == "__main__":
    unittest.main()
