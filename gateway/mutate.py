#!/usr/bin/env python3
"""Failing-first proof of the gateway: one hand mutation per rule, each must
make its named test fail. Run from gateway/: python3 -I mutate.py"""
import subprocess, sys
from pathlib import Path

M = [
 ("rights/rights.go", "if r.listed(user, b.Writers) {\n\t\treturn Write, nil", "if r.listed(user, b.Writers) {\n\t\treturn Read, nil", "rights", "TestAWriterMayAlsoRead", "a writer reads only"),
 ("rights/rights.go", "if r.listed(user, b.Readers) {\n\t\treturn Read, nil", "if r.listed(user, b.Readers) {\n\t\treturn Write, nil", "rights", "TestAWriterMayAlsoRead", "a reader writes"),
 ("rights/rights.go", "return None, fmt.Errorf(\"unknown brain %q\", brain)", "return None, nil", "rights", "TestAnUnknownBrainIsAnError", "an unknown brain is no access"),
 ("rights/rights.go", "if _, declared := r.Groups[name]; !declared {", "if _, declared := r.Groups[name]; declared && false {", "rights", "TestAMemberMustBeAGroupOrAUser", "an undeclared group passes"),
 ("rights/rights.go", "dec.KnownFields(true)", "dec.KnownFields(false)", "rights", "TestAMisspelledFieldIsRefused", "a misspelled field is ignored"),
 ("rights/rights.go", "if strings.TrimSpace(value) == \"\" {", "if false {", "rights", "TestABrainNeedsItsThreeFields", "a brain without addr passes"),
 ("rights/rights.go", "if a.AllowsRead() {\n\t\t\tout = append(out, Grant{Brain: name, Access: a})", "if true {\n\t\t\tout = append(out, Grant{Brain: name, Access: a})", "rights", "TestBrainsOfListsTheUsersBrainsInOrder", "BrainsOf lists every brain"),
 ("keys/keys.go", "if !access.AllowsRead() {", "if false {", "keys", "TestNoAccessAndNoDuplicateKey", "a key for no access"),
 ("keys/keys.go", "if k.ID == id {\n\t\t\treturn Issued{}, fmt.Errorf(\"a key is already issued", "if false {\n\t\t\treturn Issued{}, fmt.Errorf(\"a key is already issued", "keys", "TestNoAccessAndNoDuplicateKey", "a key issued twice"),
 ("keys/keys.go", "Scope: authz.Scope{Collections: []string{collection}}", "Scope: authz.Scope{}", "keys", "TestAnIssuedKeyIsConfinedToTheBrainsCollection", "an unconfined key"),
 ("keys/keys.go", "Clearance: issued.Clearance,", "Clearance: authz.ReadWrite,", "keys", "TestAnIssuedKeyIsConfinedToTheBrainsCollection", "a reader gets read-write"),
 ("keys/keys.go", "if len(kept) == len(f.Keys) {", "if false {", "keys", "TestRevokeRemovesOneKey", "revoking a missing key passes"),
 ("keys/keys.go", "if k.ID != id {\n\t\t\tkept = append(kept, k)", "if k.ID == id {\n\t\t\tkept = append(kept, k)", "keys", "TestRevokeRemovesOneKey", "revoke keeps the wrong key"),
 ("keys/keys.go", "if !ok {\n\t\t\treturn \"\", \"\", fmt.Errorf(\"key %q was not issued by the gateway\", k.ID)", "if false {\n\t\t\treturn \"\", \"\", fmt.Errorf(\"key %q was not issued by the gateway\", k.ID)", "keys", "TestHolderNamesTheUserBehindASecret", "an operator key has a holder"),
 ("keys/keys.go", "if _, err := authz.Parse(data); err != nil {", "if false {", "keys", "TestAFileTheServerWouldRefuseIsNotWritten", "a file the server refuses is written"),
 ("keys/keys.go", "tmp := path + \".tmp\"\n\terr := os.WriteFile(tmp, append(data, '\\n'), 0o600)\n\tif err == nil {\n\t\terr = os.Rename(tmp, path)\n\t}", "err := os.WriteFile(path, append(data, '\\n'), 0o600)", "keys", "TestAFailedWriteLeavesTheFileAsItWas", "the file is written in place"),
 ("relay/relay.go", "if keyBrain != name {", "if false {", "relay", "TestRefusalsHappenBeforeTheBrainIsReached", "a key of another brain is admitted"),
 ("relay/relay.go", "if !access.AllowsRead() {\n\t\treturn nil, http.StatusForbidden", "if false {\n\t\treturn nil, http.StatusForbidden", "relay", "TestRefusalsHappenBeforeTheBrainIsReached", "a withdrawn right is admitted"),
 ("relay/relay.go", "if !found || secret == \"\" {", "if false {", "relay", "TestRefusalsHappenBeforeTheBrainIsReached", "no token is admitted"),
 ("relay/relay.go", "\"authorization\", \"Bearer \"+secret", "\"authorization\", \"Bearer op-secret\"", "relay", "TestACallIsRelayedWithTheCallersKeyAndAudited", "the brain sees another key"),
 ("relay/relay.go", "entry.IDs = idsOf(args)", "entry.IDs = nil", "relay", "TestACallIsRelayedWithTheCallersKeyAndAudited", "no ids in the audit"),
 ("relay/relay.go", "entry.OK = true\n\t\t_ = s.audit.Record(entry)", "entry.OK = true", "relay", "TestACallIsRelayedWithTheCallersKeyAndAudited", "a success is not audited"),
 ("relay/relay.go", "entry.Error = err.Error()\n\t\t\t_ = s.audit.Record(entry)", "entry.Error = err.Error()", "relay", "TestAToolFailureIsAudited", "a failure is not audited"),
 ("relay/relay.go", "return &mcp.CallToolResult{IsError: true,", "return &mcp.CallToolResult{IsError: false,", "relay", "TestAToolFailureIsAudited", "a failure is a success"),
 ("relay/relay.go", "ArgsJson: args}", "ArgsJson: \"{}\"}", "relay", "TestACallIsRelayedWithTheCallersKeyAndAudited", "the arguments are dropped"),
 ("relay/audit.go", "os.O_CREATE|os.O_WRONLY|os.O_APPEND", "os.O_CREATE|os.O_WRONLY|os.O_TRUNC", "relay", "TestTheAuditFileAppendsAndNamesItselfOnFailure", "the audit file is truncated"),
 ("cmd/gc-gateway/run.go", "fmt.Fprintln(out, k.Secret)", "fmt.Fprintln(out, \"issued\", k.Secret)", "cmd/gc-gateway", "TestKeyIssueAndRevoke", "the secret is not alone on stdout"),
 ("cmd/gc-gateway/run.go", "if errors.Is(err, errUsage) {\n\t\t\tfmt.Fprintln(stderr, usage)\n\t\t\treturn 2", "if errors.Is(err, errUsage) {\n\t\t\tfmt.Fprintln(stderr, usage)\n\t\t\treturn 1", "cmd/gc-gateway", "TestTheCommandLineIsChecked", "a usage error is 1"),
 ("cmd/gc-gateway/run.go", "err = srv.Serve(lis)\n\tif errors.Is(err, http.ErrServerClosed) {\n\t\terr = nil\n\t}", "err = srv.Serve(lis)", "cmd/gc-gateway", "TestServeListensAndStops", "a clean stop is an error"),
]

bad = 0
for f, old, new, pkg, test, label in M:
    p = Path(f); src = p.read_text()
    if src.count(old) != 1:
        print(f"??    {label}: pattern not found once in {f}"); bad += 1; continue
    p.write_text(src.replace(old, new))
    try:
        r = subprocess.run(["go", "test", "./" + pkg + "/", "-run", "^" + test + "$", "-count=1"], capture_output=True, text=True)
    finally:
        p.write_text(src)
    if r.returncode != 0 and ("FAIL" in r.stdout or "FAIL" in r.stderr):
        print(f"KILLED  {label} -> {test}")
    else:
        print(f"ALIVE   {label} -> {test} (test passed with the mutation)"); bad += 1
print("alive:", bad)
sys.exit(1 if bad else 0)
