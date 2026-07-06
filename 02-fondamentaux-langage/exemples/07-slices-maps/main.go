/* ============================================================================
   Section 2.7 : Tableaux, slices et maps
   Description : Tableaux (copie par valeur), slices (longueur/capacité,
                 append et réallocation), les pièges du partage (vue, append
                 qui écrase, découpage à trois indices, slices.Clone), clear,
                 et les maps (comma-ok, delete, map nil)
   Fichier source : 07-slices-maps.md
   ============================================================================ */

package main

import (
	"fmt"
	"slices"
)

func main() {
	fmt.Println("=== Tableaux : la copie est complète ===")
	// La taille fait partie du TYPE ([3]int ≠ [4]int) ; l'affectation COPIE tout.
	b := [3]int{1, 2, 3}
	d := b   // copie intégrale : d est indépendant de b
	d[0] = 9 // ne touche QUE la copie
	fmt.Println("b =", b, "· d =", d)

	fmt.Println("=== Slices : longueur et capacité ===")
	// Une slice = un descripteur (pointeur, longueur, capacité) sur un tableau.
	s := []int{1, 2, 3}     // littéral : len 3, cap 3
	t := make([]int, 2)     // longueur 2 (remplie de zéros), cap 2
	u := make([]int, 0, 10) // longueur 0 mais capacité 10 : PRÉALLOUÉE
	fmt.Println("littéral :", len(s), cap(s), "· make(2) :", len(t), cap(t), "· make(0,10) :", len(u), cap(u))

	fmt.Println("=== append et la capacité ===")
	s2 := make([]int, 0, 2)
	fmt.Println("départ      :", len(s2), cap(s2))
	s2 = append(s2, 1, 2) // tient dans la capacité : écrit dans le tableau existant
	fmt.Println("append(1,2) :", len(s2), cap(s2))
	s2 = append(s2, 3) // capacité dépassée → NOUVEAU tableau, plus grand, recopié
	fmt.Println("append(3)   :", len(s2), cap(s2), " ← réallocation (toujours réaffecter le résultat d'append !)")

	fmt.Println("=== Piège : le partage du tableau sous-jacent ===")
	// Découper ne copie RIEN : la sous-slice regarde le même tableau.
	base := []int{10, 20, 30, 40}
	vue := base[1:3] // vue = [20 30], MÊME tableau que base
	vue[0] = 99      // modifie le tableau partagé…
	fmt.Println("base après vue[0]=99 :", base, " ← base a changé !")

	fmt.Println("=== Piège : append qui écrase ===")
	a := []int{1, 2, 3, 4, 5}
	b2 := a[:2]         // len 2, cap 5 : il RESTE de la place dans le tableau de a
	b2 = append(b2, 99) // pas de réallocation → écrit PAR-DESSUS a[2]
	fmt.Println("a après append sur b2 :", a, " ← a[2] écrasé !")

	fmt.Println("=== Parade : découpage à trois indices ===")
	a2 := []int{1, 2, 3, 4, 5}
	c2 := a2[0:2:2]     // le 3e indice BORNE la capacité à 2
	c2 = append(c2, 99) // capacité pleine → réallocation forcée, a2 intact
	fmt.Println("a2 préservé :", a2, "· c2 :", c2)

	fmt.Println("=== Parade : slices.Clone (copie détachée) ===")
	// Garder une petite sous-slice d'un GRAND tableau retient tout le tableau
	// en mémoire ; Clone fabrique une copie indépendante, le grand est libérable.
	gros := make([]int, 1000)
	gros[0], gros[1], gros[2] = 1, 2, 3
	petit := slices.Clone(gros[:3])
	fmt.Println("petit =", petit, "· capacité détachée :", cap(petit) < 1000)

	fmt.Println("=== Le package slices ===")
	nums := []int{3, 1, 2}
	slices.Sort(nums) // tri EN PLACE
	fmt.Println("Sort →", nums, "· Contains(2) →", slices.Contains(nums, 2))

	fmt.Println("=== Maps ===")
	m := map[string]int{"a": 1, "b": 2}
	m["c"] = 3            // ajout
	absent := m["absent"] // clé absente → zéro-value, PAS d'erreur ni panique
	v, ok := m["a"]       // le motif « comma-ok » distingue absent de zéro
	delete(m, "b")        // suppression (silencieuse si la clé n'existe pas)
	fmt.Println("absent =", absent, "· a =", v, ok, "· len =", len(m))

	// clear (Go 1.21) : vider une map / remettre une slice à zéro.
	clear(m)
	fmt.Println("après clear : len =", len(m))
	cs := []int{1, 2, 3}
	clear(cs) // pour une slice : chaque élément repasse à sa zéro-value
	fmt.Println("clear sur slice :", cs, " (longueur inchangée)")

	// Une map NIL se lit sans problème… mais y ÉCRIRE paniquerait.
	var p map[string]int
	fmt.Println("lecture map nil :", p["x"])
	// p["x"] = 1 // PANIQUE : assignment to entry in nil map — d'où make() d'abord
}
