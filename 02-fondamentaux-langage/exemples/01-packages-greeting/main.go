/* ============================================================================
   Section 2.1 : Structure d'un programme, packages, visibilité
   Description : Programme consommant le package greeting — import par chemin,
                 appel qualifié, visibilité (defaultName inaccessible ici),
                 preuve de l'ordre d'initialisation (init avant main)
   Fichier source : 01-structure-packages.md
   ============================================================================ */

package main

import (
	"fmt"

	"github.com/exemple/hello/greeting"
)

func main() {
	fmt.Println(greeting.Hello("Ada"))
	fmt.Println(greeting.Hello("")) // nom vide → defaultName, via l'API exportée
	// greeting.defaultName // ← erreur : identifiant non exporté, inaccessible ici

	fmt.Println("init exécuté avant main :", initRan)
	fmt.Println("startedAt initialisée   :", !startedAt.IsZero())
}
