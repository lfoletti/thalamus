# Fiches amont 02, 03 et 05 relues sur CortexDB v2.138.0

Date : 23 octobre 2026. Clone : `cortexdb/` sur `proto/granular-cortex` = `b2a1ab6` (tag v2.138.0). Sondes : un test Go jetable dans un worktree détaché de v2.138.0 (`wt02`, retiré à la fin), sans `CORTEXDB_TEST_POSTGRES`. Rien n'est poussé, aucune branche n'a changé.

Les numéros de ligne renvoient à v2.138.0. Chaque section sépare **Faits** (lus dans le code ou observés en test) et **Hypothèses** (jugements, à valider).

---

## 02. Outil `review_pending`

### Faits

- Aucun outil `review_pending` ni rien de semblable dans la toolbox (liste des `Name:` de `pkg/cortexdb/*tooldefs*.go` et `graphrag_tool_defs.go`).
- Le propriétaire a désigné lui-même la voie de listage : la description de `filters` d'`aggregate_metadata` dit « A filter on _review_status counts rows of that status only, which is how a curator lists what waits for review (_review_status: proposed) under the review gate » (`pkg/cortexdb/graphrag_tool_defs.go:178`). Le filtre devient une condition sur la table `documents` (`pkg/cortexdb/aggregate.go:33-55`, `reviewStatusIs`). `representative_records` a la même phrase (`graphrag_tool_defs.go:192`).
- La liste blanche laisse une clé confinée nommer `_review_status` dans un filtre (`pkg/authz/allowlist.go:486-491`, commentaire « that is how a curator lists what waits » ; règles `aggregate_metadata` et `representative_records` à `allowlist.go:207-210`).
- Sonde, porte de relecture active, trois documents (proposé, accepté, sans statut) :
  - `aggregate_metadata {type: group_by, group_by: [document_id], filters: {_review_status: proposed}}` rend `[{keys: {document_id: d1}, count: 1}]` : les identifiants, sans titre, auteur ni texte (`aggregate.go:153-182`, `ToolAggregateMetadataGroup`).
  - `representative_records {group_by: document_id, filter: {_review_status: proposed}}` rend un morceau par document (le médoïde), avec `content`, sans titre, sans auteur, sans `limit` (`aggregate.go:185-235`). Pour un document de plusieurs morceaux, ce n'est pas le début du texte.
  - `knowledge_search` avec `query: ""` et `plan.filters.review_status: proposed` : `cortexdb: empty text provided` (`pkg/cortexdb/knowledge_api.go:273-275`, `EmptyQueryUsesGraph: false`). Avec une requête non vide, le document proposé sort avec titre, auteur et extrait.
- `knowledge_get` n'a qu'une forme à un identifiant (`KnowledgeGetRequest`, `pkg/cortexdb/knowledge_memory_types.go:100-107` ; `GetKnowledge`, `knowledge_api.go:242-248`). Sa règle de liste blanche n'admet aucun argument hors `knowledge_id` (`allowlist.go:218`). Sonde REST, clé confinée : `knowledge_get {knowledge_id, collection}` → 403 « takes no collection argument a confined key may send ». La prémisse de la fiche tient.
- `memory_get` a une forme liste (`memory_ids`) : un identifiant hors de la portée de l'appelant va dans `missing`, comme un identifiant inexistant (`pkg/cortexdb/memory_api.go:242-273` ; règle `allowlist.go:236`, `memory_ids` classé en donnée, `oneOf` avec `memory_id`). L'aide `knowledgeReadPermitted(ctx, collection)` existe déjà côté connaissance (`pkg/cortexdb/memory_owner.go:116-121`).
- Le document proposé reste lisible par identifiant (`ReviewGateNote`, `pkg/cortexdb/review.go:30`).
- Côté Thalamus : `th_pending` fait `aggregate_metadata` puis un `knowledge_get` par document (`gateway/relay/write.go:202-261`, `listPending`), et coupe un extrait de 200 caractères (`write.go:263-273`, `snippet`).

