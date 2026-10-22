# Ce que la démarche ne couvre pas encore : risques relevés avant le code

Rédigé le 21 octobre 2026, après la lecture du code de v2.135.0 et avant la première ligne de l'amendement, à la demande de Lauro (« ce que nous n'avons pas anticipé et les sources d'erreurs possibles non couvertes par notre démarche »). Chaque risque dit ce qu'il est, comment il a été trouvé, et ce qui en est fait : **plan** (ajouté au plan, section indiquée), **outil** (un contrôle nouveau), ou **assumé** (consigné, à dire dans le corps de la PR).

## R1. Trois couches disent « not found » chacune à sa façon

Trouvé en lisant `memory_owner.go` de v2.135.0. Depuis cette version, la façade elle-même refuse un seau mémoire étranger (`errMemoryBucketWithheld` = « memory bucket not found »), notre garde `authz` dit `memory "X" not found`, et le magasin dit `embedding not found`. Le point 1 de #72 vise les deux derniers, mais la façade ajoute une troisième formulation que notre plan ne voyait pas. **Plan** (#72 point 1, D6 précisée) : les ports normalisent, pour une clé confinée seule, tout « not found » rendu par un outil en un seul texte, `not found`, le même que `ErrWithheld` ; une table de tests sur chaque champ d'identifiant compare le texte exact des deux réponses, pas seulement le code.

## R2. La garde lit les lignes à travers la façade qu'elle protège

Trouvé en cherchant où régler le point 1 de #72. La garde (`authz.Scope.GuardToolRows`) charge une ligne par `db.GetMemory` et `db.GetKnowledge`. Si une correction cache un jour les lignes étrangères dans ces deux lectures (une option naturelle avec `CallerScope`), la garde ne verrait plus la ligne étrangère, la prendrait pour absente et **laisserait passer l'écriture** : `knowledge_update` sur un identifiant étranger déplacerait le document. Le mécanisme qui protège dépend de celui qu'il protège. **Plan** : la façade ne cache rien par identifiant ; la garde reste la seule décision pour les écritures ; un test tient que, avec un `CallerScope` posé dans le contexte, la garde voit encore la ligne étrangère (`TestTheRowGuardSeesForeignRowsUnderACallerScope`). **Outil** : rien, c'est une règle de conception à écrire dans `rows.go`.

## R3. Une liste d'identifiants expurgée trahit l'ordre

Trouvé en concevant la réponse « `missing` » du point 1 de #72. Si la garde retire les identifiants étrangers d'une liste (`memory_get.memory_ids`) et que l'outil rend `missing` pour les absents, la réponse ne contient plus l'identifiant étranger nulle part, ni dans les résultats ni dans `missing` : l'oracle reste. Si le port l'ajoute à `missing` à la fin, l'**ordre** de `missing` trahit encore (un absent puis un étranger se lisent dans l'ordre inverse de la requête). **Plan** : `missing` est reconstruit dans l'ordre de la requête. Un test compare `missing` pour `[étranger, absent]` et `[absent, étranger]` avec ce qu'une clé non confinée obtient pour deux absents.

## R4. `plan.filters.document_ids` n'était pas un trou, et nous en avons fait un oracle

Le tour 1 disait « held up fine: `plan.filters.document_ids` across collections » : le filtre de collection suffisait. Notre amendement a classé le champ comme identifiant, donc une recherche qui nomme un document étranger est devenue une erreur, là où un document absent rend un résultat vide. Nous avons créé l'oracle en voulant fermer une classe. **Plan** (#72 point 1) : le champ redevient une donnée ; un test tient qu'un document étranger nommé dans le filtre ne rend aucun chunk (le filtre de collection, forcé, suffit). Leçon pour la méthode : fermer une classe peut ouvrir une autre classe (ici « réponses distinguables ») ; l'inventaire doit lister les classes **des réponses** autant que celles des entrées.

## R5. PostgreSQL n'a jamais été exécuté

Les deux index d'expression (`pkg/core/store_postgres.go`), et la transaction de dédup (`SELECT … FOR UPDATE`) à venir, n'ont jamais tourné sur PostgreSQL : pas de serveur local, pas de Docker. Les lignes sont « déclarées » dans `## Codecov`, ce qui dit au relecteur qu'elles sont non testées, pas qu'elles sont justes. Une expression d'index non immuable (`::jsonb` sur `text`) est refusée par PostgreSQL à la création : la base entière ne s'ouvrirait plus. L'amont a un harnais (`internal/pgtest`, `CORTEXDB_TEST_POSTGRES`) et neuf paquets de tests de parité. **Outil** : PostgreSQL 17 et pgvector installés par Homebrew le 21 octobre 2026 ; `make pgtest` lance les tests de parité avec un DSN local ; les tests de nos DDL y passent avant la poussée. Si l'installation échoue : assumé, dit dans le corps.

## R6. L'arbre fusionné n'est testé que par les bancs

Le propriétaire fusionne les six PR localement et sonde l'arbre fusionné. Nous testons chaque branche et, sur `proto`, seulement `go build` et les bancs de bout en bout. La suite complète `-race` sur `proto` n'est pas dans la chaîne. **Plan** (section 3, étape 6) : `go test -race ./...` sur `proto` sous le canari, en plus des bancs.

## R7. Les champs explicites n'existent que sur la surface des outils

