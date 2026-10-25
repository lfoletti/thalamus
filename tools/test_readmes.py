#!/usr/bin/env python3 -I
"""Les traductions du README suivent le README anglais : python3 -I tools/test_readmes.py"""
import re
import unittest
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
BASE = "https://github.com/lfoletti/thalamus/blob/main/"
# La phrase du bas de chaque traduction : l'anglais fait foi, les issues en anglais.
TRANSLATIONS = {
    "README.zh-CN.md": "两者不一致时，以英文版为准。请用英文提交 issue。",
    "README.ja.md": "内容が異なる場合は英語版が優先されます。issue は英語でお書きください。",
    "README.ko.md": "내용이 다를 경우 영어판이 우선합니다. issue는 영어로 작성해 주세요.",
    "README.de.md": "Bei Abweichungen gilt die englische Fassung. Bitte schreiben Sie Issues auf Englisch.",
}
READMES = ["README.md", *TRANSLATIONS]
# Les mots de CortexDB (README_CN.md, docs/GUIDE_CN.md) : « agent » et « key »
# restent en caractères latins.
NOT_CORTEXDB = {"README.zh-CN.md": {"智能体": "agent", "密钥": "key"}}
FENCE = re.compile(r"^```.*?^```", re.M | re.S)


def text(name):
    return (ROOT / name).read_text(encoding="utf-8")


def english():
    # Le résumé en français reste dans le README anglais seul.
    return text("README.md").split("## Résumé en français")[0]


def fences(s):
    return FENCE.findall(s)


def inline(s):
    return Counter(re.findall(r"`([^`\n]+)`", FENCE.sub("", s)))


def targets(s):
    # Les liens et les images, hors la ligne des langues.
    found = re.findall(r"\]\(([^)\s]+)\)", s) + re.findall(r"src=\"([^\"]+)\"", s)
    return set(found) - {BASE + r for r in READMES}


def versions(s):
    return Counter(re.findall(r"(?<![\d.])\d+\.\d+\.\d+(?!\.?\d)", s))


class Readmes(unittest.TestCase):
    def test_each_readme_links_to_the_others(self):
        for name in READMES:
            for other in READMES:
                if other != name:
                    self.assertIn(f"]({BASE}{other})", text(name), f"{name} -> {other}")

    def test_the_code_blocks_are_the_english_ones(self):
        for name in TRANSLATIONS:
            self.assertEqual(fences(text(name)), fences(english()), name)

    def test_the_inline_code_is_the_english_one(self):
        for name in TRANSLATIONS:
            self.assertEqual(inline(text(name)), inline(english()), name)

    def test_the_links_and_images_are_the_english_ones(self):
        for name in TRANSLATIONS:
            self.assertEqual(targets(text(name)), targets(english()), name)

    def test_the_versions_are_the_english_ones(self):
        for name in TRANSLATIONS:
            self.assertEqual(versions(text(name)), versions(english()), name)

    def test_each_section_link_has_its_anchor(self):
        # Une ancre de GitHub vient du titre, qui change avec la langue.
        for name in TRANSLATIONS:
            s = text(name)
            for anchor in re.findall(r"\]\(#([^)]+)\)", s):
                self.assertTrue(f'<a id="{anchor}"></a>' in s, f"{name}: #{anchor}")

    def test_the_words_are_those_of_cortexdb(self):
        for name, words in NOT_CORTEXDB.items():
            for word, theirs in words.items():
                self.assertNotIn(word, text(name), f"{name}: CortexDB writes {theirs}")

    def test_each_translation_says_that_english_prevails(self):
        for name, sentence in TRANSLATIONS.items():
            self.assertIn(sentence, text(name).rstrip().splitlines()[-1], name)


if __name__ == "__main__":
    unittest.main()
