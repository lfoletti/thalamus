# Lot C : hash de contenu, expiration, statut de relecture

Préparé le 19 octobre 2026. **Publié le 19 octobre 2026** : GC-08 = PR #75, GC-05 = PR #76, GC-06 = PR #77, ouvertes avant la fusion du lot B sur décision de Lauro (fichiers disjoints, aucune dépendance). Le propriétaire a accepté les trois sujets dans sa réponse sous #66 (18 octobre 2026), avec une condition pour le statut de relecture : le masquage des lignes non relues doit être opt-in. Le correctif 6 a été réécrit en ce sens (`CORTEXDB_REVIEW_GATE=1`, `cortexdb.WithReviewGate()`).

## État (20 octobre 2026)

**21 octobre 2026** : second tour (« changes requested » sur les trois, 00 h 25 UTC). Branches rebasées sur v2.135.0 (`cc98c6c`) sans changement de code : 08 `910e1cf`, 05 `8280fda`, 06 `30f2380`. Plan d'amendement, décisions D1 à D9 (clés réservées sous `_`, champs explicites, porte sans cache, ré-ingestion en place) et note de développement : `docs/pr/round2/`. Rien n'est poussé.

**21 octobre 2026, amendement du second tour** (en local, rien n'est poussé) : 08 `1e5ead6` (ré-ingestion en place, chunk *i* reste chunk *i* avec son graphe, seuls les chunks au-delà de la fin retirés ; dédup dans la transaction d'écriture, verrou consultatif PostgreSQL ; clés du moteur `_content_sha256`, `_chunk_size`, `_chunk_overlap`), 05 `dd69603` (champ explicite `expires_at` sur les trois outils, clé `_expires_at`, porte sans cache, quatre chemins de plus, `sweep` revérifié dans la suppression, test du balayage sous le canari), 06 `6cac298` (champs explicites `review_status`, `reviewed_by`, `reviewed_at` sur cinq outils, clés `_review_*`, décision reportée par toute écriture qui ne décide pas, filtre de statut sur les agrégats, fermeture sur panne, dédup qui voit un détenteur proposé ou rejeté, phrase de curation dans les descriptions). Mutations 12/12, 15/15, 23/23 (`docs/pr/round2/failing_first.md`). Corps : `docs/pr/round2/GC-08.body.md`, `GC-05.body.md`, `GC-06.body.md` ; réponses : `docs/pr/round2/review_replies.md` ; rapport : `docs/pr/round2/REPORT.md`. Sur `proto`, le commit de classement est `2d75ee5`.

Le propriétaire a posé « changes requested » sur les trois PR (01 h 21 UTC). Les trois branches sont amendées en place et rebasées sur v2.134.1 (`60a8831`) : `gc-08-content-hash` (`10e4b76`, +226 / −49 hors tests), `gc-05-expires-at` (`3c7a9fd`, sur 08, +505 / −22), `gc-06-review-status` (`036d1eb`, sur 05, +451 / −75). Deux écarts corrigés le 20 octobre au soir : `--sweep-expired` balaie le cerveau partagé par gRPC (outil `sweep_expired`) au lieu de refuser, et une `knowledge_update` qui décide d'une relecture fusionne ses métadonnées dans les métadonnées stockées au lieu de les remplacer. Un trou trouvé ensuite pendant l'analyse de `cortex_query` (`docs/pr/lot_D/cortex_query_allowlist.md`) : les voies de `cortex_query` ne passaient par aucune des deux portes ; corrigé dans #76 (porte appliquée à chaque voie dans `Query`) et couvert dans #77. Puis `precommit` a montré la fermeture MCP de `sweep_expired` sans test : test par session MCP ajouté dans #76 (`3c7a9fd`, test seul). Niveau 3 (gremlins, suite complète `-race`, éval) passé sur les deux têtes le 20 octobre au soir, survivants classés dans `failing_first.md`. Preuve : `failing_first.md`, section « Amendement du 20 octobre 2026 » (`mutate_amend.py 08|05|06 <worktree>`). Corps réécrits (section « Changes since review », sans pied de page). Réponses au relecteur en brouillon : `review_replies.md`. **Poussé le 20 octobre 2026 à 20 h 05** sur commande de Lauro : `gh pr ready --undo` sur les trois, `DECLARED=pkg/core/store_postgres.go git push --force-with-lease=<branche>:<ancienne tête> fork <branche>` depuis chaque copie de travail (hook vert : #75 0 bloc, #76 0 bloc + 1 `main` + 1 déclaré, #77 0 bloc + 3 `main` + 1 déclaré), corps mis à jour, un commentaire par PR, commentaire de lot sous la discussion #66. Codecov (API `compare`, 20 octobre 2026, 20 h 20) : #75 0 ligne non couverte ; #76 3 lignes (2 dans `main` de `cortexdb-mcp-stdio`, 1 dans la liste DDL) ; #77 7 lignes (3 dans `main` de `cortexdb-grpc`, 2 dans `main` de `cortexdb-mcp-stdio`, 2 dans la liste DDL) : exactement les lignes déclarées dans les sections `## Codecov`. `gh pr ready` passé sur les trois.

## Dépendance croisée avec le lot B (constatée le 20 octobre 2026)

