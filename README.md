<img src="extension/icons/icon128.png" alt="Thalamus logo: two pale lobes on an indigo square" width="96" align="right">

# Thalamus

A small governance layer over [CortexDB](https://github.com/liliang-cn/cortexdb) for the shared knowledge of an organisation: which employee may read or write which part, with their own key, and a trace of every call they make. CortexDB is the engine: documents, chunks, entities, a property graph, RDF and SPARQL, scoped API keys. Thalamus adds what an organisation needs around it and nothing more: who may read or write which brain, keys issued from those rights, one endpoint that asks every brain a person may read, and an audit line for every call.

*Version française plus bas.*

## Why

The base engine has no way to define or manage a clearance policy on a shared brain, for a company or across projects. Its key file is a list of secrets written by hand. Thalamus is a governance layer over CortexDB: it decides which employee gets which key on which brain. It brings a rights file (brains, groups, readers, writers), keys issued and revoked from it, a gateway that applies it and logs every call under a person's name, and one recall across the brains a person may read. It serves the team that runs the brains, the curators and the employees.

1. **No fork of the engine.** The fixes this layer needed went upstream as small pull requests. The layer imports CortexDB at a pinned version. No maintenance debt.
2. **The engine enforces, not a proxy.** Every employee reaches a brain with their own confined key. The gateway cannot err in the caller's favour: it decides who gets a key, the engine decides what the key may do.
3. **One key per person and per brain.** Revocation is individual. Every audit line names a person, not a shared token.
4. **Provenance on every extract.** The agent and the reader know which brain, which document and which author a sentence comes from.
5. **One recall for every readable brain.** The employee does not choose where to look, and sees only what their rights allow.
6. **Thin and proven.** A few hundred lines. Every test was written before its code and fails without it. Every rule has a recorded hand mutation that its test catches. An end-to-end bench runs a real server.

## What it does

- **Rights.** `rights.yaml` declares the brains (one CortexDB server each, or a collection of a shared server), the groups, the readers, the writers and the curators. A writer reads too. The gateway reads the file again when it changes: a withdrawn right is refused at the next call, without a restart.
- **Keys.** `th-gateway key issue` writes a key into the server's own key file, confined to the brain's collection, read-only or read-write from the right. The engine enforces the key; the gateway only decides who gets one. A brain that is its whole server gets an unconfined key, which opens the graph tools.
- **Relay.** One MCP endpoint per brain, `/mcp/<brain>`, over streamable HTTP. The caller presents the key the gateway issued. The gateway refuses before it reaches the brain (unknown key, key of another brain, right withdrawn, unknown brain), forwards each tool call over gRPC with the caller's key, and appends one line to `audit.jsonl`: who, which key, which brain, which tool, which ids, success or error.
- **Federated recall.** `/mcp` admits any key the gateway issued and names its holder. `th_brains` lists the caller's brains. `th_recall` asks every readable brain in parallel with the caller's own key for it, fuses the lists by weighted reciprocal rank, merges the same text across brains, and gives each result a fixed provenance line: `[brain] title, id, source, by author`.
- **REST.** `GET /v1/brains` and `POST /v1/add` for clients without MCP, a browser extension or a script: same key, same circuit, same audit.
- **Login.** `POST /v1/login` with the id_token of the organisation's identity provider (OpenID Connect): the gateway derives the login, issues the keys the rights allow, and returns them. `th-gateway login` and the extension's "Sign in" button do this for the employee. No key is handed out by hand.
- **Browser.** A Chrome extension (`extension/`): propose the page you read, or your selection, into a brain you may write to, for the curators. Same key, same circuit.
- **Client.** A Claude Code plugin (`plugin/`): the gateway's tools over HTTP under the employee's key, and a `UserPromptSubmit` hook that injects the best extracts with their provenance into each prompt. The hook never blocks a prompt.

## Layout

| Path | Content |
| :- | :- |
| `gateway/` | The Go module (`rights`, `keys`, `relay`, `cmd/th-gateway`, `cmd/th-hook`), its tests and `mutate.py`, the hand-mutation proof. |
| `plugin/` | The Claude Code plugin. |
| `extension/` | The Chrome extension. |
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

# Thalamus (français)

Une petite couche de gouvernance au-dessus de [CortexDB](https://github.com/liliang-cn/cortexdb) pour les connaissances partagées d'une organisation : quel employé peut lire ou écrire quelle partie, avec sa propre clé, et une trace de chaque appel qu'il fait. CortexDB est le moteur : documents, extraits, entités, graphe de propriétés, RDF et SPARQL, clés d'API à portée. Thalamus ajoute ce dont une organisation a besoin autour, et rien de plus : qui peut lire ou écrire quel cerveau, des clés émises à partir de ces droits, un point unique qui interroge tous les cerveaux qu'une personne peut lire, et une ligne d'audit par appel.

## Pourquoi

Le modèle de base ne permet pas de définir ni de gérer une politique de clearance sur un cerveau partagé, d'entreprise ou multi-projets. Son fichier de clés est une liste de secrets écrite à la main. Thalamus est une couche de gouvernance au-dessus de CortexDB : elle décide quel employé reçoit quelle clé sur quel cerveau. Elle apporte un fichier de droits (cerveaux, groupes, lecteurs, rédacteurs), des clés émises et révoquées à partir de ces droits, un gateway qui les applique et journalise chaque appel au nom d'une personne, et un rappel sur tous les cerveaux qu'une personne peut lire. Il sert à l'équipe qui exploite les cerveaux, aux curateurs et aux employés.

1. **Pas de fork du moteur.** Les correctifs dont cette couche a besoin sont partis en amont en petites PR. La couche importe CortexDB à une version épinglée. Pas de dette de maintenance.
2. **C'est le moteur qui applique, pas un proxy.** Chaque employé joint un cerveau avec sa propre clé confinée. Le gateway ne peut pas se tromper en faveur de l'appelant : il décide qui reçoit une clé, le moteur décide ce que la clé peut faire.
3. **Une clé par personne et par cerveau.** La révocation est individuelle. Chaque ligne d'audit nomme une personne, pas un jeton partagé.
4. **Provenance sur chaque extrait.** L'agent et le lecteur savent de quel cerveau, de quel document et de quel auteur vient une phrase.
5. **Un rappel pour tous les cerveaux lisibles.** L'employé ne choisit pas où chercher, et ne voit que ce que ses droits permettent.
6. **Mince et prouvé.** Quelques centaines de lignes. Chaque test a été écrit avant son code et tombe sans lui. Chaque règle a une mutation à la main consignée, que son test attrape. Un banc de bout en bout tourne sur un vrai serveur.

## Ce que ça fait

- **Droits.** `rights.yaml` déclare les cerveaux (un serveur CortexDB chacun, ou une collection d'un serveur partagé), les groupes, les lecteurs, les rédacteurs et les curateurs. Un rédacteur lit aussi. Le gateway relit le fichier quand il change : un droit retiré est refusé à l'appel suivant, sans redémarrage.
- **Clés.** `th-gateway key issue` écrit une clé dans le fichier de clés du serveur, confinée à la collection du cerveau, en lecture seule ou en lecture-écriture selon le droit. Le moteur applique la clé ; le gateway décide seulement qui en reçoit une. Un cerveau qui est tout son serveur reçoit une clé non confinée, ce qui ouvre les outils de graphe.
- **Relais.** Un point MCP par cerveau, `/mcp/<cerveau>`, en HTTP. L'appelant présente la clé émise par le gateway. Le gateway refuse avant de joindre le cerveau (clé inconnue, clé d'un autre cerveau, droit retiré, cerveau inconnu), relaie chaque appel d'outil en gRPC avec la clé de l'appelant, et ajoute une ligne à `audit.jsonl` : qui, quelle clé, quel cerveau, quel outil, quels identifiants, succès ou erreur.
- **Rappel fédéré.** `/mcp` admet toute clé émise par le gateway et nomme son porteur. `th_brains` liste les cerveaux de l'appelant. `th_recall` interroge en parallèle chaque cerveau lisible avec la clé de l'appelant pour ce cerveau, fusionne les listes par rang réciproque pondéré, réunit un même texte présent dans plusieurs cerveaux, et donne à chaque résultat une ligne de provenance fixe : `[cerveau] titre, id, source, by auteur`.
- **REST.** `GET /v1/brains` et `POST /v1/add` pour les clients sans MCP, une extension de navigateur ou un script : même clé, même circuit, même audit.
- **Connexion.** `POST /v1/login` avec l'`id_token` du fournisseur d'identité de l'organisation (OpenID Connect) : le gateway dérive le login, émet les clés que les droits permettent et les rend. `th-gateway login` et le bouton « Sign in » de l'extension le font pour l'employé. Aucune clé distribuée à la main.
- **Navigateur.** Une extension Chrome (`extension/`) : proposer la page lue, ou la sélection, dans un cerveau où l'on écrit, pour les curateurs. Même clé, même circuit.
- **Client.** Un plugin Claude Code (`plugin/`) : les outils du gateway en HTTP sous la clé de l'employé, et un hook `UserPromptSubmit` qui injecte les meilleurs extraits avec leur provenance dans chaque prompt. Le hook ne bloque jamais un prompt.

## Arborescence

| Chemin | Contenu |
| :- | :- |
| `gateway/` | Le module Go (`rights`, `keys`, `relay`, `cmd/th-gateway`, `cmd/th-hook`), ses tests et `mutate.py`, la preuve par mutations à la main. |
| `plugin/` | Le plugin Claude Code. |
| `extension/` | L'extension Chrome. |
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
