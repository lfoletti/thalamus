# Plan de publication vers l'amont CortexDB

Version du 18 octobre 2026. État : **projet, à valider par Lauro Foletti avant toute action**.

## 0. Règles

1. Aucune action publique (push vers `fork`, pull request, issue, discussion) sans validation écrite de Lauro, lot par lot.
2. Une pull request = une révision `GC-xx` = un commit de la branche `proto/granular-cortex`, rebasé sur `origin/main` du jour, **sans la modification de `CHANGELOG.md`** (section 1, règle du propriétaire).
3. Chaque pull request cite sa référence `GC-xx` et renvoie à la discussion publique qui expose le but de la série (section 6). Sans discussion publique, la référence reste dans le titre de la branche et dans ce document, pas dans le corps de la PR : un numéro privé ne dit rien au propriétaire.
4. Chaque révision a une solution de repli. Si le propriétaire la refuse, le repli s'applique et les objectifs restent atteignables.
5. Une révision refusée ne se réécrit pas dans le fork au long cours. Le fork porte au plus un petit jeu de patchs rebasés à chaque version, et seulement pour les révisions marquées « repli : patch temporaire ».

## 0 bis. Conformité aux règles de contribution amont

`CONTRIBUTING.md`, `.github/pull_request_template.md` et la pratique observée dans l'historique imposent ceci. État de conformité au 18 octobre 2026 :

