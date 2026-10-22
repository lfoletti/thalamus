## Description

Second of the three PRs for point 2 of discussion #66. Stacked on #72 (the `authz` rules): the first commit here is #72's, the second is this PR. I will rebase onto `main` as soon as #72 lands.

### Changes since the second review (21 October 2026)

Rebased on #72's amended head and on v2.135.0. The two fixes of #72 are held over gRPC by `TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC`: a foreign id and a missing id in `memory_get.memory_id` come back with one wording, `not found`, whichever side (the guard or the tool) did not find the row, while an unconfined key keeps the tool's own answer; a foreign id in `memory_get.memory_ids` comes back under `missing`; `relations[].metadata.chunk_ids` is refused; a `memory_save` refused for a foreign or a missing `supersedes` target leaves no new row, so the new id tells nothing either. All eleven row-id fields of the allowlist are held end to end, not only five: `knowledge_update` and `memory_update` with a foreign and a missing id, and the four create tools with a foreign id. A tool's own refusal that is not `NOT_FOUND` keeps its code and wording for a confined key (an empty `content`), so only the not-found answer is flattened. The list read now also checks that the key's own row is served. Null arguments are a denial, on a tool with a gate to force too, and the server keeps serving: on #72's earlier head they panicked the interceptor, which grpc-go does not recover.

**Typed writes take what the tool takes.** `SaveKnowledge`, `UpdateKnowledge`, `InsertGraphDocument`, `SaveMemory` and `UpdateMemory` now check their arguments against the allowlist line of the tool that does the same write (`knowledge_save`, `knowledge_update`, `ingest_document`, `memory_save`, `memory_update`), before the row check. A confined key that sent entities, relations, chunk ids or a `_` key through a typed RPC got them written, and `written_by` through `InsertGraphDocument`; the same call through `CallTool` was refused. Now both refuse it (`TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes`). This is a change for a confined key on the typed RPCs, named under Breaking Changes. Unconfined keys are not checked.

Two exceptions keep `main`'s answers where the allowlist line would only get in the way. The facade signs `SaveKnowledge`, `UpdateKnowledge`, `SaveMemory` and `UpdateMemory` itself (`cortexdb.WithWriter` replaces `written_by` and `updated_by`), so these four take the two keys from a confined key, and a row read back and sent again is not refused for the signature it carries; `InsertGraphDocument` is not signed and still refuses them, and so does any other spelling, which the facade does not replace (`TestATypedWriteTakesTheSignatureTheFacadeReplaces`). Through `CallTool` the tools keep #72's refusal of the two keys. And a key confined to a collection reaches no memory row, so its `SaveMemory` and `UpdateMemory` are left to the row checks: the interceptor's refusal and `NOT_FOUND`, as on `main` and as `GetMemory` and `DeleteMemory` answer (`TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks`).

**Blank ids.** The typed upsert guards take an id of blanks for no id, while the facade stores it as given, so a confined key can still take another scope's row named `" "` through `SaveKnowledge`, `UpdateKnowledge` or `SaveMemory`; `CallTool` guards it. #82 fixes that on both ports (`TestABlankUpsertIDIsGuardedLikeAnyOtherOverGRPC`); until it lands, the two paths disagree there too.

**One typed write is fixed in a separate PR.** On `main`, `InsertGraphDocument` has no row guard: a key confined to `notes` can call it with the id of a document in `secrets`, overwrite it and move it into `notes`, where `ingest_document` through `CallTool` is guarded here. That is a hole of `main`, independent of this wiring, so it has its own PR, #83 (`guardKnowledgeUpsert` on `InsertGraphDocument`, based on `main`). Until it lands, the typed RPCs and `CallTool` disagree on that one RPC. The two typed upserts of a document whose chunks are gone are withheld here (`SaveKnowledge`, `UpdateKnowledge` in `TestTheTypedRPCsWithholdADocumentTheyCannotPlace`); `InsertGraphDocument` gets the same rule through the guard once #83 lands, and I will add its case to whichever of the two lands second.

`UpdateKnowledge` from a confined key answered a missing id in the facade's words (`get knowledge: vectorstore: get_document: embedding not found`) and a foreign one as `knowledge "X" not found`, so the typed RPC told the two apart. It now answers both as `knowledge "X" not found`; an unconfined key keeps the facade's wording (`TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne`).

A document whose chunks are gone has no collection the facade can read (it names the default one), so the typed RPCs now withhold it from a confined key, as #72's tool guard does: on `main`, a key confined to the default collection could read and delete such a document of any collection through `GetKnowledge` and `DeleteKnowledge` (`TestTheTypedRPCsWithholdADocumentTheyCannotPlace`).

Thirty-six hand mutations, thirty-six killed: a typed upsert of a document without chunks let through, a document without chunks placed by the typed RPCs, the `memory bucket not found` wording passed to a confined key, the tool's wording passed through to a confined key, one wording forced on every caller, every error code flattened for a confined key, any other refusal rewritten for a confined key, the pass through `AuthorizeToolArgs` short-circuited, #72's supersede check after the write, the typed RPCs off the allowlist, and each part of a typed write left out of the check (entities, relations on both knowledge writes, a document's metadata, `SaveMemory`'s metadata, scope and namespace, `UpdateMemory`'s metadata, `UpdateKnowledge`'s collection); for the signature, each of the two keys still refused, each of the four signed writes left out, the unsigned `InsertGraphDocument` let through, another spelling let through; a collection key's memory writes held to the allowlist line again, every key let off it; `UpdateKnowledge`'s missing id in the facade's words, the withheld wording forced on the operator, and every other refusal of the facade rewritten too; the typed memory row check without the row's namespace, or without its scope (`TestTheTypedMemoryRowChecksHoldEveryMemoryField`).

