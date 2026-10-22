#!/usr/bin/env python3 -I
"""Tests du cœur pur de bodynums.py : python3 -I tools/test_bodynums.py"""
import importlib.util
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location("bodynums", Path(__file__).resolve().parent / "bodynums.py")
bn = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bn)

NUMSTAT = "\n".join([
    "100\t10\tpkg/core/a.go",
    "50\t5\tpkg/core/a_test.go",
    "7\t1\tpkg/core/testdata/x.json",
    "20\t2\tdocs/GUIDE.md",
    "-\t-\tlogo.png",
])
SCRIPT = '''
SETS = {
    "77": ("./pkg/", "X", [("a", "f", "x", "y", ["T"]), ("b", "f", "x", "y", ["T"])]),
    "77r3": ("./pkg/", "X", [("c", "f", "x", "y", ["T"])]),
    "76": ("./pkg/", "X", [("d", "f", "x", "y", ["T"])]),
}
'''


class Words(unittest.TestCase):
    def test_words(self):
        self.assertEqual([bn.words(n) for n in (0, 7, 15, 20, 41, 57, 100, 236)],
                         ["zero", "seven", "fifteen", "twenty", "forty-one", "fifty-seven",
                          "one hundred", "two hundred and thirty-six"])
        with self.assertRaises(ValueError):
            bn.words(1000)

    def test_number_reads_digits_and_words(self):
        self.assertEqual(bn.number("Forty-eight"), 48)
        self.assertEqual(bn.number("689"), 689)
        self.assertIsNone(bn.number("many"))


class Counts(unittest.TestCase):
    def test_line_counts_leave_tests_and_binaries_out(self):
        self.assertEqual(bn.line_counts(NUMSTAT), {
            "added": 120, "removed": 12, "files": 2,
            "docs_added": 20, "docs_removed": 2, "docs_files": 1})

    def test_mutation_counts_by_prefix(self):
        self.assertEqual(bn.mutation_counts(SCRIPT, "77"), {"77": 2, "77r3": 1})
        with self.assertRaises(ValueError):
            bn.mutation_counts("X = 1", "77")

    def test_codecov_lines(self):
        self.assertEqual(bn.codecov_lines(":x: Patch coverage is `97.5%` with `23 lines` in your changes missing coverage."), 23)
        self.assertEqual(bn.codecov_lines("with `1 line` in your changes missing coverage"), 1)
        self.assertEqual(bn.codecov_lines("All modified and coverable lines are covered by tests :white_check_mark:"), 0)
        self.assertIsNone(bn.codecov_lines("no report"))


class Check(unittest.TestCase):
    nums = {"lines": bn.line_counts(NUMSTAT), "mutations": 3, "codecov": 7}

    def test_a_body_that_says_the_measures_passes(self):
        body = ("Three mutations for this amendment, three killed. 120 lines outside tests over 2 files, "
                "12 removed; 20 of the added lines are in docs. 100 lines outside tests and docs. "
                "The seven lines left uncovered are inevitable.")
        self.assertEqual(bn.check(body, self.nums), [])

    def test_each_wrong_number_is_reported(self):
        body = ("Four mutations for the amendment, three killed. 119 lines outside tests over 3 files, 12 removed; "
                "21 of the added lines are in docs. 120 lines outside tests and docs. "
                "The one line left uncovered. The CODECOV_77 lines left uncovered.")
        problems = bn.check(body, self.nums)
        self.assertEqual(len(problems), 7, problems)

    def test_no_measure_no_judgement(self):
        body = "Fifty mutations for this amendment, fifty killed. The nine lines left uncovered."
        self.assertEqual(bn.check(body, {"lines": self.nums["lines"]}), [])

    def test_section(self):
        text = "## #76 (a)\n\nsix\n\n## #77 (b)\n\nseven\n\n## Discussion\n\nx\n"
        self.assertEqual(bn.section(text, 77).strip(), "seven")
        self.assertEqual(bn.section(text, 76).strip(), "six")
        with self.assertRaises(ValueError):
            bn.section(text, 7)

    def test_fill(self):
        self.assertEqual(bn.fill("The CODECOV_76 lines", 76, 15), "The fifteen lines")
        self.assertEqual(bn.fill("The CODECOV_77 lines", 76, 15), "The CODECOV_77 lines")


if __name__ == "__main__":
    unittest.main()
