Batch 1 is up as #67, #68, #69 and #70, one fix each. Point 2 in detail, before any code: **scoped keys and the toolbox.**

**Today.** A key with a `scope` in the key file is refused every `CallTool` (gRPC) and every `POST /v1/tools/<name>` (REST), because the tool arguments are opaque JSON the policy cannot see into. The typed RPCs honour the scope row by row. The MCP stdio client in remote mode only speaks `CallTool`, so a confined key cannot be given to a developer's machine at all: it is the unconfined key or nothing.

**Proposal.** Let a confined key call a tool when all of this holds, and refuse otherwise:

1. **The tool declares the confined field** in its input schema (`collection`, `user_id`, `scope`, `namespace`, typed as a string). A tool that takes no such field (`graph_statistics`, `delete_document_graph`) stays refused: it has no way to honour the confinement. The list is derived from `ToolDefinitions()`, not kept by hand.
2. **The argument is set and equals the key's value**, and so does every nested copy (a plan's `filters.collection`) and every other spelling of the key (`COLLECTION`, which `encoding/json` reads as `collection`). This is the same rule the REST body check applies to typed requests, and the one #64 tightens.
3. **A row named by id already lies within the scope.** `knowledge_update`, `knowledge_save`, `ingest_document`, `memory_save`, `memory_get` and the like load the row the id names and refuse one outside the scope with `NOT_FOUND` / `404`, worded like a missing row, so the id space is not an oracle. A row that does not exist yet passes: the write creates it, inside the scope rule 2 forced. This is what `rpcserver/ownership.go` already does for the typed RPCs.
4. **Several values per field.** A key may list `collections: ["team-a", "shared"]` (also `user_ids`, `scopes`, `namespaces`); the singular form keeps working. An empty list is a configuration error.

**What does not change.** An unconfined key, the typed RPCs, the read-only clearance, and every tool for an unconfined caller behave exactly as today. A tool with no scope field stays refused to a confined key.

**Shape.** Two PRs: the lists (4), about 190 lines, which the typed RPCs can use alone; then the toolbox rules (1 to 3), one PR of about 850 lines or three smaller ones (`authz` rules unplugged, gRPC wiring, REST wiring), as you prefer. Each rule has a test that fails before it; the row rule was found by an end-to-end run where a key confined to one collection overwrote another collection's document by passing its own collection with that document's id.

Does the principle suit you? If you would rather keep confined keys off the toolbox for good, I will stop here and run one server per team instead.
