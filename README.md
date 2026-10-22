<p align="center"><img src="extension/icons/icon128.png" alt="Thalamus logo: two pale lobes on an indigo square" width="96"></p>

# Thalamus

Une mémoire partagée pour une organisation et ses agents de code, construite sur [CortexDB](https://github.com/liliang-cn/cortexdb). CortexDB est le moteur : documents, extraits, entités, graphe de propriétés, RDF et SPARQL, clés d'API à portée. Thalamus est une couche mince autour de lui.

*English version below.*

## Pourquoi

Le moteur lit ses clés dans un fichier de secrets écrit à la main. Il ne sait pas quelle personne tient quelle clé, et il ne peut donc pas appliquer une politique d'accès à un cerveau partagé. Thalamus garde les droits de chaque personne dans un seul fichier. Il donne à chaque personne une clé pour chaque cerveau qu'elle peut utiliser, et il journalise chaque appel au nom de cette personne. Il sert à l'équipe qui exploite les cerveaux et aux personnes qui s'en servent.

Principes :

1. **Les correctifs du moteur partent en amont.** Quand Thalamus a eu besoin d'un changement dans CortexDB, ce changement est parti vers le moteur en petite PR. Thalamus importe CortexDB à une version épinglée et ne garde aucun fork.
2. **Le moteur applique les droits.** Chaque personne appelle un cerveau avec sa propre clé, confinée à ce qu'elle peut lire ou écrire. Le gateway décide qui reçoit une clé. Le moteur contrôle chaque appel avec cette clé.
3. **Une clé par personne et par cerveau.** Pour retirer l'accès d'une personne, le gateway révoque ses clés. Chaque ligne d'audit nomme la personne qui a fait l'appel.
4. **Chaque extrait montre sa source.** Un résultat nomme son cerveau, son document et son auteur.
5. **Un rappel sur les cerveaux lisibles.** L'employé pose sa question une fois. Le gateway interroge chaque cerveau que l'employé peut lire et fusionne les réponses.
6. **Petit et testé.** Environ 3 000 lignes hors tests. Chaque test a été écrit avant son code et tombe sans lui. Chaque règle a une mutation à la main consignée, que son test attrape. Un banc fait tourner le gateway devant un vrai serveur.

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
| `gateway/` | Le module Go (`rights`, `identity`, `keys`, `relay`, `login`, `cmd/th-gateway`, `cmd/th-hook`), ses tests et `mutate.py`, la preuve par mutations à la main. |
| `plugin/` | Le plugin Claude Code. |
| `extension/` | L'extension Chrome. |
| `tests/` | Les bancs de bout en bout. `e2e_gateway.sh` lance un serveur CortexDB et le gateway, puis contrôle les clés, les appels, le hook, la connexion et l'audit. `e2e_extension.sh` pilote l'extension dans Chromium. `e2e_plugin.sh` ouvre une vraie session Claude Code avec le plugin. `fake_idp.mjs` est un faux fournisseur OpenID Connect pour les bancs. |
| `docs/` | Révision d'architecture, ADR, pipeline de PR, dossiers des PR amont. |
| `cortexdb/` | Hors de ce dépôt : un clone de CortexDB, attendu à côté de `gateway/` pour le banc. |

## Amont

Les correctifs du moteur dont cette couche a besoin sont partis en amont en petites PR (discussion #66 de CortexDB) : clés confinées appliquées à la boîte à outils, hash de contenu, `expires_at`, statut de relecture opt-in, `cortex_query` tenu dans sa collection. Au 22 octobre 2026, onze sont fusionnées et trois sont en relecture (#75, #76, #77). Le gateway importe CortexDB v2.136.0 en bibliothèque. Le banc lance le serveur construit depuis la branche qui porte les correctifs ouverts.

## Construire et tester

```sh
cd gateway
go build ./...
go test -race ./...          # environ 10 s, couverture 100 % hors main
python3 -I mutate.py         # une mutation à la main par règle, chacune doit faire tomber son test
cd ../extension && npm install && node --test   # règles et icônes de l'extension
cd .. && sh tests/e2e_gateway.sh   # demande cortexdb/ à côté de gateway/
sh tests/e2e_extension.sh          # demande Playwright, installé par npm install
```

Règles de la maison : le test s'écrit avant le code et tombe sans lui, chaque ligne atteignable d'un changement est couverte avant la poussée, et une PR s'ouvre en brouillon. La procédure complète est dans `docs/pipeline_pr.md`.

## État

Au 22 octobre 2026, le gateway a douze tranches. Elles couvrent les droits, les clés, le relais, le rappel fédéré, le circuit d'écriture avec relecture, la face REST, la connexion OpenID Connect, le plugin et l'extension.

Limites connues. Les deux viennent de CortexDB, et leur correction est en cours. Chaque correctif est prêt sous forme de PR amont, dans `docs/upstream/`. Ces PR partiront quand les PR en relecture (#75 à #77) seront fusionnées et que ce dépôt sera public.

- **Redémarrage du serveur.** Le serveur CortexDB lit son fichier de clés au démarrage seulement. Il doit redémarrer après l'émission ou la révocation d'une clé. Le correctif fait relire le fichier par le serveur quand le fichier change (`docs/upstream/01-key-file-reload.md`).
- **Outils de graphe fermés aux clés confinées.** Le graphe est partagé entre les collections. Le moteur refuse donc les outils de graphe à une clé confinée à une collection. Pour un tel cerveau, le rappel fédéré fusionne seulement des extraits. Le correctif cache dans les parcours du graphe ce qui est hors du périmètre de la clé, puis ouvre ces outils (`docs/upstream/06-graph-tools-confined.md`, deux PR). Un cerveau qui est tout son serveur n'a pas cette limite.

## Licence

MIT. CortexDB est sous MIT aussi.

---

# Thalamus (English)

A shared memory for an organisation and its coding agents, built on [CortexDB](https://github.com/liliang-cn/cortexdb). 
CortexDB is the engine: documents, chunks, entities, a property graph, RDF and SPARQL, scoped API keys. 
Thalamus is a thin layer around it.

## Why

The engine reads its keys from a file of secrets written by hand. It does not know which person holds which key, so it cannot apply a clearance policy to a shared brain. Thalamus keeps the rights of each person in one file. It gives each person a key for each brain they may use, and it records every call under that person's name. It is for the team that runs the brains and for the people who use them.

Principles:

1. **Fixes to the engine go upstream.** When Thalamus needed a change in CortexDB, the change went to the engine as a small pull request. Thalamus imports CortexDB at a pinned version and keeps no fork.
2. **The engine enforces the rights.** Each person calls a brain with their own key, confined to what they may read or write. The gateway decides who gets a key. The engine checks every call against that key.
3. **One key per person and per brain.** To remove the access of one person, the gateway revokes their keys. Each audit line names the person who made the call.
4. **Each extract shows its source.** A result names its brain, its document and its author.
5. **One recall over the readable brains.** The employee asks once. The gateway asks each brain the employee may read and merges the answers.
6. **Small and tested.** About 3,000 lines outside tests. Each test was written before its code and fails without it. Each rule has a recorded hand mutation that its test catches. A bench runs the gateway against a real server.

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
| `gateway/` | The Go module (`rights`, `identity`, `keys`, `relay`, `login`, `cmd/th-gateway`, `cmd/th-hook`), its tests and `mutate.py`, the hand-mutation proof. |
| `plugin/` | The Claude Code plugin. |
| `extension/` | The Chrome extension. |
| `tests/` | The end-to-end benches. `e2e_gateway.sh` starts a CortexDB server and the gateway, then checks the keys, the calls, the hook, the login and the audit. `e2e_extension.sh` drives the extension in Chromium. `e2e_plugin.sh` opens a real Claude Code session with the plugin. `fake_idp.mjs` is a fake OpenID Connect provider for the benches. |
| `docs/` | Architecture revision, ADRs, the PR pipeline, the upstream PR folders (in French). |
| `cortexdb/` | Not in this repository: a clone of CortexDB, expected next to `gateway/` for the bench. |

## Upstream

The engine fixes this layer needs went upstream as small pull requests (CortexDB discussion #66): confined keys enforced on the tool box, a content hash, `expires_at`, an opt-in review status, and `cortex_query` kept in its collection. As of 22 October 2026, eleven are merged and three are in review (#75, #76, #77). The gateway imports CortexDB v2.136.0 as a library. The bench runs the server built from the branch that carries the open fixes.

## Build and test

```sh
cd gateway
go build ./...
go test -race ./...          # about 10 s, coverage 100 % outside main
python3 -I mutate.py         # one hand mutation per rule, each must fail its test
cd ../extension && npm install && node --test   # rules and icons of the extension
cd .. && sh tests/e2e_gateway.sh   # needs cortexdb/ next to gateway/
sh tests/e2e_extension.sh          # needs Playwright, installed by npm install
```

Rules of the house: the test is written before the code and fails without it, every reachable line of a change is covered before it is pushed, and a pull request opens as a draft. The whole procedure is in `docs/pipeline_pr.md`.

## Status

As of 22 October 2026, the gateway has twelve slices. They cover the rights, the keys, the relay, the federated recall, the write circuit with review, the REST face, the OpenID Connect login, the plugin and the extension.

Known limits. Both come from CortexDB, and their fix is in progress. Each fix is ready as an upstream PR, in `docs/upstream/`. These PRs will go when the PRs in review (#75 to #77) are merged and this repository is public.

- **Server restart.** The CortexDB server reads its key file at start only. It must restart after a key is issued or revoked. The fix makes the server read the file again when the file changes (`docs/upstream/01-key-file-reload.md`).
- **Graph tools closed to confined keys.** The graph is shared between collections. So the engine refuses the graph tools to a key confined to one collection. For such a brain, the federated recall fuses extracts only. The fix hides from the graph walks what is outside the key's scope, then opens these tools (`docs/upstream/06-graph-tools-confined.md`, two PRs). A brain that is its whole server does not have this limit.

## License

MIT. CortexDB is MIT too.
