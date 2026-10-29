package relay

import "time"

// A recall with a deadline (the hook's) answers with the brains that are
// ready and skips the others; but its connection has already reached the
// brain, and a server that starts its brains on demand starts this one,
// making room by stopping others. When the hook asks more brains than the
// server holds at once, every message wakes some and stops others, and none
// more answers. So a brain that missed a deadline is left alone by the
// recalls with one, coldFor long or until a call reaches it and it answers:
// a recall without a deadline, any th_* tool, a call through /mcp/<brain>.
// A brain that answered one call of the recall is not cold, and a brain the
// caller names is asked all the same.

// coldFor is how long a recall with a deadline leaves a cold brain alone.
// After it the brain is asked once more, and found again if it came back.
const coldFor = 10 * time.Minute

// coldReason is the skipped line of a brain left alone.
const coldReason = "cold: missed the last deadline, not woken by a recall with one"

// isCold says whether a recall with a deadline leaves the brain alone.
func (s *Server) isCold(name string) bool {
	s.coldMu.Lock()
	defer s.coldMu.Unlock()
	since, marked := s.cold[name]
	return marked && s.now().Sub(since) < coldFor
}

// markCold notes that the brain missed a deadline, from now.
func (s *Server) markCold(name string) {
	s.coldMu.Lock()
	defer s.coldMu.Unlock()
	s.cold[name] = s.now()
}

// warm takes the mark away: a call reached the brain and it answered.
func (s *Server) warm(name string) {
	s.coldMu.Lock()
	defer s.coldMu.Unlock()
	delete(s.cold, name)
}
