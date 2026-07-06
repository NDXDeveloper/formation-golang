/* ============================================================================
   Section 14.3 : PGO — l'optimisation guidée par profil
   Description : PGO ne demande AUCUN changement de code : on dépose un profil
                 CPU représentatif nommé « default.pgo » dans le répertoire du
                 package main, et « go build » le détecte et l'utilise
                 automatiquement (défaut -pgo=auto). Le compilateur s'en sert
                 pour un inlining et une dévirtualisation plus agressifs sur
                 les chemins réellement chauds. Ici, chaud() via une interface
                 (Op) donne à PGO de quoi dévirtualiser.
   Fichier source : 03-optimisations-pgo.md
   Construire avec PGO : go build .   (détecte default.pgo — cf. go build -x)
   ============================================================================ */

package main

import "fmt"

// Op : un appel de méthode d'interface — cible de la dévirtualisation PGO.
type Op interface{ Apply(int) int }

type mod7 struct{}

func (mod7) Apply(x int) int { return x % 7 }

func chaud(op Op, n int) int {
	acc := 0
	for i := 0; i < n; i++ {
		acc += op.Apply(i) // appel virtuel : PGO peut le dévirtualiser + inliner
	}
	return acc
}

func main() {
	fmt.Println(chaud(mod7{}, 1_000_000))
}
