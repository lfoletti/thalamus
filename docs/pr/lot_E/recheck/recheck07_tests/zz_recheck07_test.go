package cortexdb

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func recheckOpen(t *testing.T) *DB {
	t.Helper()
	db, err := Open(DefaultConfig(filepath.Join(t.TempDir(), "r07.db")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if db.HasEmbedder() {
		t.Fatal("want no embedder")
	}
	return db
}

func recheckSave(t *testing.T, db *DB, id, content string) {
	t.Helper()
	if _, err := db.SaveKnowledge(context.Background(), KnowledgeSaveRequest{KnowledgeID: id, Title: id, Content: content}); err != nil {
		t.Fatal(err)
	}
}

func recheckIDs(t *testing.T, db *DB, req KnowledgeSearchRequest) []string {
	t.Helper()
	req.RetrievalMode = RetrievalModeLexical
	resp, err := db.SearchKnowledge(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, h := range resp.Results {
		ids = append(ids, h.KnowledgeID)
	}
	return ids
}

func recheckText(t *testing.T, db *DB, req ToolSearchTextRequest) []string {
	t.Helper()
	resp, err := db.GraphRAGTools().SearchText(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, c := range resp.Chunks {
		ids = append(ids, c.DocumentID)
	}
	return ids
}

func has(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// E1: the early exit. A question whose words all occur in >= top_k rows
// fills top_k on idx 0, so an alternate query that alone finds a row is never run.
func TestRecheck07EarlyExit(t *testing.T) {
	db := recheckOpen(t)
	for i := 0; i < 6; i++ {
		recheckSave(t, db, fmt.Sprintf("fr%d", i), fmt.Sprintf("Le conseil fédéral règle la procédure numéro %d des cantons.", i))
	}
	recheckSave(t, db, "de-target", "Der Bundesrat regelt das Verfahren der Kantone im Einzelnen.")
	q := "conseil procédure cantons"
	alt := []string{"Bundesrat Verfahren Kantone"}
	t.Logf("queries = %q", lexicalSearchQueries(q, nil, alt))
	t.Logf("SearchText topK=3 alone   : %v", recheckText(t, db, ToolSearchTextRequest{Query: q, TopK: 3}))
	t.Logf("SearchText topK=3 +alt    : %v", recheckText(t, db, ToolSearchTextRequest{Query: q, TopK: 3, AlternateQueries: alt}))
	t.Logf("SearchText topK=10 +alt   : %v", recheckText(t, db, ToolSearchTextRequest{Query: q, TopK: 10, AlternateQueries: alt}))
	t.Logf("SearchKnowledge topK=3 +alt : %v", recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 3, AlternateQueries: alt}))
	t.Logf("SearchKnowledge topK=10 +alt: %v", recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 10, AlternateQueries: alt}))
	t.Logf("SearchKnowledge topK=3 +kw  : %v", recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 3, Keywords: []string{"Bundesrat", "Verfahren"}}))
}

// E2: a natural question (AND of all tokens) that matches nothing: the
// alternate runs, but also as an AND of all its tokens.
func TestRecheck07AlternateIsAnd(t *testing.T) {
	db := recheckOpen(t)
	recheckSave(t, db, "de-target", "Der Bundesrat regelt das Verfahren der Kantone im Einzelnen.")
	recheckSave(t, db, "fr-other", "Le peuple suisse élit le Conseil national tous les quatre ans.")
	q := "Qui règle la procédure des cantons ?"
	for _, alt := range [][]string{
		{"Bundesrat Verfahren Kantone"},
		{"Wer regelt das Verfahren der Kantone?"},
		{"Wie regelt der Bund das Verfahren der Kantone?"},
	} {
		t.Logf("alt %q -> SearchText %v ; SearchKnowledge %v", alt[0],
			recheckText(t, db, ToolSearchTextRequest{Query: q, TopK: 5, AlternateQueries: alt}),
			recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 5, AlternateQueries: alt}))
	}
	t.Logf("kw [Bund Verfahren Kantone] -> SearchKnowledge %v",
		recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 5, Keywords: []string{"Bund", "Verfahren", "Kantone"}}))
	t.Logf("queries = %q", lexicalSearchQueries(q, nil, []string{"Wie regelt der Bund das Verfahren der Kantone?"}))
}

