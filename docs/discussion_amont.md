# Discussion GitHub amont : texte final

Cible : https://github.com/liliang-cn/cortexdb/discussions (catégorie « Ideas » si elle existe, sinon « General »).
État : **publiée le 18 octobre 2026** par Lauro : https://github.com/liliang-cn/cortexdb/discussions/66

## Titre

CortexDB as the shared brain of several teams: a small series of fixes

## Corps

We run one `cortexdb-grpc` with a key file, one collection per team, and the MCP stdio client in remote mode on each developer's machine. The aim is a brain that one organisation can shape to its own granularity: one collection per team, per project or per source, and keys that reach one collection, several, or all of them. The same server then has to answer the questions a compliance review asks of it: who wrote what, from where, until when, and who approved it.

Trying this end to end showed a handful of gaps, each small and each reproduced by a test that fails before and passes after. I would like to send them one PR at a time, in three batches, so each one is easy to review or to decline on its own:

1. **Fixes with no design question.** The session-end launcher leaves a `capture-payload.*` file behind. The request log cannot say which key called which tool. `ingest_document` drops `source_url` and `author`. A `graph` search confined to a collection walks into another collection through an entity both collections mention.
2. **Scoped keys and the toolbox.** Today a confined key is refused every tool, so the MCP client cannot use one. Proposal: allow the tools that declare the confined field in their input schema, hold every nesting and every spelling of that field to the key's value, and check that a row named by id already lies within the scope. A key could also list several values per field (`collections: ["team-a", "shared"]`). I would value your view on the principle before sending the code.
3. **Document lifecycle.** A content hash on ingest, so an identical re-ingest writes nothing and a duplicate can be refused. An `expires_at` on documents with a sweep command. A `review_status` on documents and memories, so a proposed fact is hidden from search until someone accepts it.

Each PR stands alone and none changes the behaviour of an unconfined key or of a document without these fields. Happy to drop or reshape any of them.

## Pied de page des PR

> Part of a small series on running one CortexDB server for several teams (see discussion #66, item GC-xx). Each PR stands alone.
