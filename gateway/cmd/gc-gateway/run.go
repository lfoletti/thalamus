package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/lfoletti/granular-cortex/gateway/keys"
	"github.com/lfoletti/granular-cortex/gateway/relay"
	"github.com/lfoletti/granular-cortex/gateway/rights"
)

const usage = `usage:
  gc-gateway rights check <rights.yaml>
  gc-gateway rights of <rights.yaml> <user>
  gc-gateway key issue <rights.yaml> <user> <brain>
  gc-gateway key revoke <rights.yaml> <user> <brain>
  gc-gateway serve <rights.yaml> [--listen 127.0.0.1:8787] [--audit audit.jsonl]`

// run dispatches the subcommand and returns the exit code: 0 done, 1 the
// command failed, 2 the command line is wrong.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	var err error
	switch {
	case len(args) >= 1 && args[0] == "serve":
		err = serve(ctx, args[1:], stdout)
	case len(args) >= 2 && args[0] == "rights" && args[1] == "check":
		err = withRights(args[2:], 1, func(r *rights.Rights, a []string) error { return rightsCheck(r, stdout) })
	case len(args) >= 2 && args[0] == "rights" && args[1] == "of":
		err = withRights(args[2:], 2, func(r *rights.Rights, a []string) error { return rightsOf(r, a[1], stdout) })
	case len(args) >= 2 && args[0] == "key" && args[1] == "issue":
		err = withRights(args[2:], 3, func(r *rights.Rights, a []string) error { return keyIssue(r, a[1], a[2], stdout) })
	case len(args) >= 2 && args[0] == "key" && args[1] == "revoke":
		err = withRights(args[2:], 3, func(r *rights.Rights, a []string) error { return keyRevoke(r, a[1], a[2], stdout) })
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "gc-gateway: %v\n", err)
		if errors.Is(err, errUsage) {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		return 1
	}
	return 0
}

var errUsage = errors.New("wrong arguments")

// withRights loads the rights file named by the first argument, checks the
// argument count, and runs the command.
func withRights(args []string, n int, cmd func(*rights.Rights, []string) error) error {
	if len(args) != n {
		return fmt.Errorf("%w: expected %d arguments, got %d", errUsage, n, len(args))
	}
	r, err := rights.Load(args[0])
	if err != nil {
		return err
	}
	return cmd(r, args)
}

func rightsCheck(r *rights.Rights, out io.Writer) error {
	fmt.Fprintf(out, "%d brain(s), %d group(s)\n", len(r.Brains), len(r.Groups))
	for _, name := range sortedBrains(r) {
		b := r.Brains[name]
		fmt.Fprintf(out, "  %s  %s  collection=%s  readers=%d  writers=%d\n", name, b.Addr, b.Collection, len(b.Readers), len(b.Writers))
	}
	return nil
}

func rightsOf(r *rights.Rights, user string, out io.Writer) error {
	grants := r.BrainsOf(user)
	if len(grants) == 0 {
		fmt.Fprintf(out, "%s has access to no brain\n", user)
		return nil
	}
	for _, g := range grants {
		fmt.Fprintf(out, "%s  %s\n", g.Brain, g.Access)
	}
	return nil
}

// keyIssue writes the key into the brain's key file and prints the secret,
// alone on stdout, so a script can capture it.
func keyIssue(r *rights.Rights, user, brain string, out io.Writer) error {
	access, err := r.Access(user, brain)
	if err != nil {
		return err
	}
	b := r.Brains[brain]
	k, err := keys.Issue(b.KeysFile, user, brain, b.Collection, access)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, k.Secret)
	return nil
}

func keyRevoke(r *rights.Rights, user, brain string, out io.Writer) error {
	b, ok := r.Brains[brain]
	if !ok {
		return fmt.Errorf("unknown brain %q", brain)
	}
	if err := keys.Revoke(b.KeysFile, user, brain); err != nil {
		return err
	}
	fmt.Fprintf(out, "revoked %s\n", keys.ID(user, brain))
	return nil
}

// serve runs the relay until the context ends. It prints the address it
// listens on, so a caller that asked for port 0 learns the port.
func serve(ctx context.Context, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	listen := fs.String("listen", "127.0.0.1:8787", "address to listen on")
	auditPath := fs.String("audit", "audit.jsonl", "audit file, one JSON line per tool call")
	if len(args) < 1 {
		return fmt.Errorf("%w: serve needs the rights file", errUsage)
	}
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	r, err := rights.Load(args[0])
	if err != nil {
		return err
	}
	audit, err := relay.OpenAudit(*auditPath)
	if err != nil {
		return err
	}
	defer func() { _ = audit.Close() }()
	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	srv := &http.Server{Handler: relay.New(r, audit).Handler(), ReadHeaderTimeout: 10 * time.Second}
	fmt.Fprintf(out, "gc-gateway listening on http://%s (/mcp federated, /mcp/<brain> per brain), audit in %s\n", lis.Addr(), *auditPath)
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	err = srv.Serve(lis)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

func sortedBrains(r *rights.Rights) []string {
	names := make([]string, 0, len(r.Brains))
	for n := range r.Brains {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
