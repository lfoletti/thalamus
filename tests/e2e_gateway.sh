#!/bin/sh
# End-to-end trial of gc-gateway in front of one shared brain: a cortexdb-grpc
# server built from the cortexdb/ branch, a rights file with two brains on it,
# keys issued by the gateway, tool calls over MCP streamable HTTP with curl,
# and the audit file read back. Prints PASS/FAIL per check.
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
S=$(mktemp -d)
(cd "$ROOT/cortexdb" && go build -o "$S/cortexdb-grpc" ./cmd/cortexdb-grpc) || exit 1
(cd "$ROOT/gateway" && go build -o "$S/gc-gateway" ./cmd/gc-gateway && go build -o "$S/gc-hook" ./cmd/gc-hook) || exit 1
GRPC=$S/cortexdb-grpc
GW=$S/gc-gateway
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

# The federated endpoint: any gateway key names the caller; gc_brains lists
# the brains, gc_recall asks each one with the caller's own key.
brains=$(mcp "$LF" "" gc_brains '{}')
if printf '%s' "$brains" | grep -q '"brain\\":\\"org/secrets\\",\\"access\\":\\"read\\",\\"key\\":true' && printf '%s' "$brains" | grep -q 'team/acme\\",\\"access\\":\\"write\\",\\"key\\":true'; then ok "gc_brains lists both brains with their keys"; else ko "gc_brains" "$brains"; fi
rec=$(mcp "$LF" "" gc_recall '{"query":"ledger rollback","top_k":3}')
if printf '%s' "$rec" | grep -q 'spec-1' && printf '%s' "$rec" | grep -q '\[team/acme\] Spec migration, spec-1'; then ok "gc_recall finds spec-1 in team/acme with its provenance"; else ko "gc_recall ledger" "$rec"; fi
rec=$(mcp "$LF" "" gc_recall '{"query":"secret retainer"}')
if printf '%s' "$rec" | grep -q 'secret-doc' && printf '%s' "$rec" | grep -q '\[org/secrets\]'; then ok "gc_recall finds secret-doc in org/secrets with lfoletti's read key"; else ko "gc_recall retainer" "$rec"; fi
check "operator key on /mcp (401)" denied "$(mcp op-secret "" gc_brains '{}')"

# The client: the plugin's hook asks gc_recall with the employee's key and
# hands the extracts to Claude Code; a prompt too short asks nothing.
ctx=$(printf '%s' '{"prompt":"when does the ledger move to the new cluster?","cwd":"/tmp"}' | GC_GATEWAY="http://$GWADDR" GC_KEY="$LF" GC_HOOK_BIN="$S/gc-hook" "$ROOT/plugin/bin/gc-recall")
if printf '%s' "$ctx" | grep -q '"hookEventName":"UserPromptSubmit"' && printf '%s' "$ctx" | grep -q '\[team/acme\] Spec migration, spec-1'; then ok "the hook injects spec-1 with its provenance"; else ko "hook" "$ctx"; fi
ctx=$(printf '%s' '{"prompt":"ok"}' | GC_GATEWAY="http://$GWADDR" GC_KEY="$LF" GC_HOOK_BIN="$S/gc-hook" "$ROOT/plugin/bin/gc-recall")
[ -z "$ctx" ] && ok "a short prompt injects nothing" || ko "short prompt" "$ctx"
ctx=$(printf '%s' '{"prompt":"when does the ledger move to the new cluster?"}' | GC_GATEWAY="http://$GWADDR" GC_KEY="nope" GC_HOOK_BIN="$S/gc-hook" "$ROOT/plugin/bin/gc-recall" 2>/dev/null; echo "exit=$?")
[ "$ctx" = "exit=0" ] && ok "a refused key blocks nothing (exit 0, no output)" || ko "refused key" "$ctx"

# Revocation: the file changes at once; the server reads it at start, so the
# check here is on the file, not on the server.
"$GW" key revoke "$S/data/rights.yaml" mdupont team/acme >/dev/null && ok "revoke mdupont" || ko "revoke mdupont" ""
if grep -q 'gc/mdupont/team/acme' "$S/data/keys.json"; then ko "mdupont key still in the file" ""; else ok "mdupont key removed from the file"; fi
check "mdupont: revoked key (401 at the gateway)" denied "$(mcp "$MD" team/acme knowledge_search '{"query":"ledger","collection":"team-acme"}')"

echo
echo "--- audit ---"
n=$(grep -c '"tool"' "$S/data/audit.jsonl"); echo "lines: $n"
[ "$n" -eq 11 ] && ok "eleven tool calls audited (the six refusals at the gateway are not)" || ko "audit lines: $n" "$(cat "$S/data/audit.jsonl")"
grep -q '"tool":"gc_recall","ids":{"brains":\["org/secrets","team/acme"\],"knowledge_ids":\[.*team/acme:spec-1' "$S/data/audit.jsonl" && ok "gc_recall audited with the brains and the ids read" || ko "gc_recall audit line" "$(grep gc_recall "$S/data/audit.jsonl" | head -1)"
grep -q '"user":"lfoletti","key":"gc/lfoletti/team/acme","brain":"team/acme","tool":"ingest_document","ids":{"document_id":"spec-1"},"ok":true' "$S/data/audit.jsonl" && ok "ingest audited with its id" || ko "ingest audit line" "$(grep ingest_document "$S/data/audit.jsonl" | head -1)"
grep -q '"tool":"graph_statistics","ok":false,"error":.*PermissionDenied' "$S/data/audit.jsonl" && ok "refusal by the brain audited as an error" || ko "graph_statistics audit line" "$(grep graph_statistics "$S/data/audit.jsonl")"
grep -q '"user":"kmlead".*"brain":"org/secrets".*"knowledge_id":"secret-doc"' "$S/data/audit.jsonl" && ok "kmlead's save audited on org/secrets" || ko "kmlead audit line" "$(grep kmlead "$S/data/audit.jsonl")"
echo
echo "pass=$pass fail=$fail"
[ "$fail" -eq 0 ]
