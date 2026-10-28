#!/bin/sh
# End-to-end trial of th-gateway in front of one shared brain: a cortexdb-grpc
# server built from the cortexdb/ branch, a rights file with two brains on it,
# keys issued by the gateway, tool calls over MCP streamable HTTP with curl,
# and the audit file read back. Prints PASS/FAIL per check.
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
. "$ROOT/tests/cortexdb_clone.sh"
S=$(mktemp -d)
(cd "$CDB" && go build -o "$S/cortexdb-grpc" ./cmd/cortexdb-grpc && go build -o "$S/cortexdb-mcp" ./cmd/cortexdb-mcp-stdio) || exit 1
(cd "$ROOT/gateway" && go build -o "$S/th-gateway" ./cmd/th-gateway && go build -o "$S/th-hook" ./cmd/th-hook) || exit 1
# The bench owns its environment: no setting of this machine's Thalamus
# (TH_*, ~/.thalamus) reaches the hook or the gateway. HOME changes after
# the builds, which use the Go caches under the real HOME.
unset TH_GATEWAY TH_KEY TH_ENV TH_RECALL_FILE TH_RECALL_TIMEOUT TH_RECALL_TOPK TH_LOCAL_RECALL TH_LOCAL_BRAINS TH_HOOK_BIN
HOME=$S; export HOME
GRPC=$S/cortexdb-grpc
GW=$S/th-gateway
ADDR=127.0.0.1:47998
GWADDR=127.0.0.1:47997
mkdir -p "$S/data"

# The operator key is the server's own; the gateway adds the users' keys
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
    pyramid: {collection: acme-pyramid, sources: preuves}
    readers: [group:acme]
    writers: [user:lfoletti]
  team/review:
    addr: $ADDR
    keys_file: $S/data/keys.json
    collection: team-review
    writers: [user:lfoletti]
    curators: [user:kmlead]
  org/secrets:
    addr: $ADDR
    keys_file: $S/data/keys.json
    collection: secrets
    readers: [user:lfoletti]
    writers: [user:kmlead]
    weight: 0.8
    audit_reads: true
groups:
  acme: [lfoletti, mdupont]
audit:
  retention_days: 365
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
LFR=$("$GW" key issue "$S/data/rights.yaml" lfoletti team/review) || ko "issue lfoletti on team/review" "$LFR"
KMR=$("$GW" key issue "$S/data/rights.yaml" kmlead team/review) || ko "issue kmlead on team/review" "$KMR"
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
rec=$(mcp "$LF" "" th_recall '{"query":"ledger rollback","top_k":3,"deadline_ms":60000}')
if printf '%s' "$rec" | grep -q 'spec-1' && ! printf '%s' "$rec" | grep -q 'no answer within'; then ok "th_recall with deadline_ms 60000 serves every brain"; else ko "th_recall deadline" "$rec"; fi
rec=$(mcp "$LF" "" th_recall '{"query":"ledger rollback","deadline_ms":60001}')
if printf '%s' "$rec" | grep -q 'deadline_ms must be between 0 (none) and 60000'; then ok "th_recall refuses deadline_ms 60001"; else ko "th_recall deadline bound" "$rec"; fi
check "operator key on /mcp (401)" denied "$(mcp op-secret "" th_brains '{}')"

