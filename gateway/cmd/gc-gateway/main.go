// Command gc-gateway is the Granular Cortex gateway: it reads rights.yaml,
// issues and revokes the confined CortexDB keys of employees, and serves the
// MCP relay that audits every tool call.
//
//	gc-gateway rights check <rights.yaml>
//	gc-gateway rights of <rights.yaml> <user>
//	gc-gateway key issue <rights.yaml> <user> <brain>
//	gc-gateway key revoke <rights.yaml> <user> <brain>
//	gc-gateway serve <rights.yaml> [--listen 127.0.0.1:8787] [--audit audit.jsonl]
package main

import (
	"context"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
