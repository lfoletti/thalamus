The six PRs sent on 19 October 2026, all on top of v2.134.0:

| PR | Change | Own diff | Stacked on |
| :- | :- | :- | :- |
| #72 | authz rules that let a confined key call a tool, not yet wired (`pkg/authz` only) | +677 / -19 | main |
| #73 | gRPC wiring: `CallTool` checks the arguments, then the rows an id names | +286 / -30 own, +963 / -49 shown | #72 |
| #74 | REST wiring: `POST /v1/tools/<name>` under the same rules | +126 / -6 own, +803 / -25 shown | #72 |
| #75 | `ingest_document` hashes content, skips an identical re-ingest, `dedupe` refuses a copy | +297 / -6 | main |
| #76 | `expires_at` on documents, `--sweep-expired` | +643 / -3 own, +940 / -9 shown | #75 |
| #77 | `review_status`, `reviewed_by`, `reviewed_at`; hiding is opt-in | +706 / -59 own, +1613 / -35 shown | #76 |

Notes:
- #73 and #74 are independent of each other. #76 touches the same lines of `ingest_document` as #75, and #77 shares the search gate with #76, hence the stacks. After each merge I rebase the branches above it.
- Every PR has a "Failing first" section: tests applied on the base tree fail, or, where the tests do not compile without the new code, each rule was disabled by a mutation and the named test fails. Codecov: what stays uncovered is named in each body. #74, the 403 arm of the REST row guard (3 lines, defensive, refused earlier by the argument check). #75, error returns a failing row scan would take. #76, the same, plus a delete that fails mid-sweep, the default brain path and `main`. #77, the same plus `main` of `cortexdb-grpc`.
- Full `go test -race ./...` and the lexical eval pass on each head.

One more thing: tonight's updates were pushed several times as Codecov's reports came in, which was short on preparation on my side and may have cost you notifications and CI runs. Sorry for the noise. From the next batch on, patch coverage and mutation checks run locally before anything is pushed.
