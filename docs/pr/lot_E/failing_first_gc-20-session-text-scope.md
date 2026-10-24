# Failing first : gc-20-session-text-scope

Tête : `gc-20-session-text-scope` (copie `scratchpad/wtE/gc-20-session-text-scope`), base `origin/main` `a591ae1`. Un commit. Le 23 octobre 2026.

## Tests posés sur la base, sans le code

`pkg/cortexdb/extract_conversation_scope_test.go` compile sur `a591ae1`. Commande, sur une copie de la base avec ce seul fichier :

```
go test ./pkg/cortexdb/ -run 'TestExtractConversation' -count=1
```

Sortie observée :

```
--- FAIL: TestExtractConversationAnswersASessionOutsideTheCallerScopeAsMissing (0.03s)
    extract_conversation_scope_test.go:46: zeus's session was summarized under hermes's scope: {"summary":"memory: Zeus hid the Thunderbolt under Olympus. Hera must not know.","themes":[…],"entities":["Zeus","Thunderbolt","Olympus","Hera"],…}
--- FAIL: TestExtractConversationWithPersistWritesNothingForASessionOutsideTheScope (0.02s)
    extract_conversation_scope_test.go:63: zeus's session was extracted under hermes's scope
--- FAIL: TestExtractConversationFailsWhenTheSessionBucketCannotBeRead (0.01s)
    extract_conversation_scope_test.go:108: an unreadable bucket passed as readable: &{Summary:memory: Hermes tuned the Lyre for Apollo. …}
```

`TestExtractConversationStillReadsTheCallersOwnSession` passe sur la base : c'est la garde contre « tout refuser », tenue par M4.

## Mutations à la main

Rejouer : `python3 -I docs/pr/lot_E/mutate_gc-20-session-text-scope.py` depuis la racine de Thalamus. 4 mutations, 4 tuées.

| # | Mutation | Test qui tombe |
| :- | :- | :- |
| M1 | garde retirée | `TestExtractConversationAnswersASessionOutsideTheCallerScopeAsMissing`, `TestExtractConversationWithPersistWritesNothingForASessionOutsideTheScope` |
| M2 | refus rendu par une erreur propre (« session withheld ») | `TestExtractConversationAnswersASessionOutsideTheCallerScopeAsMissing` |
| M3 | erreur de lecture du seau avalée | `TestExtractConversationFailsWhenTheSessionBucketCannotBeRead` |
| M4 | condition inversée | `…AsMissing`, `…PersistWritesNothing…`, `TestExtractConversationStillReadsTheCallersOwnSession` |

## precommit

`make precommit REPO=…/wtE/gc-20-session-text-scope BASE=a591ae1` : OK (`-race` de `pkg/cortexdb` en 425 s, `patchcov` 0 bloc sans test).
