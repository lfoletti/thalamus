# Évaluer la valeur des contenus d'un cerveau

Note du 24 octobre 2026, à la demande de Lauro. Elle décrit une méthode. Aucun tri n'est fait pour l'instant.

## Copie de secours et export

- **Copie de secours :** copier tout le cerveau. Le cerveau du projet pèse 5,8 Mo. Un tri trop tôt peut perdre un contenu dont la valeur apparaît plus tard.
- **Export** (vers un autre cerveau, un nouveau projet ou un cerveau partagé de Thalamus) : trier. Un export sans tri remplit le rappel de contenus périmés ou en double.
- Une copie se fait par `sqlite3 <cerveau> ".backup '<fichier>'"`, jamais par `cp`, parce que la base a un fichier WAL.

## Contenu du cerveau le 24 octobre 2026

| Type | Nombre | Caractères |
| :- | -: | -: |
| `decision` | 37 | 23 347 |
| `project` | 35 | 34 941 |
| `feedback` | 13 | 10 776 |
| `fact` | 10 | 6 769 |
| `preference` | 6 | 3 327 |
| `knowledge` | 3 | 4 091 |
| documents de synthèse | 4 | — |

## Les cinq critères, dans l'ordre

1. **Le contenu existe-t-il ailleurs ?** C'est le critère principal. Une photo d'état qui répète GitHub, git ou `CLAUDE.md` a une valeur de copie faible. Une règle de Lauro avec sa raison n'est écrite nulle part ailleurs : sa valeur est haute. Contrôle automatique : chercher les phrases clés de chaque contenu dans `docs/`, `CLAUDE.md` et l'historique git.
2. **Le contenu est-il encore vrai ?** Une photo d'état plus récente remplace la précédente. Regrouper les contenus par sujet et garder la dernière version. Les versions précédentes sont de l'historique.
3. **Le contenu sert-il hors du projet ?** Une leçon générale vaut plus qu'un fait local (par exemple `knowledge:go-coverage:coverable-code-patterns`). Ces leçons s'exportent en premier.
4. **Qui est la source ?** Une décision ou une correction de Lauro vaut plus qu'une observation de Claude. Un fait contrôlé par un test ou un banc vaut plus qu'une affirmation.
5. **Le rappel s'en sert-il ?** Test par retrait : poser une série de vraies questions du projet, avec puis sans le contenu, et comparer les réponses. Si rien ne change, le contenu est inutile ou en double. Ce test lance des recherches répétées : le garder pour les cas douteux, et prévenir les autres sessions avant.

## Trois niveaux

- **A, à exporter :** règles et décisions de Lauro avec leur raison, leçons générales, documents de synthèse.
- **B, copie seulement :** faits du moteur et pièges, que l'on peut reconstruire à un coût élevé.
- **C, sans valeur d'export :** photos d'état présentes ailleurs, versions remplacées.

## Contenus sensibles

Le niveau ne suffit pas. Un contenu sur la sécurité peut avoir une valeur haute et rester hors d'un export vers un cerveau partagé à plusieurs personnes (par exemple une faille signalée en privé à un tiers). Marquer ces contenus « privé ». Un export vers un cerveau dont Lauro est le seul lecteur peut les garder.

## Résultat attendu d'un tri

Un tableau : identifiant, niveau, raison, où le contenu existe ailleurs, marque « privé ». Rien n'est déplacé ni effacé avant l'accord de Lauro.
