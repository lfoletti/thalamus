# Failing first: gc-21-ingest-signs-writer

Head `03c867e` on `origin/main` `a591ae1` (23 October 2026). Worktree:
`/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-21-ingest-signs-writer`.

## Tests on the base (tests posed, code absent)

All tests compile on `a591ae1`. Command:

```sh
W=<worktree>; B=<scratch dir>
git -C $W worktree add --detach $B a591ae1
git -C $W diff a591ae1 HEAD -- '*_test.go' | git -C $B apply
cd $B
go test ./pkg/cortexdb/ -run 'TestAnIngestRecordsItsWriter|TestWithoutAWriterAnIngest|TestAReIngestByAnotherKey'
go test ./pkg/httpapi/ -run 'TestAToolWriteTakesTheMarks|TestAnIngestByAnUnconfined|TestATypedWriteTakesTheSignature'
go test ./pkg/rpcserver/ -run 'TestATypedKnowledgeWriteTakesTheMarks|TestAnIngestByAnUnconfined'
go test ./pkg/authz/ -run 'TestTheDocumentToolsTakeTheSignatureBack|TestAConfinedKeyDeclaresNoEntity'
```

Observed on the base:

```
--- FAIL: TestAnIngestRecordsItsWriterAndAClientCannotClaimAnother
    written_by = zeus, want hermes
    an ingest kept the updated_by its caller sent: map[... updated_by:zeus written_by:zeus]
--- FAIL: TestAReIngestByAnotherKeyIsSignedByIt
    the same ingest by another key was unchanged, so it keeps the first key's signature
    written_by = , want zeus, who ingested last
--- FAIL: TestAToolWriteTakesTheMarksTheFacadeReplaces (pkg/httpapi)
    the ingested document is signed <nil>, want notes
--- FAIL: TestAnIngestByAnUnconfinedKeyIsSignedByTheKey (pkg/httpapi)
    the document is signed zeus / zeus, want operator and no updated_by
--- FAIL: TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces (pkg/rpcserver)
    CallTool ingest_document with the metadata it read: rpc error: code = PermissionDenied desc = denied: key "notes" is confined and ingest_document takes no metadata.updated_by from a confined key
--- FAIL: TestAnIngestByAnUnconfinedKeyIsSignedByTheKeyOverGRPC (pkg/rpcserver)
    the document is signed "zeus" / "zeus", want operator and no updated_by
--- FAIL: TestTheDocumentToolsTakeTheSignatureBack (pkg/authz)
    ingest_document with the signature it read: denied: ... takes no metadata.written_by from a confined key
```

`TestWithoutAWriterAnIngestStoresMetadataAsGiven` and `TestATypedWriteTakesTheSignatureTheFacadeReplaces` pass on the base: they hold what must not change (no writer, no signature; another spelling and the memory tools stay refused).

## Found while coding

With the allowlist change alone, `TestATypedWriteTakesTheSignatureTheFacadeReplaces` (pkg/rpcserver) fell: `InsertGraphDocument` is authorized against the `ingest_document` line, so it took `written_by` back although it does not sign. The port now refuses the two writer keys on `InsertGraphDocument` itself (`pkg/rpcserver/ownership.go`), and the test has an `updated_by` case.

## Hand mutations (one per rule)

Tool: `docs/pr/lot_E/mut_02_03_05/mutmulti.py` (copy, apply, `go test` under the canary, restore from the copy, check the restore). Specs in `docs/pr/lot_E/mut_02_03_05/m21/`. Replay:

```sh
RUN='TestAnIngestRecords|TestWithoutAWriterAnIngest|TestAReIngestByAnotherKey|TestAToolWriteTakesTheMarks|TestAnIngestByAnUnconfined|TestATypedWriteTakesTheSignature|TestATypedKnowledgeWriteTakesTheMarks|TestTheDocumentToolsTakeTheSignatureBack|TestAConfinedKeyDeclaresNoEntity'
for f in docs/pr/lot_E/mut_02_03_05/m21/M*.json; do
  python3 -I docs/pr/lot_E/mut_02_03_05/mutmulti.py --repo $W --spec $f -- -count=1 -run "$RUN" ./pkg/cortexdb/ ./pkg/authz/ ./pkg/httpapi/ ./pkg/rpcserver/
done
```

| # | Mutation | Tests that fall |
| :- | :- | :- |
| M1 | `stampWriter` removed from `IngestDocument` | `TestAnIngestRecordsItsWriterAndAClientCannotClaimAnother`, `TestAReIngestByAnotherKeyIsSignedByIt`, `TestAToolWriteTakesTheMarksTheFacadeReplaces`, `TestAnIngestByAnUnconfinedKeyIsSignedByTheKey`, `TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces`, `TestAnIngestByAnUnconfinedKeyIsSignedByTheKeyOverGRPC` |
| M2 | `stampWriter` moved after the `sameIngest` comparison (two sites) | `TestAReIngestByAnotherKeyIsSignedByIt` |
| M3 | `documentKeys` without `WriterKeys` | `TestTheDocumentToolsTakeTheSignatureBack`, `TestAToolWriteTakesTheMarksTheFacadeReplaces`, `TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces` |
| M4 | `memory_save` rewritten with `documentKeys` | `TestTheDocumentToolsTakeTheSignatureBack`, `TestATypedWriteTakesTheSignatureTheFacadeReplaces` (httpapi), and two more |
| M5 | `rewritten` compared case-insensitively | `TestAConfinedKeyDeclaresNoEntityOrRelationAndNoEngineKey`, `TestTheDocumentToolsTakeTheSignatureBack`, and four port tests |
| M6 | `stampWriter` without the "no writer" guard | `TestWithoutAWriterAnIngestStoresMetadataAsGiven` |
| M7 | `InsertGraphDocument` check removed | `TestATypedWriteTakesTheSignatureTheFacadeReplaces` (rpcserver) |
| M8 | `InsertGraphDocument` checks `written_by` only | `TestATypedWriteTakesTheSignatureTheFacadeReplaces` (rpcserver) |
| M9 | `WriterKeys` returns `written_by` only | `TestWriterKeysAreTheKeysAWriterIsRecordedUnder`, `TestTheDocumentToolsTakeTheSignatureBack`, `TestAToolWriteTakesTheMarksTheFacadeReplaces`, `TestATypedKnowledgeWriteTakesTheMarksTheFacadeReplaces` |

9 mutations, 9 killed. Full output: `docs/pr/lot_E/mut_02_03_05/m21/results.txt`.

## Class inventory

- Document writes that take metadata: `ingest_document` (signs now), `knowledge_save`, `knowledge_update` (signed on `main`), typed `SaveKnowledge`, `UpdateKnowledge`, `POST /v1/knowledge` (signed and stripped on `main`), `InsertGraphDocument` (not signed, keeps the refusal: M7, M8).
- Memory writes: `memory_save`, `knowledge_memory_remember`, `memory_update` keep the refusal on the tool route (`TestTheDocumentToolsTakeTheSignatureBack`, M4). Out of this PR (question in the body).
- Spellings: exact spelling taken back, any other refused (M5).
