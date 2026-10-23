package cortexdb

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// DEMO (v2.138.0): a caller confined to hermes's memory extracts zeus's session.
func TestDemoExtractConversationReadsAnotherUsersSession(t *testing.T) {
	ctx := context.Background()
	db, err := Open(DefaultConfig(filepath.Join(t.TempDir(), "scope.db")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	zeus, err := db.SaveMemory(ctx, MemorySaveRequest{MemoryID: "zeus-1", UserID: "zeus", Scope: "user", Content: "Zeus hid the Thunderbolt under Olympus. Hera must not know."})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.SaveMemory(ctx, MemorySaveRequest{MemoryID: "hermes-1", UserID: "hermes", Scope: "user", Content: "the lyre"}); err != nil {
		t.Fatal(err)
	}
	sess := zeus.Memory.SessionID
	hermes := WithCallerScope(ctx, CallerScope{Key: "hermes", Memory: func(b MemoryBucket) bool { return b.UserID == "hermes" }})

	// The typed read refuses: memory_get by list hides zeus-1.
	got, err := db.GetMemory(hermes, MemoryGetRequest{MemoryIDs: []string{"zeus-1"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("GetMemory list under hermes: memories=%d missing=%v", len(got.Memories), got.Missing)
	// memoryReadPermitted on the session says no.
	ok, err := db.memoryReadPermitted(hermes, sess)
	t.Logf("memoryReadPermitted(hermes, %s) = %v, %v", sess, ok, err)

	out, err := db.GraphRAGTools().Call(hermes, "extract_conversation", json.RawMessage(`{"session_id":"`+sess+`"}`))
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	body, _ := json.Marshal(out)
	t.Logf("extract_conversation under hermes on zeus's session: %s", body)
	if strings.Contains(string(body), "Thunderbolt") {
		t.Errorf("DEFECT: zeus's session text served to a caller confined to hermes")
	}
	// What a missing session answers, for the indistinguishability rule.
	_, errMissing := db.GraphRAGTools().Call(hermes, "extract_conversation", json.RawMessage(`{"session_id":"no-such-session"}`))
	t.Logf("missing session answer: %v", errMissing)
}
