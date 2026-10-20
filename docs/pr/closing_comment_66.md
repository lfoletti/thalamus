# Commentaire de clôture de la discussion #66 (en réserve)

À poster **sur commande de Lauro, après la fusion des sept PR** (#72 à #78). Décision du 20 octobre 2026.

With these seven PRs the series announced above is complete. What follows on my side, once they land: one small PR that adds `cortex_query` to the allowlist, argument by argument (`prefetch[].source` stays refused: an external source loads rows by id without a collection check), and the rebases that classify each batch's arguments for the other. Nothing else is planned for CortexDB itself. The layer above it (several brains per organisation, rights declared once and turned into keys, audit gathered across servers) lives in its own project and asks nothing more of the engine. Thank you for the reviews.
