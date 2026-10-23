import subprocess, sys
W="/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-19-key-file-reload"
KS="pkg/authz/key_source.go"
RUN_A="KeySource|SameText|SameSize|BrokenKeyFile|BrokenFile|FirstLoad|StaticKeySource"
RUN_S="KeyFile|Preloaded|AuthInterceptorFromSource|Token|Unauth|Open|Disabled|NoToken|Auth"
M=[
 ("M1 Current never stats", KS, "\tinfo, err := os.Stat(s.path)\n\tif err == nil && sameFileState(snap.info, info) {", "\tinfo, err := os.Stat(s.path)\n\tif true || err == nil && sameFileState(snap.info, info) {", "./pkg/authz/ ./pkg/rpcserver/ ./pkg/httpapi/", RUN_A+"|"+RUN_S),
 ("M2 new set not published", KS, "\t\t\t\ts.cur.Store(&keySnapshot{keys: keys, data: data, info: info})\n", "", "./pkg/authz/", RUN_A),
 ("M3a failure publishes nil", KS, "s.cur.Store(&keySnapshot{keys: snap.keys, data: snap.data, info: info})", "s.cur.Store(&keySnapshot{keys: nil, data: snap.data, info: info})", "./pkg/authz/ ./pkg/rpcserver/ ./pkg/httpapi/", RUN_A+"|"+RUN_S),
 ("M3b failure returns nil", KS, "\t}\n\treturn snap.keys\n}\n\nfunc (s *KeySource) warn", "\t}\n\treturn nil\n}\n\nfunc (s *KeySource) warn", "./pkg/authz/ ./pkg/rpcserver/ ./pkg/httpapi/", RUN_A+"|"+RUN_S),
 ("M4 never warns", KS, "\t\ts.warn(fmt.Errorf(", "\t\t_ = (fmt.Errorf(", "./pkg/authz/", RUN_A),
 ("M5 warns at each read", KS, "if key != s.failed {", "if true {", "./pkg/authz/", RUN_A),
 ("M6 failed key not reset", KS, "\t\t\t\ts.failed = \"\"\n", "", "./pkg/authz/", RUN_A),
 ("M7 always parse", KS, "if !bytes.Equal(data, snap.data) {", "if !bytes.Equal(data, nil) {", "./pkg/authz/", RUN_A),
 ("M8 no SameFile", KS, " && os.SameFile(a, b)", "", "./pkg/authz/", RUN_A),
 ("M9 broken file reread each time", KS, "\t\t\ts.cur.Store(&keySnapshot{keys: snap.keys, data: snap.data, info: info})\n", "", "./pkg/authz/", RUN_A),
 ("M10 static source stats", KS, "\tif s.path == \"\" {\n\t\treturn snap.keys\n\t}\n", "", "./pkg/authz/", RUN_A),
 ("M11 ResolveSource ignores the key file", KS, "\tif keyFilePath != \"\" {\n\t\treturn OpenKeySource", "\tif false {\n\t\treturn OpenKeySource", "./pkg/rpcserver/ ./pkg/httpapi/", RUN_S),
 ("M12 no log without Warn", KS, "\tlog.Printf(\"authz: %v\", err)", "\t_ = log.Printf", "./pkg/authz/", "WithoutWarn"),
 ("M13 gRPC captures Current once", "pkg/rpcserver/server.go", "authInterceptorFrom(src, db))", "authInterceptor(src.Current(), db))", "./pkg/rpcserver/", RUN_S),
 ("M14 REST captures Current once", "pkg/httpapi/server.go", "return authenticateFrom(src, rt), nil", "return authenticate(src.Current(), rt), nil", "./pkg/httpapi/", RUN_S),
 ("M15 gRPC rereads beside Options.Keys", "pkg/rpcserver/server.go", "\tif opts.Keys == nil {\n\t\tvar err error\n\t\tsrc, err = authz.ResolveSource", "\tif opts.KeyFile != \"\" || opts.Keys == nil {\n\t\tvar err error\n\t\tsrc, err = authz.ResolveSource", "./pkg/rpcserver/", RUN_S),
 ("M16 REST rereads beside Options.Keys", "pkg/httpapi/server.go", "\tif opts.Keys == nil {\n\t\tvar err error\n\t\tsrc, err = authz.ResolveSource", "\tif opts.KeyFile != \"\" || opts.Keys == nil {\n\t\tvar err error\n\t\tsrc, err = authz.ResolveSource", "./pkg/httpapi/", RUN_S),
 ("M17 REST disabled set not passed through", "pkg/httpapi/policy.go", "\t\tif !keys.Enabled() {\n\t\t\tnext.ServeHTTP(w, r)\n\t\t\treturn\n\t\t}\n", "", "./pkg/httpapi/", "."),
 ("M19 legacy token dropped", KS, "return StaticKeySource(LegacyToken(legacyToken)), nil", "return StaticKeySource(nil), nil", "./pkg/authz/", "ResolveSource"),
 ("M18 own server interceptor static", "pkg/rpcserver/auth_export.go", "\treturn authInterceptorFrom(src, db)", "\treturn authInterceptor(src.Current(), db)", "./pkg/rpcserver/", RUN_S),
]
only=sys.argv[1:] 
for name,f,old,new,pkgs,run in M:
    if only and name.split()[0] not in only: continue
    r=subprocess.run(["python3","-I","/Users/mini/Documents/granular_cortex/tools/handmut.py","--repo",W,"--file",f,"--old",old,"--new",new,"--","-count=1","-run",run]+pkgs.split(),capture_output=True,text=True)
    out=r.stdout+r.stderr
    fails=[l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or "build failed" in l or "[build failed]" in l]
    print(f"=== {name}: exit {r.returncode}")
    for l in fails[:6]: print("   ",l)
    if r.returncode!=0: print(out[-600:])