### Verdict : nécessaire sous une autre forme, priorité basse

Un outil `review_pending` ferait double emploi avec la voie que le propriétaire a documentée (`aggregate_metadata` + `_review_status`). Il est probable qu'il le refuse. Le coût restant est le N + 1 de la lecture. La forme qui le supprime, et qui suit un précédent du propriétaire, est **la forme liste de `knowledge_get`** (`knowledge_ids`), symétrique de `memory_get`. Thalamus passe alors de N + 1 appels à 2.

Hypothèse : le gain est modeste (un curateur liste rarement plus de 20 documents). Cette PR passe après 05 et 03.

### Conception mise à jour : `knowledge_get` lit une liste

**Règle** : `knowledge_get {knowledge_ids: [...]}` rend les documents dans l'ordre demandé. Un identifiant inexistant, ou dont la collection n'est pas permise à l'appelant (`knowledgeReadPermitted`), va dans `missing`, sans distinction possible entre les deux. La forme à un identifiant ne change pas : elle reste non confinée dans la façade, parce que le garde de lignes des ports lit à travers elle (même raison que `GetMemory`, `memory_api.go:244-249`). Un appel donne une forme au plus (`oneOf`).

**Fichiers et lignes hors tests (estimation)** :

| Fichier | Changement | Lignes |
| :- | :- | -: |
| `pkg/cortexdb/knowledge_memory_types.go` | `KnowledgeIDs []string` dans la requête ; `Knowledges []KnowledgeRecord`, `Missing []string` dans la réponse | 8 |
| `pkg/cortexdb/knowledge_api.go` | branche liste de `GetKnowledge` : `loadKnowledgeRecord`, `ErrNotFound` → `missing`, `knowledgeReadPermitted(ctx, record.Collection)` → `missing` | 25 |
| `pkg/cortexdb/knowledge_memory_tooldefs.go` | schéma : `knowledge_ids`, `knowledge_id` n'est plus requis | 4 |
| `pkg/authz/allowlist.go` | `knowledge_get` : `args: {"knowledge_ids"}`, `oneOf: {"knowledge_id","knowledge_ids"}`, commentaire | 5 |
| `pkg/rpcserver/ownership.go` (à vérifier) | `guardToolRows` ne doit rien charger pour la forme liste | 0 à 5 |
| Total | | 40 à 50 |

Le serveur MCP local passe par le même type (`pkg/cortexdb/mcp.go:329`), sans code en plus.

**Table outil par outil (clé confinée)** :

| Outil | Port | Avant | Après |
| :- | :- | :- | :- |
| `knowledge_get` (forme simple) | gRPC `CallTool`, REST `/v1/tools/knowledge_get` | garde de lignes sur `knowledge_id` | inchangé |
| `knowledge_get` (forme liste) | idem | refusé (argument non classé) | `knowledge_ids` en donnée ; la façade met dans `missing` toute collection hors portée |
| `knowledge_get` avec les deux formes | idem | refusé | refusé (`oneOf`) |
| RPC typé `GetKnowledge`, REST `GET /v1/knowledge` | | une ligne par identifiant | inchangé (pas de changement de proto) |

**Tests à écrire d'abord** :

1. `TestKnowledgeGetListsSeveralInTheOrderAsked` (`pkg/cortexdb`) : `[k2, absent, k1]` → `knowledges = [k2, k1]`, `missing = [absent]`.
2. `TestKnowledgeGetListHidesAnotherCollectionAsMissing` (`pkg/cortexdb`, `WithCallerScope` avec `Knowledge` sur `notes`) : un document de `ops` sort dans `missing`, au même rang qu'un absent.
3. `TestAConfinedKeyListsOnlyItsCollectionWithKnowledgeGet` (`pkg/rpcserver` et `pkg/httpapi`, une clé `notes`) : même chose de bout en bout, statut 200.
4. `TestKnowledgeGetTakesOneFormFromAConfinedKey` (`pkg/authz`, calqué sur `TestMemoryGetTakesOneFormFromAConfinedKey`, `allowlist_test.go:492`).
5. `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys` (existant, `allowlist_test.go:53`) tombe tant que `knowledge_ids` n'est pas classé.
6. Un document proposé sous la porte, et un document expiré : la forme liste rend ce que la forme simple rend pour le même identifiant.

