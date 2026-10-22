package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// contextBudget bounds what the hook hands Claude Code: past about ten
// thousand characters the whole context is replaced by a preview, so the
// block is cut whole extracts at a time under a margin below that.
const contextBudget = 9000

const usage = `usage: th-hook recall               (hook payload on stdin; TH_GATEWAY, TH_KEY, TH_RECALL_TOPK, TH_LOCAL_RECALL, TH_LOCAL_BRAINS)
       th-hook brains               (the brains of the automatic recall, on or off)
       th-hook on|off <brain>...    (local brains are local:<name>)
       th-hook only <brain>... | --clear
       th-hook install              (TH_LOCAL_RECALL from the CortexDB binary of this machine)`

// deadline bounds one recall, the gateway and the local brains together.
var deadline = 8 * time.Second

type hookInput struct {
	Prompt string `json:"prompt"`
}

type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// extract is one result of th_recall, as the gateway returns it.
type extract struct {
	Snippet    string   `json:"snippet"`
	Provenance string   `json:"provenance"`
	AlsoIn     []string `json:"also_in"`
}

type recallOut struct {
	Results []extract `json:"results"`
	Skipped []struct {
		Brain  string `json:"brain"`
		Reason string `json:"reason"`
	} `json:"skipped"`
}

// run dispatches: 0 done or silent, 1 a command failed, 2 wrong command
// line. The recall never fails: its problems go to stderr.
func run(args []string, stdin io.Reader, env func(string) string, stdout, stderr io.Writer) int {
	get := source(env)
	switch {
	case len(args) == 1 && args[0] == "recall":
	case len(args) == 1 && args[0] == "install":
		if err := install(get, stdout); err != nil {
			fmt.Fprintf(stderr, "th-hook: %v\n", err)
			return 1
		}
		return 0
	case len(args) == 1 && args[0] == "brains":
		return listBrains(get, stdout, stderr)
	case len(args) > 1 && (args[0] == "on" || args[0] == "off" || args[0] == "only"):
		path := recallFile(get)
		if err := writeFilter(path, readFilter(path).with(args[0], args[1:])); err != nil {
			fmt.Fprintf(stderr, "th-hook: %v\n", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
	text, problems := recall(stdin, get)
	for _, err := range problems {
		fmt.Fprintf(stderr, "th-hook: %v\n", err)
	}
	if text == "" {
		return 0
	}
	var out hookOutput
	out.HookSpecificOutput.HookEventName = "UserPromptSubmit"
	out.HookSpecificOutput.AdditionalContext = text
	_ = json.NewEncoder(stdout).Encode(out)
	return 0
}

// recall reads the prompt, then asks the gateway and the local brains at
// the same time, under the person's filter. The shared extracts come first,
// then one section per local brain. Each side that fails is a problem and
// leaves the other.
func recall(stdin io.Reader, get func(string) string) (string, []error) {
	payload, _ := io.ReadAll(stdin)
	var in hookInput
	if err := json.Unmarshal(payload, &in); err != nil {
		return "", []error{fmt.Errorf("hook input: %w", err)}
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" || strings.HasPrefix(prompt, "/") || len(strings.Fields(prompt)) < 3 {
		return "", nil
	}
	topK := 3
	if v := get("TH_RECALL_TOPK"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return "", []error{fmt.Errorf("TH_RECALL_TOPK %q is not a positive integer", v)}
		}
		topK = n
	}
	choice := readFilter(recallFile(get))
	var problems []error
	all, err := localBrains(get)
	if err != nil {
		problems = append(problems, err)
	}
	var locals []localBrain
	for _, b := range all {
		if choice.on(b.Name) {
			locals = append(locals, b)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	var wg sync.WaitGroup
	texts := make([]string, len(locals))
	errs := make([]error, len(locals))
	command := strings.Fields(get("TH_LOCAL_RECALL"))
	for i, b := range locals {
		wg.Add(1)
		go func() {
			defer wg.Done()
			texts[i], errs[i] = askLocal(ctx, command, b, payload)
		}()
	}
	var out recallOut
	var sharedErr error
	if only, skip, ask := choice.shared(); ask {
		out, sharedErr = askShared(ctx, get, prompt, topK, only, skip)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", locals[i].Name, err))
		}
	}
	if sharedErr != nil {
		problems = append(problems, sharedErr)
	}

	local := renderLocal(locals, texts, localBudget)
	budget := contextBudget
	if local != "" {
		budget -= utf8.RuneCountInString(local) + 2
	}
	shared := render(out, budget)
	if shared == "" && len(out.Results) > 0 {
		problems = append(problems, fmt.Errorf("%d extracts, not one fits the context budget", len(out.Results)))
	}
	var parts []string
	for _, p := range []string{shared, local} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "\n\n"), problems
}

// askShared asks th_recall with the person's key, under the filter.
func askShared(ctx context.Context, get func(string) string, prompt string, topK int, only, skip []string) (recallOut, error) {
	gateway, key, err := settings(get)
	if err != nil {
		return recallOut{}, err
	}
	args := map[string]any{"query": prompt, "top_k": topK}
	if len(only) > 0 {
		args["brains"] = only
	}
	if len(skip) > 0 {
		args["skip_brains"] = skip
	}
	text, err := callTool(ctx, gateway, key, "th_recall", args)
	if err != nil {
		return recallOut{}, err
	}
	var out recallOut
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return recallOut{}, fmt.Errorf("decode th_recall: %w", err)
	}
	return out, nil
}

// listBrains prints each brain of the automatic recall, on or off: the
// server brains from th_brains, then the local brains.
func listBrains(get func(string) string, stdout, stderr io.Writer) int {
	choice := readFilter(recallFile(get))
	code := 0
	var server []struct {
		Brain string `json:"brain"`
	}
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	gateway, key, err := settings(get)
	if err == nil {
		var text string
		if text, err = callTool(ctx, gateway, key, "th_brains", map[string]any{}); err == nil {
			err = json.Unmarshal([]byte(text), &server)
		}
	}
	if err != nil {
		fmt.Fprintf(stderr, "th-hook: th_brains: %v\n", err)
		code = 1
	}
	var names []string
	for _, b := range server {
		names = append(names, b.Brain)
	}
	locals, err := localBrains(get)
	if err != nil {
		fmt.Fprintf(stderr, "th-hook: %v\n", err)
		code = 1
	}
	for _, b := range locals {
		names = append(names, b.Name)
	}
	for _, n := range names {
		state := "off"
		if choice.on(n) {
			state = "on"
		}
		fmt.Fprintf(stdout, "%-4s %s\n", state, n)
	}
	return code
}

type bearer struct{ token string }

func (b bearer) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(req)
}

