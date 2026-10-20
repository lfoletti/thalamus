# ADR 0006 : pipeline de PR, tests « failing first », couverture du patch, brouillon, une poussée par lot

- Statut : acceptée le 19 octobre 2026, en vigueur pour tous les projets.
- Date : 19 octobre 2026.

## Contexte

Le propriétaire de CortexDB demande que chaque PR porte des tests qui échouent sans son code. Les premières poussées du 19 octobre 2026 ont produit des rapports Codecov partiels (#76 : 39 lignes, #77 : 45 lignes) et plusieurs amendements successifs. Lauro a tranché le soir même : « je ne veux plus voir de nouveau commit dans le vide ».

## Décision

1. Le test s'écrit avant le code et tombe sans lui. La preuve est une mutation à la main par règle, qui compile, et un test nommé qui tombe. Elle est consignée dans `docs/pr/<lot>/failing_first.md` et dans le corps de la PR.
2. Le patch est couvert à 100 % de ses lignes atteignables avant le premier push, mesuré en local (`make precommit`). Les lignes du corps de `main()` et les blocs qu'aucun test ne peut lancer (liste DDL PostgreSQL) sont déclarés dans une section `## Codecov` du corps.
3. `make mutate` en plus, chaque survivant lu et classé ; ceux des paquets `main` rejoués à la main.
4. Une PR s'ouvre en brouillon et passe en relecture quand le rapport de couverture montre zéro ligne non couverte hors celles déclarées. Une PR à amender repasse en brouillon.
5. Une poussée par lot, après la vérification complète (`make prepr` : precommit, mutate, suite `-race`, éval lexicale). Le hook `pre-push` rejoue `precommit`.
6. Commits sans ligne `Co-Authored-By`, corps de PR sans pied de page d'outil. Dans les commentaires au propriétaire, on n'explique pas ses propres outils.

## Options écartées

- Pousser au fil des rapports Codecov : c'est ce qui a produit les amendements en cascade.
- Couverture mesurée par Codecov seulement : un rapport par poussée, donc une poussée par essai.

## Conséquences

- Le code s'écrit pour être couvrable : `main` réduit à la répartition, environnement lu dans des fonctions qui rendent une erreur, pannes SQLite déterministes pour les chemins d'erreur.
- Environ 25 minutes de vérification par PR avant la poussée, et le disque à surveiller (cache Go vidé entre les suites).

## Références

- `docs/pipeline_pr.md` (procédure complète et pièges), `Makefile` à la racine.
- Mémoires CortexDB `feedback:dev-process:patch-coverage-before-push`, `feedback:dev-process:draft-pr-until-zero-uncovered`, `feedback:dev-process:no-push-before-full-local-verification`.
