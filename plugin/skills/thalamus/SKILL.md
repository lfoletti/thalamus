---
name: thalamus
description: How to use the shared brains of the organisation through the Thalamus gateway (th_brains, th_recall, th_entity, th_get, th_add, th_pending, th_review), and what the recall hook injects.
---

# Thalamus: shared brains

The gateway gives you two tools, under your own key. The engine applies your rights on each brain; the gateway records who asked what.

- `th_brains`: the brains you may read, your access (read, write or curate), and whether a key exists for each.
- `th_recall {query, top_k, brains, alternate_queries, deadline_ms}`: one question to every brain you may read, fused into one list. Each result carries a `provenance` line: `[brain] title, id, source, by author`. `skipped` names the brains that did not answer and why. Without `deadline_ms`, the gateway waits for each brain. With `deadline_ms` (at most 60000), it answers at the deadline with the brains that answered, and skips the others with the reason `no answer within N ms`.
- `th_recall` serves first the syntheses of a brain's pyramid, when the brain has one: a result with `level: synthesis` and the provenance `[brain] synthesis: title, id`. A synthesis is written by a model from the documents it cites. Open those documents with `th_descend {brain, knowledge_id}` before you rely on a figure, a date or a name of the synthesis. `syntheses` (default 4, 0 for none) sets how many synthesis chunks come first.
- `th_entity {entity, brains, skip_brains, max_depth, limit}`: one entity across the brains you may read. `known_in` lists the brains that know it, and each neighbor names the brains that serve it. A brain serves only the part of its graph that your key may read. A brain without entities answers nothing: the entities come from the producer of the brain, not from `th_add`.
- `th_get {brain, knowledge_id}`: one document whole. The hook and `th_recall` give extracts, which are cuts: when the answer is not in the extract, read the document with `th_get`, with the brain and the id of its provenance line.
- `th_add {brain, content, title, source_url, document_id, metadata}`: proposes a document into a brain you may write. It waits for a curator; your own document is served at once only when you curate that brain.
- `th_pending {brain}` and `th_review {brain, knowledge_id, decision}`: for curators, what waits and the decision (`accepted` or `rejected`).

The hook on each prompt already injects the best chunks of the documents (4 synthesis chunks and 10 passages at most), each with its provenance. A chunk is a passage of the document, not the whole document. When the extract stops before the answer, call `th_get` once; do not call `th_recall` again with the same query, it returns the same extracts. When you need more documents, or a narrower search, call `th_recall` with a precise query, and name `brains` to restrict it. The documents can be in another language than the question: give the question in that language, or in the other languages of the organization, in `alternate_queries` (at most 4). Each variant is searched apart, then fused with the query. On a French question over English documents, an English variant raised the facts found from 81 % to 85 %.

Rules:

1. Quote the provenance line when you use an extract. The reader must know which brain and which document it comes from.
2. Write through `th_add` only, and tell the user that the document waits for review. Never set `review_status` yourself: the gateway does.
3. A brain in `skipped` with "no key issued" means you have the right but no key yet: ask an administrator for `th-gateway key issue`.
