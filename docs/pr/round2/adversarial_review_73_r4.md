# Relecture adverse de #73, tour 4 partiel (22 octobre 2026, 00 h 05)

Tête `4e589f4` (`gc-04b-grpc-confined-tools`), base `d805527`. Tour partiel accepté par Lauro (« non c'est bon tu peux faire un partiel ») : seulement la différence `4f71850..4e589f4` (B1 sorti vers `gc-12-grpc-ingest-guard`, deux tests pour L3 et L4 du tour 3, textes). Agent `upstream-reviewer`, kit `adv-73-r4` avec `interdiff.diff`, invite réduite. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 18, contexte 18664 -> 65187 tokens ; écriture cache 65185, lecture cache 778971, entrée neuve 36 ; équivalent entrée (lecture x0,1, écriture x1,25) : 159378 ; outils : {'Bash': 18, 'SubagentHandback': 1}. Durée 2 min 10 s.

---

Review of #73, round 4 (only the change from 4f71850 to 4e589f4)

Verdict: no blocker. The guard came out cleanly. Every claim I checked holds on 4e589f4. All four of my mutations were killed. I found two nits of wording.

**1. Removing the guard (I read interdiff.diff in full)**
- `pkg/rpcserver/graphrag.go` is back to main's version. It is not in `stat.txt`, so #73 no longer touches it, and it cannot conflict with gc-12.
- `guardKnowledgeUpsert` is still called by SaveKnowledge and UpdateKnowledge (`knowledge.go:41`, `knowledge.go:66`). Its comment (`ownership.go:424`) and the `withheldIfMissing` comment (`ownership.go:319-323`) no longer mention InsertGraphDocument.
- The code comment at `ownership.go:140` and the tests in `confined_typed_writers_test.go:21,82` and `confined_typed_args_test.go:83,115` refer only to the argument check, which is still on the head (`ownership.go:189`). Nothing that remains depends on the removed guard.
- Round 3 N6 (the weak InsertGraphDocument chunkless case) is moot, because the case is gone.
- `both.sh -race ./pkg/rpcserver/` passes on both trees: main ok in 28.6s, head ok in 27.7s.

**2. Claims against 4e589f4**
- Nothing still claims that this PR guards InsertGraphDocument:
  - The body has a paragraph "One typed write is fixed in a separate PR".
  - The reply names #NN.
  - Breaking Changes lists InsertGraphDocument only for the argument check, which is true at `ownership.go:189`.
  - The commit message does not mention a row guard.
- The body says "where `ingest_document` through `CallTool` is guarded here". That holds: `allowlist.go:181` has `document_id` as an idKnowledge field, and `confined_tools_round2_test.go:73` tests it.
- Mutation counts: the eighth-pass table in failing_first.md has 36 rows (lines 301-336 of the file). It has no B1 row, and the new rows V, X and Y are there. The body says "Thirty-six … thirty-six killed", the reply says "Thirty-six", and the table header says "36 sur 36". They agree.
- The other round 3 points:
  - L1 (blank ids) is named in the body ("Blank ids", #82) and in the reply. It is not fixed here, which is acceptable as long as #82 lands first or alongside.
  - L3 is closed by the new block at `confined_typed_writers_test.go:127-139`. A confined empty update must keep the operator's code and message, and must not be NOT_FOUND.
  - L4 is closed by `TestTheTypedMemoryRowChecksHoldEveryMemoryField` (`confined_typed_writers_test.go:142-186`). The `lab` and `mine` keys run against stored rows through GetMemory and UpdateMemory.
  - N2 is closed: Performance Impact now says chunks are loaded and that a create reads twice.
  - N3 is closed: the subject is 62 columns.
  - N4 is closed: "New feature" is no longer labelled "non-breaking".
  - N5 is closed: the heading now reads "What this PR leaves as it is".

**3. My mutations** (`mutone.py`, `-run` limited to the two new tests)
- A. `withheldIfMissing` rewrites every code of a confined caller (`st != nil` instead of `== NotFound`): killed by TestATypedUpdateKnowledgeAnswersAMissingIdAsAForeignOne.
- B. `withheldIfMissing` keeps the code but rewrites the message of any other refusal to "refused": killed by the same test.
- C. `scopePermitsMemory` lets any `Scope == "global"` row through: killed by TestTheTypedMemoryRowChecksHoldEveryMemoryField.
- D. The UpdateMemory row guard is ignored (`memory.go:71`, `&& false`): killed by the same test.

There were no survivors.

**Findings**
- nit, commit message, last paragraph: "is withheld from a confined key on the typed RPCs as on CallTool, the typed upserts included". InsertGraphDocument is a typed upsert, and it is not withheld for a chunkless document on this head. Suggest "SaveKnowledge and UpdateKnowledge included", as the body already says.
- nit, `review_replies.md` section #73: "A document whose chunks are gone is withheld on the typed RPCs too, … writes included." This is the same over-reach. It is softened by the next sentence, which names #NN, but it reads better as "SaveKnowledge and UpdateKnowledge included".
- Not a finding, only a note: until #NN is opened, the body and the reply still say "#NN", and the body promises to add the InsertGraphDocument chunkless case to whichever of the two PRs lands second. Fill in the number before posting.