D2 ajoute `expires_at`, `review_status`, `reviewed_by`, `reviewed_at` aux requêtes des outils (JSON) et de la façade Go. Les RPC typés (`SaveKnowledge`, `UpdateMemory`, `/v1/knowledge`) sont des messages protobuf : y ajouter des champs demande de régénérer `pkg/rpc/v1` et les trois clients (`clients/node`, `python`, `rust`), hors taille et hors sujet. Le tour 1 avait mis les clés dans `metadata` précisément parce que `metadata` existe sur toutes les surfaces. Conséquence : par les RPC typés, on ne peut ni poser une expiration ni décider d'une relecture. **Assumé** : dit dans les corps de #76 et #77 (« the typed RPCs carry no review or expiry field; they go through the tools, as the layer above does »). Thalamus passe par `CallTool`, rien ne casse.

## R8. Fermer `aggregate_metadata` casse `th_pending`

`gateway/relay/write.go` liste les documents proposés par `aggregate_metadata` avec `filters: {review_status: proposed}`. Avec la porte appliquée à `aggregate_metadata` (point 2 de #76 et #77), les lignes proposées sont cachées, sauf si un filtre explicite sur le statut les demande. Si la porte ne lit pas ce filtre, `th_pending` ne liste plus rien et la curation s'arrête, en silence. **Plan** (#77 point 3) : un filtre explicite `review_status` dans `filters` ou `plan.filters` l'emporte sur la porte, sur `aggregate_metadata` comme sur `knowledge_search` ; test côté moteur, et le banc `tests/e2e_gateway.sh` (55 contrôles, dont la liste des documents proposés) le tient côté gateway. Le renommage `_review_status` touche aussi `write.go` (trois endroits) et `th_review`.

## R9. Retirer une expiration devient impossible

Avec D1 d (les clés réservées sont reportées quoi que fasse le client), un client qui veut **retirer** l'expiration d'un document n'a plus de moyen : omettre le champ garde la valeur, et `metadata` ne la porte plus. **Plan** (D2 précisée) : `expires_at` est un pointeur dans les requêtes Go (`*string`) ; absent = inchangé, `""` = plus d'expiration. Même règle pour les trois champs de relecture, avec la règle du tour 1 maintenue : vider un statut sur une ligne relue est refusé. Test : poser puis retirer.

## R10. Les valeurs posées par le moteur peuvent venir d'un client par la porte de derrière

Une clé réservée arrivée dans `metadata` est retirée (D1 c). Mais les graphies : `_Expires_At`, `_EXPIRES_AT`, ou un `_` Unicode (U+FF3F) ne sont pas notre clé et passent telles quelles, inoffensives (jamais lues). En revanche `encoding/json` lit les clés de struct sans tenir compte de la casse : un champ explicite `Review_Status` dans le JSON d'une requête d'outil alimente `ReviewStatus`. C'est voulu (c'est le champ explicite), mais la liste blanche de #72 doit classer le champ pour une clé confinée, et le relecteur l'essaiera avec une graphie. **Plan** (#72 point 4, sur `proto` seulement) : classement des nouveaux champs ; test de graphies sur les champs explicites.

## R11. Les corps de PR grossissent au-delà de la tolérance

#76 et #77 reçoivent chacun des champs explicites sur six requêtes, le report des clés, la porte sur quatre outils, le retrait du cache. Le compte hors tests passera probablement 600 lignes. **Assumé** : compte dit dans le corps. Si #77 dépasse 650, le retrait du cache (D5) et la porte sur les quatre outils vont dans #76 seul, et #77 n'ajoute que sa clé à la porte (c'est déjà la structure : la porte est commune).

## R12. Le temps dans les tests

Les tests d'expiration posent des instants passés ou futurs ; la suppression du cache de 15 s retire une source de non-déterminisme, mais `reviewed_at = now` reste. Un test qui compare deux `now` à la seconde peut tomber à minuit ou sous charge. **Plan** : les tests fixent l'horloge par un champ `now func() time.Time` sur `DB` (existe-t-il déjà ? à vérifier ; sinon injecter), jamais `time.Sleep`.

## R13. Le contexte de cette session

Le travail est long : six PR, trois niveaux, les bancs, le gateway. Une session qui se compresse perd l'état fin (quel point est fait, quelle mutation est consignée). **Outil** : `docs/pr/round2/REPORT.md` est tenu à jour **après chaque point**, pas à la fin, avec une ligne « fait / en cours / à faire » par point, pour qu'une reprise reparte de là.

## R14. Ce que l'analyse ne peut pas voir

Le propriétaire connaît des paquets que nous n'avons jamais lus (importflow, connector, agentmem, liveview) et des usages (plugins Obsidian, OpenClaw, DSH) qui lisent les métadonnées des documents. Le renommage des clés en `_…` peut toucher un lecteur de `content_sha256` ou de `expires_at` hors de `pkg/cortexdb`. **Outil** : avant la poussée, `git grep` des sept noms de clés sur tout l'arbre de `proto`, hors `pkg/cortexdb` ; chaque occurrence est classée dans `REPORT.md`.

## Leçons pour la méthode (à verser dans `docs/pipeline_pr.md` après l'amendement)

1. L'inventaire doit lister les classes des **réponses** (textes d'erreur, champs de liste, ordre) et pas seulement celles des entrées (R1, R3, R4).
2. Une couche qui vérifie en lisant par une autre couche crée une dépendance : la nommer dans le code et la tenir par un test (R2).
3. Chaque ligne « déclarée » dans `## Codecov` est une ligne non exécutée : chercher d'abord un moyen de l'exécuter (R5).
4. Un changement de nom de clé est un changement d'API pour tout lecteur de la base : grep de l'arbre entier, pas du paquet (R8, R14).
