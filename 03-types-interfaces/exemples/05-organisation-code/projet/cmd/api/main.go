/* ============================================================================
   Section 3.5 : Organisation du code : packages, internal/, layout, workspaces
   Description : Layout conventionnel — cmd/api (un binaire) qui importe le
                 package privé internal/store par son chemin depuis la racine
                 du module
   Fichier source : 05-organisation-code.md
   ============================================================================ */

package main

import (
	"fmt"

	"github.com/acme/app/internal/store"
)

func main() {
	s := store.New()
	fmt.Println(s.Greet())
}
