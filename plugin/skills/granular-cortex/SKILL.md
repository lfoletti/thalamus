---
name: granular-cortex
description: How to use the shared brains of the organisation through the granular_cortex gateway (gc_brains, gc_recall), and what the recall hook injects.
---

# granular_cortex: shared brains

The gateway gives you two tools, under your own key. The engine applies your rights on each brain; the gateway records who asked what.

- `gc_brains`: the brains you may read, your access (read or write), and whether a key exists for each.
- `gc_recall {query, top_k, brains}`: one question to every brain you may read, fused into one list. Each result carries a `provenance` line: `[brain] title, id, source, by author`. `skipped` names the brains that did not answer and why.

The hook on each prompt already injects the top 3 extracts with their provenance. When you need more, or a narrower search, call `gc_recall` with a precise query, and name `brains` to restrict it.

Rules:

1. Quote the provenance line when you use an extract. The reader must know which brain and which document it comes from.
2. Do not write to a brain through these tools: writes go through the per-brain endpoints and the review process of the brain's curators.
3. A brain in `skipped` with "no key issued" means you have the right but no key yet: ask an administrator for `gc-gateway key issue`.
