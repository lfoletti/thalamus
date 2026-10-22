# Contrôles locaux avant tout commit ou push vers l'amont CortexDB. Procédure : docs/pipeline_pr.md.
# Trois niveaux (docs/pipeline_pr.md, section 1) : precommit à chaque changement (niveau 1),
# travail de tête sur têtes gelées (niveau 2), prepr une fois par lot avant la poussée (niveau 3).
# REPO : le clone (ou la copie de travail git) à contrôler. BASE : la base du patch.
REPO ?= cortexdb
BASE ?= origin/main
# DECLARED : blocs hors main incouvrables en local, déclarés dans la PR (ex. pkg/core/store_postgres.go), séparés par des espaces.
DECLARED ?=
# SCENARIOS : dossier des scénarios différentiels de `make diffmain` (ex. docs/pr/lot_C/scenarios).
SCENARIOS ?=
GOBIN := $(shell go env GOPATH)/bin
# CANARY : aucun test n'atteint un serveur ni le cerveau de la machine. Un test qui lit
# l'environnement au lieu de le poser tombe sur « connection refused » (revue du 21 octobre 2026, #76).
# Élargi le 21 octobre 2026 (revue adverse) à l'embedder et au seuil de rappel : un test qui
# hérite d'un embedder réel ou d'un seuil tombe aussi. `main` (cc98c6c) reste vert sous ce canari.
# OPENAI_API_KEY reste vide : une clé ouvre le test réel TestMappingFromDDLWithLLM_Live de `main`.
CANARY := CORTEXDB_REMOTE=127.0.0.1:1 CORTEXDB_GRPC_TOKEN=canary CORTEXDB_PATH=/dev/null/canary \
	CORTEXDB_EMBED_BASE_URL=http://127.0.0.1:1 OPENAI_BASE_URL=http://127.0.0.1:1 OPENAI_API_KEY= \
	CORTEXDB_LLM_BASE_URL=http://127.0.0.1:1 CORTEXDB_LLM_API_KEY=canary CORTEXDB_MEMORY_SEMANTIC_FLOOR=canary

.PHONY: pgtest precommit fmt vet test patchcov mutate clock prepr benches level3 hooks inventory invcheck diffmain handmut bodynums

## precommit : gofmt, go vet, tests -race des paquets touchés, couverture du patch à 100 %.
precommit: fmt vet test patchcov
	@echo "precommit: OK ($(REPO) contre $(BASE))"

PKGS = $(shell cd $(REPO) && git diff --name-only $(BASE) -- '*.go' | xargs -n1 dirname 2>/dev/null | sort -u | sed 's|^|./|; s|$$|/|')

fmt:
	@cd $(REPO) && out=$$(gofmt -l pkg cmd); if [ -n "$$out" ]; then echo "gofmt:"; echo "$$out"; exit 1; fi

vet:
	@if [ -n "$(PKGS)" ]; then cd $(REPO) && go vet $(PKGS); fi

test:
	@if [ -n "$(PKGS)" ]; then cd $(REPO) && $(CANARY) go test -race $(PKGS); else echo "test: aucun paquet touché"; fi

## patchcov : lignes changées que les tests n'exécutent pas (sortie 1 s'il en reste).
patchcov:
	@python3 -I tools/patchcov.py --repo $(REPO) --base $(BASE) $(foreach d,$(DECLARED),--declare $(d))

## mutate : mutation automatique des lignes changées (gremlins), rapport à lire
## (docs/pipeline_pr.md, section 5). Compter une exécution des tests du paquet
## par mutant (pkg/cortexdb : ~35 s chacun).
mutate:
	@cd $(REPO) && $(GOBIN)/gremlins unleash --diff $(BASE) --timeout-coefficient 3

## clock : les tests du patch qui lisent l'heure, rejoués -race -count=COUNT (21 octobre 2026,
## #76 : 1 échec sur 12 quand une seconde passe entre deux lectures de l'horloge).
COUNT ?= 10
clock:
	@python3 -I tools/clockcheck.py --repo $(REPO) --base $(BASE) --count $(COUNT)

## prepr : tout ce qui précède une PR : precommit, mutation du diff, rejeu des tests
## qui lisent l'heure, suite complète -race et éval lexicale, comme en CI (environ 25 min).
## invcheck d'abord : un inventaire incomplet arrête prepr avant la chaîne lourde.
prepr: invcheck precommit mutate clock
	@cd $(REPO) && go clean -cache && $(CANARY) go test -race -count=1 -timeout 25m ./... && go test -count=1 ./pkg/eval -run TestLexicalRetrievalQuality
	@echo "prepr: OK ($(REPO) contre $(BASE))"

