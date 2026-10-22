package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// auditFile writes an audit with every kind of line: counted reads, named
// reads (one through th_recall), writes, a refusal, a failed write, logins,
// a line that is not JSON, and lines of another day.
func auditFile(t *testing.T) string {
	t.Helper()
	lines := []string{
		`{"ts":"2026-10-21T23:59:00Z","user":"","key":"","brain":"team/acme","tool":"reads","day":"2026-10-21","calls":5,"failed":1,"ok":true}`,
		`{"ts":"2026-10-21T10:00:00Z","user":"mdupont","key":"th/mdupont/team/acme","brain":"team/acme","tool":"ingest_document","ok":true}`,
		`{"ts":"2026-10-22T09:30:00Z","user":"lfoletti","key":"","brain":"","tool":"login","ok":true}`,
		`{"ts":"2026-10-22T08:00:00Z","user":"lfoletti","key":"","brain":"","tool":"login","ids":{"brains":["team/acme"]},"ok":true}`,
		`{"ts":"2026-10-22T08:10:00Z","user":"lfoletti","key":"","brain":"","tool":"login","ok":false,"error":"token expired"}`,
		`{"ts":"2026-10-22T09:00:00Z","user":"lfoletti","key":"th/lfoletti/org/secrets","brain":"org/secrets","tool":"knowledge_search","ok":true}`,
		`{"ts":"2026-10-22T09:01:00Z","user":"lfoletti","key":"","brain":"","tool":"th_recall","ids":{"brains":["org/secrets","team/acme"],"knowledge_ids":["org/secrets:d-1"]},"ok":true}`,
		`{"ts":"2026-10-22T09:02:00Z","user":"lfoletti","key":"th/lfoletti/team/acme","brain":"team/acme","tool":"knowledge_get","ok":false,"error":"rpc error: code = PermissionDenied desc = ` + strings.Repeat("x", 200) + `"}`,
		`{"ts":"2026-10-22T09:03:00Z","user":"lfoletti","key":"","brain":"team/acme","tool":"th_add","ids":{"document_id":"spec-2"},"ok":true}`,
		`{"ts":"2026-10-22T09:04:00Z","user":"lfoletti","key":"","brain":"team/acme","tool":"th_add","ids":{"document_id":"spec-3"},"ok":true}`,
		`{"ts":"2026-10-22T09:05:00Z","user":"mdupont","key":"","brain":"team/acme","tool":"th_add","ok":false,"error":"brain down"}`,
		`not json`,
	}
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// audit sums the file up: reads per day and brain, counted or named;
// writes per person, tool and brain; each refused or failed call; logins
// per person with the last one. A line it cannot read is counted.
func TestTheAuditSummary(t *testing.T) {
	code, out, errOut := call(t, "audit", auditFile(t))
	want := `Reads (per day and brain)
  2026-10-21  team/acme  5 calls, 1 failed, counted
  2026-10-22  org/secrets  2 calls, 0 failed, named
  2026-10-22  team/acme  2 calls, 1 failed, named
Writes (per person, tool and brain)
  lfoletti  th_add  team/acme  2
  mdupont  ingest_document  team/acme  1
Refused or failed calls
  2026-10-22T08:10:00Z  lfoletti    login  token expired
  2026-10-22T09:02:00Z  lfoletti  team/acme  knowledge_get  rpc error: code = PermissionDenied desc = ` + strings.Repeat("x", 58) + `…
  2026-10-22T09:05:00Z  mdupont  team/acme  th_add  brain down
Logins
  lfoletti  2, last 2026-10-22T09:30:00Z
1 line(s) could not be read.
The reads of the current day are written at the first call of the next day, or when the gateway stops.
`
	if code != 0 || out != want {
		t.Fatalf("code %d, stderr %q, got:\n%s", code, errOut, out)
	}
}

// --since keeps the days from that day on, --user the named lines of one
// person (counted reads have no name), --brain the lines of one brain (a
// recall counts for each brain it names). An empty section says so.
func TestTheAuditSummaryFilters(t *testing.T) {
	path := auditFile(t)
	_, out, _ := call(t, "audit", path, "--since", "2026-10-22", "--user", "mdupont")
	want := `Reads (per day and brain)
  (none)
Writes (per person, tool and brain)
  (none)
Refused or failed calls
  2026-10-22T09:05:00Z  mdupont  team/acme  th_add  brain down
Logins
  (none)
`
	if !strings.HasPrefix(out, want) {
		t.Fatalf("since and user:\n%s", out)
	}
	_, out, _ = call(t, "audit", path, "--brain", "org/secrets")
	if !strings.Contains(out, "  2026-10-22  org/secrets  2 calls, 0 failed, named\nWrites") || strings.Contains(out, "team/acme") || !strings.Contains(out, "Logins\n  (none)") {
		t.Fatalf("brain:\n%s", out)
	}
	clean := filepath.Join(t.TempDir(), "clean.jsonl")
	if err := os.WriteFile(clean, []byte(`{"ts":"2026-10-21T23:59:00Z","user":"","brain":"team/acme","tool":"reads","day":"2026-10-21","calls":5,"ok":true}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, out, _ = call(t, "audit", clean); strings.Contains(out, "could not be read") || !strings.Contains(out, "team/acme  5 calls, 0 failed, counted") {
		t.Fatalf("clean:\n%s", out)
	}
}

// A file that cannot be read is exit 1; a wrong command line is exit 2.
func TestTheAuditCommandLine(t *testing.T) {
	if code, _, errOut := call(t, "audit", filepath.Join(t.TempDir(), "absent.jsonl")); code != 1 || !strings.Contains(errOut, "absent.jsonl") {
		t.Fatalf("absent: %d %s", code, errOut)
	}
	for _, args := range [][]string{{"audit"}, {"audit", "a", "b"}, {"audit", "a", "--since", "22 October"}, {"audit", "a", "--color"}} {
		if code, _, errOut := call(t, args...); code != 2 || !strings.Contains(errOut, "usage:") {
			t.Fatalf("%v: %d %s", args, code, errOut)
		}
	}
}
