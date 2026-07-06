/* ============================================================================
   Section 2.4 : Structures conditionnelles
   Description : if / else if / else, if avec instruction d'initialisation
                 (portée confinée), switch (cas multiples, initialisation,
                 forme sans expression) et type switch
   Fichier source : 04-conditions.md
   ============================================================================ */

package main

import (
	"fmt"
	"runtime"
)

const maxLen = 10

// if AVEC INITIALISATION : n est déclarée juste avant la condition et
// n'existe QUE dans le bloc if/else — le reste de la fonction reste propre.
// C'est l'idiome central de la gestion d'erreurs (if err := …; err != nil).
func checkLen(s string) error {
	if n := len(s); n > maxLen {
		return fmt.Errorf("trop long : %d caractères", n)
	}
	// n n'existe plus ici : sa portée était limitée au if
	return nil
}

// switch CLASSIQUE : pas de fallthrough implicite (chaque case s'arrête seul),
// et un case peut regrouper PLUSIEURS valeurs.
func jourMessage(day string) string {
	switch day {
	case "samedi", "dimanche": // deux valeurs pour un même cas
		return "week-end"
	case "vendredi":
		return "presque !"
	default:
		return "au travail"
	}
}

// switch SANS EXPRESSION : chaque case porte une condition booléenne —
// une chaîne if/else if plus lisible.
func grade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	default:
		return "C"
	}
}

// TYPE SWITCH : aiguille selon le type DYNAMIQUE d'une valeur any.
// Dans chaque case, v prend le type concret correspondant.
func describe(x any) string {
	switch v := x.(type) {
	case int:
		return fmt.Sprintf("entier : %d", v) // ici v est un int
	case string:
		return fmt.Sprintf("chaîne de %d caractères", len(v)) // ici un string
	case bool:
		return fmt.Sprintf("booléen : %t", v)
	default:
		return fmt.Sprintf("type inconnu : %T", v) // %T affiche le type réel
	}
}

func main() {
	fmt.Println("=== if / else if / else ===")
	score := 85
	var g string
	if score >= 90 { // pas de parenthèses ; accolades obligatoires
		g = "A"
	} else if score >= 80 {
		g = "B"
	} else {
		g = "C"
	}
	fmt.Println("score", score, "→ note", g)

	fmt.Println("=== if avec initialisation ===")
	fmt.Println(`checkLen("court") →`, checkLen("court"))
	fmt.Println(`checkLen(long)    →`, checkLen("bien trop long pour la limite"))

	fmt.Println("=== switch ===")
	for _, d := range []string{"samedi", "vendredi", "lundi"} {
		fmt.Println(d, "→", jourMessage(d))
	}

	fmt.Println("=== switch avec initialisation ===")
	// Comme le if, le switch accepte une instruction d'initialisation :
	// os n'existe que dans le switch.
	switch os := runtime.GOOS; os {
	case "linux":
		fmt.Println("système :", os)
	case "darwin":
		fmt.Println("système :", os)
	default:
		fmt.Println("système :", os)
	}

	fmt.Println("=== switch sans expression ===")
	fmt.Println("grade(85) =", grade(85))

	fmt.Println("=== type switch ===")
	// Une tranche de any : chaque valeur garde son type dynamique.
	for _, x := range []any{42, "go", true, 3.14} {
		fmt.Println(describe(x))
	}
}
