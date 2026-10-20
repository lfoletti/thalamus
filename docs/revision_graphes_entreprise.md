# Révision d'architecture : graphes de connaissances d'entreprise pour agents de code

Date : 18 octobre 2026. Version examinée : CortexDB 2.132.0 (dépôt local du plugin, `~/.claude/plugins/marketplaces/cortexdb`).

Convention de ce document :

- **[Vérifié : fichier]** : lu dans le code ou la documentation du dépôt, fichier cité.
- **[Supposé]** : déduit, non lu dans le code.
- **[À vérifier]** : point à confirmer avant de décider.

Contrainte de méthode retenue : le projet évolue par **forks successives et merges modestes**. Chaque manque identifié devient une petite contribution remontable en amont (PR). Aucun fork divergent de longue durée.

---

## 0. Réponse courte

1. CortexDB est un bon **moteur** : stockage, graphe, faits bitemporels, outils MCP, backend PostgreSQL, mode lexical sans modèle. Il n'est pas une **plateforme de gestion des connaissances** : pas de multi-cerveaux, pas d'audit par identité, pas de classification, pas de conservation, pas de relecture humaine intégrée.
2. Le point bloquant découvert dans le code : **une clé confinée ne peut appeler aucun outil du toolbox**. Le plugin MCP en mode distant passe par ce toolbox. Donc le modèle de droits actuel et le plugin MCP sont incompatibles. C'est corrigeable par une PR de taille moyenne.
3. Décision recommandée : **CortexDB comme moteur, derrière une couche mince maison (« Thalamus ») qui porte la fédération, les droits, la classification, l'audit et la file de relecture.** La couche s'amincit à chaque PR acceptée en amont.
4. Aucune alternative examinée (Graphiti, Mem0, Cognee, LightRAG) n'est meilleure sous la contrainte « Ollama seul + MCP + conformité ». Toutes exigent un modèle pour fonctionner. Aucune ne porte des droits par graphe et par document.
5. Deux exigences sont mal posées. Voir la section 7 : « mémoire personnelle de l'employé » et « traitement exclusivement interne ».

---

## 1. Diagnostic de CortexDB pour ce besoin

### 1.1 Granularité des graphes

| Besoin | État | Détail |
| :- | :- | :- |
| Un graphe par projet d'équipe | **Couvrable par configuration** | Un serveur `cortexdb-grpc` par cerveau, une base par serveur. [Vérifié : `cmd/cortexdb-grpc/main.go`, un seul `-db`] |
| Un graphe par employé | Couvrable par configuration | Idem : un serveur ou un fichier SQLite par employé. |
| Un graphe personnel par projet | Couvrable par configuration | Idem. Le nombre de processus croît vite : N employés × M projets. |
| Plusieurs cerveaux dans un seul serveur | **Hors de portée aujourd'hui** | Le serveur ouvre une base. Il n'y a pas de notion de « brain » adressable par requête. [Vérifié : `cmd/cortexdb-grpc/main.go`] |
| Isolation logique dans une seule base | Partielle | Les mémoires portent `user_id`, `scope`, `session_id`. Les connaissances portent `namespace`, `collection`. Le graphe GraphRAG porte `collection`. [Vérifié : `pkg/authz/authz.go` lignes 73-81] |

Conclusion : la granularité se fait soit par **processus** (un serveur par cerveau, simple et étanche, lourd en exploitation), soit par **namespace** dans une base (léger, étanchéité qui dépend des clés, voir 1.2). Le code ne propose pas de troisième voie.

État de la branche locale `proto/granular-cortex` (18 octobre 2026) : les correctifs 3 et 4 ouvrent la troisième voie. Une clé peut lister plusieurs collections, et une clé confinée peut appeler les outils qui déclarent le champ confiné. Une **collection par cerveau** dans un serveur partagé devient possible. Limite vérifiée : un nœud d'entité a pour identifiant son nom normalisé (`graphEntityNodeID`), sans collection. Les entités sont donc partagées entre les collections d'un même serveur, seuls les passages sont filtrés. Voir 3.9.

### 1.2 Combinaison sans choix

| Besoin | État | Détail |
| :- | :- | :- |
| Interroger plusieurs cerveaux ensemble | **Hors de portée** | Le client MCP ouvre un fichier (`CORTEXDB_PATH`) ou un serveur (`CORTEXDB_REMOTE`). Pas de liste. [Vérifié : `cmd/cortexdb-mcp-stdio/remote.go`, `main.go`] |
| Clé lisant plusieurs périmètres | **Impossible** | `Scope` a 4 champs (`user_id`, `scope`, `namespace`, `collection`), chaque champ reçoit **une seule valeur**. Pas de liste. [Vérifié : `pkg/authz/authz.go` lignes 73-98] |
| Clé confinée + outils MCP | **Bloquant** | Une clé confinée est refusée sur **tous** les outils du toolbox, sur gRPC et sur HTTP. Motif dans le code : « a tool's arguments are opaque JSON no scope can be checked against ». [Vérifié : `pkg/httpapi/access.go` lignes 100-104, « which is also gRPC's rule »] |
| Clé confinée + RPC typés | Couvert | Les services typés (`MemoryService`, `KnowledgeService`) vérifient le périmètre par réflexion sur les champs de la requête, et l'appartenance de la ligne pour les RPC par id. [Vérifié : `pkg/rpcserver/auth.go`, `pkg/rpcserver/ownership.go`] |
| Provenance dans les résultats | Partiel | Un résultat porte `document_id`, `chunk_index`, `title` dans les métadonnées du passage. Pas de date de source, pas d'URL, pas de cerveau d'origine. [Vérifié : `pkg/cortexdb/graphrag_tool_ingest.go` lignes 80-88] |
| Fusion et classement multi-sources | Hors de portée | Rien dans le code. Le client doit le faire. |

Conséquence pratique : aujourd'hui, un employé qui utilise le plugin MCP en mode distant a besoin d'une clé **non confinée** (lecture-écriture totale sur le cerveau) ou d'une clé en lecture seule non confinée. Le confinement par `namespace` ne protège rien pour les agents de code. Le seul confinement réel est **un serveur par cerveau**, avec une clé par serveur.

État de la branche locale `proto/granular-cortex` (18 octobre 2026) : les lignes « Clé lisant plusieurs périmètres » et « Clé confinée + outils MCP » sont couvertes (commits `f21e4d7`, `8c0d1ff`). La provenance est portée sur le document et le nœud (`3be4302`), avec le hash du contenu (`7edc498`). Une limite demeure : une clé confinée à la fois sur `user_id` et sur `collections` est refusée sur tous les outils, parce qu'aucun outil ne déclare les deux champs. La clé de l'employé reste donc dans le gateway, qui tient une clé par cerveau. La fusion multi-sources reste du côté du gateway.

### 1.3 Conformité

