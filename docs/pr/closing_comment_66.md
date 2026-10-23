# Commentaire de clôture de la discussion #66

Posté le 23 octobre 2026 sur commande de Lauro, après la fusion de la série : https://github.com/liliang-cn/cortexdb/discussions/66#discussioncomment-19138549

Texte envoyé :

The series is complete. #61, #67 to #78, #82 and #83 are merged, and with #75 to #77 in v2.137.0 nothing of mine is open. Thank you for the reviews, and for fixing #84 in the same release.

One item stays open on my side: `cortex_query` is still refused to a confined key. I can send a small PR that allows it argument by argument, with `prefetch[].source` kept refused, if you want it.

The layer above CortexDB is Thalamus: several brains per organisation, rights declared once and turned into keys, recall across brains, and the audit. It is now public at https://github.com/lfoletti/thalamus. It uses CortexDB v2.138.0 as a library and asks nothing more of the engine.
