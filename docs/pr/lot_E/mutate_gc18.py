import subprocess, shutil, os, sys
R="/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-18-allowlist-cortex-query"
F=os.path.join(R,"pkg/authz/allowlist.go")
CANARY={"CORTEXDB_REMOTE":"127.0.0.1:1","CORTEXDB_GRPC_TOKEN":"canary","CORTEXDB_PATH":"/dev/null/canary"}
RUN="CortexQuery|QueriesItsCollection|MaxHops|QuerySource|Classified|OffTheAllowlist|Review"
M=[
("G1-offlist",[('"cortex_query": {scope: collectionScope','"cortex_query_off": {scope: collectionScope')]),
("G2-noscope",[('"cortex_query": {scope: collectionScope,','"cortex_query": {scope: nil,')]),
("G3-maxhops-open",[('"limit", "keywords", "alternate_queries"),','"limit", "keywords", "alternate_queries", "max_hops"),'),
                    ('"prefetch[].max_hops": "walks','"prefetch[].unused": "walks')]),
("G4-condition-arg",[('queryConditionArgs = []string{"field", "op", "value", "values", "number", "has_number"}','queryConditionArgs = []string{"field", "op", "value", "values", "number"}')]),
("G5-still-refused",[('var refusedTools = map[string]string{\n','var refusedTools = map[string]string{\n\t"cortex_query": "x",\n')]),
("G6-source-open",[('"limit", "keywords", "alternate_queries"),','"limit", "keywords", "alternate_queries", "source"),')]),
]
dirty = subprocess.run(["git","-C",R,"diff","--quiet","--","pkg/authz/allowlist.go"]).returncode
if dirty: sys.exit("dirty")
orig=open(F).read()
for name, edits in M:
    if sys.argv[1:] and name not in sys.argv[1:]: continue
    s=orig
    for old,new in edits:
        assert s.count(old)==1,(name,old)
        s=s.replace(old,new)
    open(F,"w").write(s)
    try:
        p=subprocess.run(["go","test","-count=1","-run",RUN,"./pkg/authz/","./pkg/rpcserver/","./pkg/httpapi/"],cwd=R,env=dict(os.environ,**CANARY),capture_output=True,text=True)
    finally:
        open(F,"w").write(orig)
    assert open(F).read()==orig
    out=p.stdout+p.stderr
    fails=[l for l in out.splitlines() if l.startswith("--- FAIL") or "build failed" in l]
    print("===",name,"KILLED" if fails else "SURVIVED")
    for l in fails: print("  ",l[:200])
