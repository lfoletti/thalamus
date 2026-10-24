package main

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// lateGateway is a gateway that serves one extract and says that ai-image
// did not answer within its deadline.
func lateGateway(slowInit time.Duration) *fakeGateway {
	return &fakeGateway{secret: "s3", slowInit: slowInit,
		results: []map[string]any{{"brain": "coding", "knowledge_id": "spec-1", "snippet": "The ledger moves on Sunday.", "provenance": "[coding] Spec, spec-1"}},
		skipped: []map[string]any{{"brain": "ai-image", "reason": "no answer within 1100 ms"}}}
}

// recalls is the arguments of each th_recall the gateway got.
func (g *fakeGateway) recalls() []map[string]any {
	var out []map[string]any
	for _, c := range g.calls {
		if _, ok := c["query"]; ok {
			out = append(out, c)
		}
	}
	return out
}

// The hook asks th_recall for a deadline shorter than its own timeout, by
// a fixed margin: the gateway answers in time with the brains ready. A
// brain that did not answer is a skipped brain of the block, with its
// reason. The gateway answered: the hook does not remember it as down, and
// the block has no note of a gateway skipped.
func TestTheHookAsksForADeadlineShorterThanItsOwn(t *testing.T) {
	for _, c := range []struct {
		timeout string
		want    float64
	}{{"", 1100}, {"4s", 3600}} {
		dir := t.TempDir()
		g := lateGateway(0)
		env := map[string]string{"TH_ENV": filepath.Join(dir, "env"), "TH_GATEWAY": g.start(t), "TH_KEY": "s3"}
		if c.timeout != "" {
			env["TH_RECALL_TIMEOUT"] = c.timeout
		}
		code, out, _ := hook(t, env, prompt)
		calls := g.recalls()
		if len(calls) != 1 {
			t.Fatalf("%q: calls %+v", c.timeout, g.calls)
		}
		if d := calls[0]["deadline_ms"].(float64); d > c.want || d < c.want-150 {
			t.Fatalf("%q: deadline_ms %v, want %v less the connection", c.timeout, d, c.want)
		}
		block := contextOf(t, out)
		if code != 0 || !strings.Contains(block, "The ledger moves on Sunday.") || !strings.Contains(block, "Skipped: ai-image (no answer within 1100 ms).") || strings.Contains(block, "Shared brains skipped") {
			t.Fatalf("code %d, block %q", code, block)
		}
		if _, err := os.Stat(filepath.Join(dir, "gateway-down")); !os.IsNotExist(err) {
			t.Fatalf("gateway-down: %v", err)
		}
	}
}

// The deadline is counted from the time left after the connection: a
// gateway slow to initialize leaves less of the margin, and the hook asks
// for less.
func TestTheHookCountsTheTimeOfTheConnection(t *testing.T) {
	g := lateGateway(500 * time.Millisecond)
	code, out, errOut := hook(t, map[string]string{"TH_ENV": filepath.Join(t.TempDir(), "env"), "TH_GATEWAY": g.start(t), "TH_KEY": "s3"}, prompt)
	calls := g.recalls()
	if len(calls) != 1 {
		t.Fatalf("calls %+v, stderr %q", g.calls, errOut)
	}
	if d := calls[0]["deadline_ms"].(float64); d > 600 || d < 200 {
		t.Fatalf("deadline_ms %v after an initialize of 500 ms", d)
	}
	if code != 0 || !strings.Contains(contextOf(t, out), "The ledger moves on Sunday.") {
		t.Fatalf("code %d, out %q", code, out)
	}
}

// When the connection leaves less than the margin and the floor, the hook
// does not ask th_recall: it says so on stderr, keeps the local brains, and
// remembers the gateway as down, because a connection that slow makes each
// prompt wait without a shared result.
func TestTheHookDoesNotAskUnderTheFloor(t *testing.T) {
	defer func(f func() time.Time) { now = f }(now)
	now = func() time.Time { return time.Date(2026, 10, 25, 9, 0, 0, 0, time.UTC) }
	dir := t.TempDir()
	g := lateGateway(1100 * time.Millisecond)
	code, out, errOut := hook(t, map[string]string{"TH_ENV": filepath.Join(dir, "env"), "TH_GATEWAY": g.start(t), "TH_KEY": "s3"}, prompt)
	if len(g.recalls()) != 0 || code != 0 || out != "" || !strings.Contains(errOut, "th_recall not asked:") || !strings.Contains(errOut, "ms left after the connection, under 600 ms") {
		t.Fatalf("calls %+v, code %d, out %q, stderr %q", g.calls, code, out, errOut)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "gateway-down")); err != nil || strings.TrimSpace(string(data)) != "2026-10-25T09:00:00Z" {
		t.Fatalf("gateway-down: %q %v", data, err)
	}
}

// sharedDeadline is the time left less 400 ms, and at most the 60 s that
// th_recall accepts. Under 200 ms it is a timeout of the network, that does
// not pass: th_recall is not asked.
func TestTheSharedDeadlineKeepsAMarginAndAFloor(t *testing.T) {
	for in, want := range map[time.Duration]time.Duration{
		1500 * time.Millisecond: 1100 * time.Millisecond,
		4 * time.Second:         3600 * time.Millisecond,
		600 * time.Millisecond:  200 * time.Millisecond,
		601 * time.Millisecond:  201 * time.Millisecond,
		2 * time.Minute:         60 * time.Second,
	} {
		if got, err := sharedDeadline(in); got != want || err != nil {
			t.Fatalf("%v: %v %v, want %v", in, got, err, want)
		}
	}
	for in, want := range map[time.Duration]string{
		599 * time.Millisecond: "th_recall not asked: 599 ms left after the connection, under 600 ms",
		time.Millisecond:       "th_recall not asked: 1 ms left after the connection, under 600 ms",
		-time.Second:           "th_recall not asked: -1000 ms left after the connection, under 600 ms",
	} {
		_, err := sharedDeadline(in)
		var ne net.Error
		if err == nil || err.Error() != want || !errors.As(err, &ne) || !ne.Timeout() || ne.Temporary() {
			t.Fatalf("%v: %v, a timeout that does not pass", in, err)
		}
	}
}
