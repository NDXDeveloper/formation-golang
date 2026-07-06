/* ============================================================================
   Section 3.5 : Organisation du code (workspaces)
   Description : L'application qui consomme example.com/lib — module jamais
                 publié : sans le go.work du dossier parent, ce programme ne
                 compile pas ; avec lui, la lib locale est utilisée
   Fichier source : 05-organisation-code.md
   ============================================================================ */

package main

import (
	"fmt"

	"example.com/lib"
)

func main() {
	fmt.Println("app →", lib.Message())
}
