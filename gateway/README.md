# gc-gateway

La couche mince de `granular_cortex` au-dessus d'un ou plusieurs serveurs CortexDB. Tranche 1 (20 octobre 2026) : droits, clés, relais MCP, audit. Tranche 2 (même jour) : point fédéré, cerveau dédié à clé non confinée. Le moteur reste seul à appliquer la portée d'une clé ; le gateway décide qui reçoit quelle clé et note qui a fait quoi.

## Paquets

- `rights` : lit `rights.yaml` (cerveaux, groupes, lecteurs, rédacteurs, `weight` optionnel pour le rappel fédéré). Un rédacteur lit aussi. Un cerveau sans `collection` est **tout son serveur** : ses clés ne sont pas confinées, et le moteur lui ouvre les outils de graphe (Cypher, SPARQL, ontologies), fermés à une clé confinée. Champ mal orthographié, groupe non déclaré, membre sans préfixe `user:` ou `group:` : refusés à la lecture.
- `keys` : émet et révoque les clés confinées dans le fichier de clés du serveur (`pkg/authz` de CortexDB). Identifiant `gc/<user>/<brain>`, secret aléatoire, portée = la collection du cerveau, clearance = `read-only` ou `read-write` selon le droit. Écriture en un pas, jamais d'un fichier que le serveur refuserait de charger. `Holder` retrouve l'utilisateur derrière un secret.
- `relay` : un serveur MCP (HTTP streamable, réponses JSON) par cerveau sur `/mcp/<brain>`. Le porteur présente sa clé (`Authorization: Bearer`). Le gateway refuse avant de joindre le cerveau : clé inconnue (401), clé d'un autre cerveau ou droit retiré de `rights.yaml` (403), cerveau non déclaré (404), cerveau injoignable (502). Chaque appel d'outil est relayé en gRPC avec la clé de l'appelant et écrit dans `audit.jsonl` : instant, utilisateur, clé, cerveau, outil, identifiants (`*_id`, `*_ids`, `supersedes`), succès ou erreur.
- `relay`, point fédéré `/mcp` : l'employé présente n'importe laquelle de ses clés, qui le nomme. Deux outils : `gc_brains` (ses cerveaux, son droit, clé émise ou non) et `gc_recall` (`query`, `top_k`, `brains` optionnel) qui interroge en parallèle chaque cerveau lisible par `knowledge_search` avec la clé de l'employé pour ce cerveau, puis fusionne par rang réciproque pondéré (`weight / (60 + rang)`). Un même extrait dans deux cerveaux est un résultat, qui nomme les autres dans `also_in`. Chaque résultat porte une ligne de provenance fixe : `[cerveau] titre, id, source, by auteur`. Un cerveau sans clé, injoignable ou en erreur est dans `skipped` avec sa raison. Audit : une ligne par `gc_recall` avec les cerveaux interrogés et les identifiants lus.
- `cmd/gc-gateway` : `rights check`, `rights of`, `key issue`, `key revoke`, `serve` (sert `/mcp` et `/mcp/<cerveau>`).

## Essai

```sh
cd gateway && go test -race ./...          # quatre paquets, environ 10 s
python3 -I mutate.py                        # 48 mutations à la main, chacune doit tomber
sh ../tests/e2e_gateway.sh                  # serveur gRPC depuis cortexdb/, gateway devant, 24 contrôles
```

Couverture : 100 % des instructions hors `main` (quatre lignes).

## Limites de la tranche

- `cortexdb-grpc` lit son fichier de clés au démarrage seulement : après `key issue` ou `key revoke`, le serveur redémarre. Le gateway, lui, refuse une clé révoquée aussitôt (il relit le fichier à chaque session). Prochaine étape côté amont : rechargement du fichier de clés.
- Une session MCP par clé, gardée en mémoire tant que le gateway tourne.
- Le rappel fédéré fusionne des extraits de texte, pas des nœuds : une question par entité à travers les cerveaux est une tranche ultérieure. Dédoublonnage sur le texte de l'extrait (les réponses de `knowledge_search` ne portent pas le hash).
- Pas d'OIDC, pas d'extraction, pas de contradictions : prochaines tranches.
- Le gateway importe CortexDB v2.134.1 en bibliothèque ; les sept PR ouvertes ne sont pas dedans. Le banc d'essai lance le serveur construit depuis `cortexdb/` (branche `proto/granular-cortex`), qui les contient.
