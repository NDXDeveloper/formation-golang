🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 15.2 CI/CD (GitHub Actions, GitLab CI)

La CI/CD est le point où tout ce qui précède s'exécute **automatiquement** : à chaque *push* et *pull request*, on compile, teste, analyse et scanne ; sur un *tag*, on produit des artefacts reproductibles et on publie. Les *builds* rapides et le binaire unique de Go rendent ces pipelines simples et véloces. Cette section montre leur structure sur **GitHub Actions** et **GitLab CI**, les points propres à Go (épinglage de la boîte à outils, cache, Docker pour l'intégration, matrices), et la forme d'un pipeline de *release*. Le principe directeur : un **retour rapide** (échouer tôt), reproductible et sûr — le pipeline fait autorité sur la question « ce code est-il livrable ? ».

---

## L'anatomie d'un pipeline Go

Les étapes s'enchaînent du plus rapide au plus lent, pour échouer au plus tôt :

1. **Préparation** — *checkout*, installation de Go (version épinglée), restauration des caches.
2. **Dépendances** — `go mod verify` (intégrité), et un contrôle de dérive `go mod tidy` (le diff doit être vide).
3. **Compilation** — `go build ./...` : rapide, attrape les erreurs de compilation.
4. **Analyse statique** — `go vet`, `golangci-lint run` ([§ 13.5](../13-tests-qualite/05-linters.md)), éventuellement un garde-fou `go fix -diff`.
5. **Tests** — `go test -race ./...` ([§ 13.1](../13-tests-qualite/01-tests-unitaires.md)), tests d'intégration `-tags=integration` ([§ 13.3](../13-tests-qualite/03-tests-integration.md)), corpus de graines du fuzzing.
6. **Couverture** — profil + seuil ou couverture du diff ([§ 13.6](../13-tests-qualite/06-couverture-tests-ia.md)).
7. **Sécurité** — `govulncheck ./...` ([§ 15.3](03-supply-chain.md)).
8. **Sur *tag*** — *build* de *release* reproductible ([§ 15.1](01-build-versioning.md)) et publication.

On sépare les *jobs* lents ou optionnels (intégration, benchmarks, fuzzing continu) du cœur rapide, pour que les *pull requests* obtiennent un verdict en quelques minutes.

---

## GitHub Actions

Un fichier `.github/workflows/ci.yml` décrit des *jobs* parallèles :

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read          # moindre privilège ; élevé seulement pour publier/signer

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod   # épingle la version depuis go.mod
          # cache: true (par défaut) — cache modules + build, clé sur go.sum
      - run: go mod verify
      - run: go vet ./...
      - run: go test -race -coverprofile=cover.out ./...

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v6        # requis avant l'action golangci-lint
        with:
          go-version-file: go.mod
      - uses: golangci/golangci-lint-action@v9
        with:
          version: v2.12

  vuln:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
      - run: go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Points saillants : `actions/setup-go` lit la version dans `go-version-file: go.mod` (cohérence avec le local) et **active le cache par défaut** (modules + cache de *build*, clé sur `go.sum`), ce qui rend les pipelines Go très rapides. On applique le **moindre privilège** avec `permissions: contents: read`, élevé (`packages: write`, `id-token: write`) uniquement pour les *jobs* de publication et d'attestation ([§ 15.3](03-supply-chain.md)).

Les runners `ubuntu-latest` disposent de **Docker d'emblée** : les tests d'intégration Testcontainers ([§ 13.3](../13-tests-qualite/03-tests-integration.md)) y fonctionnent sans configuration. On peut aussi déclarer des `services:` (Postgres, Redis) attachés au *job*. Enfin, une **matrice** teste plusieurs versions de Go et systèmes :

```yaml
strategy:
  matrix:
    go: ['1.25', '1.26']
    os: [ubuntu-latest, macos-latest, windows-latest]
runs-on: ${{ matrix.os }}
```

---

## GitLab CI

L'équivalent `.gitlab-ci.yml` s'appuie sur l'image officielle `golang` :

