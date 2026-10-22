package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/relay"
	"github.com/lfoletti/thalamus/gateway/rights"
)

const usage = `usage:
  th-gateway rights check <rights.yaml>
  th-gateway rights of <rights.yaml> <user>
  th-gateway key issue <rights.yaml> <user> <brain>
  th-gateway key revoke <rights.yaml> <user> <brain>
  th-gateway key list <rights.yaml>
  th-gateway key sync <rights.yaml> [--dry-run]
  th-gateway serve <rights.yaml> [--listen 127.0.0.1:8787] [--audit audit.jsonl]
  th-gateway login <gateway> [--env ~/.thalamus/env] [--no-browser]
  th-gateway audit <audit.jsonl> [--since YYYY-MM-DD] [--user login] [--brain brain]`

// run dispatches the subcommand and returns the exit code: 0 done, 1 the
// command failed, 2 the command line is wrong.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	var err error
	switch {
	case len(args) >= 1 && args[0] == "serve":
		err = serve(ctx, args[1:], stdout, stderr)
	case len(args) >= 1 && args[0] == "audit":
		err = auditCommand(args[1:], stdout)
	case len(args) >= 1 && args[0] == "login":
		err = loginCommand(ctx, args[1:], stdout)
	case len(args) >= 2 && args[0] == "rights" && args[1] == "check":
		err = withRights(args[2:], 1, func(r *rights.Rights, a []string) error { return rightsCheck(r, stdout) })
	case len(args) >= 2 && args[0] == "rights" && args[1] == "of":
		err = withRights(args[2:], 2, func(r *rights.Rights, a []string) error { return rightsOf(r, a[1], stdout) })
	case len(args) >= 2 && args[0] == "key" && args[1] == "issue":
		err = withRights(args[2:], 3, func(r *rights.Rights, a []string) error { return keyIssue(r, a[1], a[2], stdout) })
	case len(args) >= 2 && args[0] == "key" && args[1] == "revoke":
		err = withRights(args[2:], 3, func(r *rights.Rights, a []string) error { return keyRevoke(r, a[1], a[2], stdout) })
	case len(args) >= 2 && args[0] == "key" && args[1] == "list":
		err = withRights(args[2:], 1, func(r *rights.Rights, a []string) error { return keyList(r, stdout) })
	case len(args) >= 2 && args[0] == "key" && args[1] == "sync":
		err = keySync(args[2:], stdout)
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "th-gateway: %v\n", err)
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
	k, err := keys.IssueScoped(b.KeysFile, user, brain, b.Collections(), access)
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

// stale says why a held key no longer matches the right, or "" when it
// still does.
func stale(r *rights.Rights, b rights.Brain, h keys.Held) string {
	access, _ := r.Access(h.User, h.Brain)
	if !access.AllowsRead() {
		return "no right"
	}
	if access.Clearance() != h.Clearance {
		return "right is " + access.String()
	}
	if !slices.Equal(h.Collections, b.Collections()) {
		switch want := b.Collections(); len(want) {
		case 0:
			return "brain is the whole server"
		case 1:
			return "collection is " + want[0]
		default:
			return "collections are " + strings.Join(want, ", ")
		}
	}
	return ""
}

// keyList prints, per brain, the keys the gateway issued and whether each
// still matches the right.
func keyList(r *rights.Rights, out io.Writer) error {
	others := map[string]int{}
	for _, name := range sortedBrains(r) {
		b := r.Brains[name]
		held, n, err := keys.List(b.KeysFile)
		if err != nil {
			return err
		}
		others[b.KeysFile] = n
		for _, h := range held {
			if h.Brain != name {
				continue
			}
			status := "ok"
			if why := stale(r, b, h); why != "" {
				status = "stale: " + why
			}
			fmt.Fprintf(out, "%s  %s  %s  %s\n", name, h.User, h.Clearance, status)
		}
	}
	files := make([]string, 0, len(others))
	for f := range others {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		if others[f] > 0 {
			fmt.Fprintf(out, "%d other key(s) in %s\n", others[f], f)
		}
	}
	return nil
}

// keySync brings the key files to the rights: it issues the missing keys,
// reissues the stale ones (a new secret), and revokes the keys of users
// without a right. Each action is printed, with the secret when one was
// made. --dry-run prints the plan and changes nothing.
func keySync(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dry := fs.Bool("dry-run", false, "print the plan, change nothing")
	if len(args) < 1 {
		return fmt.Errorf("%w: sync needs the rights file", errUsage)
	}
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	r, err := rights.Load(args[0])
	if err != nil {
		return err
	}
	// The plan first, then the actions: one place to stop on a failure.
	type action struct {
		verb, user, brain string
		b                 rights.Brain
		access            rights.Access
	}
	var plan []action
	for _, name := range sortedBrains(r) {
		b := r.Brains[name]
		held, _, err := keys.List(b.KeysFile)
		if err != nil {
			return err
		}
		have := map[string]keys.Held{}
		for _, h := range held {
			if h.Brain == name {
				have[h.User] = h
			}
		}
		// Every login the rights name, and every holder of a key on this
		// brain: a holder the rights no longer name must lose the key.
		users := r.Users()
		for user := range have {
			users = append(users, user)
		}
		sort.Strings(users)
		users = slices.Compact(users)
		for _, user := range users {
			access, _ := r.Access(user, name)
			h, has := have[user]
			switch {
			case !has && access.AllowsRead():
				plan = append(plan, action{"issue", user, name, b, access})
			case has && !access.AllowsRead():
				plan = append(plan, action{"revoke", user, name, b, access})
			case has && stale(r, b, h) != "":
				plan = append(plan, action{"reissue", user, name, b, access})
			}
		}
	}
	if len(plan) == 0 {
		fmt.Fprintln(out, "nothing to do")
	}
	for _, a := range plan {
		if err := act(out, *dry, a.verb, a.b, a.user, a.brain, a.access); err != nil {
			return err
		}
	}
	return nil
}

// act does one sync action, or prints what it would do.
func act(out io.Writer, dry bool, verb string, b rights.Brain, user, brain string, access rights.Access) error {
	id := keys.ID(user, brain)
	if dry {
		fmt.Fprintf(out, "would %s %s\n", verb, id)
		return nil
	}
	if verb != "issue" {
		if err := keys.Revoke(b.KeysFile, user, brain); err != nil {
			return err
		}
		if verb == "revoke" {
			fmt.Fprintf(out, "revoked %s\n", id)
			return nil
		}
	}
	k, err := keys.IssueScoped(b.KeysFile, user, brain, b.Collections(), access)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%sd %s %s\n", verb, id, k.Secret)
	return nil
}

// serve runs the relay until the context ends. It prints the address it
// listens on, so a caller that asked for port 0 learns the port.
func serve(ctx context.Context, args []string, out, errOut io.Writer) error {
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
	src, err := rights.Open(args[0])
	if err != nil {
		return err
	}
	src.Warn = func(err error) { fmt.Fprintf(errOut, "th-gateway: %v\n", err) }
	audit, err := relay.OpenAudit(*auditPath)
	if err != nil {
		return err
	}
	defer func() { _ = audit.Close() }()
	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	srv := &http.Server{Handler: relay.New(src, audit).Handler(), ReadHeaderTimeout: 10 * time.Second}
	fmt.Fprintf(out, "th-gateway listening on http://%s (/mcp federated, /mcp/<brain> per brain), audit in %s\n", lis.Addr(), *auditPath)
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
