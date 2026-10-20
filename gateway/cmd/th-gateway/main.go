// Command th-gateway is the Thalamus gateway: it reads rights.yaml,
// issues and revokes the confined CortexDB keys of employees, and serves the
// MCP relay that audits every tool call.
//
//	th-gateway rights check <rights.yaml>
//	th-gateway rights of <rights.yaml> <user>
//	th-gateway key issue <rights.yaml> <user> <brain>
//	th-gateway key revoke <rights.yaml> <user> <brain>
//	th-gateway serve <rights.yaml> [--listen 127.0.0.1:8787] [--audit audit.jsonl]
//	th-gateway login <gateway> [--env ~/.thalamus/env] [--no-browser]
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
