# Fiche 08 contre CortexDB v2.138.0 : `check_against_sources`

Vérification du 23 octobre 2026. Clone `cortexdb/` à `b2a1ab6` (tag v2.138.0). Copie de travail jetable `wt08` (détachée sur v2.138.0), retirée à la fin. Aucune branche ni aucun fichier du projet n'a été modifié, rien n'a été poussé. `CORTEXDB_TEST_POSTGRES` n'a jamais été posé (`env -u`).

Convention : **Fait** = lu dans le code ou observé dans un test lancé. **Hypothèse** = déduction ou estimation non mesurée.

---

## 1. Les prémisses sur v2.138.0

### 1.1 `verify_claims` (Fait)

- Il prend des triplets, pas des phrases : `Claim{Subject, Relation, Object}` (`pkg/cortexdb/verify_claims.go:83-87`). Il cherche des arêtes du graphe et rend `supported`, `contradicted` ou `absent` (`verify_claims.go:39-43`).
- Essai (`TestRecheck08VerifyClaimsIsNotATextCheck`, copie jetable) : un document ingéré dit « Lyria RealTime was released by Google DeepMind in 2025-08. ». La revendication (`Lyria RealTime`, `released_by`, `Google DeepMind`) rend `absent / unknown_subject`. L'ingestion sans extraction ne crée pas l'entité, et `verify_claims` ne lit jamais le texte pour trancher.
- Conclusion : la prémisse de la fiche tient. `verify_claims` ne contrôle pas une phrase contre un texte.

### 1.2 `with_text` (Fait)

- `VerifyClaimsOptions.WithText` (`verify_claims.go:101-102`) charge le texte des morceaux cités par une arête de preuve, et seulement par elle. Dans `claimEvidence` (`verify_claims.go:576-601`), la condition est `withText && len(prov.ChunkIDs) > 0 && e.retractedAt.IsZero()`, puis l'appel passe à `FactProvenanceFor(ctx, e.id, true)`.
- `FactProvenanceFor` (`pkg/cortexdb/fact_provenance.go:61-121`) lit les morceaux par `GetChunks(ToolGetChunksRequest{ChunkIDs, DisableGraph: true})` (`fact_provenance.go:103-106`). Il rend les identifiants introuvables dans `Missing` (`fact_provenance.go:110-119`).
- `GetChunks` (`pkg/cortexdb/graphrag_tool_search.go:416-422`) passe par `loadToolChunks` (`pkg/cortexdb/graphrag_tool_helpers.go:86-127`), qui ne filtre aucune collection. Il cache en revanche un morceau de document expiré ou caché par la porte de relecture (`hiddenAmong`, `graphrag_tool_helpers.go:100-111` ; `pkg/cortexdb/expiry.go:174-199`).
- Essai (`TestRecheck08ChunkLoading`, base SQLite, `WithReviewGate()`) :
  - `get_chunks` sur `chunk:a:000` (collection `c1`), `chunk:b:000` (`c2`) et `chunk:p:000` (`c1`, `proposed`) rend `a` et `b`. Il passe donc d'une collection à l'autre, et il cache `p`.
  - `knowledge_get` de `p` rend le texte du document proposé. Le moteur l'écrit lui-même dans `ReviewGateNote` (`pkg/cortexdb/review.go:30`) : « knowledge_get and memory_get still read a hidden row by id ». La classe d'expiration le confirme : `"knowledge_get": "by id: the record, until the sweep"` (`pkg/cortexdb/expiry_class_test.go`, table `readToolsAndTheExpiry`).
  - `ToolChunk` ne porte pas de collection (`graphrag_tool_types.go:38-45`), et la ligne d'`embeddingsByIDs` non plus (`graph_runtime_sql.go:28-33`). La collection d'un document se lit par `knowledgeCollection` (`pkg/cortexdb/knowledge_helpers.go:150`). Un identifiant de morceau a la forme `chunk:<document>:000`.
- Conclusion : ce qui resservira, c'est `loadToolChunks` pour les morceaux (le texte, avec l'expiration et la relecture). Ce n'est pas `with_text` en tant que tel. Pour un document, `GetKnowledge`/`loadKnowledgeRecord` (`knowledge_api.go:242-248`, `knowledge_helpers.go:96-148`) n'appliquent ni l'expiration ni la porte. Le nouvel outil doit poser `hiddenAmong` lui-même.

