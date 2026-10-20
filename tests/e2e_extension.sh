#!/bin/sh
# End-to-end trial of the Chrome extension: a cortexdb-grpc server built
# from cortexdb/, th-gateway in front, a local web page, and Chromium
# (Playwright) that loads the extension, sets its options, and proposes
# the page, then a selection, into team/acme. The audit file and th_get
# prove what arrived. Needs `npm install` in extension/ once.
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
S=$(mktemp -d)
pass=0; fail=0
ok() { pass=$((pass+1)); echo "PASS  $1"; }
ko() { fail=$((fail+1)); echo "FAIL  $1: $2" | cut -c1-600; }
[ -d "$ROOT/extension/node_modules/playwright" ] || (cd "$ROOT/extension" && npm install --no-audit --no-fund >/dev/null) || exit 1
(cd "$ROOT/cortexdb" && go build -o "$S/cortexdb-grpc" ./cmd/cortexdb-grpc) || exit 1
(cd "$ROOT/gateway" && go build -o "$S/th-gateway" ./cmd/th-gateway) || exit 1
mkdir -p "$S/data" "$S/site"
ADDR=127.0.0.1:$((20000 + $$ % 10000)); GWADDR=127.0.0.1:$((30000 + $$ % 10000)); SITEPORT=$((40000 + $$ % 10000)); IDPPORT=$((50000 + $$ % 10000))
cat > "$S/data/keys.json" <<K
{"keys":[{"id":"operator","secret":"op-secret-op-secret-op-secret-op","clearance":"read-write"}]}
K
cat > "$S/data/rights.yaml" <<R
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
groups:
  acme: [lfoletti, mdupont]
identity:
  issuer: http://127.0.0.1:$IDPPORT
  audience: thalamus
  client_id: thalamus
  strip_domain: true
R
LF=$("$S/th-gateway" key issue "$S/data/rights.yaml" lfoletti team/acme) || { echo "key issue failed"; exit 1; }
LFS=$("$S/th-gateway" key issue "$S/data/rights.yaml" lfoletti org/secrets) || exit 1
cat > "$S/site/index.html" <<H
<!doctype html><meta charset="utf-8"><title>Cutover plan</title>
<h1>Cutover plan</h1>
<p>The cutover rehearsal of the Acme ledger migration takes place on Thursday 5 November 2026 at 18:00 in the Geneva office.</p>
<p id="sel">The code word for the rehearsal is MARMOT.</p>
H
"$S/cortexdb-grpc" -db "$S/data/brain.db" -addr "$ADDR" -keys "$S/data/keys.json" > "$S/data/server.log" 2>&1 &
SRV=$!
"$S/th-gateway" serve "$S/data/rights.yaml" --listen "$GWADDR" --audit "$S/data/audit.jsonl" > "$S/data/gateway.log" 2>&1 &
GWP=$!
(cd "$S/site" && python3 -m http.server "$SITEPORT" --bind 127.0.0.1 > /dev/null 2>&1) &
WEB=$!
node "$ROOT/tests/fake_idp.mjs" --port "$IDPPORT" --client-id thalamus --audience thalamus > "$S/data/idp.log" 2>&1 &
IDP=$!
trap 'kill $SRV $GWP $WEB $IDP 2>/dev/null; wait $SRV $GWP $WEB $IDP 2>/dev/null' EXIT
sleep 1.5

(cd "$ROOT/extension" && EXT="$ROOT/extension" SITE="http://127.0.0.1:$SITEPORT/" GATEWAY="http://$GWADDR" KEY="$LFS" node "$ROOT/extension/e2e_driver.mjs") > "$S/driver.out" 2> "$S/driver.err"
cat "$S/driver.out"
pass=$((pass + $(grep -c '^PASS' "$S/driver.out"))); fail=$((fail + $(grep -c '^FAIL' "$S/driver.out")))
[ -s "$S/driver.err" ] && { echo "driver stderr:"; tail -5 "$S/driver.err"; }

# What arrived: two th_add lines for the same document (same page, same
# id), the second with the selection only.
n=$(grep -c '"tool":"th_add".*"document_id":"web-' "$S/data/audit.jsonl")
[ "$n" -eq 2 ] && ok "two proposals audited under lfoletti for the page's id" || ko "audit" "$(cat "$S/data/audit.jsonl")"
id=$(grep -o '"document_id":"web-[0-9a-f]*"' "$S/data/audit.jsonl" | head -1 | cut -d'"' -f4)
init='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"bench","version":"0"}}}'
sid=$(curl -s -D - -o /dev/null -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $LF" "http://$GWADDR/mcp" --data "$init" | tr -d '\r' | awk -F': ' 'tolower($1)=="mcp-session-id"{print $2}')
curl -s -o /dev/null -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $LF" -H "Mcp-Session-Id: $sid" "http://$GWADDR/mcp" --data '{"jsonrpc":"2.0","method":"notifications/initialized"}'
got=$(curl -s -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $LF" -H "Mcp-Session-Id: $sid" "http://$GWADDR/mcp" --data "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"th_get\",\"arguments\":{\"brain\":\"team/acme\",\"knowledge_id\":\"$id\"}}}")
if printf '%s' "$got" | grep -q 'The code word for the rehearsal is MARMOT' && printf '%s' "$got" | grep -q 'kind.*selection' && printf '%s' "$got" | grep -q 'review_status.*proposed' && printf '%s' "$got" | grep -q 'author.*lfoletti' && printf '%s' "$got" | grep -q "source_url.*127.0.0.1:$SITEPORT"; then ok "the document in the brain is the selection, proposed, by lfoletti, with the page as source"; else ko "document" "$got"; fi
echo; echo "pass=$pass fail=$fail (work in $S)"
[ "$fail" -eq 0 ]
