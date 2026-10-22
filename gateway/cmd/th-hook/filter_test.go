package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// A brain is on when only is empty or names it, and skip does not name it.
// The file has one setting per line; comments and unknown keys are ignored.
func TestTheFilterSaysWhichBrainsAreOn(t *testing.T) {
	f := parseFilter("# my choice\nonly = team/acme, local:projet\nskip=team/acme\nother=x\n")
	if !reflect.DeepEqual(f, filter{Only: []string{"team/acme", "local:projet"}, Skip: []string{"team/acme"}}) {
		t.Fatalf("%+v", f)
	}
	for name, want := range map[string]bool{"team/acme": false, "local:projet": true, "org/x": false} {
		if f.on(name) != want {
			t.Fatalf("%s: %v", name, !want)
		}
	}
	if !(filter{}).on("org/x") || !(filter{Skip: []string{"a"}}).on("org/x") {
		t.Fatal("an empty only is every brain")
	}
	if got := parseFilter(f.String()); !reflect.DeepEqual(got, f) {
		t.Fatalf("round trip: %+v", got)
	}
	if (filter{}).String() != "" {
		t.Fatalf("empty: %q", (filter{}).String())
	}
}

// on removes a brain from skip and adds it to a non-empty only; off adds
// it to skip once; only sets the list, and only --clear empties it.
func TestTheFilterCommandsChangeTheChoice(t *testing.T) {
	f := filter{Only: []string{"a"}, Skip: []string{"b"}}
	for _, c := range []struct {
		cmd   string
		names []string
		want  filter
	}{
		{"on", []string{"b"}, filter{Only: []string{"a", "b"}}},
		{"on", []string{"a"}, filter{Only: []string{"a"}, Skip: []string{"b"}}},
		{"off", []string{"b"}, filter{Only: []string{"a"}, Skip: []string{"b"}}},
		{"off", []string{"a"}, filter{Only: []string{"a"}, Skip: []string{"b", "a"}}},
		{"only", []string{"c", "d"}, filter{Only: []string{"c", "d"}, Skip: []string{"b"}}},
		{"only", []string{"--clear"}, filter{Skip: []string{"b"}}},
	} {
		got := f.with(c.cmd, c.names)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("%s %v: %+v", c.cmd, c.names, got)
		}
	}
	if got := (filter{}).with("on", []string{"a"}); len(got.Only) != 0 {
		t.Fatalf("on with no only: %+v", got)
	}
}

// The hook obeys the file: skipped server brains go to th_recall as
// skip_brains, an only list of server brains goes as brains, an only list
// with no server brain asks the gateway nothing, and a skipped local brain
// does not run. TH_RECALL_FILE names another file.
func TestTheHookObeysTheFilter(t *testing.T) {
	g := &fakeGateway{secret: "s3", results: []map[string]any{{"brain": "team/acme", "knowledge_id": "spec-1", "snippet": "The ledger moves on Sunday.", "provenance": "[team/acme] Spec, spec-1"}}}
	url := g.start(t)
	dir := t.TempDir()
	projet := brainFile(t, dir, "projet.db", "- Projet memory")
	racine := brainFile(t, dir, "racine.db", "- Racine memory")
	file := filepath.Join(dir, "recall")
	env := map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_RECALL_FILE": file,
		"TH_LOCAL_RECALL": fakeRecall(t) + " --recall", "TH_LOCAL_BRAINS": "projet=" + projet + ",racine=" + racine}
	run := func(choice string) string {
		t.Helper()
		if err := os.WriteFile(file, []byte(choice), 0o600); err != nil {
			t.Fatal(err)
		}
		g.calls = nil
		_, out, errOut := hook(t, env, prompt)
		if errOut != "" {
			t.Fatalf("stderr %s", errOut)
		}
		return contextOf(t, out)
	}
	ctx := run("skip=team/other,local:racine\n")
	if len(g.calls) != 1 || !reflect.DeepEqual(g.calls[0]["skip_brains"], []any{"team/other"}) || g.calls[0]["brains"] != nil || !strings.Contains(ctx, "[local:projet]") || strings.Contains(ctx, "[local:racine]") {
		t.Fatalf("skip: %+v\n%s", g.calls, ctx)
	}
	ctx = run("only=team/acme\n")
	if len(g.calls) != 1 || !reflect.DeepEqual(g.calls[0]["brains"], []any{"team/acme"}) || strings.Contains(ctx, "[local:") {
		t.Fatalf("only server: %+v\n%s", g.calls, ctx)
	}
	ctx = run("only=local:racine\n")
	if len(g.calls) != 0 || !strings.Contains(ctx, "[local:racine]") || strings.Contains(ctx, "[local:projet]") || strings.Contains(ctx, "shared") {
		t.Fatalf("only local: %+v\n%s", g.calls, ctx)
	}
}

