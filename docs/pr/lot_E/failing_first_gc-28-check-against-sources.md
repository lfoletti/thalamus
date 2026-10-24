# Failing first: gc-28-check-against-sources

Branch `gc-28-check-against-sources`, head `4b5b37b`, on `origin/main` `a591ae1`. Worktree: `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-28-check-against-sources`.

## On the base

Test files laid on `a591ae1` without the code (`failing_first_gc-28-base.txt`, full output):

- `TestAConfinedKeyChecksOnlyItsCollection` (`pkg/authz`): `denied: check_against_sources is not a tool this server defines, want a denial saying "asks for \"secrets\""`, and the four other cases.
- `TestAConfinedKeyChecksAgainstItsOwnSourcesOverGRPC`: `rpc error: code = PermissionDenied desc = denied: check_against_sources is not a tool this server defines`.
- `TestAConfinedKeyChecksAgainstItsOwnSourcesOverREST`: `check = 403: {"error":"denied: check_against_sources is not a tool this server defines"}`.
- `pkg/cortexdb`: does not compile (`undefined: sentenceFacts`, `undefined: missingFacts`, `undefined: SentenceCheck`). The rules there are held by the mutations below.

## Hand mutations

`python3 -I docs/pr/lot_E/mutate_gc-28.py [n …]` from `/Users/mini/Documents/granular_cortex` (each through `tools/handmut.py`, under the canary, file restored from its copy). 38 mutations, 38 killed, each with a `--- FAIL` line of the named test and none by a build failure.

| # | Mutation | Test that fails |
| -: | :- | :- |
| 1 | first word excluded only when only its initial is a capital, removed (killed) | `TestTheFirstWordIsNotAFactWhenOnlyItsInitialIsCapital` |
| 2 | first word: small letters after the capital not required (killed) | `TestTheFirstWordIsNotAFactWhenOnlyItsInitialIsCapital` |
| 3 | a digit makes a word a fact, removed (killed) | `TestAFactIsANumberOrAMarkedWord` |
| 4 | a capital after the first letter makes a fact, removed (killed) | `TestAFactIsANumberOrAMarkedWord` |
| 5 | four letters or more, changed to five (killed) | `TestAFactIsANumberOrAMarkedWord` |
| 6 | a decimal comma read as a point in the sentence, removed (killed) | `TestAFactIsANumberOrAMarkedWord` |
| 7 | _ in a token, removed (killed) | `TestAFactIsANumberOrAMarkedWord` |
| 8 | edge . + - trimmed, removed (killed) | `TestTheFirstWordIsNotAFactWhenOnlyItsInitialIsCapital` |
| 9 | a digit-only token excluded from words, removed (killed) | `TestADigitOnlyTokenIsANumberNotAWord` |
| 10 | marks inside a token, removed (killed) | `TestFoldingIgnoresCaseAndAccentsOnBothSides` |
| 11 | Mn marks folded away, removed (killed) | `TestFoldingIgnoresCaseAndAccentsOnBothSides` |
| 12 | ß folded to ss, removed (killed) | `TestFoldingIgnoresCaseAndAccentsOnBothSides` |
| 13 | the fact not folded (one side only) (killed) | `TestFoldingIgnoresCaseAndAccentsOnBothSides` |
| 14 | the sources not folded (one side only) (killed) | `TestFoldingIgnoresCaseAndAccentsOnBothSides` |
| 15 | a decimal comma folded to a point in the sources, removed (killed) | `TestANumberIsFoundOnlyBetweenNonDigits` |
| 16 | a number found only between non-digits, removed (substring) (killed) | `TestANumberIsFoundOnlyBetweenNonDigits` |
| 17 | a hyphen is part of a number, removed (killed) | `TestANumberIsFoundOnlyBetweenNonDigits` |
| 18 | a number searched past its first occurrence, removed (killed) | `TestANumberIsFoundOnlyBetweenNonDigits` |
| 19 | equivalents ignored (killed) | `TestAnEquivalentFromTheRequestStandsForTheFact` |
| 20 | equivalent keys not folded (killed) | `TestAnEquivalentFromTheRequestStandsForTheFact` |
| 21 | an empty equivalent kept (killed) | `TestAnEquivalentFromTheRequestStandsForTheFact` |
| 22 | CJK cut from its neighbours, removed (killed) | `TestHanKanaAndHangulRunsAreNoFactAndAreSplitFromLatin` |
| 23 | NFKC on the sentence, removed (killed) | `TestFullwidthDigitsAreNumbers` |
| 24 | collection filter, removed (killed) | `TestASourceOfAnotherCollectionIsMissing` |
| 25 | a chunk taken to be in the collection asked for (killed) | `TestASourceOfAnotherCollectionIsMissing` |
| 26 | hiddenAmong on documents, removed (killed) | `TestAnExpiredOrProposedSourceIsMissing` |
| 27 | knowledgeReadPermitted, removed (killed) | `TestACallerScopeOutsideTheCollectionSeesNoSource` |
| 28 | holds despite a missing source (killed) | `TestAMissingSourceIsNamedAndTheSentenceDoesNotHold` |
| 29 | holds with no source (killed) | `TestASentenceWithNoSourceDoesNotHold` |
| 30 | holds despite a missing fact (killed) | `TestCheckAgainstSourcesReadsDocumentsAndChunks` |
| 31 | hold not counted (killed) | `TestCheckAgainstSourcesReadsDocumentsAndChunks` |
| 32 | a failed document read taken as a missing source (killed) | `TestACheckWhoseSourcesCannotBeReadFails` |
| 33 | bound on sentences, off by one (killed) | `TestTheCheckIsBounded` |
| 34 | bound on sources, off by one (killed) | `TestTheCheckIsBounded` |
| 35 | allowlist: collection not forced (scope removed) (killed) | `TestAConfinedKeyChecksOnlyItsCollection` |
| 36 | allowlist: sources not admitted as data (killed) | `TestAConfinedKeyChecksOnlyItsCollection` |
| 37 | MCP registration under another name (killed) | `TestCheckAgainstSourcesIsATool` |
| 38 | Call dispatch, removed (killed) | `TestCheckAgainstSourcesIsATool` |

