# 01. Rechargement du fichier de clés par `cortexdb-grpc`

**Statut** : en réserve, ne pas envoyer. **Taille estimée** : 150 lignes hors tests.

## Le manque

`cortexdb-grpc` (et `cortexdb-rest`) lisent le fichier de clés (`--keys-file`, `authz.Load`) une fois au démarrage. Une clé ajoutée ou retirée du fichier ne compte qu'après un redémarrage. Thalamus écrit dans ce fichier à chaque `key issue`, `key revoke`, `key sync` : chaque fois, le serveur redémarre, donc chaque session MCP ouverte tombe.

## Ce que Thalamus fait aujourd'hui

Rien de plus que documenter la limite (`gateway/README.md`, « Limites »). Le banc `tests/e2e_gateway.sh` redémarre le serveur après l'émission des clés.

## La proposition

Un `authz.Source` dans `pkg/authz`, même forme que `rights.Source` de Thalamus (`gateway/rights/source.go`) : `Open(path)` charge une fois et exige le succès ; `Current()` relit le fichier quand sa date ou sa taille change, garde l'ancien jeu de clés si le nouveau fichier ne se lit pas ou ne se parse pas, et le signale une fois par changement raté. Les deux serveurs appellent `Current().Lookup(secret)` à chaque requête au lieu d'un `KeySet` figé. Pas de signal, pas d'inotify : un `stat` par requête, ce qui coûte moins que l'appel lui-même.

Fichiers : `pkg/authz/source.go` (nouveau), `pkg/rpcserver/auth.go` et `pkg/httpapi/policy.go` (prendre un `*authz.Source` à la place du `*authz.KeySet`), `cmd/cortexdb-grpc/main.go` et `cmd/cortexdb-rest/main.go` (câblage, deux lignes chacun).

## Tests, écrits avant le code

- `TestASourceReloadsTheKeyFileWhenItChanges` (`pkg/authz`) : une clé ajoutée est connue à la lecture suivante ; une clé retirée est inconnue ; un fichier cassé garde l'ancien jeu et avertit une fois ; un fichier retiré pareil ; le même texte réécrit n'est pas reparsé.
- `TestARevokedKeyIsRefusedWithoutARestart` (`pkg/rpcserver`) : serveur en route, clé retirée du fichier, appel suivant `Unauthenticated`.
- La même chose en REST (`pkg/httpapi`), 401.
- Mutations à la main : « jamais relu », « nouveau jeu parsé mais pas servi », « panne jamais signalée », « panne signalée à chaque lecture ».

## Après la fusion

Thalamus retire la phrase sur le redémarrage de `gateway/README.md` et du banc. `key sync` devient une opération sans interruption.

## Risque pour le relecteur

Un `stat` par requête sur le chemin chaud. Mesurer et donner le chiffre dans le corps (quelques microsecondes). Si refusé : variante avec un intervalle minimal entre deux `stat` (une seconde), toujours sans goroutine.
