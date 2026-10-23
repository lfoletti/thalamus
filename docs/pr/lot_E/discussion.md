# Billet de discussion : présentation de Thalamus et lot E (brouillon)

Catégorie : **Show and tell**. À poster sur commande de Lauro, en même temps que les PR du lot E. Titre :

> Thalamus: shared CortexDB brains for an organisation and its coding agents

---

Hello! Discussion #66 ended with the series merged. This post introduces what the series was for: **Thalamus**, a new open-source project built on CortexDB, now public at https://github.com/lfoletti/thalamus (MIT).

## What it is

Thalamus brings the knowledge of an organisation to each employee and to their coding agent, at the moment they need it, within the rights of each person. A team keeps its knowledge in several CortexDB brains: one for the team, one for a client, one for the whole organisation. Thalamus is a thin gateway between the employees and these brains.

- **One question, every brain the person may read.** `th_recall` asks each brain with the employee's own key and fuses the answers by weighted reciprocal rank. Each passage carries its source: the brain, the document, the author.
- **Rights declared once.** A `rights.yaml` file lists the brains, the groups and who reads, writes or curates each brain. The gateway turns it into CortexDB keys, one per person and per brain, confined to the brain's collection. CortexDB checks every key on every call, so the scope of a key never depends on the gateway.
- **Login with the organisation's account** (OpenID Connect). The gateway issues the keys the rights allow.
- **Knowledge added in one step.** An employee proposes a document from Claude Code or from a page in Chrome. It enters with `review_status: proposed`, and a curator accepts or rejects it.
- **Recall at each prompt.** A Claude Code plugin adds the best passages of the shared brains and of the local brains of the machine to each prompt, with their source.
- **A pyramid of syntheses** per brain: syntheses served first, with the documents they cite one call away.
- **Questions in several languages.** The first organisation we design for is Swiss: French, German, Italian and Romansh.

## What it uses in CortexDB

Thalamus imports CortexDB v2.138.0 as a library and calls `cortexdb-grpc` with confined keys. It relies on the allowlist of tools for a confined key (#72 to #74), the content hash on ingest (#75), `expires_at` (#76), the review gate (#77), and `written_by` from the calling key. Thalamus itself is about 4,000 lines of Go outside tests. The rule we keep: a mechanism that serves every CortexDB user goes to CortexDB; Thalamus keeps only the organisation's policy (rights, groups, curators, federation between servers).

## Measures

Two benches, both in the repository.

- **A research brain** (a pilot brain on audio and AI, 41 questions in French over mostly English sources, share of the expected facts found in the context): the hook gave 51.7 % when it served the short snippets of `knowledge_search`, and 81.3 % since it serves the `chunks` of the same answer. With one English variant of each question, searched apart and fused, `th_recall` reaches 85.1 %.
- **A Swiss bench** (`tests/eval_ch/`): the 219 articles of the Federal Constitution, each stored in one language only (German, French, Italian or Romansh), and 40 questions never in the language of their article. Expected article within the first 10 results:

| Configuration | Question only | With the question in the 2 other languages |
| :- | -: | -: |
| lexical index only | 7 % | 62 % |
| `paraphrase-multilingual-MiniLM-L12-v2` | 78 % | 88 % |
| `multilingual-e5-large` | 82 % | 97 % (Romansh 91 %) |

Two of the PRs below come from these benches.

## The next PRs

Sixteen PRs, in four groups. Line counts are outside tests. `#<gc-NN>` stands for the PR numbers, filled in when they are open.

**Graph tools for a confined key.** Today the allowlist refuses every graph tool to a key confined to one collection, with good reason: under a `CallerScope`, the graph tools serve chunk text, relations, entity descriptions and even memories of other collections. These PRs make the facade hold the scope, then open the tools.

| PR | Change | Lines | On |
| :- | :- | -: | :- |
| #<gc-13> | `find_nodes` leaves out the nodes of hidden documents and memories, as `get_nodes` does (it served the text of an expired document) | 22 | main |
| #<gc-14> | Under a caller scope, the graph tools leave out the documents and memories outside it, through `hiddenAmong` | 96 | main |
| #<gc-15> | Under a caller scope, an entity no permitted collection lists is not served, and a served one keeps its name and type | 118 | #<gc-14> |
| #<gc-16> | The allowlist opens seven graph tools to a confined key, confined by the facade (`find_nodes` follows #<gc-13>) | 64 | #<gc-15> |
| #<gc-17> | The graph lane of a query with a collection walks only that collection's graph (today a 2-hop walk from a seed of another collection reaches a chunk, also through `/v1/query`) | 221 | main |
| #<gc-18> | The allowlist opens `cortex_query` to a confined key, `prefetch[].max_hops` refused until #<gc-17> | 30 | main |

**Keys and signatures.**

| PR | Change | Lines | On |
| :- | :- | -: | :- |
| #<gc-19> | The server reads the key file again when it changes: a revoked key is refused without a restart; a broken file keeps the keys in force | 184 | main |
| #<gc-20> | `extract_conversation` reads a session only when the caller scope permits its bucket | 10 | main |
| #<gc-21> | `ingest_document` signs `written_by` from the calling key, and the document tools take the signature back | 40 | main |

**Tools.**

| PR | Change | Lines | On |
| :- | :- | -: | :- |
| #<gc-22> | The request log names the rows a tool call names in its arguments | 82 | main |
| #<gc-23> | `knowledge_get` reads a list of ids, as `memory_get` does | 64 | main |
| #<gc-28> | `check_against_sources`: the facts of a sentence that its cited sources do not contain (no word list in the engine; equivalents come from the request) | 358 | main |

**Search in European languages.**

| PR | Change | Lines | On |
| :- | :- | -: | :- |
| #<gc-24> | A Swiss evaluation set from the Federal Constitution (de, fr, it, rm), with a floor per category | 11 + data | main |
| #<gc-25> | The coverage floor of v2.138.0 compares words without accents, as the word index does ("universitat" finds "Universität" again) | 60 | main |
| #<gc-26> | The words of `alternate_queries` join the OR keyword query (a variant written as a sentence returned nothing) | 5 | main |
| #<gc-27> | `alternate_queries` run even when the question fills `top_k` | −6 | main |

Order of merge, if it helps: #<gc-13> before #<gc-16>; #<gc-14>, #<gc-15>, #<gc-16> in that order; #<gc-17> before a later change that allows `max_hops`. #<gc-16> and #<gc-18> both edit the allowlist, and #<gc-13> and #<gc-14> both edit the class tests of the read tools, so the second of each pair needs a rebase, which I will do.

Questions where your design choice comes first, so no PR:

1. `unaccent` on PostgreSQL, so that accents count the same on both stores. Optional extension or required?
2. Compound words (German): #<gc-25> keeps the rule of v2.138.0 that a term inside a word does not count. A fallback on the trigram index would need that rule to change for terms of 6 letters or more. Would you accept it?
3. Stopwords for other languages: not proposed. Since v2.138.0 the rarity weight does that work without a list.
4. The grade of a synthesis that `check_against_sources` holds: `self_consistent`, or `asserted` with the check in its state? The question is also in #<gc-28>.

Each PR body has its own questions where a rule needed a choice.

Each PR stands alone, is based on `main`, keeps to about 400 lines outside tests and to one rule, and carries its failing-first tests. None changes the behaviour of a server without keys or of an unconfined key. Happy to drop or reshape any of them.

Thank you for CortexDB, and for the reviews.
