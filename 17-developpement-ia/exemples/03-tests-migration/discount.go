/* ============================================================================
   Section 17.3 : Génération de tests, migration assistée, revue de code par IA
   Description : Le code testé par le banc table-driven (discount_test.go). Le
                 vrai enseignement est dans le test : le PIÈGE N°1 de la
                 génération de tests par IA est de tester le comportement
                 ACTUEL (bugs compris) plutôt que le comportement CORRECT — un
                 bon test échoue si l'on casse le code.
   Fichier source : 17.3 (03-tests-migration-ia.md)
   ============================================================================ */

package main

import "fmt"

// Discount applique un pourcentage de remise à un prix (0..100 inclus).
func Discount(price, pct int) (int, error) {
	if pct < 0 || pct > 100 {
		return 0, fmt.Errorf("pourcentage invalide : %d", pct)
	}
	return price - price*pct/100, nil
}