## benches : les deux bancs de bout en bout sur le clone cortexdb/ (proto) et le gateway,
## environ 3 min. Le 21 octobre 2026 ils ont trouvé deux trous que precommit, les mutations
## et diffmain ne voyaient pas (th_pending refusé, banc sur l'ancienne API). Sortie 1 au premier FAIL.
benches:
	@out=$$(sh tests/e2e_shared_brain.sh 2>&1); rc=$$?; printf '%s\n' "$$out" | grep -E '^(FAIL|pass=)' | sed 's|^|e2e_shared_brain: |'; test $$rc -eq 0 || { echo "benches: e2e_shared_brain ROUGE"; exit 1; }
	@out=$$(sh tests/e2e_gateway.sh 2>&1); rc=$$?; printf '%s\n' "$$out" | grep -E '^(FAIL|pass=)' | sed 's|^|e2e_gateway: |'; test $$rc -eq 0 || { echo "benches: e2e_gateway ROUGE"; exit 1; }
	@echo "benches: OK"

## level3 : la porte complète d'une tête : prepr, puis les deux bancs.
level3: prepr benches
	@echo "level3: OK ($(REPO) contre $(BASE))"

## handmut : une mutation à la main, restaurée depuis une copie, jamais depuis HEAD
## (21 octobre 2026 : un `git checkout` après une mutation a retiré un correctif non commis).
## Usage : make handmut REPO=… FILE=… OLD='…' NEW='…' GOTEST='-count=1 -run X ./pkg/…/'
handmut:
	@python3 -I tools/handmut.py --repo $(REPO) --file '$(FILE)' --old '$(OLD)' --new '$(NEW)' -- $(GOTEST)

## pgtest : les tests de parité PostgreSQL (DDL des deux magasins, dont les index d'expression des PR)
## sur un serveur local. PGDSN par défaut : le serveur PostgreSQL 17 de Homebrew, base cortexdb_test
## (installé le 21 octobre 2026 ; démarrer : pg_ctl -D /opt/homebrew/var/postgresql@17 start).
PGDSN ?= postgres://$(USER)@127.0.0.1:5432/cortexdb_test?sslmode=disable
PGPKGS = ./internal/pgtest/ ./pkg/core/ ./pkg/graph/ ./pkg/agentmem/ ./pkg/graphflow/ ./pkg/cortexdb/
pgtest:
	@cd $(REPO) && $(CANARY) CORTEXDB_TEST_POSTGRES='$(PGDSN)' go test -count=1 $(PGPKGS)

## inventory : la liste complète de chaque classe (outils, lectures, écritures,
## identifiants, clés réservées, promesses du diff, entrées exportées, RPC typés), à tenir
## ligne par ligne dans failing_first.md (niveau 2). Sortie Markdown, à rediriger.
inventory:
	@python3 -I tools/inventory.py --repo $(REPO) --base $(BASE)

## invcheck : chaque élément des classes INVCLASSES a une ligne remplie dans FF (le
## failing_first.md ou l'inventaire du lot). Sortie 1 par élément manquant (21 octobre 2026 :
## la revue adverse a trouvé des entrées de recherche hors de l'inventaire, que rien ne comparait).
FF ?=
INVCLASSES ?= --class tools --class reads --class writes --class ids --class keys --class exports --class rpcs
invcheck:
	@test -n "$(FF)" || { echo "invcheck: FF=<failing_first.md ou inventaire du lot> manque"; exit 2; }
	@python3 -I tools/inventory.py --repo $(REPO) --base $(BASE) $(INVCLASSES) --check $(FF) | tail -40
	@python3 -I tools/inventory.py --repo $(REPO) --base $(BASE) $(INVCLASSES) --check $(FF) >/dev/null

## bodynums : les nombres du corps de la PR contre les mesures (docs/pipeline_pr.md,
## section 3 bis, règle 5). BODY : un ou plusieurs fichiers, <fichier>#<pr> pour une
## section de review_replies.md. PR=<n> lit Codecov sur la tête, FILL=1 remplit CODECOV_<n>.
bodynums:
	@python3 -I tools/test_bodynums.py >/dev/null 2>&1 || { echo "bodynums : tests de l'outil rouges"; exit 2; }
	@python3 -I tools/bodynums.py --repo $(REPO) --base $(BASE) $(if $(SETS),--sets $(SETS)) $(foreach b,$(BODY),--body $(b)) $(if $(PR),--pr $(PR)) $(if $(FILL),--fill)

## diffmain : joue les scénarios de SCENARIOS sur BASE et sur la tête de REPO, puis
## compare (niveau 2, environ 3 min). Sortie 1 si les deux arbres diffèrent : chaque
## ligne est voulue et nommée dans le corps de la PR, ou c'est un défaut.
diffmain:
	@test -n "$(SCENARIOS)" || { echo "diffmain: SCENARIOS=<dossier> manque"; exit 2; }
	@python3 -I tools/diffmain.py --repo $(REPO) --base $(BASE) --scenarios $(SCENARIOS)

## hooks : installe le hook pre-push dans le clone (vaut aussi pour ses copies de travail).
hooks:
	@H=$$(cd $(REPO) && git rev-parse --path-format=absolute --git-common-dir)/hooks; mkdir -p "$$H"; \
	printf '%s\n' '#!/bin/sh' '# Installé par `make hooks` de Thalamus : rien ne part sans precommit.' \
	  'exec make -C /Users/mini/Documents/granular_cortex precommit REPO="$$(git rev-parse --show-toplevel)"' > "$$H/pre-push"; \
	chmod +x "$$H/pre-push"; echo "hook installé : $$H/pre-push"
