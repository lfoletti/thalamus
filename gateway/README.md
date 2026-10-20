# gc-gateway

La couche mince de `granular_cortex` au-dessus d'un ou plusieurs serveurs CortexDB. Première tranche (20 octobre 2026) : droits, clés, relais MCP, audit. Le moteur reste seul à appliquer la portée d'une clé ; le gateway décide qui reçoit quelle clé et note qui a fait quoi.

## Paquets

- `rights` : lit `rights.yaml` (cerveaux, groupes, lecteurs, rédacteurs). Un rédacteur lit aussi. Champ mal orthographié, groupe non déclaré, membre sans préfixe `user:` ou `group:` : refusés à la lecture.
- `keys` : émet et révoque les clés confinées dans le fichier de clés du serveur (`pkg/authz` de CortexDB). Identifiant `gc/<user>/<brain>`, secret aléatoire, portée = la collection du cerveau, clearance = `read-only` ou `read-write` selon le droit. Écriture en un pas, jamais d'un fichier que le serveur refuserait de charger. `Holder` retrouve l'utilisateur derrière un secret.
- `relay` : un serveur MCP (HTTP streamable, réponses JSON) par cerveau sur `/mcp/<brain>`. Le porteur présente sa clé (`Authorization: Bearer`). Le gateway refuse avant de joindre le cerveau : clé inconnue (401), clé d'un autre cerveau ou droit retiré de `rights.yaml` (403), cerveau non déclaré (404), cerveau injoignable (502). Chaque appel d'outil est relayé en gRPC avec la clé de l'appelant et écrit dans `audit.jsonl` : instant, utilisateur, clé, cerveau, outil, identifiants (`*_id`, `*_ids`, `supersedes`), succès ou erreur.
- `cmd/gc-gateway` : `rights check`, `rights of`, `key issue`, `key revoke`, `serve`.

## Essai

```sh
cd gateway && go test -race ./...          # quatre paquets, environ 10 s
python3 -I mutate.py                        # 29 mutations à la main, chacune doit tomber
sh ../tests/e2e_gateway.sh                  # serveur gRPC depuis cortexdb/, gateway devant, 19 contrôles
```

Couverture : 100 % des instructions hors `main` (quatre lignes).

## Limites de la tranche

- `cortexdb-grpc` lit son fichier de clés au démarrage seulement : après `key issue` ou `key revoke`, le serveur redémarre. Le gateway, lui, refuse une clé révoquée aussitôt (il relit le fichier à chaque session). Prochaine étape côté amont : rechargement du fichier de clés.
- Une session MCP par clé, gardée en mémoire tant que le gateway tourne.
- Pas d'OIDC, pas de fédération entre cerveaux, pas d'extraction : prochaines tranches.
- Le gateway importe CortexDB v2.134.1 en bibliothèque ; les sept PR ouvertes ne sont pas dedans. Le banc d'essai lance le serveur construit depuis `cortexdb/` (branche `proto/granular-cortex`), qui les contient.
