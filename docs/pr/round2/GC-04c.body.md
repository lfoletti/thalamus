## Description

Third of the three PRs for point 2 of discussion #66. Stacked on #72 (the `authz` rules): the first commit here is #72's, the second is this PR. I will rebase onto `main` as soon as #72 lands. Independent of the gRPC wiring (#73): the two can merge in either order.

### Changes since the second review (21 October 2026)

Rebased on #72's amended head and on v2.135.0. The two fixes of #72 are held over REST by `TestTheRoundTwoRulesOfTheAllowlistHoldOverREST`, the same cases as #73's gRPC test: one wording, `not found`, for a foreign and a missing single id under a confined key, the tool's own answer for an unconfined one, a foreign id of a list under `missing`, `relations[].metadata.chunk_ids` refused, no new row left by a `memory_save` refused for its `supersedes`. Nothing for `/v1/recall`, fixed on `main`.

All eleven row-id fields of the allowlist are held end to end, as on gRPC. A tool's own refusal that is not `404` keeps its status and wording for a confined key (an empty `content`). A body that is not valid JSON, such as `{"query":`, is `400` and says "JSON object"; a body of `null` is `400` too, and the server keeps serving. The list read also checks that the key's own row is served.

**Typed writes take what the tool takes.** `POST /v1/knowledge` and `POST /v1/memory` now check the fields of their body that the allowlist line of `knowledge_save` and `memory_save` reads (collection, user scope, metadata, entities, relations) against that line, as #73 does on gRPC (`TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes`). A confined key that sent entities or a `_` key through the typed route got them written; through the tool route it was refused. Named under Breaking Changes.

The check decodes the body the way the handler does, into the facade's request type, with the same decoder: the decoder matches a field name by Unicode case folding, so `Entities` or `entitieſ` (a long s) is its `Entities`, and a check by name would miss it. Two exceptions keep `main`'s answers, as on gRPC (`pkg/httpapi/confined_typed_writers_test.go`). The facade signs both typed writes (`cortexdb.WithWriter` replaces `written_by` and drops `updated_by`), so the two routes take those two keys from a confined key, and a row read back and sent again is not refused for the signature it carries; the row keeps the facade's signature, and another spelling, which the facade does not replace, is still refused, as on the tool route. And a body the decoder refuses is the handler's `400`, as on `main`, not an allowlist refusal. That covers a key that reaches no row of the route's store (one confined to a collection on `/v1/memory`, one confined to `user_id`, `scope` or `namespace` on `/v1/knowledge`): the scope field it must send is not a field of the route, so it is `400` `unknown field`, and a body without it is the row checks' `403`.

Decoding also closes a hole of `main` on `POST /v1/knowledge`: the decoder keeps the last of two spellings, so `{"collection":"notes","Collection":""}` from a key confined to `notes` passed the row checks (which skip an empty value) and stored the document in `graphrag_chunks`. It is now `403` "leaves collection unset".

**Rows named by id.** On `main`, `POST /v1/knowledge` and `POST /v1/memory` check the scope fields of the body only, so a confined key that names another scope's id overwrites that row and moves it into its own scope. That is a hole of `main`, independent of this wiring; #82 fixes it, with blank ids, on both ports. #82 and this PR touch the same lines of `authorizeWith`: whichever lands second gets the rebase.

Twenty-six hand mutations in this round, twenty-six killed: the tool's wording passed through to a confined key, one wording forced on every key, every status flattened for a confined key, the `unknown tool` answer passed through, the pass through `AuthorizeToolArgs` short-circuited in the route, #72's supersede check after the write, broken JSON taken as an object, the typed routes off the allowlist, the declared entities and relations of a decoded body left out (on each route), the collection read by name instead of decoded, a body the decoder refuses held to the allowlist, each part of a typed body left out of the check (entities, relations, metadata on each route, the collection, the memory scope fields together, `namespace` alone, `scope` alone), empty lists of entities or relations handed over, `written_by` or `updated_by` handed over, and every spelling of `written_by` dropped.

Under a confined key, every `404` of a tool becomes the bare `not found`, `memory bucket not found` included, and so does the toolbox's `unknown tool` answer, whose text can carry an id the caller sent (`TestAToolNotFoundThatNamesAnUnknownToolIsWithheldLikeAnyOther`): a little diagnostic lost, nothing told.

