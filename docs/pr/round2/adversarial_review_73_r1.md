# Relecture adverse de #73, tour 1 (21 octobre 2026, 22 h 51)

Tête `ff24c70` (`gc-04b-grpc-confined-tools`), base `d805527`. Agent `upstream-reviewer`, kit `adv-73-r1`, invite fixe de `docs/pipeline_pr.md`, section 1 bis. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 32, contexte 19177 -> 119946 tokens ; écriture cache 119944, lecture cache 2291368, entrée neuve 64 ; équivalent entrée (lecture x0,1, écriture x1,25) : 379066 ; outils : {'Bash': 31, 'SubagentHandback': 1}. Durée 7 min 31 s.

---

## Review of #73 (gc-04b-grpc-confined-tools, ff24c70, on #72 at d805527), round 3

Read-only. I ran Go only through the kit or under the canary environment. I wrote one probe file, `zz_probe73_test.go`, into both copies and removed it afterwards (the kit-root copy too). Outside test files the diff is +155/-11, in `pkg/rpcserver/ownership.go` and `pkg/rpcserver/tools.go`. `go test -race` of `pkg/rpcserver`, `pkg/authz` and `pkg/httpapi` is green on both trees.

### Class check (from code, not from the inventory)

- **Typed RPCs a confined key can reach.** `AuthorizeRows` lets a confined key through only when the request carries the confined field, or when the RPC is in `idAddressedMethods`. From the protos and `methods.go`, the typed writes in that class are SaveKnowledge, UpdateKnowledge, InsertGraphDocument, SaveMemory, UpdateMemory, DeleteKnowledge and DeleteMemory. InsertText and InsertTextBatch carry no collection, and the graph, decision and admin writes carry no scope field, so all of those are refused outright. The five covered by `authorizeTypedArgs` are therefore complete for writes that carry arguments. The typed searches carry no path-retrieval arguments, so there is no mismatch with `pathDenied` there.
- **Row-id fields of the allowlist.** I count 11 from `confinedTools`. All 11 appear end to end in `TestTheRoundTwoRulesOfTheAllowlistHoldOverGRPC`, with the four create tools tested on a foreign id only, as the body says.
- **Typed RPCs that answer "foreign or missing".** UpdateKnowledge does not answer alike. See L1.

### Findings

