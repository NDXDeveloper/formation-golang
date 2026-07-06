/* ============================================================================
   Section 11.1 : cgo (quand l'éviter), FFI
   Description : Le bloc cgo de la section — le préambule C (compilé comme du
                 C, collé SANS ligne vide à l'import "C"), l'appel C.c_strlen
                 depuis Go, et la discipline mémoire : C.CString ALLOUE de la
                 mémoire C que le ramasse-miettes de Go ne voit pas — c'est à
                 vous de C.free (d'où le defer). Nécessite un compilateur C
                 (CGO_ENABLED=1) ; avec CGO_ENABLED=0, ce fichier est REFUSÉ
                 au build (« build constraints exclude ») — c'est le point de
                 bascule du cours.
   Fichier source : 01-cgo-ffi.md
   Lancer : CGO_ENABLED=1 go run ./cmd/strlen-cgo
   ============================================================================ */

package main

/*
#include <stdlib.h>
#include <string.h>

static size_t c_strlen(const char* s) { return strlen(s); }
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	cs := C.CString("héllo") // alloue de la mémoire C — À LIBÉRER soi-même
	defer C.free(unsafe.Pointer(cs))

	n := C.c_strlen(cs) // appel du C depuis Go
	fmt.Println(int(n)) // 6 : « héllo » fait 6 octets en UTF-8
}
