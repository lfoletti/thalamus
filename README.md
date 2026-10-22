<p align="center"><img src="extension/icons/icon128.png" alt="Thalamus logo: two pale lobes on an indigo square" width="96"></p>

# Thalamus

Thalamus brings the knowledge of an organisation to each employee and to their coding agent, at the moment they need it, within the rights of each person. It is built on [CortexDB](https://github.com/liliang-cn/cortexdb).

> **Work in progress.** The interfaces change between versions. A short French summary is at the end.

## What it does

A team keeps its knowledge in several CortexDB brains: one for the team, one for a client, one for the whole organisation. Each employee may read some of them and write to fewer. Thalamus is the gateway between the employees and these brains.

The employee's assistant reaches the brains through the gateway, which is an MCP server. When the assistant asks a question with `th_recall`, the gateway searches each brain that this employee may read, with the employee's own key, and merges the results. Each extract comes with its source: the brain, the document and the author. With `th_get`, the assistant reads the full document behind an extract.

An employee can also propose a document, from Claude Code or from a page in Chrome. The document waits for a curator of the brain, who accepts or rejects it.

## Why use it

- **Knowledge at the moment of need.** One question reaches every brain that the employee may read. The employee does not choose where to look. Each extract gives its source, so the employee can open the full document. In Claude Code, a hook brings the extracts at each prompt (see [Automatic recall in Claude Code](#automatic-recall-in-claude-code)).
- **Access without requests.** An employee logs in with the account of the organisation, and the gateway issues the keys that the rights allow. To give a person access to a team's brains, add the login to the team's group in `rights.yaml`. The change applies at the next call, without a restart.
- **Knowledge added in one step.** An employee proposes a document from Claude Code or from a page in Chrome. A curator of the brain accepts or rejects it, from the same tools.
- **Brains of the correct size.** A brain is a whole CortexDB server or one collection of a shared server. Small teams can share one server. A sensitive brain can have its own server. Each brain can have a weight in the merged results.
- **Access that stays under control.** Each person has a personal key, and CortexDB checks it on every call. To revoke one person, the gateway removes that person's keys only. The audit records the writes, the reviews and the logins.

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
- An employee logs in with the organisation's identity provider (OpenID Connect). The gateway then issues the keys that the rights allow.

## Automatic recall in Claude Code

The Thalamus plugin for Claude Code adds a hook to the tools of the gateway. At each prompt, the hook asks the gateway for the best extracts and gives them to the agent before it answers. The employee does not ask for them.

<p align="center"><img src="docs/img/recall-hook.svg" alt="The employee types a prompt. The Thalamus hook of the plugin sends it to th_recall on the Thalamus gateway, with the employee's key. The gateway searches the readable brains. The best extracts come back with their source, and the hook adds them to the prompt before the agent answers." width="760"></p>

- The hook (`th-hook`, on the `UserPromptSubmit` event) sends the prompt to `th_recall` with the employee's key. It adds the 3 best extracts to the prompt, with their source, in 9,000 characters or fewer.
- It does not ask for a prompt of fewer than 3 words, or for a `/` command.
- It never stops a prompt. If the gateway does not answer, or if the key is refused, the prompt continues without extracts.
- The agent can open the full document of an extract with `th_get`.

Planned: the hook also searches the local brain of the machine, and merges its extracts with the extracts of the server brains. The local brain stays on the machine. It does not go through the gateway.

Other assistants use the same tools through MCP. The automatic recall needs a hook in the assistant. Thalamus has one for Claude Code.

## Try it

The steps start one CortexDB server with two brains, then the gateway, then the plugin. Do all the steps in the root of this repository.

1. Build the binaries. `cortexdb/` is a clone of CortexDB next to `gateway/`.

   ```sh
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

**Done** (22 October 2026)

- Rights, keys and one MCP endpoint per brain.
- Federated recall over the brains that a person may read, with the source of each extract.
- Proposed documents and their review by a curator.
- REST endpoints for clients without MCP.
- Login with OpenID Connect, from the command line and from the extension.
- The Claude Code plugin: recall hook, tools and the `/th-add` command.
- The Chrome extension: propose a page or a selection. It keeps the main content of the page.
- A proportionate audit: named writes, reviews, logins and refusals, reads counted without a name, named reads only on the brains that ask for it, and a retention period.

**Next in Thalamus**

- The recall hook also searches the local brain of the machine, and merges its extracts with the extracts of the server brains.

**With CortexDB**

Thalamus relies on CortexDB for each mechanism that serves every CortexDB user. These parts go to CortexDB as small pull requests.

- In review: content hash and duplicate check (#75), expiry dates (#76), review status (#77). The review of proposed documents needs #77.
- Prepared: the server reads its key file again when it changes. Then a new or revoked key applies without a restart (`docs/upstream/01-key-file-reload.md`).
- Prepared: graph tools for a key confined to one collection. Then Thalamus can merge entities across brains, in addition to text extracts (`docs/upstream/06-graph-tools-confined.md`).
- Prepared, smaller: a tool that lists the documents waiting for review, the touched ids in the server's request log, and the author of a document taken from the caller's key.

The prepared proposals go to CortexDB after #75 to #77 are settled and this repository is public. The repository becomes public at the end of the current series of fixes.

## Principles

1. **The engine checks every key.** The gateway decides who gets a key. CortexDB decides what each key may do.
2. **One key per person and per brain.** The gateway can revoke the access of one person on one brain.
3. **Each extract shows its source.** The reader can open the document behind an extract with `th_get`.
4. **General mechanisms go upstream.** Thalamus keeps only the organisation's policy: the rights, the groups, the curators and the federation between servers.

## Development

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

Thalamus donne à une équipe et à ses agents de code des bases de connaissances partagées, avec des droits par personne. Il est construit sur CortexDB. Le gateway lit les droits dans `rights.yaml` et donne à chaque personne une clé par cerveau. Il cherche dans tous les cerveaux qu'une personne peut lire et donne la source de chaque extrait. Un curateur relit les documents proposés. Le projet est en construction : la feuille de route ci-dessus dit ce qui est fait et ce qui vient.
