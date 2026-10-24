"""Builds pkg/eval/testdata/dataset_ch.json (gc-24) from tests/eval_ch/articles.json and questions.json.

Usage: python3 -I make_dataset_ch.py <thalamus root> <output json>
The 24 question articles with the shortest text, in de, fr, it and rm (96 documents, group = article).
Queries: each question and its two variants, asked against the article in their own language (category
"plain"); the article's heading in each language as a short query ("heading"); each of those with its
accents removed, when that changes it ("unaccented", "heading_unaccented"); six German queries whose word
is only part of a compound in the article ("compound"). Romansh documents are distractors."""
import json, re, sys, unicodedata
from pathlib import Path

root, out = Path(sys.argv[1]), Path(sys.argv[2])
arts = json.loads((root / "tests/eval_ch/articles.json").read_text(encoding="utf-8"))
qs = json.loads((root / "tests/eval_ch/questions.json").read_text(encoding="utf-8"))

def longest(q):
    a = arts[q["article"]]
    return max(len(a[l]["text"]) for l in a)

chosen = sorted(qs, key=lambda q: (longest(q), q["id"]))[:24]
chosen.sort(key=lambda q: q["id"])

def strip(s):
    return unicodedata.normalize("NFC", "".join(c for c in unicodedata.normalize("NFD", s) if unicodedata.category(c) != "Mn"))

docs, queries = [], []
for q in chosen:
    aid = q["article"]
    for lang in ("de", "fr", "it", "rm"):
        a = arts[aid][lang]
        docs.append({"id": f"{aid}__{lang}", "title": a["title"], "content": a["text"], "group": aid})
for q in chosen:
    asked = dict(q["v"]); asked[q["lang"]] = q["q"]
    for lang in ("de", "fr", "it"):
        queries.append({"id": f"{q['id']}_{lang}", "text": asked[lang], "relevant": [f"{q['article']}__{lang}"], "category": "plain"})
for q in chosen:
    for lang in ("de", "fr", "it"):
        heading = re.sub(r"^Art\. [0-9]+( [a-z])?( [0-9]+)? ", "", arts[q["article"]][lang]["title"])
        assert heading and not heading.startswith("Art"), heading
        queries.append({"id": f"{q['id']}_{lang}_h", "text": heading, "relevant": [f"{q['article']}__{lang}"], "category": "heading"})
for q in list(queries):
    bare = strip(q["text"])
    if bare != q["text"]:
        cat = "unaccented" if q["category"] == "plain" else "heading_unaccented"
        queries.append({"id": q["id"] + "_u", "text": bare, "relevant": q["relevant"], "category": cat})

compound = [
    ("c1", "Ist die Versicherung bei Unfall obligatorisch?", "art_117"),   # Unfallversicherung
    ("c2", "Darf der Bund Versuche an Tieren regeln?", "art_80"),          # Tierversuche
    ("c3", "Welche Hilfe bekommt das Opfer einer Tat?", "art_124"),        # Straftat, Opferhilfe
    ("c4", "Ist die Freiheit des Gewissens gewährleistet?", "art_15"),     # Gewissensfreiheit
    ("c5", "Ist der Unterricht im Sport an Schulen obligatorisch?", "art_68"),  # Sportunterricht
    ("c6", "Ist die Freiheit der Wirtschaft gewährleistet?", "art_27"),    # Wirtschaftsfreiheit
]
chosen_arts = {q["article"] for q in chosen}
for cid, text, aid in compound:
    assert aid in chosen_arts, aid
    queries.append({"id": cid + "_de", "text": text, "relevant": [f"{aid}__de"], "category": "compound"})

negatives = [
    {"id": "n1_de", "text": "Wie backt man am Sonntag einen Zopf?"},
    {"id": "n2_fr", "text": "Quelle est la meilleure recette de fondue au fromage ?"},
    {"id": "n3_it", "text": "Quando apre la stagione sciistica a Zermatt?"},
    {"id": "n4_fr", "text": "Combien coûte un abonnement de ski à Verbier ?"},
]

ds = {
    "name": "cortexdb-lexical-ch",
    "description": (
        "Swiss Federal Constitution (SR 101, Fedlex, fedlex.admin.ch, version of 3 March 2024): 24 articles "
        "in the official German, French, Italian and Romansh texts, one document per article and language, "
        "group = article. Queries in German, French and Italian, each against the article in its own language; "
        "Romansh documents are distractors. Categories: plain (the question as written), unaccented (the same "
        "question with its accents removed, as people type), heading (the article's heading as a short query), "
        "heading_unaccented (the heading without its accents), compound (a German word the article holds only "
        "inside a compound). Extracted from the Fedlex HTML of the four languages. Questions and negatives written by a model (Claude), not by native speakers."
    ),
    "documents": docs,
    "queries": queries,
    "negatives": negatives,
}
def line(o):
    return json.dumps(o, ensure_ascii=False)
def block(key, rows):
    return f'  "{key}": [\n' + ",\n".join("    " + line(r) for r in rows) + "\n  ]"
text = ("{\n" + f'  "name": {line(ds["name"])},\n  "description": {line(ds["description"])},\n'
        + ",\n".join(block(k, ds[k]) for k in ("documents", "queries", "negatives")) + "\n}\n")
assert json.loads(text) == ds
out.write_text(text, encoding="utf-8")
from collections import Counter
print(len(docs), "documents;", len(queries), "queries", dict(Counter(q["category"] for q in queries)), ";", len(negatives), "negatives;", len(text.encode()), "bytes")
