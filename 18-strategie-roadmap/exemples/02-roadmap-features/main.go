/* ============================================================================
   Section 18.2 : Roadmap — les génériques « terminés avec soin »
   Description : Illustration de l'arc « finir les génériques proprement » de la
                 section, sur une nouveauté concrète de Go 1.26 : errors.AsType.
                 Là où errors.As exige de déclarer une variable de sortie et de
                 lui passer son adresse (« la danse du pointeur »), la forme
                 générique errors.AsType[E](err) renvoie directement (E, bool).
                 Les méthodes génériques — la dernière grande pièce — arrivent
                 avec Go 1.27 : voir le sous-dossier ./generics127.
   Fichier source : 18.2 (02-roadmap.md)
   Lancer : go run .
   ============================================================================ */

package main

import (
	"errors"
	"fmt"
)

type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return "champ " + e.Field }

func main() {
	err := fmt.Errorf("échec de validation : %w", &ValidationError{Field: "email"})

	// AVANT (errors.As) : déclarer la variable, passer son adresse.
	var target *ValidationError
	if errors.As(err, &target) {
		fmt.Println("errors.As      →", target.Field, "(nécessite &target)")
	}

	// Go 1.26 (errors.AsType) : la forme générique renvoie directement la valeur.
	if ve, ok := errors.AsType[*ValidationError](err); ok {
		fmt.Println("errors.AsType  →", ve.Field, "(plus de danse du pointeur)")
	}
}