**Mutations à la main** :

| Mutation | Test qui doit tomber |
| :- | :- |
| `knowledgeReadPermitted` retiré de la branche liste | 2, 3 |
| `missing` jamais rempli (absent ignoré en silence) | 1 |
| ordre perdu (tri des identifiants) | 1 |
| `oneOf` retiré de la règle | 4 |
| branche liste rendue au profit de la forme simple quand les deux sont là | 4 (refus) et 1 |
| règle de liste blanche non mise à jour | 5 |

**Après la fusion, côté Thalamus** : `listPending` fait un `aggregate_metadata` et un `knowledge_get {knowledge_ids}`. `snippet` reste, puisque le moteur ne coupe pas d'extrait. Environ 15 lignes en moins, pas 60.

---

## 03. Identifiants touchés dans le journal des requêtes

### Faits

- Le journal des requêtes est `pkg/rpcserver/trace.go`. Son enregistrement JSON a sept champs : `ts`, `method`, `key`, `tool`, `dur_ms`, `stages_ms`, `err` (`trace.go:58-66`). Aucun identifiant de ligne, aucun identifiant de requête.
- Une ligne est écrite seulement avec `CORTEXDB_TRACE=1`, ou pour un appel lent (`CORTEXDB_SLOW_MS`, 1 s par défaut) (`trace.go:126-129`). `CORTEXDB_TRACE_FILE` en fait du JSONL (`trace.go:52-57`).
- Le nom de l'outil vient du champ `name` de `CallToolRequest` (`trace.go:136-138`, `requestToolName` à `pkg/rpcserver/auth.go:103-119`). Les arguments ne sont pas lus.
- Ce journal n'existe que sur le port gRPC (`cmd/cortexdb-grpc/main.go:158`). `pkg/httpapi` n'a pas de journal des requêtes (aucun `log.Print` hors tests).
- Correction d'historique : le nom de la clé et de l'outil dans le journal vient de **#68** (`2710212`, « request log names the key and the tool of each call »), pas de #67 comme l'écrit la fiche. Le journal lui-même vient du propriétaire (`7032864`).
- Le flux de changements (`pkg/graph/changefeed.go:104-120`, outil `changes_since`) nomme les lignes écrites (`ID`) mais pas la clé, ni les lectures. Il est refusé aux clés confinées (`allowlist.go:261`).
- La liste blanche déclare, outil par outil, les chemins qui nomment des lignes (`toolRule.ids`, `allowlist.go:99-100`). `JSONPathValues(args, path)` lit déjà ces chemins, `[]` compris (`allowlist.go:615-674`).
- Deux chemins qui nomment des lignes sont classés en **donnée**, pas en `ids` : `memory_ids` de `memory_get` (`allowlist.go:236`) et `plan.filters.document_ids` (`allowlist.go:121-128`). Une fonction qui ne lirait que `ids` les manquerait.
- Côté Thalamus : `idsOf` prend les champs de premier niveau en `_id`, `_ids` et `supersedes` (`gateway/relay/relay.go:174-193`). `decide` garde une ligne nommée pour les écritures, et pour les lectures d'un cerveau `audit_reads` seulement (`gateway/relay/audit.go:70-87`). `th_recall` y ajoute les `knowledge_id` lus (`audit.go:89-115`).

### Verdict : nécessaire sous une autre forme

Le serveur ne journalise toujours aucun identifiant. Le besoin tient, et il vaut aussi pour un client qui parle au serveur sans passer par Thalamus : une clé `th/<login>/<cerveau>` marche directement contre `cortexdb-grpc` quand le port est joignable. Seul le serveur peut alors dire ce que l'appel a touché. Trois changements par rapport à la fiche :

