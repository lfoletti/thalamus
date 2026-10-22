Batch B is updated for your second review, rebased on v2.135.0, and back in review: #72 (authz rules), #73 (gRPC wiring), #74 (REST wiring). Each body has a "Changes since the second review" section, and each PR has a reply under your review.

Two holes of `main` turned up while checking the classes. They do not depend on the wiring, so each has its own small PR, based on `main`:

- #82: `POST /v1/knowledge` and `POST /v1/memory` did not check the row an upsert id names. A confined key could overwrite another scope's row and move it into its own scope. The same PR makes the upsert guards of both ports skip only an empty id, not an id of blanks.
- #83: over gRPC, `InsertGraphDocument` had no row guard. A key confined to one collection could overwrite a document of another collection.

#82 and #74 touch the same lines of the REST policy, and #83 and #73 the same guard. Whichever lands second, I rebase it and add the test cases that need both.

Batch C (#75 to #77) comes next.