What this PR leaves as it is, so you can see it: the typed RPCs keep `main`'s wording, `memory "X" not found` (`withheld`), which is the same for a foreign and a missing row and names only the id the caller gave (`UpdateKnowledge` now too, above), except a typed `SaveMemory` into another user's session, which keeps `main`'s `memory bucket not found`; `CallTool` answers a bare `not found`. A typed write refused by the allowlist line names the tool it stands for (`tool knowledge_save takes no user_id argument`), where `main` named the RPC; the code is `PERMISSION_DENIED` either way. And `TestAMemoryWriteFromACollectionKeyIsAnsweredByTheRowChecks` passes on `main` by design: it holds `main`'s answers, which this PR's first version changed. And under a confined key, every `NOT_FOUND` of a tool becomes that bare `not found`, `memory bucket not found` included: a little diagnostic lost, nothing told.

### Changes since review (20 October 2026)

`CallTool` now routes a confined key through the allowlist of #72 instead of the schema rule. The interceptor first forces the gates the key needs into the arguments (`ConfineToolArgs`), then checks them against the tool's line (`AuthorizeToolArgs`), and hands the tool the arguments it checked. The five ways out you found on #72 are held over gRPC in `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC`: consolidate refused; a user key's recall reaches no knowledge and a collection key's recall no memory, even with the gate set off in the call; `cortex_query` refused; `extract_conversation` refused a `session_id` and `persist`, allowed text; `supersedes` of another user's memory and of a missing one both `NOT_FOUND`, and the memory not retired; chunk ids on entities refused; a search pinned to another collection's document returns nothing of it.

### Problem
A key with a scope is refused every `CallTool`, because the request carries the tool's arguments as a string of JSON the interceptor cannot see into. The MCP stdio client in remote mode only speaks `CallTool`, so a confined key cannot be handed to a developer's machine at all.

### Fix
`CallTool` routes a confined key through the policy's tool rules instead of refusing it.
- The interceptor sets the gates the key needs (`authz.Key.ConfineToolArgs`), checks the arguments against the named tool's allowlist line (`authz.Key.AuthorizeToolArgs`): tool on the list, every argument on the line, confined field set and equal in every copy and spelling, gate on. It then replaces the request's arguments with the ones it checked. Arguments that are not a JSON object are a denial.
- The handler then applies the row rule (`authz.Scope.GuardToolRows`) where the database is at hand, as every id-addressed typed RPC does: a row named by id that lies outside the scope is `NOT_FOUND`, worded like a row that does not exist.
- The typed RPCs' own row checks (`scopePermitsMemory`, `scopePermitsKnowledge`) call the rules the policy exports, so the two paths cannot disagree.

Unconfined keys, `ListTools` and the read-only clearance are unchanged; the typed RPCs change for confined keys only, as above and under Breaking Changes. A tool off the allowlist (`graph_statistics`) stays refused to a confined key.

### Tests
- `pkg/rpcserver/confined_tools_review_test.go`: the regression test above, with an operator key writing the rows the confined keys must not reach.
- `pkg/rpcserver/tools_auth_test.go`: a key confined to a `user_id` searches its own memory, is refused another user's, is refused a search with no `user_id`, and is refused a tool off the list even when it passes one.
- `pkg/rpcserver/scoped_keys_test.go`: a key confined to one collection ingests and searches it through the tools; a key confined to two collections searches each through `CallTool` and is refused a third, as the typed RPC refuses it; a key confined to `team-acme` passing `collection=team-acme` with the id of a document in `secrets` gets `NOT_FOUND`, and the document is unchanged; a new row of its own passes.
- `pkg/rpcserver/tools_guard_test.go`: the two halves of the check fail closed on their own, without the interceptor in front: a message that is not a `CallToolRequest`, an unknown tool, non-object arguments and a read that fails are refused or reported, never passed; an unconfined caller is not checked.
- `go vet` and `go test -race ./...` pass.

**Failing first.** The tests use the new functions, so the proof is per wiring rule: disable it, keep the package compiling, and a test must fail. For the first review: the gate not forced before the call, the checked arguments not handed to the tool, each fail `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC`; the allowlist not consulted fails it and four tests of `scoped_keys_test.go`. For the second review, the thirty-six above. Every line of the patch has a test.

Part of a small series on running one CortexDB server for several teams (see discussion #66, item GC-04).

## Type of Change
- [x] New feature
- [x] Breaking change (fix or feature that would cause existing functionality to not work as expected), for confined keys only

## Testing
- [x] Unit tests pass
- [x] Integration tests pass
- [x] New tests have been added

## Performance Impact
For a confined key only: one JSON decode and one encode of the arguments per tool call, and one read per id argument; a document read loads its chunks, so its cost grows with the document, and a typed upsert that creates a row reads twice. Nothing for an unconfined key.

## Breaking Changes
One, for confined keys only. A typed write (`SaveKnowledge`, `UpdateKnowledge`, `InsertGraphDocument`, `SaveMemory`, `UpdateMemory`) from a confined key is now refused what the allowlist line of its tool refuses: declared entities and relations, chunk ids and `_` keys in metadata, and on `InsertGraphDocument` `written_by` and `updated_by`. Before, it wrote them. The four writes the facade signs still take `written_by` and `updated_by`, which the facade replaces, as before. A key confined to a collection gets `main`'s answers on `SaveMemory` and `UpdateMemory`. A confined key that was refused every tool is now allowed the tools on the allowlist, within its scope. Unconfined keys see no change.

