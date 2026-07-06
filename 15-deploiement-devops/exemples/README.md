# Exemples du chapitre 15 — Déploiement et DevOps

Un projet par section : un binaire **versionné et reproductible** (01), un projet avec ses **workflows CI** GitHub/GitLab (02), et une démonstration d'**atteignabilité govulncheck + SBOM** (03). Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tout a été **compilé, vérifié et exécuté** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26). Pour `01` : **make** (recettes de build). Aucun outil à pré-installer pour `02`/`03` : les scanners sont lancés via `go run …@latest` (govulncheck, cyclonedx-gomod), qui les télécharge à la volée.  
**Configuration** : aucune. **Git** recommandé pour `01` (l'estampillage VCS de `ReadBuildInfo` ne fonctionne que dans un dépôt Git — voir la note de la section 01).  
**Pas de Docker requis** pour ces exemples. *(En CI réelle, seuls les tests d'intégration Testcontainers — §13.3 — et certaines cibles GoReleaser ont besoin de Docker ; ce n'est pas le cas ici.)*

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-build-versioning/` | 15.1 | `01-build-versioning.md` | `ReadBuildInfo`, `-ldflags -X`, **build reproductible**, cross-compilation (Makefile) |
| `02-cicd/` | 15.2 | `02-cicd.md` | workflows **GitHub Actions** + **GitLab CI** complets, `go tool cover` |
| `03-supply-chain/` | 15.3 | `03-supply-chain.md` | **govulncheck** (atteignabilité), formats SARIF/OpenVEX, SBOM |

---

## 01-build-versioning — section 15.1 (`01-build-versioning.md`)

**Description** : un binaire qui connaît sa version par les deux voies de la section — `runtime/debug.ReadBuildInfo` (info VCS estampillée gratuitement) et `-ldflags -X` (étiquette de release injectée au link). Le `Makefile` porte les recettes : `build` (dev), `release` (reproductible, allégé), `cross` (multi-plateforme), `version` (inspection).  
**Lancer** :

```sh
make build      # build de dev, puis ./app
make release    # CGO_ENABLED=0 -trimpath -s -w : mêmes octets à chaque build
make cross      # app.exe (Windows), app-darwin-arm, app-linux-arm
make version    # go version -m ./app
```

**Sortie attendue** (`./app`, dans un dépôt Git) :

```text
version  = v0.0.0-abc1234       (ou le tag Git le plus proche)
commit   = abc1234
vcs.rev  = abc1234…             (dirty=false si l'arbre est propre)
```

> **Note VCS** : `vcs.rev` n'est renseigné que si l'on compile **depuis un dépôt Git**. Hors dépôt, il reste vide — c'est normal (l'estampillage se désactive aussi avec `-buildvcs=false`). Le build reproductible a été vérifié : deux `make release` successifs produisent le **même sha256**.

## 02-cicd — section 15.2 (`02-cicd.md`)

**Description** : un projet minimal (`calc.go` + test) accompagné de ses **workflows CI complets** — `.github/workflows/ci.yml` (jobs parallèles test/lint/vuln/release) et `.gitlab-ci.yml` (stages verify/test/security). Les workflows exécutent la chaîne du module : `go mod verify` → `go vet` → `go test -race` → `golangci-lint` → `govulncheck`.  
**Lancer les commandes CI localement** (ce que fait le pipeline) :

```sh
go mod verify
go vet ./...
go test -race -coverprofile=cover.out ./...
go tool cover -func=cover.out          # ligne total: → % de couverture
```

**Sortie attendue** :

```text
ok   github.com/exemple/cicd   coverage: 100.0% of statements
total:               (statements)   100.0%
```

Les deux workflows sont **valides** (YAML). Versions d'actions à jour : `actions/checkout@v7`, `actions/setup-go@v6`, `golangci/golangci-lint-action@v9`, `goreleaser/goreleaser-action@v7`. Le champ `coverage:` de GitLab extrait le pourcentage via l'expression régulière `/total:\s+\(statements\)\s+(\d+\.\d+)%/`.

## 03-supply-chain — section 15.3 (`03-supply-chain.md`)

**Description** : une démonstration concrète de l'**atteignabilité**. Le programme **appelle** `language.ParseAcceptLanguage` de `golang.org/x/text@v0.3.7`, dépendance **volontairement vulnérable** (GO-2022-1059). Le symbole fautif étant réellement atteignable, `govulncheck` le signale avec la trace d'appel.

> ⚠️ La dépendance vulnérable est **intentionnelle**, pour la démonstration. Ne jamais figer une version vulnérable en production (`go get golang.org/x/text@latest` corrigerait).

**Lancer** :

```sh
go mod tidy
govulncheck ./...                                # ou : go run golang.org/x/vuln/cmd/govulncheck@latest ./...
govulncheck -format sarif ./...  > vuln.sarif    # SARIF 2.1.0 (GitHub Code Scanning)
govulncheck -format openvex ./... > vuln.vex     # OpenVEX
go build -o app . && govulncheck -mode=binary ./app   # scan d'un binaire compilé
# SBOM CycloneDX (inventaire des composants) :
go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest mod -json -output sbom.json
```

**Sortie attendue** (`govulncheck ./...`) :

```text
Vulnerability #1: GO-2022-1059
    Found in: golang.org/x/text@v0.3.7
      #1: main.go:27:46: supplychain.main calls language.ParseAcceptLanguage
```

*Preuve de l'atteignabilité* : retirez l'appel à `ParseAcceptLanguage` (et l'import), relancez `govulncheck` → la vulnérabilité **n'est plus rapportée**, alors que la dépendance est toujours là. C'est ce qui distingue govulncheck d'un scanner naïf.

---

## Nettoyage des binaires et résidus

`go run` / `go test` ne laissent aucun binaire. Après les Makefile/commandes : `make clean` (01), et supprimer les artefacts de `03` :

```sh
rm -f app cover.out vuln.sarif vuln.vex sbom.json
go clean ./...
```

Aucun conteneur n'est lancé par ces exemples ; il n'y a donc **rien à supprimer côté Docker** (ni image, ni volume).

---

*Tous les exemples testés le 2026-07-06 (toolchain go1.26.0, Linux amd64) ; `01` dans un dépôt Git temporaire pour l'estampillage VCS. Sorties conformes au chapitre.*