| Besoin | État | Détail |
| :- | :- | :- |
| Contrôle d'accès par graphe | Couvert par configuration | Un serveur par cerveau, un fichier de clés par serveur. Clearances `read-only` et `read-write`. [Vérifié : `pkg/authz/authz.go`, `authz_test.go`] |
| Contrôle d'accès par document | **Hors de portée** | Rien dans le code. Le grain le plus fin est `collection`. |
| Classification des documents | Hors de portée, **couvrable par extension** | `ingest_document` accepte `metadata` et le copie sur chaque passage. Il ne le stocke pas sur l'enregistrement du document (`core.Document` garde ID, Title, Content, Version). [Vérifié : `graphrag_tool_ingest.go` lignes 44-50 et 85-87] |
| Durée de conservation, effacement | Partiel | `delete_document_graph`, `knowledge_delete`, `memory_delete` existent (outils MCP). Pas de date d'expiration, pas de balayage. [Vérifié : liste des outils ; absence de `expires_at` dans `pkg/core` hors `liveview`] |
| Droit à l'effacement d'une personne | Hors de portée | Pas d'index par personne concernée. L'effacement d'un document ne retire pas les faits dérivés qui ont été fusionnés dans des entités partagées. [Supposé d'après le modèle : à vérifier sur `delete_document_graph`] |
| Journal d'audit | **Non** | Il existe un journal de requêtes : méthode, durée, appels lents, `CORTEXDB_TRACE_FILE` en JSON lines. Il ne consigne **pas** l'identité de la clé. [Vérifié : `pkg/rpcserver/trace.go`, aucune référence à `key` ou `caller`] |
| Départ d'un employé | Hors de portée | Rien. Avec un serveur par cerveau, c'est une opération d'exploitation : retirer la clé, archiver la base. |
| Traitement interne (Ollama) | **Couvert** | `CORTEXDB_LLM_BASE_URL`, `CORTEXDB_EMBED_BASE_URL`, `CORTEXDB_RERANK_BASE_URL`, points d'accès compatibles OpenAI. [Vérifié : `cmd/cortexdb-mcp-stdio/llm.go`, `embedder.go`, `reranker.go`] |
| Masquage des données sensibles | Partiel | À la capture de session seulement : blocs `<private>…</private>` et motifs de secrets (clés, jetons, `password=`). Pas de masquage de données personnelles (noms, e-mails, AVS). [Vérifié : `cmd/cortexdb-mcp-stdio/scrub.go`] |
| Relecture humaine avant entrée | Partiel | `--capture-session --dry-run` produit des propositions sans écrire. `side_graph_write/read/drop` offre un graphe de côté. Pas de file de relecture, pas de statut « proposé / accepté / refusé ». [Vérifié : `capture_session.go` ligne 27 ; outils `side_graph_*`. Usage de side_graph comme file : à vérifier] |
| Historique des faits | Couvert | Faits bitemporels : `valid_from`, `valid_to`, `recorded_at`, `retracted_at`. Une rétractation garde la ligne. Utile pour l'audit des connaissances, pas pour l'audit des accès. [Vérifié : `pkg/graph/temporal.go`] |
| Transport chiffré | Par proxy | gRPC en clair. `CORTEXDB_REMOTE` accepte `https://host` derrière un proxy TLS. [Vérifié : README, `cmd/cortexdb-grpc/main.go`] |

### 1.4 Ajout de documents par les employés

| Besoin | État | Détail |
| :- | :- | :- |
| Skill « ajouter un fichier » | Partiel | `ingest_document` prend du texte. Pas d'extraction PDF ni Word dans le dépôt. [Vérifié : signature de `ToolIngestDocumentRequest` ; absence de dépendance PDF dans `go.mod`, à vérifier] |
| Choix du graphe cible | Hors de portée | Un client, un cerveau. |
| Détection de doublons | Partiel | Le `document_id` est fourni par l'appelant. Même id = réécriture. Pas de hachage de contenu. `dedup.go` concerne les mémoires. [Vérifié : `graphrag_tool_ingest.go` ; `cmd/cortexdb-mcp-stdio/dedup.go`] |
| Mise à jour quand la source change | Hors de portée côté plugin | Le paquet `connector` a un `Watcher` CDC pour des bases SQL, avec un « privacy gate ». Rien pour des fichiers ou des pages web. [Vérifié : `SKILL.md` lignes 941-955] |
| Plugin Chrome | Hors de portée | Rien. Un plugin Obsidian existe, c'est un précédent de client externe. [Vérifié : `plugins/obsidian-cortexdb`] |
| Serveur HTTP REST | Couvert | `/v1` REST à côté de gRPC, même politique de clés. Base possible pour un plugin Chrome. [Vérifié : `pkg/httpapi`, `startHTTP` dans `main.go`] |

### 1.5 Sans clé d'API

| Besoin | État | Détail |
| :- | :- | :- |
| Extraction, embeddings, reclassement par Ollama | Couvert | Voir 1.3. Le code accepte une clé vide. |
| Fonction sans aucun modèle | Couvert | Mode lexical, extraction heuristique. Qualité moyenne (entités « Le », « Voir »). [Observé en usage réel] |
| Modèle par l'abonnement Claude Code | Couvert, **à cadrer** | Le mod `cortexdb-live` appelle haiku par `$.model.complete({ model: 'haiku' })`, sur l'accès modèle de Claude Code. Aucune clé d'API. Les mods sont une fonction en accès anticipé. [Vérifié : `plugins/cortexdb-live/hooks/register.tsx` lignes 29, 198, 376 ; README ligne 65]. `claude -p` fonctionne aussi côté serveur sous abonnement, avec un siège « bot » Team ou Enterprise et un jeton `claude setup-token` (un an, prévu pour « CI pipelines, scripts »). Le quota est celui d'un siège. Voir 3.8. [Vérifié : pages « Legal and compliance » et « Authentication », code.claude.com, 18 octobre 2026] |
| Mesure de la qualité du rappel | Couvert | `--recall-eval golden.jsonl` : requête, ids attendus, ids interdits, cas négatifs. [Vérifié : `cmd/cortexdb-mcp-stdio/recall_eval.go`] |

### 1.6 Maturité et pérennité

- 725 commits, 453 étoiles, 24 forks, licence MIT, Go 1.25. [Vérifié : page GitHub du 18 octobre 2026, `go.mod`]
- 178 versions du 7 août 2025 (0.1.0) au 17 octobre 2026 (2.132.0), soit 14 mois. Plusieurs versions par jour mi-octobre. [Vérifié : `CHANGELOG.md`]
- Auteurs des 725 commits (clone complet, `git log --format='%an'`) : Liang Li 652, Rudy Mizrahi Celekli 52, « Claude » 20, Lauro Foletti 1. Un auteur principal, un second contributeur régulier. [Vérifié : clone du 18 octobre 2026, v2.132.1]
- Le code lu est soigné : commentaires de conception, tests de couverture des routes, comparaison des secrets en temps constant. [Vérifié : `pkg/httpapi/access.go`, `pkg/authz`]
- Risque : **dépendance à une personne** et **rythme de changement élevé**. Un fork doit rester petit pour suivre.

---

## 2. Décision : étendre, compléter, ou partir sur autre chose

### 2.1 Options comparées

