# Plugin thalamus (Claude Code)

Le client des cerveaux partagés derrière `th-gateway`. Il apporte :

- le serveur MCP du gateway (`.mcp.json`, transport HTTP) : outils `th_brains`, `th_recall`, `th_get`, `th_add`, `th_pending` et `th_review`, sous la clé de l'employé ;
- une commande `/th-add [cerveau] <fichier ou texte>` qui propose un document à la relecture des curateurs, par défaut dans le cerveau d'écriture que le `CLAUDE.md` du projet nomme ;
- une commande `/th-recall [brains | on <cerveau> | off <cerveau> | only <cerveau>… | only --clear]` qui choisit les cerveaux du rappel automatique ; sans argument, elle montre une liste à cocher ;
- une commande `/th-brains` qui ouvre un dialogue des cerveaux, mis en page comme `/mcp` (`hooks/panel.tsx`, un module de hooks de Claude Code) : il prend le clavier, Échap le ferme, sa hauteur suit la liste ; une ligne par cerveau, le curseur `❯` sur la ligne qui a le focus, ✔ en vert s'il répond, ✘ en rouge sinon avec la raison, `on` ou `off`, et l'accès ; Entrée change `on` ou `off` ;
- un hook `UserPromptSubmit` (`bin/th-recall` → `th-hook recall`) qui injecte dans chaque prompt les passages des cerveaux lisibles (10 passages par défaut, `TH_RECALL_TOPK`, et jusqu'à 4 passages de synthèses en tête quand un cerveau a une pyramide), avec leur provenance, et les cerveaux sautés, puis une section par cerveau local ;
- une compétence (`skills/thalamus`) qui dit à l'agent comment citer la provenance et quand appeler `th_recall`.

## Installation

```sh
go build -o plugin/bin/th-hook ./gateway/cmd/th-hook     # depuis la racine du projet
th-gateway login http://gateway.interne:8787             # connexion par le compte de l'organisation : écrit ~/.thalamus/env
plugin/bin/th-hook install                               # cerveaux locaux : TH_LOCAL_RECALL dans ~/.thalamus/env
set -a; . ~/.thalamus/env; set +a                        # TH_GATEWAY et TH_KEY dans le shell, pour le serveur MCP
claude --plugin-dir ./plugin
```

Sans fournisseur d'identité sur le gateway, exporter `TH_GATEWAY` et `TH_KEY` (clé émise par `th-gateway key issue`) à la main.

Variables : `TH_GATEWAY`, `TH_KEY` (le hook les lit aussi dans `~/.thalamus/env`, ou le fichier `TH_ENV` ; l'environnement l'emporte champ par champ), `TH_RECALL_TOPK` (10 morceaux par défaut), `TH_RECALL_TIMEOUT` (le délai du rappel, gateway et cerveaux locaux ensemble, une durée Go, 1.5s par défaut ; une valeur fausse est dite sur stderr et le défaut s'applique), `TH_HOOK_BIN` (chemin du binaire, sinon `bin/th-hook` à côté du lanceur). Le serveur MCP (`.mcp.json`) ne lit que l'environnement : d'où le `set -a`.

Le hook ne bloque jamais un prompt : sans réglage, sans binaire, gateway injoignable, clé refusée ou aucun résultat, il sort en 0 sans rien injecter. Un gateway qui ne répond pas (connexion refusée, délai dépassé) est noté dans `gateway-down`, à côté du fichier d'environnement (`~/.thalamus/`, ou le dossier de `TH_ENV`) : pendant 5 minutes le hook ne le rattend pas, les cerveaux locaux restent servis, et le bloc finit par une ligne qui le dit (« Shared brains skipped: … »). Une clé refusée n'est pas notée : le gateway a répondu. Une fois connecté, le hook demande à `th_recall` de répondre 400 ms avant la fin de son délai (`deadline_ms`, 60 s au plus). Le gateway sert alors les cerveaux prêts, et un cerveau lent figure dans la ligne `Skipped:` du bloc, avec la raison `no answer within N ms`. Le gateway a répondu : il n'est pas noté dans `gateway-down`. S'il reste moins de 600 ms après la connexion, le hook n'appelle pas `th_recall`, le dit sur stderr et note le gateway dans `gateway-down` pour 5 minutes : une connexion aussi lente fait attendre chaque prompt sans résultat partagé. Prompt vide, de moins de 3 mots, ou commande `/…` : rien n'est demandé.

## Cerveaux locaux

Le hook ajoute les cerveaux CortexDB de la machine. Pour chaque cerveau local, il lance le rappel natif de CortexDB (`cortexdb-mcp --recall`) avec `CORTEXDB_PATH` fixé sur ce cerveau, en même temps que `th_recall`. Les cerveaux locaux ne passent pas par le gateway et n'entrent pas dans l'audit.

- `TH_LOCAL_RECALL` : la commande du rappel natif, par exemple `TH_LOCAL_RECALL="$HOME/.cache/cortexdb/bin/cortexdb-mcp-darwin-arm64 --recall"`. Sans elle, aucun cerveau local.
- `TH_LOCAL_BRAINS` : les cerveaux locaux, paires `nom=chemin` séparées par des virgules, un chemin peut commencer par `~/`. Par exemple `projet=.cortexdb/cortexdb.db,racine=~/.cortexdb/cortexdb.db`. Sans elle, le seul cerveau local est celui du projet (`CORTEXDB_PATH`).
- Chaque cerveau local a sa section, `[local:<nom>]`, dans l'ordre de la liste. Les sections locales tiennent dans 3 000 caractères en tout, sur les 9 000 du bloc.
- Un cerveau absent est sauté. Un cerveau en panne ou trop lent est sauté, avec son nom sur stderr. Les autres restent.

Les deux réglages se lisent aussi dans `~/.thalamus/env`. `th-hook install` les prépare : il trouve le binaire de CortexDB de la machine (`CORTEXDB_MCP_BIN`, sinon la version la plus récente dans le dossier `data` du plugin CortexDB ou dans `~/.cache/cortexdb/bin`, sinon `cortexdb-mcp` dans le `PATH`), écrit `TH_LOCAL_RECALL` dans `~/.thalamus/env` sans toucher aux autres lignes, et coupe le rappel automatique de CortexDB, puisque le hook de Thalamus appelle le même rappel. Pour le rétablir : `cortexdb-recall --enable`.

## Choisir ses cerveaux

Chaque personne choisit les cerveaux de son rappel automatique, locaux et partagés. Le choix est dans `~/.thalamus/recall` (ou le fichier `TH_RECALL_FILE`) :

```
only=team/acme,local:projet      # filtre d'entrée : seulement ces cerveaux (vide : tous)
skip=org/archive,local:racine    # filtre de sortie : jamais ces cerveaux
```

Les commandes l'écrivent : `th-hook brains` (la liste, `on` ou `off`), `th-hook brains --json` (pour chaque cerveau : accès, `on`, et s'il répond, avec la raison sinon : pas de clé, pas de fichier, gateway muet), `th-hook on <cerveau>…`, `th-hook off <cerveau>…`, `th-hook only <cerveau>…`, `th-hook only --clear`. Dans Claude Code : `/th-recall`. Sans argument, Claude montre les cerveaux dans une liste à cocher (4 par question, 16 au plus), puis écrit le choix par `only` et `on`. `/th-brains` montre la même liste dans un panneau, lue par `th-hook brains --json` et changée par `th-hook on` et `off` : les règles restent dans `th-hook`, et le panneau dessine ses réponses. Ses tests sont dans `tests/panel.test.tsx` (`claude plugin test plugin`). Un nouveau cerveau est actif d'office quand `only` est vide. Le choix vaut pour le rappel automatique seulement : l'outil `th_recall` et les droits de `rights.yaml` ne changent pas. Côté gateway, les cerveaux partagés écartés partent dans `skip_brains` de `th_recall`, et ne sont ni interrogés ni comptés dans l'audit.

On lit plusieurs cerveaux, mais on écrit dans un seul. Le cerveau d'écriture est un réglage du projet : le `CLAUDE.md` du projet le nomme, par exemple « Thalamus write brain: `team/acme` (`th_add` with `brain: "team/acme"`). » Dans un dépôt public, `CLAUDE.local.md`, que Claude Code lit aussi, garde ce nom privé, à condition que `.gitignore` l'écarte. La compétence et `/th-add` prennent ce cerveau quand la demande n'en nomme pas, et demandent sinon. L'outil `th_add` garde son argument `brain` obligatoire : une écriture nomme toujours sa cible, et l'audit aussi.
