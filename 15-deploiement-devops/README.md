🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 15. Déploiement et DevOps

Le déploiement de Go est d'une simplicité inhabituelle : un **binaire unique** lié statiquement, aucun *runtime* à installer sur la cible, une **cross-compilation triviale** (`GOOS`/`GOARCH`), des *builds* rapides et reproductibles, et des images de conteneur minuscules (distroless / `scratch`, [§ 9.1](../09-conteneurs-cloud/01-docker.md)). Ce module transforme le code source en un **artefact fiable, traçable et sûr** : *builds* reproductibles et versioning (métadonnées injectées par `ldflags`, contrat d'intégrité `go.mod`/`go.sum`), pipelines CI/CD (GitHub Actions, GitLab CI — qui exécutent les tests, linters et scans des modules 13-14), et sécurité de la chaîne d'approvisionnement (`govulncheck`, SBOM).

Le fil conducteur : **tout automatiser**, rendre les *builds* reproductibles et traçables, et **avancer la sécurité au plus tôt** dans le cycle (*shift left*). Fidèle à l'esprit « stdlib d'abord », on s'appuie autant que possible sur la chaîne d'outils native (`go build`, `go mod`, cross-compilation) avant de recourir à des outils tiers.

---

## 🎯 Objectifs du module

À l'issue de ce module, vous saurez :

- produire un *build* reproductible et y **embarquer version et métadonnées** (`ldflags`), et comprendre le contrat d'intégrité `go.mod` / `go.sum` ;
- assembler un **pipeline CI/CD** (GitHub Actions, GitLab CI) qui compile, teste, analyse et publie un programme Go ;
- **sécuriser la chaîne d'approvisionnement** : détecter les vulnérabilités connues avec `govulncheck`, générer un SBOM.

---

## 🗺️ Plan du module

| # | Section | En bref |
|---|---------|---------|
| **15.1** | [Build reproductible, versioning (ldflags), `go.mod` / `go.sum`](01-build-versioning.md) | *Builds* reproductibles (`-trimpath`, `CGO_ENABLED=0`), versioning par `ldflags` (`-X`, `-s -w`), contrat d'intégrité `go.mod`/`go.sum` ; `go mod init` cible N-1 (Go 1.26). |
| **15.2** | [CI/CD (GitHub Actions, GitLab CI)](02-cicd.md) | Pipelines sur GitHub Actions et GitLab CI — build, tests , linters, cache, matrices, publication d'artefacts. |
| **15.3** | [Sécurité de la supply chain : `govulncheck`, SBOM](03-supply-chain.md) | Vulnérabilités connues avec `govulncheck` (analyse d'atteignabilité), et génération d'un SBOM (CycloneDX / SPDX). |

---

## 🧰 La boîte à outils

L'essentiel du module tient dans la chaîne standard, complétée par `govulncheck` :

```sh
go build -trimpath -ldflags="-s -w -X main.version=$(git describe --tags --always)"  # build versionné et allégé
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./...   # binaire statique, cross-compilé
go mod verify                                          # intégrité des dépendances (go.sum)
govulncheck ./...                                      # vulnérabilités connues (analyse d'atteignabilité)
go test -race ./... && golangci-lint run ./...         # ce que la CI exécute (module 13)
```

Un même binaire embarque son information de modules (`go version -m ./app`), ce qui rend le versioning et les SBOM particulièrement fiables en Go.

---

## 🆕 Nouveautés Go 1.24 → 1.26

- **`go mod init` cible désormais N-1 (Go 1.26)** — un module créé avec la 1.26 déclare `go 1.25.0` (avec le `.0`), pour encourager la compatibilité avec les versions encore supportées. On ajuste au besoin avec `go get go@version`. Détaillé en [§ 15.1](01-build-versioning.md) (et [§ 1.5](../01-introduction-go/05-premier-projet.md)).
- **Dépendances d'outils dans `go.mod` (`go get -tool`, Go 1.24)** — épingler les outils (linters, générateurs) comme dépendances versionnées garantit des *builds* CI reproductibles ([§ 15.2](02-cicd.md)).

---

## 🔗 Prérequis et articulation

Ce module est **le point de convergence** des modules 13-14 : le pipeline y exécute les tests ([§ 13](../13-tests-qualite/README.md)), les linters ([§ 13.5](../13-tests-qualite/05-linters.md)), le corpus de graines du fuzzing ([§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md)), la couverture ([§ 13.6](../13-tests-qualite/06-couverture-tests-ia.md)) et — avec précaution — les benchmarks ([§ 14.4](../14-performance/04-benchmarking.md)).

Il suppose acquis :

- la [chaîne d'outils et `go build`](../01-introduction-go/04-installation-outils.md) (§1.4-1.5) ;
- les [conteneurs](../09-conteneurs-cloud/README.md) , cible du déploiement (distroless / `scratch`) ;
- la [distribution CLI](../06-cli-outillage/04-distribution.md) (§6.4) : cross-compilation et GoReleaser.

Il se prolonge et se complète avec la [sécurité applicative](../16-securite/README.md) (chap. 16 — l'autre versant de la sécurité), l'[observabilité](../12-erreurs-debogage/04-observabilite.md) (§12.4) en production, et la politique de compatibilité des versions ([§ 18.1](../18-strategie-roadmap/01-gouvernance-compatibilite.md) et annexe H).

---

## Côté IDE : GoLand et VS Code

Le pipeline s'exécute en CI ; les éditeurs aident à l'**écrire et à le valider localement**.

- **GoLand** : configurations de compilation avec `ldflags` et *build tags* ; intégration Git ; greffons Docker et Kubernetes pour construire, inspecter les images et déployer ; terminal intégré pour les commandes équivalentes à la CI.
- **VS Code** (extension Go officielle) : *tasks* de build/test/lint, `"go.buildFlags"` / `"go.buildTags"` ; extensions GitHub Actions, GitLab Workflows, Docker et Kubernetes pour éditer et suivre les pipelines.

Dans les deux cas, **la CI fait autorité** : l'éditeur sert à préparer et vérifier, pas à remplacer le pipeline.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [15.1 — Build reproductible, versioning (ldflags), `go.mod` / `go.sum`](01-build-versioning.md)

⏭ [Build reproductible, versioning (ldflags), `go.mod` / `go.sum`](/15-deploiement-devops/01-build-versioning.md)
