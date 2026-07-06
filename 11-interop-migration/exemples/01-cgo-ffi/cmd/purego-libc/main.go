/* ============================================================================
   Section 11.1 : cgo (quand l'éviter), FFI
   Description : L'alternative purego de la section — appeler une bibliothèque
                 partagée du système (ici la libc) SANS cgo : chargement à
                 l'exécution (dlopen/dlsym), aucun compilateur C requis, et
                 le build reste 100 % Go (CGO_ENABLED=0 fonctionne) — on
                 conserve cross-compilation et builds rapides. Réserves du
                 cours : projet bêta, pas de filet mémoire, ABI sous votre
                 responsabilité. NB : la sortie de puts passe par le stdout
                 de la LIBC (line-buffered en terminal ; en pipe, le buffer C
                 peut ne pas être flushé — lancez dans un vrai terminal).
   Fichier source : 01-cgo-ffi.md
   Lancer : CGO_ENABLED=0 go run ./cmd/purego-libc     (Linux)
   ============================================================================ */

package main

import (
	"github.com/ebitengine/purego"
)

func main() {
	// Charge la libc à l'exécution — aucune chaîne C, build 100 % Go.
	libc, err := purego.Dlopen("libc.so.6", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}
	var puts func(string)
	purego.RegisterLibFunc(&puts, libc, "puts")
	puts("Appel de C depuis Go, sans cgo")
}