# The write circuit, on team/review: a writer proposes, the review gate hides
# the document, the curator lists what waits and accepts, the document is
# served. A brain with curators is served through /mcp only: its per-brain
# endpoint would hand the engine's raw tools to a writer's key.
check "lfoletti: /mcp/team/review refused (403, review circuit)" denied "$(mcp "$LFR" team/review knowledge_search '{"query":"x","collection":"team-review"}')"
check "kmlead: /mcp/team/review refused too (403)" denied "$(mcp "$KMR" team/review knowledge_search '{"query":"x","collection":"team-review"}')"
add=$(mcp "$LF" "" th_add '{"brain":"team/review","document_id":"spec-3","title":"Spec 3","content":"The cutover rehearsal is on Thursday evening with the whole team.","source_url":"smb://share/spec3.docx"}')
if printf '%s' "$add" | grep -q 'spec-3' && ! printf '%s' "$add" | grep -q isError; then ok "th_add proposes spec-3 into team/review"; else ko "th_add" "$add"; fi
hid=$(mcp "$LF" "" th_recall '{"query":"cutover rehearsal","top_k":5,"brains":["team/review"]}')
if printf '%s' "$hid" | grep -q 'spec-3'; then ko "proposed spec-3 served before review" "$hid"; else ok "proposed spec-3 is hidden until review"; fi
check "mdupont (reader): th_add refused" denied "$(mcp "$MD" "" th_add '{"brain":"team/acme","content":"x"}')"
pend=$(mcp "$KMR" "" th_pending '{"brain":"team/review"}')
printf '%s' "$pend" | grep -q 'spec-3' && ok "th_pending lists spec-3 for the curator" || ko "th_pending" "$pend"
check "lfoletti (writer): th_pending refused" denied "$(mcp "$LF" "" th_pending '{"brain":"team/review"}')"
rev=$(mcp "$KMR" "" th_review '{"brain":"team/review","knowledge_id":"spec-3","decision":"accepted"}')
if printf '%s' "$rev" | grep -q isError; then ko "th_review" "$rev"; else ok "th_review accepts spec-3"; fi
srv=$(mcp "$LF" "" th_recall '{"query":"cutover rehearsal","top_k":5,"brains":["team/review"]}')
printf '%s' "$srv" | grep -q 'spec-3' && ok "accepted spec-3 is served" || ko "accepted spec-3 missing" "$srv"
got=$(mcp "$KMR" "" th_get '{"brain":"team/review","knowledge_id":"spec-3"}')
printf '%s' "$got" | grep -q '_reviewed_by.*th/kmlead/team/review' && ok "the engine recorded the curator's key as reviewer" || ko "reviewer" "$got"
# A writer does not take over a document of another author: the curator's
# own note is accepted at once, and lfoletti may not propose under its id.
add=$(mcp "$KMR" "" th_add '{"brain":"team/review","document_id":"note-km","content":"The note of the curator on the cutover."}')
if printf '%s' "$add" | grep -q isError; then ko "curator th_add" "$add"; else ok "the curator adds note-km"; fi
add=$(mcp "$LF" "" th_add '{"brain":"team/review","document_id":"note-km","content":"Taken over."}')
printf '%s' "$add" | grep -q 'belongs to kmlead' && ok "lfoletti may not replace kmlead's note-km" || ko "take over" "$add"

# The client: the plugin's hook asks th_recall with the user's key and
# hands the extracts to Claude Code; a prompt too short asks nothing.
ctx=$(printf '%s' '{"prompt":"when does the ledger move to the new cluster?","cwd":"/tmp"}' | TH_GATEWAY="http://$GWADDR" TH_KEY="$LF" TH_HOOK_BIN="$S/th-hook" "$ROOT/plugin/bin/th-recall")
if printf '%s' "$ctx" | grep -q '"hookEventName":"UserPromptSubmit"' && printf '%s' "$ctx" | grep -q '\[team/acme\] Spec migration, spec-1'; then ok "the hook injects spec-1 with its provenance"; else ko "hook" "$ctx"; fi
ctx=$(printf '%s' '{"prompt":"ok"}' | TH_GATEWAY="http://$GWADDR" TH_KEY="$LF" TH_HOOK_BIN="$S/th-hook" "$ROOT/plugin/bin/th-recall")
[ -z "$ctx" ] && ok "a short prompt injects nothing" || ko "short prompt" "$ctx"
ctx=$(printf '%s' '{"prompt":"when does the ledger move to the new cluster?"}' | TH_GATEWAY="http://$GWADDR" TH_KEY="nope" TH_HOOK_BIN="$S/th-hook" "$ROOT/plugin/bin/th-recall" 2>/dev/null; echo "exit=$?")
[ "$ctx" = "exit=0" ] && ok "a refused key blocks nothing (exit 0, no output)" || ko "refused key" "$ctx"

