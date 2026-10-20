# Lot D : la voie graphe de `cortex_query` reste dans la collection (GC-10)

## État

Préparée le 20 octobre 2026 au soir, sur commande de Lauro : branche `gc-10-cortex-query-prefetch` sur `origin/main` à v2.134.1 (`60a8831`), un commit (`49c5366`). **Poussée et ouverte en brouillon le 20 octobre 2026 à 19 h 50 : PR #78** (https://github.com/liliang-cn/cortexdb/pull/78), avec le lot B. CI verte, Codecov « 100.00% of diff hit », passée « ready for review » le 20 octobre 2026 à 19 h 55. Elle répond au point 3 de la revue de #72 (20 octobre 2026) : le propriétaire demandait de passer la collection au prefetch graphe de `cortex_query`. #72 refuse l'outil aux clés confinées ; cette PR corrige l'outil lui-même. Après sa fusion, une PR à part ajoute `cortex_query` à la liste blanche, argument par argument : analyse faite dans `cortex_query_allowlist.md` (ligne proposée, `prefetch[].source` à refuser, `filter` sans oracle). Cette analyse a mis au jour un trou corrigé dans #76 et #77 : les voies de `cortex_query` ne passaient par aucune porte.

## Contenu du dossier

- `GC-10.body.md` : corps de la PR (sans pied de page).
- `failing_first.md` : preuve « failing first » (test qui tombe sur `main`, quatre mutations).
- `mutate_10.py` : script des mutations à la main.
- `cortex_query_allowlist.md` : analyse argument par argument pour la PR qui suivra (liste blanche).

## Commandes (depuis la racine `granular_cortex`, préfixe `!`, une ligne chacune), sur commande de Lauro

1. `cd cortexdb && git push fork gc-10-cortex-query-prefetch`
2. `cd cortexdb && gh pr create --draft --repo liliang-cn/cortexdb --head lfoletti:gc-10-cortex-query-prefetch --title "fix(cortexdb): the graph lane of cortex_query stays in the collection" --body-file ../docs/pr/lot_D/GC-10.body.md`
3. `gh pr ready <n>` quand Codecov montre zéro ligne non couverte.
