🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 1.5 Premier projet pas à pas

La toolchain est installée et l'éditeur est prêt ([section 1.4](04-installation-outils.md)) : créons maintenant, de bout en bout, le programme de tradition — un « hello world ». L'objectif n'est pas de retenir la syntaxe (elle est détaillée au [module 2](../02-fondamentaux-langage/README.md)), mais de **voir la boucle complète** : initialiser un module, écrire du code, l'exécuter. Suivez la démonstration ; chaque étape est expliquée au fil de l'eau.

## 1. Initialiser le module

On crée un dossier pour le projet, puis on l'initialise comme module Go :

```console
$ mkdir hello && cd hello
$ go mod init github.com/exemple/hello
go: creating new go.mod: module github.com/exemple/hello
```

La commande `go mod init` crée un fichier **`go.mod`**. Son contenu est minimal :

```
module github.com/exemple/hello

go 1.25.0
```

Deux éléments s'y trouvent :

- le **chemin du module** (`github.com/exemple/hello`) : c'est l'identifiant du projet, qui sert aussi de préfixe aux imports internes (voir [section 1.3](03-ecosysteme-go.md)). On y met par convention l'adresse du dépôt où le code vivra ; pour un projet purement local, n'importe quel chemin cohérent convient.
- la directive **`go`**, qui fixe la version du langage attendue.

À noter sur cette directive : **depuis Go 1.26, `go mod init` inscrit non plus la version courante, mais la précédente (N-1)** — avec une toolchain 1.26, la ligne indique donc `go 1.25.0`. C'est intentionnel : le module reste compatible avec l'avant-dernière version de Go, sans exiger d'emblée la toute dernière. Rien n'empêche de relever ce numéro dès qu'on veut s'appuyer sur une nouveauté propre à une version plus récente — en éditant la ligne à la main (`go 1.26.0`) ou, plus proprement, avec `go get go@1.26.0`. La gestion fine de `go.mod` et du versionnage est traitée en [section 15.1](../15-deploiement-devops/01-build-versioning.md).

## 2. Écrire le programme

On crée un fichier **`main.go`** à la racine :

```go
package main

import "fmt"

func main() {
	fmt.Println("Bonjour, le monde !")
}
```

Ce court programme suffit à illustrer l'anatomie d'un exécutable Go :

- **`package main`** — un programme exécutable appartient au paquet spécial `main`.
- **`import "fmt"`** — on importe le paquet `fmt` de la bibliothèque standard (formatage et affichage).
- **`func main()`** — la fonction `main` est le point d'entrée : c'est par elle que le programme démarre.
- **`fmt.Println(…)`** — affiche une ligne sur la sortie standard.

Chacune de ces notions — paquets, imports, fonctions — est reprise en détail au [module 2](../02-fondamentaux-langage/README.md). Inutile de tout assimiler ici.

## 3. Exécuter le programme

La façon la plus directe est `go run`, qui compile puis exécute sans laisser de binaire :

```console
$ go run .
Bonjour, le monde !
```

Le `.` désigne le paquet du répertoire courant. C'est l'outil idéal pour la boucle de développement : on modifie, on relance, on voit le résultat.

## 4. Produire un binaire

Pour obtenir l'exécutable que l'on déploiera, on utilise `go build` :

```console
$ go build
$ ./hello          # sous Windows : hello.exe
Bonjour, le monde !
```

`go build` produit un **binaire statique unique**, nommé d'après le dernier élément du chemin du module (ici `hello`). C'est le fichier autonome évoqué en [section 1.1](01-quest-ce-que-go.md) : on peut le copier sur une machine compatible et l'exécuter tel quel, sans rien installer d'autre.

La **cross-compilation** est tout aussi directe : en fixant les variables d'environnement `GOOS` (système d'exploitation cible) et `GOARCH` (architecture), on produit un binaire pour une autre plateforme, sans rien installer de plus.

```console
$ GOOS=windows GOARCH=amd64 go build   # un .exe Windows, depuis Linux ou macOS
$ GOOS=linux   GOARCH=arm64 go build   # un binaire Linux ARM64 (Raspberry Pi, serveurs Graviton…)
```

Livrer un exécutable autonome pour n'importe quelle cible, sans dépendance à installer, est l'une des raisons du succès de Go pour les CLI et le cloud-native ; la distribution multiplateforme est approfondie en [section 6.4](../06-cli-outillage/04-distribution.md).

## Le même projet dans l'IDE

Les deux éditeurs enveloppent exactement ces mêmes commandes :

- **GoLand** — *File → New → Project → Go*, en sélectionnant le SDK : GoLand exécute `go mod init` et crée la structure. Une icône d'exécution apparaît dans la gouttière, à côté de `func main`, pour lancer — ou déboguer — le programme.
- **VS Code** — on crée le dossier, on lance `go mod init` dans le terminal intégré (ou via la palette de commandes), puis on ouvre `main.go` : l'extension affiche les actions « run | debug » juste au-dessus de `func main`.

Dans les deux cas, ce qui se passe en coulisse est identique à la démonstration ci-dessus.

## Et ensuite ?

Un projet Go minimal se résume donc à peu de chose : un **`go.mod`** et un **`main.go`** dans le paquet `main`. À mesure qu'il grandit, le code s'organise en paquets et en répertoires — le layout standard est présenté en [section 3.5](../03-types-interfaces/05-organisation-code.md) et détaillé en [annexe E](../annexes/layout-projet/README.md). Le langage lui-même commence au [module 2](../02-fondamentaux-langage/README.md), et l'on ajoutera des tests dès qu'il y aura une logique à vérifier ([module 13](../13-tests-qualite/README.md)).

## En résumé

Trois gestes suffisent à obtenir un programme Go qui tourne : **`go mod init`** pour créer le module, l'écriture d'un `main.go`, puis **`go run .`** pour l'exécuter — `go build` fournissant au besoin un binaire déployable. Cette boucle, identique en ligne de commande comme dans un IDE, se transpose telle quelle à des projets de toute taille. Reste une question de fond avant d'aller plus loin : **quand choisir Go**, et face à quels langages ? C'est l'objet de la [section suivante](06-positionnement-2026.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 1.6 Positionnement 2026](06-positionnement-2026.md)

⏭ [Positionnement 2026 : quand choisir Go (grille de décision)](/01-introduction-go/06-positionnement-2026.md)