// callTool calls one tool on the gateway's federated endpoint and returns
// the text of its answer.
func callTool(ctx context.Context, gateway, key, name string, args map[string]any) (string, error) {
	tr := &mcp.StreamableClientTransport{Endpoint: gateway + "/mcp", HTTPClient: &http.Client{Transport: bearer{key}}, MaxRetries: -1}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "th-hook", Version: "0"}, nil).Connect(ctx, tr, nil)
	if err != nil {
		return "", fmt.Errorf("connect to %s: %w", gateway, err)
	}
	defer func() { _ = session.Close() }()
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	text := ""
	if len(res.Content) > 0 {
		if tc, ok := res.Content[0].(*mcp.TextContent); ok {
			text = tc.Text
		}
	}
	if res.IsError {
		return "", errors.New(text)
	}
	return text, nil
}

// render writes the context block: a header, one line per extract with its
// provenance, the skipped brains, and how to ask for more. It drops
// extracts from the end until the block fits the budget.
func render(out recallOut, budget int) string {
	if len(out.Results) == 0 {
		return ""
	}
	header := fmt.Sprintf("Thalamus: %d extracts from the shared brains, with their provenance.", len(out.Results))
	var lines []string
	for _, r := range out.Results {
		p := r.Provenance
		if len(r.AlsoIn) > 0 {
			p += " (also in " + strings.Join(r.AlsoIn, ", ") + ")"
		}
		lines = append(lines, "- "+p+": "+strings.TrimSpace(r.Snippet))
	}
	footer := "An extract is a cut: read a document whole with th_get {brain, knowledge_id} from its provenance line. Ask for more with th_recall; list the brains with th_brains."
	if len(out.Skipped) > 0 {
		var parts []string
		for _, s := range out.Skipped {
			parts = append(parts, s.Brain+" ("+s.Reason+")")
		}
		footer = "Skipped: " + strings.Join(parts, ", ") + ".\n" + footer
	}
	return fit(header, lines, footer, budget)
}

// fit joins header, lines and footer, dropping lines from the end until
// the block fits the budget, and says how many it left out. When not one
// line fits, it returns "": a heading over nothing is not worth its place.
func fit(header string, lines []string, footer string, budget int) string {
	size := func(parts ...string) int {
		n := 0
		for _, p := range parts {
			n += utf8.RuneCountInString(p) + 1
		}
		return n
	}
	kept := lines
	note := ""
	for len(kept) > 0 && size(append([]string{header, footer, note}, kept...)...) > budget {
		kept = kept[:len(kept)-1]
		note = fmt.Sprintf("(%d more not shown; ask for them with th_recall.)", len(lines)-len(kept))
	}
	if len(kept) == 0 {
		return ""
	}
	parts := append([]string{header}, kept...)
	if note != "" {
		parts = append(parts, note)
	}
	return strings.Join(append(parts, footer), "\n")
}

// source reads a setting from the environment or, when it is not there,
// from the env file that th-gateway login writes (TH_ENV, or
// ~/.thalamus/env).
func source(env func(string) string) func(string) string {
	path := env("TH_ENV")
	if path == "" {
		path = filepath.Join(env("HOME"), ".thalamus", "env")
	}
	data, _ := os.ReadFile(path)
	file := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		k, v, _ := strings.Cut(line, "=")
		file[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return func(k string) string {
		if v := strings.TrimSpace(env(k)); v != "" {
			return v
		}
		return file[k]
	}
}

// settings gives TH_GATEWAY and TH_KEY.
func settings(get func(string) string) (gateway, key string, err error) {
	gateway, key = get("TH_GATEWAY"), get("TH_KEY")
	if gateway == "" {
		return "", "", errors.New("TH_GATEWAY is not set, and no env file from th-gateway login")
	}
	if key == "" {
		return "", "", errors.New("TH_KEY is not set, and no env file from th-gateway login")
	}
	return strings.TrimRight(gateway, "/"), key, nil
}
