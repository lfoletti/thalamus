# granular_cortex

A thin layer over [CortexDB](https://github.com/liliang-cn/cortexdb) for the shared knowledge brains of an organisation, used by code agents such as Claude Code. CortexDB is the engine: documents, chunks, entities, a property graph, RDF and SPARQL, scoped API keys. granular_cortex adds what an organisation needs around it and nothing more: who may read or write which brain, keys issued from those rights, one endpoint that asks every brain a person may read, and an audit line for every call.

*Version française plus bas.*

## What it does

- **Rights.** `rights.yaml` declares the brains (one CortexDB server each, or a collection of a shared server), the groups, the readers and the writers. A writer reads too.
- **Keys.** `gc-gateway key issue` writes a key into the server's own key file, confined to the brain's collection, read-only or read-write from the right. The engine enforces the key; the gateway only decides who gets one. A brain that is its whole server gets an unconfined key, which opens the graph tools.
- **Relay.** One MCP endpoint per brain, `/mcp/<brain>`, over streamable HTTP. The caller presents the key the gateway issued. The gateway refuses before it reaches the brain (unknown key, key of another brain, right withdrawn, unknown brain), forwards each tool call over gRPC with the caller's key, and appends one line to `audit.jsonl`: who, which key, which brain, which tool, which ids, success or error.
- **Federated recall.** `/mcp` admits any key the gateway issued and names its holder. `gc_brains` lists the caller's brains. `gc_recall` asks every readable brain in parallel with the caller's own key for it, fuses the lists by weighted reciprocal rank, merges the same text across brains, and gives each result a fixed provenance line: `[brain] title, id, source, by author`.
- **Client.** A Claude Code plugin (`plugin/`): the gateway's tools over HTTP under the employee's key, and a `UserPromptSubmit` hook that injects the best extracts with their provenance into each prompt. The hook never blocks a prompt.

## Layout

| Path | Content |
| :- | :- |
| `gateway/` | The Go module (`rights`, `keys`, `relay`, `cmd/gc-gateway`, `cmd/gc-hook`), its tests and `mutate.py`, the hand-mutation proof. |
| `plugin/` | The Claude Code plugin. |
| `tests/e2e_gateway.sh` | End-to-end bench: a CortexDB server, the gateway, keys, calls over HTTP, the hook, the audit file. |
| `docs/` | Architecture revision, ADRs, the PR pipeline, the upstream PR folders (in French). |
| `cortexdb/` | Not in this repository: a clone of CortexDB, expected next to `gateway/` for the bench. |

## Upstream

The engine fixes this layer needs went upstream as small pull requests (CortexDB discussion #66): confined keys enforced on the tool box, a content hash, `expires_at`, an opt-in review status, and `cortex_query` kept in its collection. Six are merged; seven are in review as of 20 October 2026. The gateway imports CortexDB as a library at a pinned version; the bench runs the server built from the branch that carries the open fixes.

## Build and test

```sh
cd gateway
go build ./...
go test -race ./...          # about 10 s, coverage 100 % outside main
python3 -I mutate.py         # one hand mutation per rule, each must fail its test
cd .. && sh tests/e2e_gateway.sh   # needs cortexdb/ next to gateway/
```

Rules of the house: the test is written before the code and fails without it, every reachable line of a change is covered before it is pushed, and a pull request opens as a draft. The whole procedure is in `docs/pipeline_pr.md`.

## Status

Day one (20 October 2026). Three slices: rights, keys and relay; the federated endpoint; the client. Known limits: the CortexDB server reads its key file at start only, so it restarts after a key is issued or revoked; the federated recall fuses extracts, not graph nodes; no OIDC yet.

## License

MIT. CortexDB is MIT too.

---

# granular_cortex (français)

Une couche mince au-dessus de [CortexDB](https://github.com/liliang-cn/cortexdb) pour les cerveaux de connaissances partagés d'une organisation, utilisés par des agents de code comme Claude Code. CortexDB est le moteur : documents, extraits, entités, graphe de propriétés, RDF et SPARQL, clés d'API à portée. granular_cortex ajoute ce dont une organisation a besoin autour, et rien de plus : qui peut lire ou écrire quel cerveau, des clés émises à partir de ces droits, un point unique qui interroge tous les cerveaux qu'une personne peut lire, et une ligne d'audit par appel.

## Ce que ça fait

- **Droits.** `rights.yaml` déclare les cerveaux (un serveur CortexDB chacun, ou une collection d'un serveur partagé), les groupes, les lecteurs et les rédacteurs. Un rédacteur lit aussi.
- **Clés.** `gc-gateway key issue` écrit une clé dans le fichier de clés du serveur, confinée à la collection du cerveau, en lecture seule ou en lecture-écriture selon le droit. Le moteur applique la clé ; le gateway décide seulement qui en reçoit une. Un cerveau qui est tout son serveur reçoit une clé non confinée, ce qui ouvre les outils de graphe.
- **Relais.** Un point MCP par cerveau, `/mcp/<cerveau>`, en HTTP. L'appelant présente la clé émise par le gateway. Le gateway refuse avant de joindre le cerveau (clé inconnue, clé d'un autre cerveau, droit retiré, cerveau inconnu), relaie chaque appel d'outil en gRPC avec la clé de l'appelant, et ajoute une ligne à `audit.jsonl` : qui, quelle clé, quel cerveau, quel outil, quels identifiants, succès ou erreur.
- **Rappel fédéré.** `/mcp` admet toute clé émise par le gateway et nomme son porteur. `gc_brains` liste les cerveaux de l'appelant. `gc_recall` interroge en parallèle chaque cerveau lisible avec la clé de l'appelant pour ce cerveau, fusionne les listes par rang réciproque pondéré, réunit un même texte présent dans plusieurs cerveaux, et donne à chaque résultat une ligne de provenance fixe : `[cerveau] titre, id, source, by auteur`.
- **Client.** Un plugin Claude Code (`plugin/`) : les outils du gateway en HTTP sous la clé de l'employé, et un hook `UserPromptSubmit` qui injecte les meilleurs extraits avec leur provenance dans chaque prompt. Le hook ne bloque jamais un prompt.

## Arborescence

| Chemin | Contenu |
| :- | :- |
| `gateway/` | Le module Go (`rights`, `keys`, `relay`, `cmd/gc-gateway`, `cmd/gc-hook`), ses tests et `mutate.py`, la preuve par mutations à la main. |
| `plugin/` | Le plugin Claude Code. |
| `tests/e2e_gateway.sh` | Banc de bout en bout : un serveur CortexDB, le gateway, les clés, des appels en HTTP, le hook, le fichier d'audit. |
| `docs/` | Révision d'architecture, ADR, pipeline de PR, dossiers des PR amont. |
| `cortexdb/` | Hors de ce dépôt : un clone de CortexDB, attendu à côté de `gateway/` pour le banc. |

## Amont

Les correctifs du moteur dont cette couche a besoin sont partis en amont en petites PR (discussion #66 de CortexDB) : clés confinées appliquées à la boîte à outils, hash de contenu, `expires_at`, statut de relecture opt-in, `cortex_query` tenu dans sa collection. Six sont fusionnés, sept en relecture au 20 octobre 2026. Le gateway importe CortexDB en bibliothèque à une version épinglée ; le banc lance le serveur construit depuis la branche qui porte les correctifs ouverts.

## Construire et tester

```sh
cd gateway
go build ./...
go test -race ./...          # environ 10 s, couverture 100 % hors main
python3 -I mutate.py         # une mutation à la main par règle, chacune doit faire tomber son test
cd .. && sh tests/e2e_gateway.sh   # demande cortexdb/ à côté de gateway/
```

Règles de la maison : le test s'écrit avant le code et tombe sans lui, chaque ligne atteignable d'un changement est couverte avant la poussée, et une PR s'ouvre en brouillon. La procédure complète est dans `docs/pipeline_pr.md`.

## État

Jour un (20 octobre 2026). Trois tranches : droits, clés et relais ; le point fédéré ; le client. Limites connues : le serveur CortexDB lit son fichier de clés au démarrage seulement, donc il redémarre après l'émission ou la révocation d'une clé ; le rappel fédéré fusionne des extraits, pas des nœuds du graphe ; pas d'OIDC.

## Licence

MIT. CortexDB est sous MIT aussi.
