/* ============================================================================
   Section 15.2 : CI/CD (GitHub Actions, GitLab CI)
   Description : Le test que la CI exécute sous « go test -race
                 -coverprofile=cover.out ». Sa couverture est remontée dans
                 l'UI de la merge request (GitLab) via l'expression régulière
                 du champ coverage:.
   Fichier source : 02-cicd.md
   Lancer : go test -race -coverprofile=cover.out ./... && go tool cover -func=cover.out
   ============================================================================ */

package cicd

import "testing"

func TestAdd(t *testing.T) {
	cases := []struct{ a, b, want int }{
		{2, 3, 5},
		{-1, 1, 0},
		{0, 0, 0},
	}
	for _, c := range cases {
		if got := Add(c.a, c.b); got != c.want {
			t.Errorf("Add(%d,%d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
