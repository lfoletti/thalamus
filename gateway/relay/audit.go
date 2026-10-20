package relay

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// Entry is one audit line: who called what, on which brain, naming which
// rows, and whether the brain answered.
type Entry struct {
	Time  string         `json:"ts"`
	User  string         `json:"user"`
	Key   string         `json:"key"`
	Brain string         `json:"brain"`
	Tool  string         `json:"tool"`
	IDs   map[string]any `json:"ids,omitempty"`
	OK    bool           `json:"ok"`
	Error string         `json:"error,omitempty"`
}

// Audit appends JSON lines to a file, one per tool call.
type Audit struct {
	mu   sync.Mutex
	file *os.File
}

// OpenAudit opens the file for append, creating it if needed.
func OpenAudit(path string) (*Audit, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}
	return &Audit{file: f}, nil
}

// Record writes one line. The time is set here when the entry has none.
func (a *Audit) Record(e Entry) error {
	if e.Time == "" {
		e.Time = time.Now().UTC().Format(time.RFC3339Nano)
	}
	line, _ := json.Marshal(e)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.file == nil {
		return errors.New("audit file is closed")
	}
	_, err := a.file.Write(append(line, '\n'))
	return err
}

// Close closes the file; later records are refused.
func (a *Audit) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.file == nil {
		return nil
	}
	err := a.file.Close()
	a.file = nil
	return err
}
