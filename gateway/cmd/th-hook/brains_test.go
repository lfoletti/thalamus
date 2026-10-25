package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// brains --json gives the panel each brain: its access, on or off, and
// whether it answers, with the reason when it does not. A
// brain without a key, or a local brain without its file, does not answer.
// Without the gateway, the shared brains of the file are listed, none
// answers, and gateway gives the reason.
func TestTheBrainListInJSON(t *testing.T) {
	g := &fakeGateway{secret: "s3", listing: []map[string]any{{"brain": "org/ref", "access": "read", "key": false}, {"brain": "team/acme", "access": "write", "key": true}}}
	url := g.start(t)
	dir := t.TempDir()
	projet := brainFile(t, dir, "projet.db", "x")
	file := filepath.Join(dir, "recall")
	if err := os.WriteFile(file, []byte("skip=org/ref,local:absent\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"TH_GATEWAY": url, "TH_KEY": "s3", "TH_RECALL_FILE": file, "TH_LOCAL_RECALL": "x",
		"TH_LOCAL_BRAINS": "projet=" + projet + ",absent=" + filepath.Join(dir, "absent.db")}
	code, out, errOut := hook(t, env, "", "brains", "--json")
	want := `{"gateway":"","brains":[` +
		`{"brain":"org/ref","access":"read","on":false,"ok":false,"reason":"no key"},` +
		`{"brain":"team/acme","access":"write","on":true,"ok":true,"reason":""},` +
		`{"brain":"local:projet","access":"local","on":true,"ok":true,"reason":""},` +
		`{"brain":"local:absent","access":"local","on":false,"ok":false,"reason":"no file"}]}` + "\n"
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("%d %s\n%s", code, out, errOut)
	}
	if err := os.WriteFile(file, []byte("only=team/acme,org/ref\nskip=org/ref,local:absent\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env["TH_GATEWAY"] = "http://127.0.0.1:1"
	env["TH_LOCAL_BRAINS"] = "projet=" + projet
	code, out, _ = hook(t, env, "", "brains", "--json")
	var v struct {
		Gateway string
		Brains  []struct {
			Brain  string
			OK     bool
			Reason string
		}
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil || code != 1 || !strings.Contains(v.Gateway, "th_brains") {
		t.Fatalf("down: %d %q %v", code, out, err)
	}
	var names []string
	for _, b := range v.Brains {
		names = append(names, b.Brain)
		if b.Brain != "local:projet" && (b.OK || b.Reason != "gateway does not answer") {
			t.Fatalf("down: %+v", b)
		}
	}
	if !reflect.DeepEqual(names, []string{"team/acme", "org/ref", "local:projet"}) {
		t.Fatalf("down: %v", names)
	}
	bad := &fakeGateway{secret: "s3", badJSON: true}
	env["TH_GATEWAY"] = bad.start(t)
	env["TH_LOCAL_BRAINS"] = "noequals"
	code, out, errOut = hook(t, env, "", "brains", "--json")
	if code != 1 || !strings.Contains(out, `"gateway":"th_brains: invalid character`) || !strings.Contains(errOut, "TH_LOCAL_BRAINS") {
		t.Fatalf("bad answer and bad local brains: %d %q %s", code, out, errOut)
	}
}
