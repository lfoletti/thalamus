package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeRecall writes a stand-in for cortexdb-mcp --recall. The content of
// the brain file drives it: FAIL exits 1, BAD prints text that is not JSON,
// EMPTY prints nothing, SLEEP waits; any other content comes back as the
// context, with the CORTEXDB_REMOTE it saw. It keeps the payload it read
// next to the brain file.
func fakeRecall(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-recall")
	script := `#!/bin/sh
[ "$1" = "--recall" ] || exit 3
[ -n "$CORTEXDB_PATH" ] || exit 4
cat > "$CORTEXDB_PATH.in"
body="$(cat "$CORTEXDB_PATH")"
case "$body" in
FAIL) echo "no brain" >&2; exit 1 ;;
BAD) echo "not json"; exit 0 ;;
EMPTY) exit 0 ;;
SLEEP) sleep 5; exit 0 ;;
esac
printf '{"hookSpecificOutput":{"hookEventName":"UserPromptSubmit","additionalContext":"%s remote=%s"}}\n' "$body" "${CORTEXDB_REMOTE:-none}"
`
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

// brainFile writes a local brain file whose content drives fakeRecall.
func brainFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

const prompt = `{"prompt":"when does the ledger move?","cwd":"/x"}`

// The hook runs the native recall once per local brain, in the order of
// TH_LOCAL_BRAINS, with CORTEXDB_PATH set to that brain and without
// CORTEXDB_REMOTE, on the same payload. The block has the shared extracts,
// then one section per local brain. A brain whose file is absent is left
// out without a word.
func TestTheHookAddsTheLocalBrains(t *testing.T) {
	t.Setenv("CORTEXDB_REMOTE", "grpc.example:7000")
	g := &fakeGateway{secret: "s3", results: []map[string]any{{"brain": "team/acme", "knowledge_id": "spec-1", "snippet": "The ledger moves on Sunday.", "provenance": "[team/acme] Spec, spec-1"}}}
	url := g.start(t)
	dir := t.TempDir()
	projet := brainFile(t, dir, "projet.db", "- Projet memory")
	racine := brainFile(t, dir, "racine.db", "- Racine memory")
	env := map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "HOME": dir,
		"TH_LOCAL_RECALL": fakeRecall(t) + " --recall",
		"TH_LOCAL_BRAINS": "projet=" + projet + ", racine=" + racine + ",absent=" + filepath.Join(dir, "absent.db")}
	code, out, errOut := hook(t, env, prompt)
	if code != 0 || errOut != "" {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	ctx := contextOf(t, out)
	shared := strings.Index(ctx, "Thalamus: 1 extracts from the shared brains")
	p := strings.Index(ctx, "[local:projet]\n- Projet memory remote=none")
	r := strings.Index(ctx, "[local:racine]\n- Racine memory remote=none")
	if shared != 0 || p < shared || r < p || strings.Contains(ctx, "absent") {
		t.Fatalf("context:\n%s", ctx)
	}
	for _, f := range []string{projet, racine} {
		if in, _ := os.ReadFile(f + ".in"); string(in) != prompt {
			t.Fatalf("payload of %s: %q", f, in)
		}
	}
}

// Without TH_LOCAL_BRAINS the only local brain is the brain of the project,
// CORTEXDB_PATH. Without TH_LOCAL_RECALL there is no local brain at all. A
// path may start with ~/. Both settings may come from the env file.
func TestTheLocalBrainsHaveADefaultAndSettings(t *testing.T) {
	g := &fakeGateway{secret: "s3"}
	url := g.start(t)
	dir := t.TempDir()
	projet := brainFile(t, dir, "projet.db", "- Projet memory")
	fake := fakeRecall(t)
	_, out, _ := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_LOCAL_RECALL": fake + " --recall", "CORTEXDB_PATH": projet}, prompt)
	if !strings.Contains(contextOf(t, out), "[local:project]\n- Projet memory") {
		t.Fatalf("default: %s", out)
	}
	if _, out, _ := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "CORTEXDB_PATH": projet}, prompt); out != "" {
		t.Fatalf("no command: %s", out)
	}
	if _, out, _ := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_LOCAL_RECALL": fake + " --recall"}, prompt); out != "" {
		t.Fatalf("no project brain: %s", out)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".thalamus"), 0o700); err != nil {
		t.Fatal(err)
	}
	brainFile(t, dir, ".thalamus/env", "TH_GATEWAY="+url+"\nTH_KEY=s3\nTH_LOCAL_RECALL="+fake+" --recall\nTH_LOCAL_BRAINS=mine=~/projet.db\n")
	if _, out, _ := hook(t, map[string]string{"HOME": dir}, prompt); !strings.Contains(contextOf(t, out), "[local:mine]\n- Projet memory") {
		t.Fatalf("env file: %s", out)
	}
}