### 1.3 `uncited_facts` (Fait)

- L'outil existe sous ce nom (`pkg/cortexdb/graphrag_tool_defs.go:366-367`, `mcp.go:266`, code dans `fact_provenance.go:161-240`). Il liste les **arêtes du graphe** qui ne portent ni `chunk_ids`, ni `document_id`, ni `rule_id`, ni `_source`, en écartant `has_chunk` et `mentions` (`fact_provenance.go:194-223`).
- Il ne lit aucun texte. Sa propre description renvoie le contrôle du texte à `fact_provenance with_text` : « This reports what is missing, not whether a citation is any good » (`graphrag_tool_defs.go:367`). Après une seule ingestion, l'essai rend 0 fait non cité.
- Conclusion : malgré son nom, `uncited_facts` ne fait rien de ce que propose la fiche. Il répond à « quelle arête n'a pas de source », pas à « quel fait d'une phrase manque dans ses sources ».

### 1.4 Le contrat `_grade` (Fait)

- Les cinq grades sont dans `pkg/cortexdb/contract.go:106-121`. `GradeSelfConsistent` : « internal coherence only. Derived deterministically from something already stated; not checked against the world » (`contract.go:109-111`). `GradeAsserted` : « a source or a model said so and nothing has checked it » (`contract.go:112-115`).
- `KeyState` : « the producer's own word … never interpreted across producers » (`contract.go:45-47`).
- La question de la fiche reste donc ouverte. Le contrôle est déterministe, mais l'écriture par un modèle ne l'est pas.

### 1.5 Un outil existant fait-il une partie du contrôle ? (Fait)

