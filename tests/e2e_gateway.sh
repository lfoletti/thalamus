#!/bin/sh
# End-to-end trial of th-gateway in front of one shared brain: a cortexdb-grpc
# server built from the cortexdb/ branch, a rights file with two brains on it,
# keys issued by the gateway, tool calls over MCP streamable HTTP with curl,
# and the audit file read back. Prints PASS/FAIL per check.
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
S=$(mktemp -d)
(cd "$ROOT/cortexdb" && go build -o "$S/cortexdb-grpc" ./cmd/cortexdb-grpc) || exit 1
(cd "$ROOT/gateway" && go build -o "$S/th-gateway" ./cmd/th-gateway && go build -o "$S/th-hook" ./cmd/th-hook) || exit 1
GRPC=$S/cortexdb-grpc
GW=$S/th-gateway
ADDR=127.0.0.1:47998
GWADDR=127.0.0.1:47997
mkdir -p "$S/data"

# The operator key is the server's own; the gateway adds the employees' keys
# next to it.
cat > "$S/data/keys.json" <<'EOK'
{"keys":[{"id":"operator","secret":"op-secret","clearance":"read-write"}]}
EOK
cat > "$S/data/rights.yaml" <<EOR
brains:
  team/acme:
    addr: $ADDR
    keys_file: $S/data/keys.json
    collection: team-acme
    readers: [group:acme]
    writers: [user:lfoletti]
    curators: [user:kmlead]
  org/secrets:
    addr: $ADDR
    keys_file: $S/data/keys.json
    collection: secrets
    readers: [user:lfoletti]
    writers: [user:kmlead]
    weight: 0.8
groups:
  acme: [lfoletti, mdupont]
EOR

CORTEXDB_REVIEW_GATE=1 "$GRPC" -db "$S/data/brain.db" -addr "$ADDR" -keys "$S/data/keys.json" > "$S/data/server.log" 2>&1 &
SERVER=$!
trap 'kill $SERVER $GATEWAY 2>/dev/null' EXIT
i=0; until "$GRPC" -addr "$ADDR" -health >/dev/null 2>&1 || [ $i -ge 50 ]; do i=$((i+1)); sleep 0.1; done

pass=0; fail=0
ok() { pass=$((pass+1)); echo "PASS  $1"; }
ko() { fail=$((fail+1)); echo "FAIL  $1: $(printf '%s' "$2" | cut -c1-300)"; }

# Keys are issued before the server reads the file: cortexdb-grpc loads its
# key file at start, so in this slice the gateway issues, then the server
# starts (or restarts). Here the server started first to learn whether it
# re-reads the file; it does not, so it is restarted after the issue.
LF=$("$GW" key issue "$S/data/rights.yaml" lfoletti team/acme) || ko "issue lfoletti" "$LF"
LFS=$("$GW" key issue "$S/data/rights.yaml" lfoletti org/secrets) || ko "issue lfoletti on secrets" "$LFS"
MD=$("$GW" key issue "$S/data/rights.yaml" mdupont team/acme) || ko "issue mdupont" "$MD"
KM=$("$GW" key issue "$S/data/rights.yaml" kmlead org/secrets) || ko "issue kmlead" "$KM"
KMA=$("$GW" key issue "$S/data/rights.yaml" kmlead team/acme) || ko "issue kmlead on team/acme" "$KMA"
if "$GW" key issue "$S/data/rights.yaml" mdupont org/secrets >/dev/null 2>&1; then ko "mdupont on secrets issued" ""; else ok "mdupont has no key on org/secrets"; fi
kill $SERVER; wait $SERVER 2>/dev/null
CORTEXDB_REVIEW_GATE=1 "$GRPC" -db "$S/data/brain.db" -addr "$ADDR" -keys "$S/data/keys.json" > "$S/data/server.log" 2>&1 &
SERVER=$!
i=0; until "$GRPC" -addr "$ADDR" -health >/dev/null 2>&1 || [ $i -ge 50 ]; do i=$((i+1)); sleep 0.1; done

