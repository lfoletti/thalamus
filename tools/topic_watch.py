#!/usr/bin/env python3
"""Watch the GitHub topics of a repository and find rising topics near it.

Usage: python3 -I tools/topic_watch.py [owner/repo] [--out DIR] [--seeds a,b,c]

For each topic of the repository and for each candidate topic, the script
counts the public repositories with that topic (total) and the ones created
in the last 30 days (new). The ratio new/total is the momentum of the topic.
Candidates are the topics that occur most often on the most starred
repositories created in the last 30 days under the seed topics.

Each run appends one snapshot to DIR/snapshots.jsonl and writes
DIR/report-<date>.md. With a previous snapshot, the report also gives the
change of the new-repository count. The script reads GitHub through the gh
CLI and changes nothing on GitHub.
"""
import collections
import datetime
import json
import pathlib
import subprocess
import sys
import time

SEEDS = ["mcp", "claude-code", "knowledge-graph", "agent-memory", "rag", "ai-governance"]
CANDIDATES = 25        # candidate topics measured per run
MIN_HITS = 3           # a candidate occurs on at least this many sample repositories
SEARCH_PAUSE = 2.2     # GitHub search allows 30 requests per minute
MAX_TOPICS = 20        # GitHub limit per repository


def gh(path):
    out = subprocess.run(["gh", "api", path], check=True, capture_output=True, text=True).stdout
    time.sleep(SEARCH_PAUSE)
    return json.loads(out)


def count(topic, since):
    total = gh(f"search/repositories?q=topic:{topic}&per_page=1")["total_count"]
    new = gh(f"search/repositories?q=topic:{topic}+created:>{since}&per_page=1")["total_count"]
    return {"total": total, "new": new, "momentum": round(new / total, 3) if total else 0.0}


def candidates(own, since, seeds):
    hits = collections.Counter()
    for seed in seeds:
        page = gh(f"search/repositories?q=topic:{seed}+created:>{since}&sort=stars&order=desc&per_page=50")
        for repo in page["items"]:
            hits.update(set(repo.get("topics", [])))
    return [t for t, n in hits.most_common() if n >= MIN_HITS and t not in own][:CANDIDATES]


def report(repo, day, own, cand, prev):
    def row(t, m):
        p = prev.get(t)
        delta = f"{m['new'] - p['new']:+d}" if p else "n/a"
        return f"| `{t}` | {m['total']} | {m['new']} | {m['momentum']:.3f} | {delta} |"

    head = "| Topic | Total | New (30 days) | Momentum | Change of new |\n| :- | -: | -: | -: | -: |"
    weakest = sorted(own, key=lambda t: (own[t]["momentum"], own[t]["new"]))
    floor = min(m["new"] for m in own.values()) if own else 0
    median = sorted(m["momentum"] for m in own.values())[len(own) // 2] if own else 0
    rising = [t for t, m in cand.items() if m["new"] >= floor and m["momentum"] >= median]
    lines = [f"# Topic watch: {repo}, {day}", "",
             f"Momentum = new repositories in 30 days / all repositories with the topic. "
             f"The repository has {len(own)} topics of {MAX_TOPICS}.", "",
             "## Rising candidates", ""]
    if rising:
        lines.append("These candidates have at least the new count of the weakest topic and at least the median momentum of the repository topics. "
                     "Add one only if it describes the project.")
        lines += ["", head] + [row(t, cand[t]) for t in rising]
        lines += ["", "Weakest repository topics, first to replace: " + ", ".join(f"`{t}`" for t in weakest[:3]) + "."]
    else:
        lines.append("No candidate passes the two limits.")
    lines += ["", "## Repository topics", "", head] + [row(t, own[t]) for t in sorted(own, key=lambda t: -own[t]["momentum"])]
    lines += ["", "## Other candidates", "", head] + [row(t, cand[t]) for t in cand if t not in rising]
    return "\n".join(lines) + "\n"


def parse_args(argv, home):
    """[owner/repo] [--out DIR] [--seeds a,b,c] -> (repo, out, seeds)."""
    repo, out, seeds = "lfoletti/thalamus", home / ".thalamus" / "topic_watch", SEEDS
    args = list(argv)
    if "--seeds" in args:
        i = args.index("--seeds")
        seeds = [t for t in args[i + 1].split(",") if t]
        del args[i:i + 2]
    if "--out" in args:
        i = args.index("--out")
        out = pathlib.Path(args[i + 1])
        del args[i:i + 2]
    if args:
        repo = args[0]
    return repo, out, seeds


def main(argv):
    repo, out, seeds = parse_args(argv, pathlib.Path.home())
    out.mkdir(parents=True, exist_ok=True)
    today = datetime.date.today()
    since = (today - datetime.timedelta(days=30)).isoformat()

    names = json.loads(subprocess.run(["gh", "repo", "view", repo, "--json", "repositoryTopics"],
                                      check=True, capture_output=True, text=True).stdout)
    own_names = [t["name"] for t in names.get("repositoryTopics") or []]
    own = {t: count(t, since) for t in own_names}
    cand = {t: count(t, since) for t in candidates(set(own_names), since, seeds)}

    snaps = out / "snapshots.jsonl"
    prev = {}
    if snaps.exists():
        lines = [l for l in snaps.read_text().splitlines() if l.strip()]
        if lines:
            prev = json.loads(lines[-1])["topics"]
    with snaps.open("a") as f:
        f.write(json.dumps({"date": today.isoformat(), "repo": repo, "own": own_names,
                            "topics": {**own, **cand}}) + "\n")
    path = out / f"report-{today.isoformat()}.md"
    path.write_text(report(repo, today.isoformat(), own, cand, prev))
    print(path)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
