🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 6.4 Distribution : binaire unique, cross-compilation, GoReleaser

C'est ici que la promesse de Go pour l'outillage devient concrète. Un outil écrit en Go se **compile en un exécutable autonome**, se **recompile pour n'importe quelle plateforme** sans chaîne d'outils tierce, et se **publie en une commande** sous forme d'archives, de sommes de contrôle, de paquets système et même d'une formule Homebrew. Là où d'autres écosystèmes imposent à l'utilisateur d'installer un runtime et des dépendances, on livre un fichier. Cette section transforme cet avantage en processus reproductible.

## Le binaire unique

`go build` produit un exécutable unique. Pour du code Go pur, il est **lié statiquement** : aucune dépendance partagée à installer côté cible. La nuance tient à cgo : dès qu'un binaire utilise cgo (directement, ou via certaines briques de la stdlib), il se lie dynamiquement à la libc du système. Pour garantir un binaire **totalement statique**, on désactive cgo :

```console
$ CGO_ENABLED=0 go build -o monoutil ./cmd/monoutil
```

C'est précisément ce qui rend possible les images de conteneur minimales `scratch` ou *distroless*, ou l'exécution sur Alpine (musl) sans glibc — sujet approfondi au [§ 9.1](../09-conteneurs-cloud/01-docker.md). Avec `CGO_ENABLED=0`, le paquet `net` bascule d'ailleurs sur son résolveur DNS en pur Go, sans appel système à la résolution du réseau. Pour un CLI portable, désactiver cgo par défaut est un choix sain (voir le [§ 11.1](../11-interop-migration/01-cgo-ffi.md) pour les cas où cgo est réellement nécessaire).

## Amincir et estampiller le binaire

Trois options de build reviennent systématiquement pour la distribution :

- **`-ldflags="-s -w"`** retire la table des symboles (`-s`) et les informations de débogage DWARF (`-w`). Le binaire est plus petit ; en contrepartie il n'est plus débogable, ce qui est acceptable pour un artefact distribué (on garde les symboles sur ses builds de débogage).
- **`-trimpath`** remplace les chemins absolus du système de fichiers par des chemins tronqués : builds reproductibles et aucune fuite de l'arborescence locale.
- **`-ldflags="-X ..."`** injecte une valeur dans une variable `string` au moment du build — l'idiome pour estampiller la version.

```console
$ go build -trimpath -ldflags="-s -w -X main.version=1.2.3" -o monoutil ./cmd/monoutil
```

La reproductibilité du build (toolchain épinglée, `go.mod` / `go.sum`, horodatage déterministe) est traitée en détail au [§ 15.1](../15-deploiement-devops/01-build-versioning.md) ; on se limite ici à ce qui sert directement la distribution d'un CLI.

### Estamper la version : deux approches

**Par `-ldflags -X`**, on remplit des variables déclarées dans le paquet `main`. C'est le complément naturel du drapeau `--version` de Cobra ([§ 6.2](02-cobra-viper.md)) :

```go
var (
	version = "dev"     // remplacé au build par -X main.version=...
	commit  = "none"    // ... -X main.commit=...
	date    = "unknown" // ... -X main.date=...
)

// puis, sur la commande racine : &cobra.Command{Version: version, ...}
```

**Par `runtime/debug`** (depuis Go 1.18), le module et les métadonnées VCS sont déjà **incorporés automatiquement** dans le binaire ; on les lit sans aucun `-ldflags` :

```go
import "runtime/debug"

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version // « v1.2.3 » via go install …@v1.2.3, sinon « (devel) »
	}
	return "dev"
}
```

`go install exemple.com/cmd/monoutil@v1.2.3` estampille ainsi la version du module, et `go version -m ./monoutil` révèle module, révision Git et options de build d'un binaire déjà compilé. Pour un outil publié via GoReleaser (ci-dessous), l'approche `-ldflags -X` reste la plus répandue car elle donne un contrôle total sur le format affiché.

## Cross-compilation

Cibler une autre plateforme se résume à deux variables d'environnement : **`GOOS`** (système) et **`GOARCH`** (architecture). Aucun compilateur croisé à installer, contrairement au C — la démonstration de base figure au [§ 1.5](../01-introduction-go/05-premier-projet.md) ; on l'industrialise ici.

