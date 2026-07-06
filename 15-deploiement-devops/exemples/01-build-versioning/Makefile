# ============================================================================
# Section 15.1 : Build reproductible, versioning (ldflags), go.mod / go.sum
# Description : Les recettes de build de la section, comme le ferait une CI.
#              `make build`   → build de dev, VCS estampillé automatiquement.
#              `make release` → build reproductible (même source → mêmes octets)
#                               : CGO_ENABLED=0 + -trimpath + -s -w + version -X.
#              `make cross`   → cross-compilation windows/darwin/linux sans
#                               chaîne C croisée (corollaire du binaire statique).
#              `make version` → inspecte l'info de module embarquée.
# Fichier source : 01-build-versioning.md
# ============================================================================

APP     := app
PKG     := ./cmd/app
VERSION := $(shell git describe --tags --always 2>/dev/null || echo v0.0.0)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT)

.PHONY: build release cross version clean

build:                          ## build de dev (VCS estampillé)
	go build -ldflags="$(LDFLAGS)" -o $(APP) $(PKG)

release:                        ## build reproductible et allégé (release)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	  go build -trimpath -ldflags="-s -w $(LDFLAGS)" -o $(APP) $(PKG)

cross:                          ## cross-compilation multi-plateforme
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w $(LDFLAGS)" -o $(APP).exe        $(PKG)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags="-s -w $(LDFLAGS)" -o $(APP)-darwin-arm $(PKG)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags="-s -w $(LDFLAGS)" -o $(APP)-linux-arm  $(PKG)

version: build                  ## inspecte l'info de module embarquée
	go version -m $(APP)

clean:
	rm -f $(APP) $(APP).exe $(APP)-darwin-arm $(APP)-linux-arm
