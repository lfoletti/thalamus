#!/bin/sh
# End-to-end trial of the cortexdb/ branch, as the shared brain of granular_cortex:
# one cortexdb-grpc server, a key file with confined keys, and the MCP stdio
# client in remote mode driven over JSON-RPC. Prints PASS/FAIL per check.
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
S=$(mktemp -d)
(cd "$ROOT/cortexdb" && go build -o "$S/cortexdb-grpc" ./cmd/cortexdb-grpc && go build -o "$S/cortexdb-mcp-stdio" ./cmd/cortexdb-mcp-stdio) || exit 1
E=$S
GRPC=$S/cortexdb-grpc
STDIO=$S/cortexdb-mcp-stdio
ADDR=127.0.0.1:47999
rm -rf "$E/data"; mkdir -p "$E/data"

cat > "$E/data/keys.json" <<'EOF'
{"keys":[
 {"id":"operator","secret":"op-secret","clearance":"read-write"},
 {"id":"lf","secret":"lf-secret","clearance":"read-write","scope":{"collections":["team-acme","user-lf"]}},
 {"id":"reader","secret":"rd-secret","clearance":"read-only","scope":{"collection":"team-acme"}},
 {"id":"hermes","secret":"hermes-secret","clearance":"read-write","scope":{"user_id":"hermes"}}
]}
EOF

# The shared server opts in to the review gate: a proposed fact is hidden
# until a reviewer accepts it. Without CORTEXDB_REVIEW_GATE every row is served.
CORTEXDB_TRACE=1 CORTEXDB_TRACE_FILE="$E/data/trace.jsonl" CORTEXDB_REVIEW_GATE=1 \
  "$GRPC" -db "$E/data/brain.db" -addr "$ADDR" -keys "$E/data/keys.json" > "$E/data/server.log" 2>&1 &
SERVER=$!
trap 'kill $SERVER 2>/dev/null' EXIT
i=0; until "$GRPC" -addr "$ADDR" -health >/dev/null 2>&1 || [ $i -ge 50 ]; do i=$((i+1)); sleep 0.1; done

# mcp <token> <tool> <args-json>  -> prints the tool result text (one line)
mcp() {
  tok=$1; tool=$2; args=$3
  {
    printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"e2e","version":"0"}}}'
    printf '%s\n' '{"jsonrpc":"2.0","method":"notifications/initialized"}'
    printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"%s","arguments":%s}}\n' "$tool" "$args"
    # The stdio server stops on end of input: keep it open until it answered.
    sleep 1.5
  } | CORTEXDB_REMOTE=$ADDR CORTEXDB_GRPC_TOKEN=$tok "$STDIO" 2>&1 | grep -E '"id":2|discover tools|Unauthenticated' | head -1
}

pass=0; fail=0
check() { # check <name> <expect: ok|denied> <output>
  name=$1; expect=$2; out=$3
  if [ -z "$out" ]; then got=empty
  elif printf '%s' "$out" | grep -q 'PermissionDenied\|denied\|Unauthenticated\|discover tools'; then got=denied
  else got=ok; fi
  if [ "$got" = "$expect" ]; then pass=$((pass+1)); echo "PASS  $name ($got)"; else fail=$((fail+1)); echo "FAIL  $name: want $expect, got $got"; echo "      $(printf '%s' "$out" | cut -c1-300)"; fi
}

