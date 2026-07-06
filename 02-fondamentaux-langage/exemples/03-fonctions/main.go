/* ============================================================================
   Section 2.3 : Fonctions
   Description : Retours multiples (résultat, error), valeurs de retour nommées
                 et return nu, paramètres variadiques et éclatement d'un slice,
                 fonctions valeurs et closures
   Fichier source : 03-fonctions.md
   ============================================================================ */

package main

import (
	"errors"
	"fmt"
)

// Déclaration de base : les paramètres consécutifs de même type se factorisent
// (a, b int). Pas de valeur par défaut ni de surcharge en Go.
func add(a, b int) int {
	return a + b
}

// LE motif Go : renvoyer (résultat, error). L'appelant DOIT regarder err.
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division par zéro") // échec : zéro-value + erreur
	}
	return a / b, nil // succès : résultat + nil
}

// Valeurs de retour NOMMÉES (x, y) : déclarées comme des variables locales,
// initialisées à zéro ; le « return » nu les renvoie telles quelles.
// À réserver aux fonctions courtes — sinon on perd le fil de ce qui est renvoyé.
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return // renvoie x et y (return « nu »)
}

// Paramètre VARIADIQUE : ...int collecte tous les arguments dans un slice.
// C'est ainsi qu'est déclarée fmt.Println(a ...any).
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Une fonction est une VALEUR : ici, f est un paramètre de type func(int) int.
func applyTwice(f func(int) int, v int) int {
	return f(f(v)) // on appelle la fonction reçue, deux fois
}

// counter renvoie une CLOSURE : la fonction anonyme retournée capture la
// variable count et la garde vivante entre les appels.
func counter() func() int {
	count := 0 // capturée par la closure ci-dessous
	return func() int {
		count++ // chaque appel modifie LA MÊME variable capturée
		return count
	}
}

func main() {
	fmt.Println("=== Déclaration de base ===")
	fmt.Println("add(2, 3) =", add(2, 3))

	fmt.Println("=== Retours multiples (résultat, error) ===")
	q, err := divide(10, 2)
	if err != nil { // le réflexe : tester l'erreur juste après l'appel
		fmt.Println("erreur :", err)
	}
	fmt.Println("divide(10, 2) =", q)
	if _, err := divide(1, 0); err != nil { // err confinée au bloc if (§ 2.4)
		fmt.Println("divide(1, 0) → erreur :", err)
	}

	fmt.Println("=== Le motif (valeur, ok) ===")
	m := map[string]int{"clé": 9}
	v, present := m["clé"] // le booléen distingue « absent » de « zéro »
	fmt.Println("v =", v, "· présent =", present)

	fmt.Println("=== Valeurs de retour nommées, return nu ===")
	x, y := split(17) // 17*4/9 = 7 (division entière), puis 17-7 = 10
	fmt.Println("split(17) →", x, y)

	fmt.Println("=== Variadiques ===")
	fmt.Println("sum(1, 2, 3) =", sum(1, 2, 3)) // arguments individuels
	xs := []int{4, 5, 6}
	fmt.Println("sum(xs...) =", sum(xs...)) // « éclatement » d'un slice existant

	fmt.Println("=== Fonctions valeurs et closures ===")
	double := func(n int) int { return n * 2 }                    // fonction anonyme affectée à une variable
	fmt.Println("applyTwice(double, 3) =", applyTwice(double, 3)) // double(double(3)) = 12
	next := counter()
	fmt.Println("next() =", next(), "puis", next()) // 1 puis 2 : count survit entre les appels
}
