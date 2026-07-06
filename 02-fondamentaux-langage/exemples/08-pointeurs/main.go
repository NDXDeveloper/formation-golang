/* ============================================================================
   Section 2.8 : Pointeurs, sémantique valeur vs référence
   Description : Adresse (&) et déréférencement (*), tout est passé par valeur
                 (copie vs pointeur), le cas des slices (élément modifié visible,
                 append invisible), créer un pointeur (&T{}, new, new(expr) 1.26)
   Fichier source : 08-pointeurs.md
   ============================================================================ */

package main

import "fmt"

type Point struct{ X, Y int }

// Reçoit une COPIE de n : la multiplier ne change rien chez l'appelant.
func doubleValeur(n int) { n *= 2 }

// Reçoit l'ADRESSE de n : *n modifie la variable originale de l'appelant.
func doublePointeur(n *int) { *n *= 2 }

// Une slice transporte un POINTEUR vers son tableau : modifier un élément
// existant est visible par l'appelant…
func remplir(s []int) { s[0] = 99 }

// … mais append réaffecte l'EN-TÊTE local (une copie) : l'appelant ne voit
// rien. Pour rendre l'append visible : renvoyer la slice, ou passer *[]int.
func ajouter(s []int) { s = append(s, 1) }

func main() {
	fmt.Println("=== Adresse et déréférencement ===")
	x := 42
	p := &x                 // & prend l'adresse : p est un *int
	fmt.Println("*p =", *p) // * lit la valeur pointée
	*p = 7                  // * en écriture : modifie x à travers p
	fmt.Println("x après *p = 7 :", x)
	var q *int // zéro-value d'un pointeur : nil (le déréférencer paniquerait)
	fmt.Println("zéro-value d'un pointeur :", q == nil)

	fmt.Println("=== Tout est passé par valeur ===")
	x2 := 10
	doubleValeur(x2) // la copie est doublée… puis jetée
	fmt.Println("après doubleValeur  :", x2)
	doublePointeur(&x2) // on passe l'ADRESSE : la fonction modifie l'original
	fmt.Println("après doublePointeur :", x2)

	fmt.Println("=== Slices : élément visible, append invisible ===")
	nums := []int{0, 0, 0}
	remplir(nums) // écrit dans le tableau partagé → visible
	fmt.Println("après remplir :", nums)
	ajouter(nums) // n'a modifié que la copie de l'en-tête → invisible
	fmt.Println("après ajouter :", nums, " ← l'append n'a PAS d'effet ici")

	fmt.Println("=== Créer un pointeur ===")
	pt := &Point{X: 1, Y: 2} // &T{...} : alloue ET initialise en une expression
	n0 := new(int)           // new(T) : alloue une zéro-value, renvoie *T
	fmt.Println("&Point{1,2} →", *pt, "· *new(int) =", *n0)

	fmt.Println("=== new(expr) — Go 1.26 ===")
	// 🆕 new accepte une expression : fini les fonctions utilitaires « ptr() »
	// pour obtenir un pointeur vers un littéral (champs optionnels, etc.).
	r := new(42)
	fmt.Println("*new(42) =", *r)
}
