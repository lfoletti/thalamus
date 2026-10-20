# ADR 0003 : statut de relecture opt-in, décision jamais effacée

- Statut : acceptée le 20 octobre 2026, en relecture amont (PR #77).
- Date : 20 octobre 2026 (première version le 19 octobre 2026).

## Contexte

Sur un cerveau partagé, un fait extrait par un modèle ne doit pas être servi comme vrai avant qu'une personne l'ait lu. CortexDB n'avait pas de notion de statut. Le propriétaire a demandé que le masquage des lignes non relues soit opt-in, pour qu'un cerveau qui ne pose jamais le champ cherche exactement comme aujourd'hui. Sa revue du 20 octobre 2026 a ajouté : tout écrivain pouvait effacer une relecture, `reviewed_by` était du texte libre, l'acceptation effaçait les autres métadonnées, `reviewed_at` était redaté à chaque synchronisation, la porte s'ouvrait sur une panne.

## Décision

- Trois clés de métadonnées sur documents et mémoires : `review_status` (`proposed`, `accepted`, `rejected`), `reviewed_by`, `reviewed_at`.
- Le masquage est opt-in : `cortexdb.WithReviewGate()`, `CORTEXDB_REVIEW_GATE=1`. Une ligne masquée reste lisible par id, par `memory_list_all`, par son nœud de graphe et par le filtre `review_status`.
- Le relecteur d'une décision (`accepted`, `rejected`) est la clé appelante quand il y en a une, quoi que le client ait écrit. Sans clé (bibliothèque embarquée), le texte libre reste.
- Une écriture qui ne dit rien de la relecture reporte la décision stockée. Une décision ne s'efface pas : elle se remplace. `reviewed_at` est gardé sur une re-sauvegarde de la même décision par le même relecteur.
- Une `knowledge_update` qui décide d'une relecture fusionne ses métadonnées dans les métadonnées stockées au lieu de les remplacer. Une mise à jour qui ne parle pas de relecture remplace la carte, comme en amont, et reporte `expires_at` et les marques d'ingestion.
- Avec la porte active, une panne de la requête de porte ne sert rien. Sans la porte, le filtre d'expiration reste consultatif.
- Index d'expression sur `review_status`, même motif que l'expiration.

## Options écartées

- Masquage par défaut : change le comportement de tout cerveau existant. Refusé par le propriétaire.
- Appel d'acceptation dédié : un outil de plus dans la liste blanche. La fusion sur décision couvre le besoin sans nouvel outil.
- Report des seules clés système à l'acceptation (version du 20 octobre au matin) : perd les clés du client. Remplacé par la fusion le soir même, sur demande de Lauro.

## Conséquences

- Une proposition (`proposed`) garde le `reviewed_by` donné par le client : une proposition n'a pas de relecteur.
- Le relecteur amont peut encore contester la fusion partielle (remplacement hors relecture). Le corps de #77 l'explique.

## Références

- PR amont #77 (`gc-06-review-status`), discussion #66 (réponse du 18 octobre 2026 sur l'opt-in).
- `docs/pr/lot_C/GC-06.body.md`, `docs/pr/lot_C/failing_first.md`.
