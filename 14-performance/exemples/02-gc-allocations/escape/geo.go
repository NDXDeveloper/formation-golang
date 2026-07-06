/* ============================================================================
   Section 14.2 : Le GC de Go, allocations, escape analysis
   Description : L'escape analysis (§14.2) rendue visible. « go build
                 -gcflags=-m » montre les décisions du compilateur : newPoint
                 renvoie l'adresse d'une locale → « moved to heap: p » ; sumX
                 ne fait que lire son paramètre → « pts does not escape » (il
                 reste sur la pile). L'échappement n'est pas un défaut : il est
                 souvent nécessaire ; l'objectif est d'éviter les allocations
                 tas INUTILES sur le chemin chaud.
   Fichier source : 02-gc-allocations.md
   Vérifier : go build -gcflags=-m ./escape/
   ============================================================================ */

package escape

type Point struct{ X, Y int }

// p s'échappe : on renvoie son adresse, sa durée de vie dépasse newPoint.
func NewPoint() *Point {
	p := Point{X: 1, Y: 2}
	return &p
}

// p ne s'échappe pas : il naît et meurt dans la boucle, sur la pile.
func SumX(pts []Point) int {
	total := 0
	for _, p := range pts {
		total += p.X
	}
	return total
}
