import subprocess, sys
R="/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-17-query-graph-lane-scope"
F="pkg/cortexdb/query_graph_scope.go"
RUN="TestTheGraphLane|TestTheScopedWalk|TestDocumentsInCollection|TestAScopedWalk|TestALaneStep"
M=[
("M1-seed",'seen || !scope.mayEnter(step) {','seen || (depth > 0 && !scope.mayEnter(step)) {'),
("M2-entity",'if !entityKnownIn(node.Collections, s.collection) {','if false {'),
("M3-memory",'\tcase "memory":\n\t\treturn false','\tcase "memory":'),
("M4-chunkdoc",'\t\tif _, in := s.docs[node.DocumentID]; !in {\n\t\t\treturn false','\t\tif false {\n\t\t\treturn false'),
("M5-provrefused",'\t\t\treturn true\n\t\t}\n\t}\n\treturn false\n}','\t\t\treturn true\n\t\t}\n\t}\n\treturn true\n}'),
("M6-provdoc",'\tif _, in := s.docs[doc]; in {','\tif false {'),
("M7-provchunk",'\t\tif _, in := s.docs[s.nodes[chunk].DocumentID]; in {','\t\tif _, in := s.docs[s.nodes[chunk].DocumentID]; in && false {'),
("M8-provchunkread",'ids = append(append(ids, step.next), chunks...)','ids = append(append(ids, step.next), chunks[:0]...)'),
("M9-noprov",'\tif doc == "" && len(chunks) == 0 {\n\t\treturn true','\tif doc == "" && len(chunks) == 0 {\n\t\treturn false'),
("M10-limit",'if opts.Limit > 0 && len(visited)-1 >= opts.Limit {','if false {'),
("M11-types",'if len(opts.NodeTypes) == 0 || slices.Contains(opts.NodeTypes, scope.nodes[step.next].NodeType) {','if len(opts.NodeTypes) == 0 || slices.Contains(opts.NodeTypes, scope.nodes[step.next].NodeType) || true {'),
("M12-notypes",'if len(opts.NodeTypes) == 0 || slices.Contains(','if slices.Contains('),
("M13-depth",'if depth == opts.MaxDepth {','if depth == opts.MaxDepth+1 {'),
("M14-unscoped",'\tif collection == "" {\n\t\treturn db.graph.Neighbors','\tif false {\n\t\treturn db.graph.Neighbors'),
("M15-direction",'\t\t\t\tif next == id {\n\t\t\t\t\tnext = edge.ToNodeID\n\t\t\t\t}\n',''),
("M16-sqlcoll",'WHERE c.name = ? AND','WHERE c.name <> ? AND'),
("M17-seen",'if _, seen := visited[step.next]; seen || !scope.mayEnter(step) {','if !scope.mayEnter(step) {'),
("M18-seedskip",'\t\t\tif depth == 0 {\n\t\t\t\tcontinue\n\t\t\t}\n',''),
("M19-docsprov",'\t\tdocIDs = append(docIDs, doc)\n\t}','\t\tdocIDs = append(docIDs, doc[:0])\n\t}'),
("M20-sourcechunk",'\t\tchunks = append(chunks, src)','\t\tchunks = append(chunks, src[:0])'),
("M21-unread-node",'\tif !ok {\n\t\treturn false\n\t}\n\tswitch','\tif !ok {\n\t\treturn true\n\t}\n\tswitch'),
]
only = sys.argv[1:]
for name, old, new in M:
    if only and name not in only: continue
    p = subprocess.run(["python3","-I","/Users/mini/Documents/granular_cortex/tools/handmut.py","--repo",R,"--file",F,"--old",old,"--new",new,"--","-count=1","-run",RUN,"./pkg/cortexdb/"],capture_output=True,text=True)
    print("===", name, "exit", p.returncode)
    for l in (p.stdout+p.stderr).splitlines():
        if l.startswith(("--- FAIL","handmut","FAIL")) or "_test.go:" in l or "build failed" in l or ".go:" in l:
            print("  ", l[:220])
    sys.stdout.flush()
