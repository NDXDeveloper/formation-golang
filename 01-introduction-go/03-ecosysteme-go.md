🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 1.3 L'écosystème Go

Go n'est pas seulement un langage : c'est un **écosystème cohérent** où le langage, une **toolchain unifiée**, un **système de dépendances** (les modules) et un **rythme de publication prévisible** ont été pensés ensemble. Cette cohérence explique une bonne part de la productivité qu'on lui prête.

Cette section dresse la carte conceptuelle de cet écosystème. La mise en place concrète (installation, IDE) et le détail des commandes arrivent en [section 1.4](04-installation-outils.md) ; le premier projet, en [section 1.5](05-premier-projet.md).

## La toolchain : un seul outil, `go`

Là où d'autres langages exigent d'assembler un compilateur, un gestionnaire de paquets, un formateur, un lanceur de tests et un outil d'analyse, Go regroupe tout cela derrière **une commande unique : `go`**. C'est l'incarnation du principe « batteries included » au niveau de l'outillage.

La commande `go` couvre notamment :

| Besoin | Ce que fournit la toolchain |
|--------|-----------------------------|
| Compiler et exécuter | `go build`, `go run` |
| Tester et mesurer | `go test` (tests, benchmarks, fuzzing) |
| Formater et analyser | `gofmt`, `go vet` |
| Gérer les dépendances | `go mod`, `go get` |
| Consulter la documentation | `go doc` |

Le détail de ces commandes est traité en [section 1.4](04-installation-outils.md) ; retenez ici qu'elles forment **un tout homogène**, identique d'une machine et d'un projet à l'autre.

Le compilateur par défaut (dit *gc*) produit des **binaires natifs** pour de nombreux systèmes et architectures. Il existe des alternatives pour des usages spécifiques — *gccgo* (adossé à GCC) ou **TinyGo** (embarqué et WebAssembly, voir [module 11](../11-interop-migration/README.md)) — mais la toolchain officielle couvre l'immense majorité des besoins.

Enfin, la **bibliothèque standard** est livrée avec la toolchain et fait partie intégrante de l'écosystème : serveur HTTP, JSON, cryptographie, tests, journalisation structurée (`log/slog`) et bien d'autres. C'est un choix de conception fort — beaucoup se fait sans dépendance externe.

## Les modules : code et dépendances

Depuis Go 1.16, les **modules** sont le mode de gestion des dépendances par défaut ; ils ont remplacé l'ancien fonctionnement fondé sur le `GOPATH`. Un **module** est un ensemble de paquets versionnés ensemble, décrit par un fichier `go.mod` à sa racine.

Un `go.mod` minimal ressemble à ceci :

```
module github.com/exemple/mon-service

go 1.26.0

require github.com/google/uuid v1.6.0
```

On y trouve trois éléments clés :

- le **chemin du module** (`module …`), qui sert aussi de préfixe aux chemins d'import ;
- la directive **`go`**, qui indique la version du langage attendue ;
- les **dépendances** (`require`), chacune épinglée à une version précise.

À côté, le fichier **`go.sum`** enregistre des **empreintes cryptographiques** de chaque dépendance : il garantit qu'on récupère exactement le même code à chaque build et détecte toute altération. `go.mod` et `go.sum` sont versionnés avec le projet ; ils sont approfondis en [section 15.1](../15-deploiement-devops/01-build-versioning.md).

**Versionnage sémantique.** Les versions suivent la forme `vMAJEUR.MINEUR.CORRECTIF` (par ex. `v1.6.0`). À partir de la version majeure 2, ce numéro apparaît aussi dans le chemin d'import (`…/v2`), ce qui permet à plusieurs versions majeures de coexister.

**Récupération et intégrité.** Par défaut, Go télécharge les dépendances via un **proxy de modules** public (`proxy.golang.org`) et vérifie leur intégrité auprès d'une **base de sommes de contrôle** (`sum.golang.org`), le tout mis en cache localement. Ce comportement est ajustable (`GOPROXY`, `GOSUMDB`, `GOPRIVATE`) pour les dépôts privés ou les environnements confinés.

**Sélection des versions.** Go emploie l'algorithme **MVS (Minimal Version Selection)** : il retient, pour chaque dépendance, la **version minimale** qui satisfait l'ensemble des contraintes. Le résultat est **déterministe et reproductible**, sans fichier de verrou séparé.

En pratique, on manipule les modules avec quelques commandes — `go mod init` (créer), `go get` (ajouter ou mettre à jour), `go mod tidy` (synchroniser). Leur première utilisation est montrée pas à pas en [section 1.5](05-premier-projet.md).

## Le cycle de release semestriel

Go suit un **rythme de publication régulier : deux versions majeures par an**, soit environ une tous les six mois. Les versions sont numérotées **Go 1.N** — la version stable de référence pour cette formation est **Go 1.26**. Il n'y a pas de « Go 2 » en rupture : c'est la branche Go 1.x qui évolue, de façon compatible.

| Aspect | En pratique |
|--------|-------------|
| Fréquence | ~2 versions majeures par an (historiquement vers février et août) |
| Numérotation | `Go 1.N` (pas de rupture « Go 2 ») |
| Support | Les **deux dernières** versions majeures reçoivent les correctifs (Go 1.26 → jusqu'à Go 1.28) |
| Compatibilité | Promesse Go 1 : le code existant continue de compiler |

Cette régularité a une vertu : elle rend les **montées de version prévisibles**. Les nouveautés arrivent par petites touches, sans grand saut déstabilisant, et la **promesse de compatibilité Go 1** fait qu'une mise à jour se passe, en règle générale, sans douleur. La gouvernance qui encadre cette promesse est détaillée en [section 18.1](../18-strategie-roadmap/01-gouvernance-compatibilite.md), et les tableaux de référence des versions figurent en [annexe H](../annexes/versions-reference/README.md).

**Gestion de la toolchain.** Les fichiers `go.mod` peuvent indiquer la version de Go requise (directive `go`) et, au besoin, une `toolchain` précise. Couplé à la variable `GOTOOLCHAIN`, cela permet à la commande `go` de **sélectionner — voire télécharger automatiquement — la bonne version de la toolchain** pour chaque projet. Concrètement, un projet épingle la version de Go dont il a besoin, et l'outillage s'aligne.

Les évolutions récentes du langage sont signalées au fil des sections (repère 🆕) et synthétisées en [section 18.2](../18-strategie-roadmap/02-roadmap.md).

## Découvrir et documenter les paquets

L'écosystème s'appuie enfin sur **`pkg.go.dev`**, le portail central où l'on **découvre des paquets** et **consulte leur documentation**, générée automatiquement à partir du code et des commentaires de documentation. Localement, la commande `go doc` donne accès à la même information hors ligne. Cette culture de la documentation intégrée fait partie de l'expérience Go au quotidien.

## En résumé

L'écosystème Go tient en quatre piliers : une **toolchain unifiée** derrière la commande `go`, un système de **modules** pour des dépendances reproductibles et vérifiées, un **cycle semestriel prévisible** adossé à une **promesse de compatibilité**, et une **infrastructure de découverte** (proxy, base de sommes, `pkg.go.dev`). Cette cohérence d'ensemble est autant un atout qu'une commodité. Place maintenant à la pratique : la [section suivante](04-installation-outils.md) installe cet outillage et fait le tour des commandes essentielles, sous GoLand comme sous VS Code.

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 1.4 Installation et outils](04-installation-outils.md)

⏭ [Installation et outils](/01-introduction-go/04-installation-outils.md)
