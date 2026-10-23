package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// defaultTimeout bounds one recall, the gateway and the local brains
// together: a gateway answers in about 150 ms, and a prompt must not wait
// on one that does not answer.
const defaultTimeout = 1500 * time.Millisecond

// downFor is how long the hook remembers a gateway that did not answer.
const downFor = 5 * time.Minute

// now is the clock of the memory of a failed gateway; the tests set it.
var now = time.Now

// recallTimeout is TH_RECALL_TIMEOUT, a Go duration, or the default. A
// value that is not a positive duration is an error, and the default.
func recallTimeout(get func(string) string) (time.Duration, error) {
	v := get("TH_RECALL_TIMEOUT")
	if v == "" {
		return defaultTimeout, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return defaultTimeout, fmt.Errorf("TH_RECALL_TIMEOUT %q is not a positive duration such as 1.5s; the hook waits %v", v, defaultTimeout)
	}
	return d, nil
}

// downFile is gateway-down next to the env file: in the folder of TH_ENV,
// or in ~/.thalamus.
func downFile(get func(string) string) string {
	if p := get("TH_ENV"); p != "" {
		return filepath.Join(filepath.Dir(p), "gateway-down")
	}
	return filepath.Join(get("HOME"), ".thalamus", "gateway-down")
}

// downSince gives the time of the last failure of the gateway when it is
// less than downFor ago.
func downSince(path string) (time.Time, bool) {
	data, _ := os.ReadFile(path)
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
	if err != nil {
		return time.Time{}, false
	}
	return t, now().Sub(t) < downFor
}

// unreachable reports whether the gateway did not answer: the network
// failed or the time ran out. A gateway that refuses the key answered.
func unreachable(err error) bool {
	var ne net.Error
	return errors.As(err, &ne)
}

// markDown remembers that the gateway did not answer now. The folder is
// not created: without an env file there is no memory either.
func markDown(path string) {
	_ = os.WriteFile(path, []byte(now().UTC().Format(time.RFC3339)+"\n"), 0o600)
}

// skippedNote is the line that ends the block while the shared recall is
// skipped.
func skippedNote(since time.Time) string {
	const clock = "15:04:05 UTC"
	return fmt.Sprintf("Shared brains skipped: the gateway did not answer at %s; the hook asks it again from %s.", since.UTC().Format(clock), since.Add(downFor).UTC().Format(clock))
}