```console
$ GOOS=linux   GOARCH=amd64 go build -o dist/monoutil-linux-amd64       ./cmd/monoutil
$ GOOS=darwin  GOARCH=arm64 go build -o dist/monoutil-darwin-arm64      ./cmd/monoutil
$ GOOS=windows GOARCH=amd64 go build -o dist/monoutil-windows-amd64.exe ./cmd/monoutil
$ go tool dist list          # tous les couples GOOS/GOARCH pris en charge
```

Seule réserve : la cross-compilation **avec cgo** redevient difficile (il faut un compilateur C croisé pour la cible). C'est un argument de plus pour garder `CGO_ENABLED=0` sur un CLI destiné à plusieurs plateformes. Au-delà de trois ou quatre cibles, écrire et maintenir ces commandes à la main devient fastidieux : c'est le rôle de GoReleaser.

## GoReleaser : industrialiser la release

GoReleaser traite l'ingénierie de release comme de la configuration : on décrit quoi construire, empaqueter, signer et publier, puis une seule commande — ou un seul job de CI — fait la même chose à chaque fois. Il gère la matrice de builds multi-plateformes, les archives, les sommes de contrôle, le changelog, la publication de la release (GitHub, GitLab, Gitea), les paquets, les conteneurs et l'expérience « installation en une ligne » via Homebrew ou Scoop. Malgré son nom, il sait aussi publier des projets Rust, Zig, Python ou TypeScript, mais Go reste son terrain de prédilection.

Le bon réflexe idiomatique : on n'en a pas besoin pour un outil personnel (un `go build` et deux cross-compilations suffisent). GoReleaser prend son sens dès qu'on **distribue à d'autres** — plusieurs OS et architectures, sommes de contrôle, paquets, formule Homebrew — et qu'on veut garantir que la version incorporée corresponde toujours au tag Git publié.

### Le fichier `.goreleaser.yaml`

Depuis GoReleaser v2 (série toujours active en 2026), la configuration doit porter l'en-tête `version: 2`. Voici une base réaliste :

```yaml
version: 2
project_name: monoutil

before:
  hooks: # étapes de validation en lecture seule, avant tout empaquetage
    - go mod verify
    - go test ./...

builds:
  - id: monoutil
    main: ./cmd/monoutil
    binary: monoutil
    env:
      - CGO_ENABLED=0 # cross-compilation statique, sans compilateur C
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    flags:
      - -trimpath
    ldflags:
      - -s -w
      - -X main.version={{ .Version }}
      - -X main.commit={{ .Commit }}
      - -X main.date={{ .CommitDate }}
    mod_timestamp: "{{ .CommitTimestamp }}" # horodatage déterministe (reproductibilité)

archives:
  - id: default
    formats: [tar.gz]
    format_overrides:
      - goos: windows
        formats: [zip] # archive .zip sur Windows

checksum:
  name_template: checksums.txt

changelog:
  use: git
```

Les valeurs entre `{{ }}` sont des variables de gabarit fournies par GoReleaser (`.Version` provient du tag Git, d'où l'alignement automatique entre version affichée et version publiée). `env: [CGO_ENABLED=0]` conjugué à la matrice `goos`/`goarch` produit d'un coup tous les binaires statiques.

### Essayer en local, puis publier

GoReleaser s'installe via Homebrew, un paquet système, le script officiel, ou `go install github.com/goreleaser/goreleaser/v2@latest`. Le flux type commence toujours par un essai **local**, sans rien publier :

```console
$ goreleaser init                       # génère un .goreleaser.yaml de départ
$ goreleaser check                      # valide la configuration (options dépréciées/invalides)
$ goreleaser release --snapshot --clean # build complet en local, sans publier
```

`--clean` vide d'abord le répertoire `dist/`, ce qui évite que d'anciens artefacts ne faussent le résultat ; il a remplacé l'ancien `--rm-dist`, supprimé en v2.0. Une fois satisfait, on publie en poussant un tag :

```console
$ git tag -a v1.2.3 -m "v1.2.3" && git push origin v1.2.3
$ goreleaser release --clean            # construit tout et publie la release
```

### En intégration continue

