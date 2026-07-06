/* ============================================================================
   Section 18.2 §3 : Ce qui se dessine — les méthodes génériques (Go 1.27)
   Description : La dernière grande pièce manquante des génériques. Jusqu'ici,
                 une méthode ne pouvait PAS déclarer ses propres paramètres de
                 type — il fallait passer par une fonction de paquet. Go 1.27
                 lève cette limite : Map est une MÉTHODE qui déclare son propre
                 paramètre de type U. (Proposal #77273, acceptée en janvier
                 2026, livrée en 1.27.)
   Fichier source : 18.2 (02-roadmap.md)
   Prérequis : Go 1.27 (RC ou stable). Sous une toolchain < 1.27, ce fichier
               NE COMPILE PAS (les méthodes à paramètres de type sont refusées).
   Lancer : GOTOOLCHAIN=go1.27rc1 go run .   (ou go1.27 stable une fois sortie)
   ============================================================================ */

package main

import "fmt"

type Stream[T any] struct{ items []T }

// Map : une méthode qui déclare son propre paramètre de type U (Go 1.27).
func (s Stream[T]) Map[U any](f func(T) U) Stream[U] {
	out := make([]U, len(s.items))
	for i, v := range s.items {
		out[i] = f(v)
	}
	return Stream[U]{out}
}

func main() {
	s := Stream[int]{items: []int{1, 2, 3}}
	doubled := s.Map(func(n int) string { return fmt.Sprintf("<%d>", n*2) })
	fmt.Println(doubled) // {[<2> <4> <6>]}
}
