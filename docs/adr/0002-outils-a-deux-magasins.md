# ADR 0002 : outils à deux magasins, porte forcée ou refus

- Statut : acceptée le 20 octobre 2026, en relecture amont (PR #72 à #74).
- Date : 20 octobre 2026.

## Contexte

Une clé confinée à une collection ne doit atteindre que la connaissance de cette collection. Une clé confinée à un utilisateur ne doit atteindre que la mémoire de cet utilisateur. L'API typée applique déjà cette règle (`PermitsKnowledge`, `PermitsMemory`). Plusieurs outils lisent ou écrivent les deux magasins : `knowledge_memory_recall`, `knowledge_memory_build_context_pack`, `knowledge_memory_consolidate`, `promote_to_knowledge`, `reflect`, `cortex_query`.

## Décision

1. Un outil à deux magasins est admis seulement s'il a une porte booléenne qui coupe l'autre magasin. La porte est forcée dans les arguments avant l'appel (`ConfineToolArgs`) : `disable_memory=true` pour une clé de collection, `disable_knowledge=true` pour une clé d'utilisateur, de scope ou de namespace, quoi que le client ait envoyé. `recall` et `build_context_pack` sont dans ce cas.
2. Un outil à deux magasins sans porte propre est refusé : `consolidate`, `promote_to_knowledge`, `reflect`.
3. `cortex_query` est refusé tant que sa voie graphe peut servir des chunks d'une autre collection. Le correctif de la voie (GC-10) est une PR à part ; l'outil entre dans la liste après, argument par argument.
4. Les identifiants de chunks et d'arêtes (`entities[].chunk_ids`, `relations[].support_edge_ids`) sont refusés : un chunk relie le graphe partagé à n'importe quelle collection, et la garde ne charge pas les chunks.

## Options écartées

- Admettre `consolidate` avec des identifiants imbriqués (`promotion.memory_ids`, `promotion.knowledge_id`) et des portes sur `reflect.recall.*` : possible, mais trois magasins et un graphe dans un seul appel. Refusé en attendant un usage réel.
- Corriger `cortex_query` dans #73 : mélange une règle de droits et un correctif de requête.

## Conséquences

- Un client confiné ne peut pas consolider ni promouvoir : ces opérations restent à la clé non confinée, donc au serveur ou à un opérateur.
- Les réponses au relecteur et le corps de #72 disent chaque refus avec sa raison.

## Références

- PR amont #72, #73, #74 ; branche `gc-10-cortex-query-prefetch` (dossier `../thalamus_local/docs/pr/lot_D/`).
