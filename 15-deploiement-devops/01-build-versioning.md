🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 15.1 Build reproductible, versioning (ldflags), `go.mod` / `go.sum`

Un artefact déployable digne de confiance présente trois propriétés : il est **reproductible** (même source → mêmes octets), **traçable** (il connaît sa propre version et son commit), et bâti sur un **graphe de dépendances vérifié**. Cette section couvre les trois — le contrat `go.mod`/`go.sum`, l'injection de métadonnées de build par `ldflags`, et la recette d'un *build* reproductible — qui, ensemble, font une compilation auditable.

---

## Le contrat `go.mod` / `go.sum`

Deux fichiers gouvernent les dépendances. **`go.mod`** déclare l'identité du module et ses exigences ; **`go.sum`** en garantit l'intégrité.

```text
module github.com/acme/app

go 1.25.0

require (
	github.com/jackc/pgx/v5 v5.7.2
	golang.org/x/sync    v0.10.0
)

require github.com/rs/zerolog v1.33.0 // indirect

tool go.uber.org/mock/mockgen
```

La directive **`go 1.25.0`** fixe la version de langage minimale requise. Depuis Go 1.26, `go mod init` **cible la version N-1** :

```sh
$ go mod init github.com/acme/app   # avec la boîte à outils Go 1.26
$ head -3 go.mod
module github.com/acme/app

go 1.25.0        # N-1 : la 1.26 écrit 1.25.0 (avec le .0) ; une pré-version écrirait 1.24.0
```

Le but est d'**encourager la compatibilité** avec les versions de Go encore supportées (les deux dernières). Pour cibler explicitement une autre version, on enchaîne avec `go get go@1.26.0`. À ne pas confondre avec la directive **`toolchain`** (Go 1.21+), distincte : `go` indique le minimum requis pour compiler, `toolchain` indique *quelle* boîte à outils le `go` utilisera — si `go.mod` exige une version plus récente que celle installée, `go` télécharge automatiquement la bonne. Ce point relève aussi de la gouvernance des versions ([§ 18.1](../18-strategie-roadmap/01-gouvernance-compatibilite.md) et annexe H).

Le reste du fichier suit des directives connues : `require` (dépendances directes et `// indirect`), `replace`, `exclude`, `retract`, `tool` (Go 1.24, pour épingler les outils — [§ 13.2](../13-tests-qualite/02-mocks-testify.md)) et `ignore` (Go 1.25, pour exclure des répertoires).

### `go.sum` et l'intégrité

`go.sum` contient les **empreintes cryptographiques** de chaque version de module du graphe (son zip et son `go.mod`) : c'est le registre d'intégrité, la couche de détection d'altération. Deux commandes l'entretiennent et le vérifient :

```sh
go mod tidy     # aligne go.mod/go.sum sur les imports réels (ajoute, retire)
go mod verify   # vérifie que le cache de modules correspond à go.sum
```

À l'ajout d'une dépendance, `go` vérifie son empreinte contre la **base de sommes de contrôle** (`sum.golang.org`), un journal transparent qui sert de racine de confiance à l'écosystème. Pour les modules privés, `GOPRIVATE` (ou `GONOSUMDB`) contourne proxy et base publique — sujet approfondi côté chaîne d'approvisionnement ([§ 15.3](03-supply-chain.md)). Enfin, la **sélection de version minimale** (MVS) choisit la plus petite version satisfaisant toutes les exigences : les *builds* sont **déterministes par défaut**, sans surprise « dernière version en date ».

---

## Versioning : embarquer l'information de build

Un binaire doit connaître sa propre version — pour `--version`, les journaux, les rapports de bug, le SBOM. Deux voies, complémentaires.

**1. `runtime/debug.ReadBuildInfo` (automatique).** La boîte à outils **estampille seule** l'information VCS quand on compile depuis un dépôt Git (révision, horodatage, état modifié) ainsi que l'information de module :

```go
import "runtime/debug"

func buildInfo() (rev string, dirty bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	return
}
```

On inspecte de l'extérieur avec `go version -m ./app` (et, depuis Go 1.25, `go version -m -json ./app` pour une sortie exploitable par les outils). L'estampillage se désactive avec `-buildvcs=false`.

