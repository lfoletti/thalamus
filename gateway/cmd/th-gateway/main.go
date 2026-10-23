// Command th-gateway is the Thalamus gateway: it reads rights.yaml,
// issues and revokes the confined CortexDB keys of users, and serves the
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
	"syscall"
)

func main() {
	// SIGTERM is how a service manager stops the gateway: serve must close
	// the audit, which writes the counted reads of the day.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
