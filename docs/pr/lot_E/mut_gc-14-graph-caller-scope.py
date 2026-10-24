import subprocess, sys, re
W=sys.argv[1]
RUN='CallerScope|DocumentCollectionsQuery|OutOfScope|NoKnowledgeRule|GraphWalks|TheWalks|AWalk'
muts=[
 ("M1 collection rule removed","pkg/cortexdb/expiry.go","if len(names) == 0 || slices.ContainsFunc(names, func(name string) bool { return !permits(name) }) {","if len(names) == 0 || slices.ContainsFunc(names, func(name string) bool { return !permits(name) && false }) {"),
 ("M2 chunkless document shown","pkg/cortexdb/expiry.go","if len(names) == 0 || slices.ContainsFunc(","if slices.ContainsFunc("),
 ("M3 scope ignored in hiddenAmong","pkg/cortexdb/expiry.go","scoped, err = db.outOfScopeDocuments(ctx, ids)","scoped, err = db.outOfScopeDocuments(context.Background(), ids)"),
 ("M4 collections query error dropped","pkg/cortexdb/expiry.go",'return nil, fmt.Errorf("load document collections: %w", err)',"return nil, nil"),
 ("M5 hiddenAmong open on a scope error","pkg/cortexdb/expiry.go","\tif err != nil {\n\t\thidden = make(map[string]struct{}, len(ids))","\tif err != nil && false {\n\t\thidden = make(map[string]struct{}, len(ids))"),
 ("M6 advisory expiry error skips the scope","pkg/cortexdb/expiry.go","hidden, err = map[string]struct{}{}, nil","return map[string]struct{}{}"),
 ("M7 scope replaces the expiry","pkg/cortexdb/expiry.go","maps.Copy(hidden, scoped)","maps.Copy(map[string]struct{}{}, scoped); hidden = scoped"),
 ("M8 memory rule not a cause","pkg/cortexdb/graphrag_tool_search.go","if !db.reviewGate && scope.Memory == nil {","if !db.reviewGate && (scope.Memory == nil || true) {"),
 ("M9 memory rule ignored","pkg/cortexdb/graphrag_tool_search.go","\treturn permitted\n","\treturn permitted || true\n"),
 ("M10 unreadable memory shown","pkg/cortexdb/graphrag_tool_search.go","err == nil && db.memoryNodeShown(ctx, row)","err != nil || db.memoryNodeShown(ctx, row)"),
 ("M11 review gate dropped from memory nodes","pkg/cortexdb/graphrag_tool_search.go","if db.reviewGate && reviewHides(row.record.ReviewStatus) {","if false && reviewHides(row.record.ReviewStatus) {"),
 ("M12 review status hides without the gate","pkg/cortexdb/graphrag_tool_search.go","if db.reviewGate && reviewHides(row.record.ReviewStatus) {","if reviewHides(row.record.ReviewStatus) {"),
]
for name,f,old,new in muts:
    p=subprocess.run(["python3","-I","tools/handmut.py","--repo",W,"--file",f,"--old",old,"--new",new,"--","-count=1","-run",RUN,"./pkg/cortexdb/"],capture_output=True,text=True)
    out=p.stdout+p.stderr
    fails=[l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or "build failed" in l or "panic" in l][:4]
    print(f"{name}: exit {p.returncode} {'KILLED' if p.returncode==0 else 'SURVIVED/REFUSED'}")
    for l in fails: print("   ",l)
    if p.returncode!=0: print(out[-600:])
