package cortexdb

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecheck06FindNodesAndAnExpiredDocument(t *testing.T) {
	db, err := Open(DefaultConfig(filepath.Join(t.TempDir(), "recheck06b.db")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	past := "2020-01-01T00:00:00Z"
	if _, err := db.SaveKnowledge(ctx, KnowledgeSaveRequest{KnowledgeID: "x1", Content: "EXPIREDTEXT about Zephyr Gate.", ExpiresAt: &past}); err != nil {
		t.Fatal(err)
	}
	resp, err := db.GraphRAGTools().FindNodes(ctx, ToolFindNodesRequest{Names: []string{"Zephyr"}})
	raw, _ := json.Marshal(resp)
	t.Logf("find_nodes on expired doc: err=%v servesExpiredText=%v", err, strings.Contains(string(raw), "EXPIREDTEXT"))
	g, _ := db.GraphRAGTools().GetNodes(ctx, ToolGetNodesRequest{NodeIDs: []string{graphChunkNodeID("x1", 0)}})
	raw, _ = json.Marshal(g)
	t.Logf("get_nodes on expired doc: servesExpiredText=%v", strings.Contains(string(raw), "EXPIREDTEXT"))
}