GoReleaser est conçu pour tourner en CI sur un tag. L'extrait ci-dessous suffit à publier ; le traitement complet de la CI/CD (secrets, cache, matrices) est au [§ 15.2](../15-deploiement-devops/02-cicd.md).

```yaml
# .github/workflows/release.yml (extrait)
on:
  push:
    tags: ["v*"]
permissions:
  contents: write # nécessaire pour créer la release
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 } # historique complet, requis pour le changelog
      - uses: actions/setup-go@v5
        with: { go-version: stable }
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

L'action `goreleaser-action@v6` cible par défaut `~> v2`. `fetch-depth: 0` est important : sans l'historique Git complet, le changelog serait vide.

### Au-delà : paquets, conteneurs, signatures

Le même fichier peut, section par section, produire bien davantage : des **paquets système** deb/rpm/apk (via nFPM — le champ `maintainer` y devient obligatoire), une **formule/cask Homebrew** ou un manifeste **Scoop** pour l'installation en une ligne, des **images de conteneur** (renvoi au [§ 9.1](../09-conteneurs-cloud/01-docker.md)), ainsi qu'un **SBOM** et la **signature** des artefacts. Ce dernier volet relève de la sécurité de la chaîne d'approvisionnement, détaillée au [§ 15.3](../15-deploiement-devops/03-supply-chain.md). À noter enfin la distinction entre GoReleaser OSS et la version Pro : certaines fonctionnalités avancées sont réservées à cette dernière.

## Côté IDE : GoLand et VS Code

La distribution est surtout une activité de ligne de commande et de CI, mais les deux IDE aident à **cross-compiler sans quitter l'éditeur** et à lancer GoReleaser.

**GoLand** — pour un build croisé, créez une configuration *Go Build* (*Run kind : Package*) et renseignez *Environment* avec `GOOS=linux;GOARCH=arm64;CGO_ENABLED=0`, ou lancez simplement les commandes `GOOS=… GOARCH=… go build` dans le Terminal intégré. GoReleaser s'ajoute comme *External Tool* (Settings | Tools | External Tools) pointant sur le binaire avec l'argument `release --snapshot --clean`, ou s'exécute dans le Terminal.

**VS Code** — un `tasks.json` porte les cibles de build et l'appel à GoReleaser :

```jsonc
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "build linux/arm64",
      "type": "shell",
      "command": "go build -o dist/monoutil ./cmd/monoutil",
      "options": { "env": { "GOOS": "linux", "GOARCH": "arm64", "CGO_ENABLED": "0" } }
    },
    {
      "label": "goreleaser snapshot",
      "type": "shell",
      "command": "goreleaser release --snapshot --clean"
    }
  ]
}
```

Les variables `GOOS`/`GOARCH`/`CGO_ENABLED` passées via `options.env` suffisent à produire un binaire pour une autre plateforme depuis la palette de tâches.

## En résumé

- `go build` produit un binaire autonome ; `CGO_ENABLED=0` le rend totalement statique — condition des images `scratch`/*distroless* et de l'exécution sur Alpine ([§ 9.1](../09-conteneurs-cloud/01-docker.md)).
- Pour la distribution : `-ldflags="-s -w"` (amincir), `-trimpath` (reproductibilité, pas de fuite de chemins), `-ldflags="-X ..."` (estampiller la version) ; `runtime/debug.ReadBuildInfo` et `go version -m` offrent une alternative sans ldflags.
- Cross-compilation via `GOOS`/`GOARCH` (aucune chaîne C à installer, sauf avec cgo) ; `go tool dist list` énumère les cibles.
- GoReleaser (v2, en-tête `version: 2`) industrialise la release : matrice de builds, archives, sommes de contrôle, changelog, publication. Essai local avec `release --snapshot --clean`, publication sur tag avec `release --clean` (l'ancien `--rm-dist` a disparu).
- Il n'est utile que si l'on distribue à d'autres ; paquets, conteneurs et signatures se branchent dans le même fichier, avec renvois vers le [§ 9.1](../09-conteneurs-cloud/01-docker.md) (conteneurs) et le [§ 15.3](../15-deploiement-devops/03-supply-chain.md) (chaîne d'approvisionnement).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [7 — Accès aux données](../07-acces-donnees/README.md)

⏭ [Accès aux données](/07-acces-donnees/README.md)
