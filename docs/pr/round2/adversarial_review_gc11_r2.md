# Relecture adverse de la PR à part `gc-11-rest-upsert-rows`, tour 2 (21 octobre 2026, tête `afca137`)

Rapport de l'agent neuf (invite fixe). Verdict : aucun bloquant, 2 remarques probables, 5 nits.

Vérifié par l'agent : six tests sur huit tombent sur `main` (deux ne compilent pas) ; 26 mutations sur 26 tuées, rejouées sur sa copie ; `-race` vert sous le canari ; 100 % de `upsert.go` ; 25 scénarios sur chaque arbre, 9 diffèrent, tous nommés sauf `UpdateKnowledge` ; repli de casse de `encoding/json` et `strings.EqualFold` identique (K de Kelvin, ſ, ı, İ).

- **L1.** « Les clés non confinées ne sont pas contrôlées » sans test : la garde appelée pour toute clé survit ; le test ne tient que le statut 404, pas le libellé de la façade ni la mémoire `m-1` écrite.
- **L2.** Le correctif de l'identifiant blanc en gRPC ferme aussi `UpdateKnowledge` (`knowledge.go:66`), ni nommé ni testé (sur `main`, le document blanc passe de `secrets` à `notes`).
- **N1.** Le saut de la cible égale à la nouvelle mémoire compare toutes les orthographes de `memory_id`, la façade la seule que garde `encoding/json` (la dernière) : `{"Memory_Id":"new-b","memory_id":"new-a","supersedes":["new-b"]}` laisse `new-a` écrite avec un 404. Pas une évasion de portée.
- **N2.** Cible propre entourée d'espaces non tenue par un test (`TrimSpace` retiré survit).
- **N3.** Libellé `supersede "<id>": not found` non épinglé.
- **N4.** `errors.Is(err, sql.ErrNoRows)` mort : la façade traduit déjà en `core.ErrNotFound`.
- **N5.** Les gardes gRPC laissent passer une seconde lecture en panne (`ownership.go:243`, `:263`), la garde REST refuse ; la PR touche ces fonctions.
- Hors compte : la branche `db == nil` n'est atteinte que par l'aide de test `authorize` ; « une lecture par identifiant » sous-estime le coût (`GetKnowledge` charge les chunks).