"$GW" serve "$S/data/rights.yaml" --listen "$GWADDR" --audit "$S/data/audit.jsonl" > "$S/data/gateway.log" 2>&1 &
GATEWAY=$!
i=0; until grep -q listening "$S/data/gateway.log" || [ $i -ge 50 ]; do i=$((i+1)); sleep 0.1; done

# mcp <secret> <brain> <tool> <args-json> -> the tool result (one JSON line) or the HTTP error
mcp() {
  sec=$1; brain=$2; tool=$3; args=$4
  url="http://$GWADDR/mcp/$brain"; [ -z "$brain" ] && url="http://$GWADDR/mcp"
  hdr=$(curl -s -D - -o /dev/null -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $sec" "$url" \
    -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"e2e","version":"0"}}}')
  sid=$(printf '%s' "$hdr" | tr -d '\r' | awk 'tolower($1)=="mcp-session-id:"{print $2}')
  if [ -z "$sid" ]; then printf '%s' "$hdr" | head -1; return; fi
  curl -s -o /dev/null -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $sec" -H "Mcp-Session-Id: $sid" "$url" \
    -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  curl -s -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $sec" -H "Mcp-Session-Id: $sid" "$url" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool\",\"arguments\":$args}}"
}
check() { # check <name> <ok|denied> <output>
  name=$1; expect=$2; out=$3
  if [ -z "$out" ]; then got=empty
  elif printf '%s' "$out" | grep -q 'PermissionDenied\|denied\|Unauthenticated\|isError\|HTTP/1.1 4'; then got=denied
  else got=ok; fi
  if [ "$got" = "$expect" ]; then ok "$name ($got)"; else ko "$name: want $expect, got $got" "$out"; fi
}

check "lfoletti: ingest into team/acme" ok "$(mcp "$LF" team/acme ingest_document '{"document_id":"spec-1","title":"Spec migration","content":"Acme migrates the ledger to the new cluster on Sunday night. The rollback window is two hours.","collection":"team-acme"}')"
check "lfoletti: ingest into secrets through team/acme (denied by the key)" denied "$(mcp "$LF" team/acme ingest_document '{"document_id":"s-1","content":"Not yours.","collection":"secrets"}')"
check "lfoletti: no key on org/secrets (401)" denied "$(mcp "$LF" org/secrets knowledge_search '{"query":"x","collection":"secrets"}')"
check "mdupont: search team/acme" ok "$(mcp "$MD" team/acme knowledge_search '{"query":"ledger rollback","collection":"team-acme","top_k":3}')"
check "mdupont: ingest (denied, read-only key)" denied "$(mcp "$MD" team/acme ingest_document '{"document_id":"m-1","content":"x","collection":"team-acme"}')"
check "kmlead: save into org/secrets" ok "$(mcp "$KM" org/secrets knowledge_save '{"knowledge_id":"secret-doc","content":"The secret retainer is forty thousand.","collection":"secrets"}')"
check "kmlead: key of org/secrets on team/acme (403)" denied "$(mcp "$KM" team/acme knowledge_search '{"query":"x","collection":"team-acme"}')"
check "operator key is not a gateway key (401)" denied "$(mcp op-secret team/acme knowledge_search '{"query":"x","collection":"team-acme"}')"
check "unknown brain (404)" denied "$(mcp "$LF" team/acm knowledge_search '{"query":"x","collection":"team-acme"}')"
check "lfoletti: secret-doc by id as team-acme (not found)" denied "$(mcp "$LF" team/acme knowledge_update '{"knowledge_id":"secret-doc","collection":"team-acme","content":"overwritten"}')"
check "lfoletti: graph_statistics (denied, not in the allowlist)" denied "$(mcp "$LF" team/acme graph_statistics '{}')"

# The federated endpoint: any gateway key names the caller; th_brains lists
# the brains, th_recall asks each one with the caller's own key.
brains=$(mcp "$LF" "" th_brains '{}')
if printf '%s' "$brains" | grep -q '"brain\\":\\"org/secrets\\",\\"access\\":\\"read\\",\\"key\\":true' && printf '%s' "$brains" | grep -q 'team/acme\\",\\"access\\":\\"write\\",\\"key\\":true'; then ok "th_brains lists both brains with their keys"; else ko "th_brains" "$brains"; fi
rec=$(mcp "$LF" "" th_recall '{"query":"ledger rollback","top_k":3}')
if printf '%s' "$rec" | grep -q 'spec-1' && printf '%s' "$rec" | grep -q '\[team/acme\] Spec migration, spec-1'; then ok "th_recall finds spec-1 in team/acme with its provenance"; else ko "th_recall ledger" "$rec"; fi
rec=$(mcp "$LF" "" th_recall '{"query":"secret retainer"}')
if printf '%s' "$rec" | grep -q 'secret-doc' && printf '%s' "$rec" | grep -q '\[org/secrets\]'; then ok "th_recall finds secret-doc in org/secrets with lfoletti's read key"; else ko "th_recall retainer" "$rec"; fi
check "operator key on /mcp (401)" denied "$(mcp op-secret "" th_brains '{}')"

# The write circuit: a writer proposes, the review gate hides the document,
# the curator lists what waits and accepts, the document is served.
add=$(mcp "$LF" "" th_add '{"brain":"team/acme","document_id":"spec-3","title":"Spec 3","content":"The cutover rehearsal is on Thursday evening with the whole team.","source_url":"smb://share/spec3.docx"}')
if printf '%s' "$add" | grep -q 'spec-3' && ! printf '%s' "$add" | grep -q isError; then ok "th_add proposes spec-3 into team/acme"; else ko "th_add" "$add"; fi
hid=$(mcp "$LF" team/acme knowledge_search '{"query":"cutover rehearsal","collection":"team-acme","top_k":5}')
if printf '%s' "$hid" | grep -q 'spec-3'; then ko "proposed spec-3 served before review" "$hid"; else ok "proposed spec-3 is hidden until review"; fi
check "mdupont (reader): th_add refused" denied "$(mcp "$MD" "" th_add '{"brain":"team/acme","content":"x"}')"
pend=$(mcp "$KMA" "" th_pending '{"brain":"team/acme"}')
printf '%s' "$pend" | grep -q 'spec-3' && ok "th_pending lists spec-3 for the curator" || ko "th_pending" "$pend"
check "lfoletti (writer): th_pending refused" denied "$(mcp "$LF" "" th_pending '{"brain":"team/acme"}')"
rev=$(mcp "$KMA" "" th_review '{"brain":"team/acme","knowledge_id":"spec-3","decision":"accepted"}')
if printf '%s' "$rev" | grep -q isError; then ko "th_review" "$rev"; else ok "th_review accepts spec-3"; fi
srv=$(mcp "$LF" team/acme knowledge_search '{"query":"cutover rehearsal","collection":"team-acme","top_k":5}')
printf '%s' "$srv" | grep -q 'spec-3' && ok "accepted spec-3 is served" || ko "accepted spec-3 missing" "$srv"
got=$(mcp "$KMA" team/acme knowledge_get '{"knowledge_id":"spec-3"}')
printf '%s' "$got" | grep -q 'reviewed_by.*th/kmlead/team/acme' && ok "the engine recorded the curator's key as reviewer" || ko "reviewer" "$got"

# The client: the plugin's hook asks th_recall with the employee's key and
# hands the extracts to Claude Code; a prompt too short asks nothing.
ctx=$(printf '%s' '{"prompt":"when does the ledger move to the new cluster?","cwd":"/tmp"}' | TH_GATEWAY="http://$GWADDR" TH_KEY="$LF" TH_HOOK_BIN="$S/th-hook" "$ROOT/plugin/bin/th-recall")
if printf '%s' "$ctx" | grep -q '"hookEventName":"UserPromptSubmit"' && printf '%s' "$ctx" | grep -q '\[team/acme\] Spec migration, spec-1'; then ok "the hook injects spec-1 with its provenance"; else ko "hook" "$ctx"; fi
ctx=$(printf '%s' '{"prompt":"ok"}' | TH_GATEWAY="http://$GWADDR" TH_KEY="$LF" TH_HOOK_BIN="$S/th-hook" "$ROOT/plugin/bin/th-recall")
[ -z "$ctx" ] && ok "a short prompt injects nothing" || ko "short prompt" "$ctx"
ctx=$(printf '%s' '{"prompt":"when does the ledger move to the new cluster?"}' | TH_GATEWAY="http://$GWADDR" TH_KEY="nope" TH_HOOK_BIN="$S/th-hook" "$ROOT/plugin/bin/th-recall" 2>/dev/null; echo "exit=$?")
[ "$ctx" = "exit=0" ] && ok "a refused key blocks nothing (exit 0, no output)" || ko "refused key" "$ctx"

# Revocation: the file changes at once; the server reads it at start, so the
# check here is on the file, not on the server.
"$GW" key revoke "$S/data/rights.yaml" mdupont team/acme >/dev/null && ok "revoke mdupont" || ko "revoke mdupont" ""
if grep -q 'th/mdupont/team/acme' "$S/data/keys.json"; then ko "mdupont key still in the file" ""; else ok "mdupont key removed from the file"; fi
check "mdupont: revoked key (401 at the gateway)" denied "$(mcp "$MD" team/acme knowledge_search '{"query":"ledger","collection":"team-acme"}')"

# The REST face, for a browser extension or a script: same keys, same
# circuit. GET /v1/brains lists, POST /v1/add proposes, refusals by status.
rest() { # rest <method> <path> <secret> <body>
  curl -s -o "$S/rest.out" -w '%{http_code}' -X "$1" "http://$GWADDR$2" ${3:+-H "Authorization: Bearer $3"} -H 'Content-Type: application/json' ${4:+--data "$4"}
}
code=$(rest GET /v1/brains "$LF" "")
if [ "$code" = 200 ] && grep -q '"brain":"team/acme".*"access":"write".*"key":true' "$S/rest.out" && grep -q '"brain":"org/secrets".*"access":"read"' "$S/rest.out"; then ok "GET /v1/brains lists lfoletti's two brains"; else ko "GET /v1/brains" "$code $(cat "$S/rest.out")"; fi
code=$(rest POST /v1/add "$LF" '{"brain":"team/acme","document_id":"spec-4","title":"Spec 4","content":"The REST face proposes a document into the brain.","source_url":"https://intranet/spec4"}')
if [ "$code" = 200 ] && grep -q 'spec-4' "$S/rest.out"; then ok "POST /v1/add proposes spec-4 as lfoletti"; else ko "POST /v1/add" "$code $(cat "$S/rest.out")"; fi
code=$(rest POST /v1/add "$LFS" '{"brain":"org/secrets","content":"x"}')
[ "$code" = 403 ] && ok "POST /v1/add as a reader is 403" || ko "reader add" "$code $(cat "$S/rest.out")"
code=$(rest POST /v1/add "" '{"brain":"team/acme","content":"x"}')
[ "$code" = 401 ] && ok "POST /v1/add without a key is 401" || ko "no key add" "$code $(cat "$S/rest.out")"
code=$(rest POST /v1/add "$LF" '{"brain":"org/none","content":"x"}')
[ "$code" = 404 ] && ok "POST /v1/add on an unknown brain is 404" || ko "unknown brain add" "$code $(cat "$S/rest.out")"

# The rights file changes while the gateway runs: a withdrawn right is
# refused at the next call, no restart; a restored right is served again.
# lfoletti reads org/secrets by name only, so one line withdraws it all.
sed -i '' 's/    readers: \[user:lfoletti\]/    readers: []/' "$S/data/rights.yaml"
check "lfoletti: right on org/secrets withdrawn in rights.yaml (403, no restart)" denied "$(mcp "$LFS" org/secrets knowledge_search '{"query":"secret","collection":"secrets"}')"
sed -i '' 's/    readers: \[\]/    readers: [user:lfoletti]/' "$S/data/rights.yaml"
check "lfoletti: right on org/secrets restored in rights.yaml (served again)" ok "$(mcp "$LFS" org/secrets knowledge_search '{"query":"secret","collection":"secrets"}')"

# The administration: key list shows who holds what against the rights,
# key sync brings the file back to the rights (mdupont's key is reissued).
lst=$("$GW" key list "$S/data/rights.yaml")
if printf '%s' "$lst" | grep -q '^team/acme  lfoletti  read-write  ok$' && printf '%s' "$lst" | grep -q '^org/secrets  kmlead  read-write  ok$' && ! printf '%s' "$lst" | grep -q mdupont && printf '%s' "$lst" | grep -q '^1 other key(s) in '; then ok "key list shows the keys against the rights"; else ko "key list" "$lst"; fi
plan=$("$GW" key sync "$S/data/rights.yaml" --dry-run)
[ "$plan" = "would issue th/mdupont/team/acme" ] && ok "key sync --dry-run plans mdupont's key only" || ko "key sync dry run" "$plan"
done_=$("$GW" key sync "$S/data/rights.yaml")
printf '%s' "$done_" | grep -q '^issued th/mdupont/team/acme ' && ok "key sync reissues mdupont's key" || ko "key sync" "$done_"
[ "$("$GW" key sync "$S/data/rights.yaml")" = "nothing to do" ] && ok "key sync is idempotent" || ko "key sync again" ""

echo
echo "--- audit ---"
n=$(grep -c '"tool"' "$S/data/audit.jsonl"); echo "lines: $n"
[ "$n" -eq 20 ] && ok "twenty tool calls audited (the twelve refusals at the gateway are not)" || ko "audit lines: $n" "$(cat "$S/data/audit.jsonl")"
grep -q '"tool":"th_recall","ids":{"brains":\["org/secrets","team/acme"\],"knowledge_ids":\[.*team/acme:spec-1' "$S/data/audit.jsonl" && ok "th_recall audited with the brains and the ids read" || ko "th_recall audit line" "$(grep th_recall "$S/data/audit.jsonl" | head -1)"
grep -q '"user":"lfoletti","key":"th/lfoletti/team/acme","brain":"team/acme","tool":"ingest_document","ids":{"document_id":"spec-1"},"ok":true' "$S/data/audit.jsonl" && ok "ingest audited with its id" || ko "ingest audit line" "$(grep ingest_document "$S/data/audit.jsonl" | head -1)"
grep -q '"tool":"graph_statistics","ok":false,"error":.*PermissionDenied' "$S/data/audit.jsonl" && ok "refusal by the brain audited as an error" || ko "graph_statistics audit line" "$(grep graph_statistics "$S/data/audit.jsonl")"
grep -q '"user":"kmlead".*"brain":"org/secrets".*"knowledge_id":"secret-doc"' "$S/data/audit.jsonl" && ok "kmlead's save audited on org/secrets" || ko "kmlead audit line" "$(grep kmlead "$S/data/audit.jsonl")"
grep -q '"user":"kmlead".*"brain":"team/acme","tool":"th_review","ids":{"decision":"accepted","knowledge_id":"spec-3"},"ok":true' "$S/data/audit.jsonl" && ok "the review is audited with its decision" || ko "th_review audit line" "$(grep th_review "$S/data/audit.jsonl")"
echo
echo "pass=$pass fail=$fail"
[ "$fail" -eq 0 ]
