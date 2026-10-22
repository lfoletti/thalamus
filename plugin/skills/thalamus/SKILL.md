---
name: thalamus
description: How to use the shared brains of the organisation through the Thalamus gateway (th_brains, th_recall, th_get, th_add, th_pending, th_review), and what the recall hook injects.
---

# Thalamus: shared brains

The gateway gives you two tools, under your own key. The engine applies your rights on each brain; the gateway records who asked what.

- `th_brains`: the brains you may read, your access (read, write or curate), and whether a key exists for each.
- `th_recall {query, top_k, brains}`: one question to every brain you may read, fused into one list. Each result carries a `provenance` line: `[brain] title, id, source, by author`. `skipped` names the brains that did not answer and why.
- `th_recall` serves first the syntheses of a brain's pyramid, when the brain has one: a result with `level: synthesis` and the provenance `[brain] synthesis: title, id`. A synthesis is written by a model from the documents it cites. Open those documents with `th_descend {brain, knowledge_id}` before you rely on a figure, a date or a name of the synthesis. `syntheses` (default 2, 0 for none) sets how many come first.
- `th_get {brain, knowledge_id}`: one document whole. The hook and `th_recall` give extracts, which are cuts: when the answer is not in the extract, read the document with `th_get`, with the brain and the id of its provenance line.
- `th_add {brain, content, title, source_url, document_id, metadata}`: proposes a document into a brain you may write. It waits for a curator; your own document is served at once only when you curate that brain.
- `th_pending {brain}` and `th_review {brain, knowledge_id, decision}`: for curators, what waits and the decision (`accepted` or `rejected`).

The hook on each prompt already injects the top 3 extracts with their provenance. When the extract stops before the answer, call `th_get` once; do not call `th_recall` again with the same query, it returns the same extracts. When you need more documents, or a narrower search, call `th_recall` with a precise query, and name `brains` to restrict it.

Rules:

1. Quote the provenance line when you use an extract. The reader must know which brain and which document it comes from.
2. Write through `th_add` only, and tell the user that the document waits for review. Never set `review_status` yourself: the gateway does.
3. A brain in `skipped` with "no key issued" means you have the right but no key yet: ask an administrator for `th-gateway key issue`.
