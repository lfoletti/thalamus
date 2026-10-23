#!/usr/bin/env python3
"""Several replacements at once on a copy-restored set of files, go test under the canary.
Usage: mutmulti.py --repo R --spec spec.json -- <go test args>
spec.json: [{"file":..., "old":..., "new":...}, ...]. Exit 0 killed, 1 survived, 2 refused."""
import argparse, json, os, subprocess, sys
ap=argparse.ArgumentParser(); ap.add_argument('--repo',required=True); ap.add_argument('--spec',required=True); ap.add_argument('rest',nargs=argparse.REMAINDER)
a=ap.parse_args(); rest=a.rest[1:] if a.rest[:1]==['--'] else a.rest
spec=json.load(open(a.spec)); saved={}
for m in spec:
    p=os.path.join(a.repo,m['file'])
    if p not in saved: saved[p]=open(p).read()
cur={p:s for p,s in saved.items()}
for m in spec:
    p=os.path.join(a.repo,m['file'])
    if cur[p].count(m['old'])!=1: print('refused: old found',cur[p].count(m['old']),'times in',m['file']); sys.exit(2)
    cur[p]=cur[p].replace(m['old'],m['new'])
env=dict(os.environ, CORTEXDB_REMOTE='127.0.0.1:1', CORTEXDB_GRPC_TOKEN='canary', CORTEXDB_PATH='/dev/null/canary')
try:
    for p,s in cur.items(): open(p,'w').write(s)
    r=subprocess.run(['go','test']+rest,cwd=a.repo,env=env,capture_output=True,text=True)
finally:
    for p,s in saved.items(): open(p,'w').write(s)
for p,s in saved.items():
    assert open(p).read()==s, 'restore failed '+p
out=r.stdout+r.stderr
lines=[l for l in out.splitlines() if l.startswith('--- FAIL') or l.startswith('FAIL') or 'build failed' in l or l.startswith('ok')]
print('\n'.join(lines))
if any('build failed' in l for l in lines):
    print(out[-1500:]); print('BUILD-FAILED'); sys.exit(3)
killed=any(l.startswith('--- FAIL') or l.startswith('FAIL') for l in lines)
print('KILLED' if killed else 'SURVIVED'); sys.exit(0 if killed else 1)
