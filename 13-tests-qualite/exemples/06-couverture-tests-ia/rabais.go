/* ============================================================================
   Section 13.6 : Couverture de code ; génération de tests par IA
   Description : La démonstration phare de la section — une fonction avec une
                 FRONTIÈRE (seuil de 10, inclus). Son test (rabais_test.go)
                 atteint 100 % de COUVERTURE, mais n'épingle pas la frontière
                 exacte : le mutation testing (gremlins) le révèle en faisant
                 SURVIVRE le mutant « >= » → « > ». Couverture ≠ correction.
   Fichier source : 06-couverture-tests-ia.md
   ============================================================================ */

package rabais

// EligibleAuRabais : seuil de 10, INCLUS (>= 10).
func EligibleAuRabais(qté int) bool { return qté >= 10 }
