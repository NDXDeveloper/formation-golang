🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 13.5 Linters : `go vet`, staticcheck, golangci-lint

Les tests vérifient ce que le code **fait** ; les linters, eux, détectent des problèmes **sans l'exécuter** : bugs probables que le compilateur accepte pourtant, motifs non idiomatiques, code mort, erreurs non vérifiées. L'écosystème Go s'organise en trois couches — `go vet` (le plancher intégré), **staticcheck** (l'analyseur de référence) et **golangci-lint** (l'agrégateur) — auxquelles s'ajoute, depuis Go 1.26, un outil de **modernisation** : `go fix`.

---

## `go vet` — le plancher intégré

`go vet` fait partie de la chaîne d'outils. Sa règle de conception est stricte : **quasiment aucun faux positif**. Il ne signale que ce qui est presque certainement une erreur — au point qu'un rapport de vet mérite quasi toujours une correction, pas une discussion.

On le lance avec `go vet ./...`, et un sous-ensemble de ses passes s'exécute **automatiquement pendant `go test`**. Il repère notamment les formats `Printf` incohérents, les fautes dans les *struct tags*, le code inaccessible, un `cancel` de contexte perdu, la capture de variable de boucle (`loopclosure`) ou la copie d'un verrou (`copylock`).

```go
name := "Ada"
fmt.Printf("bonjour %d\n", name) // vet : le verbe %d attend un entier, pas une string
```

Ses analyseurs s'enrichissent à chaque version. Récemment : en Go 1.24, `tests` (signatures de `Test`/`Fuzz`/`Example` mal formées), `printf` sur un format non constant (`fmt.Printf(s)`), `copylock` sur une variable de boucle contenant un mutex (conséquence du changement de portée de Go 1.22) ; en Go 1.25, `waitgroup` (mauvais placement de `WaitGroup.Add`) et `hostport` :

```go
addr := fmt.Sprintf("%s:%d", host, port) // vet (hostport) : préférer net.JoinHostPort(host, port)
```

`go vet` est le plancher non négociable : toujours vert, et présent en CI.

---

## staticcheck — l'analyseur de référence

[staticcheck](https://staticcheck.dev/) (`honnef.co/go/tools`) est l'analyseur statique autonome le plus complet de l'écosystème. Il regroupe des centaines de vérifications en familles :

- **SA** — vrais bugs et problèmes de performance (déréférencement nil, assertion de type impossible, mauvais usage de la stdlib). « Hors rare faux positif, tout ce que ces analyses signalent doit être corrigé. »
- **S** — simplifications ; **ST** — style ; **QF** — corrections rapides (*quick-fixes*) ; **U** — code inutilisé (`U1000`).

Il est **conscient de la version Go** (fichier par fichier, d'après la directive `go` du `go.mod`) et activement maintenu : la version **2026.1** couvre Go 1.25 et 1.26. On l'installe et on le lance ainsi, avec une configuration facultative dans `staticcheck.conf` :

```sh
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

Un exemple parlant est la détection d'API dépréciées :

```go
data, _ := ioutil.ReadFile("config.yaml") // staticcheck SA1019 : ioutil.ReadFile est déprécié → os.ReadFile
```

En pratique, on l'exécute le plus souvent **via golangci-lint**, qui l'embarque — mais l'outil autonome reste parfait pour un usage ciblé.

---

## golangci-lint — l'agrégateur

[golangci-lint](https://golangci-lint.run/) exécute de nombreux linters en une seule passe rapide (cache, parallélisme), avec une configuration et une sortie unifiées. C'est le standard de fait en CI. Il embarque `govet`, `staticcheck`, `errcheck` (erreurs non vérifiées), `ineffassign`, `unused`, `revive`, `misspell`, `gosec`, `bodyclose`, `copyloopvar`, `modernize`, `testifylint` ([§ 13.2](02-mocks-testify.md))… des dizaines d'analyseurs.

Sa **version 2** (2025) a refondu la configuration. Le fichier `.golangci.yml` commence par `version: "2"`, remplace `enable-all`/`disable-all` par `linters.default` (`standard`, `all`, `none` ou `fast`), déplace les réglages sous `linters.settings`, et sépare le formatage dans une section `formatters` (accessible via `golangci-lint fmt`). Une commande `golangci-lint migrate` convertit une config v1.

```yaml
version: "2"

linters:
  default: standard        # govet, staticcheck, errcheck, ineffassign, unused…
  enable:
    - errorlint            # renforce l'idiome d'erreurs (§2.9)
    - bodyclose
    - misspell
  settings:
    misspell:
      locale: US
  exclusions:
    presets:
      - common-false-positives
    rules:
      - path: _test\.go
        linters: [errcheck]   # tolérer les erreurs non vérifiées dans les tests

formatters:
  enable:
    - gofmt
    - goimports
```

On lance `golangci-lint run ./...` (ajouter `--fix` applique les corrections automatisables), et l'on branche en CI l'action `golangci/golangci-lint-action@v9`. Conseil : **ne pas tout activer**. Comme pour les tests, une avalanche de linters bruyants érode la confiance. On part d'un ensemble ciblé (`default: standard` + quelques ajouts) et on l'étend progressivement ; le commentaire `//nolint:<nom-du-linter>` supprime un avertissement ponctuel, à utiliser avec parcimonie.

---

## `go fix` et les *modernizers* (Go 1.26) 🆕

Historiquement, `go fix` migrait le code après de rares changements d'API cassants. **Go 1.26 l'a entièrement réécrit** sur le framework d'analyse de `go vet`, et lui a donné une nouvelle vocation : garder le code **idiomatique**.

Il embarque désormais une vingtaine de **modernizers** — des analyseurs qui réécrivent des motifs datés vers les idiomes récents, tous préservant la sémantique, vérifiés par le compilateur et conscients de la version du `go.mod` : `interface{}` → `any`, boucles manuelles de collecte de clés → `slices.Sorted(maps.Keys(m))`, helpers maison → `min`/`max` intégrés, concaténation en boucle → `strings.Builder`, suppression de la copie obsolète `x := x` (post Go 1.22)…

```go
var v any            // était : interface{}
n := max(a, b)       // était : un if/else ou un helper maison
```

Le mécanisme complémentaire est la directive **`//go:fix inline`** : un auteur de bibliothèque annote une fonction dépréciée, et `go fix` (comme gopls) **remplace tous les appels** par leur équivalent moderne — une migration d'API « en libre-service ».

```go
package ioutilx

import "os"

// ReadFile lit le fichier nommé.
//
// Deprecated: utiliser [os.ReadFile].
//go:fix inline
func ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
// après `go fix ./...`, chaque appel ioutilx.ReadFile(f) devient os.ReadFile(f)
```

Côté usage : `go fix ./...` applique les corrections, `go fix -diff ./...` en montre le patch sans rien écrire, `-json` produit une sortie machine ; les fichiers générés sont ignorés. Une même passe peut appliquer des dizaines de corrections, réconciliées par une fusion à trois branches ; comme une correction peut en débloquer une autre, il est parfois utile de relancer `go fix` jusqu'à un point fixe (deux fois suffisent généralement). Partageant l'infrastructure de vet (`cmd/vet`, extension `-vettool`), les mêmes *modernizers* sont livrés dans **gopls** : les suggestions apparaissent en direct dans l'éditeur.

En CI, on retrouve le patron du diff vide : faire échouer le *build* si `go fix -diff ./...` renvoie quelque chose rend visible la « dette de modernisation » — exactement comme le contrôle de `go generate` vu en [§ 13.2](02-mocks-testify.md) et [§ 13.3](03-tests-integration.md).

**L'angle IA** 🤖 — les *modernizers* sont nés d'un constat de fin 2024 : les assistants LLM, entraînés sur des années de code Go existant, produisent spontanément des idiomes dépassés, et refusent parfois d'employer une fonctionnalité récente même quand on le leur demande explicitement — allant jusqu'à nier son existence. `go fix` remet ce code mécaniquement au goût du jour ; et en modernisant le corpus open-source, il améliore aussi ce sur quoi les futurs modèles s'entraîneront. Ce fil relie directement cette section à [§ 13.6](06-couverture-tests-ia.md) et aux pièges de l'IA en Go ([§ 17.2](../17-developpement-ia/02-pieges-ia.md)). À terme, ces mondes convergent : l'équipe Go prévoit d'intégrer des analyseurs de staticcheck à la commande `go` à partir de la 1.27 (certains, comme `QF1012`, sont déjà dans gopls).

---

## Une politique de qualité cohérente

Les quatre outils s'empilent : `go vet` (le plancher, toujours) → staticcheck (via golangci-lint) → golangci-lint (l'agrégateur et son jeu ciblé) → `go fix` (la modernisation). En local, l'intégration éditeur (section suivante) attrape l'essentiel avant le commit ; en CI, un job `lint` (`golangci-lint run`) doublé de `go vet ./...` et, éventuellement, d'un garde-fou `go fix -diff` fait échouer le *build* sur défaut ([§ 15.2](../15-deploiement-devops/02-cicd.md)).

Certains outils voisins sont traités ailleurs, à dessein : `govulncheck` (vulnérabilités connues) en [§ 15.3](../15-deploiement-devops/03-supply-chain.md), l'analyse de sécurité `gosec` en [§ 16.1](../16-securite/01-owasp-go.md), et le catalogue d'anti-patterns idiomatiques en annexe B. Les linters d'erreurs (`errcheck`, `errorlint`) prolongent, eux, l'idiome de gestion d'erreurs de [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md).

---

## Côté IDE : GoLand et VS Code

**GoLand** exécute en continu ses **inspections intégrées** (son propre analyseur, d'une profondeur comparable à staticcheck), avec correction rapide au vol (Alt+Entrée). Il sait aussi télécharger et lancer golangci-lint (Settings → Tools → golangci-lint, ou via un *File Watcher*) et `go vet`, et propose des transformations de type *modernize*.

**VS Code** (extension Go + gopls) fait tourner `go vet` et un ensemble d'analyseurs en direct, et les *modernizers* de `go fix` y apparaissent comme corrections rapides *inline*. On active staticcheck avec `"gopls": { "ui.diagnostic.staticcheck": true }`, et l'on branche l'agrégateur via `"go.lintTool": "golangci-lint"` et `"go.lintOnSave": "package"` (les versions récentes gèrent la config v2).

Dans les deux cas, golangci-lint est le dénominateur commun ; activer staticcheck dans gopls (VS Code) rapproche l'expérience de la profondeur native de GoLand.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [13.6 — Couverture de code ; génération de tests par IA](06-couverture-tests-ia.md)

⏭ [Couverture de code ; génération de tests par IA](/13-tests-qualite/06-couverture-tests-ia.md)