check "operator: graph_statistics" ok "$(mcp op-secret graph_statistics '{}')"
check "lf: ingest into team-acme" ok "$(mcp lf-secret ingest_document '{"document_id":"spec-1","title":"Spec migration","content":"Acme migrates the ledger to the new cluster on Sunday night. The rollback window is two hours.","collection":"team-acme","source_url":"smb://share/spec.docx","author":"mdupont","metadata":{"classification":"internal"},"expires_at":"2031-10-18T00:00:00Z"}')"
check "lf: ingest into user-lf" ok "$(mcp lf-secret ingest_document '{"document_id":"note-1","content":"Personal note: ask Acme about the retainer.","collection":"user-lf"}')"
check "lf: ingest into secrets (denied)" denied "$(mcp lf-secret ingest_document '{"document_id":"s-1","content":"Not yours.","collection":"secrets"}')"
check "lf: ingest without collection (denied)" denied "$(mcp lf-secret ingest_document '{"document_id":"s-2","content":"No collection."}')"
check "lf: knowledge_search team-acme" ok "$(mcp lf-secret knowledge_search '{"query":"ledger rollback","collection":"team-acme","top_k":3}')"
check "lf: knowledge_search user-lf" ok "$(mcp lf-secret knowledge_search '{"query":"retainer","collection":"user-lf","top_k":3}')"
check "lf: knowledge_search secrets (denied)" denied "$(mcp lf-secret knowledge_search '{"query":"ledger","collection":"secrets"}')"
check "lf: graph_statistics (denied, no collection arg)" denied "$(mcp lf-secret graph_statistics '{}')"
check "lf: re-ingest identical (unchanged)" ok "$(mcp lf-secret ingest_document '{"document_id":"spec-1","title":"Spec migration","content":"Acme migrates the ledger to the new cluster on Sunday night. The rollback window is two hours.","collection":"team-acme","source_url":"smb://share/spec.docx","author":"mdupont","metadata":{"classification":"internal"},"expires_at":"2031-10-18T00:00:00Z"}')"
unchanged=$(mcp lf-secret ingest_document '{"document_id":"spec-1","title":"Spec migration","content":"Acme migrates the ledger to the new cluster on Sunday night. The rollback window is two hours.","collection":"team-acme","source_url":"smb://share/spec.docx","author":"mdupont","metadata":{"classification":"internal"},"expires_at":"2031-10-18T00:00:00Z"}')
if printf '%s' "$unchanged" | grep -q 'unchanged'; then pass=$((pass+1)); echo "PASS  lf: unchanged reported"; else fail=$((fail+1)); echo "FAIL  lf: unchanged not reported: $(printf '%s' "$unchanged" | cut -c1-300)"; fi
dup=$(mcp lf-secret ingest_document '{"document_id":"spec-copy","content":"Acme migrates the ledger to the new cluster on Sunday night. The rollback window is two hours.","collection":"team-acme","dedupe":true}')
if printf '%s' "$dup" | grep -q 'duplicate_of'; then pass=$((pass+1)); echo "PASS  lf: dedupe names spec-1"; else fail=$((fail+1)); echo "FAIL  lf: dedupe: $(printf '%s' "$dup" | cut -c1-300)"; fi
check "reader: knowledge_search team-acme" ok "$(mcp rd-secret knowledge_search '{"query":"ledger","collection":"team-acme","top_k":3}')"
check "reader: ingest (denied, read-only)" denied "$(mcp rd-secret ingest_document '{"document_id":"r-1","content":"x","collection":"team-acme"}')"
check "hermes: memory_save own" ok "$(mcp hermes-secret memory_save '{"memory_id":"h-1","content":"Hermes prefers Friday deploys.","user_id":"hermes","scope":"user"}')"
check "hermes: memory_search own" ok "$(mcp hermes-secret memory_search '{"query":"deploys","user_id":"hermes"}')"
check "hermes: memory_search zeus (denied)" denied "$(mcp hermes-secret memory_search '{"query":"deploys","user_id":"zeus"}')"
check "hermes: knowledge_search (denied, no user_id arg)" denied "$(mcp hermes-secret knowledge_search '{"query":"ledger","collection":"team-acme"}')"
check "wrong token (denied)" denied "$(mcp nope knowledge_search '{"query":"x","collection":"team-acme"}')"
prop=$(mcp lf-secret knowledge_save '{"knowledge_id":"fact-1","content":"Acme signed the renewal on 2026-10-01.","collection":"team-acme","metadata":{"review_status":"proposed"}}')
check "lf: knowledge_save proposed" ok "$prop"
hidden=$(mcp lf-secret knowledge_search '{"query":"renewal signed","collection":"team-acme","top_k":5}')
if printf '%s' "$hidden" | grep -q 'fact-1'; then fail=$((fail+1)); echo "FAIL  proposed fact served: $(printf '%s' "$hidden" | cut -c1-300)"; else pass=$((pass+1)); echo "PASS  proposed fact hidden from search"; fi
waiting=$(mcp lf-secret knowledge_search '{"query":"renewal signed","collection":"team-acme","top_k":5,"plan":{"filters":{"review_status":"proposed"}}}')
if printf '%s' "$waiting" | grep -q 'fact-1'; then pass=$((pass+1)); echo "PASS  review_status=proposed lists fact-1"; else fail=$((fail+1)); echo "FAIL  proposed filter: $(printf '%s' "$waiting" | cut -c1-300)"; fi
acc=$(mcp lf-secret knowledge_update '{"knowledge_id":"fact-1","collection":"team-acme","metadata":{"review_status":"accepted","reviewed_by":"mdupont"}}')
check "lf: accept fact-1" ok "$acc"
served=$(mcp lf-secret knowledge_search '{"query":"renewal signed","collection":"team-acme","top_k":5}')
if printf '%s' "$served" | grep -q 'fact-1'; then pass=$((pass+1)); echo "PASS  accepted fact served"; else fail=$((fail+1)); echo "FAIL  accepted fact missing: $(printf '%s' "$served" | cut -c1-300)"; fi

