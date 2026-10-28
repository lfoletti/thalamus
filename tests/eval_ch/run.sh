#!/bin/sh
# Swiss cross-language recall bench, one configuration per run:
#   sh tests/eval_ch/run.sh lexical
#   sh tests/eval_ch/run.sh <fastembed model> <dimension>
# Builds cortexdb-grpc and th-gateway, starts a fresh brain (with the local encoder when a model
# is named), and runs eval.py. Results in tests/eval_ch/results/<label>.json.
ROOT=$(cd "$(dirname "$0")/../.." && pwd); H=$ROOT/tests/eval_ch
. "$ROOT/tests/cortexdb_clone.sh"
S=$(mktemp -d); trap 'kill $GW $GRPC $ENC 2>/dev/null; rm -rf "$S"' EXIT
(cd "$CDB" && go build -o "$S/cortexdb-grpc" ./cmd/cortexdb-grpc) || exit 1
(cd "$ROOT/gateway" && go build -o "$S/th-gateway" ./cmd/th-gateway) || exit 1
P=47930; LABEL=lexical; ENC=
if [ "$1" != lexical ]; then
  LABEL=$(basename "$1"); python3 -I "$H/encoder.py" "$1" $((P+2)) > "$S/enc.log" 2>&1 & ENC=$!
  until grep -q encoder "$S/enc.log" 2>/dev/null; do sleep 1; kill -0 $ENC || { cat "$S/enc.log"; exit 1; }; done
  set -- OPENAI_BASE_URL=http://127.0.0.1:$((P+2))/v1 OPENAI_API_KEY=local CORTEXDB_EMBED_MODEL=local CORTEXDB_EMBED_DIM=$2
else set --; fi
echo '{"keys":[{"id":"operator","secret":"op-secret","clearance":"read-write"}]}' > "$S/keys.json"
cat > "$S/rights.yaml" <<EOR
brains:
  ch:
    addr: 127.0.0.1:$P
    keys_file: $S/keys.json
    collection: ch
    readers: [user:lauro]
    writers: [user:lauro]
EOR
KEY=$("$S/th-gateway" key issue "$S/rights.yaml" lauro ch | tail -1)
env "$@" "$S/cortexdb-grpc" -db "$S/brain.db" -addr 127.0.0.1:$P -keys "$S/keys.json" > "$S/grpc.log" 2>&1 & GRPC=$!
"$S/th-gateway" serve "$S/rights.yaml" --listen 127.0.0.1:$((P+1)) --audit "$S/audit.jsonl" > "$S/gw.log" 2>&1 & GW=$!
sleep 2
python3 -I "$H/eval.py" http://127.0.0.1:$((P+1)) "$KEY" "$LABEL" || { tail -5 "$S/grpc.log" "$S/gw.log"; exit 1; }
