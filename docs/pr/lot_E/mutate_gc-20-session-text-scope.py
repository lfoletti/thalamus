import subprocess, sys
W="/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/wtE/gc-20-session-text-scope"
KS="pkg/authz/key_source.go"
RUN_A="KeySource|SameText|SameSize|BrokenKeyFile|BrokenFile|FirstLoad|StaticKeySource"
RUN_S="KeyFile|Preloaded|AuthInterceptorFromSource|Token|Unauth|Open|Disabled|NoToken|Auth"
F="pkg/cortexdb/extract_conversation.go"
R="TestExtractConversation"
M=[
 ("M1 guard removed", F, "\tif !permitted {\n\t\treturn \"\", nil\n\t}\n\tgate", "\t_ = permitted\n\tgate", "./pkg/cortexdb/", R),
 ("M2 own refusal error", F, "\tif !permitted {\n\t\treturn \"\", nil", "\tif !permitted {\n\t\treturn \"\", fmt.Errorf(\"extract_conversation: session withheld\")", "./pkg/cortexdb/", R),
 ("M3 bucket read error swallowed", F, "\tif err != nil {\n\t\treturn \"\", fmt.Errorf(\"extract_conversation: load session: %w\", err)\n\t}\n\tif !permitted", "\tif !permitted", "./pkg/cortexdb/", R),
 ("M4 condition inverted", F, "\tif !permitted {", "\tif permitted {", "./pkg/cortexdb/", R),
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
