# Relecture adverse de #74, tour 2 partiel (22 octobre 2026, 01 h 01)

Tête `ac16f92`, base `d805527`. Tour partiel sur la différence `bc2175d..ac16f92` (B1, L1, L2 du tour 1) et l'accord des textes, même méthode que le tour 4 de #73. Agent `upstream-reviewer`, kit `adv-74-r2` avec `interdiff.diff`. Texte de l'agent ci-dessous, tel quel.

Coût (`python3 -I tools/agentcost.py <journal>`) : tours 26, contexte 18774 -> 93772 tokens ; écriture cache 93770, lecture cache 1529422, entrée neuve 52 ; équivalent entrée (lecture x0,1, écriture x1,25) : 270154 ; outils : {'Bash': 26, 'SubagentHandback': 1}.

---

Partial round 2 of #74 (head ac16f92, base d805527, against round 1 on bc2175d). I found no blocker. B1, L1 and L2 are closed. N1 to N4 are each named in failing_first.md: N1 is deferred to #72, and N2 to N4 are left as they are. There are two test gaps: two of my own mutations survive the whole package.

**1. Does the check read what the handler reads?** Yes, by construction. `decodeTyped` (pkg/httpapi/policy.go:216) runs the same `json.Decoder` with `DisallowUnknownFields` on the same buffered bytes, into the same facade type that `jsonCall`/`decodeBody` uses. It differs from `decodeBody` only in being stricter-or-equal: an empty body is refused (the handler takes it as the zero request, with no entities), and trailing data is not checked (the row check already gives 400 for that). I probed 25 bodies and 2 searches on both trees:
- **Folded field names:** `ENTITIES`, `entitieſ`, `METADATA`, `metadatA`, `_Review_Status` are all 403 on head and 200 on main.
- **Duplicate keys:** `"entitieſ":null,"entities":[…]` is 403. `"entities":[…],"entitieſ":null` is 200 on both trees, and the handler sees nil because the last key wins. Two metadata objects merge (`"metadata":{"_scope":…},"Metadata":{}`), and that merged map is refused.
- **Nulls and empty lists:** `null` values and `"entities":[]` pass.
- **A second JSON value after the first:** 400 from the row check, on both trees.
- **Malformed metadata:** a non-string value in knowledge metadata is the decoder's 400 on both trees.
- **Signature keys:** `written_by` and `updated_by` pass, and the facade's signature is kept.

**The `reaches` claim holds for every scope field.** For a key of the other store, the scope field in any spelling (`collectioſ`, `uſer_id`, `NAMESPACE`, `scope`, `user_id`) is the decoder's 400 `unknown field`. A body without it is the row checks' 403 "leaves … unset". This covers the notes, hermes, lab and mine keys.

**Side effect, not claimed anywhere.** On main, the notes key with `{"collection":"notes","Collection":""}` gets 200 and the row lands in `graphrag_chunks`, outside its collection. That is a main hole: `otherSpellings` skips empty values and the decoder's last key wins. Head now refuses it with 403 "leaves collection unset". The same duplicate-empty trick on memory (`User_Id`, `Namespace`, `ſcope`) is 400 or 404 on main and 403 on head; I saw no escape there. Search with `"Collection":""` does not leak on either tree (the facade confines it).

**2. `go test -race ./pkg/httpapi/`** passes on main (23.9 s) and on head (22.2 s).

**3. Claims on ac16f92.**
- **Mutation count:** the "Sixième passage" table under #74 in failing_first.md has 23 rows. The body and the reply both say 23. It matches.
- **B1 closed:** the new rows of `TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes` hold it, and my probes agree.
- **L1 closed:** pkg/httpapi/tools.go:61-66 now sends "unknown tool" through `oneNotFound`. The probe gives `404 "not found"` to the confined key, and the operator keeps the tool's wording. An undefined tool name is refused 403 before it runs, as claimed.
- **L2 closed:** the Fix section, the Failing first paragraph ("first round's seven … the second review's are above") and the Tests list are all fixed.
- **N1 to N4:** each is named in the failing_first.md list.
- **Nit, commit message and reply:** "A body the decoder refuses, which covers a key that reaches no row of the route's store, is left to the handler's 400" omits the row checks' 403 half. The body states that half.
- **Nit, body:** "the body read by name instead of decoded (on each route)" has a single matching row in the table (`knowledgeToolArgs`). The memory route's matching row is "Entités et relations d'une mémoire non transmises".

**4. My mutations** (run with mutone.py; the survivors were rerun against the whole package):
- **M1, killed:** drop `DisallowUnknownFields` in `decodeTyped`. `TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers` fails.
- **M2, killed:** knowledge relations count forced to 0. `TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes` fails.
- **M3, killed:** `decodeTyped` replaced by `json.Unmarshal`. `TestATypedWriteFromAKeyOfTheOtherStoreKeepsMainsAnswers` fails.
- **M4, SURVIVED (likely comment), policy.go:249:** drop `authz.FieldNamespace: req.Namespace` from `memoryToolArgs`. A namespace-confined key's legitimate POST /v1/memory would then be refused 403 "leaves namespace unset". My probe shows the lab key's legitimate write is 200 on head, but no test sends one, so dropping just namespace (or just scope) is not held. The table's "Champs de portée" row disables all three fields at once and is killed only through hermes (user_id). #73 got per-field namespace and scope cases for this (its N1); #74 lacks them.
- **M5, SURVIVED (nit, effectively equivalent):** `user_id` hard-coded to "hermes" in `memoryToolArgs`. It differs only on the duplicate-empty `User_Id` body, which the facade rejects anyway (404 on main).
- **M6, SURVIVED (likely comment), policy.go:235:** collection hard-coded to "notes" in `knowledgeToolArgs`. It differs only on `{"collection":"notes","Collection":""}`, which is the escape from main that head closes without a test. Suggest a row in `TestATypedWriteTakesWhatTheAllowlistLineOfItsToolTakes` expecting 403 "leaves collection unset", plus a line in the body: it fixes a main hole on POST /v1/knowledge.

**Files**
- Kit: /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-74-r2
- Probe sources: /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-74-r2/probe_test.go, /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-74-r2/probe2_test.go, /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-74-r2/probe3_test.go. The copies placed in head/ and main/ were removed.
- mutone.py created a scratch copy at /private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/adv-74-r2/mut, which it restores after each run.

Nothing outside the kit was modified, and nothing was pushed or posted.