```yaml
stages: [verify, test, security]

default:
  image: golang:1.26
  variables:                       # déplacer les caches dans le workspace pour les persister
    GOMODCACHE: "$CI_PROJECT_DIR/.cache/mod"
    GOCACHE: "$CI_PROJECT_DIR/.cache/build"
  cache:
    key:
      files: [go.sum]
    paths: [.cache/]

lint:
  stage: verify
  script:
    - go mod verify
    - go vet ./...
    - go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12 run ./...

test:
  stage: test
  script:
    - go test -race -coverprofile=cover.out ./...
    - go tool cover -func=cover.out
  coverage: '/total:\s+\(statements\)\s+(\d+\.\d+)%/'   # % remonté dans l'UI de la MR

govulncheck:
  stage: security
  script:
    - go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Deux spécificités GitLab. Le cache est propre au *workspace* : il faut **relocaliser `GOMODCACHE` et `GOCACHE`** dans le répertoire du projet pour qu'ils survivent d'un *job* à l'autre. Et le mot-clé `coverage:` extrait le pourcentage total de la sortie de `go tool cover -func` (ligne `total:`) via une expression régulière, pour l'afficher sur la *merge request*. Pour Testcontainers, un exécuteur avec Docker-in-Docker (`services: docker:dind`) ou accès au *socket* Docker est nécessaire.

---

## Bonnes pratiques communes

- **Épingler la version de Go** (depuis `go.mod`, ou explicite) et laisser la directive `toolchain` piloter le téléchargement ([§ 15.1](01-build-versioning.md)) : le CI et le poste local compilent avec la même version.
- **Épingler l'outillage.** Plutôt que `@latest`, on fige linters et générateurs par la directive `tool` + `go tool` (Go 1.24) ou un `go run outil@vX.Y.Z` — pour éviter qu'une CI verte hier échoue aujourd'hui sur une nouvelle version d'outil.
- **Paralléliser et échouer vite** : *jobs* `lint`/`test`/`vuln` en parallèle, *jobs* lents (intégration, fuzzing, benchmarks) à part.
- **Durcir la chaîne** : `permissions` au minimum, actions épinglées (voire par empreinte SHA — [§ 15.3](03-supply-chain.md)).
- **Builds reproductibles en CI** ([§ 15.1](01-build-versioning.md)) : `-trimpath`, `CGO_ENABLED=0`, étiquette de version issue du *tag* Git.

---

## CD — livraison et release

Sur un *tag* `v*`, on construit les artefacts reproductibles et on publie. **GoReleaser** est le standard : une configuration produit les binaires cross-compilés, archives, `checksums.txt`, images de conteneur et la *release* GitHub/GitLab en une passe ([§ 6.4](../06-cli-outillage/04-distribution.md)).

```yaml
  release:
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest
    permissions:
      contents: write       # créer la release
    steps:
      - uses: actions/checkout@v7
        with: { fetch-depth: 0 }   # GoReleaser a besoin de l'historique des tags
      - uses: actions/setup-go@v6
        with: { go-version-file: go.mod }
      - uses: goreleaser/goreleaser-action@v7
        with: { args: release --clean }
        env: { GITHUB_TOKEN: "${{ secrets.GITHUB_TOKEN }}" }
```

Le reste de la *supply chain* de publication — image multi-*stage* poussée vers un registre ([§ 9.1](../09-conteneurs-cloud/01-docker.md)), génération d'un SBOM et **signature** des artefacts (cosign/Sigstore, OIDC sans clé) — relève de [§ 15.3](03-supply-chain.md). Le déploiement effectif (Kubernetes [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md), serverless [§ 9.3](../09-conteneurs-cloud/03-serverless.md)) consiste en général à pousser une image puis à déclencher un *rollout*.

---

## Rappels déjà établis

Plusieurs points transverses conditionnent la CI : les **tests d'intégration exigent Docker** (présent sur les runners GitHub, à fournir via dind sur GitLab — [§ 13.3](../13-tests-qualite/03-tests-integration.md)) ; les **benchmarks sont peu fiables en CI** (runners partagés, forte variance : suivre des tendances plutôt que bloquer sur un seuil — [§ 14.4](../14-performance/04-benchmarking.md)) ; le **fuzzing** ne joue que son corpus de graines dans la suite normale, le fuzzing continu étant un *job* programmé à part ([§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md)) ; et l'on peut faire **échouer le *build*** si `go fix -diff` ou `go mod tidy` modifie des fichiers ([§ 13.5](../13-tests-qualite/05-linters.md)).

---

## Côté IDE : GoLand et VS Code

Les éditeurs aident à **écrire et valider** les pipelines, mais ceux-ci s'exécutent en CI. Sous GoLand, l'édition YAML bénéficie de la validation par schéma (via greffons GitHub Actions/GitLab), et l'intégration Git affiche le statut CI sur les *commits*. Sous VS Code, les extensions GitHub Actions et GitLab Workflow offrent édition, validation de schéma et suivi des exécutions.

Dans les deux cas, la meilleure façon de reproduire un échec de CI est de **lancer localement les commandes exactes du pipeline** — d'où l'usage répandu d'un `Makefile` reflétant fidèlement les étapes de la CI.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [15.3 — Sécurité de la supply chain : `govulncheck`, SBOM](03-supply-chain.md)

⏭ [Sécurité de la supply chain : `govulncheck`, SBOM](/15-deploiement-devops/03-supply-chain.md)