1. **Une règle par PR** : les identifiants des arguments seulement. Les identifiants rendus (résultats de `knowledge_search`) forment une autre règle, et Thalamus les a déjà. À proposer plus tard s'il y a une demande.
2. Les chemins lus sont `ids` **plus** les deux chemins de donnée qui nomment des lignes (`memory_ids`, `plan.filters.document_ids`). Il faut une liste déclarée à part, tenue par le test de complétude, plutôt que `ids` seul.
3. gRPC seulement. Le port REST n'a pas de journal, et en ajouter un est une autre PR.

L'idée de joindre les deux journaux par un identifiant de requête du serveur ne tient pas en l'état, puisque cet identifiant n'existe pas. Hypothèse : une PR minuscule qui recopie une métadonnée gRPC `x-request-id` du client dans l'enregistrement ferait l'affaire. C'est une règle à part, à ne pas mêler à celle-ci.

### Conception mise à jour

**Règle** : quand `CallTool` journalise un appel, l'enregistrement porte `ids`, un objet `{chemin: [identifiants]}` qui contient les valeurs présentes aux chemins que la liste blanche déclare comme nommant des lignes pour cet outil. Un outil hors liste blanche n'a pas de champ `ids`. La ligne texte gagne ` ids=<n>` (le nombre, pas les valeurs, pour ne pas remplir le journal standard).

**Fichiers et lignes hors tests (estimation)** :

| Fichier | Changement | Lignes |
| :- | :- | -: |
| `pkg/authz/allowlist.go` | `rowPaths` par règle (les `ids`, plus `memory_ids` et `plan.filters.document_ids` déclarés à part) ; `func RowIDsOf(tool string, args []byte) map[string][]string` sur `JSONPathValues` | 35 |
| `pkg/rpcserver/trace.go` | lecture de `args_json` dans `CallToolRequest`, champ `IDs` de `traceRecord`, ` ids=n` | 25 |
| Total | | environ 60 |

**Table outil par outil (chemins journalisés)** : la table vient de `ConfinedToolsTable()` (`allowlist.go:676`) avec une colonne en plus. Les lignes qui changent par rapport à `ids` :

| Outil | `ids` (garde) | journalisé en plus |
| :- | :- | :- |
| `memory_get` | `memory_id` | `memory_ids` |
| `knowledge_search`, `search_text`, `search_graphrag_lexical`, `memory_search`, `knowledge_memory_recall`, `knowledge_memory_build_context_pack` | aucun | `plan.filters.document_ids` |
| tous les autres outils de la liste blanche | leurs `ids` | rien |
| outil hors liste blanche (clé non confinée) | | pas de champ `ids` |

Le comportement d'accès ne change pour aucun outil : la PR n'ajoute que de l'écriture de journal.

**Tests à écrire d'abord** :

1. `TestRowIDsOfReadsEveryDeclaredPath` (`pkg/authz`) : pour chaque outil de la liste blanche, des arguments avec une valeur à chaque chemin déclaré (tableaux `[]` compris), toutes retrouvées ; un champ `foo_id` non déclaré ignoré.
2. `TestEveryArgumentThatNamesARowIsLogged` (`pkg/authz`) : tout chemin de `ids` est dans `rowPaths`, et `memory_ids` et `plan.filters.document_ids` y sont aussi. Ce test garde la liste à jour quand un outil change.
3. `TestTheRecordNamesTheRowsTheCallNames` (`pkg/rpcserver`, à côté de `TestTheRecordNamesTheKeyAndTheTool`, `trace_test.go:56`) : `knowledge_get {knowledge_id: k1}` donne `ids: {knowledge_id: [k1]}` dans le JSONL ; `memory_get {memory_ids: [a, b]}` donne les deux ; `knowledge_search` sans pin n'a pas de champ `ids`.
4. `TestAToolOffTheAllowlistLogsNoIDs` (`pkg/rpcserver`) : `graph_cypher_query` sous une clé non confinée, pas de champ `ids`.

