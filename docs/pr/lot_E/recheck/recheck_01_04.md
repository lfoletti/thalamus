# Fiches amont 01 et 04 revues sur CortexDB v2.138.0

Date : 23 octobre 2026. Arbre lu : `/Users/mini/Documents/granular_cortex/cortexdb`, `b2a1ab6` (tag `v2.138.0`, branche `proto/granular-cortex`). Copie de travail jetable : `git worktree` détachée sur `v2.138.0`, supprimée à la fin. Rien n'a été poussé, commité ni modifié dans le dépôt. `CORTEXDB_TEST_POSTGRES` n'a jamais été posé.

Fichiers d'essai gardés hors du dépôt, pour rejouer : `/private/tmp/claude-501/-Users-mini-Documents-granular-cortex/e7588c43-d5e6-40d4-b90c-d4984f016177/scratchpad/recheck_01_04_files/` (`rpcserver_zz_reload_demo_test.go`, `httpapi_zz_reload_demo_test.go`, `zz_extract_scope_demo_test.go`, `zz_source_proto.go`, `zz_source_bench_test.go`). Pour rejouer : les copier dans `pkg/rpcserver/`, `pkg/httpapi/`, `pkg/cortexdb/` et `pkg/authz/` d'une copie sur `v2.138.0`, en retirant le préfixe `rpcserver_` ou `httpapi_` du nom.

---

## Fiche 01 : rechargement du fichier de clés

### Faits (v2.138.0)

1. **Chargement.** `authz.Load(path)` existe toujours et garde son nom : `pkg/authz/authz.go:316` (`os.ReadFile` puis `Parse`). `Parse` (`authz.go:295`) refuse les champs inconnus, les données en trop et un fichier sans clé (`authz.go:310-311`, « no keys declared »). `authz.Resolve(keyFile, token)` (`authz.go:350-355`) appelle `Load` quand un chemin est donné, sinon `LegacyToken`.
2. **Le jeu de clés est figé à la construction, des deux côtés.**
   - gRPC : `rpcserver.NewWithPolicy` (`pkg/rpcserver/server.go`, fonction `NewWithPolicy`) appelle `authz.Resolve` une fois, puis `authInterceptor(keys, db)` capture le `*authz.KeySet` (`pkg/rpcserver/auth.go:29-38` : `keys.Enabled()` puis `keys.Lookup(secret)` sur la valeur capturée).
   - REST : `httpapi.NewWithPolicy` (`pkg/httpapi/server.go:65-82`) appelle `authz.Resolve` une fois, puis `authenticate(keys, rt)` (`pkg/httpapi/policy.go:43-67`). `authenticate` teste `keys.Enabled()` **une seule fois, à la construction** (`policy.go:44-46`) et rend alors le routeur sans garde.
3. **API publique qui porte un `*authz.KeySet` :** `rpcserver.Options.Keys` (`server.go:32`), `httpapi.Options.Keys` (`server.go:44`), et la fonction exportée `rpcserver.AuthInterceptor(keys *authz.KeySet, db)` (`pkg/rpcserver/auth_export.go:21`). Changer ces types casserait des appelants : la fiche, qui proposait de « prendre un `*authz.Source` à la place du `*authz.KeySet` », est obsolète sur ce point.
4. **Câblage.** **`cmd/cortexdb-rest` n'existe pas** en v2.138.0 (`ls cmd` : `cortexdb-bench`, `cortexdb-connector-mcp`, `cortexdb-grpc`, `cortexdb-mcp-stdio`). La face REST est servie par `cortexdb-grpc` lui-même, option `-http` : `cmd/cortexdb-grpc/main.go:73` (option `-keys`, variable `CORTEXDB_KEY_FILE`), `main.go:150-153` (`rpcserver.NewWithPolicy(db, Options{Token, KeyFile: *keys, …})`), `main.go:169` puis `main.go:205-206` (`startHTTP` → `httpapi.NewWithPolicy(db, Options{Token, KeyFile: keyFile, …})`). Le chemin passe déjà par `Options.KeyFile` : la commande n'a besoin d'aucun changement pour profiter du rechargement. Chaque face construit son propre jeu de clés depuis le même chemin.
5. **Aucun rechargement.** Aucune occurrence de `reload`, `ModTime` ou `SIGHUP` dans `pkg/authz`, `pkg/rpcserver`, `pkg/httpapi`, `cmd/cortexdb-grpc` (code hors tests). Le `CHANGELOG.md` ne mentionne aucun rechargement du fichier de clés.
6. **Preuve par essai** (copie jetable, tests de démonstration qui échouent sur v2.138.0) :
   - gRPC (`TestDemoARevokedKeyStillWorksOverGRPC`, bufconn, `Admin.Health`) : fichier réécrit sans la clé `a` et avec une clé `c`. Résultat : `a` → `OK` (attendu `Unauthenticated`), `c` → `Unauthenticated` (attendu `OK`).
   - REST (`TestDemoARevokedKeyStillWorksOverREST`, `GET /v1/health`) : `a` → 200 (attendu 401), `c` → 401 (attendu 200).
