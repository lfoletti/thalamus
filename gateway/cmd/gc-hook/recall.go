package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// contextBudget bounds what the hook hands Claude Code: past about ten
// thousand characters the whole context is replaced by a preview, so the
// block is cut whole extracts at a time under a margin below that.
const contextBudget = 9000

const usage = "usage: gc-hook recall  (hook payload on stdin; GC_GATEWAY, GC_KEY, GC_RECALL_TOPK)"

type hookInput struct {
	Prompt string `json:"prompt"`
}

type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// extract is one result of gc_recall, as the gateway returns it.
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

// run dispatches: 0 done or silent, 2 wrong command line.
func run(args []string, stdin io.Reader, env func(string) string, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] != "recall" {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	text, err := recall(stdin, env)
	if err != nil {
		fmt.Fprintf(stderr, "gc-hook: %v\n", err)
		return 0
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

// recall reads the prompt, asks the gateway, and renders the context. An
// empty string with no error means there is nothing to inject.
func recall(stdin io.Reader, env func(string) string) (string, error) {
	var in hookInput
	if err := json.NewDecoder(stdin).Decode(&in); err != nil {
		return "", fmt.Errorf("hook input: %w", err)
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" || strings.HasPrefix(prompt, "/") || len(strings.Fields(prompt)) < 3 {
		return "", nil
	}
	gateway := strings.TrimRight(strings.TrimSpace(env("GC_GATEWAY")), "/")
	if gateway == "" {
		return "", errors.New("GC_GATEWAY is not set")
	}
	key := strings.TrimSpace(env("GC_KEY"))
	if key == "" {
		return "", errors.New("GC_KEY is not set")
	}
	topK := 3
	if v := strings.TrimSpace(env("GC_RECALL_TOPK")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return "", fmt.Errorf("GC_RECALL_TOPK %q is not a positive integer", v)
		}
		topK = n
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	out, err := ask(ctx, gateway, key, prompt, topK)
	if err != nil {
		return "", err
	}
	text := render(out)
	if text == "" && len(out.Results) > 0 {
		return "", fmt.Errorf("%d extracts, not one fits the context budget", len(out.Results))
	}
	return text, nil
}

type bearer struct{ token string }

func (b bearer) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(req)
}

// ask calls gc_recall on the gateway's federated endpoint.
func ask(ctx context.Context, gateway, key, prompt string, topK int) (recallOut, error) {
	tr := &mcp.StreamableClientTransport{Endpoint: gateway + "/mcp", HTTPClient: &http.Client{Transport: bearer{key}}, MaxRetries: -1}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "gc-hook", Version: "0"}, nil).Connect(ctx, tr, nil)
	if err != nil {
		return recallOut{}, fmt.Errorf("connect to %s: %w", gateway, err)
	}
	defer func() { _ = session.Close() }()
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "gc_recall", Arguments: map[string]any{"query": prompt, "top_k": topK}})
	if err != nil {
		return recallOut{}, fmt.Errorf("gc_recall: %w", err)
	}
	text := ""
	if len(res.Content) > 0 {
		if tc, ok := res.Content[0].(*mcp.TextContent); ok {
			text = tc.Text
		}
	}
	if res.IsError {
		return recallOut{}, errors.New(text)
	}
	var out recallOut
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return recallOut{}, fmt.Errorf("decode gc_recall: %w", err)
	}
	return out, nil
}

// render writes the context block: a header, one line per extract with its
// provenance, the skipped brains, and how to ask for more. It drops
// extracts from the end until the block fits the budget.
func render(out recallOut) string {
	if len(out.Results) == 0 {
		return ""
	}
	header := fmt.Sprintf("granular_cortex: %d extracts from the shared brains, with their provenance.", len(out.Results))
	var lines []string
	for _, r := range out.Results {
		p := r.Provenance
		if len(r.AlsoIn) > 0 {
			p += " (also in " + strings.Join(r.AlsoIn, ", ") + ")"
		}
		lines = append(lines, "- "+p+": "+strings.TrimSpace(r.Snippet))
	}
	footer := "Ask for more with gc_recall; list the brains with gc_brains."
	if len(out.Skipped) > 0 {
		var parts []string
		for _, s := range out.Skipped {
			parts = append(parts, s.Brain+" ("+s.Reason+")")
		}
		footer = "Skipped: " + strings.Join(parts, ", ") + ".\n" + footer
	}
	return fit(header, lines, footer, contextBudget)
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
		note = fmt.Sprintf("(%d more not shown; ask for them with gc_recall.)", len(lines)-len(kept))
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
