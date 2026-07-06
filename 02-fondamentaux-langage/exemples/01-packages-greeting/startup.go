/* ============================================================================
   Section 2.1 : Structure d'un programme, packages, visibilité
   Description : Initialisation d'un programme — variable de niveau package
                 (startedAt) puis fonction init, exécutées avant main
   Fichier source : 01-structure-packages.md
   ============================================================================ */

package main

import "time"

var startedAt = time.Now() // initialisée avant main

func init() {
	// mise en place ponctuelle, exécutée une fois au démarrage
	initRan = true // témoin observé par main
}

var initRan bool
