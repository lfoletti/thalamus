"""Swiss cross-language recall bench. Loads the corpus into the brain ch through the gateway
(knowledge_save), then asks th_recall each question, with and without the variants in the other
languages, and reports where the expected article ranks. Usage: eval.py <gateway> <key> <label>

Corpus: the articles of the Federal Constitution (SR 101, 2024-03-03), each in one language only,
de, fr, it and rm in turn. A question is never in the language of its article; no question or
variant is in Romansh."""
import json, sys, time, urllib.request
from collections import defaultdict
from pathlib import Path

HERE = Path(__file__).parent
LANGS = ("de", "fr", "it", "rm")


def corpus():
    arts = json.loads((HERE / "articles.json").read_text(encoding="utf-8"))
    docs = {}
    for i, (aid, v) in enumerate(arts.items()):
        lang = LANGS[i % 4]
        if len(v["fr"]["text"]) >= 80:  # repealed articles keep a stub only
            docs[aid] = {"lang": lang, "title": v[lang]["title"], "text": v[lang]["text"]}
    return docs


class Session:
    def __init__(self, url, key):
        self.url, self.n = url, 1
        self.h = {"Content-Type": "application/json", "Accept": "application/json, text/event-stream", "Authorization": "Bearer " + key}
        r = self._post({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "eval_ch", "version": "0"}}})
        self.h["Mcp-Session-Id"] = r.headers["Mcp-Session-Id"]
        r.read()
        self._post({"jsonrpc": "2.0", "method": "notifications/initialized"}).read()

    def _post(self, body):
        return urllib.request.urlopen(urllib.request.Request(self.url, json.dumps(body).encode(), self.h), timeout=300)

    def call(self, name, args):
        self.n += 1
        raw = self._post({"jsonrpc": "2.0", "id": self.n, "method": "tools/call", "params": {"name": name, "arguments": args}}).read().decode()
        data = [l[5:] for l in raw.splitlines() if l.startswith("data:")]
        res = json.loads(data[-1] if data else raw)["result"]
        text = res["content"][0]["text"]
        if res.get("isError"):
            raise RuntimeError(text)
        return json.loads(text)


def main():
    gw, key, label = sys.argv[1:4]
    docs = corpus()
    w = Session(f"{gw}/mcp/ch", key)
    t0 = time.time()
    for aid, d in docs.items():
        w.call("knowledge_save", {"knowledge_id": aid, "title": d["title"], "content": d["title"] + "\n" + d["text"], "collection": "ch"})
    print(f"{label}: {len(docs)} articles saved in {time.time() - t0:.0f} s", flush=True)
    r = Session(f"{gw}/mcp", key)
    qs = json.loads((HERE / "questions.json").read_text(encoding="utf-8"))
    rows = []
    for q in qs:
        row = {"id": q["id"], "article": q["article"], "doc_lang": docs[q["article"]]["lang"], "q_lang": q["lang"]}
        for arm, extra in (("query", {}), ("variants", {"alternate_queries": list(q["v"].values())})):
            got = r.call("th_recall", {"query": q["q"], "top_k": 10, "syntheses": 0, **extra})["results"]
            ids = list(dict.fromkeys(h["knowledge_id"] for h in got))
            row[arm] = ids.index(q["article"]) + 1 if q["article"] in ids else 0
        rows.append(row)
    (HERE / "results").mkdir(exist_ok=True)
    (HERE / "results" / f"{label}.json").write_text(json.dumps(rows, indent=0))
    for arm in ("query", "variants"):
        groups = defaultdict(list)
        for row in rows:
            groups["all"].append(row[arm])
            groups[row["doc_lang"]].append(row[arm])
        for g in ("all",) + LANGS:
            v = groups[g]
            if v:
                print(f"{label:8} {arm:8} {g:3} n={len(v):2}  @1 {sum(1 for x in v if x == 1)/len(v):.2f}  @5 {sum(1 for x in v if 0 < x <= 5)/len(v):.2f}  @10 {sum(1 for x in v if x)/len(v):.2f}  MRR {sum(1/x for x in v if x)/len(v):.2f}")


main()
