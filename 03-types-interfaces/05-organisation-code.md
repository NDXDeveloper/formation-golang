🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 3.5 Organisation du code : packages, `internal/`, layout standard, workspaces (`go work`)

Après avoir vu *comment* concevoir des types (structs, méthodes, interfaces), reste à savoir *où* les ranger. Go a ici des opinions nettes : le **package** est l'unité d'organisation et d'encapsulation, le répertoire **`internal/`** impose des frontières que le compilateur fait respecter, le **layout** se veut minimal et grandit avec le besoin, et **`go work`** gère le développement de plusieurs modules à la fois. Cette section clôt le module en reliant la conception à sa disposition physique.

## Le package : unité d'organisation et d'encapsulation

Un package est un **répertoire** de fichiers `.go` partageant la même clause `package X`. Règle structurante : **un seul package par répertoire** (à l'exception du package de test externe `foo_test`, [§ 13.1](../13-tests-qualite/01-tests-unitaires.md)). La visibilité, elle, est **au niveau du package**, pas du fichier — tous les fichiers d'un même package voient leurs identifiants non exportés (rappel du [§ 2.1](../02-fondamentaux-langage/01-structure-packages.md)) :

```go
// fichier : store/store.go
package store

// Store donne accès aux données (champs omis pour l'exemple).
type Store struct{}

func New() *Store { return &Store{} } // exporté (majuscule)

func connect() error { return nil } // non exporté, mais visible dans tout le package store
```

Le **chemin d'import** est le chemin du répertoire depuis la racine du module ; le **nom du package** préfixe chaque appel côté client :

```go
import "github.com/acme/app/internal/store"

s := store.New() // le nom du package fait partie de chaque site d'appel
```

`package main` produit un exécutable ; tout autre nom, une bibliothèque. Quelques conventions de nommage idiomatiques, car le nom du package se lit à *chaque* appel : court, en minuscules, sans underscore ni pluriel ; on évite les fourre-tout `util`, `common`, `helpers` (anti-pattern, [annexe B](../annexes/go-idiomatique/README.md)) et le **bégaiement** — préférez `http.Server` à `http.HTTPServer`.

Enfin, une contrainte qui façonne toute l'architecture : **les cycles d'import sont interdits**. Deux packages ne peuvent se référencer mutuellement. Loin d'être une gêne, cela force à réfléchir au *sens* des dépendances — le socle des architectures propres du [module 10](../10-architecture-services/README.md).

## `internal/` : une frontière imposée par le compilateur

Le répertoire `internal/` a un statut spécial dans la toolchain : un package situé sous `.../a/internal/` **ne peut être importé que par du code enraciné sous `.../a/`**. Toute tentative depuis l'extérieur provoque une **erreur de compilation**. C'est de l'encapsulation à l'échelle d'un sous-arbre, garantie par l'outil et non par simple convention.

```text
github.com/acme/app/
├── go.mod                 # module github.com/acme/app
├── cmd/
│   └── api/
│       └── main.go        # package main
└── internal/
    ├── store/             # importable seulement sous github.com/acme/app/…
    │   └── store.go
    └── httpapi/
        └── server.go
```

Ici, un autre module qui tenterait d'importer `github.com/acme/app/internal/store` ne compilerait pas. C'est l'outil idéal pour **cacher ses packages d'implémentation** tout en les partageant librement au sein de son propre module.

## Le layout : commencer petit, grossir au besoin

Go n'impose **aucun** layout officiel. Des conventions communautaires existent — `cmd/` (un sous-répertoire par binaire), `internal/` (code privé), parfois `pkg/` pour du code de bibliothèque public :

```text
myproj/
├── go.mod
├── cmd/
│   ├── api/main.go        # un binaire
│   └── worker/main.go     # un autre binaire
├── internal/              # code privé partagé
└── README.md
```

Deux mises en garde. D'abord, `pkg/` est **controversé** : souvent superflu, il ajoute un niveau sans réel bénéfice pour la plupart des projets — le dépôt qui l'a popularisé n'est d'ailleurs pas un standard officiel. Ensuite et surtout : **commencez petit**. Un unique package à la racine suffit à un petit programme ; on introduit `cmd/`, `internal/` et des sous-packages *au fur et à mesure* que le code grossit, jamais par mimétisme. Découpez par **domaine** (ce que fait le code) plutôt que par **couche** technique (`models/`, `controllers/`… hérités d'autres écosystèmes) — nuance développée au [§ 10.2](../10-architecture-services/02-clean-architecture.md).

Pour un layout de référence complet et commenté, voyez l'[annexe E](../annexes/layout-projet/README.md) ; cette section n'en donne que les principes.

## Workspaces : développer plusieurs modules ensemble (`go work`)

Introduits en Go 1.18, les **workspaces** répondent à un besoin précis : travailler simultanément sur plusieurs modules — typiquement une bibliothèque et l'application qui la consomme — **sans** bricoler les `go.mod` avec des directives `replace`. Un fichier `go.work` à la racine liste les modules « en cours » :

```text
workspace/
├── go.work
├── app/
│   └── go.mod             # dépend de example.com/lib
└── lib/
    └── go.mod             # module example.com/lib
```

```console
$ go work init ./app ./lib   # crée go.work avec deux directives « use »
$ go work use ./autre-module # ajoute un module à l'espace de travail
```

Le fichier obtenu est concis :

```text
go 1.25.0

use (
	./app
	./lib
)
```

*(La directive `go` du `go.work` reprend la **plus haute** des versions déclarées par les modules listés — ici `1.25.0`, celle que `go mod init` écrit sous Go 1.26, cf. [§ 1.5](../01-introduction-go/05-premier-projet.md).)*

Tant que ce `go.work` est présent, `app` voit la version **locale** de `lib` — les modifications se répercutent immédiatement, sans publication ni `replace` dans le `go.mod`. À l'ancienne, une directive `replace` pointant vers un chemin local polluait le `go.mod` et risquait d'être committée par mégarde ; `go work` garde cette configuration **hors** du `go.mod`.

`go.work` est une commodité de **développement local** : il est le plus souvent exclu du versionnement (`.gitignore`), même si certains monorepos choisissent délibérément de le committer. Un fichier compagnon `go.work.sum` peut l'accompagner. (Les bases des modules et de `go.mod` / `go.sum` relèvent des [§ 1.5](../01-introduction-go/05-premier-projet.md) et [§ 15.1](../15-deploiement-devops/01-build-versioning.md).)

## Côté IDE : GoLand et VS Code

- **GoLand** : gère nativement les projets multi-modules et `go.work` ; le refactoring *Move* déplace types et fonctions entre packages en mettant à jour tous les imports ; il détecte les cycles d'import et signale les violations de `internal/`. La vue projet distingue clairement packages et répertoires.
- **VS Code + extension Go (gopls)** : prend en charge les espaces de travail `go.work` (multi-modules), organise les imports à l'enregistrement, signale les accès interdits à `internal/`, et renomme un symbole à travers les packages (*Rename*). Le déplacement vers un autre package, lui, reste manuel — gopls n'offre pas d'équivalent au *Move* de GoLand.

## En résumé

- Le **package** = un répertoire, un `package X` ; la **visibilité** est au niveau du package (majuscule = exporté). Nom court, sans fourre-tout ni bégaiement.
- Les **cycles d'import sont interdits** : cela impose une direction claire aux dépendances ([module 10](../10-architecture-services/README.md)).
- **`internal/`** est une frontière **imposée par le compilateur** : importable seulement depuis le sous-arbre parent — idéal pour cacher l'implémentation.
- **Pas de layout officiel** : commencez petit, ajoutez `cmd/`/`internal/` au besoin, méfiez-vous de `pkg/` et découpez par domaine (layout détaillé en [annexe E](../annexes/layout-projet/README.md)).
- **`go work`** développe plusieurs modules ensemble sans `replace` ; `go.work` est une config locale, en général non versionnée.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [4. Concurrence — le point fort de Go](../04-concurrence/README.md)

⏭ [Concurrence — le point fort de Go](/04-concurrence/README.md)
