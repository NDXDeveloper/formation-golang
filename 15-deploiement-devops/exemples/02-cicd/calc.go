/* ============================================================================
   Section 15.2 : CI/CD (GitHub Actions, GitLab CI)
   Description : Du code minimal — juste ce qu'il faut pour que le pipeline ait
                 quelque chose à compiler, tester (avec couverture) et scanner.
                 Le cœur de cet exemple est ailleurs : les workflows
                 .github/workflows/ci.yml et .gitlab-ci.yml, qui exécutent la
                 chaîne go mod verify → vet → test -race → lint → govulncheck.
   Fichier source : 02-cicd.md
   ============================================================================ */

package cicd

// Add additionne deux entiers (support de la CI : compilation + test + couverture).
func Add(a, b int) int { return a + b }
