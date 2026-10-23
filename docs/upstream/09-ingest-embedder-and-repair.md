# 09. `ingest_document` ignore l'encodeur, et `vector_dimension_repair` efface la provenance

**Statut** : **corrigés en v2.137.0** (23 octobre 2026, issue #84 fermée par le propriétaire : c2fd833 pour l'encodeur, avec la marque `_vectors` ; 122cede pour la réparation, qui rend aussi `doc_id` et les métadonnées des lignes vidées, mais pas les ACL). Fiche close. Historique : 2 défauts confirmés sur `main` (`0328e0a`, v2.136.0) le 22 octobre 2026. Issue ouverte le même jour sur commande de Lauro : liliang-cn/cortexdb#84. La PR suit si le propriétaire la veut. **Taille estimée de la PR** : environ 60 lignes hors tests pour le défaut 2. Le défaut 1 est une décision de conception, à laisser au propriétaire.

Origine : le pilote de contaixt sur CortexDB (`~/Agents/science-extractor`, `CHANTIER_PILOTE_CORTEXDB.md`), avec un encodeur local servi à `cortexdb-grpc` par `OPENAI_BASE_URL`.

## Défaut 1 : `ingest_document` écrit des vecteurs lexicaux, même avec un encodeur

- `GraphRAGToolbox.IngestDocument` (`pkg/cortexdb/graphrag_tool_ingest.go`) calcule chaque vecteur par `lexicalVectorForText`, de dimension `lexicalVectorDim` (64 par défaut). Le commentaire dit : « stores lexical chunks and graph nodes without requiring an embedder ».
- `knowledge_save` utilise l'encodeur quand il existe. Il prend le plan lexical (`buildLexicalKnowledgePlan`) seulement sans encodeur.
- Les 2 voies d'ingestion ne sont donc pas d'accord quand un encodeur est branché. Les morceaux de `ingest_document` ont une autre dimension que l'encodeur. Le rapport de dimension les signale, et ils n'entrent pas dans l'index vectoriel.
- Thalamus est touché : `th_add` passe par `ingest_document`.

Reproduction (`fixedDimEmbedder` de dimension 8) :

```
embedder calls: knowledge_save 2, ingest_document 0
ingest_document vector bytes 260, embedder dimension 8 (36 bytes)
--- FAIL: TestReproIngestDocumentIgnoresTheEmbedder
```

Correctifs possibles, au choix du propriétaire :
1. `ingest_document` utilise l'encodeur quand il existe, comme `knowledge_save`.
2. `ingest_document` reste lexical par conception, et sa description le dit. L'appelant avec un encodeur utilise alors `knowledge_save`.

## Défaut 2 : `vector_dimension_repair` vide le document, les métadonnées et les droits

- `SQLiteStore.MismatchedEmbeddings` (`pkg/core/dimension_report.go`) relit chaque ligne avec 3 colonnes : `SELECT e.id, e.content, c.name`.
- `ReembedMismatchedVectors` (`pkg/cortexdb/reembed.go`) pose le nouveau vecteur sur cet objet et l'écrit par `UpsertBatch`. La ligne réécrite perd `doc_id`, `metadata` et `acl`.
- Mesure du pilote : 12 218 lignes touchées. La part des réponses dont la fiche citée est dans le contexte est tombée de 76 % à 2 %.
- Les 2 défauts s'enchaînent : un encodeur branché rend les morceaux de `ingest_document` mal dimensionnés, et leur réparation efface leur provenance.

Reproduction :

```
re-embedded 1 rows;
before: doc_id=doc-1 metadata={"chunk_index":"0","document_id":"doc-1","graph_kind":"chunk","source":"wiki","title":""} acl=["team"]
after:  doc_id=(null) metadata=(empty) acl=(null)
--- FAIL: TestReproRepairKeepsTheRowsProvenance
```

Règle du correctif : une ligne recalculée garde son document, ses métadonnées et ses droits. Seul le vecteur change.

- Voie A : `MismatchedEmbeddings` relit aussi `doc_id`, `metadata` et `acl`, sur SQLite et sur PostgreSQL (`pkg/core/store_postgres_dims.go`).
- Voie B : la réparation écrit seulement la colonne `vector` (un `UPDATE` ciblé), sans `UpsertBatch`.
- Le test : le test de reproduction ci-dessus, sur les deux moteurs.

## Les tests de reproduction

Fichier `pkg/cortexdb/repro_thalamus_test.go`, sur une copie de `main` hors du dépôt. Les deux tests utilisent les aides de `reembed_test.go` (`fixedDimEmbedder`, `newReembedDB`). Commande :

```
go test ./pkg/cortexdb/ -run 'TestRepro' -count=1 -v
```

Le texte des deux tests est dans l'issue.