7. **Côté Thalamus** (pour le risque) : `gateway/keys/keys.go:76` valide le fichier par `authz.Parse` avant d'écrire, et écrit par fichier temporaire puis `os.Rename` (`keys.go:81-83`) : chaque écriture produit un nouvel inode. Le gateway n'importe de CortexDB que `pkg/authz`, `pkg/rpc` et `pkg/rpc/v1` ; il n'appelle ni `AuthInterceptor` ni `Options.Keys`.

### Mesures (Apple M2, Go 1.26.5, darwin/arm64, `-count=3`)

Prototype de `Source` (même forme que `gateway/rights/source.go`, mutex tenu pendant le `stat`) et variante avec le `stat` hors du verrou. Un seul processus, fichier dans `t.TempDir()` (APFS).

| Banc | ns/op | alloc |
| :- | -: | :- |
| `os.Stat` seul, série | 2 958 à 2 981 | 304 B, 2 |
| `Lookup` figé, 1 clé | 17,8 à 19,2 | 0 |
| `Lookup` figé, 100 clés | 1 342 à 1 358 | 0 |
| `Lookup` figé, 1 000 clés | 13 456 à 13 624 | 0 |
| `Source.Current().Lookup`, 1 clé | 3 139 à 3 177 | 320 B, 2 |
| `Source.Current().Lookup`, 100 clés | 4 520 à 5 198 | 320 B, 2 |
| `Source.Current().Lookup`, 1 000 clés | 18 573 à 19 125 | 320 B, 2 |
| `Admin.Health` gRPC authentifié par bufconn (référence, code actuel) | 20 035 à 27 001 | 10 KB, 162 |
| Parallèle 8 fils, `Lookup` figé, 100 clés | 250 à 324 | 0 |
| Parallèle 8 fils, `os.Stat` seul | 1 665 à 1 684 | 304 B, 2 |
| Parallèle 8 fils, Source **mutex autour du `stat`** | 7 433 à 7 514 | 320 B, 2 |
| Parallèle 8 fils, Source **`stat` hors du verrou** | 1 896 à 2 617 | 320 B, 2 |

Lecture des chiffres (faits) : le surcoût est d'environ **3 µs par requête** en série. Sur l'appel gRPC le plus léger (`Health`, 20 à 27 µs par bufconn), c'est 11 à 15 %. Sous charge parallèle, le mutex tenu pendant le `stat` (la forme de `rights.Source`) sérialise les appels : 7,5 µs/op au total, soit un plafond d'environ 135 000 requêtes/s pour tout le serveur. Le `stat` hors du verrou revient au coût du `stat` lui-même (environ 2 µs/op au total, 400 000 à 500 000/s). Il faut donc **ne pas recopier `rights.Source` tel quel** : `stat` hors du verrou, instantané servi par `atomic.Pointer`, verrou pris seulement quand la date, la taille ou l'inode changent.

Estimations (non mesurées) : un appel réel (recherche, écriture) coûte des centaines de µs à des ms, et 3 µs y sont négligeables. Une variante avec un intervalle minimal d'une seconde entre deux `stat` (comparaison de `time.Now` en atomique) coûterait quelques dizaines de ns, au prix d'une seconde de délai sur une révocation.

