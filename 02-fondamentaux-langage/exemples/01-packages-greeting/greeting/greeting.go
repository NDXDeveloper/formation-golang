/* ============================================================================
   Section 2.1 : Structure d'un programme, packages, visibilité
   Description : Le package « greeting » du chapitre — anatomie d'un fichier Go
                 (clause package, imports, déclarations), identifiant exporté
                 (Hello) vs non exporté (defaultName)
   Fichier source : 01-structure-packages.md
   ============================================================================ */

// Package greeting fournit des salutations simples.
package greeting

import (
	"fmt"
	"strings"
)

const defaultName = "le monde" // non exporté (minuscule)

// Hello renvoie une salutation personnalisée.
func Hello(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}
	return fmt.Sprintf("Bonjour, %s !", name)
}
