/* ============================================================================
   Section 13.6 : Couverture de code ; génération de tests par IA
   Description : Le test « couvrant » du cours — il exécute 100 % des lignes
                 (CI verte, `go test -cover` affiche 100 %), MAIS il teste 15
                 et 3, jamais la FRONTIÈRE (qté == 10). Si l'on remplaçait
                 « >= 10 » par « > 10 », ces deux cas passeraient ENCORE : le
                 test ne s'apercevrait pas du bug. C'est ce mutant survivant
                 que gremlins révèle (voir le README). Un bon test épingle la
                 frontière (ajouter les cas 9, 10, 11).
   Fichier source : 06-couverture-tests-ia.md
   Lancer : go test -cover ./...
   Mutation : gremlins unleash ./...   (le mutant CONDITIONALS_BOUNDARY survit)
   ============================================================================ */

package rabais

import "testing"

func TestEligibleAuRabais(t *testing.T) {
	if !EligibleAuRabais(15) {
		t.Error("15 devrait être éligible")
	}
	if EligibleAuRabais(3) {
		t.Error("3 ne devrait pas l'être")
	}
	// ⚠️ manque : les cas 9, 10, 11 qui épingleraient la frontière == 10.
}