### Changes since review (20 October 2026)

- `db == nil` fails closed: with no database to read rows from, the tool route refuses a confined key (`500`, "no database to guard rows against") instead of skipping the row guard. `TestTheToolRouteRefusesAConfinedKeyWhenItHasNoDatabaseToGuardWith` holds it on `authorize` without a brain.
- The route routes a confined key through the allowlist of #72: gates forced into the body (`ConfineToolArgs`), body checked against the tool's line (`AuthorizeToolArgs`), the tool called with the body that was checked. The five ways out you found on #72 are held over REST in `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST`, the same cases as #73's gRPC test.
- `CallerContext` in the route is kept.

### Problem
A key with a scope is refused every `POST /v1/tools/<name>`, because the body is the tool's arguments, which the route's own row check cannot see into.

### Fix
The tool route routes a confined key through the policy's tool rules instead of refusing it, the same way `CallTool` does over gRPC. Once clearance has passed, the body is read; a body that is not a JSON object is `400`. The gates the key needs are set in it (`authz.Key.ConfineToolArgs`), it is checked against the named tool's allowlist line (`authz.Key.AuthorizeToolArgs`), and the row rule (`authz.Scope.GuardToolRows`) runs against the database: a row named by id that lies outside the scope is `404`, worded like a row that does not exist. The tool then reads the body the policy checked.

Unconfined keys and the read-only clearance are unchanged. The typed routes changed in the second review (above). A tool off the allowlist stays refused to a confined key.

### Tests
- `pkg/httpapi/confined_tools_review_test.go`: the regression test above, and the route without a database.
- `pkg/httpapi/scoped_keys_test.go`: a key confined to `notes` ingests and searches in `notes`, is refused `secrets`, is refused a search with no collection, is refused `graph_statistics`, gets `400` for a non-object body; naming a document of `secrets` with `collection=notes` is `404` on `knowledge_update` and on `ingest_document`, and a new row of its own passes. `TestTheToolRouteFailsClosedForAConfinedKey`: a body over the size limit is `413` before any check; a row read that fails (database closed between the argument check and the row check) is `500`. `TestARowGuardFailureHasItsOwnStatus`: the guard's failure is mapped by `toolGuardStatus` (outside the scope `404`, the guard's own refusal `403`, a database failure `500`).
- `pkg/httpapi/confined_tools_round2_test.go`: the fixes of #72 over REST (second review).
- `pkg/httpapi/confined_typed_args_test.go` and `pkg/httpapi/confined_typed_writers_test.go`: the typed routes against the allowlist, the two exceptions, the decoder's spellings, and the `unknown tool` wording.
- `go vet` and `go test -race ./...` pass.

**Failing first.** The tests use the new functions, so the proof is per wiring rule: disable it, keep the package compiling, and a test must fail. The first round's seven mutations, seven killed (the second review's are above): the gate not forced, the row guard skipped, or the checked body not handed to the tool, each fail `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST`; the allowlist not consulted fails it and `TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP`; the `db == nil` refusal removed fails `TestTheToolRouteRefusesAConfinedKeyWhenItHasNoDatabaseToGuardWith`; a non-object body admitted, or an empty body answered as a syntax error rather than by the policy, fails `TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP`. The automatic pass over the diff leaves none alive. Every line of the patch has a test.

Part of a small series on running one CortexDB server for several teams (see discussion #66, item GC-04).

## Type of Change
- [x] New feature
- [x] Breaking change (fix or feature that would cause existing functionality to not work as expected), for confined keys only

## Testing
- [x] Unit tests pass
- [x] New tests have been added

## Performance Impact
For a confined key only: the body is buffered once, decoded and encoded once, and one indexed read is made per id argument. Nothing for an unconfined key.

## Breaking Changes
One, for confined keys only. `POST /v1/knowledge` and `POST /v1/memory` from a confined key now refuse what the allowlist line of `knowledge_save` and `memory_save` refuses: declared entities and relations, chunk ids and `_` keys in metadata. Before, they wrote them. `written_by` and `updated_by`, which the facade replaces, still pass, and so does a key of the other store, which gets `main`'s answers. Unconfined keys see no change.

