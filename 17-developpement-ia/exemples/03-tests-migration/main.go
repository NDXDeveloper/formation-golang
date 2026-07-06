/* ============================================================================
   Section 17.3 : Génération de tests, migration assistée, revue de code par IA
   Description : Point d'entrée qui exerce le code des trois tâches vérifiables
                 de la section : Discount (testé en table-driven) et LoadUser
                 (migration exception→erreur). La vérification réelle est dans
                 « go test » ; ce main donne un aperçu à l'exécution.
   Fichier source : 17.3 (03-tests-migration-ia.md)
   Lancer : go run .   ·   Tester : go test -v ./...
   ============================================================================ */

package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	// Discount : un cas valide et un cas d'erreur.
	prix, _ := Discount(100, 10)
	_, err := Discount(100, 150)
	fmt.Printf("Discount(100,10)=%d · Discount(100,150) rejeté=%t\n", prix, err != nil)

	// LoadUser : présent, puis absent (errors.Is sur la sentinelle).
	ctx := context.Background()
	u, _ := LoadUser(ctx, 1)
	fmt.Printf("LoadUser(1) = %+v\n", *u)

	_, err = LoadUser(ctx, 99)
	fmt.Printf("LoadUser(99) → errors.Is(ErrNotFound)=%t : %v\n", errors.Is(err, ErrNotFound), err)
}