**Mutations à la main** :

| Mutation | Test qui doit tomber |
| :- | :- |
| `[]` ignoré dans les chemins (élément de tableau sauté) | 1 |
| `memory_ids` retiré de `rowPaths` | 2, 3 |
| `plan.filters.document_ids` retiré | 2 |
| champ `IDs` jamais rempli dans `traceRecord` | 3 |
| repli sur une heuristique `_id` pour les outils hors liste | 4 |

**Après la fusion, côté Thalamus** : `idsOf` peut rester, puisque l'audit du gateway a besoin du nom de l'employé et que le journal du serveur n'est écrit qu'avec `CORTEXDB_TRACE=1`. Hypothèse : le gain réel est que le journal du serveur couvre les appels directs. Ce n'est pas une simplification de Thalamus.

---

## 05. `author` et `written_by` depuis la clé appelante

### Faits

- `WithWriter(ctx, name)` met le nom de la clé dans le contexte (`pkg/cortexdb/written_by.go:14-25`). Les ports le posent pour toute clé, confinée ou non : `Key.CallerContext` (`pkg/authz/caller.go:22-23`), appelé par gRPC (`pkg/rpcserver/auth.go:74`) et REST (`pkg/httpapi/policy.go:66`). Ce code vient du propriétaire (`0d3e601`, 19 octobre 2026), pas du lot A.
- `stampWriter` remplace `written_by` par la clé et retire `updated_by` (`written_by.go:32-41`). `stampUpdater` garde le `written_by` d'origine et met `updated_by` à la clé (`written_by.go:43-56`). Ils sont appelés **seulement** par `SaveKnowledge` (`pkg/cortexdb/knowledge_api.go:45`), `UpdateKnowledge` (`knowledge_api.go:138`), `SaveMemory` (`pkg/cortexdb/memory_api.go:131`) et `UpdateMemory` (`memory_api.go:218`).
- **`ingest_document` ne signe pas.** `IngestDocument` (`pkg/cortexdb/graphrag_tool_ingest.go:20-160`) n'appelle pas `stampWriter`. Le propriétaire le dit lui-même : commentaire de `writerKeys`, « The facade stamps them on SaveKnowledge, UpdateKnowledge, SaveMemory and UpdateMemory, but not on ingest_document » (`pkg/authz/allowlist.go:161-164`) ; test `pkg/httpapi/confined_ingest_marks_test.go:75-78`, « A document written by ingest_document carries them and no signature ». `InsertGraphDocument` ne signe pas non plus (`pkg/rpcserver/ownership.go:139-140`).
- Sonde, `ingest_document` sous `WithWriter("th/alice/notes")` avec `metadata: {written_by: forged}` : le document garde `written_by: forged`. Sans `written_by` dans la requête, le document n'a pas de signature. Même sonde par `SaveKnowledge` : `written_by: th/alice/notes`, la valeur du client est remplacée.
- Une clé confinée ne peut pas envoyer `metadata.written_by` (`allowlist.go:506-508`). Une clé **non confinée** peut, et la valeur fausse est gardée. Chez Thalamus, c'est le cas d'un cerveau sans `collection` (serveur entier, clé non confinée).
- `author` vient toujours du client, sur `ingest_document` (`graphrag_tool_ingest.go:131`, `:151-153`) comme sur `knowledge_save` (`knowledge_api.go:60`). Aucun code ne le tire de la clé. Le schéma le décrit comme une donnée déclarée : « Optional author. Kept on the document record and returned with every search hit » (`pkg/cortexdb/graphrag_tool_defs.go:35`).
- Ce que montrent la recherche et la provenance (sonde `knowledge_search`) : chaque résultat a `author` en champ propre et `metadata` en entier, donc `written_by` quand il existe (`KnowledgeSearchHit`, `pkg/cortexdb/knowledge_memory_types.go:148-166` ; remplissage `pkg/cortexdb/knowledge_helpers.go:224-225`). `knowledge_get` rend `author` et `metadata` (`knowledge_helpers.go:140-142`).
- Nouveau, observé en sonde REST sur v2.138.0, clé confinée `notes` : un document écrit par `knowledge_save` porte `written_by: notes`. Renvoyer ses métadonnées par `/v1/tools/knowledge_save` ou `/v1/tools/ingest_document` donne **403** « takes no metadata.written_by from a confined key ». Les routes typées, elles, retirent `written_by` et `updated_by` avant la liste blanche, parce que la façade les signe (`pkg/httpapi/policy.go:229-266`, `pkg/rpcserver/ownership.go:137-163`). Sur la route des outils, `written_by` n'est pas dans les clés réécrites (`cortexdb.EngineDocumentKeys`, `graphrag_tool_ingest.go:276-278`). Le refus est sûr, mais l'aller-retour lecture-écriture ne marche pas.
- Sonde de mutation : avec `stampWriter` ajouté seul à `IngestDocument`, `pkg/authz`, `pkg/rpcserver` et `cmd/cortexdb-mcp-stdio` restent verts. `pkg/httpapi` tombe sur `TestAToolWriteTakesTheMarksTheFacadeReplaces` (`confined_ingest_marks_test.go:128`) : le document relu porte maintenant `written_by`, et la clé confinée qui le renvoie est refusée. La signature et la reprise des clés signées vont donc ensemble.
- Côté Thalamus : `th_add` envoie `author: user` (`gateway/relay/write.go:155`). Aucun code du gateway ne lit `written_by`.

