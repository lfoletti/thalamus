# Plugin granular-cortex (Claude Code)

Le client des cerveaux partagés derrière `gc-gateway`. Il apporte :

- le serveur MCP du gateway (`.mcp.json`, transport HTTP) : outils `gc_brains` et `gc_recall`, sous la clé de l'employé ;
- un hook `UserPromptSubmit` (`bin/gc-recall` → `gc-hook recall`) qui injecte dans chaque prompt les 3 meilleurs extraits des cerveaux lisibles, avec leur provenance, et les cerveaux sautés ;
- une compétence (`skills/granular-cortex`) qui dit à l'agent comment citer la provenance et quand appeler `gc_recall`.

## Installation

```sh
go build -o plugin/bin/gc-hook ./gateway/cmd/gc-hook     # depuis la racine du projet
export GC_GATEWAY=http://gateway.interne:8787             # le gateway
export GC_KEY=<la clé émise par gc-gateway key issue>     # votre clé
claude --plugin-dir ./plugin
```

Variables : `GC_GATEWAY`, `GC_KEY`, `GC_RECALL_TOPK` (3 par défaut), `GC_HOOK_BIN` (chemin du binaire, sinon `bin/gc-hook` à côté du lanceur).

Le hook ne bloque jamais un prompt : sans réglage, sans binaire, gateway injoignable, clé refusée ou aucun résultat, il sort en 0 sans rien injecter. Prompt vide, de moins de 3 mots, ou commande `/…` : rien n'est demandé.

Le hook de rappel du plugin CortexDB, s'il est installé, se coupe avec `cortexdb-recall --disable` : un seul rappel par prompt.
