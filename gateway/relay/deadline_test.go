package relay

import (
	"bytes"
	"context"
	"fmt"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/keys"
	"github.com/lfoletti/thalamus/gateway/rights"
)

// recallWith calls th_recall as lfoletti, who has a key on both brains,
// and gives the result or the text of the tool error, and how long the call
// took, without the connection.
func recallWith(t *testing.T, b *bench, args map[string]any) (recallOut, string, time.Duration) {
	t.Helper()
	k, _ := keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	_, _ = keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	s := b.federated(t, k.Secret)
	began := time.Now()
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: "th_recall", Arguments: args})
	took := time.Since(began)
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		return recallOut{}, res.Content[0].(*mcp.TextContent).Text, took
	}
	return decode[recallOut](t, res), "", took
}

// slowBench is a bench where team/other answers knowledge_search after slow.
func slowBench(t *testing.T, slow time.Duration) *bench {
	t.Helper()
	b := start(t)
	b.brain.hits = map[string][]fakeHit{
		"team-acme": {{"a-1", "Ledger plan", "The ledger moves on Sunday."}},
		"other":     {{"o-1", "Other", "Something else."}},
	}
	b.brain.slow = map[string]time.Duration{"other": slow}
	return b
}

// waitCut waits, 2 s at most, for a slow call to see its caller go away.
func waitCut(t *testing.T, b *bench) {
	t.Helper()
	for i := 0; !b.brain.cut.Load(); i++ {
		if i == 100 {
			t.Fatal("the late call was not cancelled")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// With deadline_ms, th_recall answers at the deadline with the brains that
// answered: a brain still asked is in skipped, its call is cancelled, and
// the audit names only what was read.
func TestRecallAnswersAtTheDeadlineWithTheBrainsThatAnswered(t *testing.T) {
	b := slowBench(t, 10*time.Second)
	got, msg, _ := recallWith(t, b, map[string]any{"query": "ledger", "deadline_ms": 200})
	if msg != "" {
		t.Fatal(msg)
	}
	if len(got.Results) != 1 || got.Results[0].KnowledgeID != "a-1" {
		t.Fatalf("results: %+v", got.Results)
	}
	if len(got.Skipped) != 1 || got.Skipped[0].Brain != "team/other" || got.Skipped[0].Reason != "no answer within 200 ms" {
		t.Fatalf("skipped: %+v", got.Skipped)
	}
	ids := b.auditLines(t)[0]["ids"].(map[string]any)
	if fmt.Sprint(ids["brains"]) != "[team/acme]" || fmt.Sprint(ids["knowledge_ids"]) != "[team/acme:a-1]" {
		t.Fatalf("audit: %+v", ids)
	}
	waitCut(t, b)
}

// The deadline holds both ways: th_recall answers within 900 ms of a
// deadline of 500 ms, and serves a brain that answers after three quarters
// of it.
func TestRecallKeepsItsDeadlineBothWays(t *testing.T) {
	b := slowBench(t, 10*time.Second)
	b.brain.slow["team-acme"] = 400 * time.Millisecond
	got, msg, took := recallWith(t, b, map[string]any{"query": "ledger", "deadline_ms": 500})
	if msg != "" || took > 900*time.Millisecond {
		t.Fatalf("took %v for a deadline of 500 ms: %q", took, msg)
	}
	if len(got.Results) != 1 || got.Results[0].KnowledgeID != "a-1" || len(got.Skipped) != 1 || got.Skipped[0].Brain != "team/other" {
		t.Fatalf("%+v", got)
	}
}

// direct is the server of the bench, called without MCP.
func direct(t *testing.T, b *bench) *Server {
	t.Helper()
	_, _ = keys.Issue(b.keysFile, "lfoletti", "team/acme", "team-acme", rights.Write)
	_, _ = keys.Issue(b.keysFile, "lfoletti", "team/other", "other", rights.Write)
	return New(rights.Static(b.rights), b.a)
}

// recall cancels a late call itself: a caller whose context never ends
// still sees the slow call stop.
func TestRecallCancelsTheLateCallItself(t *testing.T) {
	b := slowBench(t, 10*time.Second)
	res, err := direct(t, b).recall(context.Background(), "lfoletti", recallArgs{Query: "ledger", DeadlineMS: 200})
	if err != nil || len(res.Skipped) != 1 {
		t.Fatalf("%+v %v", res, err)
	}
	waitCut(t, b)
}

// A late brain leaves no goroutine behind: its answer goes into the room the
// channel keeps for it, after the recall has answered.
func TestRecallLeavesNoGoroutineBehind(t *testing.T) {
	b := slowBench(t, 0)
	s := direct(t, b)
	if _, err := s.recall(context.Background(), "lfoletti", recallArgs{Query: "ledger"}); err != nil {
		t.Fatal(err)
	}
	b.brain.slow["other"] = 10 * time.Second
	if _, err := s.recall(context.Background(), "lfoletti", recallArgs{Query: "ledger", DeadlineMS: 200}); err != nil {
		t.Fatal(err)
	}
	waitCut(t, b)
	for i := 0; inRecall() > 0; i++ {
		if i == 100 {
			t.Fatalf("%d goroutines of recall stay", inRecall())
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// inRecall counts the goroutines that recall started and that still run.
func inRecall() int {
	var buf bytes.Buffer
	_ = pprof.Lookup("goroutine").WriteTo(&buf, 2)
	n := 0
	for _, g := range strings.Split(buf.String(), "\n\n") {
		if strings.Contains(g, "(*Server).recall") {
			n++
		}
	}
	return n
}

// gather keeps every answer of the brains, and at the deadline the answers
// already sent, without a wait for the others.
func TestGatherKeepsTheAnswersAlreadySent(t *testing.T) {
	closed := make(chan struct{})
	close(closed)
	for range 20 {
		answers := make(chan brainAnswer, 3)
		answers <- brainAnswer{name: "a"}
		answers <- brainAnswer{name: "b"}
		if got := gather(answers, 3, closed); len(got) != 2 || got[0].name != "a" || got[1].name != "b" {
			t.Fatalf("past the deadline: %+v", got)
		}
	}
	answers := make(chan brainAnswer, 2)
	answers <- brainAnswer{name: "a"}
	answers <- brainAnswer{name: "b"}
	if got := gather(answers, 2, nil); len(got) != 2 {
		t.Fatalf("without a deadline: %+v", got)
	}
}

// The fusion follows the order of the brains, not the order of the
// answers: the same passage in two brains belongs to the first brain, even
// when it answers last.
func TestRecallFusesInTheOrderOfTheBrains(t *testing.T) {
	b := slowBench(t, 0)
	b.brain.hits["other"] = []fakeHit{{"o-1", "Copy", "The ledger moves on Sunday."}}
	b.brain.slow["team-acme"] = 100 * time.Millisecond
	got, msg, _ := recallWith(t, b, map[string]any{"query": "ledger"})
	if msg != "" || len(got.Results) != 1 || got.Results[0].Brain != "team/acme" || fmt.Sprint(got.Results[0].AlsoIn) != "[team/other]" {
		t.Fatalf("%+v %q", got, msg)
	}
}

// The skipped brains come in order of brain, not in the order of the answers.
func TestRecallSortsTheSkippedBrains(t *testing.T) {
	b := slowBench(t, 0)
	b.brain.fail = true
	b.brain.slow["team-acme"] = 100 * time.Millisecond
	got, msg, _ := recallWith(t, b, map[string]any{"query": "ledger"})
	if msg != "" || len(got.Skipped) != 2 || got.Skipped[0].Brain != "team/acme" || got.Skipped[1].Brain != "team/other" {
		t.Fatalf("%+v %q", got, msg)
	}
}

// Without deadline_ms, or with 0, th_recall waits for every brain, as before.
func TestRecallWithoutADeadlineWaitsForEveryBrain(t *testing.T) {
	for _, args := range []map[string]any{{"query": "ledger"}, {"query": "ledger", "deadline_ms": 0}} {
		b := slowBench(t, 300*time.Millisecond)
		got, msg, _ := recallWith(t, b, args)
		if msg != "" || len(got.Results) != 2 || len(got.Skipped) != 0 {
			t.Fatalf("%v: %+v %q", args, got, msg)
		}
	}
}

// A brain that failed answered: it is in skipped, not late, and the audit
// does not name it as read.
func TestRecallAuditsAFailedBrainAsNotRead(t *testing.T) {
	b := slowBench(t, 0)
	b.brain.failOn = "other"
	got, msg, _ := recallWith(t, b, map[string]any{"query": "ledger", "deadline_ms": 5000})
	if msg != "" || len(got.Results) != 1 || len(got.Skipped) != 1 || !strings.HasSuffix(got.Skipped[0].Reason, "brain says no") {
		t.Fatalf("%+v %q", got, msg)
	}
	ids := b.auditLines(t)[0]["ids"].(map[string]any)
	if fmt.Sprint(ids["brains"]) != "[team/acme]" {
		t.Fatalf("audit: %+v", ids)
	}
}

// A deadline below 0 or above a minute is refused, and asks no brain.
func TestRecallRefusesADeadlineOutOfBounds(t *testing.T) {
	for _, ms := range []int{-1, 60001} {
		b := slowBench(t, 0)
		_, msg, _ := recallWith(t, b, map[string]any{"query": "ledger", "deadline_ms": ms})
		if !strings.Contains(msg, "deadline_ms must be between 0 (none) and 60000") || len(b.brain.calls) != 0 {
			t.Fatalf("%d: %q, %d calls", ms, msg, len(b.brain.calls))
		}
	}
	for _, ms := range []int{0, 60000} {
		if _, err := recallDeadline(ms); err != nil {
			t.Fatalf("%d: %v", ms, err)
		}
	}
}

// lateBrains names each brain asked that did not answer, in the order asked.
func TestLateBrainsNameTheBrainsThatDidNotAnswer(t *testing.T) {
	asked := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	done := map[string]bool{"b": true, "d": true, "f": true, "h": true}
	got := lateBrains(asked, done, 1500*time.Millisecond)
	if fmt.Sprint(got) != "[{a no answer within 1500 ms} {c no answer within 1500 ms} {e no answer within 1500 ms} {g no answer within 1500 ms}]" {
		t.Fatalf("%v", got)
	}
	if got := lateBrains([]string{"a"}, map[string]bool{"a": true}, time.Second); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}

// The schema of th_recall offers deadline_ms with its bound.
func TestRecallSchemaOffersTheDeadline(t *testing.T) {
	b, secret := pyramidBench(t)
	tools, err := b.federated(t, secret).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		if tool.Name == "th_recall" {
			raw := fmt.Sprint(tool.InputSchema)
			if !strings.Contains(raw, "deadline_ms") || !strings.Contains(raw, "at most 60000") {
				t.Fatalf("schema: %s", raw)
			}
			return
		}
	}
	t.Fatal("no th_recall")
}
