# Sourced by the benches after ROOT is set: CDB is the CortexDB clone they build from.
# CORTEXDB_CLONE names it; without it, cortexdb/ at the root of this repository.
CDB=${CORTEXDB_CLONE:-$ROOT/cortexdb}
[ -d "$CDB/cmd/cortexdb-grpc" ] || {
  echo "no CortexDB clone at $CDB: clone github.com/liliang-cn/cortexdb there, or set CORTEXDB_CLONE to your clone" >&2
  exit 1
}