| Critère | A. CortexDB étendu (fédération dans le plugin) | B. CortexDB moteur + service maison | C. Graphiti (Zep) | D. Mem0 OSS | E. Cognee | F. Base de graphes nue (Neo4j CE, Memgraph, Oxigraph) + service maison |
| :- | :- | :- | :- | :- | :- | :- |
| Conformité (droits, audit, classification) | À construire dans le fork | À construire dans le service, **indépendant de l'amont** | Absente. Multi-tenant par `group_id` seulement [Supposé] | Absente dans l'OSS. L'MCP officiel exige une clé cloud [Recherche web, 2026] | Partielle (API + jeton) [Docs Cognee] | À construire, tout |
| Maturité, pérennité | Un auteur, très actif | Idem pour le moteur, le service est à nous | Société derrière (Zep), MCP « expérimental » [Recherche web] | Société (Mem0), OSS mince | Jeune | Très mature (Neo4j) |
| Effort de développement | Moyen, mais **fork divergent** à suivre | Moyen, **fork minimal** | Élevé : refaire les hooks et les skills | Élevé : idem, pas de graphe de documents | Élevé | Très élevé : RAG, graphe, MCP, hooks à écrire |
| Exploitation | N serveurs | N serveurs + 1 service | Neo4j ou FalkorDB + Python + modèle obligatoire | Qdrant + Neo4j + Python | LanceDB + Kuzu + Python | 1 base + service |
| Compatibilité agents de code (MCP, hooks) | Native | Native (le service parle MCP) | MCP expérimental, pas de hooks | MCP communautaire | MCP, pas de hooks | À écrire |
| Ollama seul | Oui, et mode lexical sans modèle | Oui | Modèle obligatoire, Ollama non officiel [Recherche web] | Oui pour embeddings, LLM nécessaire | Oui, mais ≥ 32B conseillé, JSON strict fragile [Recherche web] | Oui |

