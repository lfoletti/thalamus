# 02. Outil `review_pending`

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : 120 lignes hors tests. **Dépend de** : #77 (`review_status`) fusionnée.

## Le manque

Sous la porte de relecture (`CORTEXDB_REVIEW_GATE=1`), un document proposé n'est servi par aucune recherche. Un curateur n'a donc aucun outil pour voir ce qui attend. `knowledge_search` avec le filtre `review_status: proposed` exige un texte de requête (« empty text provided ») et ne liste pas. Le contournement passe par `aggregate_metadata` (groupé par `document_id`) puis un `knowledge_get` par document : N + 1 appels, et `knowledge_get` refuse l'argument `collection` d'une clé confinée (constat du banc, 20 octobre 2026).

## Ce que Thalamus fait aujourd'hui

`th_pending` (`gateway/relay/write.go`, `listPending`) : les N + 1 appels, un extrait de 200 caractères par document, une ligne d'audit avec tous les identifiants.

## La proposition

Un outil `review_pending {collection?, limit?}` dans la toolbox, qui rend `[{knowledge_id, title, author, created_at, snippet}]` des documents à `review_status: proposed`, dans l'ordre de dépôt, sous la porte de relecture ou non. Lecture seule (`Mutates: false`). Dans la liste blanche de #72 : `scope: collection`, `stores: knowledge`, `args: limit`. Le document proposé reste invisible à toute recherche : seul cet outil le montre, et seulement à une clé qui atteint sa collection.

Fichiers : `pkg/cortexdb/review_pending.go` (nouveau : la requête SQL sur l'index d'expression `review_status` de #77), `pkg/cortexdb/knowledge_memory_tooldefs.go` (définition), `pkg/authz/allowlist.go` (une ligne).

## Tests, écrits avant le code

- `TestReviewPendingListsTheProposedDocuments` : trois documents (proposé, accepté, sans statut), un seul rendu, avec titre, auteur et extrait ; `limit` honoré ; ordre de dépôt.
- `TestReviewPendingStaysInTheCollection` : deux collections, une clé confinée ne voit que la sienne (`pkg/rpcserver`).
- `TestEveryToolAndEveryArgument…` de #72 tombe tant que l'outil n'est pas classé.
- Mutations : filtre retiré (tout est listé), `limit` ignoré, extrait coupé au milieu d'un mot.

## Après la fusion

`th_pending` appelle `review_pending` une fois ; `listPending` et `snippet` disparaissent de Thalamus (environ 60 lignes).
