🔝 Retour au [Sommaire](/SOMMAIRE.md)

# Annexe E — Layout de projet standard commenté

Go n'impose **aucun** layout de projet. L'outil `go` ne reconnaît que quelques noms de répertoires spéciaux (voir plus bas) ; tout le reste est convention. Le réflexe idiomatique est de **commencer simple** et de ne structurer qu'au fil du besoin réel — dans le même esprit anti-sur-ingénierie que les annexes [B](../go-idiomatique/README.md) et [C](../bonnes-pratiques/README.md).

Cette annexe est le **compagnon annoté** de la [§3.5 (organisation du code)](../../03-types-interfaces/05-organisation-code.md) : des arborescences commentées pour différents types de projets, et le rôle exact de chaque répertoire.

> ⚠️ **Un « standard » qui n'en est pas un.** Le dépôt communautaire *golang-standards/project-layout* est très cité mais **n'est pas un standard officiel** ; l'équipe Go s'en est explicitement distanciée. À prendre comme une source d'idées, pas comme une norme : la plupart des projets Go idiomatiques sont bien plus plats que ce qu'il suggère.

---

## Principe : commencer simple

Un petit outil ou une bibliothèque tient très bien **à la racine du module**, sans sous-dossier. N'ajoutez de la structure que lorsqu'elle résout un problème concret (plusieurs binaires, code à cacher, spec d'API à versionner…).

```
monoutil/
├── go.mod            # module + version du langage
├── go.sum            # présent dès qu'il y a des dépendances
├── README.md
├── LICENSE
├── monoutil.go       # le package est à la racine
└── monoutil_test.go  # tests dans le même package (ou monoutil_test)
```

C'est un layout parfaitement idiomatique. Ne le complexifiez pas par anticipation.

---

## Les répertoires reconnus par l'outil `go`

Point de départ factuel : distinguer ce que la **chaîne d'outils impose** de ce qui n'est que convention.

| Répertoire | Reconnu par `go` | Effet |
|---|---|---|
| `internal/` | **Oui** | Import restreint au sous-arbre parent — imposé par le compilateur |
| `testdata/` | **Oui** | Ignoré par le build ; réservé aux données de test |
| `vendor/` | **Oui** | Dépendances vendorisées, utilisées par le build si le dossier est présent |
| `_nom/`, `.nom/` | **Oui** | Ignorés par l'outil `go` (non compilés) |
| `cmd/`, `pkg/`, `api/`, … | Non | Pures **conventions** communautaires |

Tout le reste (`cmd/`, `pkg/`, `api/`, `deployments/`…) relève d'habitudes partagées, utiles pour la lisibilité mais sans effet sur la compilation.

---

## `internal/` — la seule frontière imposée

C'est l'outil de structuration le plus important de Go. Règle exacte : **le code situé dans (ou sous) un répertoire nommé `internal/` n'est importable que par le code du sous-arbre enraciné au parent de ce `internal/`.**

```
github.com/org/app/
├── cmd/
│   └── app/
│       └── main.go        # ✅ peut importer internal/secret
└── internal/
    └── secret/            # importable UNIQUEMENT sous github.com/org/app/…
        └── secret.go
```

Un autre module (`github.com/tiers/x`) **ne peut pas** importer `github.com/org/app/internal/secret`. La règle vaut à **n'importe quel niveau** : `foo/internal/bar` n'est visible que sous `foo/`. Mettez en `internal/` tout ce qui n'est pas une API publique stable — c'est-à-dire la majorité du code applicatif.

---

## Layouts commentés par scénario

### Application CLI

```
moncli/
├── go.mod
├── go.sum
├── README.md
├── Makefile                 # cibles build/test/lint (cf. annexe C)
├── cmd/
│   └── moncli/
│       └── main.go          # point d'entrée : lit les flags, appelle internal/
├── internal/
│   ├── cli/                 # commandes (Cobra) — privé au module
│   │   ├── root.go
│   │   └── version.go
│   └── app/                 # logique métier, indépendante de la CLI
│       └── app.go
└── testdata/                # fixtures ignorées par le build
```

Un sous-dossier par binaire dans `cmd/` ; `main.go` reste **mince** (câblage), la logique vit dans `internal/`.

### Service backend / API

```
monservice/
├── go.mod                   # module github.com/org/monservice
├── go.sum
├── README.md
├── Makefile
├── .golangci.yml            # config du linter (cf. §13.5)
├── Dockerfile               # multi-stage, image distroless/scratch (cf. §9.1)
├── .dockerignore
├── .github/
│   └── workflows/
│       └── ci.yml           # fmt, vet, lint, test -race, govulncheck (cf. §15.2)
├── api/
│   └── openapi.yaml         # spec OpenAPI (ou .proto pour gRPC)
├── cmd/
│   └── server/
│       └── main.go          # config + injection des dépendances, puis démarrage
├── internal/
│   ├── user/                # un package PAR DOMAINE, pas par couche technique
│   │   ├── user.go          #   type et règles du domaine
│   │   ├── service.go       #   cas d'usage
│   │   ├── postgres.go      #   implémentation du dépôt (pgx)
│   │   └── http.go          #   handlers HTTP du domaine
│   ├── order/
│   │   └── …
│   └── platform/            # briques transverses privées
│       ├── config/          #   chargement de la configuration
│       ├── httpx/           #   middleware, helpers HTTP
│       └── postgres/        #   pool de connexions
├── migrations/              # migrations SQL (golang-migrate/goose, cf. §7.4)
│   ├── 0001_init.up.sql
│   └── 0001_init.down.sql
└── deployments/             # manifests Kubernetes, compose (cf. §9.2)
    └── k8s/
```

Ici, `internal/` est découpé **par domaine** (`user`, `order`) plutôt que par couche, avec un `platform/` pour le transverse. C'est un défaut raisonnable ; une autre école idiomatique place les *types* du domaine dans un package central et les *implémentations* dans des packages nommés par dépendance (`postgres/`, `http/`). Le choix d'architecture (hexagonale, clean, sans sur-ingénierie) est traité en [§10.2](../../10-architecture-services/02-clean-architecture.md).

### Dépôt multi-modules (workspaces `go work`)

Pour développer plusieurs modules ensemble sans publier de versions intermédiaires, un fichier `go.work` à la racine (cf. [§3.5](../../03-types-interfaces/05-organisation-code.md)).

```
monorepo/
├── go.work                  # déclare les modules du workspace
├── go.work.sum              # géré par l'outil (ne pas éditer à la main)
├── service-a/
│   ├── go.mod
│   └── cmd/server/main.go
├── service-b/
│   ├── go.mod
│   └── cmd/server/main.go
└── libcommune/
    ├── go.mod
    └── libcommune.go
```

```
// go.work
go 1.25.0

use (
    ./service-a
    ./service-b
    ./libcommune
)
```

On crée ce fichier avec `go work init` puis `go work use ./...` — il reste en général **hors du dépôt distant** (local au poste de développement).

---

## Organiser les packages : par domaine, pas par couche

- **Grouper par feature/domaine**, pas par nature technique. Un package `user` cohérent vaut mieux que des packages `models`, `controllers`, `services` éclatés (détail en [§3.5](../../03-types-interfaces/05-organisation-code.md) et [§10.2](../../10-architecture-services/02-clean-architecture.md)).
- **Pas de packages fourre-tout** (`util`, `common`, `helpers`) — voir [annexe B](../go-idiomatique/README.md).
- **API de package minimale** : n'exportez que le nécessaire ; le reste (`internal/`, minuscules) reste libre d'évoluer.
- **Éviter la sur-imbrication** : la profondeur des dossiers doit suivre un besoin réel, pas un plan théorique.

---

## Fichiers à la racine

- **`go.mod`** — nom du module et version du langage. 🆕 Sous Go 1.26, `go mod init` écrit la version **N-1**, soit `go 1.25.0` (avec le `.0`) ; cf. [§1.5](../../01-introduction-go/05-premier-projet.md).

```
module github.com/org/monservice

go 1.25.0

require (
    github.com/jackc/pgx/v5 vX.Y.Z   // versions gérées par `go mod tidy`
)
```

- **`go.sum`** — empreintes des dépendances ; géré par l'outil, jamais à la main.
- **`go.work` / `go.work.sum`** — workspace multi-module (souvent non versionnés).
- **`README.md`**, **`LICENSE`** — indispensables ; **`CHANGELOG.md`**, **`CONTRIBUTING.md`**, **`SECURITY.md`** selon le projet.
- **`Makefile`** (ou `Taskfile.yml`) — raccourcis `build`/`test`/`lint` (cf. [annexe C](../bonnes-pratiques/README.md)).
- **`.golangci.yml`** — configuration du linter (cf. [§13.5](../../13-tests-qualite/05-linters.md)).
- **`Dockerfile`**, **`.dockerignore`** — conteneurisation (cf. [§9.1](../../09-conteneurs-cloud/01-docker.md)).
- **`.github/workflows/`** — pipelines CI (cf. [§15.2](../../15-deploiement-devops/02-cicd.md)).
- **`.gitignore`** — exclure les binaires compilés, la couverture, les artefacts locaux.

### Ressources embarquées (`//go:embed`)

Pour livrer un **binaire unique** incluant des fichiers statiques (templates, migrations, petits assets), on les embarque avec `//go:embed` — les ressources vivent alors à côté du code qui les incorpore (cf. [§7.6](../../07-acces-donnees/06-fichiers-io.md)).

```go
import "embed"

//go:embed migrations/*.sql
var migrationsFS embed.FS
```

---

## Anti-patterns de structure

| ❌ Structure | ✅ À la place |
|---|---|
| Tout dans `pkg/` par réflexe | Racine simple ; `internal/` pour le privé ; `pkg/` seulement en cas de réel besoin d'exposer des libs |
| Dossier `src/` (réflexe Java/Maven) | Le code va à la racine du module ; pas de `src/` en Go |
| Découpage par couche (`models/`, `controllers/`, `services/`) | Découpage par domaine/feature (cf. [§3.5](../../03-types-interfaces/05-organisation-code.md), [§10.2](../../10-architecture-services/02-clean-architecture.md)) |
| Packages `util`, `common`, `helpers` | Packages au nom métier (cf. [annexe B](../go-idiomatique/README.md)) |
| Sur-imbrication (dossiers profonds pour 2 fichiers) | Aplatir ; la profondeur suit le besoin |
| Tout exporté, pas d'`internal/` | Mettre en `internal/` ce qui n'est pas une API publique |
| Un package géant « fourre-tout » | Des packages à responsabilité claire |

> Sur `pkg/` : c'est une convention **débattue**, non endossée par l'équipe Go ; elle ajoute souvent un niveau inutile. Ne l'introduisez que si vous exposez sciemment des bibliothèques réutilisables hors du module.

---

## Astuce IDE (des deux côtés)

- **GoLand** : la vue projet reflète les modules ; *Mark Directory as* (Excluded / Test Sources) pour guider l'indexation ; *File and Code Templates* pour scaffolder ; support natif de `go.work` (multi-module) ; une *Run Configuration* par binaire de `cmd/`.
- **VS Code** : `gopls` gère le multi-module et `go.work` (via un fichier *workspace*) ; le *file nesting* regroupe `x.go` et `x_test.go` ; des tâches (`tasks.json`) par binaire ; l'explorateur suit l'arborescence.

Raccourcis et réglages détaillés en [annexe D](../goland-vscode/README.md).

---

## Pour aller plus loin

- **Organisation du code, packages, `internal/`, workspaces** : [§3.5](../../03-types-interfaces/05-organisation-code.md).
- **Monolithe modulaire vs microservices** : [§10.1](../../10-architecture-services/01-monolithe-vs-microservices.md) · **Clean/hexagonale sans sur-ingénierie** : [§10.2](../../10-architecture-services/02-clean-architecture.md).
- **Conteneurs et déploiement** : [§9.1](../../09-conteneurs-cloud/01-docker.md), [§9.2](../../09-conteneurs-cloud/02-kubernetes.md).
- **Bonnes pratiques** (dépendances, CI, config) : [annexe C](../bonnes-pratiques/README.md) · **Raccourcis IDE** : [annexe D](../goland-vscode/README.md).

---

🔝 [Retour au sommaire](../../SOMMAIRE.md) · ⏭️ [Annexe F — Glossaire et acronymes](../glossaire/README.md)


⏭ [Glossaire et acronymes](/annexes/glossaire/README.md)