`TestTheEngineHoldsNoWordList` holds a property, not a rule: no mutation removes a list the code does not have. It fails if one is added (RGPD read as GDPR, or "Elle" after a full stop dropped).

## Removed while writing

- The CJK guard in `isWordFact` (a token starting with a Han, kana or hangul rune is no fact): no input reaches it. After the cut, a CJK run has no digit and no case, so the other conditions already refuse it. Removed rather than declared; the comment says why.

## Class lines

| Element | Test |
| :- | :- |
| `check_against_sources` (tools, `toolWrites`) | `TestEveryToolDeclaresWhetherItWrites`: `false`; `toolCount` 101 |
| `check_against_sources` (reads, `readToolsAndTheExpiry`) | named, not "gated": `TestAnExpiredOrProposedSourceIsMissing` |
| `check_against_sources` (allowlist) | `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys`, `TestAConfinedKeyChecksOnlyItsCollection` |
| `check_against_sources` (MCP, `Call`) | `TestCheckAgainstSourcesIsATool`, `TestEveryToolDefinitionIsReachableOverMCP`, `TestEveryDefinedToolIsReachableThroughCall` |
| `CheckAgainstSources` (exported entry, reads) | every facade test above; caller scope: `TestACallerScopeOutsideTheCollectionSeesNoSource` |

## precommit

`make precommit REPO=<worktree> BASE=a591ae1`: OK (gofmt, vet, `-race` of `pkg/authz`, `pkg/cortexdb`, `pkg/httpapi`, `pkg/rpcserver`, patchcov 0 blocks without test). A first run failed on `TestTheToolDescriptionsSayTheGateIsACurationFilter`: the description said "review gate" without the gate note; reworded to "hidden pending review".

Not run: `make prepr` (gremlins, clock, full suite, lexical eval), `make diffmain`, `make inventory`/`invcheck`, the adversarial review. Level 2 and level 3 are left for the lot.
