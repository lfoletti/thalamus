// Command gc-hook is the client side of the gateway for Claude Code: the
// UserPromptSubmit hook that asks the gateway's gc_recall with the
// employee's key and hands the extracts, with their provenance, to the
// agent as additional context.
//
//	gc-hook recall    reads the hook payload on stdin; GC_GATEWAY, GC_KEY,
//	                  GC_RECALL_TOPK (default 3)
//
// It never blocks a prompt: every problem is exit 0 with no output and the
// reason on stderr.
package main

import "os"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Getenv, os.Stdout, os.Stderr))
}
