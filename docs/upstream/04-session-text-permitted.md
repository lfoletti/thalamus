# 04. `loadSessionText` passe par `memoryBucketPermitted`

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : 60 lignes hors tests. **Dépend de** : #72, #73 fusionnées.

## Le manque

`extract_conversation` lit le texte d'une session (`loadSessionText`, `pkg/cortexdb/extract_conversation.go`) sans passer par `memoryBucketPermitted`. Pour une clé confinée, la liste blanche de #72 ferme la porte (`session_id` classé `memoryBucket`, refusé). Pour l'API typée embarquée, la porte reste ouverte : un appelant qui porte un `user_id` peut lire la session d'un autre. C'est le point 4 du plan de correction du 20 octobre 2026, laissé hors périmètre des PR en cours.

## Ce que Thalamus fait aujourd'hui

Rien : ses clés sont confinées par collection, et la liste blanche refuse `extract_conversation` dès que `session_id` est présent.

## La proposition

`loadSessionText` prend le contexte et refuse (« not found ») une session dont le `UserID` n'est pas celui de l'appelant, par `memoryBucketPermitted`, comme `SearchMemory` depuis v2.134.1 (`WithMemoryOwners`).

## Tests, écrits avant le code

- `TestExtractConversationStaysInTheCallersSessions` : deux utilisateurs, deux sessions, l'extraction de la session de l'autre rend « not found » et n'écrit rien.
- Mutation : la garde retirée.

## Après la fusion

Rien à retirer de Thalamus. La fiche existe parce que le trou est connu et documenté dans `docs/pr/` ; le signaler sans correctif serait moins utile qu'une PR de 60 lignes.
