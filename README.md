<p align="center"><img src="extension/icons/icon128.png" alt="Thalamus logo: two pale lobes on an indigo square" width="96"></p>

# Thalamus

Thalamus brings the knowledge of an organisation to each user and to their coding agent within the rights of each person. It is built on [CortexDB](https://github.com/liliang-cn/cortexdb).

> **Work in progress.** The interfaces change between versions. A short French summary is at the end.

## What it does

A team keeps its knowledge in several CortexDB brains: one for the team, one for a client, one for the whole organisation. Each user may read some of them and write to fewer. Thalamus is the gateway between the users and these brains.

The user's assistant reaches the brains through the gateway, which is an MCP server. When the assistant asks a question with `th_recall`, the gateway searches each brain that this user may read, with the user's own key, and merges the results. Each extract comes with its source: the brain, the document and the author. With `th_get`, the assistant reads the full document behind an extract.

A user can also propose a document, from Claude Code or from a page in Chrome. The document waits for a curator of the brain, who accepts or rejects it.

## Why use it

### 1. Governance

Thalamus is a thin governance layer. It lets an organisation share its knowledge with calm and with efficiency, and it keeps each access under control. The management stays easy: the rights are in one file, the keys follow the rights, and a change applies without a restart.

- **Access without requests.** A user logs in with the account of the organisation (OpenID Connect). The gateway issues the keys that the rights allow. To give a person access to a team's brains, add the login to the team's group in `rights.yaml`. The change applies at the next call, without a restart.
- **One key per person and per brain.** CortexDB checks the key on every call, so the scope of a key does not depend on the gateway. To revoke one person, the gateway removes that person's keys only.
- **Knowledge added in one step.** A user proposes a document from Claude Code or from a page in Chrome. A curator of the brain accepts or rejects it, from the same tools.
- **Simple administration.** `th-gateway key list` compares the keys with the rights, and `key sync` corrects the difference. `th-gateway audit` gives a summary of the reads, the writes, the refusals and the logins.

### 2. Compliance and sovereignty

Thalamus is built to support compliance with the GDPR, the Swiss Federal Act on Data Protection and, for a high-risk use, the EU AI Act. A tool does not make an organisation compliant. But Thalamus gives the controls that these frameworks ask for:

| Control | In Thalamus | Articles |
| :- | :- | :- |
| Data minimisation | The reads are counted per brain and per day, without a name. Only a brain with `audit_reads: true` has named reads. | GDPR 5(1)(c), 25(2). FADP 6(2), 7(3). |
| Storage limitation | `audit: retention_days` limits how long the audit lines stay. | GDPR 5(1)(e). FADP 6(4). |
| Security and traceability | Personal keys, checked by the engine. The writes, the reviews, the logins and the refused calls have a line with the name of the person. | GDPR 5(1)(f), 5(2), 32. FADP 8. |
| Data location | The organisation hosts the gateway and the CortexDB servers. The local brains stay on the machine of the user and do not go through the gateway. | GDPR 44. FADP 16. |
| Transparency | Each extract gives its brain, its document and its author. | GDPR 5(1)(a). |
| **Not covered:** the transfer to the model provider | The hook puts passages into the prompt, and the prompt goes to the model provider of the assistant. Thalamus does not control this transfer. The organisation covers it itself: a processing agreement with the provider, transfer safeguards when the provider is outside Switzerland or the EU, and the retention setting at the provider. | GDPR 28, 44 to 46. FADP 9, 16. |

A brain with `audit_reads: true` keeps a named read log, and that log is personal data: the organisation informs the users, answers their access requests and sets its retention (GDPR 13, 15. FADP 19, 25).

**If the use is high-risk.** The AI Act articles below apply to high-risk AI systems only. A recall for coding agents is not a high-risk use in general. For a deployer of a high-risk system, Thalamus gives the matching control:

| Control | In Thalamus | Articles |
| :- | :- | :- |
| Record-keeping | The writes, the reviews, the logins and the refused calls have a line with the name of the person. | AI Act 12. |
| Human oversight | A proposed document waits for a curator before the recall serves it. | AI Act 14, 26(2). |

GDPR: Regulation (EU) 2016/679. FADP: the revised Swiss Federal Act on Data Protection (SR 235.1). AI Act: Regulation (EU) 2024/1689.

### 3. Local and shared brains in one recall

- **Brains of the correct size.** A brain is a whole CortexDB server or one collection of a shared server. Small teams can share one server. A sensitive brain can have its own server.
- **Federation.** One question reaches every brain that the user may read. The gateway merges the results, with a weight per brain. The user does not choose where to look.
- **Local and shared together.** In Claude Code, a hook asks the shared brains and the local brains of the machine at the same time, at each prompt. Each person chooses the brains of the automatic recall with `/th-recall` (see [Combining local and global brains](#combining-local-and-global-brains)).

### 4. Accuracy that you can check

- **Passages with their source.** The recall serves the passages of the documents, not short extracts. With `th_get`, the agent reads the full document.
- **A pyramid of syntheses.** A brain can have a collection of syntheses above its documents. `th_recall` serves the syntheses first, and `th_descend` opens the documents that a synthesis cites.
- **Measured.** On a research pilot, the facts found by the automatic recall went from 51.7 % to 81.3 % with the passages and the syntheses, and to 85.1 % with a variant of the question in English.
- **Next.** A check of each sentence of a synthesis against its sources, proposed to CortexDB. Then the federation of the entities across brains, in addition to the passages.

### 5. Multilingual

A user asks in their language, and the document that answers is often in another one. `th_recall` takes the question in other languages (`alternate_queries`) and searches each one apart.

- The Swiss bench (`tests/eval_ch/`) has 40 questions on the Federal Constitution, in German, French, Italian and Romansh. Each question asks for an article in another language.
- With `multilingual-e5-large` and the variants, 97 % of the expected articles are in the first 10 results, and 91 % for Romansh. Without an encoder, only 7 % are.

For coding agents, Thalamus brings the knowledge into the session: through MCP for each assistant, and through an automatic recall at each prompt in Claude Code.

## Concepts

| Term | Meaning |
| :- | :- |
| Brain | A CortexDB server, or one collection of a shared server. |
| Rights | The file `rights.yaml`. It declares the brains, the groups, and who reads, writes or curates each brain. |
| Key | A CortexDB API key for one person on one brain. The gateway writes it into the server's key file, confined to the brain's collection, read-only or read-write. |
| Curator | A person who accepts or rejects the documents proposed to a brain. |

## How it works

<p align="center"><img src="docs/img/architecture.svg" alt="The Claude Code plugin, the Chrome extension and th-gateway login call the Thalamus gateway. The gateway reads rights.yaml, writes audit.jsonl, and calls the CortexDB servers with the caller's key: one key per person and per brain." width="760"></p>

- The gateway reads `rights.yaml` and issues the keys. It reads the file again when it changes, so a removed right applies at the next call.
- Each call goes to CortexDB with the caller's key. CortexDB checks the key on every call, so the scope of a key does not depend on the gateway.
- The gateway keeps an audit in `audit.jsonl`, for security and traceability. The writes, the reviews, the logins and the refused calls have a line with the name of the person. The reads are counted per brain and per day, without a name. Only a brain with `audit_reads: true` in `rights.yaml` has named reads. `audit: retention_days` limits how long the lines stay.
- A user logs in with the organisation's identity provider (OpenID Connect). The gateway then issues the keys that the rights allow.

## Combining local and global brains

The Thalamus plugin for Claude Code adds a hook to the tools of the gateway. At each prompt, the hook asks the shared brains and the local brains of the machine at the same time. It gives the best extracts to the agent before the agent answers. The user does not ask for them.

<p align="center"><img src="docs/img/recall-hook.svg" alt="The user types a prompt. The Thalamus hook sends it to th_recall on the Thalamus gateway, with the user's key, and the gateway searches the server brains. At the same time, the hook searches the local brains on the machine. The hook adds the best extracts, with their source, to the prompt. Then the agent answers." width="760"></p>

- The hook (`th-hook`, on the `UserPromptSubmit` event) sends the prompt to `th_recall` with the user's key. It adds the best passages of the documents to the prompt, with their source, in 9,000 characters or fewer: 10 passages, after 4 passages of syntheses when a brain has a pyramid of syntheses.
- It does not ask for a prompt of fewer than 3 words, or for a `/` command.
- It never stops a prompt. If the gateway does not answer, or if the key is refused, the prompt continues without extracts.
- The agent can open the full document of an extract with `th_get`.

- For the local brains, the hook runs the native recall of CortexDB, once per brain. By default, the only local brain is the brain of the project. A setting adds other local brains, for example the brain in the home directory or the brain of another project.
- The local brains stay on the machine. They do not go through the gateway, and they are not in the audit.
- The block has the shared extracts first, then one section per local brain. The local sections use 3,000 characters or fewer.
- Each person chooses the brains of the automatic recall, local and shared, with `/th-recall` (`on`, `off`, `only`). The choice does not change the rights.

Other assistants use the same tools through MCP. The automatic recall needs a hook in the assistant. Thalamus has one for Claude Code.

## Try it

The steps start one CortexDB server with two brains, then the gateway, then the plugin. Do all the steps in the root of this repository.

1. Clone CortexDB into `cortexdb/`, next to `gateway/`, then build the binaries.

   ```sh
   git clone https://github.com/liliang-cn/cortexdb.git && (cd cortexdb && git checkout v2.138.0)
   mkdir -p bin
   (cd cortexdb && go build -o ../bin/cortexdb-grpc ./cmd/cortexdb-grpc)
   (cd gateway && go build -o ../bin/th-gateway ./cmd/th-gateway && go build -o ../plugin/bin/th-hook ./cmd/th-hook)
   ```

2. Write the rights and an empty key file.

   ```yaml
   # rights.yaml
   brains:
     team/acme:
       addr: 127.0.0.1:47998
       keys_file: keys.json
       collection: team-acme
       readers: [group:acme]
       writers: [user:alice]
       curators: [user:carol]
     org/handbook:
       addr: 127.0.0.1:47998
       keys_file: keys.json
       collection: handbook
       readers: [group:acme]
       writers: [user:carol]
   groups:
     acme: [alice, bob, carol]
   ```

   ```sh
   echo '{"keys":[]}' > keys.json
   ```

   The key file is at the path that `keys_file` gives. A relative path starts from the directory where `th-gateway` runs. The gateway writes the keys into this file, and the server reads it with `-keys`.

3. Issue the keys, then start the server. The server reads its key file when it starts.

   ```sh
   bin/th-gateway key sync rights.yaml
   bin/cortexdb-grpc -db brain.db -addr 127.0.0.1:47998 -keys keys.json &
   ```

4. Start the gateway.

   ```sh
   bin/th-gateway serve rights.yaml --listen 127.0.0.1:8787 &
   ```

5. Start Claude Code with the plugin. Use one of the secrets that `key sync` printed.

   ```sh
   export TH_GATEWAY=http://127.0.0.1:8787 TH_KEY=<secret>
   claude --plugin-dir ./plugin
   ```

With an `identity` block in `rights.yaml`, `th-gateway login http://127.0.0.1:8787` replaces the export of step 5. The details are in `gateway/README.md`, `plugin/README.md` and `extension/README.md`.

## Roadmap

**Done** (23 October 2026)

- Rights, keys and one MCP endpoint per brain.
- Federated recall over the brains that a person may read, with the source of each extract.
- Proposed documents and their review by a curator.
- REST endpoints for clients without MCP.
- Login with OpenID Connect, from the command line and from the extension.
- The Claude Code plugin: recall hook, tools and the `/th-add` command.
- The Chrome extension: propose a page or a selection. It keeps the main content of the page.
- A proportionate audit: named writes, reviews, logins and refusals, reads counted without a name, named reads only on the brains that ask for it, and a retention period.
- Local and shared brains in one recall: the hook adds the local brains of the machine, and each person chooses the brains of the automatic recall.
- Recall on the passages of the documents, not on short extracts.
- A pyramid of syntheses per brain: `th_recall` serves the syntheses first, and `th_descend` opens the documents that a synthesis cites.
- Questions in several languages: `th_recall` takes the question in other languages (`alternate_queries`) and searches each one apart. A Swiss bench measures the recall from one language to another (French, German, Italian, Romansh) in `tests/eval_ch/`.

**With CortexDB**

Thalamus relies on CortexDB for each mechanism that serves every CortexDB user. These parts go to CortexDB as small pull requests.

- Merged: the allowlist of tools for a confined key (#72 to #74, #78), content hash and duplicate check (#75), expiry dates (#76), review status (#77), and other fixes (#61, #67 to #71, #82, #83). Thalamus uses CortexDB v2.138.0.
- Proposed, in discussion #86:
  - Keys and signatures: the server reads its key file again when it changes, so a new or revoked key applies without a restart (#87). Each write is signed by the caller's key (#88). A memory is read only when the caller's scope permits it (#89).
  - Graph tools for a key confined to one collection (#94 to #98). Then Thalamus can merge entities across brains, in addition to passages.
  - Three tools and an index: the touched ids in the request log (#90), `knowledge_get` on a list of ids (#91), a check of a sentence against its cited sources (#92), and two redundant edge indexes removed (#93).
  - Search in European languages: a Swiss evaluation set (#99), accents in the coverage floor (#100), and the words and runs of alternate queries (#101, #102).
- Prepared: a tool that lists the documents waiting for review.

## Principles

1. **The engine checks every key.** The gateway decides who gets a key. CortexDB decides what each key may do.
2. **One key per person and per brain.** The gateway can revoke the access of one person on one brain.
3. **Each extract shows its source.** The reader can open the document behind an extract with `th_get`.
4. **General mechanisms go upstream.** Thalamus keeps only the organisation's policy: the rights, the groups, the curators and the federation between servers.

## Development

Read [CONTRIBUTING.md](CONTRIBUTING.md) before a change, and [SECURITY.md](SECURITY.md) to report a defect.

```sh
cd gateway && go test -race ./...           # coverage 100 % outside the two main functions
python3 -I mutate.py                         # one hand mutation per rule, each must fail a test
cd ../extension && npm install && node --test
cd .. && sh tests/e2e_gateway.sh             # a real CortexDB server and the gateway
sh tests/e2e_extension.sh                    # the extension in Chromium
```

Each test is written before its code and fails without it. The full procedure is in `docs/pipeline_pr.md`. Most project documents are in French.

## License

MIT. CortexDB is also under the MIT license.

---

## Résumé en français

**Thalamus** apporte les bases de connaissance d'une organisation à chaque utilisateur et à son agent de code, au moment où il en a besoin, dans les limites de ses droits. Il est construit sur **CortexDB**. Une équipe garde sa connaissance dans plusieurs "cerveaux" CortexDB : un pour l'équipe, un pour un client, un pour toute l'organisation. Le gateway de Thalamus est un serveur MCP placé entre les utilisateurs et ces cerveaux. Quand l'assistant pose une question avec `th_recall`, le gateway cherche dans chaque cerveau que l'utilisateur peut lire, avec la clé de cet utilisateur, et fusionne les résultats. Chaque extrait donne sa source : le cerveau, le document et l'auteur.

Les droits sont dans un seul fichier, `rights.yaml`, et un changement s'applique sans redémarrage. Un utilisateur se connecte avec le compte de l'organisation (OpenID Connect), et le gateway lui donne une clé par cerveau. CortexDB vérifie cette clé à chaque appel, donc la portée d'une clé ne dépend pas du gateway. Un document proposé depuis Claude Code ou depuis une page dans Chrome attend la décision d'un curateur. Thalamus aide à la conformité au RGPD et à la LPD suisse révisée : lectures comptées sans nom, durée de conservation de l'audit, source de chaque extrait. L'organisation héberge le gateway et les serveurs, et les cerveaux locaux restent sur la machine de l'utilisateur. Thalamus ne couvre pas le transfert au fournisseur du modèle : le hook met des passages dans le prompt, et le prompt part chez le fournisseur du modèle de l'assistant. L'organisation couvre ce transfert elle-même (contrat de sous-traitance, garanties de transfert, réglage de conservation chez le fournisseur ; RGPD 28, 44 à 46, LPD 9, 16). Un journal des lectures nominatif (`audit_reads: true`) est une donnée personnelle : information des utilisateurs, droit d'accès, durée de conservation. Les articles de l'AI Act cités (12, 14, 26) valent pour un usage à haut risque, et un rappel pour agents de code n'en est pas un en général ; pour un tel usage, Thalamus donne le journal et le contrôle humain par les curateurs.

Dans Claude Code, un hook interroge à chaque prompt les cerveaux partagés et les cerveaux locaux du poste, et chaque utilisateur choisit ses cerveaux avec `/th-recall`. Le rappel sert les passages des documents, avec leur source, et `th_get` lit le document entier. Un cerveau peut avoir une pyramide de synthèses : `th_recall` sert les synthèses d'abord, et `th_descend` ouvre les documents qu'elles citent. Le rappel accepte aussi la question en plusieurs langues. Sur le banc suisse (Constitution fédérale en allemand, français, italien et romanche), 97 % des articles attendus sont dans les 10 premiers résultats avec un encodeur multilingue. Le projet est en construction : la feuille de route ci-dessus dit ce qui est fait et ce qui vient.
