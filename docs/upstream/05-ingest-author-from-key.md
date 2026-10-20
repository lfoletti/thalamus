# 05. `author` et `written_by` depuis la clé appelante

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : 80 lignes hors tests. **Dépend de** : #77 fusionnée (même règle que `reviewed_by`).

## Le manque

`ingest_document` et `knowledge_save` prennent `author` du client. Sous une clé, le client peut écrire n'importe quel nom. #77 a établi la règle pour `reviewed_by` : la clé appelante l'emporte sur ce que le client envoie. `written_by` (lot A) porte déjà l'identifiant de la clé, mais `author`, le champ que les recherches et la provenance affichent, reste libre.

## Ce que Thalamus fait aujourd'hui

`th_add` pose `author` = login de l'employé avant de relayer. Un client qui parle au serveur sans passer par Thalamus peut écrire ce qu'il veut.

## La proposition

Quand une clé appelle, `author` vide prend l'identifiant de la clé, et `author` non vide est gardé mais `written_by` reste la clé (déjà le cas). Variante plus stricte, à discuter avec le propriétaire : `author` est refusé quand il diffère de la clé, sauf pour une clé d'opérateur (clearance sans portée). La variante douce suffit à Thalamus, puisque la clé porte déjà `th/<login>/<cerveau>`.

Fichiers : `pkg/cortexdb/graphrag_tool_ingest.go`, `pkg/cortexdb/knowledge_api.go`, `written_by.go` (une fonction `authorOf(ctx, given)`).

## Tests, écrits avant le code

- `TestAnEmptyAuthorIsTheCallingKey` : sous la clé `hermes`, un ingest sans `author` stocke `hermes` ; avec `author: zeus`, stocke `zeus` et `written_by: hermes`.
- Sans clé (bibliothèque), `author` vide reste vide.
- Mutation : la règle retirée.

## Après la fusion

`th_add` n'a plus besoin de poser `author` ; il le laisse vide et le serveur écrit `th/<login>/<cerveau>`. Le gateway garde une table login ↔ clé pour l'affichage, déjà en place (`keys.Holder`).
