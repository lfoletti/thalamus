# 25. Deux tests du hook de rappel lisent l'environnement de la machine

**Statut** : **envoyée en relecture le 15 novembre 2026 (liliang-cn/cortexdb#195, tête 0ae20ed, branche `co-33-recall-tests-own-environment` du fork lfoletti/cortexdb), base `main` 7e4b22b.** **Taille** : une ligne dans `cmd/cortexdb-mcp-stdio/recall_situation_test.go`. **Dépend de** : rien.

## Le défaut

`TestAPromptWithAnImageRecallsNothing` et `TestAPromptThatIsAllMachineryRecallsNothing` lancent le hook de rappel sur un cerveau à eux, mais avec l'encodeur, le transformateur de requête et le plancher de rappel que la machine a réglés.

## Qui l'atteint

Toute machine qui pose `CORTEXDB_MEMORY_SEMANTIC_FLOOR` ou les variables d'un encodeur. Le canari de `make precommit` pose ce plancher à une valeur refusée : toute branche qui touche `cmd/cortexdb-mcp-stdio` était refusée par le hook `pre-push`.

## Preuves

`CORTEXDB_MEMORY_SEMANTIC_FLOOR=x go test -count=1 ./cmd/cortexdb-mcp-stdio/` : deux tests tombent sur `main` (`without an image the memory is recalled, got ""`), aucun sur la branche.

## Le correctif

`recallFor` appelle `withoutTheMachinesModels`, l'aide que les autres tests du paquet utilisent déjà.

## Ce qui change

Rien hors des tests.

## Ce que Thalamus y gagne

Le canari passe sur ce paquet.

Pour européens : sans effet sur les langues.
