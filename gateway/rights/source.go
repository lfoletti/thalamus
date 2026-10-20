package rights

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"
)

// Source serves the current rights of a file, and reads the file again
// when it changes: a right withdrawn in the file is withdrawn at the next
// request, without a restart. A file that cannot be read or parsed leaves
// the previous rights in force and is reported once, through Warn.
type Source struct {
	path string
	mu   sync.Mutex
	r    *Rights
	// What the last read saw: the next read is skipped while the file
	// still shows the same time and size.
	mtime time.Time
	size  int64
	data  []byte
	// The change that failed last, reported once: the file's time and
	// size, or the stat error. Empty after a success.
	failed string
	// Warn receives one error per failed change. Nil discards it.
	Warn func(error)
}

// Open loads the file once; the first load must succeed.
func Open(path string) (*Source, error) {
	s := &Source{path: path}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("read rights file: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rights file: %w", err)
	}
	r, err := Parse(data)
	if err != nil {
		return nil, err
	}
	s.r, s.mtime, s.size, s.data = r, info.ModTime(), info.Size(), data
	return s, nil
}

// Static serves one value and never reads a file.
func Static(r *Rights) *Source { return &Source{r: r} }

// Current returns the rights as of now.
func (s *Source) Current() *Rights {
	if s.path == "" {
		return s.r
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	info, err := os.Stat(s.path)
	if err == nil && info.ModTime().Equal(s.mtime) && info.Size() == s.size {
		return s.r
	}
	if err == nil {
		var data []byte
		if data, err = os.ReadFile(s.path); err == nil {
			s.mtime, s.size = info.ModTime(), info.Size()
			if bytes.Equal(data, s.data) {
				return s.r
			}
			var r *Rights
			if r, err = Parse(data); err == nil {
				s.r, s.data, s.failed = r, data, ""
				return s.r
			}
		}
	}
	// The change failed: say so once per change, keep the previous rights.
	key := "stat: " + err.Error()
	if info != nil {
		key = fmt.Sprintf("%s %d", info.ModTime(), info.Size())
	}
	if key != s.failed {
		s.failed = key
		if s.Warn != nil {
			s.Warn(fmt.Errorf("rights file %s: %v; the previous rights stay in force", s.path, err))
		}
	}
	return s.r
}