// The commands change ~/.thalamus/recall (0600, created with its folder);
// brains lists the server brains from th_brains and the local brains, each
// on or off.
func TestTheBrainCommands(t *testing.T) {
	g := &fakeGateway{secret: "s3", listing: []map[string]any{{"brain": "org/ref", "access": "read", "key": true}, {"brain": "team/acme", "access": "write", "key": true}}}
	url := g.start(t)
	home := t.TempDir()
	env := map[string]string{"HOME": home, "TH_GATEWAY": url, "TH_KEY": "s3", "TH_LOCAL_RECALL": "x", "TH_LOCAL_BRAINS": "projet=/p.db,racine=/r.db"}
	for _, args := range [][]string{{"off", "team/acme"}, {"off", "local:racine"}, {"only", "team/acme", "org/ref", "local:projet", "local:racine"}, {"on", "local:racine"}} {
		if code, _, errOut := hook(t, env, "", args...); code != 0 {
			t.Fatalf("%v: %d %s", args, code, errOut)
		}
	}
	path := filepath.Join(home, ".thalamus", "recall")
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "only=team/acme,org/ref,local:projet,local:racine\nskip=team/acme\n" {
		t.Fatalf("file %q: %v", data, err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode())
	}
	code, out, errOut := hook(t, env, "", "brains")
	want := "on   org/ref\noff  team/acme\non   local:projet\non   local:racine\n"
	if code != 0 || out != want {
		t.Fatalf("brains: %d %q %s", code, out, errOut)
	}
	if _, out, _ := hook(t, env, "", "only", "--clear"); out != "" {
		t.Fatal(out)
	}
	if data, _ := os.ReadFile(path); string(data) != "skip=team/acme\n" {
		t.Fatalf("cleared: %q", data)
	}
}

// A command without a brain is a usage error; a file that cannot be
// written, or a gateway that does not answer the listing, is an error.
// brains still lists the local brains when the gateway fails.
func TestTheBrainCommandsReportTheirProblems(t *testing.T) {
	home := t.TempDir()
	for _, args := range [][]string{{"on"}, {"off"}, {"only"}, {"brains", "x"}} {
		if code, _, errOut := hook(t, map[string]string{"HOME": home}, "", args...); code != 2 || !strings.Contains(errOut, "usage") {
			t.Fatalf("%v: %d %s", args, code, errOut)
		}
	}
	if err := os.WriteFile(filepath.Join(home, ".thalamus"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut := hook(t, map[string]string{"HOME": home}, "", "off", "a"); code != 1 || !strings.Contains(errOut, "recall file") {
		t.Fatalf("unwritable: %d %s", code, errOut)
	}
	if code, _, errOut := hook(t, map[string]string{"TH_RECALL_FILE": t.TempDir()}, "", "off", "a"); code != 1 || !strings.Contains(errOut, "recall file") {
		t.Fatalf("a folder: %d %s", code, errOut)
	}
	if code, _, errOut := hook(t, map[string]string{"HOME": home, "TH_LOCAL_RECALL": "x", "TH_LOCAL_BRAINS": "noequals"}, "", "brains"); code != 1 || !strings.Contains(errOut, "TH_LOCAL_BRAINS") {
		t.Fatalf("bad local brains: %d %s", code, errOut)
	}
	code, out, errOut := hook(t, map[string]string{"HOME": t.TempDir(), "TH_GATEWAY": "http://127.0.0.1:1", "TH_KEY": "s3", "TH_LOCAL_RECALL": "x", "TH_LOCAL_BRAINS": "projet=/p.db"}, "", "brains")
	if code != 1 || out != "on   local:projet\n" || !strings.Contains(errOut, "th_brains") {
		t.Fatalf("gateway down: %d %q %s", code, out, errOut)
	}
	if !slices.Equal(parseFilter("only=,a,,b\n").Only, []string{"a", "b"}) {
		t.Fatal("empty names are dropped")
	}
}
