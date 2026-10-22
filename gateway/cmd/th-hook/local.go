package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// localBudget bounds the sections of the local brains in the context.
const localBudget = 3000

// localBrain is a CortexDB brain on this machine, named local:<name>.
type localBrain struct{ Name, Path string }

// localBrains reads TH_LOCAL_BRAINS, name=path pairs separated by commas;
// a path may start with ~/. Without it, the brain of the project
// (CORTEXDB_PATH) is the only one. Without TH_LOCAL_RECALL there is none.
func localBrains(get func(string) string) ([]localBrain, error) {
	if get("TH_LOCAL_RECALL") == "" {
		return nil, nil
	}
	list := get("TH_LOCAL_BRAINS")
	if list == "" {
		if p := get("CORTEXDB_PATH"); p != "" {
			return []localBrain{{localPrefix + "project", p}}, nil
		}
		return nil, nil
	}
	var out []localBrain
	seen := map[string]bool{}
	for _, pair := range strings.Split(list, ",") {
		name, path, _ := strings.Cut(strings.TrimSpace(pair), "=")
		name, path = strings.TrimSpace(name), strings.TrimSpace(path)
		if name == "" || path == "" || seen[name] {
			return nil, fmt.Errorf("TH_LOCAL_BRAINS: %q is not a new name=path", pair)
		}
		seen[name] = true
		if rest, ok := strings.CutPrefix(path, "~/"); ok {
			path = filepath.Join(get("HOME"), rest)
		}
		out = append(out, localBrain{localPrefix + name, path})
	}
	return out, nil
}

// childEnv is the environment of the native recall for one brain: the
// brain's path, and no remote brain, which would take the place of the file.
func childEnv(environ []string, path string) []string {
	var out []string
	for _, kv := range environ {
		if !strings.HasPrefix(kv, "CORTEXDB_PATH=") && !strings.HasPrefix(kv, "CORTEXDB_REMOTE=") {
			out = append(out, kv)
		}
	}
	return append(out, "CORTEXDB_PATH="+path)
}

// askLocal runs the native recall of CortexDB on one brain with the hook
// payload, and returns its context. A brain whose file is absent gives
// nothing, as with the native hook.
func askLocal(ctx context.Context, command []string, b localBrain, payload []byte) (string, error) {
	if _, err := os.Stat(b.Path); err != nil {
		return "", nil
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Env = childEnv(os.Environ(), b.Path)
	cmd.Stdin = bytes.NewReader(payload)
	// A child of the command that outlives it must not hold the hook.
	cmd.WaitDelay = 100 * time.Millisecond
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return "", nil
	}
	var v hookOutput
	if err := json.Unmarshal(out, &v); err != nil {
		return "", fmt.Errorf("decode the native recall: %w", err)
	}
	return v.HookSpecificOutput.AdditionalContext, nil
}

// renderLocal writes one section per brain, in order, under budget: each
// section is cut whole lines at a time, and a brain with no line that fits
// is left out.
func renderLocal(brains []localBrain, texts []string, budget int) string {
	var parts []string
	for i, b := range brains {
		if texts[i] == "" {
			continue
		}
		head := "[" + b.Name + "]"
		used := utf8.RuneCountInString(head) + 1
		var kept []string
		for _, line := range strings.Split(texts[i], "\n") {
			n := utf8.RuneCountInString(line) + 1
			if used+n > budget {
				break
			}
			kept, used = append(kept, line), used+n
		}
		if len(kept) == 0 {
			continue
		}
		parts = append(parts, head+"\n"+strings.Join(kept, "\n"))
		budget -= used + 1
	}
	return strings.Join(parts, "\n\n")
}
