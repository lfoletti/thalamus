# Relecture adverse de `gc-12-grpc-ingest-guard`, tour 1 (22 octobre 2026, 00 h 08)

Tête `e68061d`, base `cc98c6c`. Agent `upstream-reviewer`, kit `adv-gc12-r1`, invite fixe de `docs/pipeline_pr.md`, section 1 bis. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 34, contexte 19130 -> 97253 tokens ; écriture cache 97251, lecture cache 2143736, entrée neuve 68 ; équivalent entrée (lecture x0,1, écriture x1,25) : 335937 ; outils : {'Bash': 34, 'SubagentHandback': 1}. Durée 7 min 22 s.

---

## Review of gc-12-grpc-ingest-guard (e68061d, base cc98c6c)

**Verdict: approve once two comments are dealt with. No blocker.** The fix closes B1 for ordinary ids. Two things in the PR are still open: whitespace ids still get through, and one sentence of the body is wrong.

I worked only in the kit. The probe file was removed from main, head and the mut copy at the end. The full `pkg/rpcserver` suite under `-race` passes on both trees (main 17.9 s, head 23.5 s).

### Class, taken from the code
I parsed the protos. Request messages that carry a top-level `collection` or `user_id`:
- `InsertGraphDocument`, `SearchGraphRag`, `SearchText`
- `SaveKnowledge`, `UpdateKnowledge`, `SearchKnowledge`
- `SaveMemory`, `SearchMemory`

The handlers take over the row check for the five methods in `idAddressedMethods` (`pkg/rpcserver/ownership.go:73-79`). `Rowless` is set only on `Health`, `Info` and `ListTools`, which are all reads. `CallTool` is refused outright (`pkg/authz/methods.go:239`).

Writes a confined key can reach that name a row by id: `InsertGraphDocument`, `SaveKnowledge`, `UpdateKnowledge`, `SaveMemory`, `UpdateMemory`, `DeleteMemory`, `DeleteKnowledge`. Every one is now guarded, so the body's list of members is right.

Inside `InsertGraphDocument`, the chunk, document-node and edge ids all derive from `doc.ID`. I checked the formats (`chunk:%s:%03d`, `doc:%s`, `edge:doc_chunk:%s:%d`) and found no way to collide with another document's ids. Entity nodes are shared by name. That is true on main too, and they are merged on purpose (`mergeEntityCollections`), so it is out of scope.

### Findings

**L1 (likely comment): a whitespace id still overwrites and moves a document of another collection.** `pkg/rpcserver/graphrag.go:35` → `ownership.go:255`.
- `guardKnowledgeUpsert` returns nil when `strings.TrimSpace(id) == ""`.
- The facade only rejects `doc.ID == ""` (`pkg/cortexdb/graphrag.go:169`).
- Probe: the operator saves `" "` in `secrets`, then the `notes` key calls `InsertGraphDocument{Id:" ", Collection:"notes"}`. The answer is OK on head, and the document reads back as `notes/Blank overwritten.` Same result with `"\t"`.
- So the body's sentence "A document outside the key's scope is NOT_FOUND … and is left as it was" is false for blank ids while this PR stands alone.
- The body does mention #82 ("an empty id only, not blanks"). I checked #82 at 9e3ba10: `ownership.go:255` there is `if id == "" {`, so it does close this.
- But the body never says the hole remains until #82 lands, and only promises a follow-up test for #73.
- My mutation M2 (`TrimSpace(id) == ""` → `id == ""`) **survived** the whole package. Nothing here tests blank ids in either direction.
- Fix: either take the one-line change from #82 here, or state the residual hole plainly. In both cases add a blank-id case to whichever of #82 and this PR lands second.

**L2 (likely comment): the Scope section misstates the class.**
- The body says: "Every other RPC carries no scope field, so AuthorizeRows refuses it to a confined key."
- That is wrong for `SearchGraphRag`, `SearchText`, `SearchKnowledge` and `SearchMemory`. Each declares `collection` or `user_id` and is reachable by a matching key.
- They are outside the class because they are searches, not writes named by id.
- The conclusion holds; the stated reason does not.

**N1 (nit): the Performance claim is imprecise.**
- The body says "one `GetKnowledge` per `InsertGraphDocument` that names an id".
- For a create (an id that names nothing), `guardKnowledgeUpsert` calls `GetKnowledge` twice: once inside `guardKnowledge` (`ownership.go:204`), then again at `ownership.go:263`.
- Each call goes through `loadKnowledgeRecord`, which itself does several store reads.

