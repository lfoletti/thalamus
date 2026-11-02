# 13. Fondre et séparer des cerveaux

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : quatre PR, 150 + 250 + 250 + 120 lignes hors tests. **Dépend de** : la fiche 12, PR A (écriture conditionnelle), pour l'import ; #75 (empreinte du contenu). **Motivée par** : une fusion réelle de trois cerveaux de sources en un seul, avec une archive pour ce qui n'y entre pas, faite côté client parce que ni le moteur ni le gateway n'avaient les outils.

## Ce que la fusion a demandé

Trois cerveaux d'environ 13 000 documents en tout, chacun avec ses sources, ses extractions, ses entités et sa table d'alias, devaient devenir un cerveau fusionné et une archive. Tout s'est fait dans un script du client, sans appel au modèle, en six opérations :

1. **Ranger chaque document vers une destination.** La règle naturelle va par source : une source entière va au cerveau de la majorité de ses documents. Mais une source sur trois était mixte (un blog qui traite tantôt du droit, tantôt de la vie privée) : 2 094 documents devaient aller ailleurs que leur source. Le rangement se fait donc par source, avec des exceptions par document. Un document que le rangement n'a pas su placer reste « à revoir », jamais placé par devinette ; un document arrivé après le rangement suit la majorité de sa source.
2. **Un cerveau qui en reprend d'autres, avec un filtre.** Avant de déménager quoi que ce soit, le cerveau fusionné a été servi en lisant les documents des trois cerveaux et en ne gardant que ceux rangés chez lui. Cela a permis de mesurer le résultat sans rien déplacer.
3. **Dédoublonner.** 177 documents étaient en double entre les cerveaux. Un même corps est un doublon. Une même adresse ne suffit pas : un texte de référence publie tous ses articles sous une adresse et un titre, et la première règle, sur l'adresse seule, fondait les articles distincts d'une même convention. La règle retenue : une adresse normalisée qui garde le fragment (`#art-5` désigne un article), et des corps presque identiques (indice de Jaccard des mots ≥ 0,9 ; les articles distincts mesurés montaient à 0,69). Le document gardé est le plus ancien ; les extractions de l'autre sont écartées.
4. **Fondre les tables d'alias.** Chaque cerveau avait sa table (une variante ou une traduction → le nom canonique). La fusion résout les chaînes (a → b dans un cerveau, b → c dans un autre donne a → c) et liste les désaccords (a → b ici, a → c là) au lieu d'en choisir un.
5. **Retirer ce qu'attestent les documents exclus.** Une entité citée seulement par des documents partis à l'archive ne doit pas rester dans le cerveau fusionné ; une entité aussi citée par un document gardé reste, détachée des autres.
6. **Séparer par thème.** Avec la convention des cerveaux (fiche 12, `gateway/convention/`), chaque document porte un thème et une nature. Séparer un cerveau devient un rangement dont la règle est le thème, avec les mêmes garanties.

Les garanties que les tests ont vérifiées, et que tout outil de fusion ou de séparation devrait vérifier aussi :

- **aucune perte** : chaque document d'origine a exactement une destination (un cerveau, l'archive, ou « doublon de » un document gardé) ;
- **ses extractions suivent le document** ; une extraction dont le document n'existe plus est comptée comme orpheline, pas perdue en silence ;
- **une source dans un seul cerveau** après la fusion ;
- **idempotence** : appliquer deux fois revient à appliquer une fois, et le plan recalculé après application est vide ;
- **un plan d'abord** : le même calcul, sans rien écrire, donne les chiffres par destination avant toute écriture.

## Le partage

Selon la ligne du 20 octobre 2026 (README de ce dossier) :

- **Thalamus garde la politique** : qui range, selon quelle règle (source, exception, thème), le plan et sa relecture par un curateur, les droits du cerveau fusionné, et l'ancien nom servi comme alias du nouveau pendant la transition (`rights.yaml`). Le dédoublonnage par adresse reste aussi chez le client : la normalisation d'une adresse est une règle du domaine.
- **CortexDB reçoit la mécanique** : copier des documents d'un cerveau à l'autre sans les recalculer, fondre des entités selon une table donnée, retirer le graphe d'un ensemble de documents.

## Ce qui existe déjà dans CortexDB (v2.142.0)

