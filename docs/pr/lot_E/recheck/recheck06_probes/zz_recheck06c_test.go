package cortexdb

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecheck06AMemoryNodeUnderACollectionScope(t *testing.T) {
	db, err := Open(DefaultConfig(filepath.Join(t.TempDir(), "recheck06c.db")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	bg := context.Background()
	if _, err := db.SaveKnowledge(bg, KnowledgeSaveRequest{KnowledgeID: "a1", Collection: "alpha", Content: "Kestrel Vale runs the alpha mill.",
		Entities: []ToolEntityInput{{Name: "Kestrel Vale", ChunkIDs: []string{graphChunkNodeID("a1", 0)}}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SaveMemory(bg, MemorySaveRequest{MemoryID: "m1", UserID: "zeus", Content: "ZEUSPRIVATE: Kestrel Vale owes me money.", Entities: []ToolEntityInput{{Name: "Kestrel Vale"}}}); err != nil {
		t.Fatal(err)
	}
	ctx := WithCallerScope(bg, CallerScope{Key: "alpha", Knowledge: func(c string) bool { return c == "alpha" }, Memory: func(MemoryBucket) bool { return false }})
	r, err := db.GraphRAGTools().ExpandGraph(ctx, ToolExpandGraphRequest{NodeIDs: []string{EntityNodeID("Kestrel Vale")}})
	raw, _ := json.Marshal(r)
	t.Logf("expand_graph serves zeus's memory to an alpha key: %v (err %v)", strings.Contains(string(raw), "ZEUSPRIVATE"), err)
}