// A local brain that fails, answers text that is not JSON, or does not
// answer in time is left out, with its name on stderr; the others stay. A
// gateway that is not set leaves the local brains. A wrong TH_LOCAL_BRAINS
// leaves the shared brains.
func TestALocalProblemLeavesTheRest(t *testing.T) {
	defer func(d time.Duration) { deadline = d }(deadline)
	deadline = time.Second
	g := &fakeGateway{secret: "s3", results: []map[string]any{{"brain": "team/acme", "knowledge_id": "spec-1", "snippet": "The ledger moves on Sunday.", "provenance": "[team/acme] Spec, spec-1"}}}
	url := g.start(t)
	dir := t.TempDir()
	fake := fakeRecall(t) + " --recall"
	brains := fmt.Sprintf("f=%s,b=%s,s=%s,e=%s,ok=%s", brainFile(t, dir, "f.db", "FAIL"), brainFile(t, dir, "b.db", "BAD"), brainFile(t, dir, "s.db", "SLEEP"), brainFile(t, dir, "e.db", "EMPTY"), brainFile(t, dir, "ok.db", "- Kept memory"))
	began := time.Now()
	code, out, errOut := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_LOCAL_RECALL": fake, "TH_LOCAL_BRAINS": brains}, prompt)
	if took := time.Since(began); took > 3*time.Second {
		t.Fatalf("a slow local brain held the hook %v", took)
	}
	ctx := contextOf(t, out)
	if code != 0 || !strings.Contains(ctx, "The ledger moves on Sunday.") || !strings.Contains(ctx, "[local:ok]\n- Kept memory") || strings.Contains(ctx, "[local:e]") {
		t.Fatalf("context:\n%s", ctx)
	}
	for _, want := range []string{"local:f: ", "local:b: ", "local:s: "} {
		if !strings.Contains(errOut, want) {
			t.Fatalf("stderr lacks %q: %s", want, errOut)
		}
	}
	if strings.Contains(errOut, "local:e") || strings.Contains(errOut, "local:ok") {
		t.Fatalf("stderr: %s", errOut)
	}
	_, out, errOut = hook(t, map[string]string{"HOME": dir, "TH_LOCAL_RECALL": fake, "TH_LOCAL_BRAINS": "ok=" + filepath.Join(dir, "ok.db")}, prompt)
	if !strings.Contains(contextOf(t, out), "[local:ok]") || !strings.Contains(errOut, "TH_GATEWAY") {
		t.Fatalf("no gateway: %s / %s", out, errOut)
	}
	for _, bad := range []string{"noequals", "=x.db", "a=", "a=x.db,a=y.db"} {
		_, out, errOut = hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_LOCAL_RECALL": fake, "TH_LOCAL_BRAINS": bad}, prompt)
		if !strings.Contains(contextOf(t, out), "The ledger moves on Sunday.") || !strings.Contains(errOut, "TH_LOCAL_BRAINS") {
			t.Fatalf("%q: %s / %s", bad, out, errOut)
		}
	}
	_, out, errOut = hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_LOCAL_RECALL": filepath.Join(dir, "missing-binary"), "TH_LOCAL_BRAINS": "ok=" + filepath.Join(dir, "ok.db")}, prompt)
	if !strings.Contains(contextOf(t, out), "The ledger moves on Sunday.") || !strings.Contains(errOut, "local:ok: ") {
		t.Fatalf("missing binary: %s / %s", out, errOut)
	}
}

// The local brains share localBudget, in the order of the list, cut whole
// lines at a time; a brain with no line that fits is left out. The shared
// extracts take the rest of contextBudget.
func TestTheLocalBrainsShareTheirBudget(t *testing.T) {
	results := []map[string]any{{"brain": "b", "knowledge_id": "e", "snippet": "short", "provenance": "[b] e"}}
	for i := 0; i < 8; i++ {
		results = append(results, map[string]any{"brain": "b", "knowledge_id": fmt.Sprint("d-", i), "snippet": strings.Repeat("s", 1000), "provenance": fmt.Sprint("[b] d-", i)})
	}
	g := &fakeGateway{secret: "s3", results: results}
	url := g.start(t)
	dir := t.TempDir()
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, fmt.Sprintf("- memory %02d %s", i, strings.Repeat("m", 90)))
	}
	big := brainFile(t, dir, "big.db", strings.Join(lines, `\n`))
	next := brainFile(t, dir, "next.db", "- "+strings.Repeat("n", 200))
	_, out, errOut := hook(t, map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_LOCAL_RECALL": fakeRecall(t) + " --recall", "TH_LOCAL_BRAINS": "big=" + big + ",next=" + next}, prompt)
	ctx := contextOf(t, out)
	i := strings.Index(ctx, "[local:big]")
	if i < 0 || len([]rune(ctx[i:])) > localBudget || len([]rune(ctx)) > contextBudget || !strings.Contains(ctx, "memory 00") || strings.Contains(ctx, "memory 39") || strings.Contains(ctx, "[local:next]") {
		t.Fatalf("errOut %s, len %d, local %d:\n%s", errOut, len(ctx), len(ctx)-i, ctx)
	}
	if !strings.Contains(ctx, "[b] e: short") || !strings.Contains(ctx, "more not shown") {
		t.Fatalf("shared:\n%s", ctx[:i])
	}
}
