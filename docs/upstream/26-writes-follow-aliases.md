# 26. Les écritures et `find_nodes` suivent les alias

**Statut** : envoyée le 18 novembre 2026, PR #222 en brouillon (demande de Lauro). Branche `co-35-upsert-follows-aliases` du fork lfoletti/cortexdb (tête b161411), base `main` 654bf0a (v2.156.1). **Taille** : 198 lignes hors tests, 282 lignes de tests. **Dépend de** : rien ; elle sert la PR A de la fiche 13 (`co-21-merge-entities`, tête 3a1380c), qui note désormais les noms fondus là où cette PR les lit. **Motivée par** : l'ingestion par fiche de kg (chantier K2 d'alignement, décision du 16 novembre 2026 : « éviter les opérations rationnellement évitables, comme de réécrire des fiches qui ne changent pas de contenu »).

## Le défaut

La résolution des entités (`cortexdb-mcp-stdio --resolve-entities`, et `merge_entities` de la fiche 13) fond une entité dans une autre. Elle note les noms fondus sur l'entité gardée, dans `properties.aliases`. Le rappel suit ces alias (`resolveEntityNameToNode`, pour `cortex_query`, `build_context`, la mémoire et `verify_claims`). Les écritures ne les suivent pas :

- `upsert_entities` dérive l'identifiant du nom (`graphEntityNodeID`). Un ancien nom recrée l'entité fondue à côté de celle qui l'a reçue.
- `upsert_relations` dérive ses deux extrémités de la même façon. Une relation qui nomme une entité fondue est refusée par la clé étrangère.
- `find_nodes` compare les noms au contenu des nœuds. Il répond qu'un nom fondu est inconnu.

Conséquence pour kg : son pont écrit chaque entité sous le nom résolu par la table d'alias du client. Une fusion change donc ce nom dans des centaines de documents, qui se réécrivent en entier. Mesure d'alignement (K1, 16 novembre 2026) : 607 fiches inchangées réécrites en une nuit pour cette raison.

## La correction

- Une seule fonction (`followEntityAliases`) résout par lot les noms d'une requête. Elle lit d'abord les nœuds dérivés des noms. Un nom qui garde son nœud y reste écrit. Pour un nom sans nœud, elle lit une fois les entités qui portent des alias. Un identifiant donné par l'appelant est écrit tel quel.
- Une entité écrite sous un alias arrive sur l'entité qui le porte. Celle-ci garde son nom, son contenu, sa description, son type et son vecteur. Elle gagne le document et les mentions de l'écriture.
- `upsert_relations` suit l'alias aux deux extrémités.
- `find_nodes` répond à un nom fondu par l'entité qui le porte, avec `match: "alias"`, quand les passes exacte, normalisée et par inclusion n'ont rien trouvé.
- Un index partiel (`idx_graph_nodes_aliased`, condition `properties LIKE '%"aliases"%'`) porte la lecture des alias. Sur une copie d'ai-governance (411 066 nœuds), la lecture prenait 0,6 s sans l'index et moins de 0,1 ms avec. Elle a lieu au plus une fois par appel, et seulement quand un nom ne trouve pas de nœud.
- Deux entités qui portent le même alias : le premier identifiant dans l'ordre le garde.

Aucune ouverture nouvelle pour une clé confinée : les entités sont partagées entre collections, et une clé confinée qui nomme l'entité gardée l'atteint déjà.

## Tests, écrits avant le code

Quatre tombaient avant le code : `TestUpsertEntitiesUnderAnAliasWritesTheCanonicalEntity`, `TestUpsertRelationsFollowTheAliasAtEitherEnd`, `TestFindNodesAnswersAnAliasWithTheCanonicalEntity`, `TestTheAliasReadUsesThePartialIndex`. Deux gardent le comportement voisin : `TestUpsertEntitiesKeepsANodeThatStandsUnderTheName`, `TestAnAliasTwoEntitiesCarryGoesToTheFirstID`.

Mutations, chacune prise par un test : aucun alias suivi ; l'alias avant un nœud vivant ; l'identité de l'entité gardée écrasée ; l'une ou l'autre extrémité d'une relation non suivie ; `find_nodes` sans la passe des alias ; pas d'index partiel ; le dernier porteur d'un alias le garde ; une ligne illisible qui arrête la lecture ; un identifiant typé suivi ; un identifiant donné suivi. Couverture du correctif mesurée en local : aucune ligne non couverte.

Suites `pkg/cortexdb`, `pkg/graph`, `pkg/authz`, `pkg/rpcserver` : vertes. `go test ./...` : seul `pkg/liveview` (`TestHashedAssetsAreImmutableTypedAndCompressedOnce`) tombe, et il tombe aussi sur `main` 654bf0a sans la PR.

## Ce que kg fait ensuite (newskit)

Le pont écrit chaque entité sous le nom extrait, pas sous le nom résolu. L'empreinte d'une fiche ne dépend plus de la table d'alias. Une fusion passe par `merge_entities` (fiche 13, A) : elle réécrit l'entité gardée et ses arêtes, pas les documents.
