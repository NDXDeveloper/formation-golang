🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.2 Types de données, variables, constantes, iota, zéro-values

Les packages étant en place ([section 2.1](01-structure-packages.md)), il faut maintenant les peupler de valeurs. Go est typé statiquement, avec un petit jeu de **types de base**, quelques façons concises de **déclarer des variables**, des **constantes** évaluées à la compilation (dont les puissantes constantes *non typées* et le générateur `iota`), et une notion systématique de **valeur nulle** (*zéro-value*).

Les types **composites** — slices et maps ([2.7](07-slices-maps.md)), pointeurs ([2.8](08-pointeurs.md)), structures et interfaces ([module 3](../03-types-interfaces/README.md)) — sont traités ailleurs. On se concentre ici sur les types scalaires et la mécanique des déclarations.

## Les types de base

| Catégorie | Types |
|-----------|-------|
| Booléen | `bool` |
| Entiers signés | `int`, `int8`, `int16`, `int32`, `int64` |
| Entiers non signés | `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr` |
| Flottants | `float32`, `float64` |
| Complexes | `complex64`, `complex128` |
| Chaîne | `string` |

Quelques précisions :

- `int` et `uint` ont la taille native de la plateforme (32 ou 64 bits) ; c'est le choix par défaut pour un entier.
- `byte` est un alias de `uint8`, et `rune` un alias de `int32` (utile pour le texte, voir [section 2.6](06-chaines.md)).
- `string` est **immuable** et encode de l'UTF-8 ([section 2.6](06-chaines.md)).
- Les littéraux numériques acceptent le séparateur `_` pour la lisibilité (`1_000_000`) et les préfixes `0x` (hexadécimal), `0o` (octal) et `0b` (binaire).

Point important : **Go n'effectue aucune conversion numérique implicite**. Passer d'un type à un autre est toujours explicite.

```go
var i int = 42
var f float64 = float64(i) // conversion explicite obligatoire
var u uint = uint(f)       // idem
// var x float64 = i       // ← erreur : pas de conversion implicite
```

Cette rigueur évite quantité de bugs silencieux (dépassements, pertes de précision).

## Déclarer des variables

Go propose plusieurs formes de déclaration :

```go
var a int            // déclaration seule → zéro-value (0)
var b int = 7        // type explicite + initialisation
var c = 7            // type inféré (int)
d := 7               // déclaration courte (dans une fonction uniquement)

var e, f = 1, "deux" // plusieurs variables, types mixtes
var (                // bloc de déclarations
	g bool
	h = 3.14
)
```

Deux règles à retenir :

- La **déclaration courte `:=`** n'est utilisable qu'**à l'intérieur d'une fonction** ; `var` fonctionne partout, y compris au niveau package.
- Comme pour les imports, **une variable locale déclarée mais non utilisée est une erreur de compilation**. Pour ignorer volontairement une valeur, on emploie l'**identifiant vide `_`** :

```go
valeur, _ := lireConfig() // on ignore la seconde valeur (souvent une erreur)
```

**Piège du masquage (*shadowing*).** Dans un bloc interne, `:=` crée de **nouvelles** variables, qui masquent celles du bloc englobant. Le cas classique touche `err` :

```go
var err error
if ok {
	result, err := doWork() // ← NOUVELLE err, locale au bloc if
	if err != nil {
		log.Println(err) // on croit l'avoir traitée…
	}
	use(result)
}
return err // ← toujours nil : l'err externe n'a jamais été affectée
```

Pour affecter l'`err` externe, on déclare `result` à part (par ex. `var result int`) puis on emploie `=` au lieu de `:=`. Réflexe préventif : tester l'erreur **immédiatement** après l'appel, dans le même bloc ([section 2.9](09-gestion-erreurs.md)).

## Les constantes

Une **constante** est évaluée à la compilation et ne peut porter que sur un type de base (booléen, nombre, chaîne).

```go
const maxUsers int = 100 // constante typée
const pi = 3.14159       // constante NON typée
const greeting = "salut" // non typée (chaîne)
```

