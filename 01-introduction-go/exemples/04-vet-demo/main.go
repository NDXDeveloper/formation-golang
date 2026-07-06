/* ============================================================================
   Section 1.4 : Installation et outils
   Description : Illustration de l'affirmation « go vet attrape des erreurs que
                 la compilation laisse passer (un mauvais format Printf, par
                 exemple) » — ce fichier COMPILE, mais `go vet ./...` le signale
   Fichier source : 04-installation-outils.md
   ============================================================================ */

package main

import "fmt"

func main() {
	fmt.Printf("%d\n", "oops") // %d attend un entier, reçoit une chaîne
}