| Règle amont | Source | Conforme ? | Action |
| :- | :- | :- | :- |
| Cible `main`, CI verte : `go test -race ./...` (25 min), floors `eval`, build des `examples/*`, cohérence de version | `CONTRIBUTING.md`, `ci.yml` | `eval` : passe. Suite complète : **verte** sur la branche (18 octobre 2026) | rejouer la suite complète sur chaque branche de PR avant le push |
| Avant toute modification de la recherche : `go test ./pkg/eval -run TestLexicalRetrievalQuality -v` | `CONTRIBUTING.md` | passe (GC-05, GC-06, GC-09 touchent la recherche) | rejouer à chaque rebase |
| Commits ciblés, style conventional-commit, **utilisé pour générer le changelog** | `CONTRIBUTING.md` | titres conformes. `CHANGELOG.md` : **non conforme**, 30 des 30 derniers commits qui le touchent sont du propriétaire, à la release | retirer `CHANGELOG.md` de chaque cherry-pick. Le texte écrit sert de corps de PR |
| Modèle de PR : Description, Type of Change, Testing, Checklist, Performance Impact, Breaking Changes | `pull_request_template.md` | les PR fusionnées (#42 à #61, dont la tienne) utilisent `Problem / Fix / Tests` et sont fusionnées quand même | garder `Problem / Fix / Tests` dans Description, et remplir les cases Type of Change, Testing, Performance Impact, Breaking Changes |
| Pas de bump de version par un contributeur | pratique (seules les releases du propriétaire le font) | conforme | ne jamais toucher `version.go` ni les `plugin.json` |
| Fusion en squash : une PR = un commit amont `titre (#N)` | historique | conforme avec GC-04 écrasé | — |
| `Co-Authored-By: Claude` | 20 commits sur 200 | accepté | garder |
| Type de contribution externe acceptée | historique : 23 `fix`, 1 `feat` (du propriétaire) sur 40 commits ; aucune PR externe `feat` fusionnée | **point faible** : GC-01, 03, 04, 05, 06, 08 sont des `feat` | voir section 2 : chaque `feat` est présenté par le problème qu'il corrige, et le lot A ne contient que des `fix` |
| Issues : une seule en 65 PR, modèle « Feature request » présent | historique | le propriétaire travaille par PR, pas par issue | le lot B s'ouvre par une **discussion** GitHub (activées), pas une issue, et n'attend pas plus de 7 jours une réponse avant d'ouvrir GC-03 |

## 1. Contexte amont

Faits vérifiés le 18 octobre 2026 avec `gh` :

- Propriétaire et releaser : Liang Li (652 commits). Second contributeur régulier : Rudy Mizrahi Celekli (52 commits, PR #42 à #65). Lauro Foletti : PR #61, fusionnée en 2 heures.
- Rythme de fusion : les PR #42 à #61 ont été fusionnées entre 2 heures et 1 jour après ouverture, par lots de 6 à 8.
- Taille des PR fusionnées : 35 à 204 lignes. Aucune PR fusionnée ne dépasse 250 lignes.
- Style accepté : titre conventional-commit, corps en trois sections `Problem`, `Fix`, `Tests`, pied `🤖 Generated with Claude Code`. Modèle : PR #61 et #64.
- `origin/main` n'a pas bougé depuis la base de la branche (v2.132.1, `4c43081`).
- PR ouvertes : #63 (quantization), #64 (`pkg/httpapi/policy.go`, variantes de casse des champs de portée), #65 (index IVF). La #64 touche le fichier que `GC-04` modifie aussi. Un rebase sera nécessaire après sa fusion.
- Issues et discussions GitHub sont activées. Une seule issue existe à ce jour.

Conséquences : des PR **petites, une par sujet, avec tests**, sont fusionnées vite. Une PR de 900 lignes sort du gabarit : elle exige une issue de discussion avant le code.

## 2. Objectifs de la série

### 2.1 Le but, dit au propriétaire

Un seul but, en termes CortexDB et non en termes Thalamus :

> **Un serveur CortexDB partagé par plusieurs équipes et plusieurs agents, où une clé confinée suffit à un client MCP, où un document garde sa provenance, sa durée de vie et son état de relecture, et où rien ne fuit d'une collection à l'autre.**

Chaque révision se lit comme un problème de ce serveur partagé, reproduit par un test qui échoue avant et passe après. Aucune PR ne parle de « Thalamus », de « gateway » ni de « fédération ». Ces mots sont notre couche, pas la leur.

Texte de la discussion publique : `docs/discussion_amont.md` (validé en principe le 18 octobre 2026, texte à relire avant publication).

Si la discussion n'est pas validée, le même texte, réduit à son premier paragraphe, ouvre le corps de la première PR du lot A.

### 2.2 Objectifs internes et replis

| Réf. | Objectif pour Thalamus | Révisions | Repli si tout est refusé |
| :- | :- | :- | :- |
| OBJ-1 | Audit par identité : savoir quelle clé a appelé quel outil | GC-01 | `th-gateway` journalise chaque appel dans `audit.jsonl` |
| OBJ-2 | Intégrité documentaire : provenance, hash, doublons | GC-02, GC-08 | le gateway porte la provenance en métadonnées de passage et tient son propre index de hachage |
| OBJ-3 | Clé confinée utilisable par le plugin MCP, plusieurs collections par clé | GC-03, GC-04, GC-10, GC-11 | un serveur `cortexdb-grpc` par cerveau, clé non confinée tenue par le gateway ; isolation par processus |
| OBJ-4 | Cycle de vie : rétention, expiration, balayage | GC-05 | `retention_until` en métadonnée, balayage par le gateway via `knowledge_delete` et `memory_delete` |
| OBJ-5 | Relecture humaine native dans le moteur | GC-06 | file de relecture entièrement dans le gateway (fichiers Git), seuls les faits acceptés entrent dans le cerveau |
| OBJ-6 | Étanchéité d'un serveur partagé entre collections | GC-09, GC-10, GC-11 | un serveur par équipe tant que la fuite existe |
| OBJ-7 | Hygiène du plugin (aucun fichier résiduel) | GC-07 | le plugin `thalamus` embarque son propre lanceur |

Les replis OBJ-3 et OBJ-6 coûtent de l'exploitation (N processus), pas des fonctions. Le repli OBJ-5 est le modèle principal du document de révision (section 3.6) : la révision GC-06 est un **confort**, pas une condition.

## 3. Inventaire des révisions

| Réf. | Commit | Type | Taille | Dépend de | Contenu | Risque de refus |
| :- | :- | :- | :- | :- | :- | :- |
| GC-07 | `138edb7` | fix | 30 | — | Lanceur `SessionEnd` : aucun `capture-payload.*` résiduel | faible : bug visible, correctif shell |
| GC-01 | `b7beb86` | feat | 141 | — | Journal gRPC : `key=` et `tool=` par appel | faible : additif, champ `omitempty` |
| GC-02 | `3be4302` | fix | 194 | — | `ingest_document` garde `source_url`, `author`, `metadata` sur la fiche et le nœud | faible : champs existants non remplis |
| GC-09 | `1d0fdf1` | fix | 180 | — | Expansion de graphe : un passage voisin doit être de la collection cherchée ; entités avec `collections` | faible à moyen : fuite réelle entre collections, test qui échoue avant |
| GC-03 | `f21e4d7` | feat | 264 | — | `Scope` : listes `collections`, `user_ids`, `scopes`, `namespaces` | moyen : sémantique de sécurité ; le propriétaire a écrit « deliberately left out » pour les listes par RPC, pas pour les valeurs |
| GC-04 | `8c0d1ff` | feat | 467 | GC-03 (pour `Permits`) | Clé confinée autorisée sur les outils qui déclarent le champ confiné ; arguments vérifiés | **élevé** : renverse une règle écrite (« opaque JSON ») ; doit partir **avec** GC-10 et GC-11 |
| GC-10 | `c429b74` | fix | 415 | GC-04 | Outil adressé par identifiant : la ligne existante doit être dans la portée | — : fait partie de GC-04 à la publication |
| GC-11 | `1256e4a` | fix | 60 | GC-04 | Toutes les graphies d'un champ (`USER_ID`) sont vérifiées | — : fait partie de GC-04 à la publication |
| GC-08 | `7edc498` | feat | 272 | GC-02 | `content_sha256`, ré-ingestion identique sans écriture, `dedupe` | moyen : change le comportement d'une ré-ingestion identique (ne réécrit plus) |
| GC-05 | `29c7d1e` | feat | 481 | — | `expires_at` sur documents, filtre à la lecture, `--sweep-expired` | moyen : nouvelle commande, cache de lecture |
| GC-06 | `4024f93` | feat | 395 | GC-05 (cache partagé) | `review_status`, `reviewed_by`, `reviewed_at` ; filtre `review_status` ; nœud `knowledge_save` avec métadonnées | moyen : politique produit (« proposed » caché par défaut) |

Taille = lignes ajoutées, tests compris.

## 4. Étapage

### Lot A : correctifs sans débat (4 PR indépendantes)

Ordre d'ouverture : GC-07, GC-01, GC-02, GC-09. Le même jour, 4 PR ouvertes. Chacune est rebasée sur `origin/main` et se fusionne seule.

Critère de passage au lot B : les 4 PR sont fusionnées, ou refusées avec un motif connu. Un refus de GC-02 bloque GC-08 (lot C) : GC-08 sera alors réécrit sur la seule métadonnée de passage, ou abandonné au profit du repli OBJ-2.

### Lot B : sécurité des clés (1 discussion, puis 2 PR)

1. **Discussion GitHub** (pas une issue : le dépôt n'en a jamais utilisé) : « Scoped keys and the toolbox: let a confined key call tools that declare its scope fields ». Elle expose le besoin (client MCP partagé avec clés confinées), la règle proposée (champ déclaré + argument égal + ligne existante dans la portée + toutes les graphies), les refus conservés (outil sans le champ, confinement non vérifiable) et les tests. **Attendre la réponse du propriétaire, 7 jours au plus, avant le code.** Sans réponse, GC-03 part seule et GC-04 part en PR **brouillon** (draft) avec la question dans son corps.
2. **PR GC-03** (listes dans `Scope`). Indépendante du résultat de la discussion : utile aux RPC typés seuls.
3. **PR GC-04** = commits GC-04 + GC-10 + GC-11 **écrasés en un seul** (environ 900 lignes), ou en trois PR si Lauro le choisit (section 7). Ouverte seulement si la discussion reçoit un accord de principe. Rebasée après la fusion de la #64 (même fichier `policy.go`).

Si le propriétaire refuse le principe : repli OBJ-3 (un serveur par cerveau). Les trois commits restent dans le fork comme **patch temporaire**, rebasé à chaque version, le temps du pilote. Décision de garder ou d'abandonner le patch à la fin de l'étape 1 du plan de mise en œuvre.

### Lot C : fonctions (3 PR, une à la fois)

Ordre : GC-08, GC-05, GC-06. Une PR ouverte à la fois, la suivante après fusion ou refus. GC-06 dépend de GC-05 : si GC-05 est refusée, GC-06 est réécrite sans le cache partagé (filtre direct) ou abandonnée au profit du repli OBJ-5.

### Rythme

- Au plus 4 PR ouvertes en même temps (lot A), sinon 1 à 2.
- Pas de nouvelle PR tant qu'une PR du même fichier est ouverte.
- Après chaque fusion : `git fetch origin && git rebase origin/main` sur `proto/granular-cortex`, retrait du commit fusionné, nouveau passage de `tests/e2e_shared_brain.sh`.

## 5. Procédure par PR

1. `git fetch origin`. Créer `gc-XX-<slug>` depuis `origin/main`.
2. `git cherry-pick <commit>` (ou les trois commits pour GC-04, puis `git reset --soft` et un seul commit).
3. **Retirer `CHANGELOG.md` du commit** (`git checkout origin/main -- CHANGELOG.md`, puis `commit --amend`). Le propriétaire écrit le changelog à la release, à partir du titre du commit. Le texte de notre entrée devient la section Description de la PR.
4. `gofmt -l`, `go vet`, `go build ./...`, puis **la suite complète comme en CI** : `go test -race -timeout 25m ./...`, et `go test ./pkg/eval -run TestLexicalRetrievalQuality -v` si la PR touche la recherche.
5. Rejouer `tests/e2e_shared_brain.sh` sur la branche `proto/granular-cortex` rebasée, pas sur la branche de PR.
6. **Validation de Lauro** sur le titre et le corps.
7. `git push fork gc-XX-<slug>`, puis `gh pr create -R liliang-cn/cortexdb`.
8. Suivi : répondre aux commentaires dans la journée. Un changement demandé est un commit supplémentaire sur la branche de PR, puis reporté sur `proto/granular-cortex`.
9. Après fusion : supprimer la branche sur le fork, rebaser la branche de travail, mettre à jour `CLAUDE.md` et ce plan.

## 6. Modèle de corps de PR

Le modèle amont (`.github/pull_request_template.md`) est respecté dans ses sections. Le contenu `Problem / Fix / Tests` des PR fusionnées (#61, #64) entre dans Description.

```
## Description

### Problem
<ce qui ne marchait pas, avec la trace ou le test qui échoue>

### Fix
<ce qui change, en 3 à 6 phrases. Ce qui ne change pas>

### Tests
- <test nouveau qui échoue avant, passe après>
- `go vet` and `go test -race ./...` pass; `TestLexicalRetrievalQuality` unchanged.

Part of a small series on running one CortexDB server for several teams
(see discussion #66, item GC-04). Each PR stands alone.

## Type of Change
- [x] Bug fix (non-breaking change which fixes an issue)

## Testing
- [x] Unit tests pass
- [x] New tests have been added

## Performance Impact
<une phrase : « none », ou le coût mesuré, par exemple « one indexed read per id argument on CallTool by a confined key »>

## Breaking Changes
None.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

La ligne « see discussion #66, item GC-04 » n'existe que si la discussion publique est validée (section 2.1). Sinon la phrase devient « One of a few small fixes from running one server for several teams », sans numéro.

## 7. Risques du plan

| Risque | Effet | Parade |
| :- | :- | :- |
| La #64 est fusionnée avant GC-04 | conflit dans `policy.go` | rebase, les deux changements touchent des blocs différents |
| Le propriétaire refuse le principe du lot B | OBJ-3 et OBJ-6 par processus séparés | repli documenté ; patch temporaire borné à l'étape 1 |
| GC-04 jugée trop grosse | pas de revue | proposer un découpage : (a) `authz` seul, non branché ; (b) branchement gRPC ; (c) branchement REST. Chaque morceau garde la règle complète (arguments + ligne + graphies) |
| Entrées `CHANGELOG.md` en conflit entre nos PR | rebase manuel à chaque fusion | une puce par PR, toujours en tête de `Unreleased` |
| Le propriétaire modifie une révision à la fusion | divergence avec `proto/granular-cortex` | après chaque fusion, rebaser et rejouer l'essai complet |
| Trois PR « sécurité » d'un même auteur en une semaine | lassitude du relecteur | lot B après le lot A, avec la discussion d'abord |
| Le propriétaire n'a jamais fusionné de `feat` externe | lot C et GC-01/03/04 refusés par principe | chaque `feat` est titré et présenté par le problème qu'il corrige ; GC-01 et GC-08 peuvent être retitrés `fix` sans mentir (journal incomplet, ré-ingestion qui réécrit à l'identique). GC-05 et GC-06 restent des `feat` : les replis OBJ-4 et OBJ-5 sont complets |
| Le changelog est écrit par le propriétaire | une PR qui le touche crée un conflit à chaque release | aucune PR ne touche `CHANGELOG.md` |

## 8. Décisions attendues de Lauro

1. Valider les lots A, B, C et leur ordre, ou les modifier.
2. Dire si GC-04 part en une PR (avec GC-10 et GC-11 écrasés) ou en trois morceaux (7, ligne « trop grosse »).
3. Dire si le plan est publié comme discussion GitHub amont, ou reste privé.
4. Donner le feu vert du lot A, PR par PR ou en bloc.

Rien ne part avant ces réponses.

## 9. Journal d'exécution

| Date | Événement |
| :- | :- |
| 18 octobre 2026 | Discussion amont publiée par Lauro : https://github.com/liliang-cn/cortexdb/discussions/66 |
| 18 octobre 2026 | GC-07 : PR #67 ouverte par Lauro (`gc-07-session-end-payload`, `62c293d`). Titre corrigé (retour à la ligne hérité du terminal). |
| 18 octobre 2026 | GC-01, GC-02, GC-09 : branches locales préparées (`gc-01-request-log-key-tool` `1232086`, `gc-02-ingest-provenance` `bda5816`, `gc-09-collection-confined-graph-walk` `23a5258`), corps dans `docs/pr/`, commandes dans `docs/pr/lot_A.md`. GC-09 a demandé une résolution de conflit (le commit s'appuyait sur `allowDocument` des correctifs 5 et 6) et un test rendu autonome (auxiliaires des correctifs 5, 6, 8 absents de `main`). Le test de fuite échoue sur `main` et passe sur la branche. Suites complètes par branche : en cours. |

Leçon pour les lots B et C : un commit de `proto/granular-cortex` qui s'appuie sur un correctif antérieur non fusionné se réécrit sur `main` au cherry-pick. Vérifier à chaque fois : conflit, compilation des tests (auxiliaires partagés), test qui échoue avant.
| 18 octobre 2026 | Lot A complet, ouvert par Lauro : GC-01 = PR #68, GC-02 = PR #69, GC-09 = PR #70. CI amont : `rust-client`, `dsh-plugin`, `obsidian-plugin` verts sur les quatre, `test` (suite Go) en cours. Suite locale complète : GC-01 verte (30 paquets, eval ok) ; GC-02 et GC-09 en cours. |
| 18 octobre 2026 | Lot A : suites locales complètes vertes sur les trois branches (30 paquets, eval ok). |
| 18 octobre 2026 | Lot B préparé dans `docs/pr/lot_B/` : proposition (commentaire sur #66), GC-03 `gc-03-scope-lists` `657b0f5`, GC-04 `gc-04-confined-keys-toolbox` `4ae68da` (empilée sur GC-03, correctifs 4+10+11 écrasés). PR #64 s'applique sur GC-04 sans conflit. Suites complètes : en cours. Rien n'est publié. |
| 18 octobre 2026 | Lot B : suites complètes vertes sur GC-03 (`657b0f5`) et GC-04 (`4ae68da`), eval ok. |
| 18 octobre 2026 | PR #70 : Codecov signale 14 lignes non couvertes (10 branches d'erreur, 4 fonctionnelles). L'analyse a révélé un **vrai défaut** : la règle d'entité était sous `case "entity"`, or une entité typée (`organization`) a son type comme `NodeType`. La règle ne s'appliquait donc jamais aux entités de `knowledge_save` et `ingest_document` ; seul le filtre des passages protégeait. Correctif `6648ff5` sur `gc-09-collection-confined-graph-walk` (règle dans la branche `default`, entité typée comprise, ensemble d'entités inchangé) et test direct de la règle (mention faite à la main par `upsert_entities`). Reporté sur `proto/granular-cortex` en `adbb0b2`. À pousser par Lauro après la suite complète. |

Leçon : un rapport de couverture sur une PR de sécurité se lit ligne par ligne. Une branche fonctionnelle non exécutée est soit un test manquant, soit du code mort, et du code mort dans une règle de sécurité est un trou.
| 18 octobre 2026 | GC-09 : correctif `6648ff5` poussé par Lauro vers la PR #70. Suite complète verte avant le push. |
| 18 octobre 2026 | Lot B, étape 1 : proposition publiée par Lauro en commentaire sous #66 (https://github.com/liliang-cn/cortexdb/discussions/66#discussioncomment-19021755). Délai de réponse : jusqu'au 25 octobre 2026. GC-03 prête à partir. |
| 19 octobre 2026 | Réponse du propriétaire sous #66 (18 octobre, 23 h 49 UTC) : lot A (#67 à #70) et GC-03 (#71) fusionnés, publiés en v2.133.0 (puis v2.134.0, fonction « recall » du propriétaire). #64 et #70 qualifiés de vraies fuites, reproduites sur `main`. Lot B accepté sur le principe : à envoyer en **trois PR** (règles `authz` seules, câblage gRPC, câblage REST), chacune avec ses tests qui échouent avant. Lot C : hash de contenu et `expires_at` acceptés ; pour `review_status`, le masquage des lignes non relues doit être **opt-in** (un cerveau qui ne pose jamais le champ cherche exactement comme aujourd'hui). Correctif 6 (`4024f93`) à revoir : il masque `proposed` et `rejected` par défaut. |
| 19 octobre 2026 | Branche `proto/granular-cortex` reconstruite sur v2.134.0 (ancien état : tag `backup/proto-granular-cortex-2026-10-19`). Lot B découpé en trois branches selon la demande du propriétaire : `gc-04a-authz-tool-rules` (`984cda7`, `pkg/authz` seul, rien de câblé), `gc-04b-grpc-confined-tools` (`fa67f4a`, sur 04a), `gc-04c-rest-confined-tools` (`9d07c39`, sur 04a). Somme identique au commit écrasé rebasé (`84a2739`). `go vet`, `go test -race` des paquets `authz`, `rpcserver`, `httpapi` : verts à chaque étape. Corps des PR : `docs/pr/lot_B/GC-04a.body.md`, `GC-04b.body.md`, `GC-04c.body.md`. Correctif 6 réécrit en opt-in (`3473f52`) : sans `CORTEXDB_REVIEW_GATE=1`, toute ligne est servie comme avant ; le filtre `review_status` reste utilisable sans la porte. Rien n'est publié. |
| 19 octobre 2026 | Exigence du propriétaire « each with its failing-first tests » vérifiée (`docs/pr/lot_B/failing_first.md`). 04b et 04c : leurs tests posés sur le code sans câblage échouent (5 et 2 tests). 04a : onze mutations, une règle sans test (autre graphie dans une copie imbriquée, `plan.filters.USER_ID`) ; cas ajouté, 04a amendée en `9827f2a`, 04b `19c9028`, 04c `bc214f6`, `proto/granular-cortex` rebasée (`9827f2a`..`d7fce69`). Suite `-race` des paquets touchés verte sur la branche avant l'amendement, `pkg/authz` relancé après. Essai de bout en bout : 30/30 avec `CORTEXDB_REVIEW_GATE=1` sur le serveur. |
| 19 octobre 2026 | GC-04a : PR #72 ouverte par Lauro (https://github.com/liliang-cn/cortexdb/pull/72, branche `gc-04a-authz-tool-rules`, `9827f2a`). Avant le push : `go test -race -timeout 25m ./...` sans aucun échec et `TestLexicalRetrievalQuality` vert sur la branche. 04b (`19c9028`) et 04c (`bc214f6`) attendent la fusion de #72, puis rebase. |
| 19 octobre 2026 | Lot B complet ouvert par Lauro : GC-04b = PR #73 (`gc-04b-grpc-confined-tools`, `19c9028`), GC-04c = PR #74 (`gc-04c-rest-confined-tools`, `bc214f6`), empilées sur #72. Tests `-race` des paquets changés et dépendants verts sur chaque branche avant le push. Décision de Lauro : enchaîner le lot C sans nouvelle demande quand le lot B est fusionné. |
| 19 octobre 2026 | Lot C préparé en pile sans `CHANGELOG` : `gc-08-content-hash` `fd1c4ba` (sur v2.134.0), `gc-05-expires-at` `d75c459` (sur 08 : conflit sur `ingest_document` si posée seule), `gc-06-review-status` `d0d4025` (sur 05). Suite complète `-race` sur la pile : aucun échec ; `TestLexicalRetrievalQuality` vert. En attente de la fusion du lot B pour rebase, push et PR (autorisation donnée). CI amont `test` encore en cours sur #72, #73, #74. |
| 19 octobre 2026 | CI amont `test` verte sur #72 (14 min 28 s). Codecov sur #72 : 17 lignes non couvertes dans `rows.go` et `tools.go`. Lecture ligne à ligne : deux branches fonctionnelles sans test (copie d'un champ de portée dans un tableau, borne de profondeur) et cinq chemins d'erreur de la garde des lignes. Deux tests ajoutés, garde durcie (outil inconnu refusé). Plus aucune ligne non couverte dans ces deux fichiers. 04a amendée `abe694f`, 04b `77cac1d`, 04c `ae3775a`, `proto/granular-cortex` rebasée (`abe694f`..`979b17f`), tests `-race` verts sur les trois têtes. Branches poussées de force, corps de #72 mis à jour. Lot C : 20 mutations, 20 tuées (`docs/pr/lot_C/failing_first.md`). |
| 19 octobre 2026 | Codecov sur #73 (`ownership.go`, 6 lignes) et #74 (`policy.go`, 8 lignes). Lecture ligne à ligne : chemins défensifs et d'erreur ; tests ajoutés sur les deux ports (échec fermé : message d'un autre type, outil inconnu, arguments non objet, lecture en échec, corps trop gros). 04b amendée `dd69ed9`, 04c `6b1430a`, `proto/granular-cortex` reconstruite (`abe694f`, `b68071d`, `c39a9c2`, `690b96e`, `5064567`, `df36d19`), tests `-race` verts. Seule ligne du lot B encore non couverte : le bras 403 de la garde REST, défensif. |
| 19 octobre 2026 | Lot C ouvert par Lauro sans attendre la fusion du lot B (fichiers disjoints) : GC-08 = PR #75 (`gc-08-content-hash`, `fd1c4ba`), GC-05 = PR #76 (`gc-05-expires-at`, `d75c459`, empilée sur 08), GC-06 = PR #77 (`gc-06-review-status`, `d0d4025`, empilée sur 05). Avant le push : suite complète `-race` et `eval` verts sur la pile, 20 mutations tuées. Six PR ouvertes au total (#72 à #77). |
| 19 octobre 2026 | CI de #73 : un test aléatoire amont (`TestHNSWLevelsThinOutByAFactorOfM`, 2 % d'échec par passage, hors périmètre) ; commit 04b recréé à l'identique (`b70b9d5`), CI relancée, verte. Codecov sur #75 (13 lignes) et #76 (84 lignes) : lecture ligne à ligne ; #75 : test « clé de métadonnées retirée » ; #76 : test de la marche PPR avec un passage expiré, test d'effacement d'`expires_at`, commande `--sweep-expired` refondue en fonction qui rend son code de sortie et testée ; sept mutations tuées (27 au total pour le lot C). Trois erreurs corrigées avant push (attente de version, type de métadonnées, assertion cassée par 06) : vérifier chaque code de retour explicitement, un tube le masque. Têtes : 08 `88ccd5b`, 05 `b0af1fa`, 06 `ea893b5`, poussées ; proto reconstruite (`edf488f`, `f5b20fe`, `e47a370`). Tests `-race` verts sur les trois têtes. Commentaire Codecov posté sous #74 (bras 403 défensif). |
| 19 octobre 2026 | Codecov sur #77 : lignes du commit 06 lues ; tests ajoutés (porte sur les quatre chemins de mémoire, filtres, `memory_update`, `reviewed_at`, nœud de document, `CORTEXDB_REVIEW_GATE` dans les deux binaires avec `reviewGateFromEnv` côté gRPC) ; 12 mutations tuées, 1 survivante classée défensive (`document_ids` dans la porte). #75 : second passage, l'ordre des cas du test avait découvert la comparaison des valeurs ; corrigé, deux mutations tuées. Décision de Lauro : plus de ligne `Co-Authored-By` sur les commits ; six messages réécrits, arbres identiques, six branches poussées de force : 04a `8b8163e`, 04b `49263e7`, 04c `3e999cf`, 08 `d3360fc`, 05 `45111c3`, 06 `020f2a9`. Proto reconstruite : `6c15d83`, `5ce527e`, `6a9baad`, `00f3f4a`, `4a8aafc`, `5e03106`. Tests `-race` verts sur la tête 06 (`pkg/cortexdb`, `cmd/cortexdb-mcp-stdio`, `cmd/cortexdb-grpc`). |
| 19 octobre 2026 | Rapports Codecov finaux : #72 et #73 complets ; `codecov/patch` au vert partout (pas de `codecov.yml`, seuil par défaut), la croix rouge du commentaire signifie « moins de 100 % ». Dernier passage sur les chemins d'erreur atteignables : tests « cerveau fermé » (ingestion, porte, balayage) et cas direct de `checkReview` sans statut ; 4 mutations, 3 tuées, 1 survivante (recherche de hash : l'écriture suivante échoue aussi, contrat tenu de bout en bout). Tests `-race` verts sur la tête 06. Têtes : 08 `d43088f`, 05 `13456a3`, 06 `7299044` ; proto `72c1c63`, `1a4eac9`, `4ca55a3`. Restes non couverts, injouables : lectures ligne à ligne en échec, suppression en échec en cours de balayage, chemin par défaut vers le cerveau de l'utilisateur, `main` des binaires, bras 403 de #74. |
| 19 octobre 2026 | Rapports Codecov finaux du lot C : #75 91,4 % (6 lignes), #76 83,0 % (39), #77 87,4 % (45) ; `codecov/patch` et CI au vert sur les six PR. Les lignes restantes sont celles que `make patchcov` liste en local (injouables). Commentaire 2 de la discussion #66 complété (PR incomplètes seulement, plus un mot d'excuse pour les poussées répétées). Outils locaux en place : `tools/patchcov.py`, `Makefile` (`precommit`, `patchcov`, `mutate`, `hooks`), hook `pre-push`, `gremlins` ; `make mutate` lancé sur le lot C (91 mutants). |
| 19 octobre 2026 | `make mutate` sur le lot C : 91 mutants, 76 tués, 15 survivants, 5 non couverts, en 5 min 34 s. Survivants lus un à un : 9 faux (paquets `main`, tués à la main ; limite de `gremlins` consignée dans CLAUDE.md), 1 vrai manque corrigé (`reviewed_at` d'une mémoire ; test ajouté, mutation tuée), 1 ligne amont voisine, 3 équivalents. 06 amendée `354b79e`, proto `66cc841`. Tests `-race` en cours, push de 06 ensuite. |
| 19 octobre 2026 | Deux commentaires postés sous la discussion #66 à 19 h 08 UTC (réponse au propriétaire et question sur la taille des PR : discussioncomment-19048378 ; tableau des six PR, Codecov des incomplètes, mot d'excuse : discussioncomment-19048381). Textes : `docs/discussion_66/`. |
| 19 octobre 2026 | Règle de Lauro : plus aucun commit poussé sans vérification complète en local. Lot C refait pour couvrir chaque ligne du patch hors `main` : pannes SQLite déterministes (ligne à `id` NULL, déclencheur `RAISE(ABORT)`, colonne renommée), `eachRow` pour les deux boucles de lignes, `appendReviewGate` sorti de `main`, garde-fou injouable retiré dans #75. Têtes locales non poussées : 08 `bd96d99`, 05 `17f9208`, 06 `6bb43e7`. 16 mutations à la main tuées (`docs/pr/lot_C/failing_first.md`, troisième passe). Vérification séquentielle en cours : patchcov empilé, gremlins, `-race` par tête, suite complète, éval lexicale. |
| 19 octobre 2026 | #75, #76 et #77 passées en brouillon (`gh pr ready --undo`) pendant la vérification locale, à la demande de Lauro. À repasser « ready for review » après la poussée vérifiée. |
| 19 octobre 2026 | Vérification complète du lot C refait, toute verte (19 h 27 à 19 h 55 UTC) : patchcov 3 blocs dans `main` seulement, gremlins 76/95 tués et survivants tous classés, `-race` vert sur les trois têtes, suite complète 30 paquets verts, éval lexicale verte. Prototype reconstruit : 08 `924893e`, 05 `69cccba`, 06 `99d88fe` (code identique aux têtes 08 `bd96d99`, 05 `17f9208`, 06 `6bb43e7`). En attente de la décision de Lauro pour pousser les trois branches en une fois, remettre les corps à jour et sortir les PR du brouillon. |
| 19 octobre 2026 | #75 poussée (`gc-08-content-hash` `d43088f` → `bd96d99`, poussée forcée avec bail) sur demande de Lauro, après passage du hook `pre-push` (`make precommit` : `-race` vert en 258 s, patchcov 0 bloc). Premier essai refusé par le hook : `#` dans une fonction `$(shell …)` du Makefile, lu comme commentaire par `make` ; corrigé (`sed 's|…|…|'`). Corps de #75 remis à jour. Toujours en brouillon jusqu'au rapport Codecov. #76 et #77 attendent. |
| 19 octobre 2026 | #74 : les 3 lignes du bras `403` étaient injouables derrière le contrôle d'arguments (même lecture JSON des deux côtés). Correspondance erreur → statut sortie dans `toolGuardStatus`, testée sur les trois cas (`TestARowGuardFailureHasItsOwnStatus`) ; 3 mutations tuées (404↔403, 403 supprimé, garde ignoré) ; patchcov 0 bloc ; `-race` `pkg/httpapi` vert. Commit 04c amendé en local `c18bc00` (GitHub `3e999cf`), NON poussé, corps `docs/pr/lot_B/GC-04c.body.md` prêt. Décision de Lauro. |
| 19 octobre 2026 | #74 poussée (`gc-04c-rest-confined-tools` `3e999cf` → `c18bc00`) après le hook (`-race` `pkg/authz` et `pkg/httpapi` verts, patchcov 0 bloc). Corps à jour. #76 poussée (`17f9208`), corps à jour. Prototype reconstruit : 4c `892d7ee`, 8 `6dfc41c`, 5 `912090e`, 6 `1bdf547`. Veilles CI/Codecov sur #74 et #75. |
| 19 octobre 2026 | #77 poussée (`gc-06-review-status` `7299044` → `6bb43e7`) après le hook : `-race` verts, patchcov 0 bloc hors 3 dans `main` affichés et déclarés dans le corps. Corps à jour. Les six PR sont maintenant sur leurs têtes vérifiées. Veilles CI/Codecov sur #74, #75, #76, #77. |
| 19 octobre 2026 | #75 : CI verte, Codecov « All modified and coverable lines are covered by tests », `codecov/patch` SUCCESS. Première PR du lot à zéro ligne. Toujours en brouillon ; sortie du brouillon sur accord de Lauro. |
| 19 octobre 2026 | #75 sortie du brouillon (`gh pr ready`), ouverte à la relecture du propriétaire, tête `bd96d99`. #76 et #77 restent en brouillon jusqu'à leurs rapports Codecov. |
| 19 octobre 2026 | #74 : CI verte, Codecov « All modified and coverable lines are covered by tests ». Le lot B est complet à 100 % sur ses trois PR. |
| 19 octobre 2026 | Codecov sur les nouvelles têtes : #76 99,1 % (2 lignes, `main` MCP), #77 98,6 % (5 lignes : 3 `main` gRPC, 2 `main` MCP), exactement les lignes déclarées dans les corps. CI verte. #76 et #77 sorties du brouillon à 20 h 30 UTC. Bilan du soir : six PR ouvertes à la relecture, #72 #73 #74 #75 à 100 %, #76 et #77 à 100 % hors `main`. |
| 19 octobre 2026 | Pipeline interne formalisé sur demande de Lauro : `docs/pipeline_pr.md` (règles, outils, code couvrable, procédure en 10 étapes, lecture de gremlins, pièges), cible `make prepr` (precommit + mutate + suite complète + éval), CLAUDE.md réécrit (section « Pipeline avant toute PR », règle « PR en brouillon d'abord, ready après Codecov à zéro ligne ou lignes de `main` déclarées »). |
| 19 octobre 2026 | Ligne finale « Codecov: the N lines left uncovered are inevitable. … » ajoutée aux corps de #76 (2 lignes) et #77 (5 lignes), appliquée sur GitHub. Pipeline et CLAUDE.md : le test s'écrit avant le code et tombe sans lui ; la mutation à la main reste la preuve livrée au relecteur. |
| 19 octobre 2026 | Fin de soirée (21 h UTC). Six PR ouvertes à la relecture, aucune en brouillon, aucune action en attente de notre côté. Pipeline formalisé (`docs/pipeline_pr.md`, `make prepr`, hook), règles dans CLAUDE.md projet et global, leçons dans le cerveau du projet. Reprise : réponse du propriétaire (taille des PR) ou fusions ; après chaque fusion, rebaser les PR empilées et pousser via le hook. |

