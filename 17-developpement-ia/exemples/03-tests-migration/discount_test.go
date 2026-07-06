/* ============================================================================
   Section 17.3 : Génération de tests par IA
   Description : Le test IDIOMATIQUE : table-driven avec sous-tests t.Run
                 (§13.1). Chaque cas AFFIRME un comportement attendu
                 (indépendant de l'implémentation) — contrairement au test
                 tautologique qui refait le calcul et « passe » même si le code
                 est faux. Ce test échouerait si Discount était cassé : c'est
                 le critère d'un bon test.
   Fichier source : 17.3 (03-tests-migration-ia.md)
   Lancer : go test -v ./...
   ============================================================================ */

package main

import "testing"

func TestDiscount(t *testing.T) {
	tests := []struct {
		name    string
		price   int
		pct     int
		want    int
		wantErr bool
	}{
		{"remise simple", 100, 10, 90, false},
		{"remise nulle", 100, 0, 100, false},
		{"remise totale", 100, 100, 0, false},
		{"pourcentage invalide", 100, 150, 0, true},
		{"pourcentage négatif", 100, -5, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Discount(tt.price, tt.pct)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Discount() err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Discount() = %d, want %d", got, tt.want)
			}
		})
	}
}