Sources web consultées le 18 octobre 2026 : [FalkorDB, Graphiti MCP Server](https://docs.falkordb.com/agentic-memory/graphiti-mcp-server.html), [Graphiti FastMCP (fork communautaire, Ollama)](https://glama.ai/mcp/servers/donbr/graphiti-fastmcp), [tensakulabs/mem0-mcp (auto-hébergé)](https://glama.ai/mcp/servers/tensakulabs/mem0-mcp), [Cognee MCP self-hosted](https://docs.cognee.ai/cognee-mcp/mcp-cloud-connection.md), [Self-hosting Cognee sur Ollama](https://glukhov.org/post/2025/12/selfhosting-cognee-quickstart-llms-comparison). LightRAG : accepte Ollama [Supposé d'après connaissance générale] ; pas de serveur MCP officiel connu [À vérifier].

### 2.2 Décision

**Option B, avec une méthode de l'option A.** CortexDB reste le moteur. Une couche mince Thalamus porte ce que l'amont n'a pas. Chaque fonction de la couche qui a un sens général est proposée en amont comme petite PR. Quand la PR est acceptée, la couche s'amincit.

Pourquoi pas A seul : un fork du plugin qui porte la fédération et les droits diverge vite d'un amont qui publie plusieurs versions par jour. Le coût de rebase dépasse le gain.

Pourquoi pas C à F : aucune n'apporte la conformité. Toutes exigent de réécrire l'intégration Claude Code et Codex. C, D, E exigent un modèle pour chaque opération, donc une dépendance au GPU pour le simple rappel.

Pourquoi pas « un autre projet » : le bon projet est bien **une couche de gestion des connaissances au-dessus d'un moteur**. Le moteur peut changer plus tard si la couche garde un adaptateur. Le moteur le plus proche du besoin aujourd'hui est CortexDB.

### 2.3 Petites PR candidates vers l'amont, dans l'ordre

| Ordre | PR | Taille | Ce qu'elle débloque |
| :- | :- | :- | :- |
| 1 | Journal de requêtes avec `key_id`, nom de l'outil, ids touchés, dans `CORTEXDB_TRACE_FILE` | Petite (`pkg/rpcserver/trace.go`) | Un début d'audit par identité |
| 2 | Métadonnées au niveau du document dans `ingest_document` (sur `core.Document` et le nœud `document`), et retour de ces métadonnées dans les résultats | Petite à moyenne | Classification, source, date, hachage |
| 3 | Listes dans `Scope` : `namespaces: [...]`, `collections: [...]` | Moyenne (`pkg/authz`, tests) | Une clé lit plusieurs périmètres |
| 4 | Outils du toolbox autorisés pour une clé confinée quand l'outil déclare ses champs de portée (`collection`, `namespace`) et que la requête les porte | Moyenne (`pkg/authz/tools.go`, `pkg/httpapi/access.go`, `pkg/rpcserver/auth.go`) | Le plugin MCP fonctionne avec des clés confinées |
| 5 | `expires_at` sur documents et mémoires, et un balayage `--sweep-expired` | Petite | Conservation |
| 6 | Statut `proposed/accepted/rejected` sur `knowledge_save` et `memory_save`, avec `reviewed_by` | Moyenne | Relecture humaine native |
| 7 | Hook `SessionEnd` : aucun fichier écrit quand la capture est désactivée | Très petite | Corrige les 295 fichiers par matinée |
| 8 | Hachage de contenu dans `ingest_document` pour refuser un doublon exact | Petite | Doublons |

La PR 1, 2, 7 sont des bonnes premières contributions : petites, sans débat de conception. Les PR 3 et 4 touchent la sécurité : les proposer **avec** une discussion (issue) avant le code.

État au 18 octobre 2026 : onze correctifs sont écrits et testés sur la branche locale `proto/granular-cortex` du clone (`cortexdb/`), un commit par correctif, aucun push. Ordre des commits : 7 (`138edb7`), 1 (`b7beb86`), 2 (`3be4302`), 4 (`8c0d1ff`), 3 (`f21e4d7`), 8 (`7edc498`), 5 (`29c7d1e`), 6 (`4024f93`), 9 (`1d0fdf1`, expansion de graphe confinée à la collection), 10 (`c429b74`, un outil adressé par identifiant ne touche que des lignes dans la portée), 11 (`1256e4a`, toutes les graphies d'un champ sont vérifiées). Les correctifs 10 et 11 ferment deux trous que l'essai de bout en bout (`tests/e2e_shared_brain.sh`) a révélés dans le 4 : sans eux, une clé confinée réécrivait un document d'une autre collection en passant sa propre collection avec l'identifiant du document. Le détail est dans `cortexdb/CHANGELOG.md`, section `Unreleased`. Le plan de publication est dans `docs/plan_publication.md`.

---

## 3. Architecture proposée : Thalamus

### 3.1 Composants

```
 Postes de travail                          Serveurs internes (réseau de confiance, TLS par proxy)
 ┌─────────────────────────────┐            ┌────────────────────────────────────────────────────┐
 │ Claude Code / Codex         │            │ th-gateway  (Go, service maison)                   │
 │  ├ plugin thalamus   │  MCP/HTTPS │  ├ /mcp         : outils fédérés (recall, add…)     │
 │  │  ├ hooks (recall, files) ├───────────►│  ├ /v1/brains   : cerveaux visibles par l'appelant  │
 │  │  ├ skill /th-add         │            │  ├ /v1/ingest   : ajout de document (skill, Chrome) │
 │  │  └ skill /th-review      │            │  ├ /v1/review   : file de relecture                 │
 │  └ cerveau local SQLite     │            │  ├ rights.yaml  : droits, groupes, classification   │
 │     (user/<emp>/<projet>)   │            │  └ audit.jsonl  : qui, quoi, quand, quel cerveau    │
 ├─────────────────────────────┤            │          │ gRPC + clé par cerveau                   │
 │ Chrome                      │  HTTPS     │  ┌───────┴────────┬──────────────┬───────────────┐  │
 │  └ extension thalamus├───────────►│  │cortexdb-grpc   │cortexdb-grpc │cortexdb-grpc  │  │
 └─────────────────────────────┘            │  │team/<projet-A> │team/<proj-B> │org/reference  │  │
                                            │  └───────┬────────┴──────┬───────┴──────┬────────┘  │
 Annuaire (LDAP / Entra ID)                 │          └───────────────┴──────────────┘           │
  groupes → rights.yaml                     │             PostgreSQL + pgvector (1 schéma/cerveau)│
                                            │  Ollama (extraction, embeddings, reclassement)      │
                                            │  th-sync (cron) : sources suivies, hachages, retraits│
                                            └────────────────────────────────────────────────────┘
```

Rôle de chaque composant :

- **cortexdb-grpc, un processus par cerveau partagé.** Base PostgreSQL, un schéma par cerveau. Fichier de clés par serveur : une clé pour `th-gateway` (lecture-écriture), aucune clé distribuée aux employés. [Backend PostgreSQL vérifié : `pkg/core/store_postgres.go`]
- **th-gateway.** Service Go, importe les paquets `cortexdb` comme bibliothèque et parle gRPC aux serveurs. Il est le seul client des serveurs. Il authentifie l'employé (jeton OIDC de l'annuaire, ou clé personnelle émise par l'administrateur), lit `rights.yaml`, fédère, journalise. Il expose un serveur MCP (HTTP) et une API REST.
- **Plugin `thalamus`** pour Claude Code et Codex. Il remplace les hooks du plugin CortexDB (désactivés par `cortexdb-recall --disable` et `cortexdb-session-end --disable` [Vérifié : README ligne 65]). Ses hooks appellent `th-gateway`. Le cerveau local SQLite (notes personnelles par projet) reste un fichier dans `.cortexdb/` du dépôt, non versionné.
- **Extension Chrome** (Manifest V3). Elle lit `/v1/brains?writable=1`, envoie la page ou la sélection à `/v1/ingest`.
- **th-sync.** Tâche planifiée. Elle relit la table des sources suivies, compare les hachages, ré-ingère ou retire.
- **Ollama.** Un serveur interne. Voir section 3.6.

### 3.2 Modèle de données : cerveaux

Nom d'un cerveau : `<niveau>/<identifiant>`.

| Niveau | Exemple | Hébergement | Lecture | Écriture | Curateur |
| :- | :- | :- | :- | :- | :- |
| `org/reference` | glossaire, procédures, organigramme | serveur | tous les employés | équipe documentation | responsable KM |
| `dept/<nom>` | `dept/finance` | serveur | membres du département | membres, après relecture | chef de département |
| `team/<projet>` | `team/acme-migration` | serveur | membres du projet | membres, après relecture | chef de projet ou délégué |
| `client/<nom>` | `client/acme` | serveur | membres des projets de ce client | relecture obligatoire | responsable du compte |
| `user/<employé>` | `user/lfoletti` | serveur | l'employé seul | l'employé, direct | l'employé |
| `user/<employé>/<projet>` | `user/lfoletti/acme-migration` | **poste de travail** (SQLite) | l'employé seul | l'employé, direct | l'employé |

Règle sur les niveaux : `client/` est nécessaire quand un contrat impose un cloisonnement par client. Un fait sur un client va dans `client/<nom>` ou `team/<projet>`, jamais dans `user/<employé>`. Voir 7.1.

Un cerveau est déclaré dans `rights.yaml` :

```yaml
brains:
  team/acme-migration:
    addr: cortexdb-acme.internal:47821
    key_ref: vault:th/team-acme-migration
    classification_max: confidential     # rien de plus sensible n'y entre
    retention_default: P5Y                # ISO 8601
    curators: [group:proj-acme-leads]
    readers:  [group:proj-acme]
    writers:  [group:proj-acme]          # écriture = proposition, voir 3.5
    client: acme                          # cloisonnement contractuel
```

### 3.3 Modèle de données : documents et faits

Métadonnées obligatoires d'un document, portées dans `metadata` de `ingest_document` (copiées sur chaque passage aujourd'hui, sur le document après la PR 2) :

```json
{
  "document_id": "sha256:3f9a…",           
  "source_kind": "file | web | session",
  "source_uri": "smb://…/spec.docx | https://… | session:2026-10-18T09:12Z",
  "source_hash": "sha256:…",
  "consulted_at": "2026-10-18T09:12:00+02:00",
  "added_by": "lfoletti",
  "classification": "public | internal | confidential | personal-data",
  "personal_data": "none | employees | clients | third-parties",
  "retention_until": "2031-10-18",
  "legal_basis": "contract | legitimate-interest | consent | none",
  "brain": "team/acme-migration",
  "review_status": "accepted",
  "reviewed_by": "mdupont",
  "language": "fr"
}
```

Un fait (entité ou relation) porte : `source_document_id`, `source_chunk`, `extracted_by` (`human | ollama:<modèle> | heuristic`), `review_status`, `reviewed_by`, `valid_from`, `valid_to`. Les deux derniers champs utilisent le modèle bitemporel de CortexDB. [Vérifié : `pkg/graph/temporal.go`]

### 3.4 Circuit d'un document

```
Employé                 Plugin / Chrome          th-gateway                      cortexdb-grpc        Ollama
  │ /th-add spec.docx         │                      │                               │                 │
  │   --brain team/acme ─────►│ extrait le texte     │                               │                 │
  │                           │ (pandoc, pdftotext)  │                               │                 │
  │                           │ POST /v1/ingest ────►│ 1 vérifie: writer du cerveau ?│                 │
  │                           │                      │ 2 hache, cherche doublon       │                 │
  │                           │                      │ 3 classe (règles + modèle) ───────────────────► │
  │◄── « classification       │◄─────────────────────│ 4 demande confirmation         │                 │
  │     proposée: confidential│                      │                               │                 │
  │     retention: P5Y ? »    │                      │                               │                 │
  │ ─── oui ─────────────────►│ ────────────────────►│ 5 ingest_document + metadata ─►│                 │
  │                           │                      │ 6 propose entités/relations ──────────────────► │
  │                           │                      │ 7 écrit dans la file review    │                 │
  │                           │                      │ 8 audit.jsonl: ingest, qui, où │                 │
  │◄── « 1 document, 14       │◄─────────────────────│                               │                 │
  │     passages, 9 faits     │                      │                               │                 │
  │     proposés à relire »   │                      │                               │                 │
```

Détails :

1. L'extraction du texte se fait **sur le poste** pour un fichier (pandoc pour Word et Markdown, `pdftotext` de poppler pour PDF, OCR hors périmètre du pilote). Pour une page web, l'extension envoie le texte lisible (Readability) ou la sélection.
2. La classification est une **proposition** : règles d'abord (chemin du fichier, domaine de l'URL, mots-clés, présence d'un numéro AVS ou d'un e-mail), modèle ensuite. L'employé confirme. La classification ne peut pas dépasser `classification_max` du cerveau.
3. Le doublon est détecté par `source_hash` et par `source_uri` canonique. Même URI, hachage différent : mise à jour (ré-ingestion sous le même `document_id`, version incrémentée). Même hachage, URI différente : lien vers le document existant, pas de nouvelle copie.
4. La source est inscrite dans la table `sources` de `th-sync` : `uri`, `hash`, `brain`, `document_id`, `last_checked`, `check_interval`, `status`. Un fichier retiré ou une page qui répond 404 ou 410 passe `status: gone`. Le curateur choisit : retrait du graphe (`delete_document_graph`) ou conservation marquée « source disparue ».

### 3.5 Circuit d'une question

