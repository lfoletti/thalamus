# Plugin thalamus (Claude Code)

Le client des cerveaux partagés derrière `th-gateway`. Il apporte :

- le serveur MCP du gateway (`.mcp.json`, transport HTTP) : outils `th_brains`, `th_recall`, `th_get`, `th_add`, `th_pending` et `th_review`, sous la clé de l'employé ;
- une commande `/th-add <cerveau> <fichier ou texte>` qui propose un document à la relecture des curateurs ;
- un hook `UserPromptSubmit` (`bin/th-recall` → `th-hook recall`) qui injecte dans chaque prompt les 3 meilleurs extraits des cerveaux lisibles, avec leur provenance, et les cerveaux sautés ;
- une compétence (`skills/thalamus`) qui dit à l'agent comment citer la provenance et quand appeler `th_recall`.

## Installation

```sh
go build -o plugin/bin/th-hook ./gateway/cmd/th-hook     # depuis la racine du projet
th-gateway login http://gateway.interne:8787             # connexion par le compte de l'organisation : écrit ~/.thalamus/env
set -a; . ~/.thalamus/env; set +a                        # TH_GATEWAY et TH_KEY dans le shell, pour le serveur MCP
claude --plugin-dir ./plugin
```

Sans fournisseur d'identité sur le gateway, exporter `TH_GATEWAY` et `TH_KEY` (clé émise par `th-gateway key issue`) à la main.

Variables : `TH_GATEWAY`, `TH_KEY` (le hook les lit aussi dans `~/.thalamus/env`, ou le fichier `TH_ENV` ; l'environnement l'emporte champ par champ), `TH_RECALL_TOPK` (3 par défaut), `TH_HOOK_BIN` (chemin du binaire, sinon `bin/th-hook` à côté du lanceur). Le serveur MCP (`.mcp.json`) ne lit que l'environnement : d'où le `set -a`.

Le hook ne bloque jamais un prompt : sans réglage, sans binaire, gateway injoignable, clé refusée ou aucun résultat, il sort en 0 sans rien injecter. Prompt vide, de moins de 3 mots, ou commande `/…` : rien n'est demandé.

Le hook de rappel du plugin CortexDB, s'il est installé, se coupe avec `cortexdb-recall --disable` : un seul rappel par prompt.