**N2 (nit): weak assert on the create path.** In `pkg/rpcserver/ingest_ownership_test.go`, the loop over "mine" (about lines 44-48) checks only `err == nil`. It never reads the document back to check it landed in `notes`, so the claim "`collection` already confines what it creates" has no assert here. It is held by `AuthorizeRows`; probe S11 gives `PermissionDenied` for `collection=secrets`.

**N3 (nit): empty content over a foreign id is untested.** On head the answer is `NotFound`, on main `InvalidArgument empty text`. This matches `SaveKnowledge`, but my mutation M10 (skip the guard when content is empty) survived.

**Already named in the body, confirmed:** a key confined to `graphrag_chunks` still overwrites a chunkless document of another collection (S3), on both trees. The body says this comes with #73. Leave it to whichever of the two lands second, as promised.

### Mutations I wrote (7; 4 killed, 3 survived)
| # | Mutation | Result |
| :- | :- | :- |
| M1 | `graphrag.go`: the guard's refusal is replaced by `toStatus(core.ErrNotFound)` (wording claim) | killed: TestInsertGraphDocumentCannotTakeADocumentOfAnotherCollection |
| M2 | `ownership.go:255`: `TrimSpace(id) == ""` → `id == ""` | **survived** (L1) |
| M3 | `ownership.go:263`: `gerr != nil` → `gerr == nil` | killed: the new test plus 4 others |
| M4 | `ownership.go:201`: `if !confined` → `if !confined && false` | **survived**, equivalent (a zero scope permits every row). "Unconfined keys are not checked" is a cost claim with no behavior to test. |
| M5 | `ownership.go:212`: the scope check also refuses the key's own existing documents | killed: the new test plus 2 others |
| M6 | `graphrag.go`: the guard moved after the facade write, called on a mangled id | killed: the new test |
| M10 | `graphrag.go`: guard skipped when content is empty | **survived** (N3) |

### Scenarios, the same input on both trees (keys: `notes` = `collection=notes`, `dflt` = `collection=graphrag_chunks`, `hermes` = `user_id=hermes`, `ro` = read-only `notes`)
| Scenario | main | head |
| :- | :- | :- |
| S1 notes IGD `id=sec` (secrets doc) | OK, doc becomes `notes/Overwritten by notes.` | NotFound `knowledge "sec" not found`, doc unchanged |
| S2 notes IGD `id=" "` (secrets doc) | OK, moved to notes | **OK, moved to notes** (L1) |
| S2b notes IGD `id="\t"` (secrets doc) | OK, moved | **OK, moved** (L1) |
| S3 dflt IGD over chunkless secrets doc | OK, overwritten | OK, overwritten (named, #73) |
| S4/S4b notes IGD new id, then same id | OK / OK | OK / OK |
| S5 notes IGD foreign id, empty content | InvalidArgument `empty text provided` | NotFound `knowledge "sec" not found` (named rule; untested, N3) |
| S5b notes IGD new id, empty content | InvalidArgument | InvalidArgument |
| S6 hermes IGD | PermissionDenied (no user_id field) | same |
| S7 read-only notes IGD | PermissionDenied (read-only) | same |
| S8 notes IGD empty id | Internal `graph document ID cannot be empty` | same (already on main; InvalidArgument would be the right code) |
| S9 notes IGD over operator's doc in `notes` | OK, rewritten | OK, rewritten |
| S10 operator IGD over secrets doc | OK | OK |
| S11 notes IGD `collection=secrets` | PermissionDenied | same |
| S12 notes SaveKnowledge over `sec` | NotFound `knowledge "sec" not found` | same (the wording matches IGD on head) |

The difference between the trees that the body names is S1, plus S5 as a consequence. The difference it does not state is the absence of one: S2 and S2b are unchanged.

### Test hygiene
- `ingest_ownership_test.go` has no `Sleep`, `Getenv`, `Setenv`, `LookupEnv` or `time.` calls.
- `fakeEmbedder` (`harness_test.go:18`) is deterministic, built on SHA-256.
- The test connects over bufconn and opens its database under `t.TempDir()`, so it reaches no server and does not read the machine's environment.
- Adding the variadic `dbOpts` to `newPolicyConn` broke none of its 33 call sites; the package builds and passes.
- The author's failing-first claim fits what I saw: on main, the same input (S1) overwrites and moves the document, which is exactly the two failures the body quotes.

### Before approval
- L1: take #82's change to `ownership.go:255` here, or state the residual hole and add a blank-id test case to whichever PR lands second.
- L2: fix the Scope sentence about which RPCs carry no scope field.
