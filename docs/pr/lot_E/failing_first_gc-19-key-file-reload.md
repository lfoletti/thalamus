# Failing first : gc-19-key-file-reload

Tête : `gc-19-key-file-reload` (copie `scratchpad/wtE/gc-19-key-file-reload`), base `origin/main` `a591ae1`. Un commit. Le 23 octobre 2026.

## Tests posés sur la base, sans le code

Les fichiers de test des serveurs compilent sur `a591ae1` (ils n'utilisent que `Options.KeyFile` et `Options.Keys`). Commande : copier `pkg/rpcserver/key_file_reload_test.go` et `pkg/httpapi/key_file_reload_test.go` sur une copie de `a591ae1`, puis

```
go test ./pkg/rpcserver/ ./pkg/httpapi/ -run 'KeyFile|Preloaded' -count=1
```

Sortie observée :

```
--- FAIL: TestAKeyRemovedFromTheKeyFileIsRefusedWithoutARestart (0.02s)
    key_file_reload_test.go:63: a key removed from the file: OK, want Unauthenticated
    key_file_reload_test.go:66: a key added to the file: Unauthenticated, want OK
--- FAIL: TestAKeyRemovedFromTheKeyFileGets401WithoutARestart (0.02s)
    key_file_reload_test.go:47: GET /v1/health as c-secret = 401, want 200
    key_file_reload_test.go:47: GET /v1/health as a-secret = 200, want 401
```

Sur la base, `TestAnEmptiedKeyFileDoesNotOpenTheGRPCServer`, `TestAnEmptiedKeyFileDoesNotOpenTheRESTServer`, `TestAPreloadedKeySetIsNeverReread` et `TestAPreloadedKeySetIsNeverRereadOverREST` passent : ce sont des gardes (le fichier n'est jamais relu sur la base), tenues par les mutations M3b, M15 et M16. Les tests de `pkg/authz/key_source_test.go` et `pkg/rpcserver/auth_export_source_test.go` ne compilent pas sur la base (`KeySource`, `OpenKeySource`, `StaticKeySource`, `ResolveSource`, `AuthInterceptorFromSource` sont nouveaux) : chaque règle y est tenue par une mutation.

## Mutations à la main

Rejouer : `python3 -I docs/pr/lot_E/mutate_gc-19-key-file-reload.py [M1 M2 …]` depuis la racine de Thalamus (passe par `tools/handmut.py`, restauration depuis une copie). 20 mutations, 20 tuées. Fichier `KS` = `pkg/authz/key_source.go`.

| # | Mutation | Test qui tombe (extrait) |
| :- | :- | :- |
| M1 | `Current` sert l'instantané sans comparer le `stat` | `TestAKeySourceServesAKeyAddedToTheFileAtTheNextLookup`, `…RefusesAKeyRemovedFromTheFile…`, `TestABrokenKeyFileKeepsThePreviousKeysAndWarnsOncePerChange`, `TestAFileReplacedWithTheSameSizeAndTimeIsStillReread`, les tests serveur |
| M2 | nouveau jeu rendu mais pas publié (`cur.Store` retiré sur succès) | `TestAKeySourceServesAKeyAddedToTheFileAtTheNextLookup` (« each lookup parses the file again »), `TestAKeySourceRecoversWhenABrokenFileIsFixed` |
| M3a | échec : instantané publié avec un jeu nil | `TestABrokenKeyFileKeepsThePreviousKeysAndWarnsOncePerChange`, `TestTheSameTextRewrittenIsNotParsedAgain`, `TestABrokenFileIsNotReadAgainUntilItChanges` |
| M3b | échec : `reread` rend nil | `TestABrokenKeyFileKeepsThePreviousKeysAndWarnsOncePerChange`, `TestAnEmptiedKeyFileDoesNotOpenTheGRPCServer`, `TestAnEmptiedKeyFileDoesNotOpenTheRESTServer` |
| M4 | `Warn` jamais appelé | `TestABrokenKeyFileKeepsThePreviousKeysAndWarnsOncePerChange`, `TestAKeySourceWithoutWarnLogsAFailedChange` |
| M5 | avertit à chaque lecture (`if key != s.failed` → `if true`) | `TestABrokenKeyFileKeepsThePreviousKeysAndWarnsOncePerChange` (cas « file removed », « a directory in its place ») |
| M6 | clé d'échec non remise à zéro après un succès | `TestAKeySourceRecoversWhenABrokenFileIsFixed` |
| M7 | texte toujours reparsé (`bytes.Equal` contre nil) | `TestTheSameTextRewrittenIsNotParsedAgain`, `TestABrokenKeyFileKeepsThePreviousKeysAndWarnsOncePerChange` |
| M8 | `os.SameFile` retiré de la comparaison | `TestAFileReplacedWithTheSameSizeAndTimeIsStillReread` |
| M9 | fichier cassé relu à chaque requête (instantané de l'échec non publié) | `TestABrokenFileIsNotReadAgainUntilItChanges` |
| M10 | source statique qui fait un `stat` | `TestAStaticKeySourceServesItsKeysAndReadsNoFile` |
| M11 | `ResolveSource` ignore le fichier de clés | `TestAKeyRemovedFromTheKeyFileIsRefusedWithoutARestart`, `TestAKeyFileThatCannotBeLoadedProducesAServerThatRefusesEverything`, `TestAKeyRemovedFromTheKeyFileGets401WithoutARestart`, … |
| M12 | sans `Warn`, rien dans le journal | `TestAKeySourceWithoutWarnLogsAFailedChange` |
| M13 | gRPC : `NewWithPolicy` capture `src.Current()` une fois | `TestAKeyRemovedFromTheKeyFileIsRefusedWithoutARestart` |
| M14 | REST : `NewWithPolicy` capture `src.Current()` une fois | `TestAKeyRemovedFromTheKeyFileGets401WithoutARestart` |
| M15 | gRPC : `KeyFile` lu même quand `Options.Keys` est posé | `TestAPreloadedKeySetIsNeverReread` |
| M16 | REST : idem | `TestAPreloadedKeySetIsNeverRereadOverREST` |
| M17 | REST : jeu désactivé plus laissé passer (test d'`Enabled` par requête retiré) | `TestAServerWithNoTokenAsksForNone` et cinq autres tests existants |
| M18 | `AuthInterceptorFromSource` capture `src.Current()` une fois | `TestAServerOfItsOwnRereadsTheKeyFileThroughAuthInterceptorFromSource` |
| M19 | `ResolveSource` sans fichier perd le jeton historique | `TestResolveSourcePicksThePolicyAsResolveDoes` |

Mutant équivalent connu, non compté : tester `Enabled()` à la construction dans `authenticateFrom` au lieu de chaque requête. Un fichier qui se charge a toujours au moins une clé (`Parse` refuse un fichier vide), donc une source de fichier ne passe jamais de « activé » à « désactivé ». Le test par requête est gardé parce que `authenticate` (jeu statique) passe par la même fonction.

## Constat sur le rapport de vérification

Le rapport `recheck_01_04.md` disait que M2 (jeu parsé mais pas publié) ferait tomber `…AddedToTheFile…`. Faux sur le premier jet : sans publication, chaque requête relit et reparse le fichier, et sert quand même la clé nouvelle. Ajout d'une assertion (« each lookup parses the file again », même pointeur sur deux `Current`) qui fait tomber M2. Même chose pour M9, d'où `TestABrokenFileIsNotReadAgainUntilItChanges`.

## precommit

`make precommit REPO=…/wtE/gc-19-key-file-reload BASE=a591ae1` : OK (gofmt, vet, `-race` de `pkg/authz`, `pkg/httpapi`, `pkg/rpcserver`, `patchcov` 0 bloc sans test). Aucune ligne dans un `main`.
