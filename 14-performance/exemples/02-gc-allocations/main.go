/* ============================================================================
   Section 14.2 : Le GC de Go, allocations, escape analysis
   Description : Un programme qui ALLOUE en boucle pour rendre le GC observable.
                 Lancé avec « GODEBUG=gctrace=1 », il imprime une ligne par
                 cycle de GC (tailles de tas, pauses, part CPU). Lancé avec
                 « GOMEMLIMIT=32MiB GOGC=off », il montre que la limite mémoire
                 souple (§14.2) force malgré tout le GC à travailler pour
                 éviter l'OOM. Utilise le package escape/ pour ancrer l'exemple
                 d'escape analysis au même module.
   Fichier source : 02-gc-allocations.md
   Lancer : GODEBUG=gctrace=1 go run .
   Limite : GOMEMLIMIT=32MiB GOGC=off GODEBUG=gctrace=1 go run .
   ============================================================================ */

package main

import (
	"fmt"

	"github.com/exemple/gcallocations/escape"
)

func main() {
	// escape analysis : appelle les fonctions du package escape/.
	pts := []escape.Point{{X: 1, Y: 2}, {X: 3, Y: 4}}
	fmt.Println("SumX =", escape.SumX(pts), "· NewPoint =", *escape.NewPoint())

	// pression mémoire : alloue et relâche pour déclencher des cycles de GC.
	var sink [][]byte
	for i := 0; i < 5000; i++ {
		sink = append(sink, make([]byte, 64*1024)) // 64 KiB par tour
		if len(sink) > 100 {
			sink = sink[50:] // laisse le GC récupérer la moitié
		}
	}
	fmt.Println("terminé (avec GODEBUG=gctrace=1, une ligne de GC par cycle ci-dessus)")
}
