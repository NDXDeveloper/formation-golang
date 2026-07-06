/* ============================================================================
   Section 11.1 : cgo (quand l'éviter), FFI
   Description : L'AUTRE SENS de la frontière — exposer du Go à C : chaque
                 fonction marquée //export devient visible depuis le C, et
                 `go build -buildmode=c-shared` produit une bibliothèque
                 partagée (libgo.so) AVEC son en-tête généré (libgo.h). Le
                 programme main.c ci-contre la consomme comme n'importe
                 quelle bibliothèque C — c'est ainsi qu'on offre du Go à
                 Python (ctypes), Ruby (FFI) ou un hôte C existant.
                 (Variante -buildmode=c-archive : lien statique.)
   Fichier source : 01-cgo-ffi.md
   Construire : voir le README (go build -buildmode=c-shared, puis gcc)
   ============================================================================ */

package main

import "C"
import "fmt"

//export Additionner
func Additionner(a, b C.int) C.int {
	return a + b
}

//export Saluer
func Saluer() {
	fmt.Println("bonjour depuis la bibliothèque Go")
}

func main() {} // requis en buildmode c-shared/c-archive (jamais exécuté par l'hôte C)
