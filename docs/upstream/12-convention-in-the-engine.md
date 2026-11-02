# 12. Ce que la convention d'un cerveau demande au moteur

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : quatre PR, 150 + 120 + 100 + 150 lignes hors tests. **Dépend de** : #75 (empreinte du contenu), #77 (`review_status`) fusionnées. **Motivée par** : la convention des cerveaux de Thalamus (#35, `gateway/convention/`, `gateway/relay/write.go`, `followsConvention`).

## Le partage

La convention reste dans Thalamus : quels thèmes, quelles natures, qui déclare un thème, qui tranche une proposition, ce sont des règles d'organisation, déclarées dans `rights.yaml` à côté des lecteurs et des curateurs. Le moteur n'a pas à savoir ce qu'est un thème. Il reçoit ce qui est mécanique et vaut pour tout client : écrire sous condition, refuser une métadonnée manquante, voir ce qui attend la relecture, trouver un texte presque identique. Les quatre PR ci-dessous retirent chacune une limite que le corps de #35 nomme (« Known limits »).

## A. Écriture conditionnelle (la plus utile)

**Le besoin.** Thalamus lit puis écrit : la recherche des copies proches et la lecture des thèmes déclarés précèdent `ingest_document`. Deux écritures simultanées de copies proches, ou de `theme:x` et `theme:xs`, passent toutes les deux. `th_retheme` à côté d'une réécriture peut remettre le texte qu'il a lu.

**La proposition.** `ingest_document` accepte `if_absent: true` (refus si l'id existe) et `if_sha256: "<empreinte>"` (refus si l'empreinte du contenu en place, celle de #75, a changé depuis la lecture). Le contrôle et l'écriture se font dans la même transaction. Erreur typée `FailedPrecondition`.

**Tests, écrits avant le code.** Deux écritures concurrentes de même id avec `if_absent` : une seule passe. Une réécriture avec une empreinte périmée : refusée, le contenu reste. Mutations : condition ignorée, contrôle hors transaction.

**Après la fusion.** `th_add` d'un document nouveau passe `if_absent`, `th_retheme` passe l'empreinte lue. La course disparaît sans verrou dans le gateway.

## B. Métadonnées requises à l'ingestion

**Le besoin.** Le gateway contrôle la nature, la date et le thème. Un client qui écrit au moteur sans passer par le gateway (un producteur de synthèses, un import) n'est pas contrôlé.

**La proposition.** Un schéma de métadonnées par collection, déclaré par une clé d'administration : clés requises, listes de valeurs fermées, format (`date` = `AAAA-MM-JJ`). `ingest_document` refuse un document hors schéma, quel que soit le client. Le moteur a déjà l'ontologie et SHACL pour le graphe ; il manque l'équivalent pour les métadonnées d'un document.

**Après la fusion.** Thalamus pousse au moteur la liste de natures du cerveau et retire `natureAndDate` ; la règle vaut aussi pour les écritures qui ne passent pas par lui.

## C. Les propositions visibles aux contrôles

**Le besoin.** Sous la porte de relecture, le moteur cache les lignes proposées à `knowledge_search` et à `aggregate_metadata`. Une copie proche ou un thème proche qui attend la relecture n'est pas vu, et deux propositions semblables passent.

**La proposition.** Un argument `include_proposed: true` sur `knowledge_search` et `aggregate_metadata`, réservé à une clé de relecture (celle qui peut appeler `review_pending`, fiche 02). Les lignes rendues portent leur `review_status`.

**Après la fusion.** `followsConvention` et `declaredThemes` incluent les propositions ; `th_review accepted` n'a plus de contrôle à refaire.

## D. Les textes presque identiques

**Le besoin.** Pour trouver une copie proche, Thalamus fait une recherche lexicale sur les 40 premiers mots, lit les 5 candidats en entier et calcule l'indice de Jaccard des trigrammes de mots. Cela coûte une recherche et 5 lectures à chaque écriture qui ramène des candidats.

**La proposition.** `near_duplicates {content, threshold, limit}` : le moteur garde une empreinte MinHash des trigrammes de mots de chaque document (calculée à l'ingestion, à côté de l'empreinte exacte de #75) et rend les documents dont la ressemblance estimée dépasse le seuil, avec la valeur. Lecture seule, dans la liste blanche avec `scope: collection`.

**Après la fusion.** `Likeness`, `words`, `grams` et la boucle de lecture des candidats disparaissent de Thalamus (environ 60 lignes) ; un appel au lieu de six.

## Vérifié dans le code (v2.142.0)

- **A** est préparée sur la branche `co-19-ingest-conditional` : `if_absent` et `if_sha256` sur `ingest_document`, contrôlés avant le plan puis dans la transaction d'écriture, à côté de la recherche de `dedupe` (#75) ; sur PostgreSQL, un verrou consultatif sur l'id ; `ErrPreconditionFailed` rendu `FailedPrecondition` ; les deux arguments classés dans la liste blanche. Le RPC typé `InsertGraphDocument` n'en a pas : il faudrait changer le proto.
- `knowledge_update` ne change pas les métadonnées seules : dès que `metadata` est donné, il redécoupe le document (`replaceArtifacts`, `knowledge_api.go`). `th_retheme` ne peut donc pas garder les morceaux et la relecture : c'est une cinquième PR (**E**, mise à jour des métadonnées du seul enregistrement, sans redécoupage, pour les clés que `recordOnlyKey` ne recopie pas déjà).
- **C** touche le prédicat de la porte de relecture (`review.go`, `reviewGate`), lu par la recherche et par `aggregate_metadata` : l'argument doit être refusé à toute clé qui n'est pas de relecture, et la liste blanche le classer.
- L'ordre proposé : A, puis C, puis B, D et E. A est la seule qui corrige un défaut réel ; les autres retirent du code ou étendent la règle.
