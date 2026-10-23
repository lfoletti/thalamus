package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"github.com/liliang-cn/cortexdb/v2/pkg/authz"
)

// Scratch probe for fiche 06: the typed REST routes, outside the toolbox
// allowlist, under a key confined to collection alpha.
func TestRecheck06TypedRoutesUnderAConfinedKey(t *testing.T) {
	ks, err := authz.NewKeySet([]authz.Key{
		{ID: "operator", Secret: "op", Clearance: authz.ReadWrite},
		{ID: "alpha", Secret: "alpha-secret", Clearance: authz.ReadOnly, Scope: authz.Scope{Collection: "alpha"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := newPolicyServer(t, Options{Keys: ks})
	wantStatus(t, srv, http.MethodPost, "/v1/tools/knowledge_save", "op",
		`{"knowledge_id":"a1","collection":"alpha","content":"Kestrel Vale runs the alpha mill.","entities":[{"name":"Kestrel Vale","chunk_ids":["chunk:a1:000"]}]}`, http.StatusOK)
	wantStatus(t, srv, http.MethodPost, "/v1/tools/knowledge_save", "op",
		`{"knowledge_id":"b1","collection":"beta","content":"Kestrel Vale secretly funds Project Nightjar with BETASECRET money.","entities":[{"name":"Kestrel Vale","description":"BETADESC","chunk_ids":["chunk:b1:000"]},{"name":"Project Nightjar","chunk_ids":["chunk:b1:000"]}],"relations":[{"from":"Kestrel Vale","to":"Project Nightjar","type":"funds"}]}`, http.StatusOK)

	for _, c := range []struct{ path, body string }{
		{"/v1/graph/expand", `{"collection":"alpha","node_ids":["entity:kestrel_vale"],"max_hops":2}`},
		{"/v1/graph/expand", `{"node_ids":["entity:kestrel_vale"],"max_hops":2}`},
		{"/v1/query", `{"collection":"alpha","query":"Kestrel Vale","limit":10}`},
		{"/v1/query", `{"collection":"alpha","limit":10,"prefetch":[{"kind":"graph","entity_names":["Project Nightjar"],"max_hops":2}]}`},
		{"/v1/tools/expand_graph", `{"collection":"alpha","node_ids":["entity:kestrel_vale"]}`},
	} {
		status, payload := do(t, srv, http.MethodPost, c.path, "alpha-secret", c.body)
		s := string(payload)
		t.Logf("%-24s %d leaksBeta=%v alphaText=%v body=%.160s", c.path, status,
			strings.Contains(s, "BETASECRET") || strings.Contains(s, "BETADESC") || strings.Contains(s, "Nightjar"),
			strings.Contains(s, "alpha mill"), s)
	}
}
