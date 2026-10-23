"""Hand mutations of gc-28: each rule, the test that must fail. Runs tools/handmut.py."""
import subprocess, sys
REPO = "/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-28-check-against-sources"
F = "pkg/cortexdb/check_sources.go"
A = "pkg/authz/allowlist.go"
M = [
 ("first word excluded only when only its initial is a capital, removed", F, "(i > 0 || !onlyInitialCapital(token))", "(i >= 0 || !onlyInitialCapital(token))", "TestTheFirstWordIsNotAFactWhenOnlyItsInitialIsCapital", "./pkg/cortexdb/"),
 ("first word: small letters after the capital not required", F, "i > 0 && !unicode.IsLower(r)", "false", "TestTheFirstWordIsNotAFactWhenOnlyItsInitialIsCapital", "./pkg/cortexdb/"),
 ("a digit makes a word a fact, removed", F, "return digits > 0 || capitalAfter ||", "return capitalAfter ||", "TestAFactIsANumberOrAMarkedWord", "./pkg/cortexdb/"),
 ("a capital after the first letter makes a fact, removed", F, "if i > 0 && unicode.IsUpper(r) {", "if false && i > 0 && unicode.IsUpper(r) {", "TestAFactIsANumberOrAMarkedWord", "./pkg/cortexdb/"),
 ("four letters or more, changed to five", F, "runes >= 4", "runes >= 5", "TestAFactIsANumberOrAMarkedWord", "./pkg/cortexdb/"),
 ("a decimal comma read as a point in the sentence, removed", F, 'add(strings.ReplaceAll(n, ",", "."))', "add(n)", "TestAFactIsANumberOrAMarkedWord", "./pkg/cortexdb/"),
 ("_ in a token, removed", F, 'strings.ContainsRune("_.+-", r)', 'strings.ContainsRune(".+-", r)', "TestAFactIsANumberOrAMarkedWord", "./pkg/cortexdb/"),
 ("edge . + - trimmed, removed", F, 'strings.Trim(string(run), ".+-")', "string(run)", "TestTheFirstWordIsNotAFactWhenOnlyItsInitialIsCapital", "./pkg/cortexdb/"),
 ("a digit-only token excluded from words, removed", F, "if digits == runes {", "if false && digits == runes {", "TestADigitOnlyTokenIsANumberNotAWord", "./pkg/cortexdb/"),
 ("marks inside a token, removed", F, "unicode.IsLetter(r) || unicode.IsMark(r) ||", "unicode.IsLetter(r) ||", "TestFoldingIgnoresCaseAndAccentsOnBothSides", "./pkg/cortexdb/"),
 ("Mn marks folded away, removed", F, "case unicode.Is(unicode.Mn, r):", "case false:", "TestFoldingIgnoresCaseAndAccentsOnBothSides", "./pkg/cortexdb/"),
 ("ß folded to ss, removed", F, 'b.WriteString("ss")', "b.WriteRune(r)", "TestFoldingIgnoresCaseAndAccentsOnBothSides", "./pkg/cortexdb/"),
 ("the fact not folded (one side only)", F, "folded := foldFact(fact)", "folded := fact", "TestFoldingIgnoresCaseAndAccentsOnBothSides", "./pkg/cortexdb/"),
 ("the sources not folded (one side only)", F, 'text := foldFact(strings.Join(texts, "\\n"))', 'text := strings.Join(texts, "\\n")', "TestFoldingIgnoresCaseAndAccentsOnBothSides", "./pkg/cortexdb/"),
 ("a decimal comma folded to a point in the sources, removed", F, "case r == ',':", "case false:", "TestANumberIsFoundOnlyBetweenNonDigits", "./pkg/cortexdb/"),
 ("a number found only between non-digits, removed (substring)", F, "if !unicode.IsDigit(before) && !unicode.IsDigit(after) {", "if !unicode.IsDigit(before) || true || !unicode.IsDigit(after) {", "TestANumberIsFoundOnlyBetweenNonDigits", "./pkg/cortexdb/"),
 ("a hyphen is part of a number, removed", F, "r != '.' && r != '-'", "r != '.'", "TestANumberIsFoundOnlyBetweenNonDigits", "./pkg/cortexdb/"),
 ("a number searched past its first occurrence, removed", F, "from = i + 1", "return false", "TestANumberIsFoundOnlyBetweenNonDigits", "./pkg/cortexdb/"),
 ("equivalents ignored", F, "|| slices.ContainsFunc(alternatives[folded]", "|| false && slices.ContainsFunc(alternatives[folded]", "TestAnEquivalentFromTheRequestStandsForTheFact", "./pkg/cortexdb/"),
 ("equivalent keys not folded", F, "key := foldFact(fact)\n\t\tfor _, alt", "key := fact\n\t\tfor _, alt", "TestAnEquivalentFromTheRequestStandsForTheFact", "./pkg/cortexdb/"),
 ("an empty equivalent kept", F, 'alt = foldFact(alt); alt != ""', "alt = foldFact(alt); true", "TestAnEquivalentFromTheRequestStandsForTheFact", "./pkg/cortexdb/"),
 ("CJK cut from its neighbours, removed", F, " || len(run) > 0 && isCJK(run[len(run)-1]) != isCJK(r)", "", "TestHanKanaAndHangulRunsAreNoFactAndAreSplitFromLatin", "./pkg/cortexdb/"),
 ("NFKC on the sentence, removed", F, "sentence = norm.NFKC.String(sentence)", "_ = norm.NFKC", "TestFullwidthDigitsAreNumbers", "./pkg/cortexdb/"),
 ("collection filter, removed", F, " || collection != \"\" && src.collection != collection", "", "TestASourceOfAnotherCollectionIsMissing", "./pkg/cortexdb/"),
 ("a chunk taken to be in the collection asked for", F, "found[c.ID] = source{c.Content, emb.Collection}", "found[c.ID] = source{c.Content, collection + emb.Collection[:0]}", "TestASourceOfAnotherCollectionIsMissing", "./pkg/cortexdb/"),
 ("hiddenAmong on documents, removed", F, "_, isHidden := hidden[id]", '_, isHidden := hidden[""]', "TestAnExpiredOrProposedSourceIsMissing", "./pkg/cortexdb/"),
 ("knowledgeReadPermitted, removed", F, " || !knowledgeReadPermitted(ctx, src.collection)", "", "TestACallerScopeOutsideTheCollectionSeesNoSource", "./pkg/cortexdb/"),
 ("holds despite a missing source", F, " && len(check.SourcesMissing) == 0", "", "TestAMissingSourceIsNamedAndTheSentenceDoesNotHold", "./pkg/cortexdb/"),
 ("holds with no source", F, "check.Holds = len(s.Sources) > 0 && ", "check.Holds = ", "TestASentenceWithNoSourceDoesNotHold", "./pkg/cortexdb/"),
 ("holds despite a missing fact", F, " && len(check.Missing) == 0\n", "\n", "TestCheckAgainstSourcesReadsDocumentsAndChunks", "./pkg/cortexdb/"),
 ("hold not counted", F, "resp.Counts.Hold++", "resp.Counts.Hold += 0", "TestCheckAgainstSourcesReadsDocumentsAndChunks", "./pkg/cortexdb/"),
 ("a failed document read taken as a missing source", F, "if errors.Is(err, core.ErrNotFound) {", "if errors.Is(err, core.ErrNotFound) || err != nil {", "TestACheckWhoseSourcesCannotBeReadFails", "./pkg/cortexdb/"),
 ("bound on sentences, off by one", F, "len(req.Sentences) > maxCheckSentences", "len(req.Sentences) > maxCheckSentences+1", "TestTheCheckIsBounded", "./pkg/cortexdb/"),
 ("bound on sources, off by one", F, "len(s.Sources) > maxCheckSources", "len(s.Sources) > maxCheckSources+1", "TestTheCheckIsBounded", "./pkg/cortexdb/"),
 ("allowlist: collection not forced (scope removed)", A, '"check_against_sources": {scope: collectionScope, stores: storeKnowledge,', '"check_against_sources": {stores: storeKnowledge,', "TestAConfinedKeyChecksOnlyItsCollection", "./pkg/authz/"),
 ("allowlist: sources not admitted as data", A, '"sentences[].text", "sentences[].sources", ', '"sentences[].text", ', "TestAConfinedKeyChecksOnlyItsCollection", "./pkg/authz/"),
 ("MCP registration under another name", "pkg/cortexdb/mcp.go", 'definitions["check_against_sources"]', 'definitions["check_against_sources_off"]', "TestCheckAgainstSourcesIsATool", "./pkg/cortexdb/"),
 ("Call dispatch, removed", "pkg/cortexdb/graphrag_tool_defs.go", 'case "check_against_sources":', 'case "check_against_sources_off":', "TestCheckAgainstSourcesIsATool", "./pkg/cortexdb/"),
]
only = sys.argv[1:]
killed = 0
for i, (what, f, old, new, test, pkg) in enumerate(M, 1):
    if only and str(i) not in only:
        continue
    cmd = ["python3", "-I", "/Users/mini/Documents/granular_cortex/tools/handmut.py", "--repo", REPO, "--file", f, "--old", old, "--new", new, "--", "-count=1", "-run", "^" + test + "$", pkg]
    r = subprocess.run(cmd, capture_output=True, text=True)
    out = r.stdout + r.stderr
    fails = [l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or (".go:" in l and "_test.go" in l)]
    if any("[build failed]" in l or "setup failed" in l for l in out.splitlines()):
        r = subprocess.CompletedProcess(cmd, 3)
        print(out[-400:])
    state = {0: "killed", 1: "SURVIVED", 2: "REFUSED", 3: "BUILD FAILED"}.get(r.returncode, str(r.returncode))
    killed += r.returncode == 0
    print(f"{i:2d}. {state}: {what} -> {test}")
    for l in fails[:2]:
        print("      " + l[:220])
    if r.returncode == 2:
        print(out[-600:])
print(f"{killed}/{len(M) if not only else len(only)} killed")
