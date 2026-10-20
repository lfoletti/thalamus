Thank you for the review and the merges. Next batches and requests are in:

- Batch B is three PRs, as asked: authz rules unplugged (#72), gRPC wiring (#73), REST wiring (#74). Each carries its failing-first tests, with a "Failing first" section in the body that names, for each rule, the test that fails without it.
- Batch C is three PRs as well: content hash (#75), `expires_at` (#76), `review_status` (#77). Hiding unreviewed rows is opt-in: `CORTEXDB_REVIEW_GATE=1` on the server, or `cortexdb.WithReviewGate()`. A brain that never sets the field searches exactly as today, and a test holds that.

One question on size. Some of these PRs are long: the own change of #72 and of #77 is about 700 lines each, half of it tests, and #76 and #77 show about 900 and 1600 lines because they are stacked on the PR below them. I am concerned about readability. Would you prefer smaller PRs? If so, an approximate line limit would help me cut the next ones to the right size. I can also split the open ones, for example tests in a separate commit, or one rule per PR.
