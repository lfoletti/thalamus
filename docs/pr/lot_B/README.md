# Lot B : clés confinées et toolbox

Préparé le 18 octobre 2026, redécoupé le 19 octobre 2026 après la réponse du propriétaire sous #66. **Rien n'est publié.** Chaque étape attend l'accord de Lauro et sa commande.

## État

- **20 octobre 2026** : le propriétaire a posé « changes requested » sur les trois PR (01 h 21 UTC). Les trois branches sont amendées en place et rebasées sur v2.134.1 (`60a8831`) : `gc-04a-authz-tool-rules` (`3c93d00`, liste blanche explicite par outil, `pkg/authz/allowlist.go`), `gc-04b-grpc-confined-tools` (`49949eb`), `gc-04c-rest-confined-tools` (`4b3ff4c`). Tailles hors tests : 04a +727 / −12 (dont environ 200 lignes de liste), 04b +61 / −10, 04c +54 / −6. Preuve : `failing_first.md`, section « Amendement du 20 octobre 2026 » (`mutate_amend.py`, `mutate_wiring.py`). Corps réécrits (`GC-04*.body.md`, section « Changes since review », sans pied de page). Réponses au relecteur en brouillon : `review_replies.md`. **Rien n'est poussé** : poussée du lot sur commande de Lauro (`gh pr ready --undo` sur les trois, `git push --force-with-lease fork <branche>` par branche, puis commentaires).

- Proposition publiée le 18 octobre 2026 (commentaire sur #66). Réponse du propriétaire le même jour : accord de principe, **trois PR** demandées (règles `authz` seules, câblage gRPC, câblage REST), chacune avec ses tests qui échouent avant.
- GC-03 : PR #71, fusionnée, publiée dans v2.133.0.
- GC-04 : découpée en trois branches, toutes basées sur `origin/main` à v2.134.0 (`5965a15`). **Les trois PR sont ouvertes le 19 octobre 2026** : GC-04a = #72, GC-04b = #73, GC-04c = #74 (https://github.com/liliang-cn/cortexdb/pulls/lfoletti). #73 et #74 sont empilées sur #72 (leur premier commit est celui de #72) ; à rebaser sur `main` dès la fusion de #72.

## Contenu du dossier

| Fichier | Rôle |
| :- | :- |
| `00_proposal_comment.md`, `00_proposal_comment.body.md` | proposition publiée sous #66 (archive) |
| `GC-03.body.md` | corps de la PR #71 (archive) |
| `GC-04a.body.md` | corps de la PR « règles `authz`, non câblées » |
| `GC-04b.body.md` | corps de la PR « câblage gRPC » (`#NA` remplacé par #72) |
| `GC-04c.body.md` | corps de la PR « câblage REST » (`#NA` remplacé par #72 ; remplacer `#NB` par le numéro de la PR 04b) |
| `failing_first.md`, `mutate.py`, `mutate_amend.py`, `mutate_wiring.py` | preuve que les tests de chaque PR échouent sans son code (première version, puis amendement du 20 octobre 2026) |
| `review_replies.md` | brouillons des réponses au relecteur (20 octobre 2026) |

## Branches locales

| Réf. | Branche | Base | Contenu | Taille |
| :- | :- | :- | :- | :- |
| GC-04a | `gc-04a-authz-tool-rules` (`8b8163e`) | `origin/main` | `pkg/authz` : `tools.go`, `rows.go` et leurs tests | 4 fichiers, +677 / -19 |
| GC-04b | `gc-04b-grpc-confined-tools` (`49263e7`) | `gc-04a-authz-tool-rules` | `pkg/rpcserver` : `ownership.go`, `tools.go` et leurs tests | 5 fichiers, +286 / -30 |
| GC-04c | `gc-04c-rest-confined-tools` (`c18bc00`) | `gc-04a-authz-tool-rules` | `pkg/httpapi` : `policy.go` et son test | 2 fichiers, +126 / -6 |

La somme des trois est identique au commit écrasé `4ae68da` rebasé sur v2.134.0 (`84a2739`, tag `backup/gc-04-2026-10-19`), plus un cas de test ajouté le 19 octobre 2026 (autre graphie dans une copie imbriquée). Preuve « failing first » : `failing_first.md`. 04b et 04c dépendent de 04a seulement, pas l'une de l'autre.

Sûreté de l'état intermédiaire : après 04a seule, `AuthorizeRows` refuse toujours `CallTool` à une clé confinée et `policy.go` refuse toujours la route des outils. Rien ne s'ouvre avant le câblage.

## Ordre

1. **GC-04a** : PR #72 ouverte après suite complète `-race` verte et test `eval` vert.
2. **GC-04b** et **GC-04c** : PR #73 et #74 ouvertes le même jour, empilées sur #72, sur demande de Lauro (le propriétaire attend les trois ensemble). Après la fusion de #72 : `git rebase origin/main` de chaque branche et push forcé vers `fork`.

## Poussée de l'amendement (20 octobre 2026, 19 h 50, sur commande de Lauro)

Les trois PR repassées en brouillon, puis poussées depuis leur copie de travail (`git push --force-with-lease=<branche>:<ancienne tête> fork <branche>`, hook `precommit` vert, 0 bloc sans test) : #72 `8b8163e` → `3c93d00`, #73 `49263e7` → `49949eb`, #74 `c18bc00` → `4b3ff4c`. Corps mis à jour (`gh pr edit --body-file`), un commentaire de réponse par PR (`review_replies.md`). GC-10 poussée et ouverte en brouillon dans le même lot : PR #78. Codecov sur les trois têtes : « 100.00% of diff hit » ; `gh pr ready` passé sur #72, #73, #74 à 19 h 50. Le commentaire sous la discussion #66 part avec le lot C.

## Commandes (depuis la racine `granular_cortex`, préfixe `!`, une ligne chacune)

GC-04a :

```
cd cortexdb && git push fork gc-04a-authz-tool-rules && gh pr create -R liliang-cn/cortexdb --head lfoletti:gc-04a-authz-tool-rules --title "feat(authz): rules that let a confined key call a tool, not yet wired" --body-file ../docs/pr/lot_B/GC-04a.body.md
```

GC-04b, après fusion de #72 :

```
cd cortexdb && git checkout gc-04b-grpc-confined-tools && git fetch origin && git rebase origin/main && git checkout proto/granular-cortex
```

```
cd cortexdb && git push fork gc-04b-grpc-confined-tools && gh pr create -R liliang-cn/cortexdb --head lfoletti:gc-04b-grpc-confined-tools --title "feat(rpcserver): a confined key can call the toolbox within its scope over gRPC" --body-file ../docs/pr/lot_B/GC-04b.body.md
```

GC-04c, après fusion de #72 (remplacer `#NB` dans le corps) :

```
cd cortexdb && git checkout gc-04c-rest-confined-tools && git fetch origin && git rebase origin/main && git checkout proto/granular-cortex
```

```
cd cortexdb && git push fork gc-04c-rest-confined-tools && gh pr create -R liliang-cn/cortexdb --head lfoletti:gc-04c-rest-confined-tools --title "feat(httpapi): a confined key can call the toolbox within its scope over REST" --body-file ../docs/pr/lot_B/GC-04c.body.md
```