### Verdict : nécessaire sous une autre forme

La prémisse « `written_by` porte déjà l'identifiant de la clé » est fausse pour `ingest_document`, le chemin que prend `th_add`. La partie `author` de la fiche est à abandonner. Le propriétaire a séparé `author`, donnée déclarée par le client, de `written_by`, signature posée par le moteur, et les deux sont rendus. Faire de `author` la clé irait contre ce partage. Ce serait faux aussi pour une page web proposée par l'extension, dont l'auteur n'est pas l'employé. Le vrai trou : **`ingest_document` ne signe pas**. Le combler avec la reprise des clés signées sur la route des outils est une seule règle, celle que le propriétaire a déjà écrite pour `SaveKnowledge`.

### Conception mise à jour : toute écriture de document signe son auteur

**Règle** : `ingest_document` signe comme `knowledge_save`. Avec un écrivain dans le contexte, `written_by` est la clé et `updated_by` est retiré, quoi que le client envoie. Sans écrivain (bibliothèque, serveur sans clés), les métadonnées sont gardées telles quelles. Comme la façade réécrit `written_by` et `updated_by` sur les trois écritures de document, la liste blanche les reprend sous leur graphie exacte sur ces trois outils, comme les marques (`rewritten`). Une autre graphie (`Written_By`) reste refusée. Les écritures de mémoire et `InsertGraphDocument` restent hors de cette PR.

**Décision à prendre, à tester explicitement** : la signature est posée avant `sameIngest` (`graphrag_tool_ingest.go:358`). Une ré-ingestion identique par la même clé reste « unchanged ». La même ré-ingestion par une autre clé réécrit le document et le signe de la nouvelle clé, comme `SaveKnowledge` remplace `written_by` à chaque sauvegarde. Hypothèse : c'est le comportement attendu, et une tâche de synchronisation tourne sous une seule clé.

**Fichiers et lignes hors tests (estimation)** :

| Fichier | Changement | Lignes |
| :- | :- | -: |
| `pkg/cortexdb/graphrag_tool_ingest.go` | `stampWriter(ctx, metadata, plainString)` après `carryExpiry` ; commentaire | 4 |
| `pkg/cortexdb/written_by.go` | `func WriterKeys() []string` exporté (`written_by`, `updated_by`) | 6 |
| `pkg/authz/allowlist.go` | `rewritten` des trois outils de document = `EngineDocumentKeys` + `WriterKeys` ; commentaire de `writerKeys` (« but not on ingest_document » devient faux) | 10 |
| Total | | environ 20 |