# A local brain: a memory saved through the CortexDB stdio server into a file
# of this machine. The hook runs the native recall on it next to th_recall,
# and the person's filter chooses the brains. The first native recall starts
# cold and may take longer than the hook's 1.5 s default on a CI runner, so
# the bench gives the recall 10 s: the deadline has its own tests in th-hook.
{ printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"bench","version":"0"}}}' '{"jsonrpc":"2.0","method":"notifications/initialized"}' '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memory_save","arguments":{"memory_id":"fact:bench:otter","content":"The ledger moves to the new cluster under the local code word OTTER."}}}'; sleep 2; } | env -i HOME="$S" PATH="$PATH" CORTEXDB_PATH="$S/local.db" "$S/cortexdb-mcp" >/dev/null 2>&1
hookl() { printf '%s' '{"prompt":"when does the ledger move to the new cluster?","cwd":"/tmp"}' | env HOME="$S" TH_GATEWAY="http://$GWADDR" TH_KEY="$LF" TH_HOOK_BIN="$S/th-hook" TH_RECALL_FILE="$S/recall" TH_RECALL_TIMEOUT=10s TH_LOCAL_RECALL="$S/cortexdb-mcp --recall" TH_LOCAL_BRAINS="projet=$S/local.db" CORTEXDB_REMOTE="127.0.0.1:1" "$ROOT/plugin/bin/th-recall" "$@"; }
ctx=$(hookl)
if printf '%s' "$ctx" | grep -q '\[team/acme\] Spec migration, spec-1' && printf '%s' "$ctx" | grep -q '\[local:projet\]\\n.*OTTER'; then ok "the hook adds the local brain after the shared extracts"; else ko "local brain" "$ctx"; fi
hookl off team/acme && ctx=$(hookl)
if printf '%s' "$ctx" | grep -q 'spec-1' || ! printf '%s' "$ctx" | grep -q 'OTTER'; then ko "skip team/acme" "$ctx"; else ok "a skipped shared brain is left out of the recall"; fi
hookl only local:projet && ctx=$(hookl)
if printf '%s' "$ctx" | grep -q 'OTTER' && ! printf '%s' "$ctx" | grep -q 'shared brains'; then ok "only the local brain: the gateway is not asked"; else ko "only local" "$ctx"; fi
got=$(hookl brains 2>/dev/null)
if printf '%s' "$got" | grep -q '^off  team/acme$' && printf '%s' "$got" | grep -q '^on   local:projet$'; then ok "th-recall brains lists each brain on or off"; else ko "brains" "$got"; fi
hookl only --clear && hookl on team/acme && [ ! -s "$S/recall" ] && ok "the filter comes back to every brain" || ko "clear" "$(cat "$S/recall")"

# Revocation: the file changes at once; the server reads it at start, so the
# check here is on the file, not on the server.
"$GW" key revoke "$S/data/rights.yaml" mdupont team/acme >/dev/null && ok "revoke mdupont" || ko "revoke mdupont" ""
if grep -q 'th/mdupont/team/acme' "$S/data/keys.json"; then ko "mdupont key still in the file" ""; else ok "mdupont key removed from the file"; fi
check "mdupont: revoked key (401 at the gateway)" denied "$(mcp "$MD" team/acme knowledge_search '{"query":"ledger","collection":"team-acme"}')"

# th_get reads a document whole, under the caller's right on the brain.
get=$(mcp "$LF" "" th_get '{"brain":"team/acme","knowledge_id":"spec-1"}')
if printf '%s' "$get" | grep -q '"id\\":\\"spec-1' && ! printf '%s' "$get" | grep -q isError; then ok "th_get reads spec-1 whole"; else ko "th_get" "$get"; fi
check "kmlead: th_get on an unknown brain (refused)" denied "$(mcp "$KMR" "" th_get '{"brain":"org/none","knowledge_id":"secret-doc"}')"

