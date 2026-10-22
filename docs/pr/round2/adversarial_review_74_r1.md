# Relecture adverse de #74, tour 1 (22 octobre 2026, 00 h 46)

Tête `bc2175d` (`gc-04c-rest-confined-tools`), base `d805527` (#72). Agent `upstream-reviewer`, kit `adv-74-r1`, invite fixe de `docs/pipeline_pr.md`, section 1 bis. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 30, contexte 19317 -> 127245 tokens ; écriture cache 127243, lecture cache 2353184, entrée neuve 60 ; équivalent entrée (lecture x0,1, écriture x1,25) : 394372 ; outils : {'Bash': 29, 'SubagentHandback': 1}. Durée 10 min 6 s.

---

I found one blocker on #74 at bc2175d (base d805527). A confined key can still declare entities and relations through the typed routes by spelling the field with U+017F (the long s, `ſ`). That breaks the main claim of this round.

## Review of #74 (gc-04c-rest-confined-tools, bc2175d, on #72 at d805527)

**Verdict: one blocker, two likely comments, four nits.** The tool-route wiring holds on every probe I ran. The typed-write rule added this round does not hold against case folding.

### What I ran
- One probe file, put in both trees and removed afterwards. It covers 39 requests over three keys: operator, `hermes` (user_id) and `notes` (collection). A fourth key, `both` (collection + user_id), was used in the first pass.
- Eight mutations of my own with mutone.py, all with `-run .` on `./pkg/httpapi/`.
- `both.sh -race ./pkg/httpapi/ ./pkg/authz/`: green on both trees.

### Class check (from the code)
- **Fields of a typed write body that `encoding/json` maps to `Entities`/`Relations`.** The decoder matches keys by Unicode simple folding (`encoding/json/fold.go`, `foldRune` via `unicode.SimpleFold`), so `entitieſ` lands in `Entities` and `relationſ` in `Relations`. `typedToolArgs` picks keys with `strings.ToLower(k)` (`policy.go:229`), which leaves `ſ` as it is, so those keys are dropped before the allowlist sees them.
  - The tool route is safe: `checkPaths` refuses an unknown key (S5).
  - The scope fields are safe: `otherSpellings` and `scopeValues` use `strings.EqualFold`, which folds `ſ` too.
  - `metadata` and `collection` have no `s` or `k`, so they cannot be spelled this way.
- **Row-id fields of the allowlist.** I count 11 from `allowlist.go:181-225`. All 11 are exercised end to end in `TestTheRoundTwoRulesOfTheAllowlistHoldOverREST`.
- **Typed REST writes a confined key can reach.** `routeAccess` has seven POST write routes. The three graph routes (`/v1/graph/entities`, `/v1/graph/relations`, `/v1/graph/triples`) carry no scope field and stay 403, as tested. That leaves `POST /v1/knowledge` and `POST /v1/memory`, both in `typedWriteTools`. The `reaches` predicate matches what the row checks let through: S8 and S9 give the same 403 on both trees for a key confined to collection and user_id.
- **404s a tool returns to a confined key.** One path skips `oneNotFound`: the "unknown tool" substring branch (L1).

### Findings

**B1 (blocker): a confined key writes shared-graph entities and relations through `POST /v1/knowledge` and `POST /v1/memory` by spelling the field with U+017F.**
- `typedToolArgs`, `pkg/httpapi/policy.go:229`: `switch strings.ToLower(k)` matches neither `entitieſ` nor `relationſ`, so the allowlist never sees them. The handler's decoder folds them into `Entities` and `Relations`.
- Scenario T1, notes key, both trees, 200:
  `{"knowledge_id":"n2","collection":"notes","content":"x","entitieſ":[{"name":"Nightshade","type":"project","description":"DEFACED"},{"name":"Orpheus","type":"person"}],"relationſ":[{"from":"Nightshade","to":"Orpheus","type":"funds_by_notes"}]}`
  The response has `entity_node_ids:["entity:nightshade","entity:orpheus"]` and a `relation_edge_ids` entry. Nightshade had been created by the operator in `secrets`. Afterwards the operator's SPARQL finds `entity:nightshade` with literal `DEFACED` (T3) and the edge `rel:funds_by_notes` (T4).
- Same on `/v1/memory` with the hermes key: 200 on both trees, description `defaced by memory` written (U3/U4).
- The ASCII spelling is refused on head (S1, 403), so the hole is spelling-only.
- This contradicts the commit message ("a confined key cannot write through a typed route what the tool route refuses (declared entities and relations …)"), the body's Breaking Changes, and "as on gRPC". gRPC has no such hole because proto field names are exact.
- Mutation M1 (`switch strings.ToLower(k)` → `switch k`) **survives**: no test sends an outer key in another spelling on a typed route.
- Suggested fix: match with `strings.EqualFold` against each name, or better, decode the body into the facade request struct and hand the decoded `Entities`/`Relations`/`Metadata` to the allowlist. Then the check reads what the handler will read.

**L1 (likely comment): "every 404 a tool returns is spelled `not found`" has an exception that is an oracle for some ids.**
- `tools.go:60` writes `err.Error()` unchanged whenever the error text contains "unknown tool", before `oneNotFound` runs.
- The tool name has already been checked by the policy at this point, so the branch now fires on any error whose text carries a caller's id.
- Head scenario, hermes key, `memory_save` with `supersedes`:
  - foreign `["unknown tool z"]` → `404 {"error":"not found"}` (U1)
  - missing `["unknown tool q"]` → `404 {"error":"supersede \"unknown tool q\": embedding not found"}` (U2)
- It is narrow, since the id must contain the phrase, but it is the exact distinction D6 closes. Pass this branch through `oneNotFound` too, or match the toolbox's not-found sentinel instead of its text.

**L2 (likely comment): the body contradicts itself.**
- The Fix section still says "Unconfined keys, the typed routes and the read-only clearance are unchanged", while the second-round section and Breaking Changes say the typed routes now refuse entities, relations and `_` keys.
- The older "Failing first" paragraph says "Seven mutations, seven killed" next to "Twenty-one".
- The Tests list leaves out `confined_tools_round2_test.go`, `confined_typed_args_test.go` and `confined_typed_writers_test.go`.

**N1 (nit, from #72, now visible over REST):** `memory_update` from a collection key answers `… tool memory_update reads memory, which the call must switch off with =true` (S19). The gate name is empty because the tool has no gate for that field. It should say the tool reaches no row of the key's store.

**N2 (nit, weak assert):** the db-closed case of `TestTheToolRouteFailsClosedForAConfinedKey` (`scoped_keys_test.go`, the `knowledge_update` → 500) would also pass if the guard let the call through, because the tool itself fails 500 on a closed database. Its comment promises "never as a pass", which it does not show. Mutation M4 (500 mapped to 404) is killed, so only the "pass" half is unheld.

**N3 (nit):** `TestARowGuardFailureHasItsOwnStatus` unit-tests an `ErrDenied` → 403 branch that `GuardToolRows` reaches only for a tool off the allowlist. `AuthorizeToolArgs` refuses such a tool first, so the branch is unreachable from the route. Its `ErrWithheld` case also uses the old wording (`knowledge "n-1" not found`) that the guard no longer produces.

**N4 (nit, equivalent mutants):**
- M2: `r.ContentLength = int64(len(body))` removed (`policy.go:162`) survives. Nothing downstream reads it; keep it, but note it is not tested.
- M5: running the confined block for unconfined keys too (`policy.go:128`, `&& !key.Scope.IsZero()` removed) survives. It is equivalent for well-formed bodies, but an operator's non-object body would then get the policy's 400 wording instead of the handler's. Nothing pins the operator's wording on that path (U5 sends broken JSON, which still reaches the handler's `unexpected EOF`).

### My mutations (mutone.py, `-run .`, `./pkg/httpapi/`)
| # | Mutation | Result |
| :- | :- | :- |
| M1 | `policy.go:229` `switch strings.ToLower(k)` → `switch k` | **survived** (B1) |
| M2 | `policy.go:162` ContentLength line → `_ = len(body)` | survived (equivalent, N4) |
| M3 | `toolGuardStatus` ErrDenied → 404 | killed: TestARowGuardFailureHasItsOwnStatus |
| M4 | `toolGuardStatus` default 500 → 404 | killed: TestTheToolRouteFailsClosedForAConfinedKey, TestARowGuardFailureHasItsOwnStatus |
| M5 | `policy.go:128` drop `&& !key.Scope.IsZero()` | survived (equivalent in effect, N4) |
| M6 | `!(isJSONObject(body) && json.Valid(body))` → `!(json.Valid(body))` | killed: TestTheRoundTwoRulesOfTheAllowlistHoldOverREST, TestAConfinedKeyMayCallToolsWithinItsScopeOverHTTP |
| M7 | `return` removed after a failed `bufferBody` in the tool block | killed: TestTheToolRouteFailsClosedForAConfinedKey |
| M8 | allowlist refusal on the tool route 403 → 400 | killed: four tests (WaysOut, RoundTwo, Signature, MayCallTools) |

That is 8 mutations: 5 killed, 3 survived, of which 1 is real (M1) and 2 are equivalent (M2, M5).

### Scenarios, same probe on both trees
| Scenario | main | head |
| :- | :- | :- |
| S1 typed knowledge, `entities` (ASCII), notes | 200, written | 403 allowlist |
| T1 typed knowledge, `entitieſ` + `relationſ`, notes | 200, entity defaced, edge added | **200, same** (B1) |
| S3 typed knowledge, `relationſ` alone | 500 FK (write attempted) | 500 FK (write attempted) |
| U3 typed memory, `entitieſ`, hermes | 200, entity defaced | **200, same** (B1) |
| S5 tool `knowledge_save`, `entitieſ` | 403 (confined refused) | 403 `takes no entitieſ argument` |
| S7 typed knowledge, `metadata` + `Metadata{_review_status}` | 200, `_review_status` stored | 403 |
| S8 typed knowledge, collection+user_id key | 403 user_id unset | 403 same |
| S9 typed memory, collection+user_id key | 403 collection unset | 403 same |
| S10 typed knowledge, body `null` | 403 collection unset | 403 same |
| S13/S14 tool, empty body, gated or ungated | 403 confined | 403 collection unset |
| S15 tool, trailing second object | 403 confined | 400 JSON object |
| S16 typed `entities` as an object | 400 decode | 400 decode |
| S17 typed `metadata.written_by` | 200, facade signs | 200, facade signs |
| S18 tool `knowledge_delete` missing | 403 confined | 404 `not found` |
| S19 tool `memory_update`, collection key | 403 confined | 403 `… with =true` (N1) |
| S20 typed memory, supersedes foreign | 404 `supersede … embedding not found` | same |
| T5–T8 tool get/update/delete, missing id "unknown tool 1" | 403 confined | 404 `not found` |
| T9–T11 operator, missing ids | 404 store wording | same |
| T13 tool `memory_save`, foreign `session_id` | 403 confined | 200 (stored in hermes's default bucket) |
| U1 tool supersedes foreign "unknown tool z" | 403 confined | 404 `not found` |
| U2 tool supersedes missing "unknown tool q" | 403 confined | **404 raw store wording** (L1) |
| U5 operator, broken JSON on the tool route | 400 `unexpected EOF` | same |

Every head/main difference except the B1 and L1 rows is named in the body.

### Tests that sleep, read the environment, or pass without the code
- No added line calls `Sleep`, `Getenv`, `Setenv`, `LookupEnv`, `time.Now` or `os.`.
- The suites are green under the canary on both trees.
- Weak asserts: N2 and N3.
- The "non-404 refusal keeps its wording" case checks only that the answer is not `not found` and not 404. It never compares against the operator's wording.
- Commit message: no line over 72 columns.

### Before approval
1. B1: make the typed-route check read what the decoder reads, and add `entitieſ`/`relationſ` cases on both routes.
2. L1: put the "unknown tool" branch through `oneNotFound`, or replace the substring with a sentinel.
3. L2: fix the stale body sections.

All probe files and the `mut` copy are removed; the worktrees and kit copies are unchanged. The kit is at /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-74-r1.
