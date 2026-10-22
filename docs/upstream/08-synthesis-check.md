# 08. Le contrôle d'une synthèse contre ses sources

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : une PR, environ 250 lignes hors tests. **Dépend de** : rien dans le moteur. Les conditions de départ du `README.md` s'appliquent.

## Le besoin

Un producteur écrit des synthèses avec un modèle et les stocke dans CortexDB : une synthèse = un document, avec ses preuves en métadonnée (voir `docs/notes/pyramide.md`). Une synthèse vaut quelque chose seulement si chaque fait qu'elle avance se lit dans ses preuves. Aujourd'hui, ce contrôle se fait hors du moteur, dans contaixt (`scripts/synthese.py`, `faits`, `fait_present`, `verifier`). Chaque producteur doit donc l'écrire une nouvelle fois.

Mesure de contaixt (5 octobre 2026, ai-music) : le contrôle a retiré 162 phrases sur 1 723. Le banc a donné ensuite 0 réponse fausse sur 50 questions.

## Ce que le moteur a déjà

- `verify_claims` (`pkg/cortexdb/verify_claims.go`) : il contrôle des triplets (sujet, relation, objet) contre le graphe. Il ne contrôle pas une phrase contre un texte.
- Le contrat `_grade` (`pkg/cortexdb/contract.go`) : `verified`, `self_consistent`, `asserted`, `held`, `refused`.
- `knowledge_get` lit le texte d'un document par son identifiant.

## La proposition : un outil `check_against_sources`

Règle : pour chaque phrase, l'outil rend les faits de la phrase qui ne se lisent dans aucune des sources citées. Une phrase sans fait absent tient.

L'entrée :

```json
{
  "sentences": [
    {"text": "Lyria RealTime est publié en 2025-08 par Google DeepMind.",
     "sources": ["sources/arxiv_music/2508.04651.md"]}
  ],
  "equivalents": {"rgpd": ["gdpr", "dsgvo"], "etat": ["state", "staat"]}
}
```

La sortie, pour chaque phrase : `holds` (vrai ou faux), `missing` (les faits absents), `sources_missing` (les sources que le moteur ne trouve pas).

Les règles, chacune avec son test :

1. **Les faits d'une phrase** : les nombres, les dates, et les mots avec une majuscule après la 1re lettre, un chiffre, ou une majuscule initiale et 4 lettres ou plus. Le 1er mot d'une phrase n'est pas un fait.
2. **Le pliage** : la comparaison se fait sans casse et sans accents, sur les deux côtés.
3. **Les équivalents** viennent de la requête, pas du code. Le moteur ne contient aucune liste de langue. Cette règle suit le principe du propriétaire : une liste de mots sert seulement les langues que quelqu'un a énumérées.
4. **Les sources** sont des identifiants de documents ou de morceaux (chunks). L'autre session de contaixt propose des morceaux, pour une provenance plus fine. Le moteur lit leur texte. Une source absente est dite, et la phrase ne tient pas.
5. **Le périmètre** : sous une clé confinée, une source hors de la collection de la clé est une source absente. L'outil entre dans la liste blanche avec `scope: collectionScope`.
6. **Le calcul est pur** : `(phrase, textes, équivalents) → faits absents`, testé sans base. Une seule fonction lit les textes.

Le moteur ne fait rien d'autre : il n'écrit pas, il ne retire pas de phrase et ne pose pas de grade. Le producteur décide.

## Points à soumettre au propriétaire

- **Le grade d'une synthèse.** Le pilote donne `self_consistent`. Mais la définition dit « derived deterministically from something already stated », et une synthèse écrite par un modèle n'est pas déterministe. Une synthèse contrôlée par cet outil est-elle `self_consistent` ? Ou est-elle `asserted`, avec le résultat du contrôle dans `KeyState` ?
- **La règle des faits** de contaixt est faite pour le français et l'anglais. Le propriétaire connaît le cas du chinois, où une majuscule n'existe pas. Pour une écriture CJK, la règle 1 se limite peut-être aux nombres et aux dates.

## Ce qui reste chez le producteur

- L'écriture des synthèses par un modèle, le choix des notions, la cascade.
- La décision de retirer une phrase, et la table des équivalents de son domaine.

## Avant l'envoi

- Relire `verify_claims.go` : une partie du chargement des textes (`with_text`) sert peut-être déjà.
- Rejouer le contrôle de contaixt et l'outil sur les mêmes synthèses d'ai-music : les mêmes phrases doivent tomber.
