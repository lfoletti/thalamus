#!/usr/bin/env python3 -I
"""Preuve « failing first » de l'amendement du lot C (20 octobre 2026).

Pour chaque règle corrigée après la revue : une mutation à la main qui
compile, les tests du paquet touché, le ou les tests attendus doivent tomber.
Le fichier est restauré après chaque mutation (depuis la copie en mémoire, donc
un arbre non commité convient). Usage : python3 -I mutate_amend.py 08|05|06 <worktree>.
"""
import subprocess
import sys
from pathlib import Path

which, wt = sys.argv[1], Path(sys.argv[2])
I = "pkg/cortexdb/graphrag_tool_ingest.go"
K = "pkg/cortexdb/knowledge_api.go"
T = "pkg/cortexdb/knowledge_tx.go"
E = "pkg/cortexdb/expiry.go"
P = "pkg/cortexdb/retrieval_plan.go"
H = "pkg/cortexdb/graphrag_tool_helpers.go"
S = "pkg/cortexdb/graphrag_tool_search.go"
R = "pkg/cortexdb/review.go"
C = "pkg/core/store_init.go"
W = "cmd/cortexdb-mcp-stdio/sweep_expired.go"

SETS = {
    "08": ("./pkg/cortexdb/", "Ingest|Dedupe|Hash|Unchanged|Reingest|ChunkIDs|Unreadable", [
        ("Dédup hors collection", I,
         "\t\tif storedIn == collection {\n\t\t\treturn id, nil",
         "\t\tif storedIn == collection || true {\n\t\t\treturn id, nil",
         ["TestDedupeIsDecidedWithinTheCollection"]),
        ("Découpage non comparé", I,
         "\tmetadata[chunkSizeKey] = strconv.Itoa(opts.ChunkSize)",
         "\tmetadata[chunkSizeKey] = \"0\"",
         ["TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking"]),
        ("Collection non comparée", I,
         "\t\tif storedIn == ingestOpts.Collection {",
         "\t\tif storedIn == ingestOpts.Collection || true {",
         ["TestAnUnchangedReingestMeansTheSameCollectionAndTheSameChunking"]),
        ("Mise à jour sans re-hash", K,
         "\t\tstampIngestMarks(metadata, content, resolvedIngestOptions(collection, chunkSize, chunkOverlap))",
         "\t\tstampIngestMarks(metadata, existing.Content, resolvedIngestOptions(collection, chunkSize, chunkOverlap))",
         ["TestTheContentHashFollowsTheContentThroughKnowledgeWrites"]),
        ("knowledge_save sans hash", K,
         "\tstampIngestMarks(metadata, req.Content, resolvedIngestOptions(req.Collection, req.ChunkSize, req.ChunkOverlap))",
         "\tstampIngestMarks(metadata, \"\", resolvedIngestOptions(req.Collection, req.ChunkSize, req.ChunkOverlap))",
         ["TestTheContentHashFollowsTheContentThroughKnowledgeWrites"]),
        ("Hash absent du nœud à la sauvegarde", T,
         "\t\tdocNode.Properties[contentHashKey] = hash\n",
         "\t\t_ = hash\n",
         ["TestTheContentHashFollowsTheContentThroughKnowledgeWrites"]),
        ("Ré-ingestion sans retrait des anciens chunks", I,
         "\tif err := t.db.applyKnowledgeMutation(ctx, req.DocumentID, existing != nil, plan); err != nil {",
         "\tif err := t.db.applyKnowledgeMutation(ctx, req.DocumentID, false, plan); err != nil {",
         ["TestAnIngestWritesEverythingOrNothing"]),
        ("Chunks triés par texte", I,
         "\t\treturn a < b\n",
         "\t\t_, _ = a, b\n\t\treturn chunks[i].ID < chunks[j].ID\n",
         ["TestTheChunkIDsOfAnUnchangedReingestComeInChunkOrder"]),
    ]),
    "05": ("./pkg/cortexdb/ ./cmd/cortexdb-mcp-stdio/", "Expir|Sweep|Chunk|Vector|Indexed|ToolCount", [
        ("Chemin vectoriel non filtré", P,
         "\t\tif db.allowDocument(ctx, filters, result.DocID) {",
         "\t\tif db.allowDocument(ctx, filters, result.DocID) || true {",
         ["TestAnExpiredDocumentIsLeftOutOfTheVectorPathToo"]),
        ("fermeture MCP sans balayage", "pkg/cortexdb/expiry_tool.go",
         "\t\treturn db.SweepExpiredTool(ctx, req)\n",
         "\t\treturn &ExpiredSweepReport{}, nil\n",
         ["TestTheSweepIsAToolOfTheSharedBrain"]),
        ("cortex_query non filtré", "pkg/cortexdb/query_api.go",
         "\t\tresults = db.filterScoredEmbeddings(ctx, results, nil)\n",
         "",
         ["TestAnExpiredDocumentIsLeftOutOfCortexQuery"]),
        ("Outils par chunks non filtrés", H,
         "\t\tif !ok || !t.db.allowDocument(ctx, nil, emb.DocumentID) {",
         "\t\tif !ok {",
         ["TestTheChunkToolsLeaveAnExpiredDocumentOut"]),
        ("expand_graph : nœuds gardés", S,
         "&& !db.allowDocument(ctx, nil, docID) {",
         "&& docID == \"\\x00\" {",
         ["TestTheChunkToolsLeaveAnExpiredDocumentOut"]),
        ("expand_graph : arêtes gardées", S,
         "\t\tif _, gone := dropped[edge.FromNodeID]; gone {\n\t\t\tcontinue\n\t\t}\n\t\tif _, gone := dropped[edge.ToNodeID]; gone {\n\t\t\tcontinue\n\t\t}",
         "\t\tif _, gone := dropped[edge.FromNodeID]; gone && false {\n\t\t\tcontinue\n\t\t}\n\t\tif _, gone := dropped[edge.ToNodeID]; gone && false {\n\t\t\tcontinue\n\t\t}",
         ["TestTheChunkToolsLeaveAnExpiredDocumentOut"]),
        ("Balayage local malgré CORTEXDB_REMOTE", W,
         "\tif addr, token, ok := remoteConfigured(); ok {",
         "\tif addr, token, ok := remoteConfigured(); ok && false {",
         ["TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet"]),
        ("dry_run non transmis au cerveau partagé", W,
         "cortexdb.SweepExpiredToolRequest{DryRun: dryRun}",
         "cortexdb.SweepExpiredToolRequest{DryRun: false}",
         ["TestTheSweepRunsOnTheSharedBrainWhenRemoteIsSet"]),
        ("Outil sweep_expired non déclaré", "pkg/cortexdb/graphrag_tool_defs.go",
         "\tdefinitions = append(definitions, expiryToolDefinitions()...)\n",
         "",
         ["TestTheSweepIsAToolOfTheSharedBrain", "TestToolCountIsUnchangedByTheCursorWork"]),
        ("Outil sweep_expired non dispatché", "pkg/cortexdb/expiry_tool.go",
         "\tif name != \"sweep_expired\" {",
         "\tif name != \"\\x00\" {",
         ["TestTheSweepIsAToolOfTheSharedBrain"]),
        ("dry_run de l'outil ignoré", "pkg/cortexdb/expiry_tool.go",
         "return db.SweepExpired(ctx, time.Now(), req.DryRun)",
         "return db.SweepExpired(ctx, time.Now(), false)",
         ["TestTheSweepIsAToolOfTheSharedBrain"]),
        ("Index sur une autre expression", C,
         "ON documents ((CASE WHEN json_valid(metadata) = 1 THEN json_extract(metadata, '$.expires_at') ELSE NULL END));",
         "ON documents (author);",
         ["TestTheDocumentExpiryIsIndexed"]),
        ("Prédicat que l'index ne sert pas", E,
         "` FROM documents WHERE ` + expr + ` > ''`",
         "` FROM documents WHERE ` + expr + ` IS NOT NULL`",
         ["TestTheDocumentExpiryIsIndexed"]),
    ]),
    "06": ("./pkg/cortexdb/", "Review|Gate|Reviewer|Cleared|Acceptance|Indexed|Decision|Expir", [
        ("Relecteur du document : le client, pas la clé", R,
         "\t\tmetadata[reviewedByKey] = writer\n\t}\n\tat, err := checkReview(status, metadata[reviewedByKey]",
         "\t\t_ = writer\n\t}\n\tat, err := checkReview(status, metadata[reviewedByKey]",
         ["TestTheReviewerIsTheKeyThatDecided"]),
        ("Relecteur forcé sur une proposition (document)", R,
         "(status == ReviewAccepted || status == ReviewRejected) {\n\t\tmetadata[reviewedByKey] = writer\n\t}\n\tat, err",
         "(status == ReviewAccepted || status != ReviewRejected) {\n\t\tmetadata[reviewedByKey] = writer\n\t}\n\tat, err",
         ["TestTheReviewerIsTheKeyThatDecided"]),
        ("Relecteur forcé sur une proposition (mémoire)", R,
         "(status == ReviewAccepted || status == ReviewRejected) {\n\t\tmetadata[reviewedByKey] = writer\n\t}\n\tby, _ :=",
         "(status == ReviewAccepted || status != ReviewRejected) {\n\t\tmetadata[reviewedByKey] = writer\n\t}\n\tby, _ :=",
         ["TestTheReviewerIsTheKeyThatDecided"]),
        ("Relecteur de la mémoire : le client, pas la clé", R,
         "\t\tmetadata[reviewedByKey] = writer\n\t}\n\tby, _ := stringFromAny",
         "\t\t_ = writer\n\t}\n\tby, _ := stringFromAny",
         ["TestTheReviewerIsTheKeyThatDecided"]),
        ("Effacement admis sur un document", R,
         "\tstatus := metadata[reviewStatusKey]\n\tif status == \"\" {\n\t\tif prev.status != \"\" {",
         "\tstatus := metadata[reviewStatusKey]\n\tif status == \"\" {\n\t\tif prev.status != \"\" && false {",
         ["TestAReviewDecisionIsNotClearedSilently"]),
        ("Effacement admis sur une mémoire", R,
         "\tprev := reviewOf(before)\n\tif status == \"\" {\n\t\tif prev.status != \"\" {",
         "\tprev := reviewOf(before)\n\tif status == \"\" {\n\t\tif prev.status != \"\" && false {",
         ["TestAReviewDecisionIsNotClearedSilently"]),
        ("Décision non reportée quand la requête n'en parle pas", R,
         "\tif !named {\n\t\tif prev.status != \"\" {",
         "\tif !named {\n\t\tif prev.status != \"\" && false {",
         ["TestAReviewDecisionIsNotClearedSilently"]),
        ("Statut non chaîne lu comme absent", R,
         "\tif present && raw != nil && !isString {",
         "\tif present && raw != nil && !isString && false {",
         ["TestAReviewDecisionIsNotClearedSilently"]),
        ("reviewed_at redaté à chaque sauvegarde", R,
         "reviewedBy == prev.by && prev.at != \"\" {",
         "reviewedBy == prev.by && prev.at != \"\" && false {",
         ["TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate"]),
        ("expires_at perdu à l'acceptation", R,
         "\tif at, _ := stringFromAny(before[expiresAtKey]); at != \"\" {",
         "\tif at, _ := stringFromAny(before[expiresAtKey]); at != \"\" && false {",
         ["TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate"]),
        ("Acceptation qui remplace au lieu de fusionner", K,
         "\tcase reviewNamed(req.Metadata):\n",
         "\tcase reviewNamed(req.Metadata) && false:\n",
         ["TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate"]),
        ("Décision stockée gardée sous la nouvelle", R,
         "\tdelete(stored, reviewedAtKey)\n",
         "",
         ["TestAnAcceptanceKeepsTheOtherSystemKeysAndItsDate"]),
        ("Porte ouverte sur panne", E,
         "\t\treturn !db.reviewGate\n",
         "\t\treturn true\n",
         ["TestTheGateFailsClosedWhenTheBrainCannotBeRead"]),
        ("Index review_status sur une autre expression", C,
         "idx_documents_review_status ON documents ((CASE WHEN json_valid(metadata) = 1 THEN json_extract(metadata, '$.review_status') ELSE NULL END));",
         "idx_documents_review_status ON documents (author);",
         ["TestTheReviewStatusIsIndexed"]),
        ("Prédicat review_status que l'index ne sert pas", E,
         "` > '' OR ` + review + ` > ''`",
         "` > '' OR ` + review + ` IS NOT NULL`",
         ["TestTheReviewStatusIsIndexed"]),
    ]),
}
pkgs, run, MUTATIONS = SETS[which]