// E3: accents on SQLite: FTS5 unicode61 folds, the rarity floor does not.
func TestRecheck07Accents(t *testing.T) {
	db := recheckOpen(t)
	recheckSave(t, db, "etat", "Le Conseil d'État du canton exerce le pouvoir exécutif et dirige l'administration.")
	recheckSave(t, db, "uni", "Die Universität Zürich untersteht dem kantonalen Recht und der Aufsicht des Regierungsrates.")
	recheckSave(t, db, "kv", "Die Krankenversicherung ist für alle Personen mit Wohnsitz in der Schweiz obligatorisch.")
	recheckSave(t, db, "filler", "Le parlement cantonal vote le budget annuel et contrôle les comptes de l'administration.")
	for _, q := range []string{"conseil etat pouvoir", "conseil état pouvoir", "universitat zurich recht", "universität zürich recht", "versicherung obligatorisch schweiz", "krankenversicherung obligatorisch"} {
		t.Logf("%-36q SearchText %v ; SearchKnowledge %v", q,
			recheckText(t, db, ToolSearchTextRequest{Query: q, TopK: 5}),
			recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 5}))
	}
	// floor off
	db2, err := Open(DefaultConfig(filepath.Join(t.TempDir(), "r07b.db")), WithKnowledgeFloors(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	recheckSave(t, db2, "etat", "Le Conseil d'État du canton exerce le pouvoir exécutif et dirige l'administration.")
	recheckSave(t, db2, "uni", "Die Universität Zürich untersteht dem kantonalen Recht und der Aufsicht des Regierungsrates.")
	for _, q := range []string{"conseil etat pouvoir", "universitat zurich recht"} {
		t.Logf("floors off %-26q SearchKnowledge %v", q, recheckIDs(t, db2, KnowledgeSearchRequest{Query: q, TopK: 5}))
	}
	t.Logf("containsWord(état,etat)=%v containsWord(krankenversicherung,versicherung)=%v",
		containsWord("conseil d'état", "etat"), containsWord("die krankenversicherung", "versicherung"))
	t.Logf("lexicalQueryKeywords fr = %q", lexicalQueryKeywords("Quel est le rôle du Conseil d'État dans le canton ?"))
	t.Logf("lexicalQueryKeywords de = %q", lexicalQueryKeywords("Wer ist für die Aufsicht über die Universität zuständig?"))
}

func TestRecheck07RawIndexes(t *testing.T) {
	db := recheckOpen(t)
	recheckSave(t, db, "uni", "Die Universität Zürich untersteht dem kantonalen Recht.")
	recheckSave(t, db, "kv", "Die Krankenversicherung ist obligatorisch.")
	recheckSave(t, db, "etat", "Le Conseil d'État du canton.")
	ctx := context.Background()
	for _, idx := range []string{"chunks_fts", "chunks_fts_cjk"} {
		for _, term := range []string{"universitat", "universität", "versicherung", "etat", "état", "Versicherung"} {
			var n int
			if err := db.queryRow(ctx, "SELECT count(*) FROM "+idx+" WHERE "+idx+" MATCH ?", `"`+term+`"`).Scan(&n); err != nil {
				t.Logf("%s %s err %v", idx, term, err)
				continue
			}
			t.Logf("%-15s %-14s %d", idx, term, n)
		}
	}
	for _, q := range []string{"versicherung", "Versicherung", "universitat"} {
		t.Logf("SearchText %q -> %v ; SearchKnowledge %v", q, recheckText(t, db, ToolSearchTextRequest{Query: q, TopK: 5}), recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 5}))
	}
}

func TestRecheck07FloorOnVariants(t *testing.T) {
	ctx := context.Background()
	for _, floors := range []bool{true, false} {
		var opts []Option
		if !floors {
			opts = append(opts, WithKnowledgeFloors(0, 0))
		}
		db, err := Open(DefaultConfig(filepath.Join(t.TempDir(), fmt.Sprintf("f%v.db", floors))), opts...)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		recheckSave(t, db, "de-target", "Der Bundesrat regelt das Verfahren der Kantone im Einzelnen.")
		recheckSave(t, db, "fr-other", "Le peuple suisse élit le Conseil national tous les quatre ans.")
		for i := 0; i < 20; i++ {
			recheckSave(t, db, fmt.Sprintf("fr-f%d", i), fmt.Sprintf("La loi %d règle des questions de la vie des communes et qui les paie.", i))
		}
		q := "Qui règle la procédure des cantons ?"
		t.Logf("floors=%v alt short -> %v", floors, recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 5, AlternateQueries: []string{"Bundesrat Verfahren Kantone"}}))
		t.Logf("floors=%v kw        -> %v", floors, recheckIDs(t, db, KnowledgeSearchRequest{Query: q, TopK: 5, Keywords: []string{"Bundesrat", "Verfahren", "Kantone"}}))
		sc, err := db.chunkRarityScorer(ctx, "", newRarityQuery(q, []string{"Bundesrat", "Verfahren", "Kantone"}))
		if err != nil {
			t.Fatal(err)
		}
		m := sc.coverage("Der Bundesrat regelt das Verfahren der Kantone im Einzelnen.")
		t.Logf("floors=%v coverage of de-target = %.3f words=%d", floors, m.share, m.words)
	}
}