**L1 (likely comment, close to blocker given this round's theme): typed `UpdateKnowledge` tells a missing id from a foreign one, contrary to the body.**
- The body says: "the typed RPCs keep main's wording, `memory "X" not found` (`withheld`), which is the same for a foreign and a missing row."
- Probe, key confined to `notes`, both trees:
  - foreign id: `NotFound "knowledge \"sec\" not found"`
  - missing id: `NotFound "get knowledge: vectorstore: get_document: embedding not found"`
- Cause: `guardKnowledgeUpsert` (`ownership.go:389`) passes a missing id. The facade's own wording then comes back unflattened through `knowledge.go:88`.
- The same call through `CallTool` is flattened by `oneNotFound`, so after this PR the two paths do disagree.
- This is not a regression, but the claim is false. Either name it, or apply the same flattening on the typed upsert path.

**L2 (likely comment): typed `UpdateKnowledge` relations are not tested.**
- Own mutation: `ownership.go:167` `graph(r.GetEntities(), r.GetRelations())` → `graph(r.GetEntities(), nil)`. It **survived** the whole `pkg/rpcserver` suite.
- `TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes` covers relations only on SaveKnowledge, and entities only on UpdateKnowledge.
- The body lists "relations" among what each typed write now refuses.

**L3 (likely comment): the `written_by` rationale is wrong for four of the five typed writes, and the new refusal breaks read-modify-write.**
- Body and commit message: "A confined key that sent … `written_by` … through a typed RPC got them written."
- On main, the facade already replaces `written_by` and drops `updated_by` on SaveKnowledge, UpdateKnowledge, SaveMemory and UpdateMemory (`stampWriter` and `stampUpdater`, `pkg/cortexdb/written_by.go`).
- Probe on main: the `notes` key saves `metadata.written_by=operator`, and the stored row reads `written_by:notes`. So nothing was "written" there.
- Only InsertGraphDocument (ingest) does not stamp. I could not confirm that one on main, because ingest fails without an embedder; I am going by the `writerKeys` comment in `allowlist.go`.
- Head now refuses the field:
  - hermes `UpdateMemory` of **its own** memory with `metadata.written_by` is `OK` on main and `PermissionDenied` on head.
  - A client that reads a row (which carries the engine's `written_by`) and sends the metadata back is now refused.
- Breaking Changes names the refusal but gives a false "Before, it wrote them". Please correct the rationale and say that round-tripped metadata is now refused.
- No test sends `written_by` or `updated_by` through any typed RPC.

**L4 (likely comment): unnamed code change on `UpdateMemory` for a collection-confined key, with a broken message.**
- main: `NotFound "memory \"hermes-1\" not found"`.
- head: `PermissionDenied "... tool memory_update reads memory, which the call must switch off with =true"`.
- The gate name is empty because `memory_update` has no gates (`allowlist.go:344`, formatted with `rule.gates[c.field] == ""`).
- `authorizeTypedArgs` now runs before the `idAddressedMethods` waiver (`ownership.go:92`), so the typed RPC inherits the message.
- GetMemory and DeleteMemory for the same key still answer NOT_FOUND, so the three id-addressed memory RPCs now disagree.
- Not an oracle, because the answer does not depend on the id. But it is not under Breaking Changes, and the message is nonsense. The wording is #72's, and this PR exposes it on a typed RPC.

**L5 (likely comment): "a tool's own refusal that is not NOT_FOUND keeps its code and wording" is not held.**
- Own mutation in `oneNotFound` (`ownership.go:285`): for a confined caller, rewrite any other error to `Internal "refused"`. It **survived**.
- The assertion at `confined_tools_round2_test.go:91-93` only checks that the code is not NotFound or OK and that the message is not "not found". It should compare against the tool's own code and message, for example the operator's answer to the same call.

**N1 (nit): typed SaveMemory scope and namespace are untested on head.**
- Own mutation: `ownership.go:173` `authz.FieldNamespace: r.GetNamespace()` → `""`. It **survived**.
- No test uses a namespace- or scope-confined key on a typed SaveMemory. Such a key would be refused every SaveMemory under this mutant.

**N2 (nit): weak assert on InsertGraphDocument.**
- `confined_typed_args_test.go:110-111` accepts any error other than PermissionDenied, because the test brain has no embedder. It proves only that the policy passed.

**N3 (nit): the first check in `authorizeTypedArgs` and `authorizeToolArgs` is redundant.**
- The `key.Scope.IsZero()` early return in both functions duplicates the checks inside `AuthorizeToolArgs` and `ConfineToolArgs`, so removing it would be an equivalent mutant.
- Likewise, after this PR `AuthorizeRows` on UpdateKnowledge is subsumed by the `knowledge_update` collection check. An own mutation that skips it for UpdateKnowledge survives as an equivalent mutant.
- Worth a line in the comment so nobody reads these as the guard.

**N4 (nit): the PermitsMemory and PermitsKnowledge rule is still written out twice.**
- "So the two paths cannot disagree" holds for `scopePermits*` now. `authz.Key.CallerContext` (`pkg/authz/caller.go`, not in this diff) still spells both rules out separately, and has no chunkless rule.

Tests that sleep or read the environment: none in the new or changed test files (I grepped for `Sleep`, `Getenv`, `Setenv`, `LookupEnv` and `UserHomeDir`).

### Own mutations (`mutone.py`, `-run .` on `./pkg/rpcserver/`)

| # | Mutation | Result |
|---|---|---|
| 1 | UpdateKnowledge relations dropped from the typed check (`ownership.go:167`) | **survived** |
| 2 | `guardToolRows`: `ErrDenied` mapped to NotFound | killed (`TestTheCallToolChecksFailClosedOnTheirOwn`) |
| 3 | `guardToolRows`: a read failure returns nil | killed (same test) |
| 4 | `authorizeToolArgs`: a message that is not a CallToolRequest returns nil | killed (same test) |
| 5 | SaveMemory typed check drops namespace (`ownership.go:173`) | **survived** |
| 6 | `oneNotFound`: any other error from a confined call becomes `Internal "refused"` | **survived** |
| 7 | AuthorizeRows skipped for UpdateKnowledge | survived, equivalent (N3) |

That is 7 mutations: 3 killed, 3 real survivors, 1 equivalent.

### Scenarios on main and head (same probe on both)

| Scenario | main | head |
|---|---|---|
| `notes` SaveKnowledge `metadata.written_by=operator` | OK, stored `written_by:notes` | PermissionDenied |
| `notes` SaveKnowledge `_review_status`, `updated_by` | OK, `_review_status:accepted` stored | PermissionDenied |
| hermes SaveMemory `written_by` | OK | PermissionDenied |
| hermes UpdateMemory own row, `written_by` | OK | PermissionDenied (unnamed in effect, L3) |
| hermes UpdateMemory foreign / missing | NotFound `memory "id" not found` (both) | same |
| `notes` (collection key) UpdateMemory | NotFound | PermissionDenied, "with =true" (L4) |
| `notes` GetMemory | NotFound | NotFound |
| `notes` UpdateKnowledge, no collection | PermissionDenied | PermissionDenied (reworded) |
| `notes` UpdateKnowledge foreign | NotFound `knowledge "sec" not found` | same |
| `notes` UpdateKnowledge missing | NotFound `...embedding not found` | same (L1) |
| `notes` SaveKnowledge foreign id | NotFound | NotFound |
| `notes` SaveKnowledge entities | OK | PermissionDenied (named) |
| `notes` CallTool `knowledge_search` own collection | PermissionDenied | OK (named) |
| read-only confined `knowledge_save` / `knowledge_search` | denied / denied | denied / OK |
| hermes `memory_get` foreign / missing / `Memory_ID` foreign | denied (opaque JSON) | NotFound `"not found"` (all three) |
| hermes `memory_get` empty id / no args | denied | InvalidArgument `memory_id is required` |
| hermes `memory_delete` missing, `memory_update` missing | denied | NotFound `"not found"` |
| `notes` `knowledge_get` missing | denied | NotFound `"not found"` |
| hermes recall with `disable_knowledge:false` | denied | OK (gate forced) |
| `notes` `search_text` collection=secrets | denied | PermissionDenied (asks for "secrets") |
| `notes` `memory_search` | denied | PermissionDenied (no collection argument) |
| `notes` `graph_statistics` | denied | PermissionDenied (off the allowlist) |
| `notes` args `[]` | denied | PermissionDenied (not an object) |
| unknown tool | PermissionDenied | PermissionDenied |

### Verdict

No blocker found in the wiring itself. The gate is forced and then checked, the checked arguments reach the tool, the row guard holds all 11 fields, and confined NOT_FOUNDs are flattened. Before approval I would ask for:
- L1: name the typed UpdateKnowledge oracle, or flatten it.
- L2 and L5: tighten the two tests.
- L3: fix the `written_by` rationale and name the round-trip break.
- L4: name the UpdateMemory change and fix the empty gate in the message (in #72).
