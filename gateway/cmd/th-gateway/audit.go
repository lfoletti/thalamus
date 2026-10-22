package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lfoletti/thalamus/gateway/relay"
)

// auditFilter keeps the lines of the summary: from a day on, of one person,
// of one brain. Empty fields keep everything.
type auditFilter struct{ since, user, brain string }

// readKey groups reads by day, brain, and whether they are counted or named.
type readKey struct{ day, brain, kind string }

type tally struct{ calls, failed int }

type writeKey struct{ user, tool, brain string }

type logins struct {
	n    int
	last string
}

// auditSummary is what the audit file says, grouped for a person to read.
type auditSummary struct {
	reads      map[readKey]*tally
	writes     map[writeKey]int
	failures   []relay.Entry
	logins     map[string]*logins
	unreadable int
}

// summarize reads the audit lines under the filter. It is pure: the command
// reads the file and prints what this returns.
func summarize(data []byte, f auditFilter) auditSummary {
	s := auditSummary{reads: map[readKey]*tally{}, writes: map[writeKey]int{}, logins: map[string]*logins{}}
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e relay.Entry
		if json.Unmarshal(line, &e) != nil {
			s.unreadable++
			continue
		}
		day := e.Day
		if day == "" && len(e.Time) >= 10 {
			day = e.Time[:10]
		}
		if day < f.since || (f.user != "" && e.User != f.user) {
			continue
		}
		brains := brainsOf(e)
		if f.brain != "" {
			if !slices.Contains(brains, f.brain) {
				continue
			}
			brains = []string{f.brain}
		}
		switch {
		case e.User == "" && e.Tool == "reads":
			s.add(readKey{day, e.Brain, "counted"}, e.Calls, e.Failed)
		case relay.Reads(e.Tool):
			for _, b := range brains {
				s.add(readKey{day, b, "named"}, 1, boolInt(!e.OK))
			}
		case e.Tool == "login":
			if e.OK {
				l := s.logins[e.User]
				if l == nil {
					l = &logins{}
					s.logins[e.User] = l
				}
				l.n++
				l.last = max(l.last, e.Time)
			}
		case e.OK:
			s.writes[writeKey{e.User, e.Tool, e.Brain}]++
		}
		if !e.OK {
			s.failures = append(s.failures, e)
		}
	}
	return s
}

// brainsOf is the brain of a line, or the brains a recall or a login names.
func brainsOf(e relay.Entry) []string {
	if e.Brain != "" {
		return []string{e.Brain}
	}
	names, _ := e.IDs["brains"].([]any)
	var out []string
	for _, n := range names {
		if s, ok := n.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (s auditSummary) add(k readKey, calls, failed int) {
	t := s.reads[k]
	if t == nil {
		t = &tally{}
		s.reads[k] = t
	}
	t.calls += calls
	t.failed += failed
}

// errorWidth bounds the error of a failed call in the summary.
const errorWidth = 100

// print writes the summary, each section sorted.
func (s auditSummary) print(out io.Writer) {
	section := func(title string, lines []string) {
		fmt.Fprintln(out, title)
		if len(lines) == 0 {
			lines = []string{"(none)"}
		}
		for _, l := range lines {
			fmt.Fprintln(out, "  "+l)
		}
	}
	var lines []string
	for k, t := range s.reads {
		lines = append(lines, fmt.Sprintf("%s  %s  %d calls, %d failed, %s", k.day, k.brain, t.calls, t.failed, k.kind))
	}
	sort.Strings(lines)
	section("Reads (per day and brain)", lines)
	lines = nil
	for k, n := range s.writes {
		lines = append(lines, fmt.Sprintf("%s  %s  %s  %d", k.user, k.tool, k.brain, n))
	}
	sort.Strings(lines)
	section("Writes (per person, tool and brain)", lines)
	lines = nil
	for _, e := range s.failures {
		msg := e.Error
		if utf8.RuneCountInString(msg) > errorWidth {
			msg = string([]rune(msg)[:errorWidth]) + "…"
		}
		lines = append(lines, fmt.Sprintf("%s  %s  %s  %s  %s", e.Time, e.User, strings.Join(brainsOf(e), ","), e.Tool, msg))
	}
	sort.Strings(lines)
	section("Refused or failed calls", lines)
	lines = nil
	for u, l := range s.logins {
		lines = append(lines, fmt.Sprintf("%s  %d, last %s", u, l.n, l.last))
	}
	sort.Strings(lines)
	section("Logins", lines)
	if s.unreadable > 0 {
		fmt.Fprintf(out, "%d line(s) could not be read.\n", s.unreadable)
	}
	fmt.Fprintln(out, "The reads of the current day are written at the first call of the next day, or when the gateway stops.")
}

// auditCommand is th-gateway audit <audit.jsonl> [--since day] [--user
// login] [--brain brain].
func auditCommand(args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: audit needs the audit file", errUsage)
	}
	fs := flag.NewFlagSet("audit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var f auditFilter
	fs.StringVar(&f.since, "since", "", "first day, YYYY-MM-DD")
	fs.StringVar(&f.user, "user", "", "one person")
	fs.StringVar(&f.brain, "brain", "", "one brain")
	if err := fs.Parse(args[1:]); err != nil || fs.NArg() > 0 {
		return fmt.Errorf("%w: audit <audit.jsonl> [--since YYYY-MM-DD] [--user login] [--brain brain]", errUsage)
	}
	if f.since != "" {
		if _, err := time.Parse(time.DateOnly, f.since); err != nil {
			return fmt.Errorf("%w: --since %q is not YYYY-MM-DD", errUsage, f.since)
		}
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	summarize(data, f).print(out)
	return nil
}
