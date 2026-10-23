package authz

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"
)

// protoSource: scratch prototype of the planned Source, for the benchmark only.
type protoSource struct {
	path   string
	mu     sync.Mutex
	ks     *KeySet
	mtime  time.Time
	size   int64
	data   []byte
	failed string
	Warn   func(error)
}

func protoOpen(path string) (*protoSource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ks, err := Parse(data)
	if err != nil {
		return nil, err
	}
	return &protoSource{path: path, ks: ks, mtime: info.ModTime(), size: info.Size(), data: data}, nil
}

func (s *protoSource) Current() *KeySet {
	s.mu.Lock()
	defer s.mu.Unlock()
	info, err := os.Stat(s.path)
	if err == nil && info.ModTime().Equal(s.mtime) && info.Size() == s.size {
		return s.ks
	}
	if err == nil {
		var data []byte
		if data, err = os.ReadFile(s.path); err == nil {
			s.mtime, s.size = info.ModTime(), info.Size()
			if bytes.Equal(data, s.data) {
				return s.ks
			}
			var ks *KeySet
			if ks, err = Parse(data); err == nil {
				s.ks, s.data, s.failed = ks, data, ""
				return s.ks
			}
		}
	}
	key := "stat: " + err.Error()
	if info != nil {
		key = fmt.Sprintf("%s %d", info.ModTime(), info.Size())
	}
	if key != s.failed {
		s.failed = key
		if s.Warn != nil {
			s.Warn(err)
		}
	}
	return s.ks
}

// currentStatOutsideLock: stat before taking the lock, lock only to compare.
func (s *protoSource) currentStatOutsideLock() *KeySet {
	info, err := os.Stat(s.path)
	s.mu.Lock()
	if err == nil && info.ModTime().Equal(s.mtime) && info.Size() == s.size {
		ks := s.ks
		s.mu.Unlock()
		return ks
	}
	s.mu.Unlock()
	return s.Current()
}
