# Réponses au relecteur, lot C (brouillons du 20 octobre 2026)

À poster **sur commande de Lauro**, après la poussée du lot. Un commentaire par PR, factuel, point par point. Pas de commentaire sur ses outils.

## #75 (`gc-08-content-hash`)

Thanks for the review. Amended in place, rebased on v2.134.1. Each point has a test written first (`pkg/cortexdb/graphrag_ingest_review_test.go`):

1. Dedupe is decided within the collection: `duplicate_of` never names a document of another one.
2. "Unchanged" compares the chunking (now stored as `chunk_size`, `chunk_overlap`) and the collection the chunks sit in, not only the hash.
3. The hash follows the content on `knowledge_update` and `knowledge_save`, and sits on the document node for both write paths.
4. `ingest_document` writes through the same mutation plan as `knowledge_save`, in one transaction, and retires the chunks of the previous text. It fit in this PR: 226 lines outside tests.
5. The ids of an unchanged re-ingest come in `chunk_index` order (held at 1001 chunks).

Eight mutations for the amendment, eight killed (body).

## #76 (`gc-05-expires-at`)

Amended in place, rebased on #75's new head. Each point has a test written first:

1. The vector path: hybrid seeds and `search_text` go through the same gate as the lexical path, and so does each lane of `cortex_query`, which reads the store on its own.
2. The chunk tools and `expand_graph` leave an expired document out, with its edges.
3. `--sweep-expired` with `CORTEXDB_REMOTE` set sweeps the shared brain: the sweep is a tool of the toolbox now (`sweep_expired`, with `dry_run`), and the command calls it over gRPC with the token. The local file, if any, is not touched.
4. An expression index on `expires_at` in both backends, and the gate query uses the predicate it serves. The PostgreSQL statement is not run here: no PostgreSQL in my environment.

Thirteen mutations, thirteen killed (body). 505 lines outside tests, 105 over your 400; the tool and the remote sweep are about 180 of them.

## #77 (`gc-06-review-status`)

Amended in place, rebased on #76's new head. Each point has a test written first (`pkg/cortexdb/review_review_test.go`):

1. The hybrid path is gated, through #76's gate, `cortex_query`'s lanes included.
2. `reviewed_by` of a decision is the calling key when there is one, whatever the client wrote.
3. No silent clearing: a write that says nothing about the review carries the decision over; an empty status on a reviewed row is refused; a non-string status is refused.
4. A `knowledge_update` that decides on the review is merged into the stored metadata rather than replacing it, so an acceptance that names only the review keys keeps `expires_at`, the ingest marks and the client's own keys. An update that says nothing about the review replaces the map as before.
5. `reviewed_at` is kept on a re-save of the same decision, so an identical re-ingest of an accepted document is `unchanged`.
6. With the gate on, an unreadable brain serves nothing.
7. An expression index on `review_status`, same shape as #76's.

Fifteen mutations, fifteen killed (body). 451 lines outside tests, 51 over your 400; the carry-over and merge rules (points 3 and 4) are most of the excess.