### Conception révisée

**Règle unique.** Un serveur construit sur un fichier de clés (`Options.KeyFile`) applique le contenu actuel du fichier à chaque requête : une clé ajoutée sert, une clé retirée est refusée, sans redémarrage. Un fichier qui ne se lit pas ou ne se valide pas (y compris un fichier vidé, `{"keys":[]}`) laisse le jeu précédent en vigueur et le signale une fois par changement raté. `Options.Keys` (jeu préchargé) et `AuthInterceptor` restent statiques ; aucune signature publique ne change.

**Fichiers.**

| Fichier | Changement | Lignes hors tests (estimation) |
| :- | :- | -: |
| `pkg/authz/source.go` (nouveau) | `type KeySource`, `OpenKeySource(path) (*KeySource, error)` (premier chargement obligatoire), `(*KeySource).Current() *KeySet` : `os.Stat` hors verrou, comparaison date + taille + `os.SameFile` contre l'instantané ; si changé, verrou, relecture, `bytes.Equal` contre le texte précédent, `Parse`, échange par `atomic.Pointer` ; échec : jeu précédent, `Warn` une fois par clé d'échec (date et taille, ou erreur de `stat`). `Warn` nil → `log.Printf`. | 80 à 100 |
| `pkg/rpcserver/auth.go` | `authInterceptor` prend `func() *authz.KeySet` ; `AuthInterceptor` (exporté) l'appelle avec une fonction qui rend le jeu figé. | 8 à 12 |
| `pkg/rpcserver/server.go` | `NewWithPolicy` : `KeyFile` → `authz.OpenKeySource` et `src.Current` ; `Keys` ou jeton → jeu figé. L'échec du premier chargement garde `refuseAllInterceptor`. | 10 à 15 |
| `pkg/httpapi/server.go` | même aiguillage. | 10 à 15 |
| `pkg/httpapi/policy.go` | `authenticate` prend `func() *authz.KeySet` et teste `Enabled()` à chaque requête (aujourd'hui une fois, `policy.go:44`). | 8 à 12 |
| `pkg/authz/authz.go` | `Resolve` inchangé ; commentaire de `Load` qui renvoie au rechargement. | 2 à 5 |
| `cmd/cortexdb-grpc/main.go` | rien d'obligatoire ; ligne de démarrage `auth=scoped keys (<chemin>, reloaded on change)`. | 0 à 2 |
| **Total** | | **120 à 160**, sous les 400 |

La fiche annonçait 150 lignes et un câblage de deux lignes par commande dans deux commandes : il n'y a qu'une commande et elle n'a rien à changer.

**Tests à écrire d'abord** (noms proposés) :

- `pkg/authz` :
  - `TestAKeySourceServesAKeyAddedToTheFileAtTheNextLookup`
  - `TestAKeySourceRefusesAKeyRemovedFromTheFileAtTheNextLookup`
  - `TestABrokenKeyFileKeepsThePreviousKeysAndWarnsOncePerChange` (cas en table : JSON cassé, champ inconnu, `{"keys":[]}`, fichier retiré ; deux lectures par cas, un seul avertissement ; un second changement raté avertit de nouveau)
  - `TestAKeySourceRecoversWhenABrokenFileIsFixed` (clé d'échec remise à zéro)
  - `TestTheSameTextRewrittenIsNotParsedAgain` (même pointeur `*KeySet` rendu après réécriture du même texte avec une autre date)
  - `TestAFileReplacedWithTheSameSizeAndTimeIsStillReread` (nouveau fichier par `os.Rename`, `os.Chtimes` à la même date, même taille : la clé nouvelle sert)
  - `TestTheFirstLoadOfAKeySourceMustSucceed`
- `pkg/rpcserver` : `TestAKeyRemovedFromTheKeyFileIsRefusedWithoutARestart` (et une clé ajoutée sert) ; `TestAnEmptiedKeyFileDoesNotOpenTheGRPCServer` (appel sans jeton toujours `Unauthenticated`, ancienne clé toujours servie).
- `pkg/httpapi` : `TestAKeyRemovedFromTheKeyFileGets401WithoutARestart` ; `TestAnEmptiedKeyFileDoesNotOpenTheRESTServer`.
- Garde-fou de compatibilité : `TestAPreloadedKeySetIsNeverReread` (`Options.Keys` posé, fichier jamais lu).

Les tests de démonstration de `recheck_01_04_files/` en sont le premier jet pour gRPC et REST : ils tombent sur v2.138.0, c'est la preuve « failing first » sans code.

**Mutations à la main, chacune doit faire tomber au moins le test nommé :**

| Mutation | Test qui tombe |
| :- | :- |
| `Current` rend l'instantané sans `stat` (jamais relu) | `…AddedToTheFile…`, `…RemovedFromTheFile…`, les deux tests serveur |
| nouveau jeu parsé mais pas publié (`atomic.Store` retiré) | `…AddedToTheFile…` |
| sur échec, publier un jeu nil ou vide au lieu de garder l'ancien | `…KeepsThePreviousKeys…`, `…EmptiedKeyFileDoesNotOpen…` (gRPC et REST) |
| `Warn` jamais appelé | `…WarnsOncePerChange` |
| `Warn` à chaque lecture (comparaison de la clé d'échec retirée) | `…WarnsOncePerChange` |
| clé d'échec non remise à zéro après un succès | `…RecoversWhenABrokenFileIsFixed` (le second échec n'avertit plus) |
| comparaison `bytes.Equal` retirée (toujours reparser) | `…NotParsedAgain` |
| comparaison `os.SameFile` retirée | `…SameSizeAndTimeIsStillReread` |
| `NewWithPolicy` appelle `src.Current()` une fois et capture le résultat (gRPC, puis REST) | test serveur de la face mutée |
| `authenticate` teste `Enabled()` à la construction au lieu de chaque requête | aucun aujourd'hui (le jeu d'un fichier n'est jamais vide). À noter dans `failing_first.md` comme équivalent, ou à garder tel quel avec le test « emptied » comme filet. |
| `Options.Keys` relu depuis `KeyFile` quand les deux sont posés | `TestAPreloadedKeySetIsNeverReread` |

**Risques pour le relecteur.**

1. *Un `stat` par requête sur le chemin chaud* : environ 3 µs en série, 11 à 15 % de l'appel gRPC le plus léger, plafond de 135 000 req/s si le verrou entoure le `stat`, 400 000 à 500 000 req/s s'il ne l'entoure pas (mesures ci-dessus). Donner les chiffres dans le corps, et proposer la variante à intervalle minimal d'une seconde en repli, sans goroutine.
2. *Changement de comportement* : `KeyFile` était lu une fois ; un fichier valide mais faux s'applique désormais tout de suite. À écrire dans le corps (pas de section « Breaking Changes » : aucune signature ne change, mais un opérateur qui préparait le fichier à l'avance sans redémarrer est surpris).
3. *Résolution de la date* : sur un système de fichiers à la seconde (HFS+, certains montages réseau), une réécriture sur place de même taille dans la même seconde passerait inaperçue sans `os.SameFile`. La comparaison d'inode couvre l'écriture par renommage (celle de Thalamus) ; une réécriture sur place de même taille dans la même seconde reste invisible jusqu'au changement suivant. À écrire comme limite.
4. *Deux faces, deux sources* : `cortexdb-grpc -http` ouvre deux `KeySource` sur le même chemin, donc deux avertissements par changement raté (un par face). Une requête ne touche qu'une face, donc un seul `stat`. Partager une source demanderait un champ public de plus dans `Options` : hors de cette PR.
5. *Lecture d'un fichier à moitié écrit* : `Parse` échoue, l'ancien jeu reste, un avertissement ; le `stat` suivant voit la taille finale et relit. Pas de fenêtre ouverte, puisque l'échec garde l'ancien jeu et qu'un jeu vide est refusé par `Parse`.

---

## Fiche 04 : `loadSessionText` et la portée de l'appelant

### Faits (v2.138.0)

1. **`memoryBucketPermitted` n'existe plus.** Il a été introduit par `84539de` (v2.134.1) et remplacé par `5e02220` (v2.135.0, `CallerScope`). Le contrôle de lecture actuel est `(*DB).memoryReadPermitted(ctx, bucketID)` (`pkg/cortexdb/memory_owner.go:103-113`) : sans `CallerScope` ou sans `Memory`, tout passe ; un seau absent passe (« no rows to give away ») ; sinon `scope.Memory(bucket)`. Appelé par les lectures de mémoire : `pkg/cortexdb/memory_api.go:262`, `:352`, `:1375`. `WithCallerScope` est exporté (`memory_owner.go:48`) ; seul `authz.Key.CallerContext` l'appelle dans l'arbre (`pkg/authz/caller.go:22-43`).
2. **`loadSessionText` garde son nom et ignore toujours la portée.** `pkg/cortexdb/extract_conversation.go:176-205` : une requête `SELECT … FROM messages WHERE session_id = ?` (ligne 183), filtrée par la seule porte de relecture (lignes 177-181, ajoutée par `477a27d`). Aucun appel à `memoryReadPermitted`. Appelée par `ExtractConversation` (`extract_conversation.go:57`) quand `text` est vide et `session_id` donné. Joignable par `GraphRAGToolbox.Call("extract_conversation", …)` (`pkg/cortexdb/graphrag_tool_defs.go:702-707`) et par le serveur MCP (`pkg/cortexdb/mcp.go:279-285`).
3. **Le trou existe toujours sur l'API embarquée.** Essai `TestDemoExtractConversationReadsAnotherUsersSession` (copie jetable) : deux mémoires, `zeus-1` (utilisateur `zeus`) et `hermes-1` ; contexte `WithCallerScope` limité à `hermes`.
   - `GetMemory` par liste sous `hermes` : `memories=0 missing=[zeus-1]` (la lecture typée refuse).
   - `memoryReadPermitted(hermes, "memory:user:zeus:default")` : `false`.
   - `extract_conversation {"session_id":"memory:user:zeus:default"}` sous `hermes` : **réponse complète** : `summary` = « memory: Zeus hid the Thunderbolt under Olympus. Hera must not know. », `themes` (8), `entities` [Zeus, Thunderbolt, Olympus, Hera], `relations` co_occurs. Le test tombe.
   - Session inexistante : erreur `extract_conversation: text or session_id is required`.
4. **Ce qu'un appelant confiné lit par ce trou** : les deux premières phrases de la session d'un autre (400 caractères au plus), 8 thèmes, jusqu'à 30 entités et leurs cooccurrences. Avec `persist: true`, en plus, le résumé écrit comme document `conversation:<session>` dans la collection demandée (par défaut `conversations`) et les entités dans le graphe partagé.
5. **Par les serveurs, une clé confinée n'atteint pas le trou.** La liste blanche refuse `session_id` et `persist` pour `extract_conversation` (`pkg/authz/allowlist.go:211-216`), testé par `pkg/authz/allowlist_test.go:208` et `:223`. L'outil est classé `collectionScope` : une clé confinée par `user_id` se voit refuser l'outil. Une clé non confinée n'a pas de `CallerScope` (`caller.go:24-26`). **Qui est exposé** : un programme qui embarque `pkg/cortexdb`, pose `WithCallerScope` lui-même et appelle `ExtractConversation` ou `Call`. Le contrat écrit de `CallerScope` (`memory_owner.go:15-20` : « The checks are on what a request resolves to, not on its fields ») n'est pas tenu par cet outil.
6. **Le correctif ne permet pas d'ouvrir `session_id` dans la liste blanche avec le modèle actuel.** Pour une clé confinée par collection, `CallerScope.Memory` exige `scope.Permits(FieldCollection, "")` (`caller.go:34`), qui est faux : toute session lui serait refusée par la façade. Une clé confinée par utilisateur n'a pas l'outil. Le refus de la liste blanche reste donc juste ; la PR est une défense en profondeur de la façade, pas une ouverture.
7. **Essai du correctif** (copie jetable, retiré ensuite) : cinq lignes en tête de `loadSessionText` (`memoryReadPermitted` ; erreur enveloppée ; refus → texte vide). Sous `hermes`, la session de `zeus` rend alors exactement l'erreur d'une session inexistante (`text or session_id is required`). `TestExtractConversation` et `TestConversationExtractionLeavesOutAMemoryTheGateHides` restent verts.

### Conception révisée

**Règle unique.** `extract_conversation` ne lit une session que si la portée de l'appelant permet son seau (`memoryReadPermitted`). Refusée, la session se comporte comme une session vide ou inexistante : même erreur, rien d'écrit, même avec `persist`.

**Fichiers.** `pkg/cortexdb/extract_conversation.go` seul : 5 à 8 lignes, plus 2 à 3 lignes de commentaire. Environ **10 lignes hors tests**, loin des 60 de la fiche et des 400 du plafond. Pas de changement de `pkg/authz`.

Dépendances : la fiche dépendait de #72 et #73 ; les deux sont fusionnées (v2.138.0 contient toute la série). Plus de dépendance.

**Tests à écrire d'abord** (`pkg/cortexdb`, sans serveur ni embedder) :

- `TestExtractConversationAnswersASessionOutsideTheCallerScopeAsMissing` : deux utilisateurs ; sous la portée de `hermes`, la session de `zeus` rend la même erreur qu'une session inexistante, et aucun texte de `zeus` n'apparaît.
- `TestExtractConversationWithPersistWritesNothingForASessionOutsideTheScope` : `persist: true`, puis `knowledge_get` sur `conversation:<session de zeus>` rend « not found » et le graphe n'a pas `entity:thunderbolt`.
- `TestExtractConversationStillReadsTheCallersOwnSession` : la session de `hermes` sous `hermes`, et celle de `zeus` sans portée, sont résumées comme avant (garde contre « tout refuser »).
- `TestExtractConversationFailsWhenTheSessionBucketCannotBeRead` : `UPDATE sessions SET created_at = 'not a time'` (même panne que `memory_get_scope_test.go:67-73`) ; l'appel échoue au lieu de passer.

**Mutations à la main :**

| Mutation | Test qui tombe |
| :- | :- |
| garde retirée | `…OutsideTheCallerScopeAsMissing`, `…PersistWritesNothing…` |
| refus rendu par une erreur propre (« withheld ») au lieu du texte vide | `…AsMissing` (l'erreur diffère de celle d'une session inexistante) |
| erreur de lecture du seau avalée (`!ok` seul testé) | `…BucketCannotBeRead` |
| condition inversée (refuser ce qui est permis) | `…StillReadsTheCallersOwnSession` |

**Risques pour le relecteur.**

1. *Valeur* : aucun serveur n'expose le trou, puisque la liste blanche refuse `session_id`. Le relecteur peut demander pourquoi. Réponse : `WithCallerScope` est public et promet que la façade contrôle ce qu'une requête résout ; cet outil est le seul lecteur de `messages` par identifiant de session qui ne le fait pas (les autres lectures de `memory_api.go` appellent `memoryReadPermitted`). Écrire aussi que la liste blanche garde son refus, et pourquoi (fait 6).
2. *Indiscernabilité* : rendre le texte vide plutôt qu'une erreur propre, pour qu'une session d'autrui et une session absente répondent pareil. Le relecteur peut préférer une erreur « not found » explicite ; elle trahirait l'existence de la session.
3. *Coût* : un `GetSession` de plus par extraction depuis une session, seulement quand une portée est posée (`memoryReadPermitted` sort tout de suite sinon).
4. *Pour Thalamus* : rien à retirer ni à ouvrir. La PR est de faible priorité ; son intérêt est de fermer un trou connu et documenté avec un correctif de dix lignes.

---

## Ce qui est sûr, ce qui est supposé

- **Sûr (lu ou exécuté)** : les numéros de ligne cités ; l'absence de `cmd/cortexdb-rest` et de tout rechargement ; les quatre essais rouges sur v2.138.0 (gRPC, REST, extraction) ; les mesures du tableau ; le comportement du correctif de la fiche 04 essayé en copie jetable.
- **Supposé** : les estimations de lignes ; le coût de la variante à intervalle (non mesuré) ; la part du `stat` sur un appel réel autre que `Health` (non mesurée, déduite du coût des recherches) ; l'accueil du relecteur.
