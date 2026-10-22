# Relecture adverse de la PR à part `gc-11-rest-upsert-rows`, tour 3 (21 octobre 2026, tête `91ea926`)

Première relecture selon la méthode révisée (section 1 bis : kit `tools/advkit.py`, mutations propres à l'agent, tableau des scénarios). Agent `general-purpose` avec les consignes de `upstream-reviewer` en tête d'invite (l'agent dédié n'était pas encore chargé dans la session).

Coût (`tools/agentcost.py`) : 45 tours, contexte 46 079 -> 151 394 tokens, lecture cache 4 699 963, écriture cache 151 392, environ 659 000 tokens d'entrée en équivalent (tour 2 : 743 000). 19 scénarios sur chaque arbre, 10 mutations propres, 3 survivantes utiles.

Verdict : aucun bloquant, 3 remarques probables, 4 nits.

- **L1.** Les lignes sans propriétaire ne sont pas sous test : mémoire du seau global (`user_id` vide), document de la collection par défaut (`graphrag_chunks`). Mutations P (`upsert.go:129`) et R (`:120`) survivent ; le code les refuse aujourd'hui.
- **L2.** La sentinelle n'est pas distinguée d'une lecture du texte : mutation H (`errors.Is` remplacé par un test sur « not found ») survit.
- **L3.** Quatrième copie de la règle ; avec #74, le port REST donnera trois libellés pour le même refus (`not found` sur la route des outils, `knowledge "<id>" not found` ici, `supersede "x": embedding not found` de la façade après #73) ; même bloc de `authorizeWith` modifié par les deux PR.
- **N1.** Ordre des refus changé, non nommé : un identifiant étranger répond 404 avant le 400 du gestionnaire ; un corps sans `memory_id` avec une cible étrangère ou absente répond 404 au lieu de `memory_id is required`.
- **N2.** Le test sans base tombe sur `main` seulement par la panique du gestionnaire ; `NewWithPolicy(nil, …)` n'est pas une configuration réelle, la branche `db == nil` garde un état injouable.
- **N3.** Un test épingle le libellé de la façade, qui n'est pas celui de la PR.
- **N4.** La garde dépend des lectures par identifiant de la façade non confinées ; un commentaire le dirait.

Tableau des scénarios et mutations : rapport de l'agent, conservé dans la conversation du 21 octobre 2026 ; résumé ci-dessus.