echo
echo "--- trace (key= tool= on each CallTool) ---"
grep -c '"tool"' "$E/data/trace.jsonl" 2>/dev/null | sed 's/^/records with a tool name: /'
grep '"key":"lf"' "$E/data/trace.jsonl" | grep -c 'denied\|PermissionDenied' | sed 's/^/lf denials logged: /'
echo
echo "pass=$pass fail=$fail"

echo
echo "--- rows outside the scope, named by id ---"
mcp op-secret knowledge_save '{"knowledge_id":"secret-doc","content":"The secret retainer is forty thousand.","collection":"secrets"}' >/dev/null
mcp op-secret memory_save '{"memory_id":"zeus-1","content":"Zeus private note.","user_id":"zeus","scope":"user"}' >/dev/null
check "lf: knowledge_update secret-doc as team-acme (not found)" denied "$(mcp lf-secret knowledge_update '{"knowledge_id":"secret-doc","collection":"team-acme","content":"overwritten"}' | sed 's/NotFound/denied/')"
check "lf: ingest_document secret-doc as team-acme (not found)" denied "$(mcp lf-secret ingest_document '{"document_id":"secret-doc","collection":"team-acme","content":"overwritten"}' | sed 's/NotFound/denied/')"
check "lf: knowledge_save secret-doc as team-acme (not found)" denied "$(mcp lf-secret knowledge_save '{"knowledge_id":"secret-doc","collection":"team-acme","content":"overwritten"}' | sed 's/NotFound/denied/')"
check "hermes: memory_save zeus-1 as hermes (not found)" denied "$(mcp hermes-secret memory_save '{"memory_id":"zeus-1","content":"overwritten","user_id":"hermes","scope":"user"}' | sed 's/NotFound/denied/')"
back=$(mcp op-secret knowledge_get '{"knowledge_id":"secret-doc"}')
if printf '%s' "$back" | grep -q 'forty thousand' && printf '%s' "$back" | grep -q '\\"collection\\":\\"secrets\\"'; then pass=$((pass+1)); echo "PASS  secret-doc untouched"; else fail=$((fail+1)); echo "FAIL  secret-doc changed: $(printf '%s' "$back" | cut -c1-300)"; fi
back=$(mcp op-secret memory_get '{"memory_id":"zeus-1"}')
if printf '%s' "$back" | grep -q 'Zeus private' && printf '%s' "$back" | grep -q '\\"user_id\\":\\"zeus\\"'; then pass=$((pass+1)); echo "PASS  zeus-1 untouched"; else fail=$((fail+1)); echo "FAIL  zeus-1 changed: $(printf '%s' "$back" | cut -c1-300)"; fi
echo "pass=$pass fail=$fail"
