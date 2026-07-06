🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 1.4 Installation et outils

Place à la pratique : mettons en place un environnement de travail complet. Trois étapes se succèdent — **installer la toolchain Go** (commune à tout le monde), **choisir un éditeur**, puis **connaître les commandes de base** que tout éditeur finit par exécuter.

Conformément au parti pris de cette formation, l'éditeur est traité **en parallèle pour GoLand et pour VS Code** : les deux sont pleinement pris en charge, et rien dans la suite n'impose l'un plutôt que l'autre. La comparaison « lequel choisir ? » fait l'objet de la [section 1.6.2](06.2-goland-vs-vscode.md), et les raccourcis du quotidien de l'[annexe D](../annexes/goland-vscode/README.md).

## Installer la toolchain Go

La source de référence est **`go.dev/dl`**, qui fournit un installeur pour chaque système. Les gestionnaires de paquets fonctionnent aussi, avec une réserve : ils sont parfois en retard d'une version.

| Système | Méthode recommandée |
|---------|---------------------|
| Windows | Installeur `.msi` depuis `go.dev/dl` (ou `winget`, `scoop`, `choco`) |
| macOS | Paquet `.pkg` depuis `go.dev/dl` (ou `brew install go`) |
| Linux | Archive `.tar.gz` extraite dans `/usr/local/go`, puis ajout au `PATH` (ou le gestionnaire de la distribution) |

Une fois l'installation faite, on la vérifie :

```bash
go version   # affiche la version installée, ex. « go version go1.26.0 linux/amd64 »
go env       # affiche la configuration (GOROOT, GOPATH, GOMODCACHE, …)
```

Deux variables de configuration reviennent souvent dans cette sortie : **`GOROOT`** (le répertoire où Go lui-même est installé — géré automatiquement, on n'y touche pas) et **`GOPATH`** (par défaut `~/go`, où atterrissent les modules téléchargés et les outils installés via `go install`). À l'ère des modules, vous n'avez plus à placer votre code sous `GOPATH` : un projet vit où vous le souhaitez.

Grâce à la gestion automatique de la toolchain (voir [section 1.3](03-ecosysteme-go.md)), une version récente de Go suffit : chaque projet pourra, au besoin, récupérer la version exacte qu'il réclame.

> 💡 **Essayer sans rien installer.** Le **Go Playground** ([go.dev/play](https://go.dev/play/)) exécute et partage des extraits de Go directement dans le navigateur. Il est volontairement limité (pas d'état persistant, entrées/sorties restreintes, pas d'accès réseau arbitraire), mais il est idéal pour les tout premiers pas et pour partager un exemple reproductible.

## GoLand (JetBrains)

**GoLand** est l'IDE commercial de JetBrains entièrement dédié à Go (le même support existe via le plugin Go d'IntelliJ IDEA Ultimate). Après installation, il détecte le SDK Go du système ; il ne reste qu'à ouvrir un module pour être opérationnel.

Ses points forts correspondent à ce qu'on attend d'un IDE complet :

- **Inspections** — une analyse statique riche, en temps réel, assortie de corrections rapides (*quick-fixes*).
- **Refactorings sûrs** — renommage, extraction de fonction ou de variable, changement de signature, déplacement, *inline*… appliqués de façon cohérente à l'ensemble du projet.
- **Débogueur intégré** — fondé sur Delve, entièrement graphique (points d'arrêt, inspection des variables, vue des goroutines), sans configuration manuelle.

S'y ajoutent un lanceur de tests visuel, des configurations d'exécution, l'intégration du gestionnaire de versions, un client HTTP et des outils base de données. GoLand est un produit payant, avec des licences gratuites pour les étudiants et les projets open source.

## VS Code + extension Go

**Visual Studio Code** est un éditeur gratuit et très répandu. Son **extension Go officielle**, maintenue par l'équipe Go, s'appuie sur deux outils clés :

- **`gopls`** — le *serveur de langage* officiel : complétion, diagnostics, aller à la définition et aux usages, documentation au survol, renommage et organisation des imports. C'est le même moteur que celui utilisé par d'autres éditeurs (Neovim, Emacs, Zed…), ce qui rend l'expérience Go portable au-delà de VS Code.
- **`dlv`** (**Delve**) — le débogueur de référence pour Go, sur lequel l'extension greffe une expérience de débogage graphique.

La mise en route est simple : installer VS Code, ajouter l'extension « Go », puis lancer la commande **« Go: Install/Update Tools »** — l'extension le propose d'elle-même — pour installer `gopls`, `dlv` et les outils associés. On obtient un environnement léger, gratuit et amplement suffisant pour la majorité des tâches.

## Les commandes essentielles en ligne de commande

Quel que soit l'éditeur retenu, tout repose *in fine* sur la commande **`go`** : les IDE ne font que l'actionner pour vous. Voici les cinq commandes du quotidien :

```bash
go run .              # compile puis exécute, sans laisser de binaire
go build -o bin/app . # compile le paquet courant en un exécutable nommé
go build ./...        # compile tous les paquets du module (vérifie que tout compile)
go test ./...         # exécute les tests de tout le module
go vet ./...          # signale des constructions suspectes que le compilateur accepte
gofmt -w .            # reformate les fichiers au format canonique de Go
```

Quelques précisions utiles :

- **Le motif `./...`** désigne le paquet du répertoire courant **et tous ses sous-paquets**, récursivement. C'est la cible habituelle pour agir sur l'ensemble d'un module.
- **`go run` vs `go build`** : `go run` sert la boucle de développement (compiler-exécuter d'un trait) ; `go build` produit l'artefact que l'on déploiera.
- **`gofmt` vs `go fmt`** : `gofmt` est l'outil de formatage ; `go fmt ./...` en est un raccourci qui l'applique à des paquets. Le format n'est pas négociable — c'est l'esprit de Go (voir [section 1.2](02-histoire-philosophie.md)).
- **`go vet`** attrape des erreurs de logique que la compilation laisse passer (un mauvais format `Printf`, par exemple). Il constitue le socle de l'analyse ; des outils plus poussés (staticcheck, golangci-lint) sont présentés en [section 13.5](../13-tests-qualite/05-linters.md).

Ces commandes sont aussi celles qu'on retrouve en intégration continue et dans les environnements sans interface graphique — d'où l'intérêt de les connaître, même en travaillant au quotidien dans un IDE. Le débogage avec Delve est approfondi en [section 12.2](../12-erreurs-debogage/02-debogage-delve.md), et l'automatisation en [section 15.2](../15-deploiement-devops/02-cicd.md).

## En résumé

Un environnement Go se met en place en trois temps : **installer la toolchain** depuis `go.dev/dl`, **choisir son éditeur** — GoLand ou VS Code, tous deux couverts dans cette formation — et **maîtriser les cinq commandes** `go run`, `go build`, `go test`, `go vet` et `gofmt`, qui restent la fondation commune. Tout est prêt : la [section suivante](05-premier-projet.md) crée un premier projet de bout en bout.

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 1.5 Premier projet pas à pas](05-premier-projet.md)

⏭ [Premier projet pas à pas (`go mod init`, hello world)](/01-introduction-go/05-premier-projet.md)
