## Description

Third of the three PRs for point 2 of discussion #66. Stacked on #72 (the `authz` rules): the first commit here is #72's, the second is this PR. I will rebase onto `main` as soon as #72 lands. Independent of the gRPC wiring (#73): the two can merge in either order.

### Changes since review (20 October 2026)

- `db == nil` fails closed: with no database to read rows from, the tool route refuses a confined key (`500`, "no database to guard rows against") instead of skipping the row guard. `TestTheToolRouteRefusesAConfinedKeyWhenItHasNoDatabaseToGuardWith` holds it on `authorize` without a brain.
- The route routes a confined key through the allowlist of #72: gates forced into the body (`ConfineToolArgs`), body checked against the tool's line (`AuthorizeToolArgs`), the tool called with the body that was checked. The five ways out you found on #72 are held over REST in `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST`, the same cases as #73's gRPC test.
- Rebased on v2.134.1; `CallerContext` in the route is kept.

### Problem
A key with a scope is refused every `POST /v1/tools/<name>`, because the body is the tool's arguments, which the route's own row check cannot see into.

### Fix
The tool route routes a confined key through the policy's tool rules instead of refusing it, the same way `CallTool` does over gRPC. Once clearance has passed, the body is read; a body that is not a JSON object is `400`. The gates the key needs are set in it (`authz.Key.ConfineToolArgs`), it is checked against the named tool's allowlist line (`authz.Key.AuthorizeToolArgs`), and the row rule (`authz.Scope.GuardToolRows`) runs against the database: a row named by id that lies outside the scope is `404`, worded like a row that does not exist. The tool then reads the body the policy checked.

Unconfined keys, the typed routes and the read-only clearance are unchanged. A tool off the allowlist stays refused to a confined key.

### Tests
- `pkg/httpapi/confined_tools_review_test.go`: the regression test above, and the route without a database.
- `pkg/httpapi/scoped_keys_test.go`: a key confined to `notes` ingests and searches in `notes`, is refused `secrets`, is refused a search with no collection, is refused `graph_statistics`, gets `400` for a non-object body; naming a document of `secrets` with `collection=notes` is `404` on `knowledge_update` and on `ingest_document`, and a new row of its own passes. `TestTheToolRouteFailsClosedForAConfinedKey`: a body over the size limit is `413` before any check; a row read that fails (database closed between the argument check and the row check) is `500`. `TestARowGuardFailureHasItsOwnStatus`: the guard's failure is mapped by `toolGuardStatus` (outside the scope `404`, the guard's own refusal `403`, a database failure `500`).
- `go vet` and `go test -race ./...` pass.

**Failing first.** The tests use the new functions, so the proof is per wiring rule: disable it, keep the package compiling, and a test must fail. Seven mutations, seven killed: the gate not forced, the row guard skipped, or the checked body not handed to the tool, each fail `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverREST`; the allowlist not consulted fails it and `TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP`; the `db == nil` refusal removed fails `TestTheToolRouteRefusesAConfinedKeyWhenItHasNoDatabaseToGuardWith`; a non-object body admitted, or an empty body answered as a syntax error rather than by the policy, fails `TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP`. The automatic pass over the diff leaves none alive. Every line of the patch has a test.

Part of a small series on running one CortexDB server for several teams (see discussion #66, item GC-04).

## Type of Change
- [x] New feature (non-breaking change which adds functionality)

## Testing
- [x] Unit tests pass
- [x] New tests have been added

## Performance Impact
For a confined key only: the body is buffered once, decoded and encoded once, and one indexed read is made per id argument. Nothing for an unconfined key.

## Breaking Changes
None.
