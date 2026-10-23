package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lfoletti/thalamus/gateway/login"
)

// loginCommand signs the user in through the browser and writes
// TH_GATEWAY and TH_KEY into the env file, for the plugin and the hook.
func loginCommand(ctx context.Context, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	envPath := fs.String("env", "", "file that receives TH_GATEWAY and TH_KEY (default ~/.thalamus/env)")
	noBrowser := fs.Bool("no-browser", false, "print the sign-in address, do not open the browser")
	if len(args) < 1 {
		return fmt.Errorf("%w: login needs the gateway address", errUsage)
	}
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	if *envPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home directory: %w", err)
		}
		*envPath = filepath.Join(home, ".thalamus", "env")
	}
	f := &login.Flow{Gateway: args[0], Out: out, Open: openBrowser}
	if *noBrowser {
		f.Open = nil
	}
	a, err := f.Run(ctx)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*envPath), 0o700); err != nil {
		return fmt.Errorf("env file: %w", err)
	}
	text := "TH_GATEWAY=" + strings.TrimRight(args[0], "/") + "\nTH_KEY=" + a.Key + "\n"
	if err := os.WriteFile(*envPath, []byte(text), 0o600); err != nil {
		return fmt.Errorf("env file: %w", err)
	}
	brains := make([]string, 0, len(a.Brains))
	for _, b := range a.Brains {
		brains = append(brains, b.Brain+" ("+b.Access+")")
	}
	fmt.Fprintf(out, "Brains: %s\nTH_GATEWAY and TH_KEY written to %s\nThe hook reads it. For the plugin's MCP server, load it in your shell: set -a; . %s; set +a\n", strings.Join(brains, ", "), *envPath, *envPath)
	return nil
}

// openBrowser opens the address with TH_BROWSER, or the system opener. A
// variable, so that a test can act as the browser.
var openBrowser = func(u string) error {
	return exec.Command(openerName(runtime.GOOS), u).Start()
}

func openerName(goos string) string {
	if n := os.Getenv("TH_BROWSER"); n != "" {
		return n
	}
	if goos == "darwin" {
		return "open"
	}
	return "xdg-open"
}