- Aucun outil de v2.138.0 ne compare une phrase à un texte. La liste des noms d'outils (`graphrag_tool_defs.go` et les `*_tooldefs.go`) ne contient ni `check_against_sources` ni rien d'approchant. Une recherche de `grounded`, `faithful`, `hallucinat` et `against_sources` dans `pkg/` et `cmd/` ne trouve aucun contrôle de ce genre.
- Les morceaux réutilisables : `loadToolChunks` (texte des morceaux, avec expiration et relecture), `knowledgeCollection`, `hiddenAmong`/`allowDocument`, `knowledgeReadPermitted` (`pkg/cortexdb/memory_owner.go:117-120`, un seul usage aujourd'hui, `knowledge_api.go:283`).
- `foldName` (`graphrag_tool_find_nodes.go:169-178`) plie la casse et retire la ponctuation, mais garde les accents. Le moteur n'a aucun pliage des accents en Go. `golang.org/x/text` v0.36.0 est déjà dans `go.mod`, en dépendance indirecte (`go.mod:58`). L'utiliser directement n'ajoute aucun module.

### 1.6 Les clés confinées (Fait)

- Essai (`TestRecheck08Refused`, `pkg/authz`) : `verify_claims`, `uncited_facts` et `fact_provenance` sont refusés à une clé confinée avec la raison par défaut « reaches state that no scope field confines » (`pkg/authz/allowlist.go:271`, et le remplissage dans `init`, `allowlist.go:276-284`).
- `get_chunks` est refusé pour « names chunks, whose document the guard does not load » (`allowlist.go:255`). `knowledge_get` est admis, par identifiant seul (`allowlist.go:218`).
- Le précédent utile est `plan.filters.document_ids`, traité comme une **donnée** et non comme un identifiant de ligne. Le filtre de collection, forcé, décide de ce que l'épingle trouve : un document d'une autre collection ne correspond à rien, exactement comme un document qui n'existe pas (`allowlist.go:120-127`). Les `sources` de l'outil suivent la même voie, ce qui rend la règle 5 de la fiche tenable pour les morceaux aussi.

### 1.7 Le contrôle actuel de contaixt (Fait, lecture seule)

Fichier `~/Agents/science-extractor/scripts/synthese.py` :

- `faits` (`synthese.py:130-149`) :
  - les nombres viennent de `\d+(?:[.,]\d+)*`, avec la virgule changée en point ;
  - les mots viennent de `[\w][\w.+-]*` quand ils ont une majuscule après la 1re lettre, un chiffre, ou une majuscule initiale et 4 caractères ou plus ; un mot fait de chiffres seuls est exclu ;
  - le 1er mot est retiré seulement s'il a la forme `[A-ZÀ-Ý][a-zà-ÿ]+` ;
  - après une ponctuation interne, un mot n'est retiré que s'il est dans `DEBUTS`, une **liste de mots français** (`synthese.py:120-126`).
- `fait_present` (`synthese.py:152-165`) : sous-chaîne pliée, **sans borne de mot**. Il ajoute un pluriel en `s` pour 4 lettres ou plus, l'ordinal `_ORDINAL` (`er|ere|re|e|eme|st|nd|rd|th`, une liste française et anglaise, `synthese.py:127`) et `EQUIVALENTS` (`synthese.py:99-119`).
- `verifier` (`synthese.py:168-171`) plie les sources par `_plier` (`interroger_graphe.py:322-325` : NFKD, `casefold`, sans diacritiques) et change la virgule en point.

Écarts entre la fiche et ce code (Fait) :

- La règle 1 de la fiche ne dit ni `DEBUTS`, ni l'exclusion des mots faits de chiffres, ni la forme exacte du 1er mot. « IBM » en tête reste un fait.
- La fiche ne dit ni le pluriel ni l'ordinal. Les deux sont des règles de langue, ce qui contredit sa propre règle 3.

Essais sur une copie des fonctions (`py08/probe.py`, Python `-I`), avec ce que chaque cas montre :

| Phrase | Source | Résultat de contaixt | Ce que le cas montre |
| :- | :- | :- | :- |
| « Le modèle AI a 7 couches. » | « it is said there are 17 layers » | tient | `7` est trouvé dans `17` et `ai` dans `said` : faux positifs de la sous-chaîne. |
| « 谷歌于2025年发布Lyria。 » | « 谷歌在2025年发布了Lyria模型。 » | tombe | Le fait absent est **toute la suite** `谷歌于2025年发布lyria`. En Python, `\w` couvre les idéogrammes et un texte CJK n'a pas d'espace. Lyria n'est jamais isolé. |
| « La CNIL a dit. … Elle confirme. » | « CNIL … Candice Odgers » | tient | `Elle` est écarté par `DEBUTS`, une liste française. |

Piège de portage (Fait, connu de Go) : en Go, `\w` et `\d` du paquet `regexp` ne couvrent que l'ASCII, alors qu'ils sont Unicode en Python. Un portage littéral perdrait « Müller » et les chiffres pleine chasse.

---

## 2. La PR est-elle encore nécessaire ?

**Oui, sous une forme changée.** Aucun outil de v2.138.0 ne fait ce contrôle (section 1.5). `uncited_facts` a un nom proche mais un autre objet (section 1.3). La fiche doit changer sur six points :

1. **Le chargement des sources.** Il passe par `loadToolChunks` pour les morceaux, et par une lecture du document **suivie de `hiddenAmong`** pour les documents, au lieu de `knowledge_get`. Une source expirée, ou cachée par la porte de relecture, compte comme absente. Une source d'une autre collection aussi. `with_text` n'est pas réutilisé tel quel.
2. **Plus aucune liste de langue dans le moteur.** `DEBUTS`, le pluriel en `s` et `_ORDINAL` sortent du moteur. Le producteur garde ces règles chez lui, ou les passe par `equivalents`.
3. **Le cas CJK.** Une suite d'idéogrammes, de kana ou de hangul se coupe de ses voisins latins et chiffrés, et n'est jamais un fait. Ce qui reste vérifiable en CJK : les nombres et les noms latins.
4. **Les nombres** se cherchent entre deux non-chiffres. C'est un écart volontaire à contaixt (le cas `7` dans `17`).
5. **L'entrée dans la liste blanche** se fait avec `sources` en donnée (précédent de `plan.filters.document_ids`) et non en identifiants. Sans cela, le motif de refus de `get_chunks` s'appliquerait.
6. **La taille.** La fiche estimait environ 250 lignes ; l'estimation est maintenant d'environ 300 à 340 lignes hors tests (section 3.4).

---

## 3. Le dessin mis à jour

### 3.1 L'outil

Nom : `check_against_sources`. Il ne fait que lire (`Mutates: false`).

Entrée :

```json
{
  "collection": "ai-music",
  "sentences": [{"text": "…", "sources": ["doc-id", "chunk:doc-id:003"]}],
  "equivalents": {"rgpd": ["gdpr", "dsgvo"]}
}
```

Sortie :

```json
{"results": [{"holds": false, "missing": ["…"], "sources_missing": ["…"]}],
 "counts": {"sentences": N, "hold": H}}
```

Les résultats suivent l'ordre de la requête.

### 3.2 Les règles, une par étape (pur sauf l'étape 6)

**Étape 1, normaliser.** La phrase passe en NFKC (`golang.org/x/text/unicode/norm`) avant l'extraction : les chiffres et les lettres pleine chasse deviennent ASCII, la casse est gardée.

**Étape 2, découper.** Un jeton est une suite maximale de lettres Unicode, de chiffres Unicode et de `_ . + -`. Ses `. - +` de bord sont ôtés. Le jeton est coupé à chaque passage entre une rune CJK (Han, Hiragana, Katakana, Hangul) et une rune non CJK.

**Étape 3, extraire les faits.**
- (a) Les nombres : `\d+(?:[.,]\d+)*` sur la phrase normalisée, la virgule devenue point.
- (b) Un jeton est un fait de mot si :
  - il n'est pas CJK ;
  - il n'est pas fait de chiffres seuls ;
  - et il a soit un chiffre, soit une majuscule après sa 1re rune, soit une majuscule initiale et 4 lettres ou plus.
- (c) Le 1er jeton de la phrase n'est pas un fait s'il a la forme Majuscule puis minuscules seules.

Aucune liste de mots. Une phrase = une entrée : c'est le producteur qui découpe.

**Étape 4, plier.** Minuscules, NFKD, retrait des marques `Mn`, `ß` → `ss`, virgule → point. Le pliage s'applique des deux côtés : faits, sources et équivalents (clés et valeurs).

**Étape 5, chercher.** Un fait de mot est présent s'il est sous-chaîne de la source pliée, comme chez contaixt. Un fait numérique (chiffres, `.` et `-` seuls) est présent seulement entre deux non-chiffres. Un fait est aussi présent si l'un de ses équivalents pliés de la requête l'est, selon la même règle.

**Étape 6, lire les sources (la seule entrée-sortie).** Un identifiant `chunk:` se lit par `loadToolChunks`, les autres par la ligne du document. Pour l'un comme pour l'autre :
- la collection du document est lue par `knowledgeCollection` ;
- la source est absente si :
  - elle n'existe pas ;
  - elle est cachée par `hiddenAmong(ctx, nil, ids)` (expirée, ou cachée par la porte) ;
  - `collection` est donnée et diffère ;
  - ou `knowledgeReadPermitted(ctx, collection)` est faux.
- Un seul point d'étranglement : `func (db *DB) sourceTexts(ctx, collection, ids) (texts map[string]string, missing []string, err error)`.

**Étape 7, trancher.** `holds = len(sources) > 0 && len(sources_missing) == 0 && len(missing) == 0`. Une phrase sans source ne tient pas. Les faits sont cherchés dans le texte joint des sources trouvées.

**Bornes.** 500 phrases et 50 sources par phrase au plus ; au-delà, l'outil rend une erreur. **Hypothèse** : les valeurs sont à fixer par le propriétaire.

### 3.3 Les fichiers

- `pkg/cortexdb/check_sources.go` (nouveau) : le cœur pur (étapes 1 à 5 et 7), `sourceTexts` (étape 6), les types de requête et de réponse, `CheckAgainstSourcesTool`.
- `pkg/cortexdb/graphrag_tool_defs.go` : la définition et le schéma (à côté de `verify_claims`, ligne 376), puis un `case` dans l'aiguillage de `Call` (vers les lignes 860-875).
- `pkg/cortexdb/mcp.go` : `addGraphRAGMCPTool` (à côté de la ligne 260).
- `pkg/authz/allowlist.go` : une ligne dans `confinedTools`.
- `go.mod` : `golang.org/x/text` passe de dépendance indirecte à directe.
- Tests : `pkg/cortexdb/check_sources_test.go`, et une entrée chacune dans `tool_mutates_test.go` (`false`) et `expiry_class_test.go` (`"gated: TestAnExpiredOrProposedSourceIsMissing"`). Le test de complétude `pkg/authz/allowlist_test.go:53` tient la ligne de la liste blanche, un test confiné va dans `pkg/authz`.

### 3.4 Les lignes estimées hors tests (Hypothèse, appuyée par un prototype)

Le prototype du cœur pur (`recheck08_core.go`, copie jetable) fait 152 lignes non vides et non commentées. Il couvre les étapes 1 à 5 et 7, et il manque encore l'exclusion des jetons faits de chiffres seuls.

| Partie | Lignes |
| :- | -: |
| Cœur pur | 150 à 160 |
| `sourceTexts` | 50 à 60 |
| Types et outil | 40 à 50 |
| Définition et schéma | 35 à 40 |
| Aiguillage, MCP, liste blanche, `go.mod` | environ 20 |
| **Total** | **environ 300 à 340** |

Le total reste sous la limite de 400 lignes. Pour « une règle par PR », on peut couper si le propriétaire le demande : PR A, l'outil sans entrée confinée (il tombe alors dans `defaultRefusal`) ; PR B, la ligne de la liste blanche avec son test confiné (environ 5 lignes hors tests).

### 3.5 Tableau outil par outil (ce qu'une clé confinée atteint)

| Outil | Sous une clé confinée à `collection` | Argument | Classe |
| :- | :- | :- | :- |
| `check_against_sources` | admis, `scope: collectionScope`, `stores: storeKnowledge` | `collection` | portée, forcée à la valeur de la clé |
| | | `sentences`, `sentences[].text` | donnée |
| | | `sentences[].sources` | donnée : la façade filtre par collection, une source étrangère est `sources_missing` comme une source inexistante (précédent `allowlist.go:120-127`) |
| | | `equivalents.*` | donnée |
| `verify_claims`, `uncited_facts`, `fact_provenance` | refusés (inchangé) | | `defaultRefusal` |
| `get_chunks` | refusé (inchangé) | | identifiants de morceaux non gardés |

### 3.6 Les tests à écrire d'abord (noms proposés)

Cœur pur, sans base :

- `TestAFactIsANumberOrAMarkedWord` : MusicLM, DeepMind, GPT-4, Oslo au milieu d'une phrase, « 1,5 » qui donne `1.5`.
- `TestTheFirstWordIsNotAFactWhenOnlyItsInitialIsCapital` : « Lyria … » écarté, « IBM … » gardé.
- `TestADigitOnlyTokenIsANumberNotAWord`.
- `TestFoldingIgnoresCaseAndAccentsOnBothSides` : Müller contre MULLER, Straße contre STRASSE.
- `TestANumberIsFoundOnlyBetweenNonDigits` : `7` contre « 17 layers » est absent ; `2025` contre « in 2025, » est présent.
- `TestAnEquivalentFromTheRequestStandsForTheFact` : RGPD contre GDPR.
- `TestTheEngineHoldsNoWordList` : RGPD contre GDPR sans équivalent est absent ; « Elle » après un point reste un fait.
- `TestHanKanaAndHangulRunsAreNoFactAndAreSplitFromLatin` : 谷歌于2025年发布Lyria donne {2025, lyria}.
- `TestFullwidthDigitsAreNumbers`.

Façade, SQLite :

- `TestCheckAgainstSourcesReadsDocumentsAndChunks`
- `TestAMissingSourceIsNamedAndTheSentenceDoesNotHold`
- `TestASentenceWithNoSourceDoesNotHold`
- `TestASourceOfAnotherCollectionIsMissing`
- `TestAnExpiredOrProposedSourceIsMissing` : l'expiration, puis la porte de relecture.
- `TestACallerScopeOutsideTheCollectionSeesNoSource` : `CallerScope`.

Clé confinée, `pkg/authz` :

- `TestAConfinedKeyChecksOnlyItsCollection` : `collection` étrangère refusée, `collection` absente refusée, `sentences[].sources` admis en donnée.

### 3.7 Les mutations à la main (règle → test qui doit tomber)

| Mutation | Test qui doit tomber |
| :- | :- |
| Exclusion du 1er mot retirée | `TestTheFirstWordIsNotAFactWhenOnlyItsInitialIsCapital` |
| Condition « chiffre » retirée de l'étape 3b | `TestAFactIsANumberOrAMarkedWord` (GPT-4) |
| Condition « majuscule après la 1re rune » retirée | idem (MusicLM) |
| Seuil `>= 4` changé en `>= 5` | idem (Oslo) |
| Exclusion des jetons faits de chiffres seuls retirée | `TestADigitOnlyTokenIsANumberNotAWord` |
| Retrait des marques `Mn` supprimé | `TestFoldingIgnoresCaseAndAccentsOnBothSides` |
| Pliage d'un seul côté | idem |
| Borne des nombres retirée (sous-chaîne) | `TestANumberIsFoundOnlyBetweenNonDigits` |
| `equivalents` ignorés | `TestAnEquivalentFromTheRequestStandsForTheFact` |
| Coupe CJK retirée | `TestHanKanaAndHangulRunsAreNoFactAndAreSplitFromLatin` |
| NFKC retiré | `TestFullwidthDigitsAreNumbers` |
| Filtre de collection retiré | `TestASourceOfAnotherCollectionIsMissing` |
| `hiddenAmong` retiré | `TestAnExpiredOrProposedSourceIsMissing` |
| `knowledgeReadPermitted` retiré | `TestACallerScopeOutsideTheCollectionSeesNoSource` |
| `holds` vrai malgré une source absente | `TestAMissingSourceIsNamedAndTheSentenceDoesNotHold` |
| `holds` vrai sans source | `TestASentenceWithNoSourceDoesNotHold` |
| `scope` ôté de la ligne de la liste blanche | `TestAConfinedKeyChecksOnlyItsCollection` |

---

## 4. Questions ouvertes pour le propriétaire

1. **Le grade d'une synthèse** (`contract.go:109-115`). Une synthèse écrite par un modèle et contrôlée par cet outil : `self_consistent`, ou `asserted` avec le résultat dans `KeyState` (par exemple `checked: 41/43`) ? Proposition (Hypothèse) : `asserted`, puisque la dérivation n'est pas déterministe. L'outil ne pose de toute façon aucun grade.
2. **Les écritures CJK.** La proposition réduit le contrôle aux nombres et aux noms latins. Le propriétaire veut-il des suites Han de N caractères ou plus comme faits ? Il faudrait alors une segmentation, que le moteur n'a pas.
3. **Le pluriel et l'ordinal.** Ils sortent du moteur. Faut-il un champ de requête `suffixes` (mécanisme neutre, données venues de la requête), ou laisser `equivalents` suffire ?
4. **La sous-chaîne pour les mots.** Elle donne des faux positifs (`ai` dans `said`), mais elle tolère la flexion. Faut-il une borne de mot, plus stricte que contaixt ?
5. **Une source proposée hors porte.** Sans `CORTEXDB_REVIEW_GATE`, un document proposé compte comme source. Faut-il l'écarter toujours ?
6. **La dépendance directe** à `golang.org/x/text`, ou un pliage écrit à la main ?
7. **Les bornes** : nombre de phrases, de sources, octets de texte.

## 5. Avant l'envoi (reste à faire)

- Rejouer contaixt et l'outil sur les mêmes synthèses d'ai-music. **Hypothèse** : l'outil fera tomber **plus** de phrases que contaixt, pour trois raisons : la fin de `DEBUTS`, la fin du pluriel et de l'ordinal, et la borne des nombres. L'écart doit être mesuré et justifié phrase par phrase ; il n'a pas été mesuré ici.
- Mettre à jour la fiche `docs/upstream/08-synthesis-check.md` d'après ce rapport (non fait : la consigne interdisait de modifier le projet).