def run_tests():
    p = subprocess.run(["go", "test", *pkgs.split(), "-run", run], cwd=wt, capture_output=True, text=True)
    failed = {ln.split()[2] for ln in p.stdout.splitlines() if ln.startswith("--- FAIL")}
    return p.returncode, failed, p.stdout + p.stderr


ok = True
print("| Règle désactivée | Mutation | Test qui échoue |")
print("| :- | :- | :- |")
for rule, fname, before, after, expected in MUTATIONS:
    f = wt / fname
    src = f.read_text()
    if src.count(before) != 1:
        print(f"| {rule} | **motif introuvable** ({src.count(before)}) | |")
        ok = False
        continue
    f.write_text(src.replace(before, after))
    try:
        code, failed, out = run_tests()
    finally:
        f.write_text(src)
    if "build failed" in out:
        print(f"| {rule} | `{fname}` | **ne compile pas** |")
        ok = False
        continue
    missing = [t for t in expected if t not in failed]
    mark = "" if not missing else " **(attendu non tombé : " + ", ".join(missing) + ")**"
    if missing:
        ok = False
    print(f"| {rule} | `{fname}` : `{(after.strip().splitlines() or ['(ligne retirée)'])[-1][:60]}` | {', '.join(sorted(failed)) or '**aucun**'}{mark} |")

code, failed, _ = run_tests()
print()
print("Arbre restauré, suite verte :", code == 0 and not failed)
sys.exit(0 if ok and code == 0 else 1)
