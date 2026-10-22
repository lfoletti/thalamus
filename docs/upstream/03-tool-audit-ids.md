# 03. Identifiants touchés dans le journal des requêtes

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : 200 lignes hors tests. **Dépend de** : #72 (liste blanche) fusionnée.

## Le besoin de Thalamus

Le journal des requêtes de `cortexdb-grpc` (lot A, PR #67) dit quelle clé a appelé quel outil et si l'appel a réussi. Pour l'audit des cerveaux sensibles, Thalamus a besoin aussi des documents ou mémoires que l'appel a touchés. Le serveur a déjà cette information : la liste blanche de #72 connaît, outil par outil, les champs porteurs d'identifiant (`ids`, chemin complet).

## Ce que Thalamus fait aujourd'hui

`gateway/relay/audit.go` et `idsOf` dans `relay.go` : à chaque relais, les champs `*_id`, `*_ids` et `supersedes` des arguments vont dans `audit.jsonl`, et le rappel fédéré ajoute les `knowledge_id` lus. C'est une copie partielle de ce que la liste blanche sait mieux.

## La proposition

Le journal des requêtes du serveur ajoute un champ `ids` par appel d'outil : les identifiants présents dans les arguments, aux chemins que la liste blanche déclare, plus, pour les outils de lecture, les identifiants rendus (`knowledge_id` des résultats de `knowledge_search`, `memory_id` de `memory_search`). Une ligne JSON par appel, même format qu'aujourd'hui.

Fichiers : `pkg/authz/allowlist.go` (une fonction `IDsOf(tool, args) map[string][]string` qui lit les chemins d'`ids`), `pkg/rpcserver/request_log.go` (le champ), `pkg/httpapi` (idem).

## Tests, écrits avant le code

- `TestIDsOfReadsEveryDeclaredPath` : pour chaque outil de la liste blanche, des arguments avec un identifiant à chaque chemin déclaré, tous retrouvés ; un chemin non déclaré ignoré.
- `TestTheRequestLogNamesWhatTheCallTouched` : `knowledge_get` journalise le `knowledge_id` ; `knowledge_search` journalise les identifiants rendus.
- Mutations : un chemin imbriqué sauté, les identifiants rendus oubliés.

## Après la fusion

Thalamus garde `audit.jsonl` pour une seule raison : le nom de l'employé, que le serveur n'a pas. `idsOf` disparaît ; la ligne d'audit du gateway cite l'identifiant de requête du serveur pour joindre les deux journaux.
