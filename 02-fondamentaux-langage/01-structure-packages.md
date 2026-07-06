🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.1 Structure d'un programme, packages, visibilité

Au [module 1](../01-introduction-go/README.md), on a exécuté un premier programme sans le disséquer. Voyons à présent comment le code Go s'organise réellement : chaque fichier appartient à un **package**, les packages regroupent du code apparenté, les **imports** les relient, et une règle unique — la casse de la première lettre — décide de ce qui est **visible** hors d'un package.

*(L'organisation à plus grande échelle — répertoire `internal/`, layout standard, *workspaces* — relève de la [section 3.5](../03-types-interfaces/05-organisation-code.md).)*

## L'anatomie d'un fichier Go

Tout fichier `.go` suit le même ordre : une **clause `package`**, puis les **imports**, puis les **déclarations de niveau supérieur** (`const`, `var`, `type`, `func`), dans n'importe quel ordre entre elles.

```go
// Package greeting fournit des salutations simples.
package greeting

import (
	"fmt"
	"strings"
)

const defaultName = "le monde" // non exporté (minuscule)

// Hello renvoie une salutation personnalisée.
func Hello(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}
	return fmt.Sprintf("Bonjour, %s !", name)
}
```

On y distingue trois parties :

- la **clause `package`** (obligatoire, en tête), souvent précédée d'un commentaire de documentation ;
- le **bloc d'imports**, qui déclare les packages utilisés ;
- les **déclarations** (ici une constante et une fonction).

À noter : **l'ordre des déclarations de niveau supérieur n'a pas d'importance**. On peut employer une fonction ou un type déclaré plus bas dans le fichier — ou dans un autre fichier du même package.

## Les packages

Un **package** est l'ensemble des fichiers `.go` d'un même répertoire, qui déclarent tous le même nom de package. C'est l'unité de compilation et d'encapsulation de Go.

Quelques principes :

- **Convention de nommage** : le nom du package correspond au dernier élément du chemin du répertoire ; il est court, en minuscules, sans tiret bas ni majuscule interne, et de préférence au singulier.
- **Portée partagée** : les identifiants de niveau package sont visibles dans **tous les fichiers du même package**, sans import. On répartit donc librement le code d'un package sur plusieurs fichiers.
- **Éviter le bégaiement** : le nom du package qualifie ses identifiants au point d'appel (`fmt.Println`, `http.Client`). On évite donc les répétitions du type `http.HTTPServer` — `http.Server` suffit.
- **Le package `main`** est spécial : il produit un exécutable et doit contenir la fonction `main` (voir [section 1.5](../01-introduction-go/05-premier-projet.md)). Les autres packages sont des bibliothèques.

## Les imports

On importe un package par son **chemin**, et on y fait référence par son **nom**. La forme simple s'écrit `import "fmt"` ; au-delà d'un import, on utilise un bloc, avec la convention de **séparer la bibliothèque standard des dépendances externes** :

```go
import (
	"fmt" // bibliothèque standard
	"net/http"

	"github.com/google/uuid" // dépendance externe
)
```

Trois variantes plus rares méritent d'être connues :

- **Alias** — pour lever un conflit de noms : `import crand "crypto/rand"` (utile quand deux packages `rand` cohabitent).
- **Import « blanc »** — `import _ "github.com/lib/pq"` : on n'utilise pas directement le package, mais on déclenche ses effets de bord (typiquement l'enregistrement d'un driver ; voir [section 7.2](../07-acces-donnees/02-drivers.md)).
- **Import « point »** — `import . "…"` : importe les identifiants sans qualificatif ; **déconseillé**, car il nuit à la lisibilité.

Enfin, une rigueur propre à Go : **un import inutilisé est une erreur de compilation**. C'est volontaire — le code reste débarrassé des dépendances mortes. En pratique, l'outil `goimports` (ou l'IDE) ajoute et retire les imports automatiquement.

Voici comment un programme consomme le package `greeting` défini plus haut :

```go
package main

import (
	"fmt"

	"github.com/exemple/hello/greeting"
)

func main() {
	fmt.Println(greeting.Hello("Ada"))
	// greeting.defaultName // ← erreur : identifiant non exporté, inaccessible ici
}
```

## La visibilité : majuscule = exporté

C'est **la** règle d'accès de Go, et elle tient en une phrase :

> Un identifiant dont la première lettre est **majuscule** est **exporté** (visible hors du package) ; s'il commence par une **minuscule**, il est **non exporté** (privé au package).

Elle s'applique à tout : types, fonctions, variables, constantes, mais aussi **champs de structure** et **méthodes**. Il n'existe ni `public`, ni `private`, ni `protected` — la casse *est* le contrôle d'accès. Dans l'exemple précédent, `Hello` est accessible depuis `main`, tandis que `defaultName` reste confiné à son package.

Deux précisions :

- **« Non exporté » ne veut pas dire « invisible partout »** : un identifiant en minuscule reste visible dans tout son package (tous ses fichiers), mais pas au-delà.
- La visibilité des **champs de structure** a des conséquences concrètes : les outils fondés sur la réflexion, comme l'encodage JSON, ne voient que les champs exportés (voir [section 5.3](../05-backend-http/03-json.md)).

## L'initialisation : variables de package, `init`, `main`

Avant que `main` ne s'exécute, Go initialise le programme dans un ordre précis :

1. les packages importés sont initialisés d'abord (récursivement) ;
2. puis les **variables de niveau package**, dans l'ordre de leurs dépendances ;
3. puis les fonctions **`init`** ;
4. et enfin, pour le package `main`, la fonction `main`.

La fonction `func init()` est particulière : elle s'exécute **automatiquement** à l'initialisation du package, ne prend aucun argument et ne peut être ni appelée ni référencée. Un package peut même en déclarer plusieurs (réparties sur ses fichiers).

```go
var startedAt = time.Now() // initialisée avant main

func init() {
	// mise en place ponctuelle, exécutée une fois au démarrage
}
```

Mieux vaut cependant **utiliser `init` avec parcimonie** : c'est une initialisation implicite, difficile à suivre. Une initialisation explicite (dans `main`, ou via une fonction de construction) est souvent préférable — un point repris parmi les anti-patterns de l'[annexe B](../annexes/go-idiomatique/README.md).

## En résumé

Un fichier Go se résume à une **clause `package`**, des **imports** et des **déclarations**. Les packages regroupent les fichiers d'un répertoire et forment l'unité d'encapsulation ; les imports les relient par leur chemin ; et une seule règle — **majuscule = exporté** — constitue tout le modèle de visibilité du langage. Restent à peupler ces packages : place aux **types, variables et constantes**, en [section suivante](02-types-variables.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.2 Types, variables, constantes](02-types-variables.md)

⏭ [Types de données, variables, constantes, `iota`, zéro-values](/02-fondamentaux-langage/02-types-variables.md)
