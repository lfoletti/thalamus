# Billet de discussion : les lots de PR qui viennent (brouillon)

Catégorie : **Ideas** (ou « General », au choix de Lauro). À poster sur commande de Lauro, en même temps que les PR du lot A (`gc-19`, `gc-20`, `gc-21`). Les lots B, C et D partent ensuite un par un, chacun quand le précédent est relu. Les nombres de lignes et les numéros `#<gc-NN>` se remplissent après les amendements du niveau 2 et l'ouverture des PR ; les PR des lots suivants sont nommées par leur sujet tant qu'elles n'ont pas de numéro. Titre :

> Next PR series: search in European languages, graph tools for confined keys, key file reload, three tools

---

Hello! Discussion #66 ended with the series merged. Here are the improvements we propose next: sixteen PRs in four lots. This post opens lot A as drafts. Lots B, C and D follow one at a time, each one when you have read the one before. `#<gc-NN>` stands for the PR numbers.

They have two aims:

- **More languages.** Extend the search of CortexDB to more languages, beyond English and Chinese, starting with the European context: accents, several languages in one brain, a question in one language and the answer in another. The rule stays the one you set: no word list in the engine, because a list serves only the languages that someone wrote down.
- **One shared server for several teams.** Open the graph tools and `cortex_query` to a key confined to one collection. Also: a revoked key is refused without a restart, each ingested document is signed by its writer, and the request log names the rows that a call names.

Each PR keeps to one rule and to about 400 lines outside tests, and carries its failing-first tests. Each one is based on `main`, except two PRs of lot C, which are stacked. These are proposals: happy to drop or reshape any of them.

## Lot A: keys and signatures (opened with this post)

| PR | Change | Lines | On |
| :- | :- | -: | :- |
| #<gc-19> | The server reads the key file again when it changes: a revoked key is refused without a restart, and a broken file keeps the keys in force | 214 | main |
| #<gc-20> | `extract_conversation` reads a session only when the caller scope permits its bucket | 10 | main |
| #<gc-21> | `ingest_document` signs `written_by` from the calling key, as the other document tools do | 40 | main |

## Lot B: three tools (next)

| PR | Change | Lines |
| :- | :- | -: |
| Request log | The request log names the rows that a tool call names in its arguments | 82 |
| `knowledge_get` | `knowledge_get` reads a list of ids, as `memory_get` does | 64 |
| `check_against_sources` | For each sentence, the facts that its cited sources do not contain. No word list in the engine; equivalents between languages come from the request | 358 |

## Lot C: graph tools for a key confined to one collection (after lot B)

Today the allowlist keeps the graph tools closed to a confined key. This lot proposes that the facade applies the caller scope to each graph tool: chunk text, relations, entity descriptions and memories. Then the allowlist can open the tools.

| PR | Change | Lines |
| :- | :- | -: |
| `find_nodes` | `find_nodes` leaves out the nodes of hidden documents and memories, as `get_nodes` does | 22 |
| Scope of the graph tools | Under a caller scope, the graph tools leave out the documents and memories outside it, through `hiddenAmong` | 96 |
| Scoped entities | Under a caller scope, an entity that no permitted collection lists is not served; a served one keeps its name and type. Stacked on the PR above | 118 |
| Allowlist, graph tools | The allowlist opens seven graph tools to a confined key, confined by the facade. Stacked on the PR above | 64 |
| Graph lane of a query | The graph lane of a query with a collection walks only that collection's graph, also through `/v1/query` | 221 |
| Allowlist, `cortex_query` | The allowlist opens `cortex_query` to a confined key, with `prefetch[].max_hops` refused until the graph lane PR | 30 |

## Lot D: search in European languages (after lot C)

These PRs come from a Swiss bench, in `tests/eval_ch/` of the repository named at the end: the 219 articles of the Federal Constitution, each stored in one language only (German, French, Italian or Romansh), and 40 questions never in the language of their article. With the lexical index only, 7 % of the expected articles are in the first 10 results. With the question also given in the two other languages, 65 %.

| PR | Change | Lines |
| :- | :- | -: |
| Swiss evaluation set | A Swiss evaluation set from the Federal Constitution (de, fr, it, rm), with a floor per category | 11 + data |
| Accents in the coverage floor | The coverage floor compares words without accents, as the word index does ("universitat" finds "Universität" again) | 60 |
| Words of the variants | The words of `alternate_queries` join the OR keyword query. A word that is very frequent in the brain does not count as a content word, in any language, without a word list | 5 |
| Variants that fill `top_k` | When a question has alternate queries or keywords, they run even when the question fills `top_k`, and the lists are merged by rank. A plain question is unchanged | −6 |

## Order of merge

- Lot A: the three PRs are independent.
- Lot B: the request log PR and the `knowledge_get` PR meet on one line: the log must name `knowledge_ids`. I will rebase the second one.
- Lot C: `find_nodes` before the graph tools allowlist. The scope, scoped entities and allowlist PRs in that order. The graph lane PR before a later change that allows `max_hops`.
- Lot D: the evaluation set before the accents PR, which raises a floor of the Swiss set.

## Questions where your design choice comes first

1. `unaccent` on PostgreSQL, so that accents count the same on both stores: optional extension or required? (lot D)
2. Compound words (German): the accents PR keeps the rule that a term inside a word does not count. A fallback on the trigram index would need that rule to change for terms of 6 letters or more. Would you accept it? (lot D)
3. The grade of a synthesis that `check_against_sources` holds: `self_consistent`, or `asserted` with the check in its state? (lot B)

Each PR body has its own questions where a rule needed a choice.

## Where these PRs come from

They come from **Thalamus** (https://github.com/lfoletti/thalamus, MIT), a thin gateway built on CortexDB. It gives each user in an organisation, and their coding agent, the knowledge of that organisation, from several brains, within the rights of each person. A mechanism that serves every CortexDB user goes to CortexDB, and Thalamus keeps only the organisation's policy.

Thank you for CortexDB, and for the reviews.
