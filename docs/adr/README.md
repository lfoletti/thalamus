# Décisions d'architecture (ADR)

Un fichier par décision, numéroté, jamais réécrit : une décision qui change reçoit un nouvel ADR qui remplace l'ancien. Champs : Statut, Date, Contexte, Décision, Options écartées, Conséquences, Références. Les ADR décrivent les décisions de Thalamus et des correctifs que le projet pousse vers CortexDB. Ils ne sont pas repris dans les PR amont : le propriétaire lit les corps de PR et tient seul son changelog.

| N° | Titre | Statut |
| :- | :- | :- |
| [0001](0001-liste-blanche-par-outil.md) | Liste blanche explicite par outil pour les clés confinées | Acceptée, en relecture amont (#72) |
| [0002](0002-outils-a-deux-magasins.md) | Outils à deux magasins : porte forcée ou refus | Acceptée, en relecture amont (#72 à #74) |
| [0003](0003-relecture-opt-in.md) | Statut de relecture opt-in, décision jamais effacée | Acceptée, en relecture amont (#77) |
| [0004](0004-expiration-filtre-et-balayage.md) | Expiration : filtre consultatif, balayage comme garantie, balayage exposé comme outil | Acceptée, en relecture amont (#76) |
| [0005](0005-dependance-croisee-des-lots.md) | Liste blanche tenue au schéma : le lot fusionné en second complète | Acceptée |
| [0006](0006-pipeline-de-pr.md) | Pipeline de PR : tests « failing first », couverture du patch, brouillon, une poussée par lot | Acceptée, en vigueur |
