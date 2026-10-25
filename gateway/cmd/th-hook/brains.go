package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// serverBrain is one brain of th_brains.
type serverBrain struct {
	Brain  string `json:"brain"`
	Access string `json:"access"`
	Key    bool   `json:"key"`
}

// serverBrains asks th_brains for the brains the person may read, within
// listDeadline.
func serverBrains(get func(string) string) ([]serverBrain, error) {
	ctx, cancel := context.WithTimeout(context.Background(), listDeadline)
	defer cancel()
	gateway, key, err := settings(get)
	if err != nil {
		return nil, fmt.Errorf("th_brains: %w", err)
	}
	text, err := callTool(ctx, gateway, key, "th_brains", noArgs)
	if err != nil {
		return nil, fmt.Errorf("th_brains: %w", err)
	}
	var out []serverBrain
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return nil, fmt.Errorf("th_brains: %w", err)
	}
	return out, nil
}

// listed is one brain as brains --json gives it.
type listed struct {
	Brain  string `json:"brain"`
	Access string `json:"access"`
	On     bool   `json:"on"`
	OK     bool   `json:"ok"`
	Reason string `json:"reason"`
}

// brainsJSON lists the brains for the panel: the server brains of
// th_brains, or those of the file when the gateway does not answer, then
// the local brains.
func brainsJSON(get func(string) string, stdout, stderr io.Writer) int {
	choice := readFilter(recallFile(get))
	code := 0
	out := struct {
		Gateway string   `json:"gateway"`
		Brains  []listed `json:"brains"`
	}{Brains: []listed{}}
	add := func(name, access string, ok bool, reason string) {
		out.Brains = append(out.Brains, listed{name, access, choice.on(name), ok, reason})
	}
	server, err := serverBrains(get)
	if err != nil {
		out.Gateway = err.Error()
		code = 1
		seen := map[string]bool{}
		for _, n := range append(append([]string{}, choice.Only...), choice.Skip...) {
			if !strings.HasPrefix(n, localPrefix) && !seen[n] {
				seen[n] = true
				add(n, "", false, "gateway does not answer")
			}
		}
	}
	for _, b := range server {
		reason := ""
		if !b.Key {
			reason = "no key"
		}
		add(b.Brain, b.Access, b.Key, reason)
	}
	locals, err := localBrains(get)
	if err != nil {
		fmt.Fprintf(stderr, "th-hook: %v\n", err)
		code = 1
	}
	for _, b := range locals {
		_, err := os.Stat(b.Path)
		reason := ""
		if errors.Is(err, os.ErrNotExist) {
			reason = "no file"
		}
		add(b.Name, "local", reason == "", reason)
	}
	_ = json.NewEncoder(stdout).Encode(out)
	return code
}
