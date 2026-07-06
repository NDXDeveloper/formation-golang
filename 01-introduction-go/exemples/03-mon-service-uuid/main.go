/* ============================================================================
   Section 1.3 : L'écosystème Go (toolchain, modules, cycle de release)
   Description : Programme complet construit autour du go.mod d'exemple de la
                 section (extrait → projet exécutable) : consomme la dépendance
                 github.com/google/uuid v1.6.0 et affiche un UUID
   Fichier source : 03-ecosysteme-go.md
   ============================================================================ */

package main

import (
	"fmt"

	// Dépendance EXTERNE : déclarée dans go.mod (require … v1.6.0), téléchargée
	// via le proxy de modules au premier build, vérifiée par go.sum.
	"github.com/google/uuid"
)

func main() {
	// NewString() génère un UUID v4 aléatoire — différent à chaque exécution.
	// L'appel est qualifié par le NOM du package (uuid.), pas par son chemin.
	fmt.Println(uuid.NewString())
}
