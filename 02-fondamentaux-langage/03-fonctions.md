🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.3 Fonctions

Les fonctions sont l'unité première du comportement en Go. Au-delà de la déclaration de base, quatre traits façonnent le code idiomatique : les **retours multiples** (fondement de la gestion d'erreurs), les **valeurs de retour nommées**, les paramètres **variadiques**, et les **closures** — car en Go, une fonction est une valeur comme une autre.

*(Les méthodes, qui sont des fonctions dotées d'un receveur, relèvent de la [section 3.1](../03-types-interfaces/01-structs-methodes.md) ; les fonctions génériques, de la [section 3.4](../03-types-interfaces/04-generiques.md) ; `defer`, `panic` et `recover`, de la [section 2.10](10-defer-panic-recover.md).)*

## Déclarer une fonction

La forme de base associe un nom, des paramètres et un type de retour :

```go
func add(a, b int) int {
	return a + b
}
```

Lorsque des paramètres consécutifs partagent le même type, on peut le factoriser (`a, b int`). À noter, dans l'esprit de simplicité de Go : **pas de valeurs de paramètres par défaut, ni de surcharge de fonctions**. À chaque signature son nom.

## Retours multiples

Une fonction Go peut renvoyer **plusieurs valeurs** — un trait distinctif du langage. Le cas le plus courant est le couple `(résultat, error)`, socle de la gestion d'erreurs :

```go
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division par zéro")
	}
	return a / b, nil
}

q, err := divide(10, 2)
if err != nil {
	// gérer l'erreur
}
```

Un autre motif fréquent est le couple `(valeur, ok)`, où un booléen signale le succès — par exemple lors de la lecture d'une map :

```go
v, ok := m["clé"] // ok vaut false si la clé est absente
```

L'idiome complet de gestion des erreurs (`errors.Is`/`As`, *wrapping*) est détaillé en [section 2.9](09-gestion-erreurs.md).

## Valeurs de retour nommées

On peut **nommer les valeurs de retour** dans la signature. Elles deviennent des variables locales, initialisées à leur zéro-value, et un `return` nu (« *naked return* ») les renvoie telles quelles :

```go
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return // renvoie x et y
}
```

Deux atouts : la signature **documente** ce qui est renvoyé, et une fonction différée peut **modifier** ces valeurs avant le retour (voir [section 2.10](10-defer-panic-recover.md)). En contrepartie, le `return` nu nuit à la lisibilité des fonctions longues : on le réserve aux fonctions **courtes**, ou aux cas où `defer` doit intervenir.

## Fonctions variadiques

Le dernier paramètre d'une fonction peut être **variadique**, noté `...T` : il collecte un nombre quelconque d'arguments dans un slice.

```go
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

sum(1, 2, 3) // 6

xs := []int{4, 5, 6}
sum(xs...) // « éclatement » d'un slice existant
```

C'est ainsi qu'est construite une fonction familière comme `fmt.Println`. On peut passer des arguments individuels, ou **éclater** un slice existant avec la syntaxe `slice...`.

## Les fonctions sont des valeurs

En Go, une fonction est une **valeur de première classe** : on peut l'affecter à une variable, la passer en argument, ou la renvoyer. Le type d'une fonction s'écrit comme sa signature (`func(int) int`).

```go
func applyTwice(f func(int) int, v int) int {
	return f(f(v))
}

double := func(n int) int { return n * 2 } // fonction anonyme
applyTwice(double, 3) // 12
```

Cette propriété ouvre la voie aux fonctions d'ordre supérieur (qui prennent ou renvoient d'autres fonctions) — et aux closures.

## Les closures

Une **closure** est une fonction anonyme qui **capture les variables** de son environnement. La fonction et ces variables restent liées, même après la sortie de la portée d'origine :

```go
func counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

next := counter()
next() // 1
next() // 2
```

Ici, chaque `next` conserve son propre `count`. La capture se fait **par référence à la variable**, pas par copie de sa valeur — ce qui rend les closures idéales pour les compteurs, les fonctions de rappel, les générateurs à état ou les *middlewares* HTTP ([module 5](../05-backend-http/README.md)).

Un point d'attention classique concerne la capture de la **variable de boucle** ; sa sémantique (une variable par itération depuis Go 1.22) est traitée en [section 2.5](05-boucles.md).

## En résumé

Les fonctions Go se distinguent par leurs **retours multiples** — fondement du couple `(résultat, error)` —, leurs **valeurs de retour nommées** (à utiliser avec mesure), leurs paramètres **variadiques**, et leur statut de **valeurs de première classe** qui rend possibles les **closures**. Reste à orchestrer leur exécution : place au **contrôle de flot**, en [section suivante](04-conditions.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.4 Structures conditionnelles](04-conditions.md)

⏭ [Structures conditionnelles (`if` avec initialisation, `switch`, `type switch`)](/02-fondamentaux-langage/04-conditions.md)