# The pyramid of team/acme: a synthesis in its own collection, citing spec-1.
# The key of team/acme holds both collections; th_recall serves the synthesis
# first, and th_descend opens what it cites.
check "lfoletti: write a synthesis into the pyramid of team/acme" ok "$(mcp "$LF" team/acme ingest_document '{"document_id":"pyramide/ledger","title":"Ledger migration","content":"The ledger migration happens on Sunday night, with a rollback window of two hours.","collection":"acme-pyramid","metadata":{"preuves":"[\"spec-1\"]"}}')"
rec=$(mcp "$LF" "" th_recall '{"query":"ledger rollback","top_k":3,"brains":["team/acme"]}')
first=$(printf '%s' "$rec" | sed -n 's/.*"results\\":\[{\\"brain\\":\\"team\/acme\\",\\"knowledge_id\\":\\"\([^\\]*\)\\".*/\1/p')
if [ "$first" = "pyramide/ledger" ] && printf '%s' "$rec" | grep -q 'synthesis: Ledger migration, pyramide/ledger' && printf '%s' "$rec" | grep -q 'spec-1'; then ok "th_recall serves the synthesis first, then the passages"; else ko "th_recall with a pyramid (first: $first)" "$rec"; fi
desc=$(mcp "$LF" "" th_descend '{"brain":"team/acme","knowledge_id":"pyramide/ledger"}')
if printf '%s' "$desc" | grep -q 'knowledge_id\\":\\"spec-1' && printf '%s' "$desc" | grep -q 'rollback window is two hours' && ! printf '%s' "$desc" | grep -q isError; then ok "th_descend opens spec-1, the source of the synthesis"; else ko "th_descend" "$desc"; fi
check "lfoletti: th_descend on org/secrets, which has no pyramid (refused)" denied "$(mcp "$LF" "" th_descend '{"brain":"org/secrets","knowledge_id":"secret-doc"}')"

# Entities across brains. The operator of the server declares Acme in two
# collections (a confined key may not declare entities). th_entity asks each
# brain with the caller's key. Through the key of team/acme alone, the
# neighbor that only secrets states is never served.
# opmcp <tool> <args-json>: the CortexDB stdio client with the operator key.
opmcp() {
  { printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"e2e","version":"0"}}}' '{"jsonrpc":"2.0","method":"notifications/initialized"}'
    printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"%s","arguments":%s}}\n' "$1" "$2"; sleep 1.5
  } | CORTEXDB_REMOTE=$ADDR CORTEXDB_GRPC_TOKEN=op-secret "$S/cortexdb-mcp" 2>&1 | grep '"id":2'
}
check "operator: Acme and Ledger in team-acme" ok "$(opmcp knowledge_save '{"knowledge_id":"ent-acme","content":"Acme runs the ledger.","collection":"team-acme","entities":[{"name":"Acme","type":"Company"},{"name":"Ledger","type":"System"}],"relations":[{"from":"Acme","to":"Ledger","type":"runs"}]}')"
check "operator: Acme and Retainer in secrets" ok "$(opmcp knowledge_save '{"knowledge_id":"ent-secret","content":"Acme pays the retainer.","collection":"secrets","entities":[{"name":"Acme","type":"Company"},{"name":"Retainer","type":"Contract"}],"relations":[{"from":"Acme","to":"Retainer","type":"pays"}]}')"
ent=$(mcp "$LF" "" th_entity '{"entity":"Acme"}')
if printf '%s' "$ent" | grep -q 'known_in\\":\[\\"org/secrets\\",\\"team/acme\\"\]' && printf '%s' "$ent" | grep -q 'entity:ledger' && printf '%s' "$ent" | grep -q 'entity:retainer' && ! printf '%s' "$ent" | grep -q 'vector\|isError'; then ok "th_entity: lfoletti finds Acme in both brains, with Ledger and Retainer"; else ko "th_entity lfoletti" "$ent"; fi
ent=$(mcp "$LF" "" th_entity '{"entity":"Acme","brains":["team/acme"]}')
if printf '%s' "$ent" | grep -q 'known_in\\":\[\\"team/acme\\"\]' && printf '%s' "$ent" | grep -q 'entity:ledger' && ! printf '%s' "$ent" | grep -q 'retainer\|isError'; then ok "th_entity: the key of team/acme serves Acme and Ledger, without Retainer"; else ko "th_entity team/acme" "$ent"; fi
# An entity mentioned by more chunks than the limit keeps its neighbors: the
# walk does not follow the mentions, which would fill the limit with passages.
# The mentions are written as a graph builder writes them: a document, then
# its entities tied to its chunk (upsert_entities, chunk_ids). The last note
# names Zeta with Acme, with no relation between them: from depth 2, the
# walk that follows the mentions still serves Zeta.
for n in 1 2 3; do
  zeta=''; [ $n = 3 ] && zeta=',{"name":"Zeta","type":"Company","chunk_ids":["chunk:note-3:000"]}'
  check "operator: note $n in team-acme" ok "$(opmcp ingest_document '{"document_id":"note-'$n'","content":"Acme note '$n'.","collection":"team-acme"}')"
  check "operator: note $n mentions Acme" ok "$(opmcp upsert_entities '{"document_id":"note-'$n'","entities":[{"name":"Acme","type":"Company","chunk_ids":["chunk:note-'$n':000"]}'"$zeta"']}')"
