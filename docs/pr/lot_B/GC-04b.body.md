## Description

Second of the three PRs for point 2 of discussion #66. Stacked on #72 (the `authz` rules): the first commit here is #72's, the second is this PR. I will rebase onto `main` as soon as #72 lands.

### Changes since review (20 October 2026)

`CallTool` now routes a confined key through the allowlist of #72 instead of the schema rule. The interceptor first forces the gates the key needs into the arguments (`ConfineToolArgs`), then checks them against the tool's line (`AuthorizeToolArgs`), and hands the tool the arguments it checked. The five ways out you found on #72 are held over gRPC in `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC`: consolidate refused; a user key's recall reaches no knowledge and a collection key's recall no memory, even with the gate set off in the call; `cortex_query` refused; `extract_conversation` refused a `session_id` and `persist`, allowed text; `supersedes` of another user's memory and of a missing one both `NOT_FOUND`, and the memory not retired; chunk ids on entities refused; a search pinned to another collection's document `NOT_FOUND`. Rebased on v2.134.1.

### Problem
A key with a scope is refused every `CallTool`, because the request carries the tool's arguments as a string of JSON the interceptor cannot see into. The MCP stdio client in remote mode only speaks `CallTool`, so a confined key cannot be handed to a developer's machine at all.

### Fix
`CallTool` routes a confined key through the policy's tool rules instead of refusing it.
- The interceptor sets the gates the key needs (`authz.Key.ConfineToolArgs`), checks the arguments against the named tool's allowlist line (`authz.Key.AuthorizeToolArgs`): tool on the list, every argument on the line, confined field set and equal in every copy and spelling, gate on. It then replaces the request's arguments with the ones it checked. Arguments that are not a JSON object are a denial.
- The handler then applies the row rule (`authz.Scope.GuardToolRows`) where the database is at hand, as every id-addressed typed RPC does: a row named by id that lies outside the scope is `NOT_FOUND`, worded like a row that does not exist.
- The typed RPCs' own row checks (`scopePermitsMemory`, `scopePermitsKnowledge`) call the rules the policy exports, so the two paths cannot disagree.

Unconfined keys, the typed RPCs, `ListTools` and the read-only clearance are unchanged. A tool off the allowlist (`graph_statistics`) stays refused to a confined key.

### Tests
- `pkg/rpcserver/confined_tools_review_test.go`: the regression test above, with an operator key writing the rows the confined keys must not reach.
- `pkg/rpcserver/tools_auth_test.go`: a key confined to a `user_id` searches its own memory, is refused another user's, is refused a search with no `user_id`, and is refused a tool off the list even when it passes one.
- `pkg/rpcserver/scoped_keys_test.go`: a key confined to one collection ingests and searches it through the tools; a key confined to two collections searches each through `CallTool` and is refused a third, as the typed RPC refuses it; a key confined to `team-acme` passing `collection=team-acme` with the id of a document in `secrets` gets `NOT_FOUND`, and the document is unchanged; a new row of its own passes.
- `pkg/rpcserver/tools_guard_test.go`: the two halves of the check fail closed on their own, without the interceptor in front: a message that is not a `CallToolRequest`, an unknown tool, non-object arguments and a read that fails are refused or reported, never passed; an unconfined caller is not checked.
- `go vet` and `go test -race ./...` pass.

**Failing first.** The tests use the new functions, so the proof is per wiring rule: disable it, keep the package compiling, and a test must fail. Three mutations, three killed: the gate not forced before the call, the checked arguments not handed to the tool, each fail `TestTheWaysOutOfAConfinementFoundInReviewAreClosedOverGRPC`; the allowlist not consulted fails it and four tests of `scoped_keys_test.go`. Every line of the patch has a test.

Part of a small series on running one CortexDB server for several teams (see discussion #66, item GC-04).

## Type of Change
- [x] New feature (non-breaking change which adds functionality)

## Testing
- [x] Unit tests pass
- [x] Integration tests pass
- [x] New tests have been added

## Performance Impact
For a confined key only: one JSON decode and one encode of the arguments per tool call, and one indexed read per id argument. Nothing for an unconfined key.

## Breaking Changes
None. A confined key that was refused every tool is now allowed the tools on the allowlist, within its scope; every other answer is unchanged.