La liste blanche de #72 est tenue au schéma des outils par un test de complétude. Les arguments et l'outil que le lot C ajoute (`dedupe` sur `ingest_document` ; `expires_at` sur `ingest_document`, `knowledge_save`, `knowledge_update` ; `review_status`, `reviewed_by`, `reviewed_at` sur ces trois outils et sur `memory_save`, `memory_update` ; filtre `plan.filters.review_status` sur six outils de recherche ; outil `sweep_expired` à refuser aux clés confinées) ne sont pas dans la liste de #72, qui part de `main`. Chaque PR est cohérente seule ; la réunion des deux lots fait tomber `TestEveryToolAndEveryArgumentIsClassifiedForConfinedKeys` et l'essai de bout en bout. Sur `proto/granular-cortex`, un septième commit classe ces arguments (`pkg/authz/allowlist.go`). **À faire au rebasage** : quand #72 est fusionnée, chaque PR du lot C ajoute ses propres arguments à la liste (08 : `dedupe` ; 05 : `expires_at` et `sweep_expired` dans `refusedTools` ; 06 : `plan.filters.review_status`), et le corps le dit. Si le lot C est fusionné d'abord, c'est #72 qui les ajoute.

## Ordre

Le lot C part après le lot B, pour ne pas faire relire trois sujets en même temps. Les trois PR sont indépendantes du lot B.

Les trois PR forment une pile, ouvertes ensemble comme le lot B, chacune rebasée sur `main` dès la fusion de la précédente :

1. **GC-08** (hash de contenu) : sur `origin/main`.
2. **GC-05** (`expires_at`, balayage) : **empilée sur GC-08**. Les deux ajoutent un argument à `ingest_document` aux mêmes lignes (`graphrag_tool_defs.go`, `graphrag_tool_ingest.go`, `graphrag_tool_types.go`) ; posée seule sur `main`, GC-05 est en conflit (vérifié le 19 octobre 2026).
3. **GC-06** (`review_status`) : **empilée sur GC-05** : la porte de recherche (`expiry.go`, une requête en cache pour `expires_at` et `review_status`) est commune.

## Contenu du dossier

| Fichier | Rôle |
| :- | :- |
| `GC-08.body.md` | corps de la PR #75 |
| `GC-05.body.md` | corps de la PR #76 (`#N8` remplacé par #75) |
| `GC-06.body.md` | corps de la PR #77 (`#NE` remplacé par #76) |
| `failing_first.md`, `mutate_lotC.py`, `mutate_amend.py` | preuve que chaque règle du lot a un test qui tombe sans elle (20 mutations, puis 36 pour l'amendement : 8 sur #75, 13 sur #76, 15 sur #77, plus 11 et 5 des paquets `main` rejouées à la main) |
| `review_replies.md` | brouillons des réponses au relecteur (20 octobre 2026) |
| `scenarios/pkg/cortexdb/reingest_scenario_test.go` | scénarios différentiels de `make diffmain` (21 octobre 2026) : ré-ingestion et graphe attaché, clés de métadonnées du client. Sur `036d1eb` contre `origin/main`, le diff montre les points 1 et 2 de #75 et le point 1 de #76 et #77 (second tour de revue). |
| `scenarios/pkg/cortexdb/reingest2_scenario_test.go` | second scénario différentiel (21 octobre 2026) : autre texte sur un graphe attaché. Il a montré que le premier jet du second tour retirait le chunk 0, qui survit sur `main` ; sur `6cac298` contre `origin/main`, le diff ne porte plus que les chunks au-delà de la fin du nouveau texte, nommés dans le corps de #75. Les scénarios de dédup (champs absents de `main`) sont des tests unitaires de la branche. |

## Branches locales

À créer depuis `proto/granular-cortex`, commit par commit, **sans l'entrée `CHANGELOG`** (section 0 bis du plan) :

| Réf. | Branche (tête) | Commit source sur `proto/granular-cortex` | Base |
| :- | :- | :- | :- |
| GC-08 | `gc-08-content-hash` (`bd96d99`) | `45c897b` | `origin/main` (v2.134.0) |
| GC-05 | `gc-05-expires-at` (`17f9208`) | `ae46541` | `gc-08-content-hash` |
| GC-06 | `gc-06-review-status` (`6bb43e7`) | `d7fce69` | `gc-05-expires-at` |

Branches créées et poussées le 19 octobre 2026 (contenu identique à `proto/granular-cortex` moins `CHANGELOG.md`). Après la fusion de #75 : rebase de `gc-05-expires-at` et `gc-06-review-status` sur `origin/main`, push forcé. Après la fusion de #76 : idem pour `gc-06-review-status`.

## Commandes de publication (préfixe `!`, une ligne chacune, après accord de Lauro)

```
cd cortexdb && git push fork gc-08-content-hash && gh pr create -R liliang-cn/cortexdb --head lfoletti:gc-08-content-hash --title "feat(graphrag): ingest_document hashes content, skips identical re-ingests, can refuse duplicates" --body-file ../docs/pr/lot_C/GC-08.body.md
```

```
cd cortexdb && git push fork gc-05-expires-at && gh pr create -R liliang-cn/cortexdb --head lfoletti:gc-05-expires-at --title "feat(cortexdb): documents expire, and a sweep removes what has expired" --body-file ../docs/pr/lot_C/GC-05.body.md
```

```
cd cortexdb && git push fork gc-06-review-status && gh pr create -R liliang-cn/cortexdb --head lfoletti:gc-06-review-status --title "feat(cortexdb): documents and memories carry a review status, hidden on request" --body-file ../docs/pr/lot_C/GC-06.body.md
```