done
ent=$(mcp "$LF" "" th_entity '{"entity":"Acme","brains":["team/acme"],"limit":2}')
if printf '%s' "$ent" | grep -q 'entity:ledger' && ! printf '%s' "$ent" | grep -q isError; then ok "th_entity: Acme, mentioned by three chunks, keeps Ledger at limit 2"; else ko "th_entity past the mentions" "$ent"; fi
ent=$(mcp "$LF" "" th_entity '{"entity":"Acme","brains":["team/acme"],"max_depth":2}')
if printf '%s' "$ent" | grep -q 'entity:ledger' && printf '%s' "$ent" | grep -q 'entity:zeta' && ! printf '%s' "$ent" | grep -q isError; then ok "th_entity: at depth 2, Acme also has Zeta, named in the same passage"; else ko "th_entity co-mention" "$ent"; fi

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

# Sign-in through the identity provider: 404 without an identity block;
# with one (added hot, the rights file reloads), a bad token is 401, and
# th-gateway login, driven by curl as the browser, brings lfoletti's key
# into the env file: the key the gateway issued before.
code=$(rest GET /v1/identity "" "")
[ "$code" = 404 ] && ok "GET /v1/identity without an identity block is 404" || ko "identity 404" "$code $(cat "$S/rest.out")"
# Port 0: the system gives a free port, and the provider prints it once it
# listens. A port chosen here can be in use: 50000 to 59999 is in the
# ephemeral range of Linux, and the login failed in the CI on 52172.
node "$ROOT/tests/fake_idp.mjs" --port 0 --client-id thalamus-cli --audience thalamus > "$S/data/idp.log" 2>&1 &
IDP=$!
IDPPORT=
for i in $(seq 1 100); do IDPPORT=$(sed -n 's|^fake idp at http://127.0.0.1:||p' "$S/data/idp.log"); [ -n "$IDPPORT" ] && break; sleep 0.2; done
[ -n "$IDPPORT" ] && ok "the fake provider listens on a free port" || ko "fake provider" "$(cat "$S/data/idp.log")"
printf 'identity:\n  issuer: http://127.0.0.1:%s\n  audience: thalamus\n  client_id: thalamus-cli\n  strip_domain: true\n' "$IDPPORT" >> "$S/data/rights.yaml"
code=$(rest POST /v1/login "not-a-token" "")
[ "$code" = 401 ] && grep -q 'three parts' "$S/rest.out" && ok "POST /v1/login with a bad token is 401" || ko "login 401" "$code $(cat "$S/rest.out")"
code=$(rest GET /v1/identity "" "")
[ "$code" = 200 ] && grep -q '"client_id":"thalamus-cli"' "$S/rest.out" && ok "GET /v1/identity names the provider and the client" || ko "identity" "$code $(cat "$S/rest.out")"
"$S/th-gateway" login "http://$GWADDR" --env "$S/data/env" --no-browser > "$S/data/login.out" 2>&1 &
LOGIN=$!
for i in 1 2 3 4 5 6 7 8 9 10; do grep -q '^http' "$S/data/login.out" && break; sleep 0.3; done
curl -sL "$(grep '^http' "$S/data/login.out" | head -1)" > "$S/data/browser.out"
wait $LOGIN; lcode=$?
if [ "$lcode" = 0 ] && grep -q 'Signed in as lfoletti' "$S/data/login.out" && grep -q 'Signed in. You can close this window' "$S/data/browser.out"; then ok "th-gateway login signs lfoletti in through the provider"; else ko "login" "$lcode $(cat "$S/data/login.out") $(cat "$S/data/browser.out")"; fi
[ "$(cat "$S/data/env")" = "TH_GATEWAY=http://$GWADDR
TH_KEY=$LFS" ] && ok "the env file holds the gateway and the first brain's key, issued before" || ko "env file" "$(cat "$S/data/env")"
grep -q '"user":"lfoletti".*"tool":"login".*"brains":\["org/secrets","team/acme","team/review"\].*"ok":true' "$S/data/audit.jsonl" && ok "the login is audited with the brains" || ko "login audit" "$(grep login "$S/data/audit.jsonl")"
kill $IDP 2>/dev/null; wait $IDP 2>/dev/null
# sed -i takes its suffix differently in BSD and GNU sed: edit through a copy.
edit() { sed "$1" "$2" > "$2.new" && mv "$2.new" "$2"; }
edit '/^identity:/,$d' "$S/data/rights.yaml"
sleep 0.5

# The rights file changes while the gateway runs: a withdrawn right is
# refused at the next call, no restart; a restored right is served again.
# lfoletti reads org/secrets by name only, so one line withdraws it all.
edit 's/    readers: \[user:lfoletti\]/    readers: []/' "$S/data/rights.yaml"
check "lfoletti: right on org/secrets withdrawn in rights.yaml (403, no restart)" denied "$(mcp "$LFS" org/secrets knowledge_search '{"query":"secret","collection":"secrets"}')"
edit 's/    readers: \[\]/    readers: [user:lfoletti]/' "$S/data/rights.yaml"
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
# The gateway stops on SIGTERM, as a service manager stops it: it writes the counted reads of the day. Reads of
# org/secrets (audit_reads) are named; reads of team/acme are counted.
kill -TERM $GATEWAY; wait $GATEWAY 2>/dev/null
A="$S/data/audit.jsonl"
grep -q '"tool":"th_brains"' "$A" && ko "th_brains audited" "$(grep th_brains "$A")" || ok "th_brains, which reads no brain, leaves no line"
grep -q '"tool":"th_recall","ids":{"brains":\["org/secrets"\],"knowledge_ids":\[.*org/secrets:secret-doc' "$A" && ok "th_recall names the reads of org/secrets only" || ko "th_recall audit line" "$(grep th_recall "$A")"
grep -q 'team/acme:spec' "$A" && ko "a read of team/acme named" "$(grep 'team/acme:spec' "$A")" || ok "the reads of team/acme are not named"
grep -q '"brain":"team/acme","tool":"knowledge_search"' "$A" && ko "a search of team/acme named" "$(grep '"brain":"team/acme","tool":"knowledge_search"' "$A")" || ok "the searches of team/acme are not named"
grep -q '"user":"","key":"","brain":"team/acme","tool":"reads","day":"[0-9-]*","calls":[0-9]' "$A" && ok "the reads of team/acme are counted for the day, without a name" || ko "team/acme count line" "$(grep '"reads"' "$A")"
grep -q '"user":"lfoletti","key":"th/lfoletti/org/secrets","brain":"org/secrets","tool":"knowledge_search".*"ok":true' "$A" && ok "a search of org/secrets is named" || ko "org/secrets search line" "$(grep org/secrets "$A")"
grep -q '"user":"lfoletti","key":"th/lfoletti/team/acme","brain":"team/acme","tool":"ingest_document","ids":{"document_id":"spec-1"},"ok":true' "$A" && ok "ingest audited with its id" || ko "ingest audit line" "$(grep ingest_document "$A" | head -1)"
grep -q '"user":"lfoletti".*"tool":"graph_statistics","ok":false,"error":.*PermissionDenied' "$A" && ok "a read the brain refuses is named, as an error" || ko "graph_statistics audit line" "$(grep graph_statistics "$A")"
grep -q '"user":"kmlead".*"brain":"org/secrets".*"knowledge_id":"secret-doc"' "$A" && ok "kmlead's save audited on org/secrets" || ko "kmlead audit line" "$(grep kmlead "$A")"
grep -q '"user":"kmlead".*"brain":"team/review","tool":"th_review","ids":{"decision":"accepted","knowledge_id":"spec-3"},"ok":true' "$A" && ok "the review is audited with its decision" || ko "th_review audit line" "$(grep th_review "$A")"
sum=$("$GW" audit "$A")
if printf '%s' "$sum" | grep -q '^  [0-9-]*  team/acme  [0-9]* calls, [0-9]* failed, counted$' && printf '%s' "$sum" | grep -q '^  kmlead  th_review  team/review  1$' && printf '%s' "$sum" | grep -q 'graph_statistics  .*PermissionDenied'; then ok "th-gateway audit sums up the audit file"; else ko "audit summary" "$sum"; fi
echo
echo "pass=$pass fail=$fail"
[ "$fail" -eq 0 ]