La distinction **typée / non typée** est un mécanisme important. Une constante non typée n'a pas de type figé : elle possède un **type par défaut** (`int`, `float64`, `string`, `bool`, `rune`…) mais **s'adapte au contexte** là où un type concret est attendu, sans conversion explicite :

```go
var rayon float64 = 2
circonference := 2 * pi * rayon // pi employée comme float64, sans conversion
```

C'est ce qui permet d'écrire des constantes numériques souples tout en conservant l'absence de conversion *implicite* pour les variables.

## `iota` : générer des énumérations

Go n'a pas de mot-clé `enum`. À la place, le générateur **`iota`** produit des suites de constantes dans un bloc `const`. Il vaut `0` sur la première ligne du bloc et s'incrémente de `1` à chaque ligne suivante.

```go
type Weekday int

const (
	Sunday    Weekday = iota // 0
	Monday                   // 1
	Tuesday                  // 2
	Wednesday                // 3
	Thursday                 // 4
	Friday                   // 5
	Saturday                 // 6
)
```

Combiné à une expression, `iota` engendre par exemple des **drapeaux binaires** (chaque valeur étant une puissance de deux) :

```go
type Permission uint

const (
	Read    Permission = 1 << iota // 1
	Write                          // 2
	Execute                        // 4
)
```

On peut aussi **sauter des valeurs** avec `_` (par exemple ignorer le `0` initial). C'est l'idiome classique pour définir des énumérations lisibles et sûres.

## Les zéro-values

En Go, **toute variable déclarée sans initialisation reçoit la valeur nulle de son type** — il n'existe pas de variable « non initialisée » au sens de certains langages.

| Type | Zéro-value |
|------|-----------|
| Numériques (`int`, `float64`…) | `0` |
| `bool` | `false` |
| `string` | `""` (chaîne vide) |
| Pointeurs, slices, maps, channels, fonctions, interfaces | `nil` |

Un principe de conception en découle : **rendre la zéro-value utile**. De nombreux types de la bibliothèque standard sont directement exploitables sans initialisation.

```go
var mu sync.Mutex // prêt à l'emploi, sans construction
mu.Lock()
defer mu.Unlock()

var buf bytes.Buffer // tampon vide, immédiatement utilisable
buf.WriteString("ok")
```

*(Inutile de comprendre déjà ces deux types : les verrous `sync.Mutex` sont l'objet du [module 4 (§ 4.3)](../04-concurrence/03-synchronisation.md), et `bytes.Buffer` reparaîtra avec les E/S en [section 7.6](../07-acces-donnees/06-fichiers-io.md) — retenez seulement qu'ils fonctionnent sans initialisation.)*

Concevoir ses propres types pour que leur zéro-value soit exploitable est une bonne pratique idiomatique.

## Allouer avec `new` (et sa nouveauté)

La fonction intégrée `new(T)` alloue une valeur de type `T`, l'initialise à sa zéro-value et renvoie un **pointeur** vers elle (`*T`).

```go
p := new(int) // p : *int, pointe vers 0
*p = 42
```

🆕 **Depuis Go 1.26, `new` accepte aussi une expression d'initialisation.** `new(expr)` alloue une valeur, l'initialise avec le résultat de `expr`, et renvoie un pointeur vers elle :

```go
q := new(42)   // q : *int pointant vers 42
r := new("ok") // r : *string pointant vers "ok"
```

Cette forme supprime le besoin d'une fonction utilitaire lorsqu'on veut un **pointeur vers une valeur littérale** (fréquent pour les champs optionnels). Les pointeurs eux-mêmes — sémantique, usage, pièges — sont détaillés en [section 2.8](08-pointeurs.md).

## En résumé

Go combine un **petit ensemble de types de base**, des déclarations de variables **concises mais explicites** (sans conversion implicite), des **constantes** dont les non typées s'adaptent au contexte, le générateur `iota` pour les **énumérations**, et une **zéro-value** systématique qu'on cherche à rendre utile. Ces valeurs prennent tout leur sens une fois manipulées par des **fonctions** : c'est l'objet de la [section suivante](03-fonctions.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.3 Fonctions](03-fonctions.md)

⏭ [Fonctions (retours multiples, valeurs nommées, variadiques, closures)](/02-fondamentaux-langage/03-fonctions.md)
