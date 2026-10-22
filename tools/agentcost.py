"""Coût d'un agent : tours, tokens, outils, d'après son journal JSONL.

Usage : python3 -I tools/agentcost.py <fichier .output de l'agent>
La lecture du cache domine : chaque tour relit tout le contexte.
"""
import collections, json, sys

turns, seen, tools = [], set(), collections.Counter()
for line in open(sys.argv[1]):
    try:
        d = json.loads(line)
    except ValueError:
        continue
    m = d.get("message") or {}
    if d.get("type") != "assistant":
        continue
    u, mid = m.get("usage") or {}, m.get("id")
    if u and mid not in seen:
        seen.add(mid)
        turns.append((u.get("input_tokens", 0), u.get("cache_creation_input_tokens", 0),
                      u.get("cache_read_input_tokens", 0), u.get("output_tokens", 0)))
    for c in m.get("content") or []:
        if c.get("type") == "tool_use":
            tools[c["name"]] += 1
if not turns:
    sys.exit("aucun tour lu")
cw, cr = sum(t[1] for t in turns), sum(t[2] for t in turns)
print(f"tours {len(turns)}, contexte {sum(turns[0][:3])} -> {sum(turns[-1][:3])} tokens")
print(f"écriture cache {cw}, lecture cache {cr}, entrée neuve {sum(t[0] for t in turns)}")
print(f"équivalent entrée (lecture x0,1, écriture x1,25) : {int(cr * 0.1 + cw * 1.25)}")
print("outils :", dict(tools))
