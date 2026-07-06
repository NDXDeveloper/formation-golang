/* ============================================================================
   Section 2.5 : Boucles et itérateurs
   Description : Le for unique (3 formes), break étiqueté, variable de boucle
                 par itération (Go 1.22), range sur slice/chaîne/entier, et
                 itérateur range-over-func avec iter.Seq (Go 1.23)
   Fichier source : 05-boucles.md
   ============================================================================ */

package main

import (
	"fmt"
	"iter"
	"slices"
)

// Compte renvoie un ITÉRATEUR (Go 1.23) : une fonction que range sait piloter.
// Le principe : l'itérateur PRODUIT chaque valeur en appelant yield ; si la
// boucle appelante s'arrête (break), yield renvoie false et on abandonne.
func Compte(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < n; i++ {
			if !yield(i) { // false = le range d'en face a fait break
				return
			}
		}
	}
}

func main() {
	fmt.Println("=== Les trois formes du for ===")
	// 1. Forme complète (init ; condition ; post) — le for « classique ».
	somme := 0
	for i := 0; i < 3; i++ {
		somme += i
	}
	fmt.Println("somme 0+1+2 =", somme)
	// 2. Condition seule : l'équivalent du while des autres langages.
	cond, tours := true, 0
	for cond {
		tours++
		cond = false
	}
	fmt.Println("tours =", tours)
	// 3. Boucle infinie — toujours bornée par un break quelque part.
	for {
		break
	}
	fmt.Println("boucle infinie quittée par break")

	fmt.Println("=== break étiqueté (sort des deux boucles) ===")
	grille := [][]int{{1, 2}, {3, 4}}
	cible, visites := 3, 0
Recherche: // l'étiquette nomme la boucle EXTERNE
	for _, ligne := range grille {
		for _, cellule := range ligne {
			visites++
			if cellule == cible {
				break Recherche // sort des DEUX boucles d'un coup
			}
		}
	}
	fmt.Println("cible", cible, "trouvée après", visites, "cellules visitées")

	fmt.Println("=== Variable de boucle par itération (Go 1.22) ===")
	// Chaque itération a SON i : les closures capturent des variables distinctes.
	// Avant Go 1.22, toutes partageaient le même i et affichaient souvent 3 3 3.
	funcs := make([]func(), 3)
	for i := 0; i < 3; i++ {
		funcs[i] = func() { fmt.Print(i, " ") } // capture l'i de CETTE itération
	}
	for _, f := range funcs {
		f() // 0 1 2 — et non 3 3 3
	}
	fmt.Println()

	fmt.Println("=== range sur une chaîne (indices d'octet, runes) ===")
	// i est l'indice d'OCTET (il saute après le « é » multi-octets), r la rune.
	for i, r := range "héllo" {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()

	fmt.Println("=== range sur un entier (Go 1.22) ===")
	for i := range 3 { // équivaut à : for i := 0; i < 3; i++
		fmt.Print(i, " ")
	}
	fmt.Println()

	fmt.Println("=== Itérateur range-over-func (Go 1.23) ===")
	// La même syntaxe range que pour un slice… sur NOTRE itérateur.
	for v := range Compte(3) {
		fmt.Print(v, " ")
	}
	fmt.Println()

	fmt.Println("=== slices.Values : itérer sur un slice ===")
	// La stdlib fournit déjà des itérateurs prêts à l'emploi.
	nums := []int{10, 20}
	total := 0
	for v := range slices.Values(nums) {
		total += v
	}
	fmt.Println("total =", total)
}
