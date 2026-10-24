import subprocess, sys
W=sys.argv[1]
A="pkg/authz/allowlist.go"
AUTHZ=["-count=1","-run","GraphWalks|WholeGraph|EveryToolAndEvery|AllowlistTable","./pkg/authz/"]
E2E=["-count=1","-run","WalksTheGraphOfItsCollection","./pkg/rpcserver/","./pkg/httpapi/"]
muts=[
 ("P1 the facade still asks for a gate",A,"other != 0 && !rule.facade {","other != 0 {",AUTHZ+["./pkg/rpcserver/","./pkg/httpapi/"]),
 ("P2 a walk off the allowlist",A,'"expand_graph": {stores:','"expand_graph_x": {stores:',AUTHZ),
 ("P3 the table hides the facade",A,'gates = append(gates, "by the facade (CallerScope)")','gates = append(gates, "")',AUTHZ),
 ("P4 an argument left unclassified",A,'"get_nodes": {stores: storeKnowledge | storeMemory, facade: true, args: []string{"node_ids"}},','"get_nodes": {stores: storeKnowledge | storeMemory, facade: true},',AUTHZ),
 ("P5 find_nodes without its reason",A,'\t"find_nodes":',"\t\"find_nodes_x\":",AUTHZ),
 ("P6 a whole-graph read without its reason",A,'\t"graph_list_all":',"\t\"graph_list_all_x\":",AUTHZ),
 ("P7 the facade ignores the scope (documents)","pkg/cortexdb/expiry.go","scoped, err = db.outOfScopeDocuments(ctx, ids)","scoped, err = db.outOfScopeDocuments(context.Background(), ids)",E2E),
 ("P8 the facade serves every entity","pkg/cortexdb/entity_collections.go","if !slices.ContainsFunc(toStringSlice(node.Properties[entityCollectionsKey]), permits) {","if false && !slices.ContainsFunc(toStringSlice(node.Properties[entityCollectionsKey]), permits) {",E2E),
 ("P9 the facade serves the entity whole","pkg/cortexdb/entity_collections.go","served.Properties = map[string]any{}","served.Properties = node.Properties",E2E),
 ("P10 the facade ignores the memory rule","pkg/cortexdb/graphrag_tool_search.go","\treturn permitted\n","\treturn permitted || true\n",E2E),
]
for name,f,old,new,gt in muts:
    p=subprocess.run(["python3","-I","tools/handmut.py","--repo",W,"--file",f,"--old",old,"--new",new,"--"]+gt,capture_output=True,text=True)
    out=p.stdout+p.stderr
    fails=[l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or "build failed" in l or "panic:" in l][:4]
    print(f"{name}: {'KILLED' if p.returncode==0 else 'SURVIVED/REFUSED'}")
    for l in fails: print("   ",l)
    if p.returncode!=0: print(out[-600:])
