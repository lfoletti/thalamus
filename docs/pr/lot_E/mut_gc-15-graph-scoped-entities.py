import subprocess, sys
W=sys.argv[1]
RUN='CallerScope|Entit.*Scope|Scope.*Entit|PathSeed|AWalkDoesNotStart|NoCollectionList|SeedsThatCannot|GraphWalks|TheWalks|AWalk'
E="pkg/cortexdb/entity_collections.go"; G="pkg/cortexdb/graphrag_tool_search.go"; B="pkg/cortexdb/brain.go"; P="pkg/cortexdb/graphrag_paths.go"
muts=[
 ("N1 entity rule removed",E,"if !slices.ContainsFunc(toStringSlice(node.Properties[entityCollectionsKey]), permits) {","if false && !slices.ContainsFunc(toStringSlice(node.Properties[entityCollectionsKey]), permits) {"),
 ("N2 entity with no list served",E,"if !slices.ContainsFunc(toStringSlice(node.Properties[entityCollectionsKey]), permits) {","if len(toStringSlice(node.Properties[entityCollectionsKey])) > 0 && !slices.ContainsFunc(toStringSlice(node.Properties[entityCollectionsKey]), permits) {"),
 ("N3 properties kept",E,"served.Properties = map[string]any{}","served.Properties = node.Properties"),
 ("N4 vector kept",E,"served.Vector = nil","served.Vector = node.Vector"),
 ("N5 memory node under the entity rule",E,"permits == nil || memory ||","permits == nil || (memory && false) ||"),
 ("N6 chunk node under the entity rule",E,'|| node.NodeType == "chunk" ||','|| (node.NodeType == "chunk" && false) ||'),
 ("N7 reduced copy not served",G,"keptNodes = append(keptNodes, served)","_ = served\n\t\tkeptNodes = append(keptNodes, node)"),
 ("N8 original nodes returned when nothing dropped",G,"return keptNodes, db.withoutExpiredProvenance(ctx, edges)","return nodes, db.withoutExpiredProvenance(ctx, edges)"),
 ("N9 search_chunks_by_entities seeds unfiltered",G,"range t.db.servedSeeds(ctx, entityIDs)","range entityIDs"),
 ("N10 expand_graph seeds unfiltered",G,"seeds := t.db.servedSeeds(ctx, req.NodeIDs)","seeds := req.NodeIDs"),
 ("N11 seeds unfiltered under a memory rule alone",G,"scope.Knowledge == nil && scope.Memory == nil","scope.Knowledge == nil"),
 ("N12 unreadable seeds kept",G,"\tif err != nil {\n\t\treturn nil\n\t}","\tif err != nil {\n\t\treturn ids\n\t}"),
 ("N13 neighbors seed unfiltered",B,"if len(b.db.servedSeeds(ctx, []string{nodeID})) == 0 {","if false {"),
 ("N14 shortest path keeps the unreduced nodes",B,"path.Nodes, path.Edges = nodes, edges","path.Nodes, path.Edges = path.Nodes, edges"),
 ("N15 path seeds unfiltered",P,"served := t.db.servedSeeds(ctx, ids)","served := ids"),
]
for name,f,old,new in muts:
    p=subprocess.run(["python3","-I","tools/handmut.py","--repo",W,"--file",f,"--old",old,"--new",new,"--","-count=1","-run",RUN,"./pkg/cortexdb/"],capture_output=True,text=True)
    out=p.stdout+p.stderr
    fails=[l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or "build failed" in l or "panic:" in l][:4]
    print(f"{name}: {'KILLED' if p.returncode==0 else 'SURVIVED/REFUSED'}")
    for l in fails: print("   ",l)
    if p.returncode!=0: print(out[-600:])
