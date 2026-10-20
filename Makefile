# Contrôles locaux avant tout commit ou push vers l'amont CortexDB. Procédure : docs/pipeline_pr.md.
# Trois niveaux (docs/pipeline_pr.md, section 1) : precommit à chaque changement (niveau 1),
# travail de tête sur têtes gelées (niveau 2), prepr une fois par lot avant la poussée (niveau 3).
# REPO : le clone (ou la copie de travail git) à contrôler. BASE : la base du patch.
REPO ?= cortexdb
BASE ?= origin/main
# DECLARED : blocs hors main incouvrables en local, déclarés dans la PR (ex. pkg/core/store_postgres.go), séparés par des espaces.
DECLARED ?=
GOBIN := $(shell go env GOPATH)/bin

.PHONY: precommit fmt vet test patchcov mutate prepr hooks

## precommit : gofmt, go vet, tests -race des paquets touchés, couverture du patch à 100 %.
precommit: fmt vet test patchcov
	@echo "precommit: OK ($(REPO) contre $(BASE))"

PKGS = $(shell cd $(REPO) && git diff --name-only $(BASE) -- '*.go' | xargs -n1 dirname 2>/dev/null | sort -u | sed 's|^|./|; s|$$|/|')

fmt:
	@cd $(REPO) && out=$$(gofmt -l pkg cmd); if [ -n "$$out" ]; then echo "gofmt:"; echo "$$out"; exit 1; fi

vet:
	@if [ -n "$(PKGS)" ]; then cd $(REPO) && go vet $(PKGS); fi

test:
	@if [ -n "$(PKGS)" ]; then cd $(REPO) && go test -race $(PKGS); else echo "test: aucun paquet touché"; fi

## patchcov : lignes changées que les tests n'exécutent pas (sortie 1 s'il en reste).
patchcov:
	@python3 -I tools/patchcov.py --repo $(REPO) --base $(BASE) $(foreach d,$(DECLARED),--declare $(d))

## mutate : mutation automatique des lignes changées (gremlins), rapport à lire
## (docs/pipeline_pr.md, section 5). Compter une exécution des tests du paquet
## par mutant (pkg/cortexdb : ~35 s chacun).
mutate:
	@cd $(REPO) && $(GOBIN)/gremlins unleash --diff $(BASE) --timeout-coefficient 3

## prepr : tout ce qui précède une PR : precommit, mutation du diff, suite
## complète -race et éval lexicale, comme en CI (environ 25 min).
prepr: precommit mutate
	@cd $(REPO) && go clean -cache && go test -race -count=1 -timeout 25m ./... && go test -count=1 ./pkg/eval -run TestLexicalRetrievalQuality
	@echo "prepr: OK ($(REPO) contre $(BASE))"

## hooks : installe le hook pre-push dans le clone (vaut aussi pour ses copies de travail).
hooks:
	@H=$$(cd $(REPO) && git rev-parse --path-format=absolute --git-common-dir)/hooks; mkdir -p "$$H"; \
	printf '%s\n' '#!/bin/sh' '# Installé par `make hooks` de granular_cortex : rien ne part sans precommit.' \
	  'exec make -C /Users/mini/Documents/granular_cortex precommit REPO="$$(git rev-parse --show-toplevel)"' > "$$H/pre-push"; \
	chmod +x "$$H/pre-push"; echo "hook installé : $$H/pre-push"
