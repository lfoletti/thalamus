package relay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lfoletti/thalamus/gateway/rights"
)

// Entry is one audit line: who called what, on which brain, naming which
// rows, and whether the brain answered. A line of counted reads names no
// one: it gives the brain, the day, the calls and the failed calls.
type Entry struct {
	Time   string         `json:"ts"`
	User   string         `json:"user"`
	Key    string         `json:"key"`
	Brain  string         `json:"brain"`
	Tool   string         `json:"tool"`
	Day    string         `json:"day,omitempty"`
	Calls  int            `json:"calls,omitempty"`
	Failed int            `json:"failed,omitempty"`
	IDs    map[string]any `json:"ids,omitempty"`
	OK     bool           `json:"ok"`
	Error  string         `json:"error,omitempty"`
	// Denied marks an error that is the refusal of a key; it is not written.
	Denied bool `json:"-"`
}

// denied reports whether err is a refusal of the caller's key, wrapped or
// not.
func denied(err error) bool {
	c := status.Code(err)
	return c == codes.PermissionDenied || c == codes.Unauthenticated
}

// count is one read the audit counts for a brain, without a name.
type count struct {
	Brain  string
	Failed bool
}

// readsOf are the gateway's own tools that read; th_add, th_review and
// login write. An engine tool reads when the engine says so.
var readsOf = map[string]bool{"th_recall": true, "th_entity": true, "th_get": true, "th_descend": true, "th_pending": true, "th_brains": true}

// Reads reports whether a tool reads: a gateway tool that reads, or an
// engine tool that the engine classes as a read. The audit summary of
// th-gateway uses it.
func Reads(tool string) bool { return reads(tool) }

func reads(tool string) bool {
	if readsOf[tool] {
		return true
	}
	access, known := authz.LookupTool(tool)
	return known && access == authz.Read
}

// decide is the audit policy: what one call leaves as named lines and as
// counted reads. A write keeps a named line, and so does a tool the engine
// does not know. A read keeps a named line on a brain with audit_reads, or
// when the brain refused the key, and is counted elsewhere. th_recall and
// th_entity follow the rule brain by brain; th_brains reads no brain and
// leaves nothing.
func decide(e Entry, r *rights.Rights) ([]Entry, []count) {
	switch {
	case !reads(e.Tool):
		return []Entry{e}, nil
	case e.Tool == "th_brains":
		return nil, nil
	case e.Tool == "th_recall", e.Tool == "th_entity":
		return splitRecall(e, r)
	case r.Brains[e.Brain].AuditReads, e.Denied:
		return []Entry{e}, nil
	}
	return nil, []count{{Brain: e.Brain, Failed: !e.OK}}
}

// splitRecall keeps, in one named line, the brains of a federated read that
// have audit_reads and the ids read there, with the other ids of the call,
// and counts the other brains.
func splitRecall(e Entry, r *rights.Rights) ([]Entry, []count) {
	refusals := refusedLines(e)
	brains, _ := e.IDs["brains"].([]string)
	read, _ := e.IDs["knowledge_ids"].([]string)
	var named, ids []string
	var counted []count
	for _, b := range brains {
		if r.Brains[b].AuditReads {
			named = append(named, b)
		} else {
			counted = append(counted, count{Brain: b})
		}
	}
	if len(named) == 0 {
		return refusals, counted
	}
	for _, id := range read {
		for _, b := range named {
			if strings.HasPrefix(id, b+":") {
				ids = append(ids, id)
			}
		}
	}
	out := e
	out.IDs = map[string]any{}
	for k, v := range e.IDs {
		out.IDs[k] = v
	}
	delete(out.IDs, "denied")
	out.IDs["brains"], out.IDs["knowledge_ids"] = named, ids
	return append([]Entry{out}, refusals...), counted
}

