import subprocess
RUN = "Ingest|Dedupe|Hash|Expir|Sweep|Review|Gate|Reviewer"
I = "pkg/cortexdb/graphrag_tool_ingest.go"; E = "pkg/cortexdb/expiry.go"; R = "pkg/cortexdb/review.go"; M = "pkg/cortexdb/memory_api.go"
MUT = [
 ("08 hash not stored",        I, 'metadata[contentHashKey] = contentSHA256(content)', 'metadata[contentHashKey+"_off"] = contentSHA256(content)'),
 ("08 unchanged never detected", I, 'sameIngest(existing, req, metadata) {', 'sameIngest(existing, req, metadata) && false {'),
 ("08 dedupe ignored",         I, 'if req.Dedupe {', 'if req.Dedupe && false {'),
 ("05 bad expires_at accepted", E, 't, err := time.Parse(time.RFC3339, value)\n\tif err != nil {', 't, err := time.Parse(time.RFC3339, value)\n\tif err != nil && false {'),
 ("05 zone not normalized",    E, 'return t.UTC().Format(time.RFC3339), nil', '_ = t\n\treturn value, nil'),
 ("05 expired still served",   E, 'if t, ok := until[documentID]; ok && !t.After(time.Now().UTC()) {', 'if t, ok := until[documentID]; ok && !t.After(time.Now().UTC()) && false {'),
 ("05 sweep removes nothing",  E, 'if dryRun {\n\t\treturn report, nil\n\t}', 'if dryRun || true {\n\t\treturn report, nil\n\t}'),
 ("05 dry run deletes",        E, 'if dryRun {\n\t\treturn report, nil\n\t}', 'if false {\n\t\treturn report, nil\n\t}'),
 ("05 memories not swept",     E, 'report.Memories, err = db.expiredMemoryIDs(ctx, now)', 'report.Memories, err = []string{}, nil'),
 ("06 reviewer not required",  R, 'if strings.TrimSpace(reviewedBy) == "" {', 'if strings.TrimSpace(reviewedBy) == "" && false {'),
 ("06 made-up status accepted", R, 'default:\n\t\treturn "", fmt.Errorf("%s must be %s, %s or %s, got %q",', 'default:\n\t\treturn reviewedAt, nil\n\t\treturn "", fmt.Errorf("%s must be %s, %s or %s, got %q",'),
 ("06 bad reviewed_at accepted", R, 'if _, err := time.Parse(time.RFC3339, reviewedAt); err != nil {', 'if _, err := time.Parse(time.RFC3339, reviewedAt); err != nil && false {'),
 ("06 gate ignored (documents)", R, 'return !db.reviewGate || !reviewHides(status)', 'return true'),
 ("06 gate always on (documents)", R, 'return !db.reviewGate || !reviewHides(status)', 'return !reviewHides(status)'),
 ("06 gate ignored (memories)", R, 'return db.reviewGate && reviewHides(record.ReviewStatus)', 'return false'),
 ("06 gate always on (memories)", R, 'return db.reviewGate && reviewHides(record.ReviewStatus)', 'return reviewHides(record.ReviewStatus)'),
 ("06 filter by status ignored", R, '\treturn status == want\n}', '\treturn true\n}'),
 ("06 filter accepted ignored", R, 'return status == "" || status == ReviewAccepted', 'return true'),
 ("06 WithReviewGate no-op",   R, 'db.reviewGate = true', 'db.reviewGate = false'),
 ("06 status not on MemoryRecord", M, 'record.ReviewStatus, _ = stringFromAny(record.Metadata[reviewStatusKey])', 'record.ReviewStatus = ""'),
]
for name, path, old, new in MUT:
    src = open(path, encoding="utf-8").read()
    if src.count(old) != 1:
        print(f"{name:34s} SKIP (pattern found {src.count(old)} times)", flush=True); continue
    open(path, "w", encoding="utf-8").write(src.replace(old, new))
    r = subprocess.run(["go", "test", "./pkg/cortexdb/", "-run", RUN], capture_output=True, text=True)
    fails = [l.split()[2] for l in r.stdout.splitlines() if l.startswith("--- FAIL")]
    if r.returncode == 0: v = "SURVIVED (no test fails)"
    elif fails: v = "KILLED by " + ", ".join(fails)
    else: v = "BUILD ERROR: " + (r.stderr or r.stdout).strip().splitlines()[-1][:120]
    print(f"{name:34s} {v}", flush=True)
    subprocess.run(["git", "checkout", "HEAD", "--", path], check=True)
print(subprocess.run(["git", "status", "--short"], capture_output=True, text=True).stdout or "tree clean")
