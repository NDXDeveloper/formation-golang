/* ============================================================================
   Section 18.1 : GODEBUG — faire évoluer le langage sans casser le code
   Description : Le mécanisme de compatibilité de Go, rendu tangible sur son
                 exemple canonique : panic(nil). Depuis Go 1.21, panic(nil)
                 déclenche un *runtime.PanicNilError (pour que recover()
                 indique de façon fiable si l'on panique) — un changement
                 « compatible mais cassant », donc adossé à un réglage GODEBUG.
                 Le défaut suit la ligne `go` du go.mod (ici « go 1.26.0 »).
                 L'ancien comportement (recover() renvoie nil) se restaure de
                 TROIS façons, du plus large au plus ciblé (voir le README) :
                   - variable d'environnement : GODEBUG=panicnil=1
                   - directive en tête du main : //go:debug panicnil=1  (Go 1.21)
                   - ligne dans go.mod :          godebug panicnil=1     (Go 1.23)
   Fichier source : 18.1 (01-gouvernance-compatibilite.md)
   Lancer : go run .   (défaut)   ·   GODEBUG=panicnil=1 go run .   (ancien)
   ============================================================================ */

package main

import "fmt"

func main() {
	defer func() {
		r := recover()
		fmt.Printf("recover() renvoie : %T (valeur : %v)\n", r, r)
		if r != nil {
			fmt.Println("→ comportement Go 1.21+ : panic(nil) est un *runtime.PanicNilError")
		} else {
			fmt.Println("→ ancien comportement restauré (GODEBUG=panicnil=1) : recover() renvoie nil")
		}
	}()

	panic(nil) // depuis Go 1.21 : *runtime.PanicNilError, sauf GODEBUG=panicnil=1
}