**2. `-ldflags="-X ..."` (explicite).** Pour une valeur qui ne se déduit pas du VCS — typiquement une version sémantique de *release* —, on injecte des chaînes dans des variables de package au moment de l'édition de liens :

```go
package main

var (
	version = "dev" // surchargé au link : -X main.version=...
	commit  = "none"
)
```

```sh
go build -ldflags="-X main.version=$(git describe --tags --always) \
                   -X main.commit=$(git rev-parse --short HEAD)" ./cmd/app
```

`-X importpath.name=value` ne fonctionne que sur une **variable** (jamais une constante) de type **chaîne**, déjà déclarée. En pratique, on privilégie `ReadBuildInfo` pour l'info VCS/module (gratuite, sans plomberie) et `-X` pour l'étiquette de version — ce que fait GoReleaser ([§ 6.4](../06-cli-outillage/04-distribution.md)).

---

## Le build reproductible

L'objectif : mêmes source et boîte à outils → binaire **identique au bit près**. C'est ce qui rend une compilation vérifiable et une chaîne d'approvisionnement auditable. Quelques sources de non-déterminisme et leurs parades composent la recette :

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w -X main.version=v1.2.3" -o app ./cmd/app
```

- **`CGO_ENABLED=0`** produit un binaire **pur Go, lié statiquement** : aucune dépendance à la libc du système, portable et reproductible d'une machine à l'autre (indispensable pour une image `scratch`, [§ 9.1](../09-conteneurs-cloud/01-docker.md)).
- **`-trimpath`** retire les **chemins du système de fichiers** local du binaire — sans quoi le répertoire de compilation s'y retrouve, brisant la reproductibilité (et fuitant des chemins internes).
- **`-s -w`** supprime la table des symboles et les informations de débogage DWARF : binaire plus léger. Contrepartie : le débogage avec Delve en pâtit ([§ 12.2](../12-erreurs-debogage/02-debogage-delve.md)) — on réserve donc ces drapeaux aux *builds* de *release*.

Le reste découle de la chaîne d'outils : **MVS et `go.sum`** rendent le graphe de dépendances déterministe, et l'on **épingle la version de la boîte à outils** (directive `toolchain` ou étape CI). Pour un déterminisme strict au bit près, on maîtrise aussi l'estampillage VCS (au besoin `-buildvcs=false`).

---

## Cross-compilation

Corollaire du binaire statique : `GOOS` et `GOARCH` permettent de compiler pour **n'importe quelle cible depuis n'importe quel hôte**, sans chaîne de compilation croisée — une force rare de Go. `go tool dist list` énumère les couples disponibles.

```sh
GOOS=windows GOARCH=amd64 go build -o app.exe ./cmd/app
GOOS=darwin  GOARCH=arm64 go build -o app     ./cmd/app
```

`CGO_ENABLED=0` conserve cette portabilité ; avec CGO, il faudrait une chaîne C croisée. La distribution multi-plateforme complète (archives, `checksums.txt`, images multi-arch) est traitée en [§ 6.4](../06-cli-outillage/04-distribution.md).

---

## Côté IDE : GoLand et VS Code

**Drapeaux et environnement.** Sous GoLand, la configuration d'exécution accueille `-ldflags`/`-trimpath` dans les **Go tool arguments**, et `CGO_ENABLED`, `GOOS`/`GOARCH` dans les variables d'environnement (plus un champ *build tags*). Sous VS Code (extension Go), on règle `"go.buildFlags": ["-trimpath"]`, `"go.toolsEnvVars": {"CGO_ENABLED": "0"}` et `"go.buildTags"`.

**Gestion de `go.mod`.** Dans les deux cas, gopls gère le fichier (ajout/retrait, `tidy`, indications de mise à jour) ; GoLand offre un éditeur `go.mod` avec complétion de versions, VS Code des *CodeLens* de mise à jour des dépendances.

Les *builds* de *release* reproductibles restent toutefois pilotés par la **CI**, un `Makefile` ou GoReleaser plutôt que par l'éditeur, qui sert à préparer et vérifier localement.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [15.2 — CI/CD (GitHub Actions, GitLab CI)](02-cicd.md)

⏭ [CI/CD (GitHub Actions, GitLab CI)](/15-deploiement-devops/02-cicd.md)