C'est bien en dessous de 400 lignes, avec une seule règle.

**Table outil par outil (clé confinée)** :

| Outil | `metadata.written_by` / `updated_by` envoyé, avant | après | ligne stockée, après |
| :- | :- | :- | :- |
| `ingest_document` | refusé (403) | repris | `written_by` = clé, `updated_by` absent |
| `knowledge_save` | refusé (403) | repris | `written_by` = clé (déjà le cas) |
| `knowledge_update` | refusé (403) | repris | `written_by` d'origine, `updated_by` = clé (déjà le cas) |
| `Written_By` (autre graphie) sur les trois | refusé | refusé | |
| `memory_save`, `knowledge_memory_remember`, `memory_update` | refusé | refusé (hors PR) | inchangé |
| RPC typés et routes REST typées | retirés par le port | inchangé | inchangé |
| Clé non confinée, `ingest_document` | valeur du client **gardée** | remplacée par la clé | `written_by` = clé |

**Tests à écrire d'abord** :

1. `TestAnIngestRecordsItsWriterAndAClientCannotClaimAnother` (`pkg/cortexdb/written_by_test.go`) : sous `WithWriter("hermes")`, `ingest_document` avec `metadata: {written_by: zeus, updated_by: zeus}` stocke `written_by: hermes` et pas d'`updated_by` ; `author: zeus` est gardé tel quel.
2. `TestWithoutAWriterAnIngestStoresMetadataAsGiven` (même fichier, calqué sur `TestWithoutAWriterMetadataIsStoredAsGiven`, `written_by_test.go:76`).
3. `TestAReIngestByAnotherKeyIsSignedByIt` (`pkg/cortexdb`) : même texte, clé `hermes` puis `hermes` → `unchanged: true` ; puis clé `zeus` → `unchanged: false`, `written_by: zeus`.
4. `TestAToolWriteTakesTheMarksTheFacadeReplaces` (existant, `pkg/httpapi/confined_ingest_marks_test.go`) étendu : le document relu porte `written_by`, il est renvoyé par les trois outils avec `written_by: forged` (200), et la ligne garde la clé. Son jumeau gRPC aussi (`pkg/rpcserver`).
5. `TestAnotherSpellingOfTheSignatureIsStillRefused` (`pkg/authz`) : `metadata.Written_By` → refus sur les trois outils ; `metadata.written_by` sur `memory_save` → refus.

**Mutations à la main** :

| Mutation | Test qui doit tomber |
| :- | :- |
| `stampWriter` retiré d'`IngestDocument` | 1, 3, 4 (la ligne ne porte pas la clé) |
| `stampWriter` placé après `sameIngest` (signature hors comparaison) | 3 |
| `WriterKeys` absent du `rewritten` d'`ingest_document` | 4 (403, observé en sonde) |
| `WriterKeys` ajouté au `rewritten` de `memory_save` | 5 |
| comparaison de clés insensible à la casse dans `rewritten` | 5 |
| `stampWriter` sans le garde « pas d'écrivain » | 2 |

**Après la fusion, côté Thalamus** : `th_add` garde `author: user`, qui est une donnée lisible. Le moteur signe `written_by: th/<login>/<cerveau>`, que personne ne peut falsifier, même sur un cerveau à clé non confinée. Le rappel peut citer `written_by` comme preuve. Hypothèse : l'affichage reste `author`, avec `written_by` en contrôle dans l'audit.

---

## Ordre proposé

1. **05 sous sa nouvelle forme** (environ 20 lignes) : trou de signature avéré, reconnu dans le code du propriétaire, avec un effet de bord REST observé.
2. **03 réduit** (environ 60 lignes) : identifiants des arguments dans le journal gRPC.
3. **02 sous la forme `knowledge_get` liste** (40 à 50 lignes), priorité basse.

Le dossier `docs/upstream/` n'est pas modifié. Les fiches sont à réécrire sur ce rapport si Lauro valide.