- `delete_document_graph` retire tout ce qu'un document a mis dans le graphe ; les entités qu'il est seul à citer sont supprimées, les autres détachées, avec `dry_run`. C'est exactement l'opération 5, mais document par document.
- `GraphStore.MergeEntities(canonical, aliases)` fond des entités en gardant l'historique (raison `merged`, `superseded_by`). Il n'est atteint que par `cortexdb-mcp-stdio --resolve-entities`, qui choisit lui-même les paires (avec un modèle) : aucun outil n'applique une table d'alias donnée par le client.
- `ingest_document` avec `dedupe` refuse un contenu identique déjà présent et nomme le détenteur (`duplicate_of`) : le doublon exact est couvert, pas le presque identique (fiche 12, PR D).
- `knowledge_update` peut changer la collection d'un document, mais redécoupe et réencode tout (fiche 12, « À vérifier » : une cinquième PR).
- `knowledge_graph_import` et `knowledge_graph_export` échangent des triplets RDF, pas des documents avec leurs morceaux, vecteurs et provenance.
- `graph_snapshot` et `graph_diff` permettent de vérifier après coup ce qu'une fusion a changé.

## A. Un outil qui applique une table d'alias

**Le besoin.** La table fondue de l'opération 4 est connue du client ; il ne lui manque que de l'appliquer au graphe.

**La proposition.** `merge_entities {canonical, aliases[], dry_run}`, ou une liste de paires, au-dessus de `MergeEntities` : mêmes garanties d'historique, bilan (arêtes repointées, boucles retirées). Pour une clé confinée, seules les entités de sa collection.

**Tests, écrits avant le code.** Une fusion repointe les arêtes et garde la version d'avant en historique ; `dry_run` n'écrit rien ; un alias d'une autre collection est refusé à une clé confinée. Mutations : historique non écrit, garde de collection retirée.

## B. Exporter des documents avec ce qu'ils portent

**Le besoin.** Copier un document d'un cerveau à l'autre passe aujourd'hui par une nouvelle ingestion : morceaux redécoupés, vecteurs recalculés (des heures pour 6 000 documents), entités et relations à reverser.

**La proposition.** `export_documents {ids | filter, with_vectors}` rend un flux (JSONL) : le document, ses morceaux, leurs vecteurs si demandé, ses mentions d'entités et ses relations avec leur provenance. Lecture seule, filtrée comme `knowledge_search` pour une clé confinée.

## C. Importer ce flux

**La proposition.** `import_documents {stream, dry_run}` : chaque document est écrit avec ses identifiants, sous la condition `if_absent` de la fiche 12 (un document déjà présent est rapporté, pas écrasé), dans une transaction par document. Les vecteurs sont repris s'ils ont la dimension du cerveau, recalculés sinon. Bilan : écrits, déjà présents, refusés.

**Tests, écrits avant le code.** Export puis import dans un cerveau vide : même contenu, mêmes morceaux, mêmes vecteurs, même provenance (`graph_diff` vide entre les deux). Un second import n'écrit rien. Mutations : condition ignorée, vecteurs recalculés alors qu'ils étaient fournis.

## D. Retirer le graphe d'un ensemble de documents

**Le besoin.** L'opération 5 portait sur 4 000 documents : 4 000 appels de `delete_document_graph`, et entre deux appels un état où une entité est détachée d'un document mais pas encore de l'autre.

**La proposition.** `delete_document_graph` accepte `ids[]` ou un filtre de métadonnées, avec `dry_run` qui rend les comptes, et calcule « seul à citer » sur l'ensemble, pas document par document.

## Après la fusion de ces PR

Le client ne garde que la politique : le plan de rangement, ses tests et sa relecture. Le déménagement devient export, import avec condition, table d'alias appliquée, retrait du graphe des exclus ; une séparation par thème, un export filtré par `theme`. Côté Thalamus, un outil de curateur pourrait calculer et montrer le plan (`th_split` ou `th_merge` en lecture seule), puis l'appliquer par ces quatre outils ; il n'est pas proposé avant que le moteur les ait.

## À vérifier avant d'écrire

- Si `MergeEntities` est atteignable par un autre chemin public que l'option de `cortexdb-mcp-stdio`.
- La taille d'un flux pour 6 000 documents avec vecteurs, et s'il faut le paginer.
- L'ordre proposé : A (le plus petit, utile tout de suite), puis B et C ensemble, puis D.
