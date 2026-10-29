package relay

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// counting puts a TCP proxy in front of team/other that counts the
// connections it accepts: on the VPS, a connection is what wakes a brain.
func counting(t *testing.T, b *bench) *atomic.Int32 {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lis.Close() })
	other := b.rights.Brains["team/other"]
	target := other.Addr
	other.Addr = lis.Addr().String()
	b.rights.Brains["team/other"] = other
	n := &atomic.Int32{}
	go func() {
		for {
			c, err := lis.Accept()
			if err != nil {
				return
			}
			n.Add(1)
			go func() {
				defer func() { _ = c.Close() }()
				up, err := net.Dial("tcp", target)
				if err != nil {
					return
				}
				defer func() { _ = up.Close() }()
				go func() { _, _ = io.Copy(up, c) }()
				_, _ = io.Copy(c, up)
			}()
		}
	}()
	return n
}

// coldServer is the bench's server with team/other slow and counted, and a
// clock the test moves.
func coldServer(t *testing.T) (*bench, *Server, *atomic.Int32, *time.Time) {
	t.Helper()
	b := slowBench(t, 10*time.Second)
	n := counting(t, b)
	s := direct(t, b)
	now := time.Date(2026, 10, 29, 22, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	return b, s, n, &now
}

func recallAt(t *testing.T, s *Server, deadlineMS int) recallResult {
	t.Helper()
	res, err := s.recall(context.Background(), "lfoletti", recallArgs{Query: "ledger", DeadlineMS: deadlineMS})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func reasonOf(res recallResult, brain string) string {
	for _, k := range res.Skipped {
		if k.Brain == brain {
			return k.Reason
		}
	}
	return ""
}

// A recall with a deadline does not reach a brain that missed the last
// one: it is skipped as cold, and no connection wakes it.
func TestARecallWithADeadlineLeavesAColdBrainAlone(t *testing.T) {
	b, s, n, _ := coldServer(t)
	if r := reasonOf(recallAt(t, s, 200), "team/other"); r != "no answer within 200 ms" {
		t.Fatalf("first recall: %q", r)
	}
	waitCut(t, b)
	before := n.Load()
	res := recallAt(t, s, 200)
	if r := reasonOf(res, "team/other"); r != coldReason {
		t.Fatalf("second recall: %q", r)
	}
	if n.Load() != before {
		t.Fatalf("%d connection(s) reached the cold brain", n.Load()-before)
	}
	if len(res.Results) != 1 || res.Results[0].KnowledgeID != "a-1" {
		t.Fatalf("the warm brain still answers: %+v", res.Results)
	}
}

// A recall without a deadline asks a cold brain, and its answer takes the
// mark away: the next recall with a deadline asks it again.
func TestARecallWithoutADeadlineWakesAColdBrain(t *testing.T) {
	b, s, n, _ := coldServer(t)
	recallAt(t, s, 200)
	waitCut(t, b)
	b.brain.slow["other"] = 0
	if res := recallAt(t, s, 0); len(res.Skipped) != 0 || len(res.Results) != 2 {
		t.Fatalf("without a deadline: %+v", res)
	}
	before := n.Load()
	if res := recallAt(t, s, 2000); len(res.Skipped) != 0 || len(res.Results) != 2 || n.Load() == before {
		t.Fatalf("after its answer: %+v, %d connection(s)", res, n.Load()-before)
	}
}

// Any call that reaches the brain takes the mark away, not only a recall.
func TestAnyCallThatReachesTheBrainWakesIt(t *testing.T) {
	b, s, _, _ := coldServer(t)
	recallAt(t, s, 200)
	waitCut(t, b)
	b.brain.slow["other"] = 0
	if _, err := s.askBrain(context.Background(), b.rights.Brains["team/other"], "lfoletti", "team/other", "ledger", 5, "other"); err != nil {
		t.Fatal(err)
	}
	if r := reasonOf(recallAt(t, s, 2000), "team/other"); r != "" {
		t.Fatalf("still skipped: %q", r)
	}
}

// A call that fails leaves the mark: the brain did not answer.
func TestAFailedCallLeavesTheMark(t *testing.T) {
	b, s, _, _ := coldServer(t)
	recallAt(t, s, 200)
	waitCut(t, b)
	b.brain.slow["other"] = 0
	b.brain.failOn = "other"
	if _, err := s.askBrain(context.Background(), b.rights.Brains["team/other"], "lfoletti", "team/other", "ledger", 5, "other"); err == nil {
		t.Fatal("the call should fail")
	}
	if r := reasonOf(recallAt(t, s, 200), "team/other"); r != coldReason {
		t.Fatalf("after a failed call: %q", r)
	}
}

// The mark lasts coldFor: after it, a recall with a deadline asks the brain
// once more, so a brain that came back is found again.
func TestTheColdMarkLastsColdFor(t *testing.T) {
	b, s, n, now := coldServer(t)
	recallAt(t, s, 200)
	waitCut(t, b)
	*now = now.Add(coldFor - time.Second)
	if r := reasonOf(recallAt(t, s, 200), "team/other"); r != coldReason {
		t.Fatalf("within coldFor: %q", r)
	}
	b.brain.slow["other"] = 0
	*now = now.Add(time.Second)
	before := n.Load()
	if r := reasonOf(recallAt(t, s, 2000), "team/other"); r != "" || n.Load() == before {
		t.Fatalf("after coldFor: %q, %d connection(s)", r, n.Load()-before)
	}
}

// A deadline missed again restarts the mark from then.
func TestAMissedDeadlineRestartsTheMark(t *testing.T) {
	b, s, _, now := coldServer(t)
	recallAt(t, s, 200)
	waitCut(t, b)
	*now = now.Add(coldFor)
	b.brain.cut.Store(false)
	recallAt(t, s, 200) // asked again, misses again
	waitCut(t, b)
	*now = now.Add(coldFor - time.Second)
	if r := reasonOf(recallAt(t, s, 200), "team/other"); r != coldReason {
		t.Fatalf("the mark did not restart: %q", r)
	}
}
