#!/bin/sh
# End-to-end trial of the Claude Code plugin in a real, unattended session:
# a cortexdb-grpc server built from cortexdb/, th-gateway in front, one
# document in the brain, then `claude -p --plugin-dir plugin` twice from an
# empty directory. The first prompt asks a question whose answer sits only
# in the shared brain and in a local brain of the machine: the hook must
# inject both, the whole passage, with the synthesis of the pyramid first.
# The second asks for th_brains: the MCP server of the plugin must answer.
# The third asks what the source of a synthesis says: the answer is only in
# a document the synthesis cites, so the session must call th_descend. The
# audit file proves the reads (team/acme has audit_reads).
# Usage: sh tests/e2e_plugin.sh [model]   (default: haiku; three short calls; 11 checks)
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
MODEL=${1:-haiku}
S=$(mktemp -d)
pass=0; fail=0
ok() { pass=$((pass+1)); echo "PASS  $1"; }
ko() { fail=$((fail+1)); echo "FAIL  $1: $2" | cut -c1-600; }
(cd "$ROOT/cortexdb" && go build -o "$S/cortexdb-grpc" ./cmd/cortexdb-grpc && go build -o "$S/cortexdb-mcp" ./cmd/cortexdb-mcp-stdio) || exit 1
(cd "$ROOT/gateway" && go build -o "$S/th-gateway" ./cmd/th-gateway && go build -o "$S/th-hook" ./cmd/th-hook) || exit 1
mkdir -p "$S/data" "$S/work"
ADDR=127.0.0.1:$((20000 + $$ % 10000)); GWADDR=127.0.0.1:$((30000 + $$ % 10000))
cat > "$S/data/keys.json" <<K
{"keys":[{"id":"operator","secret":"op-secret-op-secret-op-secret-op","clearance":"read-write"}]}
K
cat > "$S/data/rights.yaml" <<R
brains:
  team/acme:
    addr: $ADDR
    keys_file: $S/data/keys.json
    collection: team-acme
    pyramid: {collection: acme-pyramid, sources: preuves}
    audit_reads: true
    readers: [group:acme]
    writers: [user:lfoletti]
groups:
  acme: [lfoletti, mdupont]
