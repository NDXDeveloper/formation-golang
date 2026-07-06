/* ============================================================================
   Section 11.1 : cgo (quand l'éviter), FFI
   Description : Le programme C qui consomme la bibliothèque Go — il inclut
                 l'en-tête GÉNÉRÉ par le build c-shared (libgo.h) et appelle
                 Additionner et Saluer comme des fonctions C ordinaires.
   Fichier source : 01-cgo-ffi.md
   Construire : gcc main.c -o prog_c -I.. -L.. -lgo -Wl,-rpath,'$ORIGIN/..'
   (main.c vit dans consommateur/ : s'il restait à côté de lib.go, go build
   le compilerait comme fichier cgo du package AVANT que libgo.h n'existe)
   ============================================================================ */

#include <stdio.h>
#include "libgo.h"

int main(void) {
    printf("C appelle Go : Additionner(19, 23) = %d\n", (int)Additionner(19, 23));
    Saluer();
    return 0;
}
