"""Extracts the articles of the Federal Constitution (Fedlex HTML, SR 101, 2024-03-03) into articles.json:
{article id: {lang: {"title", "text"}}}, for de, fr, it and rm."""
import html, json, re
from pathlib import Path
HERE = Path(__file__).parent
out = {}
for lang in ("de", "fr", "it", "rm"):
    src = (HERE / "sources" / f"cst_{lang}.html").read_text(encoding="utf-8")
    for m in re.finditer(r'<article id="(art_[0-9a-z_]+)">(.*?)</article>', src, re.S):
        aid, body = m.group(1), m.group(2)
        head = re.search(r"<h6[^>]*>(.*?)</h6>", body, re.S)
        title = re.sub(r"<[^>]+>", " ", head.group(1)) if head else ""
        body = re.sub(r"<h6.*?</h6>", " ", body, flags=re.S)
        body = re.sub(r"<sup>.*?</sup>", "", body, flags=re.S)
        body = re.sub(r'<div class="footnotes">.*', "", body, flags=re.S)
        text = re.sub(r"<[^>]+>", " ", body)
        clean = lambda s: re.sub(r"\s+", " ", html.unescape(s)).strip()
        title = re.sub(r"(\s+\d+)?(\s*\*)?$", "", clean(title))
        out.setdefault(aid, {})[lang] = {"title": title, "text": clean(text)}
(HERE / "articles.json").write_text(json.dumps(out, ensure_ascii=False, indent=0), encoding="utf-8")
print(len(out), "articles;", sum(1 for a in out.values() if len(a) == 4), "in 4 languages")
