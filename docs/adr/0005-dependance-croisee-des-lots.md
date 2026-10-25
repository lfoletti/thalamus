# ADR 0005 : liste blanche tenue au schéma, le lot fusionné en second complète

- Statut : acceptée le 20 octobre 2026.
- Date : 20 octobre 2026.

## Contexte

La liste blanche de #72 (ADR 0001) est tenue au schéma des outils par un test de complétude. Le lot C (#75 à #77) ajoute des arguments (`dedupe`, `expires_at` sur `ingest_document`, filtre `plan.filters.review_status`) et un outil (`sweep_expired`). Chaque lot part de `main` et passe seul. La réunion des deux lots sur `proto/granular-cortex` a fait tomber le test de complétude et 5 contrôles sur 30 de l'essai de bout en bout.

## Décision

- Chaque PR est cohérente seule, sur sa base.
- Le lot fusionné en second ajoute à la liste blanche ce que l'autre a introduit, au rebasage : si le lot B est fusionné d'abord, chaque PR du lot C classe ses propres arguments et outils ; si le lot C est fusionné d'abord, #72 les classe.
- Sur `proto/granular-cortex`, un commit propre à la branche classe les éléments du lot C, entre le dernier commit du lot C et les correctifs suivants. Il n'est pas poussé.
- Les corps de PR et les réponses au relecteur annoncent cette règle.

## Options écartées

- Faire dépendre le lot C du lot B (empilement) : trois sujets à relire en même temps, et un lot bloqué par l'autre.
- Retirer le test de complétude : c'est lui qui protège la liste (ADR 0001).

## Conséquences

- Tout ajout d'argument ou d'outil dans une PR future doit s'accompagner de sa ligne dans la liste blanche ou dans `refusedTools`, avec sa raison.
- L'essai de bout en bout se rejoue sur `proto` reconstruite après chaque modification d'un lot.

## Références

- `../thalamus_local/docs/pr/lot_C/README.md`, section « Dépendance croisée avec le lot B ».
- Mémoire CortexDB `knowledge:granular-cortex:allowlist-cross-lot-dependency`.