// refusedLines is one named line per brain of a federated read that
// refused the caller's key, as for any refused read.
func refusedLines(e Entry) []Entry {
	denied, _ := e.IDs["denied"].(map[string]string)
	var out []Entry
	for _, b := range sortedKeys(denied) {
		out = append(out, Entry{User: e.User, Key: e.Key, Brain: b, Tool: e.Tool, Error: denied[b], Denied: true})
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// keep drops the lines dated before cutoff. A line whose time cannot be
// read stays: the audit never drops what it cannot date.
func keep(data []byte, cutoff time.Time) []byte {
	var out bytes.Buffer
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		var e struct {
			Time string `json:"ts"`
		}
		if json.Unmarshal(line, &e) == nil {
			if t, err := time.Parse(time.RFC3339Nano, e.Time); err == nil && t.Before(cutoff) {
				continue
			}
		}
		out.Write(line)
	}
	return out.Bytes()
}

// tally is the counted reads of one brain for the current day.
type tally struct{ calls, failed int }

// Audit appends JSON lines to a file, under the policy that Note applies.
type Audit struct {
	mu     sync.Mutex
	path   string
	file   *os.File
	now    func() time.Time
	day    string
	counts map[string]*tally
}

// OpenAudit opens the file for append, creating it if needed.
func OpenAudit(path string) (*Audit, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}
	return &Audit{path: path, file: f, now: time.Now, counts: map[string]*tally{}}, nil
}

var errClosed = errors.New("audit file is closed")

// Note records one call under the policy of r. At the first call of a day
// it writes the counted reads of the day before, then drops the lines older
// than the retention of r.
func (a *Audit) Note(e Entry, r *rights.Rights) error {
	named, counted := decide(e, r)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.file == nil {
		return errClosed
	}
	if day := a.now().UTC().Format(time.DateOnly); day != a.day {
		if err := a.flush(); err != nil {
			return err
		}
		if err := a.drop(r.RetentionDays()); err != nil {
			return err
		}
		a.day = day
	}
	for _, n := range named {
		if err := a.write(n); err != nil {
			return err
		}
	}
	for _, c := range counted {
		t := a.counts[c.Brain]
		if t == nil {
			t = &tally{}
			a.counts[c.Brain] = t
		}
		t.calls++
		if c.Failed {
			t.failed++
		}
	}
	return nil
}

// write appends one line. The time is set here when the entry has none.
func (a *Audit) write(e Entry) error {
	if e.Time == "" {
		e.Time = a.now().UTC().Format(time.RFC3339Nano)
	}
	line, _ := json.Marshal(e)
	_, err := a.file.Write(append(line, '\n'))
	return err
}

// flush writes one line per brain for the counted reads of the day, sorted
// by brain, and starts the counts again.
func (a *Audit) flush() error {
	brains := make([]string, 0, len(a.counts))
	for b := range a.counts {
		brains = append(brains, b)
	}
	sort.Strings(brains)
	for _, b := range brains {
		t := a.counts[b]
		if err := a.write(Entry{Brain: b, Tool: "reads", Day: a.day, Calls: t.calls, Failed: t.failed, OK: true}); err != nil {
			return err
		}
	}
	a.counts = map[string]*tally{}
	return nil
}

// drop rewrites the file without the lines older than days; 0 keeps every
// line. It runs at most once a day.
func (a *Audit) drop(days int) error {
	if days == 0 {
		return nil
	}
	data, err := os.ReadFile(a.path)
	if err != nil {
		return fmt.Errorf("drop old audit lines: %w", err)
	}
	kept := keep(data, a.now().Add(-time.Duration(days)*24*time.Hour))
	if len(kept) == len(data) {
		return nil
	}
	if err := a.replace(kept); err != nil {
		return fmt.Errorf("drop old audit lines of %s: %w", a.path, err)
	}
	return nil
}

// syncFile flushes a file to the disk; the tests make it fail.
var syncFile = (*os.File).Sync

// replace writes data to a temporary file next to the audit, syncs it, and
// renames it over the audit, so that a stop at any point leaves either the
// old file or the new one, whole. The temporary file's descriptor, opened
// for append, becomes the audit's descriptor. On a failure the temporary
// file goes and the old descriptor stays.
func (a *Audit) replace(data []byte) error {
	tmp := a.path + ".tmp"
	_ = os.Remove(tmp) // left by a stop before the rename
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = syncFile(f)
	}
	if err == nil {
		err = os.Rename(tmp, a.path)
	}
	if err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	_ = a.file.Close()
	a.file = f
	return nil
}

// Close writes the counted reads of the day and closes the file; later
// calls are refused.
func (a *Audit) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.file == nil {
		return nil
	}
	err := a.flush()
	if cerr := a.file.Close(); err == nil {
		err = cerr
	}
	a.file = nil
	return err
}