R
LF=$("$S/th-gateway" key issue "$S/data/rights.yaml" lfoletti team/acme) || { echo "key issue failed"; exit 1; }
"$S/cortexdb-grpc" -db "$S/data/brain.db" -addr "$ADDR" -keys "$S/data/keys.json" > "$S/data/server.log" 2>&1 &
SRV=$!
"$S/th-gateway" serve "$S/data/rights.yaml" --listen "$GWADDR" --audit "$S/data/audit.jsonl" > "$S/data/gateway.log" 2>&1 &
GWP=$!
trap 'kill $SRV $GWP 2>/dev/null; wait $SRV $GWP 2>/dev/null' EXIT
sleep 1.5
# One document, through the per-brain endpoint, as lfoletti (a writer).
init='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"bench","version":"0"}}}'
sid=$(curl -s -D - -o /dev/null -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $LF" "http://$GWADDR/mcp/team/acme" --data "$init" | tr -d '\r' | awk -F': ' 'tolower($1)=="mcp-session-id"{print $2}')
curl -s -o /dev/null -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $LF" -H "Mcp-Session-Id: $sid" "http://$GWADDR/mcp/team/acme" --data '{"jsonrpc":"2.0","method":"notifications/initialized"}'
ing=$(curl -s -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $LF" -H "Mcp-Session-Id: $sid" "http://$GWADDR/mcp/team/acme" --data '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"ingest_document","arguments":{"document_id":"cutover-plan","title":"Cutover plan","collection":"team-acme","source_url":"smb://share/cutover.docx","content":"The cutover rehearsal of the Acme ledger migration takes place on Thursday 5 November 2026 at 18:00 in the Geneva office, with the whole team. The code word for the rehearsal is MARMOT."}}}')
printf '%s' "$ing" | grep -q 'cutover-plan' && ok "document ingested in team/acme" || ko "ingest" "$ing"
# The pyramid: a synthesis that cites cutover-plan and a budget memo. The
# memo shares no word with the questions, so only th_descend reaches it.
call() { curl -s -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Authorization: Bearer $LF" -H "Mcp-Session-Id: $sid" "http://$GWADDR/mcp/team/acme" --data "$1"; }
memo=$(call '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"ingest_document","arguments":{"document_id":"budget-memo","title":"Finance committee memo","collection":"team-acme","content":"The finance committee approved 48000 CHF for the operation, signed on 2 November 2026."}}}')
syn=$(call '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"ingest_document","arguments":{"document_id":"pyramide/cutover","title":"Acme ledger migration","collection":"acme-pyramid","content":"The Acme ledger migration has a cutover rehearsal in the Geneva office and an approved budget; see the sources.","metadata":{"preuves":"[\"cutover-plan\", \"budget-memo\"]"}}}}')
printf '%s%s' "$memo" "$syn" | grep -q 'budget-memo.*pyramide/cutover' && ok "a synthesis in the pyramid of team/acme cites cutover-plan and budget-memo" || ko "pyramid" "$memo $syn"
: > "$S/data/audit.jsonl"

# One memory in a local brain of the machine, through the CortexDB stdio server.
{ printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"bench","version":"0"}}}' '{"jsonrpc":"2.0","method":"notifications/initialized"}' '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memory_save","arguments":{"memory_id":"fact:bench:cutover-room","content":"For the cutover rehearsal of the Acme ledger migration, my team meets in room Lavaux."}}}'; sleep 2; } | env -i HOME="$S" PATH="$PATH" CORTEXDB_PATH="$S/data/local.db" "$S/cortexdb-mcp" >/dev/null 2>&1
[ -s "$S/data/local.db" ] && ok "memory saved in the local brain" || ko "local brain" ""

export TH_GATEWAY="http://$GWADDR" TH_KEY="$LF" TH_HOOK_BIN="$S/th-hook" TH_RECALL_FILE="$S/data/recall"
export TH_LOCAL_RECALL="$S/cortexdb-mcp --recall" TH_LOCAL_BRAINS="projet=$S/data/local.db"
run_claude() { # run_claude <prompt> <out>
  (cd "$S/work" && claude -p "$1" --plugin-dir "$ROOT/plugin" --model "$MODEL" --max-turns 4 --output-format json \
     --allowedTools "mcp__plugin_thalamus_thalamus__th_brains" "mcp__plugin_thalamus_thalamus__th_recall" "mcp__plugin_thalamus_thalamus__th_get" "mcp__plugin_thalamus_thalamus__th_descend" "mcp__plugin_thalamus_thalamus__th_add" < /dev/null > "$2" 2> "$2.err")
}
# 1. The hook: the answer is only in the brains, no tool is named.
run_claude "What is the code word for the cutover rehearsal of the Acme ledger migration, on which day is it, and in which room does my team meet? Answer in one sentence." "$S/a1.json"
a1=$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1])).get("result",""))' "$S/a1.json" 2>/dev/null)
printf '%s' "$a1" | grep -qi 'MARMOT' && printf '%s' "$a1" | grep -qi 'thursday' && printf '%s' "$a1" | grep -qi 'Lavaux' && ok "the session answered from the brain: $(printf '%s' "$a1" | tr '\n' ' ' | cut -c1-120)" || ko "answer 1" "$a1 $(cat "$S/a1.json.err" | tail -3)"
grep -q '"user":"lfoletti".*"tool":"th_recall"' "$S/data/audit.jsonl" && ok "the hook's recall is audited under lfoletti" || ko "hook audit" "$(cat "$S/data/audit.jsonl")"
grep -q '"tool":"th_recall".*"team/acme:cutover-plan".*"team/acme:pyramide/cutover"' "$S/data/audit.jsonl" && ok "the hook served the passage and the synthesis" || ko "hook reads" "$(cat "$S/data/audit.jsonl")"
# 2. The MCP server of the plugin: a tool call on demand.
run_claude "Call the th_brains tool and reply with the exact list of brain names it returns, nothing else." "$S/a2.json"
a2=$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1])).get("result",""))' "$S/a2.json" 2>/dev/null)
printf '%s' "$a2" | grep -q 'team/acme' && ok "th_brains answered through the plugin's MCP server: $(printf '%s' "$a2" | tr '\n' ' ' | cut -c1-80)" || ko "answer 2" "$a2 $(cat "$S/a2.json.err" | tail -3)"
grep -q '"tool":"th_brains"' "$S/data/audit.jsonl" && ko "th_brains audit" "$(cat "$S/data/audit.jsonl")" || ok "th_brains reads no brain and leaves no audit line"
# 3. The descent: what the sources of a synthesis say.
run_claude "Thalamus has a synthesis about the Acme ledger migration. Open the sources it cites and tell me the money figure they give, in one sentence." "$S/a3.json"
a3=$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1])).get("result",""))' "$S/a3.json" 2>/dev/null)
printf '%s' "$a3" | grep -q '48' && ok "the session answered from a source of the synthesis: $(printf '%s' "$a3" | tr '\n' ' ' | cut -c1-120)" || ko "answer 3" "$a3 $(cat "$S/a3.json.err" | tail -3)"
tail -n 2 "$S/data/audit.jsonl" | head -n 1 | grep -q 'budget-memo' && ko "the hook already served budget-memo" "$(tail -n 2 "$S/data/audit.jsonl")" || ok "the hook did not serve budget-memo: only the descent reaches it"
grep -q '"tool":"th_descend".*"budget-memo"' "$S/data/audit.jsonl" && ok "the session opened the sources with th_descend" || ko "th_descend audit" "$(cat "$S/data/audit.jsonl")"
echo; echo "audit:"; cat "$S/data/audit.jsonl" | cut -c1-200
echo; echo "pass=$pass fail=$fail (model $MODEL, work in $S)"
[ "$fail" -eq 0 ]