```
Hook UserPromptSubmit          th-gateway                             cerveaux accessibles
  │ prompt, projet courant,       │                                        │
  │ identité ───────────────────► │ 1 plan de requête (mots-clés, entités) │
  │                               │   lexical, ou haiku via Claude Code    │
  │                               │ 2 cerveaux = readers(employé) ∩        │
  │                               │   pertinents(projet courant)           │
  │                               │ 3 fan-out parallèle, même requête ────►│ team/acme, user/lfoletti,
  │                               │                                        │ user/lfoletti/acme (local),
  │                               │◄───────────────────────────────────────│ org/reference
  │                               │ 4 filtre: classification injectable ?  │
  │                               │ 5 fusion RRF, dédoublonnage par hash   │
  │                               │ 6 contradictions → marquage            │
  │                               │ 7 audit.jsonl: recall, ids lus         │
  │◄── 3 extraits + provenance ───│                                        │
```

Règles de fusion et de classement :

- **Fusion** : Reciprocal Rank Fusion sur les listes de chaque cerveau, `k = 60`. Pas de comparaison des scores bruts entre cerveaux : un score lexical SQLite et un score vectoriel ne sont pas comparables.
- **Pondération par niveau** : `team/<projet courant>` ×1.0, `user/<employé>/<projet courant>` ×1.0, `org/reference` ×0.8, `user/<employé>` ×0.7, autres projets ×0.3. Les coefficients sont dans `rights.yaml` et ajustés par le jeu de référence (section 4.5).
- **Dédoublonnage** : deux extraits avec le même `source_hash` et le même `chunk_index` sont un seul extrait. La provenance liste alors tous les cerveaux qui le contiennent.
- **Provenance** : chaque extrait injecté porte une ligne `— [team/acme-migration] spec.docx §3, source 2026-09-02, ajouté par mdupont, interne`. Le format est fixe pour que l'agent et l'humain le lisent.
- **Contradictions** : deux faits sur la même entité et le même attribut, avec des valeurs différentes, dans deux cerveaux. Règle : le fait **accepté par un curateur** dans un cerveau partagé l'emporte sur un fait personnel. Entre deux cerveaux partagés, le plus récent par `valid_from` l'emporte. Dans tous les cas, l'extrait porte la mention `⚠ contredit par [user/lfoletti] note du 2026-10-01`. Le gateway crée une entrée dans la file de relecture du curateur.
- **Classification injectable** : voir 7.2. Par défaut, un extrait `confidential` ou `personal-data` est injecté seulement si le contrat avec le fournisseur de l'agent le permet, sinon il est remplacé par `— [team/acme] 1 extrait confidentiel non affiché, ouvrir avec /th-show <id>`.

Règle d'écriture d'un fait nouveau :

| Origine | Destination par défaut | Qui décide |
| :- | :- | :- |
| Capture de session (automatique) | `user/<employé>/<projet>` (local) | personne, c'est la zone brouillon |
| `/th-remember` explicite de l'employé | `user/<employé>` ou `user/<employé>/<projet>` au choix | l'employé |
| Fait extrait d'un document ajouté à `team/<projet>` | file de relecture de `team/<projet>` | le curateur |
| Promotion d'une note personnelle | file de relecture du cerveau cible | l'employé propose, le curateur accepte |

### 3.6 Circuit d'un fait extrait

```
Source (document, session)          th-gateway                                   Curateur
  │                                    │                                             │
  │ texte ──────────────────────────►  │ 1 Ollama extrait entités + relations        │
  │                                    │   (JSON contraint, schéma d'ontologie)      │
  │                                    │ 2 filtre: mots vides, longueur, stop-list   │
  │                                    │ 3 vérification: chaque fait cite un passage │
  │                                    │   qui contient ses termes, sinon rejet      │
  │                                    │ 4 écrit review/<brain>/<id>.json           │
  │                                    │   status: proposed                          │
  │                                    │ ─────────────── /th-review team/acme ─────► │ lit, pour chaque fait:
  │                                    │                                             │ accepte / corrige / refuse (motif)
  │                                    │ ◄─────────────────────────────────────────  │
  │                                    │ 5 accepté → knowledge_graph_upsert avec     │
  │                                    │   provenance, reviewed_by, valid_from       │
  │                                    │ 6 refusé → gardé avec motif, jamais reproposé│
  │                                    │ 7 audit.jsonl                               │
```

Points fermes :

- Rien d'extrait par un modèle n'entre dans un cerveau partagé sans un clic humain. C'est la leçon du premier usage réel (« l'étude a lu la réserve »).
- L'étape 3 est une vérification mécanique, pas un second modèle : les termes du fait doivent apparaître dans le passage cité. Elle élimine une part des hallucinations à coût nul. [Outil `verify_claims` de CortexDB : à vérifier s'il fait cela]
- La file de relecture est un répertoire de fichiers JSON dans un dépôt Git interne par cerveau, ou une table PostgreSQL de `th-gateway`. Le dépôt Git a l'avantage de la revue par PR, familière aux équipes de code. Pour le pilote : Git.
- Un fait refusé reste visible du gateway avec son motif. Si Ollama le repropose, le gateway le filtre.

### 3.7 Modèles Ollama

| Tâche | Modèle plausible | Mémoire (quantisation Q4) | Perte face à un modèle de pointe |
| :- | :- | :- | :- |
| Extraction d'entités et relations, français | `qwen3:14b` pour le pilote, `qwen3:32b` ou `gemma3:27b` si le matériel le permet | 10 Go pour 14B, 20 Go pour 32B | Rappel des relations plus faible, plus de relations inventées, nuances juridiques en français moins bien lues. Ordre de grandeur : 10 à 25 points de F1 en moins sur l'extraction de relations [Supposé, d'après l'état de l'art ; à mesurer sur le jeu de référence] |
| Embeddings multilingues | `bge-m3` (1024 dim) ou `qwen3-embedding:0.6b` | 1 à 2 Go | Faible perte en recherche de passages en français. Les modèles ouverts sont proches des modèles propriétaires sur cette tâche. [Supposé] |
| Reclassement | `bge-reranker-v2-m3` | 1 Go | Faible perte. Ollama ne sert pas nativement les rerankers : un conteneur `text-embeddings-inference` (Hugging Face) ou vLLM expose le point d'accès compatible attendu par `CORTEXDB_RERANK_BASE_URL`. [À vérifier : format attendu par `reranker.go`] |
| Classification de documents | le même modèle que l'extraction, avec un prompt fermé (4 classes) | partagé | Faible perte sur 4 classes avec des règles en amont. |
| Planification de requête, capture de session, relecture assistée, extraction de documents `public` et `internal` | haiku ou sonnet via `claude -p` (siège de l'employé ou siège bot du gateway, abonnement), ou `qwen3:8b` sur Ollama | 6 Go | Voir 3.8 pour le cadre contractuel et 7.2 pour la sortie de données. |

Matériel : un serveur avec un GPU de 24 Go (RTX 4090 ou 5090) couvre 14B + embeddings + reranker pour une équipe de 10 à 20 personnes. Un GPU de 48 Go (L40S) ou un Mac Studio 128 Go couvre 32B. Prix d'ordre : 3 000 à 5 000 CHF pour le premier, 10 000 à 15 000 CHF pour le second. [Supposé, prix du marché à vérifier]

Le mode lexical sans modèle reste le repli : le rappel fonctionne, l'extraction heuristique est désactivée (elle crée des entités parasites), seules les entités validées par un humain existent.

### 3.8 Où `claude -p` sous abonnement est utilisable

Sources : pages « Legal and compliance » et « Authentication » de Claude Code, code.claude.com, lues le 18 octobre 2026. [Vérifié]

Ce que les pages disent :

1. « Advertised usage limits for Pro and Max plans assume ordinary, individual usage of Claude Code and the Agent SDK. » L'usage de `claude -p` sous abonnement est prévu.
2. `claude setup-token` crée un jeton OAuth d'un an « for CI pipelines, scripts, or other environments where interactive browser login isn't available ». Le jeton « requires a Pro, Max, Team, or Enterprise plan ». Il se passe dans `CLAUDE_CODE_OAUTH_TOKEN`. Un usage **sans personne au clavier** est donc documenté par Anthropic.
3. Les interdictions visent deux cas : un tiers qui route les requêtes de ses utilisateurs par des identifiants Free, Pro ou Max, et un client qui revend ou intermédie l'usage pour les utilisateurs finaux de son produit. Un pipeline interne de l'entreprise sur son propre siège n'est ni l'un ni l'autre.
4. `claude --bare -p` ne lit pas `CLAUDE_CODE_OAUTH_TOKEN` et exige une clé d'API. Le gateway n'utilise donc pas `--bare`. Il passe le contexte par `--mcp-config`, `--append-system-prompt` et `--json-schema`.

Règles pour Thalamus :

| Usage | Cadre | Décision |
| :- | :- | :- |
| Planification de requête, capture de session, dans la session Claude Code de l'employé | Usage individuel ordinaire, siège de l'employé | **Admis.** C'est le modèle de `cortexdb-live`. |
| Relecture assistée : l'employé lance `/th-review`, le modèle de sa session résume les faits proposés | Siège de l'employé | **Admis.** |
| Siège « bot » Team ou Enterprise, jeton `setup-token` dans le gateway, pour l'extraction d'entités, la classification et le résumé des documents `public` et `internal` | Usage documenté (CI, scripts). Quota d'un siège. | **Admis.** Pratique courante. Le gateway surveille le quota et bascule sur Ollama à l'approche du plafond. |
| Le même siège bot pour les documents `confidential` et `personal-data` | Licence : admis. Données : dépend du contrat Enterprise (DPA, pas d'entraînement, localisation) | **À trancher par l'entreprise.** Voir 7.2 et question 1 de 7.4. Par défaut : Ollama. |
| Ingestion de masse (centaines de documents, réindexation complète, embeddings) | Admis, mais hors de l'échelle d'un siège | **Ollama.** Les embeddings et le reclassement sont toujours sur Ollama : Claude Code ne fournit pas d'embeddings. |

Conséquence sur l'architecture : le gateway porte **un** jeton de siège bot, stocké dans le coffre, avec un compteur d'usage. Il envoie à `claude -p` les tâches qui méritent un modèle de pointe et dont la classification le permet. Il envoie à Ollama les tâches de masse, les embeddings, le reclassement, et tout document dont la classification interdit la sortie. Un deuxième siège bot est une décision de budget, pas de licence.

Format d'appel depuis le gateway :

```bash
CLAUDE_CODE_OAUTH_TOKEN="$(vault read th/bot-seat)" \
claude -p "Extrais les entités et relations du document ci-joint." \
  --append-system-prompt-file prompts/extract.md \
  --output-format json --json-schema schemas/facts.json \
  --permission-mode dontAsk --permission-prompts none \
  --max-budget-usd 0.50 < document.txt
```

`--json-schema` donne un JSON contraint par le schéma d'ontologie. `--max-budget-usd` borne chaque appel. Le mode `dontAsk` refuse tout outil non listé. [Options vérifiées : page « Run Claude Code programmatically », code.claude.com]

---

### 3.9 Granularité et liens entre cerveaux

Réponse courte : l'unité de granularité est le **cerveau**. Les cerveaux ne portent **aucune arête physique** entre eux. Les liens se font par identifiants communs et par nœuds de référence, et le gateway les résout à la lecture. L'écriture reste explicite (7.3).

#### Granularité : trois voies, et laquelle pour quel niveau

| Voie | Étanchéité | Coût d'exploitation | État |
| :- | :- | :- | :- |
| Un serveur par cerveau | processus | N processus, N ports, N clés | amont |
| Une **collection** par cerveau, un serveur partagé | clé confinée, vérifiée sur chaque outil | 1 processus par niveau de sensibilité | branche locale (correctifs 3 et 4) |
| SQLite local | fichier de l'employé | nul | amont |

Choix par niveau :

| Niveau | Voie | Raison |
| :- | :- | :- |
| `org/`, `dept/`, `team/` | collection dans **un** serveur partagé | une clé par employé liste ses collections ; un seul processus à exploiter |
| `client/<nom>` | **un serveur par client** | les nœuds d'entité sont partagés entre les collections d'un serveur (voir ci-dessous) ; un cloisonnement contractuel exige le processus |
| `user/<e>` | collection `user-<e>` sur le serveur partagé | lisible par l'employé seul, par clé |
| `user/<e>/<p>` | SQLite sur le poste | aucune exploitation, aucune fuite réseau |

Limite vérifiée : un nœud d'entité a pour identifiant son nom normalisé (`pkg/cortexdb/graphrag_helpers.go`, `graphEntityNodeID`). Il ne porte pas de collection. Dans un serveur partagé, « Acme » de l'équipe A et « Acme » de l'équipe B sont le **même nœud**, avec les relations des deux équipes. Les passages (chunks) sont filtrés par collection à la lecture (`chunksInScope`), les entités et leurs relations ne le sont pas. Pour un glossaire d'organisation, c'est utile. Pour un client sous contrat, c'est une fuite : d'où un serveur par client. Correctif candidat 9 : les nœuds d'entité portent la liste des collections qui les mentionnent, et l'expansion filtre dessus. Avec ce correctif, `client/` pourra rejoindre le serveur partagé.

Seconde limite : une clé confinée sur `user_id` **et** sur `collections` est refusée partout, parce que `knowledge_search` ne déclare pas `user_id` et `memory_search` ne déclare pas `collection`. La règle « champ non déclaré = refus » est volontaire et reste. Conséquence : la clé de l'employé ne sort pas du gateway. Le gateway tient une clé par cerveau et fait l'éventail (3.5).

#### Liens d'un graphe à l'autre : trois mécanismes

Aucun cerveau ne stocke une arête vers un autre cerveau. SQLite ne le permet pas, et un droit par cerveau serait impossible à faire respecter sur une telle arête. Les liens sont logiques, résolus par le gateway, par coût croissant.

**A. Identité par convention.** La même chose a le même identifiant dans tous les cerveaux :

| Objet | Identifiant | État |
| :- | :- | :- |
| document | `doc:sha256:<contenu>` | le hash est calculé et stocké (`content_sha256`, correctif 8) ; le gateway l'utilise comme `document_id` |
| passage | `chunk:<document_id>:<index>` | amont |
| entité | nom normalisé, canonisé par l'ontologie (alias) | amont ; à durcir avec `ontology_*` |

Un lien inter-graphe est alors une **égalité d'identifiants**. Le gateway fusionne par identifiant (RRF, 3.5) et la provenance liste tous les cerveaux qui contiennent l'objet. Rien à stocker. Ce mécanisme couvre la plupart des cas.

**B. Nœud de référence.** Quand un cerveau doit pointer vers un objet d'un autre cerveau, il crée un nœud local :

```
id:     ref:team/acme-migration#doc:sha256:3f9a…
type:   external_ref
props:  brain, node_id, snapshot_at, label (titre seul, jamais le contenu)
```

et une arête locale vers lui : `promoted_to`, `contradicts`, `see_also`. Le gateway résout la référence si l'employé est lecteur du cerveau cible. Sinon il affiche « 1 référence vers [client/acme], non accessible ». Usages : la promotion (4.3), les contradictions (3.5), les renvois d'une note personnelle vers un document d'équipe.

**C. Index de fédération dans th-gateway.** Une table `canonical_id → [cerveau, dernière vue]`, remplie à l'acceptation d'un fait et à l'ingestion. Elle répond à « où Acme est-il connu ? » sans interroger tous les cerveaux. Étape 3 du plan, facultative pour le pilote.

#### Ce qui n'est pas proposé

- Un graphe global unique : il supprime l'isolation que les niveaux ont pour but de donner.
- Des arêtes physiques entre bases : aucun droit ne peut être vérifié dessus.
- Une écriture « sans choix » : la lecture est fédérée, l'écriture reste explicite (7.3).

---

## 4. Gestion des connaissances : règles

### 4.1 Création d'un cerveau

1. Un chef de projet, de département ou de compte demande un cerveau. Il nomme un curateur et un suppléant.
2. L'administrateur ajoute l'entrée dans `rights.yaml`, crée le schéma PostgreSQL, démarre le serveur, émet la clé du gateway.
3. Un cerveau `user/<employé>` est créé à l'arrivée de l'employé. Un cerveau `user/<employé>/<projet>` est créé par le plugin au premier usage dans le dépôt.

### 4.2 Curation

- Le curateur relit la file au moins une fois par semaine. Un fait proposé depuis plus de 30 jours est refusé automatiquement avec le motif `expired-review`.
- Le curateur garde le **graphe de vérité** du projet dans un fichier versionné `knowledge/<projet>.yaml` (entités, relations, sources). Le cerveau est reconstruit depuis ce fichier et les documents suivis. C'est ce qui a bien marché en usage réel : identifiants stables, aucune réécriture d'une section inchangée, avertissement pour tout fichier non classé.
- Les documents ajoutés par les employés sont indexés tels quels. Seuls les **faits** passent par la relecture.

### 4.3 Promotion d'un niveau à l'autre

| De | Vers | Procédure |
| :- | :- | :- |
| `user/<e>/<p>` | `team/<p>` | `/th-promote <id>`. Le fait part dans la file de `team/<p>`. Le curateur accepte. Le fait original reste chez l'employé avec `promoted_to`. |
| `team/<p>` | `dept/`, `org/reference` | Le curateur du cerveau source propose. Le curateur du cerveau cible accepte. |
| `team/<p>` | `client/<c>` | Idem, avec vérification de la classification et du contrat. |
| Tout | `user/<e>` | Jamais de promotion automatique d'un fait de projet vers la mémoire transverse d'un employé. L'employé copie à la main. Voir 7.1. |

### 4.4 Retrait d'un fait faux

1. Toute personne qui lit le cerveau peut marquer un fait `disputed` avec un motif (`/th-dispute <id> "raison"`).
2. Le fait reste rappelé, avec la mention `⚠ contesté`.
3. Le curateur tranche : rétractation (`retracted_at` posé, le fait reste dans l'historique, il n'est plus rappelé) ou confirmation.
4. Si le fait vient d'un document, le curateur vérifie les autres faits du même document.

### 4.5 Mesure de la qualité du rappel

Oui, un jeu de questions de référence est nécessaire. Sans lui, aucun changement de modèle, de pondération ou de version de CortexDB n'est jugeable.

- Format : le `golden.jsonl` de `--recall-eval` [Vérifié : `recall_eval.go`] : `query`, `expected` (ids qui doivent sortir), `forbidden` (ids périmés qui ne doivent pas sortir), cas négatifs.
- Taille : 30 à 50 questions par cerveau d'équipe, écrites par le curateur à partir de vraies questions posées aux agents. 10 questions négatives.
- Mesures : `hit@3` (l'attendu est dans les 3 extraits injectés), taux d'interdits dans le top 3, taux de bruit sur les négatifs. Seuils du pilote en section 5.
- Fréquence : à chaque changement de modèle, de version de CortexDB, ou de pondération. Un rapport mensuel.
- Mesure de l'extraction : sur 20 documents, le curateur compte les faits proposés, acceptés tels quels, corrigés, refusés. Le taux d'acceptation sans correction est l'indicateur.

---

## 5. Plan de mise en œuvre

### Étape 0 : socle, sans fork (3 semaines, 1 personne)

Livrables :

- 1 serveur `cortexdb-grpc` pour `team/<pilote>`, PostgreSQL, proxy TLS, fichier de clés avec 1 clé gateway. Déploiement d'après `deploy/docker` et `deploy/systemd`. [Vérifié : répertoires présents]
- `th-gateway` v0 : MCP + REST, 1 cerveau, authentification par jeton OIDC ou clé personnelle, `rights.yaml`, `audit.jsonl` (qui, quoi, quand, quel cerveau, quels ids).
- Plugin `thalamus` v0 : hook de rappel, skill `/th-add` (Markdown, PDF, Word), skill `/th-review`.
- Ollama avec `qwen3:14b` et `bge-m3`. Mode lexical en repli.
- Golden set de 30 questions.

Risques : format exact des RPC gRPC à suivre à chaque version de l'amont (le proto change). Mitigation : épingler la version de CortexDB, mettre à jour tous les mois.

### Étape 1 : pilote sur un projet d'équipe (6 semaines, 1 personne + 1 curateur à 10 %)

Livrables : 5 à 10 employés, 50 à 200 documents, file de relecture Git, rapport de qualité.

Critères de succès mesurables :

| Mesure | Seuil |
| :- | :- |
| `hit@3` sur le golden set | ≥ 0,70 |
| Faits interdits (périmés) dans le top 3 | ≤ 5 % |
| Extraits hors sujet injectés sur des prompts négatifs | ≤ 20 % |
| Faits proposés par Ollama acceptés sans correction | ≥ 50 % |
| Documents ajoutés par des employés (hors curateur) | ≥ 30 |
| Entrées d'audit incomplètes (sans identité ou sans ids) | 0 |
| Délai médian de relecture d'un fait | ≤ 7 jours |
| Retour des utilisateurs « le rappel m'a aidé » (sondage hebdomadaire) | ≥ 60 % |

Si `hit@3` reste sous 0,55 après 3 semaines de réglage, arrêt du pilote et révision du moteur.

### Étape 2 : premières PR vers l'amont (en parallèle de l'étape 1, 2 semaines cumulées)

PR 1 (identité dans le trace), PR 2 (métadonnées du document), PR 7 (fichier de capture). Ouvrir une issue de conception pour les PR 3 et 4.

Risque : l'auteur refuse ou réécrit autrement. Mitigation : petites PR, tests inclus, style du dépôt respecté. Si refus, la couche garde la fonction.

### Étape 3 : fédération et cerveaux personnels (6 semaines, 1 personne)

Livrables : 3 cerveaux (team, user, user/projet local), fan-out, RRF, provenance, contradictions, `/th-promote`, `/th-dispute`. Ajout de `org/reference`.

Risque : bruit du cerveau personnel transverse dans le rappel (observé en usage réel avec la base globale). Mitigation : pondération 0,7 et golden set étendu.

### Étape 4 : ajout par Chrome et synchronisation (4 semaines, 1 personne)

Livrables : extension Chrome MV3 (sélection ou page, choix du cerveau parmi les cerveaux autorisés, classification à l'ajout), `th-sync` (table des sources, hachages, retraits), déploiement de l'extension par la politique d'entreprise de Chrome.

Risque : pages derrière authentification, contenu dynamique. Mitigation : l'extension envoie le DOM rendu, pas l'URL seule.

### Étape 5 : durcissement conformité (4 semaines, 1 personne + juriste)

Livrables : balayage de conservation, procédure de départ d'un employé (clé révoquée, `user/<e>` archivé chiffré 90 jours puis effacé, `user/<e>/<p>` locaux effacés par la procédure de restitution du poste), registre des traitements, procédure de droit d'accès et d'effacement (recherche par nom dans tous les cerveaux, rapport, effacement avec rétractation des faits dérivés), masquage des données personnelles à l'ingestion (motifs AVS, e-mail, téléphone, puis modèle).

### Étape 6 : extension à 3 à 5 équipes (continu)

Un cerveau par équipe, un curateur par cerveau, revue trimestrielle des pondérations et des golden sets.

Effort total avant l'extension : environ 27 semaines-personne, sur 6 à 8 mois calendaires avec une personne.

---

## 6. Risques

| Risque | Gravité | Mitigation |
| :- | :- | :- |
| Dépendance à un auteur unique, rythme élevé | Haute | Couche mince, adaptateur moteur, version épinglée, PR en amont, plan B : geler une version et maintenir un fork minimal |
| Clé confinée inutilisable avec le toolbox | Haute aujourd'hui | Un serveur par cerveau et une seule clé gateway. PR 3 et 4 à moyen terme |
| Extraction fausse acceptée sans lecture | Haute | File de relecture obligatoire, vérification mécanique des termes, délai de relecture mesuré |
| Fuite par le rappel vers le fournisseur de l'agent | Haute | Filtre de classification injectable, voir 7.2 |
| Bruit du rappel (hors sujet) | Moyenne | Golden set, pondérations, cas négatifs, seuil d'arrêt |
| Transport gRPC en clair | Moyenne | Proxy TLS obligatoire, réseau de confiance, serveurs sur `127.0.0.1` derrière le proxy |
| Coût matériel sous-estimé | Moyenne | Pilote sur 14B, mesure, décision 32B après |
| Charge de curation | Moyenne | 10 % d'un curateur par cerveau d'équipe, refus automatique à 30 jours |
| Cerveau personnel = données d'employé | Moyenne | Voir 7.1 |

---

## 7. Exigences mal posées et questions ouvertes

### 7.1 « Mémoire personnelle de l'employé, tous projets confondus »

Problème : un cerveau `user/<employé>` qui mélange tous les projets viole le cloisonnement par client quand un contrat l'impose. Il pose aussi une question de droit du travail : sous la nLPD, les notes de travail d'un employé sont des données de l'entreprise, mais une « mémoire personnelle » peut contenir des données sur l'employé lui-même (préférences, habitudes). Au départ de l'employé, l'entreprise doit savoir si elle garde ou efface.

Reformulation proposée : `user/<employé>` contient des **préférences de travail et du savoir-faire transverse** (outils, conventions, raccourcis), sans fait sur un client ni sur un projet. Les notes de projet vont dans `user/<employé>/<projet>`, cerveau local, effacé à la restitution du poste ou promu dans `team/<projet>` avant le départ. Le nom « mémoire de poste » est plus juste que « mémoire personnelle ».

### 7.2 « Traitement exclusivement interne »

Problème : l'exigence est contradictoire avec l'usage de Claude Code ou de Codex. Le hook de rappel **injecte les extraits dans le contexte**, donc les envoie au fournisseur du modèle de l'agent. Ollama interne ne change rien à cela. Le document lui-même reste interne, mais les extraits rappelés sortent.

Reformulation proposée : « Aucun traitement d'indexation, d'extraction ou de synthèse ne sort de l'entreprise. Les extraits rappelés sont soumis au même contrat que le code que l'agent lit déjà. Une classification `confidential` ou `personal-data` n'est injectée que si le contrat avec le fournisseur de l'agent couvre ce niveau. Sinon l'extrait est remplacé par une référence. » La même règle s'applique à la planification de requête par haiku (mod `cortexdb-live`) : le prompt sort déjà, mais pas les documents.

### 7.3 « Combinaison sans choix » et « règle d'écriture »

Les deux sont compatibles si la lecture est fédérée et l'écriture est **toujours explicite**. Un fait écrit « sans choix » irait au mauvais endroit une fois sur dix, et ce serait une fuite. La règle de la section 3.5 fixe la destination par défaut à la zone brouillon locale.

### 7.4 Questions à faire trancher par l'entreprise

Juridique :

1. Le contrat avec Anthropic (et OpenAI pour Codex) couvre-t-il les données `confidential` et `personal-data` dans le contexte des agents ? Avec quel sous-traitant et quelle localisation ?
2. Faut-il un cerveau `client/<nom>` cloisonné pour chaque client sous NDA, ou le cerveau `team/<projet>` suffit-il ?
2b. Le contrat Enterprise (DPA, absence d'entraînement, localisation) couvre-t-il l'envoi de documents `confidential` et `personal-data` par le siège bot du gateway ? Si oui, Ollama ne sert qu'aux embeddings, au reclassement et à la masse. Si non, Ollama traite ces deux classes. Voir 3.8.
3. Durée de conservation par classification. Proposition de départ : `public` et `internal` 5 ans, `confidential` durée du contrat + 2 ans, `personal-data` durée du besoin, revue annuelle.
4. Statut des cerveaux `user/` : données de l'entreprise ou données de l'employé ? Information à donner aux employés (règlement d'utilisation).

Sécurité :

5. Réseau de confiance : Tailscale, VPN existant, ou proxy TLS avec certificats internes ?
6. Authentification du gateway : OIDC de l'annuaire (préférable) ou clés personnelles émises à la main ?
7. Qui a le droit d'exporter un cerveau entier ?

Budget matériel :

8. GPU 24 Go (pilote, 14B) ou 48 Go (32B) ? Décision après la mesure du pilote.
9. Hébergement PostgreSQL : instance existante ou dédiée ?

Organisation :

10. Qui est le responsable de la gestion des connaissances (propriétaire de `rights.yaml` et de `org/reference`) ?
11. Qui porte les contributions vers l'amont et avec quelle adresse (compte d'entreprise ou personnel) ? Licence MIT : pas de contrainte de rétrocession, mais une politique de contribution interne est utile.
12. Le curateur à 10 % par cerveau : accepté par les chefs de projet ?

---

## Annexe A : ce qui reste à vérifier dans le code

1. ~~Nombre de contributeurs réel~~ : fait, voir 1.6.
2. `delete_document_graph` : retire-t-il les entités fusionnées depuis plusieurs documents, ou seulement les passages ?
3. `verify_claims` : vérifie-t-il la présence des termes d'un fait dans le passage cité ?
4. Format du point d'accès attendu par `reranker.go` (compatible Cohere, Jina, ou TEI ?).
5. Le changement de schéma proto entre versions : fréquence sur les 20 dernières versions, pour estimer le coût de suivi du gateway.
6. `side_graph_*` : convient-il comme file de relecture, ou est-il conçu pour un autre usage ?
7. Le `Watcher` CDC du paquet `connector` : réutilisable pour une table `sources` de fichiers et de pages ?
8. Le plugin Codex : mêmes hooks que Claude Code, ou sous-ensemble ?
